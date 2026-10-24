use luna_contracts::{
    CONTRACT_VERSION, CreateGoal, Evidence, GoalView, PolicyRequest, Principal, SourceStatus,
    TickResult,
};
use luna_policy_broker::PolicyService;
use rusqlite::{Connection, OptionalExtension, params};
use sha2::{Digest, Sha256};
use std::{
    fs::OpenOptions,
    path::Path,
    sync::{Arc, Mutex},
};
use thiserror::Error;
use uuid::Uuid;

const FIXTURE: &[u8] = b"LunaOS public checkpoint v1\n";

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
        Ok(Self {
            db: Arc::new(Mutex::new(db)),
            broker: Arc::new(broker),
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
        tx.execute("INSERT INTO job VALUES(?1,'queued')", [&id])?;
        tx.execute("INSERT INTO goal_event(goal_id,kind,code,created_at_utc) VALUES(?1,'goal.created','OWNER_REQUEST',strftime('%Y-%m-%dT%H:%M:%fZ','now'))", [&id])?;
        tx.commit()?;
        read_goal(&db, &id)
    }

    pub fn tick(&self) -> Result<TickResult, EngineError> {
        let mut db = self.db.lock().expect("sqlite mutex poisoned");
        let tx = db.transaction()?;
        let next: Option<(String,String,String)> = tx.query_row("SELECT g.id,g.action,g.policy_code FROM goal g JOIN job j ON j.goal_id=g.id WHERE j.state='queued' AND g.score>0 AND NOT EXISTS (SELECT 1 FROM goal_dependency gd JOIN goal prerequisite ON prerequisite.id=gd.depends_on_goal_id WHERE gd.goal_id=g.id AND prerequisite.status!='completed') ORDER BY g.score DESC,g.created_at_utc,g.id LIMIT 1", [], |r| Ok((r.get(0)?,r.get(1)?,r.get(2)?))).optional()?;
        let Some((id, action, _)) = next else {
            let queued: bool = tx.query_row(
                "SELECT EXISTS(SELECT 1 FROM job WHERE state='queued')",
                [],
                |r| r.get(0),
            )?;
            tx.commit()?;
            return Ok(TickResult {
                outcome: if queued {
                    "waiting_for_dependencies"
                } else {
                    "idle_success"
                }
                .into(),
                goal_id: None,
                reason: if queued {
                    "DEPENDENCIES_NOT_COMPLETED"
                } else {
                    "NO_AUTHORIZED_USEFUL_WORK"
                }
                .into(),
            });
        };
        let grant: String = tx.query_row(
            "SELECT grant FROM goal_grant WHERE goal_id=?1",
            [&id],
            |r| r.get(0),
        )?;
        if let Err(error) = self.broker.verify(&id, &action, &grant) {
            if error == luna_policy_broker::BrokerError::GrantExpired {
                tx.execute("UPDATE goal SET status='blocked' WHERE id=?1", [&id])?;
                tx.execute("UPDATE job SET state='blocked' WHERE goal_id=?1", [&id])?;
                tx.execute("INSERT INTO goal_event(goal_id,kind,code,created_at_utc) VALUES(?1,'goal.authorization.expired','GRANT_RENEWAL_REQUIRED',strftime('%Y-%m-%dT%H:%M:%fZ','now'))", [&id])?;
                tx.commit()?;
                return Ok(TickResult {
                    outcome: "blocked".into(),
                    goal_id: Some(id),
                    reason: "GRANT_RENEWAL_REQUIRED".into(),
                });
            }
            tx.commit()?;
            return Err(EngineError::Policy("EXECUTION_REAUTH_REQUIRED".into()));
        }
        if action != "local.checksum" {
            return Err(EngineError::Policy("ACTION_REVOKED".into()));
        }
        let digest = hex::encode(Sha256::digest(FIXTURE));
        const EXPECTED_SHA256: &str =
            "07be6a1cc7b1b323fd3baca70abf34bfea13e09f9d28cf417d11aca77d4220ce";
        let verified = digest == EXPECTED_SHA256 && FIXTURE.len() == 28;
        if verified {
            tx.execute(
                "INSERT INTO evidence VALUES(?1,?2,?3,'local_public_fixture_v1',1)",
                params![id, digest, FIXTURE.len()],
            )?;
            tx.execute("UPDATE goal SET status='completed' WHERE id=?1", [&id])?;
            tx.execute("UPDATE job SET state='succeeded' WHERE goal_id=?1", [&id])?;
            tx.execute("INSERT INTO goal_event(goal_id,kind,code,created_at_utc) VALUES(?1,'goal.completed','ARTIFACT_VERIFIED',strftime('%Y-%m-%dT%H:%M:%fZ','now'))",[&id])?;
        } else {
            tx.execute("UPDATE goal SET status='failed' WHERE id=?1", [&id])?;
            tx.execute("UPDATE job SET state='failed' WHERE goal_id=?1", [&id])?;
            tx.execute("INSERT INTO goal_event(goal_id,kind,code,created_at_utc) VALUES(?1,'goal.failed','ARTIFACT_MISMATCH',strftime('%Y-%m-%dT%H:%M:%fZ','now'))", [&id])?;
        }
        tx.commit()?;
        Ok(TickResult {
            outcome: if verified { "succeeded" } else { "failed" }.into(),
            goal_id: Some(id),
            reason: if verified {
                "FIXTURE_DIGEST_AND_LENGTH_VERIFIED"
            } else {
                "ARTIFACT_MISMATCH"
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

    pub fn queued_count(&self) -> Result<u64, EngineError> {
        let db = self.db.lock().expect("sqlite mutex poisoned");
        Ok(
            db.query_row("SELECT count(*) FROM job WHERE state='queued'", [], |r| {
                r.get(0)
            })?,
        )
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
}

fn read_goal(db: &Connection, id: &str) -> Result<GoalView, EngineError> {
    db.query_row("SELECT id,action,status,score,policy_code,provenance_status,(SELECT state='queued' FROM job WHERE goal_id=goal.id) FROM goal WHERE id=?1",[id],|r| Ok(GoalView {contract_version:CONTRACT_VERSION,id:r.get(0)?,action:r.get(1)?,status:r.get(2)?,score:r.get(3)?,policy_code:r.get(4)?,provenance_status:match r.get::<_,String>(5)?.as_str(){"verified"=>SourceStatus::Verified,"disputed"=>SourceStatus::Disputed,_=>SourceStatus::Unknown},queued:r.get::<_,bool>(6)?})).map_err(|e| if e == rusqlite::Error::QueryReturnedNoRows { EngineError::NotFound } else { EngineError::Storage(e) })
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
