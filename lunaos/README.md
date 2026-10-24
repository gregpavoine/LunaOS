# LunaOS — native AI operating system

LunaOS is an AI-native, local-first desktop operating system built on Arch Linux with Hyprland/Wayland. It is independently installable and owns its system configuration, package manifests, installer, desktop shell, services, updates and recovery. Omarchy is a reference implementation and source of reusable Arch/Hyprland components; Omarchy is not a mandatory LunaOS runtime dependency.

The MacBook Air M2 is the initial development/controller host. It is not the LunaOS runtime target: do not install Omarchy or Linux system components on macOS. Native system behavior must be verified on Linux. No remote Linux host may be accessed until the owner provides an SSH alias and explicitly authorizes the connection.

> **Release truth (2026-10-24):** the repository has a frozen Go/Python prototype and a newly started Rust migration slice. This is not a bootable LunaOS image, completed V0, or Mont Blanc release. The GTK shell has not been built/run on Arch, no local LLM model has performed inference, and no target hardware or independent ISO has been tested.

## Architecture

- **Operating system:** Arch Linux, systemd, Hyprland/Wayland. Omarchy components may be reused after their behavior, license and KEEP obligations are reviewed.
- **Control plane:** Rust 2024 and Tokio. Rust owns LUNA Core, Goal Engine, Scheduler, Policy Broker, Agent Runner, Memory Engine, Identity Engine, Strategy Evaluator and Resource Manager.
- **Native integration:** D-Bus, private Unix sockets, systemd and typed Linux API adapters. OS actions are deny-by-default, owner-authenticated, scoped and independently evaluated.
- **Desktop:** native Rust GTK4 shell. The UI choice and Slint comparison are documented in [native architecture](docs/architecture/native-rust.md).
- **Inference:** local llama.cpp adapter. No response or model capability is claimed until a real local model is configured and measured.
- **Persistence:** SQLite WAL/FULL for the initial offline store, behind a storage boundary that can grow to a multi-service database. Sensitive vault data requires separate encryption/key domains.
- **Python boundary:** isolated ML workloads, ComfyUI and third-party tools that require Python. Python is not required to boot LunaOS or run the OS controller, Goal Engine, policy broker or shell.

The Rust workspace in [`rust/`](rust/) contains versioned service contracts, a deny-by-default policy crate, persistent/scored goals, the Core REST surface, the llama.cpp adapter and the first GTK shell source. The existing `luna/` and `services/policy/` Go/Python implementation remains preserved as a frozen behavior/API/test reference; do not expand it while Rust migration is underway.

## Product behavior and safety

LUNA is a persistent assistant that helps the owner progress on projects, daily wellbeing and learning. It can plan, schedule, execute and evaluate authorized work. It must distinguish observation, inference, hypothesis, unknown data and verified evidence. It does not claim human feelings, and it does not autonomously seek employment.

The goal score uses normalized impact (0.30), urgency (0.20), user value (0.20), confidence (0.15) and efficiency (0.15) as a provisional engineering formula. Hard policy decisions precede scoring; a score cannot authorize an action. Missing authority is denied. Guest identity and model output cannot grant privileges. Payments are refused when no owner-configured budget exists. Unknown-source claims remain hypotheses. When no authorized useful work exists, idle is a successful scheduler result.

Private information stays in separated, encrypted domains. Sensor use requires independent visible grants. `BLIND` stops capture and processing and cannot be overridden by an agent/model. Subject consent is scoped, explicit and revocable; an owner's grant cannot consent for another person. Revocation stops collection and initiates required source/derivative deletion. Private images, biometric data, secrets and consent proofs never belong in general logs or this repository. Until those controls are real and verified, capture, private-media ingestion and sensitive ML remain unavailable.

## Required Mont Blanc 1.0 scope

Every scope below remains mandatory; the current status is tracked in the [feature matrix](docs/release/MONT_BLANC_FEATURE_MATRIX.md). No missing hardware or provider capability may be replaced by a mock and reported as complete.

1. **Core agency:** persistent identity/persona revisions; owner sessions; goal CRUD, hierarchical bounded planner and DAG; explainable score; adaptive scheduler, durable leases and resource modes; sandboxed agents; independent outcome/strategy evaluation; memory provenance, contradiction handling and evidence; event outbox/inbox.
2. **System and development:** independent Arch image, installer, update and recovery; systemd/D-Bus/Wayland integrations; system diagnostics and safe rollback; local model routing/benchmarks/adapters; project-aware software development, CI, Git and release workflows; resource/GPU manager and real ComfyUI image/video workflows.
3. **Privacy and media:** camera/screen/audio brokers; trusted BLIND latch; visible, subject-specific consent and revocation; encrypted private vault and deletion proofs; adult-only lawful Intimacy Core; consented body/face scopes and model validation; Discovery provider connectors, provenance, licensing and media indexing; voice, local perception and private notifications.
4. **Devices and remote services:** iPhone pairing, media import, encrypted backup and contacts archive; authorized WhatsApp archive where available; encrypted NAS replication and restore tests; hardened, revocable remote administration; explicit recipient-scoped third-party communications; budget-scoped financial intent/reservation/settlement; supported peripherals and IoT.
5. **Release:** complete G-001–G-030, U-001–U-040, BL-001–BL-030, Discovery, iPhone and Series 26 acceptance suites; real target-hardware performance, privacy, recovery, reboot and power-loss evidence; SBOM/license inventory; signed reproducible packages and independent ISO.

The full owner-controlled Goal Engine and Mont Blanc source documents remain the detailed behavioral authority. Public decisions and blockers are summarized in [`OPEN_DECISIONS.md`](docs/OPEN_DECISIONS.md).

## Initial Rust API

The current local Core API listens on `127.0.0.1:7766` by default and uses an owner bearer token for goal operations:

| Method | Path | Behavior |
|---|---|---|
| `POST` | `/v1/goals` | Create/idempotently return an owner goal; policy is checked before scoring |
| `GET` | `/v1/goals/{id}` | Read goal status, score and provenance state |
| `POST` | `/v1/scheduler/tick` | Advance one authorized queued action or return idle success |
| `GET` | `/v1/system/capabilities` | Report current implementation/blocker state |

This small contract is not the full source API; broader routes for goal lifecycle/plans/evaluation, decisions, identity, remote sessions and consent remain required and are listed in the architecture doc. The current Core route is local development software, not a hardened public network service.

## Development

Rust is the native implementation and the only runtime required by Core. GTK4 development packages are needed to build the shell, normally on Arch; they are not installed on the Mac host by this repository.

```sh
cd rust
cargo test --workspace --locked
cargo build --locked --release -p luna-core
# On Arch Linux with GTK4 development libraries:
cargo build --locked --release -p luna-shell --features gtk-ui
```

The frozen prototype can be checked for migration parity with `make test check` from this directory. It is not the LunaOS control-plane runtime. The Linux baseline at `scripts/linux-baseline.sh` is read-only but must not be executed until the owner authorizes it.

## Project documents

- [Native architecture, service contracts, APIs and incremental plan](docs/architecture/native-rust.md)
- [Resolved conflicts, open owner decisions and risks](docs/OPEN_DECISIONS.md)
- [Current architecture decisions](docs/DECISIONS.md) and [ADRs](docs/adr/)
- [Implementation milestones](docs/IMPLEMENTATION_PLAN.md)
- [Implemented, verified and blocked status](docs/STATUS.md)
- [Mont Blanc release feature matrix](docs/release/MONT_BLANC_FEATURE_MATRIX.md)
- [Threat model](docs/threat-model/overview.md)
