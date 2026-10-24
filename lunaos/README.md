# LunaOS — native Arch operating system

LunaOS is an AI-native, independently installable Linux distribution. Arch Linux
is the base, Hyprland provides the Wayland compositor, and Omarchy is a reference
for reusable configurations and desktop integrations rather than a required
runtime dependency. LunaOS owns its package manifests, installer, shell, system
services, updates and recovery.

## Architecture

- **OS control plane:** Rust workspace, Tokio runtime, systemd services, D-Bus,
  Unix sockets and Linux APIs. No Python or Go is intended in the boot path or LUNA
  Core.
- **Goal Engine and persistence:** Rust Goal Engine and Scheduler with SQLite as
  the initial database; migrations and storage adapters must preserve an upgrade
  path.
- **Authority:** independent owner-authenticated Rust Policy Broker. It denies by
  default, checks policy before scoring and again before execution, and requires
  explicit scoped grants. Budgets start at zero; external messages, payment,
  remote administration and destructive system actions require separate grants.
- **Desktop:** Hyprland/Wayland with a native GTK4 shell. GTK4 is selected for its
  existing Arch/Wayland accessibility, theming and D-Bus ecosystem.
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

The project is in **native migration**, not release. Rust currently has a local
Goal Engine/Core slice and a separate Policy Broker process. The only executable
system action is a checksum of a public fixture. GTK4 source exists but is not
verified on Linux; llama.cpp has no configured model; D-Bus, GPU, ComfyUI,
installation, ISO boot, recovery and real hardware behavior remain unverified or
unimplemented. The existing Go/Python implementation is frozen parity reference
code and is not the target OS runtime. Mont Blanc is not released.

## Development on macOS or Linux

Rust and Cargo are the native development path. Install `curl`, `jq` and OpenSSL
locally if you want to use the live API smoke or dev launcher.

```sh
make test          # Rust workspace tests
make check         # formatting, clippy and cargo check
make build         # native Core and Policy Broker release binaries
make native-smoke  # real local API + separate broker process + Core restart
make dev           # temporary local services, database and owner token
```

The smoke test binds only `127.0.0.1:17766`, stores state in a temporary directory,
and removes its test credential and database when it exits. It proves local process
behavior only; it does not test Linux systemd or confer host privileges. GTK4 is
opt-in with `make shell` and requires GTK development dependencies.

Run frozen reference tests separately:

```sh
make test-reference
make check-reference
```

## Linux target and safety

The read-only inventory script is `scripts/linux-baseline.sh` and is not run
automatically. Do not connect to an existing Linux host until the owner provides
an SSH alias and explicitly authorizes the connection and baseline. Do not modify
any current Linux installation without an authorized read-only baseline, a verified
recovery plan and explicit authorization for the particular mutation. Never run
cleanup scripts, remove Omarchy packages or alter boot/storage based only on a spec.

## Project documents

- [Native architecture](docs/architecture/native-rust.md)
- [Goal Engine specification](docs/LUNA_GOAL_ENGINE_CODEX_SPEC_EN.md)
- [Mont Blanc unified specification](docs/LUNAOS_MONT_BLANC_1_0_CODEX_UNIFIED_SPEC.md)
- [Implementation plan](docs/IMPLEMENTATION_PLAN.md)
- [Current status and test evidence](docs/STATUS.md)
- [Open decisions and risks](docs/OPEN_DECISIONS.md)
- [Mont Blanc release matrix](docs/release/MONT_BLANC_FEATURE_MATRIX.md)
