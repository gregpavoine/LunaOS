# LunaOS implementation status — Omarchy-aligned architecture checkpoint

**LunaOS V0: INCOMPLETE. Mont Blanc 1.0: NOT RELEASED.** On 2026-10-25 the owner
changed the target implementation to follow Omarchy's Arch package, Bash,
Hyprland and Quickshell/QML techniques, with Go for long-running Luna services.
This direction is documented in ADR 0004. The Go Core/Broker migration has a
tested portable first slice; Quickshell/QML migration has not started. No Arch
image has been produced or booted and there is no verified Linux/hardware run.
Specifications, schema-only contracts and service templates are not counted as
completed OS capabilities.

## Implemented and verified locally

- Rust workspace in `rust/`: `luna-contracts`, `luna-policy-broker`,
  `luna-agent-runner`, `luna-identity`, `luna-goal-engine`, `luna-core`,
  `luna-llm-adapter`, `luna-shell`.
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
  recovery after engine restart, stale-token rejection, authenticated lease renewal
  with a persisted heartbeat timestamp, and a default limit of three expired leases.
  Schema migrations now preserve records through v6. This covers the local pure
  action only; independent heartbeat processes, configurable retry policy, concurrent
  service processes and power-loss recovery remain unverified.
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
- Owners can retrieve verified artifact evidence over a dedicated authenticated
  route; unknown goals or goals without evidence return 404. Only checksum fixture
  evidence exists in this slice.
- Native Agent Runner interface is a separate Rust crate. Its only implementation
  computes verifiable evidence for a fixed public fixture; it has no shell, arbitrary
  filesystem, process or network operation. It remains linked in-process and is not
  a sandboxed worker service.
- Native owner credential provisioning CLI creates a random private token file
  without overwriting an existing credential; a `--reveal-token` flag is required
  to print the secret. Temporary-path creation, file mode and duplicate refusal
  are tested. No Linux first-boot flow or credential rotation/revocation UX exists.
- Rust loopback-only llama.cpp adapter source. One test rejects external endpoint
  configuration; no inference is claimed.
- systemd service unit sources for Core and Policy Broker. These files are
  deployment templates and have not run under systemd. The PKGBUILD and archiso
  staging script now include Core, Broker, owner setup, sysusers configuration and
  the GTK shell, but have not been built or executed. The owner setup command is
  manual only; installer/recovery flow is absent, so services are not enabled by
  the image.
- The prior Go/Python Task 1 implementation remains frozen for behavioral reference.
  The native Make targets no longer require Go or Python.
- Architecture direction changed after these verified Rust results: the Rust
  workspace is now frozen and must not be expanded. Go service parity and the
  Quickshell/QML desktop are pending; no Rust runtime removal has occurred.
- Go Policy Broker checks Linux Unix-socket peer UID at runtime and fails closed
  on non-Linux. Handler tests pass on macOS, including UID comparison. The Linux
  `SO_PEERCRED` path, runtime socket ACLs and child-process IPC have not been run.
- Go Goal Engine/Core in `services/core` implements SQLite WAL/FULL storage,
  authenticated goal create/read/list, policy authorization before weighted score,
  one benign checksum action, idempotency, unknown-source hypothesis provenance,
  durable evidence, queued-work recovery after Engine reopen, and idle queue success.
  Unit and HTTP handler tests pass. The handler tests inject a policy test double;
  no live Linux broker process has been connected to Core.
- `services/core/cmd/luna-core` is a runnable process with a private owner token,
  SQLite database, broker Unix socket, loopback REST API and two-second scheduler.
  Native macOS binaries build; no process-level smoke has yet been run.
- Intimacy control foundation is implemented in Go Core: `/v1/intimacy/stop`
  persists STOP_INTIMACY in SQLite, `/v1/intimacy/relationship/state` reports only
  privacy-safe control metadata, and `/v1/intimacy/conversation` blocks while STOP
  is active or privacy is unknown. Tests verify owner-only stop, blocked conversation
  and stop persistence after reopening the database. First boot defaults to stopped
  and privacy `unknown`; the stop controller itself accepts and stores no intimate
  payload. A separate optional vault handles explicitly consented owner memory.
