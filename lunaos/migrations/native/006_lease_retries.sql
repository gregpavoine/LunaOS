BEGIN IMMEDIATE;
ALTER TABLE job ADD COLUMN max_attempts INTEGER NOT NULL DEFAULT 3 CHECK(max_attempts BETWEEN 1 AND 10);
ALTER TABLE job ADD COLUMN heartbeat_at_unix INTEGER;
INSERT OR IGNORE INTO schema_migration(version,applied_at_utc)
VALUES(6,strftime('%Y-%m-%dT%H:%M:%fZ','now'));
COMMIT;
