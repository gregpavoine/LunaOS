BEGIN IMMEDIATE;
CREATE TABLE event_outbox (
  event_id INTEGER PRIMARY KEY REFERENCES goal_event(id),
  attempts INTEGER NOT NULL DEFAULT 0 CHECK(attempts>=0),
  queued_at_utc TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  delivered_at_utc TEXT
);
CREATE TABLE event_inbox (
  event_id INTEGER PRIMARY KEY REFERENCES goal_event(id),
  received_at_utc TEXT NOT NULL
);
INSERT INTO event_outbox(event_id) SELECT id FROM goal_event;
CREATE TRIGGER goal_event_enqueue AFTER INSERT ON goal_event
BEGIN
  INSERT INTO event_outbox(event_id) VALUES(NEW.id);
END;
CREATE TRIGGER goal_event_no_update BEFORE UPDATE ON goal_event
BEGIN
  SELECT RAISE(ABORT,'goal_event is append-only');
END;
CREATE TRIGGER goal_event_no_delete BEFORE DELETE ON goal_event
BEGIN
  SELECT RAISE(ABORT,'goal_event is append-only');
END;
INSERT OR IGNORE INTO schema_migration(version,applied_at_utc)
VALUES(4,strftime('%Y-%m-%dT%H:%M:%fZ','now'));
COMMIT;
