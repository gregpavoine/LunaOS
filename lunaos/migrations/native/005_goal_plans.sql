BEGIN IMMEDIATE;
CREATE TABLE goal_parent (
  goal_id TEXT PRIMARY KEY REFERENCES goal(id),
  parent_goal_id TEXT NOT NULL REFERENCES goal(id),
  created_at_utc TEXT NOT NULL,
  CHECK(goal_id<>parent_goal_id)
);
CREATE INDEX goal_parent_parent_idx ON goal_parent(parent_goal_id);
CREATE TRIGGER goal_dependency_limit
BEFORE INSERT ON goal_dependency
WHEN (SELECT count(*) FROM goal_dependency WHERE goal_id=NEW.goal_id)>=32
 AND NOT EXISTS (SELECT 1 FROM goal_dependency WHERE goal_id=NEW.goal_id AND depends_on_goal_id=NEW.depends_on_goal_id)
BEGIN
  SELECT RAISE(ABORT,'goal dependency limit exceeded');
END;
INSERT OR IGNORE INTO schema_migration(version,applied_at_utc)
VALUES(5,strftime('%Y-%m-%dT%H:%M:%fZ','now'));
COMMIT;