- A separate SQLite intimate vault is now implemented in `services/intimacy` and
  wired into Core when both `LUNA_INTIMACY_DATABASE` and
  `LUNA_INTIMACY_KEY_FILE` are set. AES-256-GCM encrypts memory payloads with row
  metadata as associated data; the DB and directory must be owner-private. Owner
  consent is explicit, scoped to memory read/write, expires, and can be revoked;
  reads/writes fail after revocation while deletion stays available. API consent
  records require an owner confirmation and adult self-attestation; this is not
  independent age or identity verification. Tests verify encrypted SQLite/WAL
  contents, scope enforcement, cross-subject denial, tamper detection, expiry,
  revocation and deletion. This store has no UI/key provisioning, recovery-key
  workflow, third-party consent, automatic retention/derivative erasure, or
  process-isolated service identity yet.
- The declared-preferences feature is implemented on branch
  `feature/intimacy-preferences`: the separate vault stores explicit owner values
  encrypted with AES-256-GCM, validates a fixed dimension/timescale schema, records
  only `declared` provenance, separates read/write consent, supports correction
  counters, and allows owner deletion after revocation. Core exposes owner-token
  REST endpoints for create/list/correction/delete; write responses omit values.
  Unit and HTTP tests pass locally. This is private preference storage only; it
  does not implement preference inference, NSFW conversation, safety evaluation,
  image/video discovery or generation.
- A fail-closed Discovery metadata policy is implemented in the same encrypted
  vault/Core boundary: source, creator, category, model family and every tag must
  be explicitly allowlisted; policy update, read and evaluation use separate
  expiring consent scopes. The default is disabled, policy writes require owner
  confirmation, and evaluation stores no candidate or media. Go service and HTTP
  tests verify allow/deny, scope enforcement, revocation, explicit updates and
  encrypted policy data. This is policy enforcement only; no Discovery connector,
  media fetch, age/consent classifier or content safety model exists.
- The private vault now provides an owner-confirmed, atomic in-library key
  rotation primitive. It re-encrypts memory, declared preferences and Discovery
  policy in one SQLite transaction; tests verify reopen with the new key, rejection
  of old-key decryption, and rollback when one stored ciphertext is invalid. No
  key file replacement, service quiescence/lock, or lost-key recovery is
  implemented, so operational key lifecycle remains incomplete.
- The vault now supports encrypted backup and restore as library operations:
  SQLite produces a consistent snapshot; chunked AES-256-GCM uses a separate
  backup key; restore verifies database integrity and every encrypted payload and
  refuses to overwrite an existing destination. Tests cover round-trip, chunk
  boundaries, wrong keys, tampering, plaintext absence and no-overwrite behavior.
  The `luna-vault` CLI exposes keygen, backup and restore and is included in local
  builds and the Arch package recipe. OS key-store integration, key-file
  replacement, Linux service coordination and lost-key recovery remain
  absent/unverified.
- No NSFW conversation model, preference learner, consent/subject
  identity workflow, adult media index/classifier/recommender, ComfyUI generation,
  GPU controls or global UI/voice stop binding exists. The local model backend is
  unavailable and the API reports that honestly.
- systemd templates, Arch PKGBUILD and archiso staging now refer to Go Core/Broker.
  They have not been built by makepkg, run on Arch or booted. The image does not
  yet include a functional Luna shell or Go owner-credential provisioning.

## Frozen Rust checkpoint verification, 2026-10-24

Host: macOS `Darwin 27.0.0 arm64`; local test results do not count as Linux or OS
installation evidence.

Detailed command record: [native-migration-2026-10-24.txt](test-evidence/native-migration-2026-10-24.txt).

| Command | Result |
|---|---|
| `make test` | PASS: 29 Rust unit tests plus 4 integration tests (Agent Runner, Identity and child-process Broker); 0 failures |
| `make check` | PASS: `cargo fmt --check`, Clippy across workspace/all targets with warnings denied, and `cargo check` |
| `make native-smoke` | PASS: release Core and broker binaries; live loopback REST, unauthorized 401, owner create, duplicate idempotency, checksum evidence, then Core process restart with completed state retained |
| `make test-reference check-reference` at that checkpoint | PASS: existing Go policy tests/vet, 16 Python reference integration tests/compileall, Linux baseline script syntax only |

## Omarchy-aligned command verification, 2026-10-25

Host: same macOS Apple Silicon checkout. These checks exercise Go Core/Broker
logic and native macOS builds; they do not establish Linux peer-credential behavior
or OS integration.

