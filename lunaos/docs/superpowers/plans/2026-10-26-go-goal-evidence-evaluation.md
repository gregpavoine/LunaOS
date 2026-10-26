# Goal Evidence Evaluation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the documented owner-triggered goal evidence evaluation for the one real, safe local checksum action.

**Architecture:** The evaluator reads persisted goal and evidence state, independently recomputes the expected checksum/byte count from the known public fixture, classifies unsupported or missing evidence as insufficient, and appends a privacy-safe evaluation event transactionally. It never upgrades goal state or treats model output as proof.

**Tech Stack:** Go, `database/sql`, SQLite, existing `net/http` handler and tests.

**Spec:** `lunaos/contracts/openapi/local.json`, `POST /v1/goals/{id}/evaluate`; `lunaos/docs/LUNA_GOAL_ENGINE_CODEX_SPEC_EN.md`, evidence-backed evaluation and no false completion.

## Global Constraints

- Owner authentication remains required through the common Core handler.
- Evaluation is read-only with respect to goal/job state and appends only a safe audit event.
- A missing or unsupported evidence source is `insufficient_evidence`; never fabricate success.
- This restricted checksum evaluator is not the full strategy optimizer or a general artifact verifier.
- Do not access Linux hosts or modify installed systems.

## Review Focus

- An unauthorized caller must not trigger evaluation or create an audit event.
- Missing goals return not found; incomplete and unsupported goals return insufficient evidence.
- Only a completed `local.checksum` goal with the exact fixture digest, length and provenance can be verified.
- Tampered persisted evidence must not be reported as verified.
- Evaluation must not alter goal/job status or evidence and must append one audit event per accepted request.

---

### Task 1: Bounded evidence evaluation

**Files:**
- Modify: `lunaos/services/core/types.go`
- Modify: `lunaos/services/core/engine.go`
- Modify: `lunaos/services/core/http.go`
- Test: `lunaos/services/core/http_test.go`
- Modify: `lunaos/contracts/openapi/local.json`
- Modify: `lunaos/docs/STATUS.md`
- Modify: `lunaos/docs/release/MONT_BLANC_FEATURE_MATRIX.md`

**Interfaces:**
- Produces: `Engine.EvaluateEvidence(ctx context.Context, goalID string) (GoalEvaluation, error)`; outcome is `verified`, `insufficient_evidence`, or `mismatch`.

- [x] Add failing tests for owner-only route, empty body, successful fixture verification, missing evidence, mismatch, audit, and unchanged state.
- [x] Implement deterministic fixture verification and transactional append-only audit.
- [x] Update OpenAPI and truthfully scope docs as local implementation, not strategy evaluation.
- [x] Run focused test red/green, full `make test check`, Core race suite, and inspect diff.
