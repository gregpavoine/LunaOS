BEGIN IMMEDIATE;
ALTER TABLE goal ADD COLUMN parent_id TEXT REFERENCES goal(id);
CREATE TABLE IF NOT EXISTS goal_dependency (
 goal_id TEXT NOT NULL REFERENCES goal(id),
 depends_on_goal_id TEXT NOT NULL REFERENCES goal(id),
 dependency_type TEXT NOT NULL CHECK(dependency_type='blocks'),
 created_at TEXT NOT NULL,
 PRIMARY KEY(goal_id, depends_on_goal_id),
 CHECK(goal_id <> depends_on_goal_id)
);
CREATE INDEX IF NOT EXISTS goal_dependency_by_dependency ON goal_dependency(depends_on_goal_id, goal_id);
CREATE TRIGGER IF NOT EXISTS goal_parent_no_cycle_insert
BEFORE INSERT ON goal WHEN NEW.parent_id IS NOT NULL
BEGIN
 SELECT CASE WHEN NEW.parent_id=NEW.id OR EXISTS (
  WITH RECURSIVE ancestors(id) AS (
   SELECT NEW.parent_id
   UNION
   SELECT g.parent_id FROM goal g JOIN ancestors a ON g.id=a.id WHERE g.parent_id IS NOT NULL
  )
  SELECT 1 FROM ancestors WHERE id=NEW.id
 ) THEN RAISE(ABORT,'parent_cycle') END;
END;
CREATE TRIGGER IF NOT EXISTS goal_parent_no_cycle_update
BEFORE UPDATE OF parent_id ON goal WHEN NEW.parent_id IS NOT NULL
BEGIN
 SELECT CASE WHEN NEW.parent_id=NEW.id OR EXISTS (
  WITH RECURSIVE ancestors(id) AS (
   SELECT NEW.parent_id
   UNION
   SELECT g.parent_id FROM goal g JOIN ancestors a ON g.id=a.id WHERE g.parent_id IS NOT NULL
  )
  SELECT 1 FROM ancestors WHERE id=NEW.id
 ) THEN RAISE(ABORT,'parent_cycle') END;
END;
CREATE TRIGGER IF NOT EXISTS goal_dependency_limit_insert
BEFORE INSERT ON goal_dependency
BEGIN
 SELECT CASE WHEN (SELECT count(*) FROM goal_dependency WHERE goal_id=NEW.goal_id) >= 32
 THEN RAISE(ABORT,'dependency_limit') END;
 SELECT CASE WHEN EXISTS (
  WITH RECURSIVE reachable(id) AS (
   SELECT NEW.depends_on_goal_id
   UNION
   SELECT d.depends_on_goal_id FROM goal_dependency d JOIN reachable r ON d.goal_id=r.id
  )
  SELECT 1 FROM reachable WHERE id=NEW.goal_id
 ) THEN RAISE(ABORT,'dependency_cycle') END;
END;
CREATE TRIGGER IF NOT EXISTS goal_dependency_no_cycle_update
BEFORE UPDATE OF goal_id, depends_on_goal_id ON goal_dependency
BEGIN
 SELECT CASE WHEN (
  SELECT count(*) FROM goal_dependency
  WHERE goal_id=NEW.goal_id
  AND NOT (goal_id=OLD.goal_id AND depends_on_goal_id=OLD.depends_on_goal_id)
 ) >= 32 THEN RAISE(ABORT,'dependency_limit') END;
 SELECT CASE WHEN EXISTS (
  WITH RECURSIVE reachable(id) AS (
   SELECT NEW.depends_on_goal_id
   UNION
   SELECT d.depends_on_goal_id FROM goal_dependency d JOIN reachable r ON d.goal_id=r.id
   WHERE NOT (d.goal_id=OLD.goal_id AND d.depends_on_goal_id=OLD.depends_on_goal_id)
  )
  SELECT 1 FROM reachable WHERE id=NEW.goal_id
 ) THEN RAISE(ABORT,'dependency_cycle') END;
END;
INSERT OR IGNORE INTO schema_migration VALUES(2,strftime('%Y-%m-%dT%H:%M:%fZ','now'));
COMMIT;
