CREATE TABLE IF NOT EXISTS items (
  provider TEXT NOT NULL,
  external_id TEXT NOT NULL,
  blob_name TEXT NOT NULL DEFAULT '',
  content_level TEXT NOT NULL,
  license TEXT NOT NULL,
  bytes INTEGER NOT NULL,
  fetched_at INTEGER NOT NULL,
  meta BLOB NOT NULL,
  PRIMARY KEY (provider, external_id)
);
CREATE INDEX IF NOT EXISTS items_blob_name ON items(blob_name) WHERE blob_name <> '';
CREATE TABLE IF NOT EXISTS cursors (
  provider TEXT NOT NULL,
  scope TEXT NOT NULL,
  cursor TEXT NOT NULL,
  PRIMARY KEY (provider, scope)
);
