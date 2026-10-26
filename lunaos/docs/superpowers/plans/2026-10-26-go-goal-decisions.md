# Goal Decisions API Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the already documented owner-authenticated goal decisions endpoint over the existing append-only audit log.

**Architecture:** Core queries only goal event ID, kind, decision code and UTC timestamp, validates that the goal exists, orders the result chronologically, and returns at most the latest 100 events in chronological order. HTTP authentication remains centralized in the existing handler.

**Tech Stack:** Go `database/sql`, SQLite, existing `net/http` handler and Go tests.

**Spec:** `lunaos/contracts/openapi/local.json`, `/v1/goals/{id}/decisions`; `lunaos/docs/LUNA_GOAL_ENGINE_CODEX_SPEC_EN.md`, REST contracts and append-only safe events.

## Global Constraints

- Owner authentication is required.
- Events remain append-only and privacy-safe; never return goal descriptions, assertions, credentials, grants or evidence payloads from this endpoint.
- Preserve existing behavior and keep Linux/systemd/device access behind existing adapters.
- Do not claim full strategy evaluation from returning audit decisions.

## Review Focus

- A caller without owner authentication must not read a goal's event history.
- Unknown goal IDs must return not found rather than an empty event collection.
- More than 100 events must return the latest 100 in chronological order.
- Empty history must serialize as an empty array, not null.
- Response fields must not expose private goal or policy grant data.

---

### Task 1: Read-only decision history

**Files:**
- Modify: `lunaos/services/core/types.go`
- Modify: `lunaos/services/core/engine.go`
- Modify: `lunaos/services/core/http.go`
- Test: `lunaos/services/core/http_test.go`
- Modify: `lunaos/contracts/openapi/local.json`
- Modify: `lunaos/docs/release/MONT_BLANC_FEATURE_MATRIX.md`
- Modify: `lunaos/docs/STATUS.md`

**Interfaces:**
- Produces: `Engine.Decisions(ctx context.Context, goalID string) ([]GoalDecision, error)`; each `GoalDecision` contains `id`, `kind`, `code`, and `created_at_utc` only.

- [x] Add handler tests for owner auth, missing goal, ordered safe fields, and bounded history; run them and verify expected failures.
- [x] Implement the query and route with a 100-event cap, preserving chronological response order.
- [x] Update contract and truthful status/matrix wording.
- [x] Run focused tests, full `make test check`, and race tests; inspect the diff.
