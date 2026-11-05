package core

import _ "embed"

//go:embed schema.sql
var schemaSQL string

//go:embed migrations/003_goal_dependency.sql
var goalDependencyMigration string

//go:embed migrations/004_goal_metadata.sql
var goalMetadataMigration string

//go:embed migrations/005_privacy_control.sql
var privacyControlMigration string

//go:embed migrations/006_strategy_trials.sql
var strategyTrialMigration string
