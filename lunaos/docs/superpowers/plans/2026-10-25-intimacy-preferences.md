# Intimacy Private Preferences Implementation Plan

> **For agentic workers:** execute each checkbox on branch `feature/intimacy-preferences`; do not merge until the complete feature verification passes.

**Goal:** Add a consent-scoped, encrypted owner preference store with explicit declaration, correction, listing and deletion over the local Core API.

**Architecture:** Extend the existing separate AES-GCM vault with a second encrypted table for preferences. Keep category, value, timescale and evidence inside the ciphertext; expose only owner-authenticated routes that require `preferences.read` or `preferences.write` consent. User declaration and correction are the only supported preference sources; no automatic inference is claimed.

**Tech Stack:** Go 1.20, `database/sql`, existing CGO SQLite driver, standard-library AES-256-GCM, local authenticated HTTP routes.

**Spec:** `docs/LUNAOS_MONT_BLANC_1_0_CODEX_UNIFIED_SPEC.md`, sections 6.2, 6.3, 6.8, 7, and 8.

## Global Constraints

- Private payloads stay in the separate encrypted vault; do not add them to Core SQLite, normal logs, errors, or write responses.
- Preference routes require an active owner consent with the exact read/write scope; unknown, expired, revoked, cross-subject or underscoped grants fail closed.
- Preference origin is only `declared` in this feature. Do not invent automatic or observed inference.
- Corrections replace declared values, preserve `first_seen_at`, set confidence to 1.0, increment evidence count and refresh `last_seen_at`.
- Deletion remains available after consent revocation.
- No models, intimate samples, or real personal preference values in tests or repository fixtures.
- Complete and test this branch before merging it into `mont-blanc`.

---

### Task 1: Encrypted preference persistence

**Files:**
- Modify: `services/intimacy/vault.go`
- Modify: `services/intimacy/vault_test.go`

**Interfaces:**
- Consumes: `Vault`, `Scope`, owner-only consent checks and AES-GCM helpers.
- Produces: `Preference`, `CreatePreference`, `ListPreferences`, `CorrectPreference`, `DeletePreference`.

- [ ] Add tests for encrypted round trip, consent scope, correction provenance, safe bounds, ciphertext integrity, and deletion after revoke.
- [ ] Run `go test ./services/intimacy -run Preference -count=1`; confirm missing API or expected behavior fails.
- [ ] Add `preferences.read` and `preferences.write` scopes. Store a JSON preference payload encrypted with associated data `preference_id || owner || preferences.v1`; leave only opaque ID and timestamps in SQLite.
- [ ] Implement declaration, list, correction and deletion. Restrict dimensions/timescales to documented enums, require explicit owner declaration, and never accept inferred/observed source in this API.
- [ ] Run `go test ./services/intimacy -run Preference -count=1`; all new cases pass.

### Task 2: Owner-only REST endpoints

**Files:**
- Modify: `services/core/http.go`
- Create: `services/core/preference_http_test.go`
- Modify: `contracts/openapi/local.json`

**Interfaces:**
- Consumes: the preference methods from Task 1 and the existing optional `HTTPHandler.vault`.
- Produces: `POST/GET /v1/intimacy/preferences`, `POST /v1/intimacy/preferences/{id}/correction`, and `DELETE /v1/intimacy/preferences/{id}`.

- [ ] Add tests for owner authentication, missing/revoked/underscoped consent, declaration persistence, correction precedence, no plaintext in create response, private GET, and delete after revoke.
- [ ] Run `go test ./services/core -run Preference -count=1`; verify the new endpoints fail before implementation.
- [ ] Implement strict bounded JSON routes. `POST` accepts only explicit declared preferences; `GET` requires `consent_id`; correction requires active `preferences.write`; DELETE is owner-only and remains usable after revocation.
- [ ] Add matching OpenAPI schemas, enums, consent errors and response limits.
- [ ] Run Core preference tests and `python3 -m json.tool contracts/openapi/local.json`.

### Task 3: Docs, release status and complete branch verification

**Files:**
- Modify: `docs/architecture/intimacy.md`
- Modify: `docs/STATUS.md`
- Modify: `docs/release/MONT_BLANC_FEATURE_MATRIX.md`
- Modify: `README.md`
- Modify: `Makefile`

**Interfaces:**
- Consumes: tested store/API from Tasks 1–2.
- Produces: an accurate PARTIAL status and repeatable feature test target.

- [ ] Document routes, declared-only provenance, encryption boundary, consent and deletion behavior, and remaining inference/UI/model gaps.
- [ ] Include `services/intimacy` tests in `make test` and `make check`.
- [ ] Run `make test`, `make check`, `make build`, `go test -race -count=1 ./services/policy ./services/core/... ./services/intimacy/...`, and `git diff --check`.
- [ ] Review the staged diff, commit the feature branch, switch to `mont-blanc`, fast-forward merge the feature, rerun `make test` and `make check`, then leave `mont-blanc` clean.

## Spec coverage and limits

This feature implements only the explicit encrypted preference storage/correction slice. Automatic preference inference, preference-history UX, relationship memory, local conversation, adult media indexing/recommendation, consent proof/age verification, BLIND integration and ComfyUI generation are deliberately outside this feature and remain unimplemented.
