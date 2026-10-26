# Go Goal Metadata Plan

## Goal

Add persistent owner-visible goal title/description fields and authenticated metadata editing in Go Core without changing action authorization or invalidating old idempotency keys.

## Rules

- Persist bounded plain-text title (1–256 Unicode characters) and description (0–4096 Unicode characters); title defaults to the action for legacy clients.
- Add a transactional forward-only SQLite migration and preserve existing rows, grants, evidence and event history.
- Permit metadata edits only for approved or policy-blocked queued work; reject running and terminal/paused/cancelled states.
- Owner authentication is mandatory; updates are atomic and append-only audited. Metadata edits cannot alter action, policy grant, score, provenance, evidence, or execution state.
- Preserve canonical request hashes for legacy create payloads that omit the new fields.

## Tasks

1. Add failing tests for create/default metadata, upgrade compatibility, owner-only edit, validation, state conflicts, persistence and unchanged execution authorization.
2. Implement migration 4, storage/read models, engine update and PATCH operation.
3. Update OpenAPI, feature matrix and status, then run focused/full Go checks and race tests.
4. Commit and push the feature branch, merge to `mont-blanc`, and verify workflow results.

## Acceptance

Existing databases reopen without lost rows; old omitted-field create requests retain their prior idempotency hash; metadata survives restart; invalid/unauthorized edits do not mutate; execution action/grant/status remain unchanged; tests and `make test check` pass. No claim of complete goal CRUD or Mont Blanc release readiness.
