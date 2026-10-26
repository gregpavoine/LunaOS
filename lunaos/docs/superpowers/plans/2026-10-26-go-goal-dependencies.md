# Go Goal Dependency Parity Plan

## Goal

Port the bounded persistent dependency behavior from the frozen Rust checkpoint into the Go Core, without expanding the Python prototype. Preserve existing databases and keep the scheduler policy recheck in place.

## Scope and invariants

- Add an additive, once-only SQLite migration for `goal.parent_id` and `goal_dependency`.
- Owner-authenticated `POST /v1/goals/{id}/plan` atomically replaces one goal's parent and blocking dependency list; maximum 32 dependencies.
- Reject malformed, duplicate, self, missing and cyclic references. SQL constraints/triggers provide a second integrity boundary.
- Scheduler only claims a goal when every dependency's persisted goal status is `completed`; policy evaluation still runs before execution.
- Update the Go OpenAPI contract and feature/status evidence. No model planning, new privileged action, or target-hardware claim.

## Tasks

1. Add failing DB/API/scheduler tests for plan auth, validation, atomic replacement, cycles, prerequisite ordering and reopen persistence.
2. Add safe forward-only migration and typed planner API.
3. Gate scheduler claims on completed prerequisites and add REST/OpenAPI behavior.
4. Run focused and full Go checks, update the feature matrix and status, inspect diff, commit one atomic feature change, and push this feature branch if remote access succeeds.

## Acceptance

The focused tests and `make test check` pass; existing SQLite rows survive migration; denied owner calls cannot mutate plans; blocked prerequisites cannot execute early; policy revalidation remains covered. Arch/systemd/ISO readiness remains outside this milestone.
