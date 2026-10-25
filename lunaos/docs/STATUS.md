# LunaOS implementation status — Omarchy-aligned architecture checkpoint

**LunaOS V0: INCOMPLETE. Mont Blanc 1.0: NOT RELEASED.** On 2026-10-25 the owner
changed the target implementation to follow Omarchy's Arch package, Bash,
Hyprland and Quickshell/QML techniques, with Go for long-running Luna services.
This direction is documented in ADR 0004. The Go Core/Broker migration has a
tested portable slice and a Linux process-level integration test in Docker
Desktop; Quickshell/QML migration has not started. No Arch image has been
produced or booted and there is no verified Arch or target-hardware run.
Specifications, schema-only contracts and service templates are not counted as
completed OS capabilities.

## Implemented and verified locally

The Rust workspace and GTK experiment are frozen historical checkpoints, not active
LunaOS services or desktop functionality. The current implementation path is Go
services, Bash integration and a not-yet-built Quickshell/QML shell.

- Go Policy Broker checks Linux Unix-socket peer UID at runtime and fails closed
  on non-Linux. Handler tests pass on macOS, including UID comparison. The Linux
  `SO_PEERCRED` path and child-process IPC passed in a disposable Debian
  `golang:1.20-bookworm` container. Arch systemd sandboxing, separate service UIDs
  and target socket lifecycle remain unverified.
- Go Goal Engine/Core in `services/core` implements SQLite WAL/FULL storage,
  authenticated goal create/read/list, policy authorization before weighted score,
  one benign checksum action, idempotency, unknown-source hypothesis provenance,
  durable evidence, queued-work recovery after Engine reopen, and idle queue success.
  Unit and HTTP handler tests pass. A Linux process test connects Core to the real
  broker binary; handler unit tests still use a policy test double.
- `services/core/cmd/luna-core` is a runnable process with a private owner token,
  SQLite database, broker Unix socket, loopback REST API and two-second scheduler.
  A Linux process test starts the real Go broker and Core binaries, loads a queued
  goal from SQLite after Core startup, checks owner authentication and API
  idempotency, then verifies completed checksum evidence. This is a Debian
  container result, not an Arch image or systemd result.
- Go `services/identity` and `luna-owner-setup` create the service credential as
  32 random bytes encoded as 64 lowercase hex characters. The CLI uses the format
  Core and Broker require, creates a new credential directory as `0700`, writes a
  `0600` regular file atomically, refuses existing paths and prints the token only
  when explicitly asked. Unit tests and a local binary smoke passed.
- The x86_64 `lunaos-native` Arch package built and installed in a disposable Arch
  Docker container under x86_64 emulation. Its sysusers hook created the `lunaos`
  account; the packaged credential CLI, Broker and Core then ran as that account
  against a temporary database and completed a checksum goal with verified
  evidence. This did not boot systemd. `systemd-analyze verify` could not validate
  ExecStart paths in this emulated environment: it returned `Invalid argument` for
  the packaged binary and also for `/usr/bin/true`.
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

Host: macOS Apple Silicon checkout plus a disposable Docker Desktop Linux
container (`golang:1.20-bookworm`). Linux process checks below exercise Unix peer
credentials and real Go child processes; they do not establish Arch packaging,
systemd, boot or target-hardware behavior.

