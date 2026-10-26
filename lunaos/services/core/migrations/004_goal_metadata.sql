ALTER TABLE goal ADD COLUMN title TEXT NOT NULL DEFAULT '' CHECK (length(title) <= 256);
ALTER TABLE goal ADD COLUMN description TEXT NOT NULL DEFAULT '' CHECK (length(description) <= 4096);
UPDATE goal SET title = action WHERE title = '';
INSERT OR IGNORE INTO schema_migration(version, applied_at_utc)
VALUES (4, strftime('%Y-%m-%dT%H:%M:%fZ', 'now'));
