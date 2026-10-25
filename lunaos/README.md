# LunaOS — native Arch operating system

LunaOS is an AI-native, independently installable Arch Linux distribution. It
follows Omarchy's proven composition techniques: upstream Arch packages,
configuration and Bash integration, Hyprland/Wayland, and Quickshell/QML for the
desktop shell. Omarchy is a reusable reference, not a runtime dependency. LunaOS
owns its package manifests, installer, shell, services, updates and recovery.

## Architecture

- **System composition:** Arch package manifests, systemd units and idempotent
  Bash integration scripts. Reuse the upstream Linux kernel and drivers; do not
  write a LunaOS kernel.
- **Native builds and speed:** every Luna-owned core service must ship as a
  compiled native binary for the target CPU. Package-time QML cache/AOT is used
  where Quickshell supports it. Performance is measured on the actual target
  before setting budgets or claiming an optimization; portable images avoid
  unverified CPU-specific compiler flags.
- **Luna services:** Go for long-running Goal Engine, Scheduler and Policy Broker
  processes, with systemd, D-Bus, Unix sockets and Linux APIs as needed. SQLite is
  the initial database.
- **Authority:** independent owner-authenticated Policy Broker. It denies by
  default, checks policy before scoring and again before execution, and requires
  explicit scoped grants. Budgets start at zero; external messages, payment,
  remote administration and destructive system actions require separate grants.
- **Owner credential setup:** the Go `luna-owner-setup` command creates separate
  32-byte hex credentials for the owner API and Core-to-Broker channel, stores them
  without overwriting existing paths, and prints the owner token only with explicit
  `--reveal-token`. A first-boot systemd unit is present; login-session delivery,
  rotation and revocation remain unfinished.
- **Desktop:** Hyprland/Wayland with a Quickshell/QML shell, following Omarchy's
  compositor-integrated desktop approach.
- **Local AI:** llama.cpp or another documented native local inference runtime.
  Model output is untrusted input; it cannot authorize actions. Model files and
  licenses belong in a manifest, not silently inside the source tree.
- **Data and privacy:** versioned local memory and evidence, provenance/confidence,
  encrypted private-data compartments, consent/revocation and BLIND sensor cutoff.
  Sensitive data never enters general logs or cloud training without explicit scope.
- **Mont Blanc scope:** goal planning and execution, evidence-backed evaluation,
  strategy trials, identity/persona continuity, memory, resource/GPU scheduling,
  ComfyUI and media workflows, privacy/consent controls, Intimacy and Discovery
  features, iPhone sync, NAS/backup, remote administration, update/rollback and
  recoverable independent installation. See the linked specifications and release
  matrix; a listed requirement is not a completed feature.

## Current implementation status

The project is in **architecture migration**, not release. The frozen Rust
checkpoint remains available as migration evidence. The active Go slice now has a
SQLite-backed goal engine, scoring after policy authorization, an idempotent create
API, a fixed benign checksum action, persistent evidence and a scheduler that can
resume queued work after reopening the database. The separate Go broker enforces
owner authentication and Linux Unix-socket peer UID checks. Go tests run locally on
macOS; a disposable Linux container process test also verifies the live broker/Core
boundary and queued-job recovery, including a restricted owner Unix socket for the
local UI. It is not Arch/systemd/hardware validation. The Intimacy
control foundation persists an owner-only STOP endpoint and fails closed on unknown
privacy state. A separate encrypted memory store and scoped/revocable owner consent
API exist, but the vault is disabled until its private key and database are configured.
The vault also supports encrypted owner-declared preferences and a disabled-by-
default Discovery metadata allowlist with explicit read/write/run consent scopes.
It also exposes an owner-confirmed transactional key-rotation primitive. The
`luna-vault` command backs up and restores the vault using separate owner-only
vault and backup key files. It will not overwrite a restore destination; key
generation is owner-only and does not print key bytes. OS key-store integration,
key-file rotation and lost-key recovery are not available.
These are storage and policy components only: there is no conversation model,
media connector, age/consent classifier, or generation backend configured.
A first Quickshell/QML command-center source and an Archiso live profile now exist;
the UI has not been run under Wayland, and the updated Arch package only built
under x86_64 emulation. The ISO build is not yet verified. Target systemd
validation, installer, recovery and real local inference remain unimplemented.
Mont Blanc is not released.

## Development on macOS or Linux

The current development targets build Go services with CGO enabled for the native
SQLite driver. Build and run them locally using:

```sh
make test          # Go broker and persistent Core tests
make check         # gofmt, go vet and shell syntax checks
make check-qml     # Quickshell/Qt QML lint (Arch/Qt environment)
make build         # local Core, Broker, vault and owner-setup binaries
make test-process-linux # real Core/Broker process test; Linux only
./scripts/dev-native.sh # temporary Go Core/Broker development run; Linux only
make build-linux-amd64 # broker only, generic Linux x86_64, CGO disabled
make build-linux-amd64-v3 # broker only, optional x86_64-v3 profile
make test-rust-checkpoint     # frozen Rust checkpoint tests
make check-rust-checkpoint    # frozen Rust formatting, lint and build checks
make smoke-rust-checkpoint    # frozen local Rust Core/Broker process smoke
```

The Rust checkpoint smoke test binds only `127.0.0.1:17766`, stores state in a temporary directory,
and removes its test credential and database when it exits. It proves local process
behavior only; it does not test Linux systemd or confer host privileges. The GTK4
source is a frozen experiment; target UI direction is Quickshell/QML.

Run frozen Python reference tests separately:

```sh
make test-reference    # includes frozen Python Task 1 tests
make check-reference   # includes frozen Python syntax checks
```

## Linux target and safety

The read-only inventory script is `scripts/linux-baseline.sh` and is not run
automatically. Do not connect to an existing Linux host until the owner provides
an SSH alias and explicitly authorizes the connection and baseline. Do not modify
any current Linux installation without an authorized read-only baseline, a verified
recovery plan and explicit authorization for the particular mutation. Never run
cleanup scripts, remove Omarchy packages or alter boot/storage based only on a spec.

## Project documents

- [Omarchy-aligned architecture](docs/architecture/omarchy-aligned.md)
- [Intimacy and private-memory checkpoint](docs/architecture/intimacy.md)
- [Historical Rust checkpoint](docs/architecture/native-rust.md)
- [Goal Engine specification](docs/LUNA_GOAL_ENGINE_CODEX_SPEC_EN.md)
- [Mont Blanc unified specification](docs/LUNAOS_MONT_BLANC_1_0_CODEX_UNIFIED_SPEC.md)
- [Implementation plan](docs/IMPLEMENTATION_PLAN.md)
- [Current status and test evidence](docs/STATUS.md)
- [Open decisions and risks](docs/OPEN_DECISIONS.md)
- [Mont Blanc release matrix](docs/release/MONT_BLANC_FEATURE_MATRIX.md)
