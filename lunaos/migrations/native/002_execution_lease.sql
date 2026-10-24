BEGIN IMMEDIATE;
ALTER TABLE job RENAME TO job_pre_lease;
CREATE TABLE job (
  goal_id TEXT PRIMARY KEY REFERENCES goal(id),
  state TEXT NOT NULL CHECK(state IN ('queued','running','succeeded','blocked','failed')),
  lease_owner TEXT,
  lease_token TEXT,
  lease_expires_at_unix INTEGER,
  attempt_count INTEGER NOT NULL DEFAULT 0 CHECK(attempt_count>=0),
  CHECK((state='running' AND lease_owner IS NOT NULL AND lease_token IS NOT NULL AND lease_expires_at_unix IS NOT NULL)
     OR (state!='running' AND lease_owner IS NULL AND lease_token IS NULL AND lease_expires_at_unix IS NULL))
);
INSERT INTO job(goal_id,state,attempt_count)
SELECT goal_id,state,0 FROM job_pre_lease;
DROP TABLE job_pre_lease;
INSERT OR IGNORE INTO schema_migration(version,applied_at_utc)
VALUES(2,strftime('%Y-%m-%dT%H:%M:%fZ','now'));
COMMIT;
