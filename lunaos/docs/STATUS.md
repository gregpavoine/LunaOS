# Task 1 status — 2026-10-24

**Working local bootstrap checkpoint. Full LunaOS V0 is NOT finished. Mont Blanc
release remains PLANNED. No target Linux integration has been verified.**

## Implemented

- Requirement/contradiction/open-decision map, architectural service boundaries,
  incremental V0 plan, OS/stack ADRs, threat model and release feature matrix.
- Separate Go owner-authenticated deny-default policy broker, private Unix socket,
  signed action/goal-specific expiring grants. Guest/voice strings do not authenticate.
- Python/SQLite persistent registry, versioned migration, queue, weighted scoring
  after policy gate, execution recheck and durable idempotency.
- Real `local.checksum` action on a public fixture, independent digest/length
  evaluation, evidence, measured strategy trial and transactional completion event.
- Durable lifecycle outbox records and one local transactional completion consumer
  receipt. General asynchronous event delivery is not implemented.
- Owner-authenticated local REST API, native client, background worker and isolated
  no-pip venv instructions. Linux operations remain an interface only.
- Read-only Linux baseline script authored. **Not executed**; syntax-check only.

## Verified locally (real executed results)

Evidence: [task-1.txt](test-evidence/task-1.txt).

| Command | Actual result |
|---|---|
| `make env` | PASS; local venv created without pip/download |
| `make test check` | PASS; 3 Go test functions + 16 Python integration tests; vet, compileall and baseline syntax pass |
| `go test -buildvcs=false -race -count=1 ./services/policy` | PASS |
| `.venv/bin/python scripts/smoke.py` | PASS; real launcher stop/restart, same persisted goal and exactly one verified digest |
| Git diff whitespace check | PASS |
| OpenAPI JSON parse | PASS; full OpenAPI schema-validator conformance not executed |
| Independent read-only review | Two privacy findings fixed; follow-up reviewer confirmed both resolved and found no new actionable issues |

The six requested acceptance cases all pass: owner-only privileges, absent-budget
payment refusal, policy denial before scoring, duplicate-job idempotency,
unknown-source hypothesis, and inactive/zero-utility idle success. Additional tests
cover score ordering, real UDS API auth, broker outage, grant forgery/scope/expiry,
sensitive-denial canary, immutable events, concurrent workers, new-process queue
resume and SIGKILL during execution followed by successful retry.

Checksum fixture: `LunaOS public checkpoint v1` plus newline (28 bytes).
Verified SHA256: `07be6a1cc7b1b323fd3baca70abf34bfea13e09f9d28cf417d11aca77d4220ce`.
This is a real checksum execution, not evidence of a model or hardware action.

## Failures encountered and resolved

- Default Git path encountered Xcode licence gate; alternate `/usr/local/bin/git`
  was incompatible (`bad CPU type in executable`). Used existing Command Line
  Tools Git; did not accept licence, install tools or change system configuration.
- Go VCS stamping initially selected that incompatible Git, failing build. Added
  explicit `-buildvcs=false`; build/tests pass. Commit remains actual provenance.
- Test-first baseline lacked implementation modules, then passed after implementation.
- Review regressions reproduced sensitive-denied assertion persistence and duplicate
  denial log growth. Sensitive assertions now discarded before hashing/persistence;
  identical consecutive denials deduplicated. Both regression tests pass.

## Blocked / not verified / limits

- Linux: no SSH alias/authorization. No remote connection, baseline execution,
  cleanup, packages, mounts, service mutation, boot or disk operation performed.
- Hardware: GPU/NVIDIA/CUDA, ComfyUI image/video, Omarchy shell/KEEP tests, actual
  reboot/power-loss, sensors/BLIND, iPhone/WhatsApp, NAS and recovery NOT VERIFIED.
- Model router, full planner DAG/CRUD, strategy optimizer, production identity/MFA,
  grant revocation registry, resource leases, encrypted private vault, consent,
  full Intimacy Core, Discovery and signed images NOT IMPLEMENTED in this task.
- Owner key is private development file, not production credential vault. Same-UID
  hostile process is outside assurance; no privileged/untrusted tools enabled.
- DB supports non-sensitive public metadata only; no at-rest encryption or DLP claim.
- Assertion claim/source input is transient and not stored; provenance metadata alone cannot verify user claims.
- Pure action recovery uses serialized SQL transactions, not distributed exactly-once
  external effects. No payment/message adapter or configured financial budget.
- Hosted CI authored, NOT EXECUTED. Linux portability inferred from standard-library
  interfaces, not claimed as tested. Full spec acceptance suites remain open.

## Next smallest milestone

Complete owner session expiry/revocation and separate production service identities,
then goal DAG/planner + migration tests and protected private import vault. In the
Linux lane, the first dependency is owner-supplied SSH alias **and explicit connection
and baseline authorization**; then read-only baseline and recovery verification.
No data is needed from the owner to use this local checkpoint.
