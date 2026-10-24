PRAGMA foreign_keys=OFF;
BEGIN IMMEDIATE;
CREATE TABLE goal_v3 (
  id TEXT PRIMARY KEY,
  action TEXT NOT NULL,
  status TEXT NOT NULL CHECK(status IN ('idea','planned','approved','active','blocked','paused','completed','cancelled','failed')),
  score REAL CHECK(score BETWEEN 0 AND 100),
  policy_code TEXT NOT NULL,
  provenance_status TEXT NOT NULL CHECK(provenance_status IN ('unknown','verified','disputed','stale')),
  idempotency_key TEXT UNIQUE NOT NULL,
  request_hash TEXT NOT NULL,
  created_at_utc TEXT NOT NULL
);
INSERT INTO goal_v3(id,action,status,score,policy_code,provenance_status,idempotency_key,request_hash,created_at_utc)
SELECT id,action,status,score,policy_code,provenance_status,idempotency_key,request_hash,created_at_utc FROM goal;
DROP TABLE goal;
ALTER TABLE goal_v3 RENAME TO goal;

CREATE TABLE job_v3 (
  goal_id TEXT PRIMARY KEY REFERENCES goal(id),
  state TEXT NOT NULL CHECK(state IN ('queued','running','paused','succeeded','blocked','cancelled','failed')),
  lease_owner TEXT,
  lease_token TEXT,
  lease_expires_at_unix INTEGER,
  attempt_count INTEGER NOT NULL DEFAULT 0 CHECK(attempt_count>=0),
  CHECK((state='running' AND lease_owner IS NOT NULL AND lease_token IS NOT NULL AND lease_expires_at_unix IS NOT NULL)
     OR (state!='running' AND lease_owner IS NULL AND lease_token IS NULL AND lease_expires_at_unix IS NULL))
);
INSERT INTO job_v3(goal_id,state,lease_owner,lease_token,lease_expires_at_unix,attempt_count)
SELECT goal_id,state,lease_owner,lease_token,lease_expires_at_unix,attempt_count FROM job;
DROP TABLE job;
ALTER TABLE job_v3 RENAME TO job;

INSERT OR IGNORE INTO schema_migration(version,applied_at_utc)
VALUES(3,strftime('%Y-%m-%dT%H:%M:%fZ','now'));
COMMIT;
PRAGMA foreign_keys=ON;
