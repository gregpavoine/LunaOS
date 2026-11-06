-- Visual privacy controller (BLIND). One row holds the mode and a capture
-- epoch: every entry into BLIND bumps the epoch, which invalidates every
-- capture lease issued before it. The audit table keeps coarse transitions
-- only, never frames, identities or anatomy, and cannot be rewritten.
CREATE TABLE IF NOT EXISTS privacy_control (
  singleton INTEGER PRIMARY KEY CHECK (singleton = 1),
  mode TEXT NOT NULL CHECK (mode IN ('STARTUP_LOCKED','NORMAL_EPHEMERAL_DETECTION','AUTHORIZED_CAPTURE_SESSION','BLIND_LATCHED','SENSOR_UNAVAILABLE')),
  capture_epoch INTEGER NOT NULL CHECK (capture_epoch >= 0),
  updated_at_utc TEXT NOT NULL
);
INSERT OR IGNORE INTO privacy_control(singleton, mode, capture_epoch, updated_at_utc)
VALUES (1, 'STARTUP_LOCKED', 0, strftime('%Y-%m-%dT%H:%M:%fZ','now'));

CREATE TABLE IF NOT EXISTS privacy_event (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  kind TEXT NOT NULL,
  code TEXT NOT NULL,
  created_at_utc TEXT NOT NULL
);
CREATE TRIGGER IF NOT EXISTS privacy_event_no_update BEFORE UPDATE ON privacy_event
BEGIN SELECT RAISE(ABORT, 'privacy_event is append-only'); END;
CREATE TRIGGER IF NOT EXISTS privacy_event_no_delete BEFORE DELETE ON privacy_event
BEGIN SELECT RAISE(ABORT, 'privacy_event is append-only'); END;

INSERT OR IGNORE INTO schema_migration(version, applied_at_utc)
VALUES (5, strftime('%Y-%m-%dT%H:%M:%fZ', 'now'));
