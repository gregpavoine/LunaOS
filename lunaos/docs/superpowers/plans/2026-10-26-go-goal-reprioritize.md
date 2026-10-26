# Go Goal Reprioritization Plan

## Goal

Complete the owner-controlled reprioritize operation in the Go Goal Engine API while preserving deny-before-score and scheduler authorization boundaries.

## Rules

- `PATCH /v1/goals/{id}` with `operation=reprioritize` requires all five bounded 0..1 score inputs and the owner credential.
- Only non-running, non-paused, non-cancelled goals may be reprioritized.
- Re-run Policy Broker authorization before computing or storing the new score. A denial clears the grant and blocks the job; it never stores a score.
- Successful updates atomically persist normalized metrics, score, policy code, scoped grant, queue state and append-only audit events.
- Do not change the action, expected cost, evidence or provenance. No model or hardware claims.

## Tasks

1. Add failing engine/API tests for owner checks, required/bounded metrics, policy denial before scoring, successful score persistence and protected lifecycle states.
2. Implement the reprioritize transaction in Go Core and expose it through the existing owner-only PATCH endpoint.
3. Update OpenAPI, status and feature matrix; run full checks and focused race tests.
4. Commit, publish feature branch, merge to `mont-blanc` and verify the CI results.

## Acceptance

All five metrics are explicitly provided and validated; unauthorized or invalid requests do not mutate state; policy denial leaves a blocked goal with no score/grant; authorized reprioritization changes priority and remains queued; tests and `make test check` pass. No claim of full CRUD or Mont Blanc release readiness.
