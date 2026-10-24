# Task 1 architecture and V0 boundaries

This is a local goal-engine bootstrap, not the LunaOS V0 exit scenario.

## Current executable path

Native client → owner-authenticated Python Core UDS API → Go broker UDS →
SQLite registry/outbox → weighted serial scheduler → policy recheck → public
fixture checksum executor → separate exact-digest/length evaluator → atomic
result/evidence/strategy measurement/event commit. Both sockets are 0600 inside
0700 owner directories. No TCP listener, SSH, root action, shell tool or sensor.

| Boundary | Input / ownership | Current implementation | Later V0 requirement |
|---|---|---|---|
| Core API | Owner key possession; strict JSON; caller cannot supply role | `luna/api.py` | OS-authenticated shell/device session, revocable credentials |
| Policy broker | Separate Go process; signed goal/action/owner/expiry grant | `services/policy` | Separate UID, peer credentials, OS-protected keys, owner grant UI |
| Goal registry | Canonical request, metrics and provenance | `luna/engine.py`, SQLite | Full goal DAG/planner, policies, CRUD lifecycle |
| Scorer | Hard policy gates first, normalized 0..1 metrics | Pure weighted formula 0..100; explanation stored | Aging bonus capped at 5, project fairness, deadlines/resource factors benchmarked |
| Scheduler | Approved queued jobs with utility >0 | One write transaction serializes workers; score DESC, oldest first | Leases/heartbeats, quotas, modes, fairness and resource admission |
| Agent executor | Single allowlisted typed action; no arguments | SHA256 of public fixture bytes | OS-sandboxed agent tools; scoped files/service/ComfyUI adapters |
| Outcome evaluator | Actual digest + byte count | Independent check before completion | Artifact decode, real tests/benchmarks; no LLM completion claim |
| Strategy evaluator | Baseline hypothesis + measured result | Immutable trial/result comparison for one strategy | Bounded experimental replanning/rollback; not self-modification |
| Evidence/events | Opaque IDs, generic codes, public digest | Append-only DB event triggers, durable outbox, local completion inbox | Protected audit store and asynchronous dispatcher with inbox dedup |
| Linux adapter | Typed baseline/service/workflow interface | Protocol only, unavailable capability response | Approved SSH inventory first, real target implementation |
| Sensitive vault/perception | No supported input path | Not implemented; sensitive declared goals denied | Independent sensor gates, encrypted compartments, consent/revocation |

Default scoring weights: .30 impact, .20 urgency, .20 user_value, .15 confidence,
.15 efficiency. Defaults are engineering choices, not Series28 approvals. Missing
metrics default to 1 for an explicit owner-created checksum request. No automatic
unrequested goal creation. Zero-utility jobs remain queued and return idle success.
Aging/fairness and full resource modes above are planned contracts, not implemented.

## Authentication and authorization

`Authorization: Bearer <32 random bytes hex>` is read from a private regular key
file. Core checks it without trusting caller labels; broker independently checks
it and returns a signed capability scoped to a particular goal and action for one
hour. The scheduler rechecks that capability with the live broker before execution.
Unknown action, unknown privacy scope, missing/forged/expired grant or broker outage
never yields execution. Payments always return BUDGET_UNSET; no ledger or payment
adapter exists. Owner key is never stored in the DB or returned by API. Signed
grant is operational metadata and omitted from responses; it is not sufficient
without authenticated broker access. Stop the broker to cut new execution.

This development identity boundary is intentionally limited to nonprivileged pure
actions. It does not establish production MFA, OS isolation from same-UID malware,
root authorization or protection from a compromised trusted Core. No LLM or untrusted
agent is given the key. Production owner session and separate service identity are
required before unsafe tools; core and runner must not share a grant-issuing key.

## Persistence, state and crash behavior

Migration 001 uses FK/unique/check constraints, UUID IDs, UTC timestamps,
FULL synchronous SQLite rollback journal and BEGIN IMMEDIATE. Created goal,
authorization decision and queue entry commit together; duplicate idempotency key
with same canonical payload returns existing goal, changed payload is a conflict.
One job per goal; database uniqueness plus serialization prevents duplicate completion.

Execution traverses queued → running → verifying → succeeded/failed inside a single
write transaction. Kill before commit rolls back to queued; commit stores result,
evidence, trial, completion event and one idempotent local completion receipt.
This is re-execution of a pure deterministic action, not continuation of an external
operation. Never carry this transaction-across-work pattern to long or network jobs.
Outbox records for other lifecycle events remain durable and pending; no general
message bus or dispatcher is claimed. Goal reads expose event history directly.
Denied jobs are blocked at admission; an unavailable/expired broker at execution
leaves queued work intact. No automatic grant renewal. Unchanged repeated denials are deduplicated; bounded retention/backoff is required
before production long-running use. Only the latest 100 events are exposed, in chronological order.

## Logical data schema vs implemented subset

Migration is the exact executable schema. Goal Engine §3 remains the full logical
model: goal (owner/project/domain/source/title/description/status/metrics/cost/score/
explanation/evidence/privacy/timestamps/version/parent), goal_dependency (DAG edges),
goal_policy (scope/runtime/CPU/VRAM/network/spend/expiry/tools/grant), goal_execution
(agent/attempt/state/lease/idempotency/artifact/outcome), goal_event (append-only),
strategy_trial (hypothesis/tests/measurements/baseline), policy_grant (principal,
auth-strength/action/targets/data/recipients/money/currency/validity/revocation/consent),
expense_ledger (intent/reservation/settlement/reconciliation), memory_assertion
(claim/source/status/confidence/supersession/privacy/freshness), persona_revision,
external_conversation_grant, consent_receipt and data_erasure_job.

Task 1 implements only public-fixture goal fields, one signed scoped grant, executions,
evidence, events/outbox/inbox, a single measured strategy and optional unknown-source
assertion. The remaining entities are specified above and in source §3, not empty
production tables or APIs masquerading as implemented services. Full planner DAG,
ledger reservations, persona, external communication and consent are not implemented.
SQLite→PostgreSQL migration path is ADR0002; concurrency/external execution needs
leases, at-least-once outbox delivery, scoped inbox keys and provider-side reconciliation.

## REST and event contracts

`contracts/openapi/local.json` describes only currently implemented UDS HTTP endpoints.
Create accepts action, optional metrics, privacy class and an optional assertion input. Claim/source text is validated in memory and discarded before hashing or persistence; only confidence and unknown/hypothesis status are retained for the provenance test;
unknown fields rejected. Idempotency-Key required; owner derived from authenticated
configuration. Errors use fixed codes, never echo raw bodies/secrets. Origin-bearing
browser requests are rejected; native UDS client only. Results bounded to 100 records.

Planned Goal §8 routes (PATCH, plan, evaluate, remote/session, consent, persona) remain
unimplemented (404). Event payloads have UUID event_id/goal_id/correlation_id, UTC
occurred_at, schema_version=1, event_type and decision_code. No prompt/body/credential
in event payload. Exact HTTP broker grant/evaluation contracts are in OpenAPI.
