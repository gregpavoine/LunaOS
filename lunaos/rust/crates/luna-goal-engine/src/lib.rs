use luna_agent_runner::{ActionRunner, LocalActionRunner};
use luna_contracts::{
    CONTRACT_VERSION, CreateGoal, Evidence, ExecutionLease, GoalEventView, GoalLifecycleCommand,
    GoalPlan, GoalPlanView, GoalView, PolicyRequest, Principal, SourceStatus, TickResult,
};
use luna_policy_broker::PolicyService;
use rusqlite::{Connection, OptionalExtension, TransactionBehavior, params};
use sha2::{Digest, Sha256};
use std::{
    fs::OpenOptions,
    path::Path,
    sync::{Arc, Mutex},
    time::{Duration, SystemTime, UNIX_EPOCH},
};
use thiserror::Error;
use uuid::Uuid;

#[derive(Debug, Error)]
pub enum EngineError {
    #[error("invalid request: {0}")]
    Invalid(String),
    #[error("storage error: {0}")]
    Storage(#[from] rusqlite::Error),
    #[error("policy denied: {0}")]
    Policy(String),
    #[error("goal not found")]
    NotFound,
}

#[derive(Clone)]
pub struct GoalEngine {
    db: Arc<Mutex<Connection>>,
    broker: Arc<dyn PolicyService>,
    runner: Arc<dyn ActionRunner>,
}

impl GoalEngine {
    pub fn open<P: PolicyService + 'static>(
        path: impl AsRef<Path>,
        broker: P,
    ) -> Result<Self, EngineError> {
        if let Some(parent) = path.as_ref().parent() {
            std::fs::create_dir_all(parent).map_err(|e| EngineError::Invalid(e.to_string()))?;
        }
        #[cfg(unix)]
        {
            use std::os::unix::fs::OpenOptionsExt;
            OpenOptions::new()
                .create(true)
                .write(true)
                .truncate(false)
                .mode(0o600)
                .open(path.as_ref())
                .map_err(|e| EngineError::Invalid(e.to_string()))?;
        }
        #[cfg(not(unix))]
        OpenOptions::new()
            .create(true)
            .write(true)
            .open(path.as_ref())
            .map_err(|e| EngineError::Invalid(e.to_string()))?;
        let db = Connection::open(path.as_ref())?;
        db.busy_timeout(Duration::from_secs(5))?;
        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt;
            std::fs::set_permissions(path.as_ref(), std::fs::Permissions::from_mode(0o600))
                .map_err(|e| EngineError::Invalid(e.to_string()))?;
        }
        db.pragma_update(None, "journal_mode", "WAL")?;
        db.pragma_update(None, "synchronous", "FULL")?;
        db.pragma_update(None, "foreign_keys", "ON")?;
        db.execute_batch(include_str!("../../../../migrations/native/001_core.sql"))?;
        let lease_migration_applied: bool = db.query_row(
            "SELECT EXISTS(SELECT 1 FROM schema_migration WHERE version=2)",
            [],
            |row| row.get(0),
        )?;
        if !lease_migration_applied {
            db.execute_batch(include_str!(
                "../../../../migrations/native/002_execution_lease.sql"
            ))?;
        }
        let lifecycle_migration_applied: bool = db.query_row(
            "SELECT EXISTS(SELECT 1 FROM schema_migration WHERE version=3)",
            [],
            |row| row.get(0),
        )?;
        if !lifecycle_migration_applied {
            db.execute_batch(include_str!(
                "../../../../migrations/native/003_goal_lifecycle.sql"
            ))?;
        }
        let event_migration_applied: bool = db.query_row(
            "SELECT EXISTS(SELECT 1 FROM schema_migration WHERE version=4)",
            [],
            |row| row.get(0),
        )?;
        if !event_migration_applied {
            db.execute_batch(include_str!(
                "../../../../migrations/native/004_event_outbox.sql"
            ))?;
        }
        let plan_migration_applied: bool = db.query_row(
            "SELECT EXISTS(SELECT 1 FROM schema_migration WHERE version=5)",
            [],
            |row| row.get(0),
        )?;
        if !plan_migration_applied {
            db.execute_batch(include_str!(
                "../../../../migrations/native/005_goal_plans.sql"
            ))?;
        }
        let retry_migration_applied: bool = db.query_row(
            "SELECT EXISTS(SELECT 1 FROM schema_migration WHERE version=6)",
            [],
            |row| row.get(0),
        )?;
        if !retry_migration_applied {
            db.execute_batch(include_str!(
                "../../../../migrations/native/006_lease_retries.sql"
            ))?;
        }
        Ok(Self {
            db: Arc::new(Mutex::new(db)),
            broker: Arc::new(broker),
            runner: Arc::new(LocalActionRunner),
        })
    }

    pub fn create(
        &self,
        body: &CreateGoal,
        owner_token: &str,
        idempotency_key: &str,
    ) -> Result<GoalView, EngineError> {
        validate(body, idempotency_key)?;
        // Do not persist user-authored source text, even as a guessable hash.
        let canonical_request = (
            &body.action,
            &body.metrics,
            &body.privacy_class,
            body.expected_cost_minor,
        );
        let request_hash = format!(
            "{:x}",
            Sha256::digest(
                serde_json::to_vec(&canonical_request)
                    .map_err(|e| EngineError::Invalid(e.to_string()))?
            )
        );
        let mut db = self.db.lock().expect("sqlite mutex poisoned");
        if let Some((id, old_hash)) = db
            .query_row(
                "SELECT id,request_hash FROM goal WHERE idempotency_key=?1",
                [idempotency_key],
                |r| Ok((r.get::<_, String>(0)?, r.get::<_, String>(1)?)),
            )
            .optional()?
        {
            if old_hash != request_hash {
                return Err(EngineError::Invalid("IDEMPOTENCY_CONFLICT".into()));
            }
            return read_goal(&db, &id);
        }
        let id = Uuid::new_v4().to_string();
        let policy_request = PolicyRequest {
            owner_token: owner_token.into(),
            principal: Principal::Owner,
            action: body.action.clone(),
            privacy_class: body.privacy_class.clone(),
            expected_cost_minor: body.expected_cost_minor,
            configured_budget_minor: None,
        };
        let grant = self
            .broker
            .authorize(&policy_request, &id)
            .map_err(|e| EngineError::Policy(e.to_string()))?;
        // This check intentionally precedes score calculation.
        let score = Some(score(&body.metrics));
        let policy_code = "OWNER_SCOPED_GRANT";
        let tx = db.transaction()?;
        tx.execute("INSERT INTO goal VALUES(?1,?2,'approved',?3,?4,'unknown',?5,?6,strftime('%Y-%m-%dT%H:%M:%fZ','now'))", params![id, body.action, score, policy_code, idempotency_key, request_hash])?;
        tx.execute("INSERT INTO goal_grant VALUES(?1,?2)", params![id, grant])?;
        tx.execute("INSERT INTO job(goal_id,state) VALUES(?1,'queued')", [&id])?;
        tx.execute("INSERT INTO goal_event(goal_id,kind,code,created_at_utc) VALUES(?1,'goal.created','OWNER_REQUEST',strftime('%Y-%m-%dT%H:%M:%fZ','now'))", [&id])?;
        tx.commit()?;
        read_goal(&db, &id)
    }

    pub fn claim_next(
        &self,
        worker: &str,
        now_unix: i64,
        lease_seconds: i64,
    ) -> Result<Option<ExecutionLease>, EngineError> {
        if worker.is_empty()
            || worker.len() > 64
            || !worker
                .chars()
                .all(|c| c.is_ascii_alphanumeric() || "_.-".contains(c))
            || !(1..=300).contains(&lease_seconds)
        {
            return Err(EngineError::Invalid("INVALID_LEASE_REQUEST".into()));
        }
        let mut db = self.db.lock().expect("sqlite mutex poisoned");
        let tx = db.transaction_with_behavior(TransactionBehavior::Immediate)?;
        tx.execute(
            "INSERT INTO goal_event(goal_id,kind,code,created_at_utc)
             SELECT goal_id,'job.lease.expired','LEASE_EXPIRED',strftime('%Y-%m-%dT%H:%M:%fZ','now')
             FROM job WHERE state='running' AND lease_expires_at_unix<=?1",
            [now_unix],
        )?;
        tx.execute(
            "UPDATE job SET state=CASE WHEN attempt_count>=max_attempts THEN 'failed' ELSE 'queued' END,
             lease_owner=NULL,lease_token=NULL,lease_expires_at_unix=NULL,heartbeat_at_unix=NULL
             WHERE state='running' AND lease_expires_at_unix<=?1",
            [now_unix],
        )?;
        tx.execute(
            "UPDATE goal SET status='failed' WHERE id IN
             (SELECT goal_id FROM job WHERE state='failed') AND status NOT IN ('failed','completed','cancelled')",
            [],
        )?;
        tx.execute(
            "INSERT INTO goal_event(goal_id,kind,code,created_at_utc)
             SELECT goal_id,'job.retry.exhausted','MAX_ATTEMPTS_REACHED',strftime('%Y-%m-%dT%H:%M:%fZ','now')
             FROM job WHERE state='failed' AND attempt_count>=max_attempts
               AND goal_id IN (SELECT id FROM goal WHERE status='failed')
               AND NOT EXISTS (SELECT 1 FROM goal_event WHERE goal_id=job.goal_id AND kind='job.retry.exhausted')",
            [],
        )?;
        let next: Option<(String, String)> = tx
            .query_row(
                "SELECT g.id,g.action FROM goal g JOIN job j ON j.goal_id=g.id
             WHERE j.state='queued' AND j.attempt_count<j.max_attempts AND g.score>0
               AND NOT EXISTS (SELECT 1 FROM goal_dependency gd JOIN goal prerequisite
                 ON prerequisite.id=gd.depends_on_goal_id WHERE gd.goal_id=g.id
                   AND prerequisite.status!='completed')
             ORDER BY g.score DESC,g.created_at_utc,g.id LIMIT 1",
                [],
                |row| Ok((row.get(0)?, row.get(1)?)),
            )
            .optional()?;
        let Some((goal_id, action)) = next else {
            tx.commit()?;
            return Ok(None);
        };
        let grant: String = tx.query_row(
            "SELECT grant FROM goal_grant WHERE goal_id=?1",
            [&goal_id],
            |row| row.get(0),
        )?;
        if let Err(error) = self.broker.verify(&goal_id, &action, &grant) {
            if error == luna_policy_broker::BrokerError::GrantExpired {
                tx.execute("UPDATE goal SET status='blocked' WHERE id=?1", [&goal_id])?;
                tx.execute(
                    "UPDATE job SET state='blocked' WHERE goal_id=?1",
                    [&goal_id],
                )?;
                tx.execute("INSERT INTO goal_event(goal_id,kind,code,created_at_utc) VALUES(?1,'goal.authorization.expired','GRANT_RENEWAL_REQUIRED',strftime('%Y-%m-%dT%H:%M:%fZ','now'))", [&goal_id])?;
                tx.commit()?;
                return Ok(None);
            }
            tx.commit()?;
            return Err(EngineError::Policy("EXECUTION_REAUTH_REQUIRED".into()));
        }
        if action != "local.checksum" {
            tx.commit()?;
            return Err(EngineError::Policy("ACTION_REVOKED".into()));
        }
        let token = Uuid::new_v4().to_string();
        let expires_at_unix = now_unix + lease_seconds;
        tx.execute(
            "UPDATE job SET state='running',lease_owner=?2,lease_token=?3,
             lease_expires_at_unix=?4,attempt_count=attempt_count+1,heartbeat_at_unix=?5 WHERE goal_id=?1 AND state='queued'",
            params![goal_id, worker, token, expires_at_unix, now_unix],
        )?;
        tx.execute("INSERT INTO goal_event(goal_id,kind,code,created_at_utc) VALUES(?1,'job.lease.claimed','AUTHORIZED_WORKER',strftime('%Y-%m-%dT%H:%M:%fZ','now'))", [&goal_id])?;
        tx.commit()?;
        Ok(Some(ExecutionLease {
            goal_id,
            action,
            grant,
            owner: worker.to_owned(),
            token,
            expires_at_unix,
        }))
    }

    pub fn renew_claim(
        &self,
        lease: &ExecutionLease,
        now_unix: i64,
        lease_seconds: i64,
    ) -> Result<ExecutionLease, EngineError> {
        if !(1..=300).contains(&lease_seconds) {
            return Err(EngineError::Invalid("INVALID_LEASE_REQUEST".into()));
        }
        let mut db = self.db.lock().expect("sqlite mutex poisoned");
        let tx = db.transaction_with_behavior(TransactionBehavior::Immediate)?;
        let active: Option<(String, String, String, i64)> = tx
            .query_row(
                "SELECT j.lease_owner,j.lease_token,gg.grant,j.lease_expires_at_unix
                 FROM job j JOIN goal_grant gg ON gg.goal_id=j.goal_id
                 WHERE j.goal_id=?1 AND j.state='running'",
                [&lease.goal_id],
                |row| Ok((row.get(0)?, row.get(1)?, row.get(2)?, row.get(3)?)),
            )
            .optional()?;
        let Some((owner, token, grant, expires_at)) = active else {
            return Err(EngineError::Invalid("LEASE_LOST".into()));
        };
        if owner != lease.owner
            || token != lease.token
            || grant != lease.grant
            || now_unix >= expires_at
        {
            return Err(EngineError::Invalid("LEASE_LOST".into()));
        }
        if self
            .broker
            .verify(&lease.goal_id, &lease.action, &lease.grant)
            .is_err()
        {
            tx.execute(
                "UPDATE job SET state='blocked',lease_owner=NULL,lease_token=NULL,
                 lease_expires_at_unix=NULL,heartbeat_at_unix=NULL WHERE goal_id=?1",
                [&lease.goal_id],
            )?;
            tx.execute(
                "UPDATE goal SET status='blocked' WHERE id=?1",
                [&lease.goal_id],
            )?;
            tx.execute(
                "INSERT INTO goal_event(goal_id,kind,code,created_at_utc)
                 VALUES(?1,'goal.authorization.revoked','LEASE_RENEWAL_DENIED',strftime('%Y-%m-%dT%H:%M:%fZ','now'))",
                [&lease.goal_id],
            )?;
            tx.commit()?;
            return Err(EngineError::Policy("EXECUTION_REAUTH_REQUIRED".into()));
        }
        let next_expiry = now_unix + lease_seconds;
        tx.execute(
            "UPDATE job SET lease_expires_at_unix=?2,heartbeat_at_unix=?3
             WHERE goal_id=?1 AND state='running' AND lease_owner=?4 AND lease_token=?5",
            params![
                lease.goal_id,
                next_expiry,
                now_unix,
                lease.owner,
                lease.token
            ],
        )?;
        tx.commit()?;
        let mut renewed = lease.clone();
        renewed.expires_at_unix = next_expiry;
        Ok(renewed)
    }

    pub fn complete_claim(
        &self,
        lease: &ExecutionLease,
        now_unix: i64,
    ) -> Result<TickResult, EngineError> {
        let mut db = self.db.lock().expect("sqlite mutex poisoned");
        let tx = db.transaction_with_behavior(TransactionBehavior::Immediate)?;
        let active: Option<(String, String, String, String, i64)> = tx
            .query_row(
                "SELECT j.lease_owner,j.lease_token,g.action,gg.grant,j.lease_expires_at_unix
             FROM job j JOIN goal g ON g.id=j.goal_id JOIN goal_grant gg ON gg.goal_id=g.id
             WHERE j.goal_id=?1 AND j.state='running'",
                [&lease.goal_id],
                |row| {
                    Ok((
                        row.get(0)?,
                        row.get(1)?,
                        row.get(2)?,
                        row.get(3)?,
                        row.get(4)?,
                    ))
                },
            )
            .optional()?;
        let Some((owner, token, action, persisted_grant, expires_at)) = active else {
            return Err(EngineError::Invalid("LEASE_LOST".into()));
        };
        if owner != lease.owner
            || token != lease.token
            || action != lease.action
            || persisted_grant != lease.grant
            || now_unix >= expires_at
        {
            return Err(EngineError::Invalid("LEASE_LOST".into()));
        }
        if self
            .broker
            .verify(&lease.goal_id, &lease.action, &lease.grant)
            .is_err()
        {
            tx.execute(
                "UPDATE goal SET status='blocked' WHERE id=?1",
                [&lease.goal_id],
            )?;
            tx.execute("UPDATE job SET state='blocked',lease_owner=NULL,lease_token=NULL,lease_expires_at_unix=NULL WHERE goal_id=?1", [&lease.goal_id])?;
            tx.execute("INSERT INTO goal_event(goal_id,kind,code,created_at_utc) VALUES(?1,'goal.authorization.revoked','EXECUTION_REAUTH_REQUIRED',strftime('%Y-%m-%dT%H:%M:%fZ','now'))", [&lease.goal_id])?;
            tx.commit()?;
            return Err(EngineError::Policy("EXECUTION_REAUTH_REQUIRED".into()));
        }
        const EXPECTED_SHA256: &str =
            "07be6a1cc7b1b323fd3baca70abf34bfea13e09f9d28cf417d11aca77d4220ce";
        let artifact = self.runner.execute(&lease.action).ok();
        let verified = artifact.as_ref().is_some_and(|evidence| {
            lease.action == "local.checksum"
                && evidence.sha256 == EXPECTED_SHA256
                && evidence.byte_count == 28
                && evidence.provenance == "local_public_fixture_v1"
        });
        if verified {
            let evidence = artifact.expect("verified artifact must be present");
            tx.execute(
                "INSERT INTO evidence VALUES(?1,?2,?3,'local_public_fixture_v1',1)",
                params![lease.goal_id, evidence.sha256, evidence.byte_count],
            )?;
            tx.execute(
                "UPDATE goal SET status='completed' WHERE id=?1",
                [&lease.goal_id],
            )?;
            tx.execute("UPDATE job SET state='succeeded',lease_owner=NULL,lease_token=NULL,lease_expires_at_unix=NULL WHERE goal_id=?1", [&lease.goal_id])?;
            tx.execute("INSERT INTO goal_event(goal_id,kind,code,created_at_utc) VALUES(?1,'goal.completed','ARTIFACT_VERIFIED',strftime('%Y-%m-%dT%H:%M:%fZ','now'))", [&lease.goal_id])?;
        } else {
            tx.execute(
                "UPDATE goal SET status='failed' WHERE id=?1",
                [&lease.goal_id],
            )?;
            tx.execute("UPDATE job SET state='failed',lease_owner=NULL,lease_token=NULL,lease_expires_at_unix=NULL WHERE goal_id=?1", [&lease.goal_id])?;
            tx.execute("INSERT INTO goal_event(goal_id,kind,code,created_at_utc) VALUES(?1,'goal.failed','ARTIFACT_MISMATCH',strftime('%Y-%m-%dT%H:%M:%fZ','now'))", [&lease.goal_id])?;
        }
        tx.commit()?;
        Ok(TickResult {
            outcome: if verified { "succeeded" } else { "failed" }.into(),
            goal_id: Some(lease.goal_id.clone()),
            reason: if verified {
                "FIXTURE_DIGEST_AND_LENGTH_VERIFIED"
            } else {
                "ARTIFACT_MISMATCH"
            }
            .into(),
        })
    }

    pub fn tick(&self) -> Result<TickResult, EngineError> {
        let now_unix = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .map_err(|error| EngineError::Invalid(error.to_string()))?
            .as_secs() as i64;
        if let Some(lease) = self.claim_next("luna-core", now_unix, 30)? {
            return self.complete_claim(&lease, now_unix);
        }
        let db = self.db.lock().expect("sqlite mutex poisoned");
        let queued: bool = db.query_row(
            "SELECT EXISTS(SELECT 1 FROM job WHERE state='queued')",
            [],
            |row| row.get(0),
        )?;
        let running: bool = db.query_row(
            "SELECT EXISTS(SELECT 1 FROM job WHERE state='running')",
            [],
            |row| row.get(0),
        )?;
        Ok(TickResult {
            outcome: if queued {
                "waiting_for_dependencies"
            } else if running {
                "work_in_progress"
            } else {
                "idle_success"
            }
            .into(),
            goal_id: None,
            reason: if queued {
                "DEPENDENCIES_NOT_COMPLETED"
            } else if running {
                "AUTHORIZED_JOB_LEASED"
            } else {
                "NO_AUTHORIZED_USEFUL_WORK"
            }
            .into(),
        })
    }

    pub fn evidence(&self, goal_id: &str) -> Result<Option<Evidence>, EngineError> {
        let db = self.db.lock().expect("sqlite mutex poisoned");
        Ok(db.query_row("SELECT goal_id,sha256,byte_count,provenance,verified FROM evidence WHERE goal_id=?1",[goal_id],|r| Ok(Evidence {goal_id:r.get(0)?,sha256:r.get(1)?,byte_count:r.get(2)?,provenance:r.get(3)?,verified:r.get::<_,i64>(4)? == 1})).optional()?)
    }

    pub fn get(&self, goal_id: &str) -> Result<GoalView, EngineError> {
        let db = self.db.lock().expect("sqlite mutex poisoned");
        read_goal(&db, goal_id)
    }

    pub fn list(&self) -> Result<Vec<GoalView>, EngineError> {
        let db = self.db.lock().expect("sqlite mutex poisoned");
        let ids = {
            let mut statement = db.prepare(
                "SELECT id FROM goal ORDER BY COALESCE(score,0) DESC,created_at_utc,id LIMIT 100",
            )?;
            let rows = statement.query_map([], |row| row.get::<_, String>(0))?;
            rows.collect::<Result<Vec<_>, _>>()?
        };
        ids.iter().map(|id| read_goal(&db, id)).collect()
    }

    pub fn events(&self, goal_id: &str, limit: u32) -> Result<Vec<GoalEventView>, EngineError> {
        if !(1..=100).contains(&limit) {
            return Err(EngineError::Invalid("EVENT_LIMIT_OUT_OF_RANGE".into()));
        }
        let db = self.db.lock().expect("sqlite mutex poisoned");
        let exists: bool = db.query_row(
            "SELECT EXISTS(SELECT 1 FROM goal WHERE id=?1)",
            [goal_id],
            |row| row.get(0),
        )?;
        if !exists {
            return Err(EngineError::NotFound);
        }
        let mut statement = db.prepare(
            "SELECT id,goal_id,kind,code,created_at_utc FROM (
               SELECT id,goal_id,kind,code,created_at_utc FROM goal_event
               WHERE goal_id=?1 ORDER BY id DESC LIMIT ?2
             ) ORDER BY id ASC",
        )?;
        let rows = statement.query_map(params![goal_id, limit], |row| {
            Ok(GoalEventView {
                event_id: row.get(0)?,
                goal_id: row.get(1)?,
                kind: row.get(2)?,
                decision_code: row.get(3)?,
                created_at_utc: row.get(4)?,
            })
        })?;
        Ok(rows.collect::<Result<Vec<_>, _>>()?)
    }

    /// Deliver only to the local durable inbox. This does not claim an external
    /// message bus or network delivery guarantee.
    pub fn dispatch_events(&self, limit: u32) -> Result<u32, EngineError> {
        if !(1..=100).contains(&limit) {
            return Err(EngineError::Invalid("EVENT_LIMIT_OUT_OF_RANGE".into()));
        }
        let mut db = self.db.lock().expect("sqlite mutex poisoned");
        let tx = db.transaction_with_behavior(TransactionBehavior::Immediate)?;
        let event_ids = {
            let mut statement = tx.prepare(
                "SELECT event_id FROM event_outbox WHERE delivered_at_utc IS NULL
                 ORDER BY event_id LIMIT ?1",
            )?;
            let rows = statement.query_map([limit], |row| row.get::<_, i64>(0))?;
            rows.collect::<Result<Vec<_>, _>>()?
        };
        for event_id in &event_ids {
            tx.execute(
                "INSERT OR IGNORE INTO event_inbox(event_id,received_at_utc)
                 VALUES(?1,strftime('%Y-%m-%dT%H:%M:%fZ','now'))",
                [event_id],
            )?;
            tx.execute(
                "UPDATE event_outbox SET attempts=attempts+1,
                 delivered_at_utc=strftime('%Y-%m-%dT%H:%M:%fZ','now')
                 WHERE event_id=?1 AND delivered_at_utc IS NULL",
                [event_id],
            )?;
        }
        tx.commit()?;
        Ok(event_ids.len() as u32)
    }

    pub fn event_delivery_counts(&self) -> Result<(u64, u64), EngineError> {
        let db = self.db.lock().expect("sqlite mutex poisoned");
        Ok((
            db.query_row(
                "SELECT count(*) FROM event_outbox WHERE delivered_at_utc IS NULL",
                [],
                |row| row.get(0),
            )?,
            db.query_row("SELECT count(*) FROM event_inbox", [], |row| row.get(0))?,
        ))
    }

    pub fn queued_count(&self) -> Result<u64, EngineError> {
        let db = self.db.lock().expect("sqlite mutex poisoned");
        Ok(
            db.query_row("SELECT count(*) FROM job WHERE state='queued'", [], |r| {
                r.get(0)
            })?,
        )
    }

    pub fn lease_attempts(&self, goal_id: &str) -> Result<u32, EngineError> {
        let db = self.db.lock().expect("sqlite mutex poisoned");
        Ok(db.query_row(
            "SELECT attempt_count FROM job WHERE goal_id=?1",
            [goal_id],
            |row| row.get(0),
        )?)
    }

    pub fn add_dependency(
        &self,
        goal_id: &str,
        depends_on_goal_id: &str,
    ) -> Result<(), EngineError> {
        if goal_id == depends_on_goal_id {
            return Err(EngineError::Invalid("DEPENDENCY_CYCLE".into()));
        }
        let mut db = self.db.lock().expect("sqlite mutex poisoned");
        let tx = db.transaction()?;
        let target_exists: bool = tx.query_row(
            "SELECT EXISTS(SELECT 1 FROM goal WHERE id=?1)",
            [goal_id],
            |row| row.get(0),
        )?;
        let prerequisite_exists: bool = tx.query_row(
            "SELECT EXISTS(SELECT 1 FROM goal WHERE id=?1)",
            [depends_on_goal_id],
            |row| row.get(0),
        )?;
        if !target_exists || !prerequisite_exists {
            return Err(EngineError::NotFound);
        }
        let creates_cycle: bool = tx.query_row(
            "WITH RECURSIVE reachable(id) AS (
               SELECT ?2
               UNION
               SELECT gd.depends_on_goal_id FROM goal_dependency gd
                 JOIN reachable r ON gd.goal_id=r.id
             )
             SELECT EXISTS(SELECT 1 FROM reachable WHERE id=?1)",
            params![goal_id, depends_on_goal_id],
            |row| row.get(0),
        )?;
        if creates_cycle {
            return Err(EngineError::Invalid("DEPENDENCY_CYCLE".into()));
        }
        tx.execute(
            "INSERT OR IGNORE INTO goal_dependency(goal_id,depends_on_goal_id,created_at_utc) VALUES(?1,?2,strftime('%Y-%m-%dT%H:%M:%fZ','now'))",
            params![goal_id, depends_on_goal_id],
        )?;
        tx.commit()?;
        Ok(())
    }

    /// Atomically replace the hierarchy and prerequisites for a goal. Plans are
    /// immutable while work is leased, terminal, or waiting for renewed policy.
    pub fn set_goal_plan(
        &self,
        goal_id: &str,
        plan: &GoalPlan,
    ) -> Result<GoalPlanView, EngineError> {
        const MAX_DEPENDENCIES: usize = 32;
        const MAX_PARENT_DEPTH: usize = 16;
        if plan.dependency_goal_ids.len() > MAX_DEPENDENCIES {
            return Err(EngineError::Invalid("PLAN_TOO_LARGE".into()));
        }
        let mut unique = std::collections::HashSet::new();
        if plan.dependency_goal_ids.iter().any(|id| !unique.insert(id)) {
            return Err(EngineError::Invalid("DUPLICATE_DEPENDENCY".into()));
        }
        if plan.parent_goal_id.as_deref() == Some(goal_id)
            || plan.dependency_goal_ids.iter().any(|id| id == goal_id)
        {
            return Err(EngineError::Invalid("DEPENDENCY_CYCLE".into()));
        }

        let mut db = self.db.lock().expect("sqlite mutex poisoned");
        let tx = db.transaction_with_behavior(TransactionBehavior::Immediate)?;
        let target: Option<(String, String)> = tx
            .query_row(
                "SELECT g.status,j.state FROM goal g JOIN job j ON j.goal_id=g.id WHERE g.id=?1",
                [goal_id],
                |row| Ok((row.get(0)?, row.get(1)?)),
            )
            .optional()?;
        let Some((status, job_state)) = target else {
            return Err(EngineError::NotFound);
        };
        if !["approved", "paused"].contains(&status.as_str())
            || !["queued", "paused", "blocked"].contains(&job_state.as_str())
        {
            return Err(EngineError::Invalid("PLAN_LOCKED_BY_GOAL_STATE".into()));
        }
        for prerequisite in plan
            .dependency_goal_ids
            .iter()
            .chain(plan.parent_goal_id.iter())
        {
            let prerequisite_status: Option<String> = tx
                .query_row(
                    "SELECT status FROM goal WHERE id=?1",
                    [prerequisite],
                    |row| row.get(0),
                )
                .optional()?;
            let Some(prerequisite_status) = prerequisite_status else {
                return Err(EngineError::NotFound);
            };
            if ["blocked", "cancelled", "failed"].contains(&prerequisite_status.as_str()) {
                return Err(EngineError::Invalid("DEPENDENCY_NOT_RUNNABLE".into()));
            }
        }

        // Check hierarchy depth and repeats before changing persisted state.
        let mut ancestors = std::collections::HashSet::from([goal_id.to_owned()]);
        let mut cursor = plan.parent_goal_id.clone();
        let mut depth = 0;
        while let Some(parent) = cursor {
            depth += 1;
            if depth > MAX_PARENT_DEPTH {
                return Err(EngineError::Invalid("PLAN_TOO_LARGE".into()));
            }
            if !ancestors.insert(parent.clone()) {
                return Err(EngineError::Invalid("DEPENDENCY_CYCLE".into()));
            }
            cursor = tx
                .query_row(
                    "SELECT parent_goal_id FROM goal_parent WHERE goal_id=?1",
                    [&parent],
                    |row| row.get(0),
                )
                .optional()?;
        }

        tx.execute("DELETE FROM goal_dependency WHERE goal_id=?1", [goal_id])?;
        tx.execute("DELETE FROM goal_parent WHERE goal_id=?1", [goal_id])?;
        for candidate in plan
            .dependency_goal_ids
            .iter()
            .chain(plan.parent_goal_id.iter())
        {
            let creates_cycle: bool = tx.query_row(
                "WITH RECURSIVE edges(source,target) AS (
                   SELECT goal_id,depends_on_goal_id FROM goal_dependency
                   UNION SELECT goal_id,parent_goal_id FROM goal_parent
                 ), reachable(id) AS (
                   SELECT ?1 UNION SELECT edges.target FROM edges JOIN reachable ON edges.source=reachable.id
                 ) SELECT EXISTS(SELECT 1 FROM reachable WHERE id=?2)",
                params![candidate, goal_id],
                |row| row.get(0),
            )?;
            if creates_cycle {
                return Err(EngineError::Invalid("DEPENDENCY_CYCLE".into()));
            }
        }
        for prerequisite in &plan.dependency_goal_ids {
            tx.execute(
                "INSERT INTO goal_dependency(goal_id,depends_on_goal_id,created_at_utc) VALUES(?1,?2,strftime('%Y-%m-%dT%H:%M:%fZ','now'))",
                params![goal_id, prerequisite],
            )?;
        }
        if let Some(parent) = &plan.parent_goal_id {
            tx.execute(
                "INSERT INTO goal_parent(goal_id,parent_goal_id,created_at_utc) VALUES(?1,?2,strftime('%Y-%m-%dT%H:%M:%fZ','now'))",
                params![goal_id, parent],
            )?;
        }
        tx.execute(
            "INSERT INTO goal_event(goal_id,kind,code,created_at_utc) VALUES(?1,'goal.plan.replaced','OWNER_PLAN_UPDATED',strftime('%Y-%m-%dT%H:%M:%fZ','now'))",
            [goal_id],
        )?;
        let view = GoalPlanView {
            goal_id: goal_id.to_owned(),
            parent_goal_id: plan.parent_goal_id.clone(),
            dependency_goal_ids: plan.dependency_goal_ids.clone(),
        };
        tx.commit()?;
        Ok(view)
    }

    pub fn get_goal_plan(&self, goal_id: &str) -> Result<GoalPlanView, EngineError> {
        let db = self.db.lock().expect("sqlite mutex poisoned");
        let exists: bool = db.query_row(
            "SELECT EXISTS(SELECT 1 FROM goal WHERE id=?1)",
            [goal_id],
            |row| row.get(0),
        )?;
        if !exists {
            return Err(EngineError::NotFound);
        }
        let parent_goal_id = db
            .query_row(
                "SELECT parent_goal_id FROM goal_parent WHERE goal_id=?1",
                [goal_id],
                |row| row.get(0),
            )
            .optional()?;
        let mut statement = db.prepare(
            "SELECT depends_on_goal_id FROM goal_dependency WHERE goal_id=?1 ORDER BY depends_on_goal_id",
        )?;
        let rows = statement.query_map([goal_id], |row| row.get(0))?;
        let dependency_goal_ids = rows.collect::<Result<Vec<String>, _>>()?;
        Ok(GoalPlanView {
            goal_id: goal_id.to_owned(),
            parent_goal_id,
            dependency_goal_ids,
        })
    }

    pub fn transition(
        &self,
        goal_id: &str,
        command: GoalLifecycleCommand,
    ) -> Result<GoalView, EngineError> {
        let mut db = self.db.lock().expect("sqlite mutex poisoned");
        let tx = db.transaction_with_behavior(TransactionBehavior::Immediate)?;
        let status = tx
            .query_row("SELECT status FROM goal WHERE id=?1", [goal_id], |row| {
                row.get::<_, String>(0)
            })
            .optional()?
            .ok_or(EngineError::NotFound)?;
        let (next_goal, next_job, event, code) = match command {
            GoalLifecycleCommand::Pause
                if ["approved", "active", "blocked"].contains(&status.as_str()) =>
            {
                ("paused", "paused", "goal.paused", "OWNER_PAUSED")
            }
            GoalLifecycleCommand::Resume if status == "paused" => {
                ("active", "queued", "goal.resumed", "OWNER_RESUMED")
            }
            GoalLifecycleCommand::Cancel
                if !["completed", "cancelled"].contains(&status.as_str()) =>
            {
                (
                    "cancelled",
                    "cancelled",
                    "goal.cancelled",
                    "OWNER_CANCELLED",
                )
            }
            _ => return Err(EngineError::Invalid("INVALID_GOAL_TRANSITION".into())),
        };
        tx.execute(
            "UPDATE goal SET status=?2 WHERE id=?1",
            params![goal_id, next_goal],
        )?;
        tx.execute(
            "UPDATE job SET state=?2,lease_owner=NULL,lease_token=NULL,lease_expires_at_unix=NULL WHERE goal_id=?1",
            params![goal_id, next_job],
        )?;
        tx.execute(
            "INSERT INTO goal_event(goal_id,kind,code,created_at_utc) VALUES(?1,?2,?3,strftime('%Y-%m-%dT%H:%M:%fZ','now'))",
            params![goal_id, event, code],
        )?;
        let goal = read_goal(&tx, goal_id)?;
        tx.commit()?;
        Ok(goal)
    }
}

