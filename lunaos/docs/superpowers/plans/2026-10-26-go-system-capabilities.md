# System Capabilities API Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the documented owner-only capabilities endpoint with live Core, SQLite and Policy Broker readiness and explicit unavailable states for unimplemented LLM and unprobed hardware.

**Architecture:** Core reports that its API is live, probes its SQLite handle, and asks the existing Policy Broker over the authenticated Unix socket for readiness. Capability failures are represented as statuses, not fabricated successes; LLM and hardware stay unavailable/unverified until real integrations exist.

**Tech Stack:** Go, `net/http`, `database/sql`, Unix sockets, existing Policy Broker and tests.

**Spec:** `lunaos/contracts/openapi/local.json`, `GET /v1/system/capabilities`; `lunaos/docs/LUNAOS_MONT_BLANC_1_0_CODEX_UNIFIED_SPEC.md`, truthful readiness and hardware status.

## Global Constraints

- Owner authentication is mandatory on the Core endpoint and the broker readiness probe requires the Core credential and peer checks.
- Hardware and local model capability must not be reported as ready without real discovery/configuration.
- Keep all Linux system access behind existing adapters; make no host changes or remote connections.

## Review Focus

- Unauthenticated callers cannot inspect system readiness.
- Database readiness comes from a live SQLite ping.
- Broker readiness comes from a real authenticated Unix-socket request, and reports unavailable on failure.
- Model and hardware remain explicitly not ready.
- Capability reporting never mutates goals or policy state.

---

### Task 1: Truthful readiness reporting

**Files:**
- Modify: `lunaos/services/policy/main.go`
- Modify: `lunaos/services/policy/main_test.go`
- Modify: `lunaos/services/core/types.go`
- Modify: `lunaos/services/core/policy_client.go`
- Modify: `lunaos/services/core/engine.go`
- Modify: `lunaos/services/core/http.go`
- Test: `lunaos/services/core/http_test.go`
- Modify: `lunaos/contracts/openapi/local.json`
- Modify: `lunaos/docs/STATUS.md`
- Modify: `lunaos/docs/release/MONT_BLANC_FEATURE_MATRIX.md`

**Interfaces:**
- Produces: authenticated broker `GET /v1/system/health`; Core `GET /v1/system/capabilities` returns readiness per component.

- [x] Add failing tests for broker auth, endpoint owner auth, SQLite ready, broker unavailable/ready, and non-claims for LLM/hardware.
- [x] Implement the authenticated broker health endpoint and Core probe.
- [x] Update contract and status with exact unverified boundaries.
- [x] Run focused red/green tests, full `make test check`, Core race tests and local binary build.
