ALTER TABLE goal ADD COLUMN parent_id TEXT REFERENCES goal(id);

CREATE TABLE IF NOT EXISTS goal_dependency (
  goal_id TEXT NOT NULL REFERENCES goal(id) ON DELETE CASCADE,
  depends_on_goal_id TEXT NOT NULL REFERENCES goal(id) ON DELETE RESTRICT,
  dependency_type TEXT NOT NULL CHECK (dependency_type = 'blocks'),
  created_at_utc TEXT NOT NULL,
  PRIMARY KEY (goal_id, depends_on_goal_id),
  CHECK (goal_id <> depends_on_goal_id)
);
CREATE INDEX IF NOT EXISTS goal_dependency_by_dependency ON goal_dependency(depends_on_goal_id, goal_id);

CREATE TRIGGER IF NOT EXISTS goal_parent_no_cycle_insert
BEFORE INSERT ON goal WHEN NEW.parent_id IS NOT NULL
BEGIN
  SELECT CASE WHEN EXISTS (
    WITH RECURSIVE edges(src,dst) AS (
      SELECT goal_id, depends_on_goal_id FROM goal_dependency
      UNION ALL SELECT id, parent_id FROM goal WHERE parent_id IS NOT NULL
    ), reachable(id) AS (
      SELECT NEW.parent_id
      UNION
      SELECT edges.dst FROM edges JOIN reachable ON edges.src = reachable.id
    )
    SELECT 1 FROM reachable WHERE id = NEW.id
  ) THEN RAISE(ABORT, 'parent_cycle') END;
END;
CREATE TRIGGER IF NOT EXISTS goal_parent_no_cycle_update
BEFORE UPDATE OF parent_id ON goal WHEN NEW.parent_id IS NOT NULL
BEGIN
  SELECT CASE WHEN EXISTS (
    WITH RECURSIVE edges(src,dst) AS (
      SELECT goal_id, depends_on_goal_id FROM goal_dependency
      UNION ALL SELECT id, parent_id FROM goal WHERE parent_id IS NOT NULL AND id != NEW.id
    ), reachable(id) AS (
      SELECT NEW.parent_id
      UNION
      SELECT edges.dst FROM edges JOIN reachable ON edges.src = reachable.id
    )
    SELECT 1 FROM reachable WHERE id = NEW.id
  ) THEN RAISE(ABORT, 'parent_cycle') END;
END;
CREATE TRIGGER IF NOT EXISTS goal_dependency_limit_insert
BEFORE INSERT ON goal_dependency
BEGIN
  SELECT CASE WHEN (SELECT count(*) FROM goal_dependency WHERE goal_id = NEW.goal_id) >= 32
    THEN RAISE(ABORT, 'dependency_limit') END;
  SELECT CASE WHEN EXISTS (
    WITH RECURSIVE edges(src,dst) AS (
      SELECT goal_id, depends_on_goal_id FROM goal_dependency
      UNION ALL SELECT id, parent_id FROM goal WHERE parent_id IS NOT NULL
    ), reachable(id) AS (
      SELECT NEW.depends_on_goal_id
      UNION
      SELECT edges.dst FROM edges JOIN reachable ON edges.src = reachable.id
    )
    SELECT 1 FROM reachable WHERE id = NEW.goal_id
  ) THEN RAISE(ABORT, 'dependency_cycle') END;
END;

INSERT OR IGNORE INTO schema_migration(version, applied_at_utc)
VALUES (3, strftime('%Y-%m-%dT%H:%M:%fZ', 'now'));