| Command | Result |
|---|---|
| `make test` | PASS: Go Broker/Core/Intimacy vault, REST, backup CLI and restore tests; 0 failures |
| `make test test-process-linux` in `golang:1.20-bookworm` | PASS: Go service tests and real Core/Broker process integration; queued recovery, owner-authenticated API, duplicate idempotency and verified evidence |
| `go vet -buildvcs=false ./services/policy ./services/core/... ./services/intimacy/... ./services/identity/...` in `golang:1.20-bookworm` | PASS |
| Arch `makepkg --cleanbuild --noconfirm` in x86_64 Docker container | PASS: built `lunaos-native-0.1.0-1-x86_64.pkg.tar.zst` with Core, Broker, vault and Go owner setup |
| `pacman -U` in disposable Arch container | PASS: package installed; sysusers created `lunaos`; packaged credential CLI and Broker/Core processes ran as that account and completed a goal with verified evidence |
| `systemd-analyze verify` in emulated Arch container | BLOCKED: returned `Invalid argument` for all ExecStart paths, including `/usr/bin/true`; systemd unit validation remains open |
| Go owner setup CLI smoke on macOS | PASS: created 64-hex credential, directory mode `0700`, file mode `0600`, refused duplicate provisioning; secret was not printed |
| `make check` | PASS: gofmt, go vet, Bash syntax (including Arch PKGBUILD) and OpenAPI JSON validation |
| `make build` | PASS: native macOS arm64 Core, Broker, `luna-vault` and `luna-owner-setup` binaries built with CGO enabled |
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
credential and removed them on exit. The Go process test uses only a temporary
directory, synthetic owner key and public checksum fixture; it does not access a
host installation. No hosted CI run is being counted. The vault CLI smoke also
used only a temporary directory.
Intimacy vault and API tests use temporary databases and
synthetic text only; tests verify that writes omit plaintext, encrypted backups
round-trip into new database paths, and corrupted restores do not publish a DB.

## Not implemented or blocked

- No real Omarchy or Linux installation was accessed. No packages/services were
  removed, disabled, retained or replaced on a host; **no cleanup occurred**. The
  read-only Linux baseline script remains unexecuted.
- No standalone ISO, installer, boot, update, rollback or recovery result exists.
  The archiso script is syntax-checked only; `mkarchiso` is unavailable on this Mac.
- No runtime performance benchmarks have been recorded. The broker's prior Go
  cross-builds prove compilation for x86_64 baselines; the v3 artifact has not run
  on a feature-checked Ryzen 9 5950X. Core/Broker/vault/identity binaries are in an
  Arch package built under emulation, not tested on target hardware. No boot-to-shell,
  idle resource, scheduler/API latency or throughput claim is made.
- No Quickshell/QML shell has been implemented or run. The frozen GTK4 source is a
  historical experiment; no Wayland/Hyprland target was used.
- No llama.cpp server or licensed model is configured; no real local inference.
- No GPU/NVIDIA/CUDA, ComfyUI, systemd, D-Bus, sensor/BLIND, iPhone, NAS,
  production privacy-vault operations, consent/erasure or remote-administration
  capability has been exercised. The optional private vault has portable tests,
  but no production key or device integration.
- Full goal lifecycle beyond list/pause/resume/cancel and CRUD, natural-language plan generation and strategy scoring,
independent heartbeat workers/configurable retry policy, external event delivery and resource quotas, sandboxed Agent Runner service,
  strategy optimizer, production identity/session
  management, encrypted memory/evidence vault, and broad Mont Blanc modules remain
  incomplete. Several contracts are types only.
- The current authorization credential is a local bearer secret shared with the
  broker. The broker is a separate process, but both local service units use the
  same unprivileged account. Production provisioning, independent service identity,
  owner login, revocation UX and Linux security review remain release blockers.
- Go owner-token setup is included in the Arch package and works in local tests.
  Automated first-boot orchestration, owner login/session distribution, and
  credential rotation/revocation workflows remain incomplete.

The portable Go development launcher is syntax-checked but has not been run on
Linux in this checkpoint; it intentionally exits on macOS.

## Next milestone

Validate the Go systemd units on a non-emulated Arch VM and test service restart,
credential loading and restrictions there. The package and same-UID Go process
flow have passed in an emulated x86_64 Arch container, but it does not boot systemd
or replace a VM/target run. Then build a functional Quickshell/QML shell against the
loopback API. No real Linux installation has been accessed or changed; a remote
baseline still requires an owner-provided SSH alias and explicit connection
authorization, and any mutation requires a verified recovery plan and explicit
authorization for that mutation.
