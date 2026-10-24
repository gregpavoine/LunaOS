# Goal Dependency Planner Implementation Plan

> **For agentic workers:** Implement this plan task by task with tests before production changes. The owner authorized continuing and pushing the work on the current feature branch.

**Goal:** Persist bounded goal parent/dependency plans, reject cycles, and prevent a queued goal from running until all blocking prerequisites have completed.

**Architecture:** Add a forward-only SQLite migration for parent and dependency metadata. Keep planner mutations in the authenticated Core API and `Engine`; validate the complete replacement plan in one write transaction, with SQL constraints/triggers as a second integrity boundary. The scheduler reads persisted dependencies and continues to call the existing independent policy broker before execution.

**Tech Stack:** Python 3 standard library, SQLite recursive CTEs and triggers, Go policy broker unchanged, OpenAPI 3.1 JSON.

**Spec:** Local owner source `docs/LUNA_GOAL_ENGINE_CODEX_SPEC_EN.md` §§2–5, 8; public-repo architecture and implementation plan.

## Global Constraints

- Keep LunaOS additive under `lunaos/`; preserve the Omarchy parent tree.
- SQLite migrations must upgrade existing task-1 databases without deleting or rewriting goal, evidence, event, or execution data.
- Owner authentication and the policy gate remain mandatory; planning cannot grant execution permission.
- A dependency is ready only when its goal status is `completed`; missing, blocked, paused, failed, or active prerequisites never permit dependent execution.
- The submitted graph is bounded to 32 prerequisites per goal; reject duplicate, self, missing, and cyclic references.
- Do not add a model-generated planner, hardware claim, or external side effect in this milestone.

---

### Task 1: Persist parent and dependency graph

**Files:**
- Create: `migrations/core/002_goal_dependency.sql`
- Modify: `luna/store.py`
- Test: `tests/test_slice.py`

**Interfaces:**
- Store loads ordered numbered migrations and applies each once using `schema_migration`.
- Migration adds nullable `goal.parent_id` and `goal_dependency(goal_id, depends_on_goal_id, dependency_type, created_at)` with owner-safe foreign keys, uniqueness, and cycle/self-reference protection.

- [x] Add a test that creates and closes a version-1 database with one goal, reopens it through the new Store, and verifies version 2 plus preserved row/evidence/event contents.
- [x] Run that test and confirm it fails because migration 2 is absent/unapplied.
- [x] Add migration 2 and ordered, repeat-safe migration discovery.
- [x] Re-run the migration test with `python3 -m unittest discover -s tests -p test_slice.py -v`; existing goal, evidence and event rows survive.
- [x] Commit the schema/upgrade unit in the Milestone 1b checkpoint.

### Task 2: Validate owner-authenticated bounded plans

**Files:**
- Modify: `luna/engine.py`
- Test: `tests/test_slice.py`

**Interfaces:**
- `Engine.set_plan(goal_id, depends_on, parent_id, token)` atomically replaces one goal's dependency list and optional parent; it returns the plan with `parent_id` and ordered dependency IDs.
- `Engine.get(goal_id)` includes the persisted plan.
- Raise `Invalid('PLAN_TOO_LARGE')` above 32 dependencies and `Invalid('PLAN_INVALID')` for malformed/duplicate/self/missing/cyclic input; raise `Unauthorized` for a non-owner token.

- [x] Add real-database tests for valid prerequisites, replacement, missing IDs, duplicate IDs, self-edge, a multi-edge dependency cycle, parent cycles, and unauthenticated mutation.
- [x] Run plan tests and observe expected failure because `Engine.set_plan` was absent.
- [x] Implement plan validation and graph writes in one transaction; SQL triggers reject dependency and parent cycles.
- [x] Run focused tests and confirm all plan failure codes and rows are atomic.
- [x] Commit the registry unit in the Milestone 1b checkpoint.

### Task 3: Gate scheduler and expose REST plan contract

**Files:**
- Modify: `luna/engine.py`
- Modify: `luna/api.py`
- Modify: `contracts/openapi/local.json`
- Test: `tests/test_slice.py`

**Interfaces:**
- `POST /v1/goals/{id}/plan` accepts exactly `depends_on` (array of goal IDs) and nullable `parent_id`; it returns the authenticated goal representation including its plan.
- `tick()` selects queued goals only if all blocking dependency goals are completed; it still rechecks the broker before any execution.

- [x] Add API contract/auth tests for successful plan write, unauthenticated rejection, malformed request, and a nonexistent target.
- [x] Add an integration test proving a goal cannot execute ahead of its prerequisite and runs only after a subsequent tick observes prerequisite completion.
- [x] Run the new API test and observe 404 before the endpoint implementation.
- [x] Implement the endpoint, OpenAPI schemas/responses, and SQL readiness condition.
- [x] Run focused tests, then `make test check`, Go race tests, restart smoke, OpenAPI JSON parsing, and `git diff --check`.
- [x] Update architecture, incremental plan, status, and test evidence with actual results; commit and push the verified slice.

## Review Checklist

- No planner request can bypass the policy broker or alter a grant.
- Cycle checks work for direct database writes as well as API requests.
- Existing v1 database contents survive migration; applying Store initialization repeatedly is safe.
- Scheduler idles safely when prerequisites are unresolved and resumes when they complete.
- OpenAPI describes only behavior implemented and tested here.
- Linux, GPU, model routing, production MFA, and full Mont Blanc completion remain explicitly unverified or unimplemented.
