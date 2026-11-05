-- Bounded strategy trials. A failed goal is retried only as an explicit new
-- attempt that names its strategy and hypothesis, starts from the failure it
-- answers (baseline_code) and records its measured outcome. Five attempts is
-- the ceiling: replanning cannot loop.
CREATE TABLE IF NOT EXISTS strategy_trial (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  goal_id TEXT NOT NULL REFERENCES goal(id),
  attempt INTEGER NOT NULL CHECK (attempt BETWEEN 1 AND 5),
  strategy TEXT NOT NULL CHECK (length(strategy) BETWEEN 1 AND 64),
  hypothesis TEXT NOT NULL CHECK (length(hypothesis) BETWEEN 1 AND 1024),
  baseline_code TEXT NOT NULL,
  outcome TEXT NOT NULL CHECK (outcome IN ('pending','succeeded','failed','blocked')),
  outcome_code TEXT,
  created_at_utc TEXT NOT NULL,
  finished_at_utc TEXT,
  UNIQUE (goal_id, attempt)
);
CREATE INDEX IF NOT EXISTS strategy_trial_pending ON strategy_trial(goal_id) WHERE outcome = 'pending';

INSERT OR IGNORE INTO schema_migration(version, applied_at_utc)
VALUES (6, strftime('%Y-%m-%dT%H:%M:%fZ', 'now'));