| Command | Result |
|---|---|
| `make test` | PASS: Go Broker/Core/Intimacy vault, REST, backup CLI and restore tests; 0 failures |
| `make check` | PASS: gofmt, go vet, Bash syntax (including Arch PKGBUILD) and OpenAPI JSON validation |
| `make build` | PASS: native macOS arm64 Core, Broker and `luna-vault` binaries built with CGO enabled |
| `luna-vault` process smoke | PASS: generated two owner-only keys, backed up an empty vault and restored it to a new path in a temporary private directory; both key files were mode 0600 |
| `make build-linux-amd64-v3` | PASS: optional Go policy prototype cross-compiled for Linux x86_64-v3; not run on target hardware |
| `go test -race -count=1 ./services/policy` | PASS: race-enabled Go policy tests; 0 failures |
| `go test -race -count=1 ./services/policy ./services/core/... ./services/intimacy/...` | PASS: race-enabled service tests; 0 failures |
| `make test-reference check-reference` | PASS: Go reference tests/vet and Python compile/syntax checks; 16 frozen Python process tests SKIPPED because this host is macOS and the Linux broker requires `SO_PEERCRED` |
| `make test-rust-checkpoint` | PASS: Rust workspace unit, integration, process-boundary and doc tests; 0 failures on final run |
| `make check-rust-checkpoint` | PASS: Rust formatting, Clippy with warnings denied, and workspace check |
| `git diff --check` | PASS |

The first Go build attempt failed because Go invoked an incompatible Intel Git at
`/usr/local/bin/git` while stamping VCS metadata. The Make target now passes
`-buildvcs=false`; the retry passed. This changes only build metadata generation.

The Rust smoke test at the earlier checkpoint used a temporary database and test
credential and removed them on exit. No Go Core/Broker process smoke or hosted CI
run is being counted. The vault CLI smoke also used only a temporary directory.
Intimacy vault and API tests use temporary databases and
synthetic text only; tests verify that writes omit plaintext, encrypted backups
round-trip into new database paths, and corrupted restores do not publish a DB.

## Not implemented or blocked

- No real Omarchy or Linux installation was accessed. No packages/services were
  removed, disabled, retained or replaced on a host; **no cleanup occurred**. The
  read-only Linux baseline script remains unexecuted.
- No Arch package has been built, and no standalone ISO, installer, boot, update,
  rollback or recovery result exists. The archiso script is syntax-checked only;
  `mkarchiso` is unavailable on this Mac.
- No runtime performance benchmarks have been recorded. The broker's prior Go
  cross-builds prove compilation for x86_64 baselines; the v3 artifact has not run
  on a feature-checked Ryzen 9 5950X. The CGO Core has not been cross-compiled for
  Linux. No service startup, boot-to-shell, idle resource,
  scheduler/API latency or throughput claim is made.
- GTK4 shell source is present but has not been built or run. `pkg-config` fails on
  this host with `bad CPU type in executable`; no Wayland/Hyprland target was used.
- No llama.cpp server or licensed model is configured; no real local inference.
- No GPU/NVIDIA/CUDA, ComfyUI, systemd, D-Bus, sensor/BLIND, iPhone, NAS, privacy
  vault, consent/erasure or remote-administration capability has been exercised.
- Full goal lifecycle beyond list/pause/resume/cancel and CRUD, natural-language plan generation and strategy scoring,
independent heartbeat workers/configurable retry policy, external event delivery and resource quotas, sandboxed Agent Runner service,
  strategy optimizer, production identity/session
  management, encrypted memory/evidence vault, and broad Mont Blanc modules remain
  incomplete. Several contracts are types only.
- The current authorization credential is a local bearer secret shared with the
  broker. The broker is a separate process, but both local service units use the
  same unprivileged account. Production provisioning, independent service identity,
  owner login, revocation UX and Linux security review remain release blockers.
- Owner-token setup exists only as a local Rust checkpoint command, but it is not
  installed in the Go image path. Target provisioning and credential
  rotation/revocation workflows remain untested and incomplete.

## Next milestone

Run process-level Core/Broker tests in an isolated Arch VM, including owner auth,
peer UID, payment denial, grant revalidation, queued recovery and systemd startup.
Then add Go credential provisioning and a functional Quickshell/QML shell using the
loopback API. No real Linux installation has been accessed or changed; a remote
baseline still requires an owner-provided SSH alias and explicit connection
authorization, and any mutation requires a verified recovery plan and explicit
authorization for that mutation.
