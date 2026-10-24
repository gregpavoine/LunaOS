BEGIN IMMEDIATE;
CREATE TABLE IF NOT EXISTS schema_migration(version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS goal (
 id TEXT PRIMARY KEY, owner_id TEXT NOT NULL CHECK(owner_id='owner'),
 action TEXT NOT NULL, privacy_class TEXT NOT NULL CHECK(privacy_class IN ('general','intimate','biometric')),
 status TEXT NOT NULL CHECK(status IN ('approved','blocked','completed','failed')),
 metrics TEXT NOT NULL, computed_priority REAL CHECK(computed_priority BETWEEN 0 AND 100),
 priority_explanation TEXT NOT NULL, decision_code TEXT NOT NULL, assertion TEXT,
 grant_token TEXT NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL,
 version INTEGER NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS goal_execution (
 id TEXT PRIMARY KEY, goal_id TEXT NOT NULL UNIQUE REFERENCES goal(id),
 idempotency_key TEXT NOT NULL UNIQUE, request_hash TEXT NOT NULL,
 execution_state TEXT NOT NULL CHECK(execution_state IN ('queued','running','verifying','succeeded','failed','blocked')),
 attempt INTEGER NOT NULL DEFAULT 0, finished_at TEXT
);
CREATE TABLE IF NOT EXISTS evidence (
 id TEXT PRIMARY KEY, goal_id TEXT NOT NULL UNIQUE REFERENCES goal(id),
 sha256 TEXT NOT NULL CHECK(length(sha256)=64), byte_count INTEGER NOT NULL,
 provenance TEXT NOT NULL, verified_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS goal_event (
 event_id TEXT PRIMARY KEY, goal_id TEXT REFERENCES goal(id), event_type TEXT NOT NULL,
 decision_code TEXT NOT NULL, schema_version INTEGER NOT NULL DEFAULT 1,
 correlation_id TEXT NOT NULL, occurred_at TEXT NOT NULL
);
CREATE TRIGGER IF NOT EXISTS goal_event_no_update BEFORE UPDATE ON goal_event BEGIN SELECT RAISE(ABORT,'append only'); END;
CREATE TRIGGER IF NOT EXISTS goal_event_no_delete BEFORE DELETE ON goal_event BEGIN SELECT RAISE(ABORT,'append only'); END;
CREATE TABLE IF NOT EXISTS event_outbox (
 event_id TEXT PRIMARY KEY REFERENCES goal_event(event_id), delivered_at TEXT
);
CREATE TABLE IF NOT EXISTS event_inbox (
 event_id TEXT PRIMARY KEY REFERENCES event_outbox(event_id), received_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS strategy_trial (
 id TEXT PRIMARY KEY, goal_id TEXT NOT NULL UNIQUE REFERENCES goal(id),
 version INTEGER NOT NULL, hypothesis TEXT NOT NULL, measurements TEXT NOT NULL,
 status TEXT NOT NULL CHECK(status IN ('passed','failed')), comparison_to_baseline TEXT NOT NULL
);
INSERT OR IGNORE INTO schema_migration VALUES(1,strftime('%Y-%m-%dT%H:%M:%fZ','now'));
COMMIT;
