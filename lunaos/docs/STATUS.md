# LunaOS implementation status — Omarchy-aligned architecture checkpoint

## Single-file installer candidate rebuild — 2026-10-26 (not released)

The current installer source no longer embeds the complete Arch package dependency
closure a second time in the live ISO. It packages the LunaOS-native target
locally and configures Archinstall to fetch Arch packages from standard mirrors;
the installer now states that network access is required. The disposable QEMU
installer test receives a user-mode network adapter. The Arch ISO workflow now
reports image build, size/checksum, live boot, and install-to-disk as separate
steps, and refuses an ISO at or above 1,900,000,000 bytes so the release is one
file within GitHub's per-asset limit.

**Verified locally:** `make test check` passes on macOS arm64, including the
Archinstall configuration test for online official mirrors and the retained local
LunaOS repository. The Linux-only process test is skipped on macOS by design. The
Rust checkpoint's `cargo test --workspace --locked` also passed on this checkout.
The portable GitHub checks pass at commit `88334b56`; that run predates this ISO
rebuild.

**Not yet verified:** this change has not yet been built with Archiso, size-checked,
booted or installed in QEMU. The previous grouped ISO workflow run
[40328871429](https://github.com/gregpavoine/LunaOS/actions/runs/40328871429) was
cancelled before its hidden build/install step produced diagnostics. Public
`v0.1.1-rc.1` still contains split ISO parts and has no install-to-disk acceptance;
it is superseded. No clean single-file replacement has been published yet.

## Disposable installer VM gate implementation — 2026-10-26

On local branch `feature/installer-vm-acceptance`, an Archinstall test-input
generator and a QEMU/OVMF install acceptance harness are implemented. The
generator creates private CI-only credentials and a fixed `/dev/vda` layout.
The harness creates its own temporary qcow2 disk, injects a CI-only smoke package
through a temporary package repository, invokes the same interactive
`lunaos-install` entrypoint shipped in the ISO, then requires the installed Core
and Policy Broker to be active and a real owner-authorized `local.checksum` goal
to complete with the fixture SHA-256, byte count, provenance and verified flag.
The CI-only package removes its temporary repository stanza before reporting
success. It is not part of the release ISO package manifest.

**Verified locally:** `make test check` passed on macOS arm64, including Go tests
and compilation of the smoke helper, the installer boundary test, private
Archinstall config test, unsupported-host preflight regression, gofmt, vet, Bash
syntax and OpenAPI JSON validation. `git diff --check` passes. The host preflight
rejects this Mac before creating any VM artifact.

**GitHub Actions results:** run [40323710952](https://github.com/gregpavoine/LunaOS/actions/runs/40323710952)
built the Arch target package and ISO, and the live image passed its UEFI/QEMU
boot smoke. The new install acceptance harness stopped at preflight because
`qemu-img` was missing from the CI container. Run [40325027426](https://github.com/gregpavoine/LunaOS/actions/runs/40325027426)
also built the ISO and passed live boot, then stopped before starting the install
VM because the config generator was invoked without executable permissions.
Run [40326179529](https://github.com/gregpavoine/LunaOS/actions/runs/40326179529)
also built the ISO and passed live boot, then reached CI package creation but
failed because Go's build cache defaulted to `/.cache/go-build`, which is not
writable by the unprivileged package builder. None of these runs attempted disk
installation or uploaded an ISO artifact. The current fix directs Go caches to
the package builder's temporary home; CI must run again to verify it. The
installed target, systemd services and goal/evidence path remain unverified.
Run [40327343802](https://github.com/gregpavoine/LunaOS/actions/runs/40327343802)
built the ISO, passed live boot, compiled the CI smoke package and started QEMU,
but QEMU rejected the boot-order argument. The same run selected the debug
package instead of its executable package and the diagnostic log's private mode
blocked upload. The current changes fix the QEMU argument, select packages by
Pacman metadata and redact diagnostics before making them readable by the
Actions runner. No disk installation has yet been attempted. The Mac itself has
no Archiso, QEMU or OVMF runtime. No real Omarchy or Linux host was accessed or
modified.

## Standalone installer source checkpoint — 2026-10-26 (not released)

On `feature/standalone-installable-iso`, installer source now adds a root-only
interactive launcher, an archinstall guided configuration with no disk or user
credentials preselected, an installed-target finalizer, a separate target package
manifest, a live desktop launcher restricted by a command-specific sudoers rule,
and an ISO-local Pacman repository assembly path that downloads the target package
dependency closure. The Archiso profile version is 0.1.1; the GitHub workflow is
configured to build a temporary artifact on this feature branch and `mont-blanc` updates and no longer
publishes a release from this installer-preview workflow.

**Not verified:** this Mac cannot run Archiso or boot the resulting x86_64 image.
No Arch package build, ISO build, guided install, disk write, reboot to installed
system, systemd service check, visible desktop check or benign-goal acceptance has
been run for this change. The Archiso workflow is configured to build an artifact
from this feature branch or `mont-blanc` after push. Until an Arch x86_64 runner builds it and a disposable
UEFI VM completes the entire installation and reboot flow, the project has no
verified installable image. The existing `v0.1.0` artifact remains a live-only
image. This work is source implementation, not an installer release.

Source checks for this checkpoint passed: `bash -n` on the modified build,
installer, finalizer and customization scripts; `python3 -m json.tool` on the
archinstall template; a `jq` expansion check confirmed the installed package list
contains `lunaos-native` and no preselected `disk_config` or `users`; `git diff --check`;
and `bash tests/installer-boundary-test.sh` (protected interpreter
ignored a decoy `BASH_ENV`, and the installer rejected the decoy `-p` argument).
Go suites, `make test`, Arch package/ISO builds and VM installation were not run.
The local Docker CLI exists, but `docker info` did not respond within the 10-second
probe and that probe was terminated; no container or ISO build was started.

## Current release checkpoint — 2026-10-26

**LunaOS 0.1.0 is published.** The `v0.1.0` tag workflow rebuilt the Arch x86_64
live ISO and booted it with QEMU TCG and OVMF UEFI before publishing. Its captured
serial log confirms systemd reached `Graphical Interface`, first-boot owner and
broker credentials were created, and systemd started the Policy Broker and Core
services. This does not verify a visible Hyprland session, physical hardware, disk
installation, upgrades, recovery or Secure Boot.

- Source commit: `1b2688bafd81512291ece4656bc4d8c9dc3aa77c`
- Build and UEFI smoke: [GitHub Actions 40287955011](https://github.com/gregpavoine/LunaOS/actions/runs/40287955011)
- Temporary ISO artifact: [lunaos-0.1.0-x86_64-iso](https://github.com/gregpavoine/LunaOS/actions/runs/40287955011/artifacts/12635192570) (1,647,567,072 bytes)
- ISO SHA-256 from the build job: `0325c55fa07a39053b6606aa8a78157e881a9176bdf1b2b5402b3e95d1855868`
- `make test check` on the same source: PASS (Go tests, gofmt, vet, shell syntax and OpenAPI JSON validation)
- Published release: [LunaOS 0.1.0 — First Bootable ISO](https://github.com/gregpavoine/LunaOS/releases/tag/v0.1.0), created by [tag workflow 40288575830](https://github.com/gregpavoine/LunaOS/actions/runs/40288575830)
- Published ISO: `lunaos-0.1.0-x86_64.iso` (1,658,187,776 bytes)
- Published ISO SHA-256: `fa52b6c2048bd1fe2277786fd7c5b24e6903b23fa196cf0e783efcae2ab7ac6d` (matches the `SHA256SUMS` asset and GitHub asset digest)

This is a live technical preview, not the completed Mont Blanc 1.0 scope.

## Latest implementation checkpoint — 2026-10-26

At the pushed checkpoint `7f2830ec`, three Go Core goal features were merged into `mont-blanc`: bounded dependency planning (`67dd84ab`, merge `67ebc099`), policy-checked reprioritization (`521c69ee`, merge `6d631fb7`), and persistent metadata/editing (`b78f2d76`, merge `ec32e0d0`). On macOS arm64, `make test check` and `go test -buildvcs=false -race -count=1 ./services/core` passed. The repo's portable CI run [40292836837](https://github.com/gregpavoine/LunaOS/actions/runs/40292836837) passed its Ubuntu and macOS jobs. The Arch ISO run [40292836869](https://github.com/gregpavoine/LunaOS/actions/runs/40292836869) passed the x86_64 build and UEFI/QEMU smoke test, which confirmed the graphical systemd target and startup of first-boot owner setup, Policy Broker and Core.

The [test ISO artifact](https://github.com/gregpavoine/LunaOS/actions/runs/40292836869/artifacts/12636767777)
contains `lunaos-0.1.0-x86_64.iso` and `SHA256SUMS`; ISO SHA-256 is
`8956099794c9f7c2044f6e24a527e5a0e08f6c92fa1a72c618b81003642396e6a`. It was
built from source commit `ec32e0d0402060bd5356fbd2ffea014a1f6bc49d` and expires
2026-11-09. It is a temporary Actions artifact, not a new GitHub release tag; the
published release remains `v0.1.0`. The smoke test does not verify a visible
Hyprland/QML session, disk installation or recovery, physical hardware, or full
Mont Blanc 1.0. Reprioritization and planner REST behavior are covered by local Go
tests; the QEMU smoke only confirms the boot path and service startup. These
features are not full goal CRUD, natural-language planning or a completed Mont
Blanc 1.0 release.

Goal metadata support was implemented on `feature/go-goal-metadata` and merged into
`mont-blanc` at `ec32e0d0` (feature commit `b78f2d76`). Migration 4 adds bounded
title/description columns and backfills existing titles from each action; create
requests without the new fields preserve their previous idempotency hashes. The
owner edit endpoint updates text only for queued goals before execution, preserves
authorization and evidence state, and appends an audit event. `make test check` and
the Go Core race suite passed locally. The ISO build from the merge commit passed
the QEMU boot smoke; it confirms service startup, not use of the metadata route in
the live image.

The goal decisions endpoint was committed as `440591b3` and fast-forward merged locally into `mont-blanc`; that local commit has not been pushed and is not included in an ISO. It implements the documented `GET /v1/goals/{id}/decisions` contract using append-only SQLite audit events. The response is owner-authenticated, returns only event ID, kind, code and timestamp, and is limited to the latest 100 events in chronological order. Its focused handler test first failed on the missing route and then passed. `make test check` and `go test -buildvcs=false -race -count=1 ./services/core` passed locally on macOS arm64. Linux process and Arch VM validation were not run. This audit-history endpoint is not the evidence-backed strategy evaluator.

The `feature/go-goal-evaluation` change adds a limited owner-triggered evidence check for `local.checksum` and is fast-forward merged locally into `mont-blanc`; it is not pushed or included in an ISO. The evaluator recomputes the known public fixture SHA-256 and byte count, returns `insufficient_evidence` for uncompleted goals or unsupported provenance, reports mismatches, and appends an evaluation event without changing goal state. It does not evaluate strategies or general artifacts. The focused test, `make test check`, the Core race suite and `make build-local` pass locally on macOS arm64. Linux process and Arch VM checks were not run.

The `feature/go-system-capabilities` change implements owner-authenticated `GET /v1/system/capabilities`, with a live SQLite ping and a Core-credential Policy Broker health request over the configured Unix socket. The broker health endpoint also checks the configured peer UID. Broker failure reports `unavailable`; the local LLM is `unavailable` and hardware is `unverified`. Policy handler/client/API tests, `make test check`, the Core race suite, and `make build-local` pass locally on macOS arm64. This branch has been fast-forward merged locally into `mont-blanc`, but is not pushed or included in an ISO; Linux peer-credential and systemd validation were not run.

## Release 0 build status — 2026-10-25

Fresh local verification on macOS arm64: `make test && make check && make build`
passed, including Go unit tests, formatting, vet, shell syntax, OpenAPI JSON parsing
and all four development binaries. `make test-rust-checkpoint &&
make check-rust-checkpoint` also passed for the frozen Rust checkpoint, including
workspace tests, formatting, Clippy with warnings denied and `cargo check`.

The Mac cannot run `mkarchiso`, and Docker Desktop did not provide a usable Linux
engine. The GitHub Actions Arch x86_64 image-build workflow at
`.github/workflows/lunaos-archiso.yml` completed successfully and uploaded
`lunaos-0.1.0-x86_64.iso` as a temporary artifact. The artifact was downloaded and
`sha256sum --check` passed. ISO boot, installation, systemd session, Wayland shell
and target-hardware verification were outstanding at that checkpoint; the later
boot verification is recorded above.

- Build run: [GitHub Actions 40220619493](https://github.com/gregpavoine/LunaOS/actions/runs/40220619493)
- Artifact: [lunaos-0.1.0-x86_64-iso](https://github.com/gregpavoine/LunaOS/actions/runs/40220619493/artifacts/12608679636), 1,296,797,696 bytes
- ISO SHA-256: `dd7c9cfc3bd7b5fd8e811998f4ef3a2b5ff373ed8a34ee5b47563021f2a54752`

**LunaOS V0: INCOMPLETE. Mont Blanc 1.0: NOT RELEASED.** On 2026-10-25 the owner
changed the target implementation to follow Omarchy's Arch package, Bash,
Hyprland and Quickshell/QML techniques, with Go for long-running Luna services.
This direction is documented in ADR 0004. The Go Core/Broker migration has a
tested portable slice and a Linux process-level integration test in Docker
Desktop. A first Quickshell/QML dashboard, Arch live profile and separate
Core/Broker service identities have source implementations, but their target
runtime has not been verified beyond the QEMU UEFI smoke test above. There is no
verified Arch installation or target-hardware run.
Specifications, schema-only contracts and service templates are not counted as
completed OS capabilities.

## Implemented and verified locally

The Rust workspace and GTK experiment are frozen historical checkpoints, not active
LunaOS services or desktop functionality. The current implementation path is Go
services, Bash integration and a Quickshell/QML source shell not yet run under
Wayland.

- Go Policy Broker checks `SO_PEERCRED` against the configured Core service UID,
  distinct from its own service UID, and fails closed on non-Linux. It uses a
  dedicated Core-to-Broker secret separate from the owner API credential. Unit tests
  cover UID allow/deny and owner-secret rejection. Linux process tests verify the
  real UDS boundary in a disposable Debian `golang:1.20-bookworm` container. The
  systemd templates now assign separate `lunaos` and `luna-policy` service users,
  separate owner/Core credentials, and an owner-group-restricted Core API socket.
  The configuration has not yet been exercised under systemd.
- Go Goal Engine/Core in `services/core` implements SQLite WAL/FULL storage,
  authenticated goal create/read/list, policy authorization before weighted score,
  one benign checksum action, idempotency, unknown-source hypothesis provenance,
  durable evidence, queued-work recovery after Engine reopen, and idle queue success.
  Unit and HTTP handler tests pass. A Linux process test connects Core to the real
  broker binary; handler unit tests still use a policy test double.
- Go Core now also supports an owner-authenticated bounded parent/dependency plan
  at `POST /v1/goals/{id}/plan`. Migration 3 adds the plan tables without
  replacing existing goal rows; SQL guards reject combined hierarchy/dependency
  cycles and limit each goal to 32 prerequisites. The scheduler skips a queued
  goal until every prerequisite is completed and still revalidates Policy Broker
  authorization before the action. Tests verify unauthorized and malformed plans,
  duplicate/self/missing/cyclic references, cross-type cycle rejection, legacy
  row preservation, plan persistence across restart, and prerequisite ordering.
  This is local Go behavior evidence only; no target Arch/systemd validation or
  natural-language planning is implemented.
- Go Core also exposes owner-authenticated reprioritization through
  `PATCH /v1/goals/{id}`. It requires explicit finite 0..1 values for every score
  metric, calls Policy Broker authorization before scoring, atomically replaces
  score inputs and grant only on allow, and moves denied work to blocked without a
  score or grant. Running, paused, cancelled, completed and failed goals cannot be
  reprioritized. Engine and HTTP tests cover owner denial, metric validation,
  broker denial, persisted successful score and lifecycle restrictions. This
  changes scheduling priority only; it does not edit goal action, cost or content.
- `services/core/cmd/luna-core` is a runnable process with a private owner token,
  SQLite database, broker Unix socket, loopback REST API and two-second scheduler.
  A Linux process test starts the real Go broker and Core binaries, loads a queued
  goal from SQLite after Core startup, checks owner authentication and API
  idempotency, then verifies completed checksum evidence. The private Unix owner
  socket also accepts owner-group local UI requests without embedding a secret in
  QML; loopback TCP still requires its bearer token. This is a Debian container
  result, not an Arch image or systemd result.
- Go `services/identity` and `luna-owner-setup` create independent owner API and
  Core-to-Broker credentials as 32 random bytes encoded as 64 lowercase hex
  characters. The CLI creates new credential directories as `0700`, writes `0600`
  regular files without replacement, and prints only the owner token when explicitly
  asked. Unit tests cover distinct values, permissions and no-overwrite behavior.
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
| `make test && make test-process-linux && go vet -buildvcs=false ./services/policy ./services/core/... ./services/intimacy/... ./services/identity/...` in `golang:1.20-bookworm` | PASS: Go unit suites, real Core/Broker process integration including the owner Unix socket and queued recovery, plus vet; 0 failures |
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

## Mont Blanc local implementation checkpoint, 2026-10-25

This work adds distinct Core/Broker identities and credentials, first-boot
credential setup, a group-restricted owner Unix API for the local shell, a
Quickshell/QML dashboard source, and a live Arch profile that applies LunaOS
sysusers before creating the interactive account. The hosted Archiso build below
produced an image, but this checkpoint has not verified boot or installation.

| Check | Result |
|---|---|
| `make test check build` on macOS arm64 | PASS: Go unit tests, gofmt, vet, Bash syntax, OpenAPI JSON and four Go binaries |
| Debian Docker service and process checks | PASS: unit suites, real Core/Broker process test including bearerless owner Unix socket, queued-goal recovery and evidence, plus `go vet` |
| Arch `systemd-sysusers --root=<temporary-root>` | PASS: separate Broker/Core users and socket groups created without warnings |
| Arch `qmllint shell/shell.qml` | PASS: exit 0 with one `PanelWindow` uncreatable-type warning; no other warnings after cleanup |
| Updated Arch `makepkg --cleanbuild --noconfirm` | PASS: `lunaos-native-0.1.0-1-x86_64.pkg.tar.zst` built under x86_64 emulation; this exact artifact has not been installed or booted |
| `scripts/build-archiso.sh` final attempt | BLOCKED: profile validation, package installation, live-user customization, SquashFS creation (1,145 MiB) and UEFI systemd-boot setup completed; xorriso could not write the 633,189-sector ISO because Docker reported only 121 MiB free. Docker Desktop then reported a containerd metadata/blob I/O error. No ISO file was produced. |
| GitHub Actions Arch x86_64 build, run `40220619493` | PASS: produced and uploaded `lunaos-0.1.0-x86_64.iso` (1,296,797,696 bytes); downloaded artifact passed `sha256sum --check`; SHA-256 `dd7c9cfc3bd7b5fd8e811998f4ef3a2b5ff373ed8a34ee5b47563021f2a54752`. This verifies build and artifact integrity, not boot. |
| Follow-up `make build` on macOS arm64 | PASS: `luna-core`, `luna-policy-broker`, `luna-vault` and `luna-owner-setup` built under `build/local/`; these are host development binaries, not the Arch release |
| Docker Desktop retry after the ISO failure | BLOCKED: `docker info` reports `Docker Desktop is unable to start`; `open -a Docker` did not restore the engine. The host subsequently reported 6.5 GiB free. |
| Updated package installation and separate-UID runtime smoke | NOT RUN for this checkpoint |
| ISO boot, installation, systemd boot, Wayland shell session | NOT RUN |

The first Archiso attempt exposed and fixed an incorrect Go module working
directory. A subsequent attempt exposed obsolete boot-mode names and missing
bootloader files; the profile now uses UEFI systemd-boot. The final Mac Docker
attempt got past those stages and failed while writing the ISO after exhausting
Docker Desktop storage. The host had 340 MiB free immediately afterward and 6.5
GiB at a later check; Docker system cleanup was not run because it could affect
unrelated user images and containers. The successful hosted build above supersedes
that local build blocker.

The first Go build attempt failed because Go invoked an incompatible Intel Git at
`/usr/local/bin/git` while stamping VCS metadata. The Make target now passes
`-buildvcs=false`; the retry passed. This changes only build metadata generation.

The hosted CI ISO result is a real image-build artifact, but it does not prove the
image boots or runs on target hardware. The Rust smoke test at the earlier
checkpoint used a temporary database and test credential and removed them on exit.
The Go process test uses only a temporary directory, synthetic owner key and public
checksum fixture; it does not access a host installation. The vault CLI smoke also
used only a temporary directory.
Intimacy vault and API tests use temporary databases and
synthetic text only; tests verify that writes omit plaintext, encrypted backups
round-trip into new database paths, and corrupted restores do not publish a DB.

## Not implemented or blocked

- No real Omarchy or Linux installation was accessed. No packages/services were
  removed, disabled, retained or replaced on a host; **no cleanup occurred**. The
  read-only Linux baseline script remains unexecuted.
- A standalone Arch x86_64 live ISO has been built and booted to systemd's
  graphical target in QEMU; it has not been installed. No installer, update,
  rollback or recovery result exists. No real Omarchy or Linux host was modified.
- No runtime performance benchmarks have been recorded. The broker's prior Go
  cross-builds prove compilation for x86_64 baselines; the v3 artifact has not run
  on a feature-checked Ryzen 9 5950X. Core/Broker/vault/identity binaries are in an
  Arch package built under emulation, not tested on target hardware. No boot-to-shell,
  idle resource, scheduler/API latency or throughput claim is made.
- A minimal Quickshell/QML shell source now exposes Core connection state, goal
  listing, one benign checksum action and privacy-stop control. Arch `qmllint`
  exits 0 with one `PanelWindow` type-resolution warning; the shell has not run
  under Wayland/Hyprland. Frozen GTK4 source is still a historical experiment.
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
- Core and Broker have distinct accounts and credentials in the current service
  templates; the local shell uses a restricted Unix socket instead of embedding
  the owner secret. The first-boot credential setup and Core/Broker service starts
  were observed under systemd in the QEMU live boot. Their installed-system
  lifecycle, API behavior after systemd launch, production owner login, credential
  revocation UX and Linux security review remain unverified.

The portable Go development launcher is syntax-checked but has not been run on
Linux in this checkpoint; it intentionally exits on macOS.

## Next milestone

Publish the tagged v0.1.0 release from the verified `mont-blanc` commit. Install
the updated package in a disposable Arch environment, then validate unit restart,
credential loading and service identity separation in an installed VM. Earlier
package installation and same-UID process tests in emulated Arch containers do not
replace those tests. Run the shell visibly under Wayland and complete an
installer/recovery path before calling Mont Blanc 1.0 complete. No real Linux installation
has been accessed or changed; a remote
baseline still requires an owner-provided SSH alias and explicit connection
authorization, and any mutation requires a verified recovery plan and explicit
authorization for that mutation.

## Goal lifecycle controls checkpoint, 2026-10-26

Go Core now supports owner-authenticated `PATCH /v1/goals/{id}` pause, resume and
cancel operations. An additive SQLite `goal_control` table preserves the base
goal/job schema and records paused/cancelled overlays; state transitions append
events. Scheduler claims exclude controlled goals. Pause/cancel conflict while
execution is already running, and resume is limited to a paused queued job. Unit
and HTTP tests cover owner rejection, persistence through reopen, scheduler
exclusion, idempotent cancellation, invalid transitions and event recording.
This is a bounded Go Goal Engine feature, not full lifecycle parity or a Mont
Blanc release. Cross-service, systemd, Wayland, Arch boot and hardware behavior
remain unverified.
