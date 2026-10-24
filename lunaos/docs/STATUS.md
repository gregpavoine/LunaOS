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
- Rust loopback-only llama.cpp adapter source. One test rejects external endpoint
  configuration; no inference is claimed.
- systemd service unit sources for Core and Policy Broker. These files are
  deployment templates and have not run under systemd.
- The prior Go/Python Task 1 implementation remains frozen for behavioral reference.
  The native Make targets no longer require Go or Python.

## Commands executed on this Mac checkout

Host: macOS `Darwin 27.0.0 arm64`; local test results do not count as Linux or OS
installation evidence.

Detailed command record: [native-migration-2026-10-24.txt](test-evidence/native-migration-2026-10-24.txt).

| Command | Result |
|---|---|
| `make test` | PASS: 14 Rust unit tests plus 1 child-process Unix-socket broker integration test; 0 failures |
| `make check` | PASS: `cargo fmt --check`, Clippy across workspace/all targets with warnings denied, and `cargo check` |
| `make smoke-native` | PASS: release Core and broker binaries; live loopback REST, unauthorized 401, owner create, duplicate idempotency, checksum evidence, then Core process restart with completed state retained |
| `make test-reference check-reference` | PASS: existing Go policy tests/vet, 25 Python reference integration tests/compileall, Linux baseline script syntax only |

The smoke test used a temporary database and test credential and removed them on
exit. Evidence is a live local process test, not CI-only evidence. No hosted CI run
is being counted.

## Not implemented or blocked

- No real Omarchy or Linux installation was accessed. No packages/services were
  removed, disabled, retained or replaced on a host; **no cleanup occurred**. The
  read-only Linux baseline script remains unexecuted.
- No standalone ISO, installer, Arch package, boot, update, rollback or recovery
  result exists. `mkarchiso` is unavailable on this Mac.
- GTK4 shell source is present but has not been built or run. `pkg-config` fails on
  this host with `bad CPU type in executable`; no Wayland/Hyprland target was used.
- No llama.cpp server or licensed model is configured; no real local inference.
- No GPU/NVIDIA/CUDA, ComfyUI, systemd, D-Bus, sensor/BLIND, iPhone, NAS, privacy
  vault, consent/erasure or remote-administration capability has been exercised.
- Full goal lifecycle and CRUD, dependency DAG/planner, durable leases/resource
  quotas, standalone Agent Runner, strategy optimizer, production identity/session
  management, encrypted memory/evidence vault, and broad Mont Blanc modules remain
  incomplete. Several contracts are types only.
- The current authorization credential is a local bearer secret shared with the
  broker. The broker is a separate process, but both local service units use the
  same unprivileged account. Production provisioning, independent service identity,
  owner login, revocation UX and Linux security review remain release blockers.

## Next milestone

Complete owner credential lifecycle and a full Rust goal state model/DAG with
migration and process-restart tests. Then add durable execution leases and evidence
events before connecting any Linux action adapter. The next target-side milestone
requires an owner-provided SSH alias and explicit connection/baseline authorization;
no current Linux installation may be modified without a verified recovery plan and
explicit authorization for each mutation.