fn read_goal(db: &Connection, id: &str) -> Result<GoalView, EngineError> {
    db.query_row("SELECT id,action,status,score,policy_code,provenance_status,(SELECT state IN ('queued','running') FROM job WHERE goal_id=goal.id) FROM goal WHERE id=?1",[id],|r| Ok(GoalView {contract_version:CONTRACT_VERSION,id:r.get(0)?,action:r.get(1)?,status:r.get(2)?,score:r.get(3)?,policy_code:r.get(4)?,provenance_status:match r.get::<_,String>(5)?.as_str(){"verified"=>SourceStatus::Verified,"disputed"=>SourceStatus::Disputed,_=>SourceStatus::Unknown},queued:r.get::<_,bool>(6)?})).map_err(|e| if e == rusqlite::Error::QueryReturnedNoRows { EngineError::NotFound } else { EngineError::Storage(e) })
}

fn validate(body: &CreateGoal, idempotency_key: &str) -> Result<(), EngineError> {
    if idempotency_key.is_empty()
        || idempotency_key.len() > 128
        || !idempotency_key
            .chars()
            .all(|c| c.is_ascii_alphanumeric() || "_.:-".contains(c))
    {
        return Err(EngineError::Invalid("IDEMPOTENCY_KEY_REQUIRED".into()));
    }
    for n in [
        body.metrics.impact,
        body.metrics.urgency,
        body.metrics.user_value,
        body.metrics.confidence,
        body.metrics.efficiency,
    ] {
        if !n.is_finite() || !(0.0..=1.0).contains(&n) {
            return Err(EngineError::Invalid("METRICS_MUST_BE_FINITE_0_TO_1".into()));
        }
    }
    if body.action.is_empty()
        || body.action.len() > 80
        || !body
            .action
            .chars()
            .all(|c| c.is_ascii_lowercase() || c.is_ascii_digit() || "_.".contains(c))
    {
        return Err(EngineError::Invalid("INVALID_ACTION".into()));
    }
    if body
        .provenance
        .claim
        .as_ref()
        .is_some_and(|v| v.len() > 512)
        || body
            .provenance
            .source_id
            .as_ref()
            .is_some_and(|v| v.len() > 256)
    {
        return Err(EngineError::Invalid("INVALID_PROVENANCE_INPUT".into()));
    }
    Ok(())
}

