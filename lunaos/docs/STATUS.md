# LunaOS status — native migration bootstrap — 2026-10-24

**Mont Blanc release: BLOCKED / NOT RELEASED. LunaOS V0: NOT COMPLETE.** The native Rust migration has a locally tested control-plane slice, but no bootable LunaOS installation, production-isolated broker, real LLM inference, Linux/Wayland runtime or target-hardware evidence has been verified.

## Implemented in source

- Owner-directed architecture correction documented: independent Arch distribution; Hyprland/Wayland; Rust/Tokio; systemd/D-Bus/Unix socket boundaries; GTK4 selection; llama.cpp adapter; SQLite initial store. Prior overlay and Go/Python stack ADRs are marked superseded.
- Rust workspace and lockfile with contracts, policy broker crate, goal engine, Tokio/Axum Core, local-only llama.cpp HTTP adapter and GTK4 shell source.
- Durable SQLite goal/job/evidence/event schema migration; WAL/FULL sync; database file mode 0600; deterministic weighted score; idempotency; goal/action-bound one-hour HMAC grant; deny-before-score; fixed public fixture checksum and independent constant digest/length verification; restart queue resume; idle success.
- Owner-only bearer-gated REST routes for goal create/read/tick; background scheduler loop; llama.cpp endpoint restricted to HTTP loopback; capability route reports absent target/model/image blockers.
- Rust logical contracts for goals, dependencies, executions, policy grants, events, strategy trials, memory assertions, persona revisions, expense intents, external conversation grants, consent receipts and erasure jobs. Most are schema-only, not live database/services.
- systemd Core unit, Arch PKGBUILD template, archiso profile and build script authored. These are preparation artifacts, not a package or image result.
- Existing Go/Python prototype preserved as a frozen behavior/API/test reference. New control-plane work is Rust.
- Requirement conflicts, release constraints, risks, and unanswered owner decisions documented in `OPEN_DECISIONS.md`; Linux baseline script remains unexecuted.

## Verified locally

Commands executed on macOS Apple Silicon in this repository:

| Command | Result |
|---|---|
| `make test check` | PASS: Rust workspace tests, `cargo fmt --check` and workspace `cargo check` |
| `cargo test --workspace --locked` | PASS: 10 Rust unit tests; policy owner-only, absent-budget payment refusal, deny-before-score, duplicate idempotency, unknown provenance, idle success, fixture evidence and process-reopen queue resume are covered |
| `cargo build --locked --release -p luna-core` | PASS: optimized Core binary built locally |
| `scripts/smoke-native-core.sh` | PASS: unauthenticated request denied; owner goal created; duplicate key idempotent; checksum action completed; state remained readable after Core restart |
| `make test-reference check-reference` | PASS for frozen reference: Go tests/vet, 25 Python integration tests/compileall, Linux baseline shell syntax only |
| Native OpenAPI JSON parse | PASS |
| Archiso script and PKGBUILD `bash -n` | PASS (syntax only) |
| OpenAPI JSON parse and Archiso/profile/PKGBUILD/smoke script syntax | PASS at local syntax-check time; no Arch behavior asserted |
| Git whitespace check | PASS on the final pre-commit diff |

Evidence details: [`native-bootstrap-2026-10-24.txt`](test-evidence/native-bootstrap-2026-10-24.txt). This evidence is local-only. The old prototype tests validate reference behavior; they do not make Python a LunaOS runtime dependency. No target Linux command, baseline, SSH connection, GTK build, local model inference, Arch package build, ISO build, installer, target reboot, NVIDIA/ComfyUI workflow, privacy sensor test or restore test was executed.

## Partially implemented / blocked

- **Policy boundary:** real Rust authorization logic and signed scoped grants are tested in the same process as Core. Separate service/UID, Unix peer credentials, credential provisioning, grant revocation UI and production key protection are not present. No privileged host action is enabled.
- **Goal engine:** one-action schema and scheduler slice only. Full goal lifecycle/CRUD, bounded planner/DAG, durable leases, resource admission, complete events/outbox/inbox, migration parity and multi-process crash recovery remain incomplete.
- **LLM:** real llama.cpp HTTP adapter source accepts loopback only; no server/model was configured, licensed, launched or measured. LLM is not integrated into planning or action authority.
- **Shell:** GTK4 source reads live capability status; GTK build could not be attempted because the Mac host's `/usr/local/bin/pkg-config` fails to execute with `bad CPU type in executable`. No Arch/Hyprland login, keyboard, accessibility or recovery run.
- **OS packaging:** PKGBUILD/profile/build script exist but are not a completed package pipeline; license metadata/key lifecycle, installer, system configuration, account/credential bootstrap, update/rollback, recovery and signed artifact flow remain unresolved. The archiso script was not executed.
- **Mont Blanc modules:** memory/identity, resource/GPU manager, system and D-Bus adapters, privacy/BLIND/consent/vault, Intimacy, Discovery, iPhone, NAS, remote admin, payments, communications and all named hardware capabilities remain unimplemented or blocked. See the feature matrix.

## Exact next milestone

1. Move policy evaluation to an independent Rust Unix-socket broker; bind requests to OS peer credentials and isolate service identities before adding any privileged adapter.
2. Port remaining Goal Engine behaviors from the frozen reference and implement versioned lifecycle, DAG and event delivery with integration tests.
3. Resolve source/package licensing and credential bootstrap, then build the GTK shell and package on disposable Arch CI/VM infrastructure.
4. Obtain SSH alias plus explicit authorization before running the read-only Linux baseline or doing any target integration.

Do not tag or announce Mont Blanc until the complete source acceptance matrix and release gates pass on the actual target and a signed independent image is installable and recoverable.
