PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;
PRAGMA synchronous = FULL;

CREATE TABLE IF NOT EXISTS goal (
  id TEXT PRIMARY KEY,
  action TEXT NOT NULL,
  privacy_class TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('approved','blocked','completed','failed')),
  metrics_json TEXT NOT NULL,
  score REAL CHECK (score IS NULL OR score BETWEEN 0 AND 100),
  policy_code TEXT NOT NULL,
  provenance_status TEXT NOT NULL CHECK (provenance_status IN ('unknown','verified','disputed')),
  assertion_json TEXT,
  idempotency_key TEXT NOT NULL UNIQUE,
  request_hash TEXT NOT NULL,
  grant_token TEXT,
  created_at_utc TEXT NOT NULL,
  updated_at_utc TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS job (
  goal_id TEXT PRIMARY KEY REFERENCES goal(id),
  state TEXT NOT NULL CHECK (state IN ('queued','running','succeeded','blocked','failed')),
  attempt_count INTEGER NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
  max_attempts INTEGER NOT NULL DEFAULT 3 CHECK (max_attempts BETWEEN 1 AND 10),
  lease_owner TEXT,
  lease_token TEXT,
  lease_expires_at_unix INTEGER,
  CHECK ((state='running' AND lease_owner IS NOT NULL AND lease_token IS NOT NULL AND lease_expires_at_unix IS NOT NULL)
      OR (state!='running' AND lease_owner IS NULL AND lease_token IS NULL AND lease_expires_at_unix IS NULL))
);

CREATE TABLE IF NOT EXISTS evidence (
  goal_id TEXT PRIMARY KEY REFERENCES goal(id),
  sha256 TEXT NOT NULL CHECK (length(sha256)=64),
  byte_count INTEGER NOT NULL CHECK (byte_count >= 0),
  provenance TEXT NOT NULL,
  verified INTEGER NOT NULL CHECK (verified IN (0,1)),
  created_at_utc TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS goal_event (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  goal_id TEXT NOT NULL REFERENCES goal(id),
  kind TEXT NOT NULL,
  code TEXT NOT NULL,
  created_at_utc TEXT NOT NULL
);
CREATE TRIGGER IF NOT EXISTS goal_event_no_update BEFORE UPDATE ON goal_event
BEGIN SELECT RAISE(ABORT, 'goal_event is append-only'); END;
CREATE TRIGGER IF NOT EXISTS goal_event_no_delete BEFORE DELETE ON goal_event
BEGIN SELECT RAISE(ABORT, 'goal_event is append-only'); END;

CREATE TABLE IF NOT EXISTS schema_migration (
  version INTEGER PRIMARY KEY,
  applied_at_utc TEXT NOT NULL
);
INSERT OR IGNORE INTO schema_migration(version, applied_at_utc)
VALUES (1, strftime('%Y-%m-%dT%H:%M:%fZ','now'));

CREATE TABLE IF NOT EXISTS intimacy_control (
  singleton INTEGER PRIMARY KEY CHECK (singleton = 1),
  stopped INTEGER NOT NULL CHECK (stopped IN (0,1)),
  privacy_mode TEXT NOT NULL CHECK (privacy_mode IN ('unknown','private','public')),
  updated_at_utc TEXT NOT NULL
);
INSERT OR IGNORE INTO intimacy_control(singleton, stopped, privacy_mode, updated_at_utc)
VALUES (1, 1, 'unknown', strftime('%Y-%m-%dT%H:%M:%fZ','now'));

-- Additive overlay keeps older goal rows readable and persists owner controls.
CREATE TABLE IF NOT EXISTS goal_control (
  goal_id TEXT PRIMARY KEY REFERENCES goal(id),
  state TEXT NOT NULL CHECK (state IN ('paused','cancelled')),
  updated_at_utc TEXT NOT NULL
);
INSERT OR IGNORE INTO schema_migration(version, applied_at_utc)
VALUES (2, strftime('%Y-%m-%dT%H:%M:%fZ','now'));