pub fn score(metrics: &luna_contracts::Metrics) -> f64 {
    (100.0
        * (0.30 * metrics.impact
            + 0.20 * metrics.urgency
            + 0.20 * metrics.user_value
            + 0.15 * metrics.confidence
            + 0.15 * metrics.efficiency)
        * 1_000_000.0)
        .round()
        / 1_000_000.0
}

#[cfg(test)]
mod tests {
    use super::*;
    use luna_contracts::{CreateGoal, Metrics, PrivacyClass, ProvenanceInput, SourceStatus};
    use luna_policy_broker::PolicyBroker;
    fn body(action: &str) -> CreateGoal {
        CreateGoal {
            action: action.into(),
            metrics: Metrics::default(),
            privacy_class: PrivacyClass::General,
            expected_cost_minor: None,
            provenance: ProvenanceInput {
                source_id: None,
                claim: None,
            },
        }
    }
    fn engine() -> (GoalEngine, String) {
        let path = std::env::temp_dir().join(format!("luna-test-{}.sqlite", Uuid::new_v4()));
        (
            GoalEngine::open(path, PolicyBroker::new("owner-test-key", None)).unwrap(),
            "owner-test-key".into(),
        )
    }

    #[test]
    fn deny_before_score_and_no_job_for_denied_action() {
        let (engine, key) = engine();
        assert!(engine.create(&body("system.exec"), &key, "deny-1").is_err());
        assert_eq!(engine.queued_count().unwrap(), 0);
    }
    #[test]
    fn duplicate_job_is_idempotent() {
        let (engine, key) = engine();
        let a = engine
            .create(&body("local.checksum"), &key, "same")
            .unwrap();
        let b = engine
            .create(&body("local.checksum"), &key, "same")
            .unwrap();
        assert_eq!(a.id, b.id);
        assert_eq!(engine.queued_count().unwrap(), 1);
    }
    #[test]
    fn unknown_source_remains_unknown() {
        let (engine, key) = engine();
        let mut request = body("local.checksum");
        request.provenance.source_id = Some("untrusted-article".into());
        request.provenance.claim = Some("The claim has not been checked".into());
        let g = engine.create(&request, &key, "prov").unwrap();
        assert_eq!(g.provenance_status, SourceStatus::Unknown);
    }
    #[test]
    fn idle_is_success_and_action_records_verified_evidence() {
        let (engine, key) = engine();
        assert_eq!(engine.tick().unwrap().outcome, "idle_success");
        let goal = engine.create(&body("local.checksum"), &key, "run").unwrap();
        assert_eq!(engine.tick().unwrap().outcome, "succeeded");
        assert!(engine.evidence(&goal.id).unwrap().unwrap().verified);
        assert_eq!(engine.tick().unwrap().outcome, "idle_success");
    }

