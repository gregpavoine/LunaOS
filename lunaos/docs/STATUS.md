# LunaOS implementation status — native migration checkpoint

**LunaOS V0: INCOMPLETE. Mont Blanc 1.0: NOT RELEASED.** This checkout has a
portable Rust control-plane slice and separate Rust Policy Broker process. It has
not produced or booted an Arch image and has no verified Linux/hardware run.
Specifications, schema-only contracts and service templates are not counted as
completed OS capabilities.

## Implemented and verified locally

- Rust workspace in `rust/`: `luna-contracts`, `luna-policy-broker`,
  `luna-goal-engine`, `luna-core`, `luna-llm-adapter`, `luna-shell`.
- Separate Rust Policy Broker binary over an owner-only Unix socket. It checks the
  owner credential, socket peer UID, deny-by-default action list, payment budget,
  and goal/action-bound expiring HMAC grants. Only `local.checksum` is allowed.
- Rust/SQLite Goal Engine and Core API slice: owner-authenticated create/read/tick,
  deterministic score after authorization, idempotency, durable queued state,
  process-reopen resume, fixed public fixture checksum, independent digest/length
  evidence, unknown source status and idle success. Owners can add dependency edges;
  the engine rejects cycles, waits for prerequisites and distinguishes queued
  dependency waits from idle success.
- Durable SQLite job leases: atomic worker claims, unique lease tokens, expiry
  recovery after engine restart, stale-token rejection and a v1-to-v2 data-preserving
  migration. This is verified for one local pure action; heartbeats, bounded retries,
  multiple service processes and power-loss recovery remain unverified.
- Goal lifecycle subset in Rust: owner-authenticated pause, resume and cancel routes;
  pause revokes an active lease, resume requeues work, cancellation survives restart,
  and completed/cancelled states reject further transitions. SQLite v3 migration
  broadens state constraints while preserving goals, grants, dependencies and jobs.
- Owner-authenticated goal listing returns at most 100 goals ordered by score and
  creation time. Editing and full CRUD semantics are still absent.
- Owner-authenticated goal plan GET/PUT replaces hierarchy and prerequisite edges
  atomically. The planner caps direct prerequisites at 32 and parent depth at 16,
  rejects duplicate/missing/unrunnable edges and cycles across the combined graph,
  and locks plans once a goal is running or terminal. It does not yet generate
  plans from natural language or provide strategy evaluation.
- Goal events are append-only, privacy-safe decision codes. Event creation inserts
  a durable outbox record transactionally; Core dispatches idempotently into a local
  inbox, and owners can read the last 100 events per goal. No external bus is wired.
- Rust loopback-only llama.cpp adapter source. One test rejects external endpoint
  configuration; no inference is claimed.
- systemd service unit sources for Core and Policy Broker. These files are
  deployment templates and have not run under systemd. The PKGBUILD and archiso
  staging script now include both native services and the GTK shell, but have not
  been built or executed. Owner-token first-boot provisioning and installer/recovery
  flow are absent, so services are not enabled by the image.
- The prior Go/Python Task 1 implementation remains frozen for behavioral reference.
  The native Make targets no longer require Go or Python.

## Commands executed on this Mac checkout

Host: macOS `Darwin 27.0.0 arm64`; local test results do not count as Linux or OS
installation evidence.

Detailed command record: [native-migration-2026-10-24.txt](test-evidence/native-migration-2026-10-24.txt).

| Command | Result |
|---|---|
| `make test` | PASS: 26 Rust unit tests plus 1 child-process Unix-socket broker integration test; 0 failures |
| `make check` | PASS: `cargo fmt --check`, Clippy across workspace/all targets with warnings denied, and `cargo check` |
| `make native-smoke` | PASS: release Core and broker binaries; live loopback REST, unauthorized 401, owner create, duplicate idempotency, checksum evidence, then Core process restart with completed state retained |
| `make test-reference check-reference` | PASS: existing Go policy tests/vet, 25 Python reference integration tests/compileall, Linux baseline script syntax only |

The smoke test used a temporary database and test credential and removed them on
exit. Evidence is a live local process test, not CI-only evidence. No hosted CI run
is being counted.

## Not implemented or blocked

- No real Omarchy or Linux installation was accessed. No packages/services were
  removed, disabled, retained or replaced on a host; **no cleanup occurred**. The
  read-only Linux baseline script remains unexecuted.
- No Arch package has been built, and no standalone ISO, installer, boot, update,
  rollback or recovery result exists. The archiso script is syntax-checked only;
  `mkarchiso` is unavailable on this Mac.
- GTK4 shell source is present but has not been built or run. `pkg-config` fails on
  this host with `bad CPU type in executable`; no Wayland/Hyprland target was used.
- No llama.cpp server or licensed model is configured; no real local inference.
- No GPU/NVIDIA/CUDA, ComfyUI, systemd, D-Bus, sensor/BLIND, iPhone, NAS, privacy
  vault, consent/erasure or remote-administration capability has been exercised.
- Full goal lifecycle beyond list/pause/resume/cancel and CRUD, natural-language plan generation and strategy scoring,
lease renewal/heartbeats/retry policy, external event delivery and resource quotas, standalone Agent Runner,
  strategy optimizer, production identity/session
  management, encrypted memory/evidence vault, and broad Mont Blanc modules remain
  incomplete. Several contracts are types only.
- The current authorization credential is a local bearer secret shared with the
  broker. The broker is a separate process, but both local service units use the
  same unprivileged account. Production provisioning, independent service identity,
  owner login, revocation UX and Linux security review remain release blockers.
- No owner-token bootstrap or interactive installer exists; the current service
  units require an administrator to provision `/etc/lunaos/owner-token` manually.

## Next milestone

Complete owner credential lifecycle and full goal CRUD/state semantics.
Then add lease renewal/heartbeats, bounded retry policy and strategy evaluation
before connecting any Linux action adapter. The next target-side milestone
requires an owner-provided SSH alias and explicit connection/baseline authorization;
no current Linux installation may be modified without a verified recovery plan and
explicit authorization for each mutation.
