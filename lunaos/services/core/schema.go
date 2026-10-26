package core

import _ "embed"

//go:embed schema.sql
var schemaSQL string

//go:embed migrations/003_goal_dependency.sql
var goalDependencyMigration string