    #[test]
    fn active_lease_prevents_idle_success_and_duplicate_claims() {
        let (engine, key) = engine();
        engine
            .create(&body("local.checksum"), &key, "leased-once")
            .unwrap();
        let now = 1_800_000_000;
        let lease = engine.claim_next("worker-a", now, 30).unwrap().unwrap();
        assert!(engine.claim_next("worker-b", now, 30).unwrap().is_none());
        let tick = engine.tick().unwrap();
        assert_eq!(tick.outcome, "work_in_progress");
        assert_eq!(tick.reason, "AUTHORIZED_JOB_LEASED");
        assert_eq!(lease.owner, "worker-a");
    }

    #[test]
    fn expired_lease_is_reclaimed_after_engine_restart() {
        let path =
            std::env::temp_dir().join(format!("luna-lease-restart-{}.sqlite", Uuid::new_v4()));
        let key = "owner-test-key";
        let (goal_id, first_token) = {
            let engine = GoalEngine::open(&path, PolicyBroker::new(key, None)).unwrap();
            let goal = engine
                .create(&body("local.checksum"), key, "lease-restart")
                .unwrap();
            let claim = engine
                .claim_next("worker-a", 1_800_000_000, 30)
                .unwrap()
                .unwrap();
            (goal.id, claim.token)
        };
        let engine = GoalEngine::open(&path, PolicyBroker::new(key, None)).unwrap();
        let recovered = engine
            .claim_next("worker-b", 1_800_000_031, 30)
            .unwrap()
            .unwrap();
        assert_eq!(recovered.goal_id, goal_id);
        assert_ne!(recovered.token, first_token);
        assert_eq!(engine.lease_attempts(&goal_id).unwrap(), 2);
        assert!(
            matches!(engine.complete_claim(&recovered, 1_800_000_032), Ok(result) if result.outcome == "succeeded")
        );
        assert!(
            matches!(engine.complete_claim(&recovered, 1_800_000_033), Err(EngineError::Invalid(code)) if code == "LEASE_LOST")
        );
    }

