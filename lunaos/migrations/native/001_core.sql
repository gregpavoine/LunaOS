CREATE TABLE IF NOT EXISTS schema_migration (
  version INTEGER PRIMARY KEY,
  applied_at_utc TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS goal (
  id TEXT PRIMARY KEY,
  action TEXT NOT NULL,
  status TEXT NOT NULL CHECK(status IN ('approved','blocked','completed','failed')),
  score REAL CHECK(score BETWEEN 0 AND 100),
  policy_code TEXT NOT NULL,
  provenance_status TEXT NOT NULL CHECK(provenance_status IN ('unknown','verified','disputed')),
  idempotency_key TEXT UNIQUE NOT NULL,
  request_hash TEXT NOT NULL,
  created_at_utc TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS job (
  goal_id TEXT PRIMARY KEY REFERENCES goal(id),
  state TEXT NOT NULL CHECK(state IN ('queued','succeeded','blocked','failed'))
);

CREATE TABLE IF NOT EXISTS goal_grant (
  goal_id TEXT PRIMARY KEY REFERENCES goal(id),
  grant TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS evidence (
  goal_id TEXT PRIMARY KEY REFERENCES goal(id),
  sha256 TEXT NOT NULL CHECK(length(sha256)=64),
  byte_count INTEGER NOT NULL CHECK(byte_count>=0),
  provenance TEXT NOT NULL,
  verified INTEGER NOT NULL CHECK(verified IN (0,1))
);

CREATE TABLE IF NOT EXISTS goal_event (
  id INTEGER PRIMARY KEY,
  goal_id TEXT NOT NULL REFERENCES goal(id),
  kind TEXT NOT NULL,
  code TEXT NOT NULL,
  created_at_utc TEXT NOT NULL
);

INSERT OR IGNORE INTO schema_migration VALUES(1,strftime('%Y-%m-%dT%H:%M:%fZ','now'));