    #[test]
    fn lease_renewal_requires_the_live_owner_token_and_extends_expiry() {
        let (engine, key) = engine();
        let goal = engine
            .create(&body("local.checksum"), &key, "lease-renewal")
            .unwrap();
        let now = 1_800_000_000;
        let lease = engine.claim_next("worker-a", now, 30).unwrap().unwrap();
        let renewed = engine.renew_claim(&lease, now + 10, 60).unwrap();
        assert_eq!(renewed.expires_at_unix, now + 70);

        let mut stale = renewed.clone();
        stale.token = "forged-stale-token".into();
        assert!(
            matches!(engine.renew_claim(&stale, now + 11, 60), Err(EngineError::Invalid(code)) if code == "LEASE_LOST")
        );
        assert_eq!(engine.get(&goal.id).unwrap().status, "approved");
    }

    #[test]
    fn expired_lease_stops_after_the_bounded_retry_limit() {
        let (engine, key) = engine();
        let goal = engine
            .create(&body("local.checksum"), &key, "lease-retry-limit")
            .unwrap();
        let first = engine
            .claim_next("worker-a", 1_800_000_000, 10)
            .unwrap()
            .unwrap();
        let second = engine
            .claim_next("worker-b", 1_800_000_010, 10)
            .unwrap()
            .unwrap();
        let third = engine
            .claim_next("worker-c", 1_800_000_020, 10)
            .unwrap()
            .unwrap();
        assert_ne!(first.token, second.token);
        assert_ne!(second.token, third.token);
        assert!(
            engine
                .claim_next("worker-d", 1_800_000_030, 10)
                .unwrap()
                .is_none()
        );
        assert_eq!(engine.lease_attempts(&goal.id).unwrap(), 3);
        assert_eq!(engine.get(&goal.id).unwrap().status, "failed");
    }

    #[test]
    fn version_one_database_migrates_queued_work_without_loss() {
        let path =
            std::env::temp_dir().join(format!("luna-v1-migration-{}.sqlite", Uuid::new_v4()));
        let key = "owner-test-key";
        let goal_id = Uuid::new_v4().to_string();
        let broker = PolicyBroker::new(key, None);
        let grant = broker
            .authorize(
                &PolicyRequest {
                    owner_token: key.into(),
                    principal: Principal::Owner,
                    action: "local.checksum".into(),
                    privacy_class: PrivacyClass::General,
                    expected_cost_minor: None,
                    configured_budget_minor: None,
                },
                &goal_id,
            )
            .unwrap();
        {
            let db = Connection::open(&path).unwrap();
            db.execute_batch(include_str!("../../../../migrations/native/001_core.sql"))
                .unwrap();
            db.execute(
                "INSERT INTO goal VALUES(?1,'local.checksum','approved',100,'OWNER_SCOPED_GRANT','unknown','legacy-key','legacy-hash',strftime('%Y-%m-%dT%H:%M:%fZ','now'))",
                [&goal_id],
            ).unwrap();
            db.execute(
                "INSERT INTO goal_grant VALUES(?1,?2)",
                params![goal_id, grant],
            )
            .unwrap();
            db.execute("INSERT INTO job VALUES(?1,'queued')", [&goal_id])
                .unwrap();
        }
        let engine = GoalEngine::open(&path, broker).unwrap();
        assert_eq!(engine.get(&goal_id).unwrap().id, goal_id);
        assert_eq!(engine.lease_attempts(&goal_id).unwrap(), 0);
        let db = engine.db.lock().unwrap();
        let broken_reference: Option<String> = db
            .query_row(
                "SELECT \"table\" FROM pragma_foreign_key_check LIMIT 1",
                [],
                |row| row.get(0),
            )
            .optional()
            .unwrap();
        assert!(broken_reference.is_none());
        drop(db);
        assert_eq!(engine.tick().unwrap().outcome, "succeeded");
        assert_eq!(engine.lease_attempts(&goal_id).unwrap(), 1);
    }

    #[test]
    fn pause_prevents_claim_and_resume_requeues_goal() {
        let (engine, key) = engine();
        let goal = engine
            .create(&body("local.checksum"), &key, "pause-resume")
            .unwrap();
        let paused = engine
            .transition(&goal.id, GoalLifecycleCommand::Pause)
            .unwrap();
        assert_eq!(paused.status, "paused");
        assert!(!paused.queued);
        assert!(
            engine
                .claim_next("worker", 1_800_000_000, 30)
                .unwrap()
                .is_none()
        );
        let resumed = engine
            .transition(&goal.id, GoalLifecycleCommand::Resume)
            .unwrap();
        assert_eq!(resumed.status, "active");
        assert!(resumed.queued);
        assert_eq!(engine.tick().unwrap().outcome, "succeeded");
    }

    #[test]
    fn pause_revokes_an_active_lease_and_terminal_goals_cannot_change() {
        let (engine, key) = engine();
        let goal = engine
            .create(&body("local.checksum"), &key, "pause-lease")
            .unwrap();
        let claim = engine
            .claim_next("worker", 1_800_000_000, 30)
            .unwrap()
            .unwrap();
        assert_eq!(
            engine
                .transition(&goal.id, GoalLifecycleCommand::Pause)
                .unwrap()
                .status,
            "paused"
        );
        assert!(
            matches!(engine.complete_claim(&claim, 1_800_000_001), Err(EngineError::Invalid(code)) if code == "LEASE_LOST")
        );
        assert_eq!(
            engine
                .transition(&goal.id, GoalLifecycleCommand::Cancel)
                .unwrap()
                .status,
            "cancelled"
        );
        assert!(
            matches!(engine.transition(&goal.id, GoalLifecycleCommand::Resume), Err(EngineError::Invalid(code)) if code == "INVALID_GOAL_TRANSITION")
        );
        let complete = engine
            .create(&body("local.checksum"), &key, "terminal-complete")
            .unwrap();
        engine.tick().unwrap();
        assert_eq!(engine.get(&complete.id).unwrap().status, "completed");
        assert!(
            matches!(engine.transition(&complete.id, GoalLifecycleCommand::Pause), Err(EngineError::Invalid(code)) if code == "INVALID_GOAL_TRANSITION")
        );
    }

    #[test]
    fn cancellation_is_durable_across_engine_restart() {
        let path =
            std::env::temp_dir().join(format!("luna-cancel-restart-{}.sqlite", Uuid::new_v4()));
        let key = "owner-test-key";
        let goal_id = {
            let engine = GoalEngine::open(&path, PolicyBroker::new(key, None)).unwrap();
            let goal = engine
                .create(&body("local.checksum"), key, "cancel-restart")
                .unwrap();
            engine
                .transition(&goal.id, GoalLifecycleCommand::Cancel)
                .unwrap();
            goal.id
        };
        let engine = GoalEngine::open(&path, PolicyBroker::new(key, None)).unwrap();
        assert_eq!(engine.get(&goal_id).unwrap().status, "cancelled");
        assert_eq!(engine.tick().unwrap().outcome, "idle_success");
    }

    #[test]
    fn goal_listing_is_bounded_and_ordered_by_score_then_creation() {
        let (engine, key) = engine();
        let low = CreateGoal {
            metrics: Metrics {
                impact: 0.2,
                urgency: 0.2,
                user_value: 0.2,
                confidence: 0.2,
                efficiency: 0.2,
            },
            ..body("local.checksum")
        };
        let low_goal = engine.create(&low, &key, "list-low").unwrap();
        let high_goal = engine
            .create(&body("local.checksum"), &key, "list-high")
            .unwrap();
        let goals = engine.list().unwrap();
        assert_eq!(goals.len(), 2);
        assert_eq!(goals[0].id, high_goal.id);
        assert_eq!(goals[1].id, low_goal.id);
    }

    #[test]
    fn goal_events_are_private_listable_and_delivered_idempotently() {
        let (engine, key) = engine();
        let goal = engine
            .create(&body("local.checksum"), &key, "event-outbox")
            .unwrap();
        let events = engine.events(&goal.id, 100).unwrap();
        assert_eq!(events.len(), 1);
        assert_eq!(events[0].kind, "goal.created");
        assert_eq!(engine.event_delivery_counts().unwrap(), (1, 0));
        assert_eq!(engine.dispatch_events(100).unwrap(), 1);
        assert_eq!(engine.event_delivery_counts().unwrap(), (0, 1));
        assert_eq!(engine.dispatch_events(100).unwrap(), 0);
        assert_eq!(engine.event_delivery_counts().unwrap(), (0, 1));
        let db = engine.db.lock().unwrap();
        assert!(
            db.execute(
                "UPDATE goal_event SET code='tampered' WHERE id=?1",
                [events[0].event_id]
            )
            .is_err()
        );
        assert!(
            db.execute("DELETE FROM goal_event WHERE id=?1", [events[0].event_id])
                .is_err()
        );
    }

    #[test]
    fn goal_plan_replacement_is_atomic_and_detects_mixed_graph_cycles() {
        let (engine, key) = engine();
        let first = engine
            .create(&body("local.checksum"), &key, "plan-first")
            .unwrap();
        let second = engine
            .create(&body("local.checksum"), &key, "plan-second")
            .unwrap();
        let third = engine
            .create(&body("local.checksum"), &key, "plan-third")
            .unwrap();
        engine
            .set_goal_plan(
                &first.id,
                &GoalPlan {
                    parent_goal_id: Some(second.id.clone()),
                    dependency_goal_ids: vec![second.id.clone()],
                },
            )
            .unwrap();
        let replaced = engine
            .set_goal_plan(
                &first.id,
                &GoalPlan {
                    parent_goal_id: None,
                    dependency_goal_ids: vec![third.id.clone()],
                },
            )
            .unwrap();
        assert_eq!(replaced.dependency_goal_ids, vec![third.id.clone()]);
        assert!(matches!(engine.set_goal_plan(&third.id, &GoalPlan {
            parent_goal_id: None,
            dependency_goal_ids: vec![first.id.clone()],
        }), Err(EngineError::Invalid(code)) if code == "DEPENDENCY_CYCLE"));
        assert_eq!(engine.get_goal_plan(&first.id).unwrap(), replaced);

        engine
            .set_goal_plan(
                &first.id,
                &GoalPlan {
                    parent_goal_id: Some(second.id.clone()),
                    dependency_goal_ids: vec![],
                },
            )
            .unwrap();
        assert!(matches!(engine.set_goal_plan(&second.id, &GoalPlan {
            parent_goal_id: None,
            dependency_goal_ids: vec![first.id.clone()],
        }), Err(EngineError::Invalid(code)) if code == "DEPENDENCY_CYCLE"));
    }

    #[test]
    fn goal_plan_rejects_oversized_or_unrunnable_dependencies() {
        let (engine, key) = engine();
        let target = engine
            .create(&body("local.checksum"), &key, "plan-limit-target")
            .unwrap();
        let dependencies = (0..33)
            .map(|index| {
                engine
                    .create(
                        &body("local.checksum"),
                        &key,
                        &format!("plan-limit-{index}"),
                    )
                    .unwrap()
                    .id
            })
            .collect::<Vec<_>>();
        assert!(matches!(engine.set_goal_plan(&target.id, &GoalPlan {
            parent_goal_id: None,
            dependency_goal_ids: dependencies,
        }), Err(EngineError::Invalid(code)) if code == "PLAN_TOO_LARGE"));
        let blocked = engine
            .create(&body("local.checksum"), &key, "plan-blocked-dependency")
            .unwrap();
        {
            let db = engine.db.lock().unwrap();
            db.execute(
                "UPDATE goal SET status='blocked' WHERE id=?1",
                [&blocked.id],
            )
            .unwrap();
            db.execute(
                "UPDATE job SET state='blocked' WHERE goal_id=?1",
                [&blocked.id],
            )
            .unwrap();
        }
        assert!(matches!(engine.set_goal_plan(&target.id, &GoalPlan {
            parent_goal_id: None,
            dependency_goal_ids: vec![blocked.id],
        }), Err(EngineError::Invalid(code)) if code == "DEPENDENCY_NOT_RUNNABLE"));
    }
    #[test]
    fn queued_goal_resumes_after_engine_reopens() {
        let path = std::env::temp_dir().join(format!("luna-restart-{}.sqlite", Uuid::new_v4()));
        let key = "owner-test-key";
        let original = {
            let engine = GoalEngine::open(&path, PolicyBroker::new(key, None)).unwrap();
            engine
                .create(&body("local.checksum"), key, "resume")
                .unwrap()
                .id
        };
        let engine = GoalEngine::open(&path, PolicyBroker::new(key, None)).unwrap();
        assert_eq!(
            engine
                .create(&body("local.checksum"), key, "resume")
                .unwrap()
                .id,
            original
        );
        assert_eq!(engine.tick().unwrap().outcome, "succeeded");
    }

    #[test]
    fn dependency_prevents_a_higher_scored_goal_from_running_early() {
        let (engine, key) = engine();
        let mut low_priority = body("local.checksum");
        low_priority.metrics = Metrics {
            impact: 0.1,
            urgency: 0.1,
            user_value: 0.1,
            confidence: 0.1,
            efficiency: 0.1,
        };
        let prerequisite = engine
            .create(&low_priority, &key, "dependency-prerequisite")
            .unwrap();
        let blocker = engine
            .create(&body("local.checksum"), &key, "dependency-blocker")
            .unwrap();
        engine
            .add_dependency(&blocker.id, &prerequisite.id)
            .unwrap();
        let first = engine.tick().unwrap();
        assert_eq!(first.outcome, "succeeded");
        assert_eq!(first.goal_id.as_deref(), Some(prerequisite.id.as_str()));
        assert_eq!(engine.get(&blocker.id).unwrap().status, "approved");
        assert_eq!(engine.tick().unwrap().outcome, "succeeded");
    }

    #[test]
    fn dependency_cycles_are_rejected() {
        let (engine, key) = engine();
        let first = engine
            .create(&body("local.checksum"), &key, "cycle-first")
            .unwrap();
        let second = engine
            .create(&body("local.checksum"), &key, "cycle-second")
            .unwrap();
        engine.add_dependency(&first.id, &second.id).unwrap();
        assert!(matches!(
            engine.add_dependency(&second.id, &first.id),
            Err(EngineError::Invalid(code)) if code == "DEPENDENCY_CYCLE"
        ));
    }

    #[test]
    fn queued_goal_waiting_on_blocked_dependency_is_not_reported_idle() {
        let (engine, key) = engine();
        let prerequisite = engine
            .create(&body("local.checksum"), &key, "waiting-prerequisite")
            .unwrap();
        let dependent = engine
            .create(&body("local.checksum"), &key, "waiting-dependent")
            .unwrap();
        engine
            .add_dependency(&dependent.id, &prerequisite.id)
            .unwrap();
        {
            let db = engine.db.lock().unwrap();
            db.execute(
                "UPDATE goal SET status='blocked' WHERE id=?1",
                [&prerequisite.id],
            )
            .unwrap();
            db.execute(
                "UPDATE job SET state='blocked' WHERE goal_id=?1",
                [&prerequisite.id],
            )
            .unwrap();
        }
        let tick = engine.tick().unwrap();
        assert_eq!(tick.outcome, "waiting_for_dependencies");
        assert_eq!(tick.reason, "DEPENDENCIES_NOT_COMPLETED");
        assert!(engine.get(&dependent.id).unwrap().queued);
    }
}
