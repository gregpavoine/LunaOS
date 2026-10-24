# LunaOS native architecture

This architecture supersedes ADR 0001/0002 wherever they describe LunaOS as an overlay or a Python/Go runtime. The existing `luna/` Python and `services/policy/` Go code are frozen behavioral references; new control-plane behavior belongs in `rust/`.

## Deployment shape

LunaOS is an independently installable Arch Linux distribution. It owns the package manifest, installer, system configuration, desktop shell, systemd services, update channel and recovery path. Omarchy remains a source of proven Arch/Hyprland packages and configuration patterns, not a required runtime dependency. Hyprland/Wayland is the desktop session. Rust 2024 and Tokio implement the control plane; Python stays outside it in separately constrained ML, ComfyUI and third-party workload processes.

The initial Rust workspace is split into `luna-contracts`, `luna-policy-broker`, `luna-goal-engine`, `luna-core`, `luna-llm-adapter` and `luna-shell`. These crates are build boundaries; they do not yet imply separate deployable processes. The goal store uses SQLite WAL with FULL synchronous writes. Linux-only device, service, GPU, camera, audio, USB and filesystem operations are traits/adapters and remain unavailable until real implementations pass target tests.

## Service responsibilities and trust boundaries

| Component | Responsibility | Current Rust status |
|---|---|---|
| LUNA Core API | Authenticated owner REST surface, capability reporting, lifecycle coordination | Tokio/Axum loopback API for goal create/read/tick; credential bootstrap and hardened listener pending |
| Goal Registry | Durable goals, idempotency, execution queue, evidence and event records | SQLite persistence and restart resume for one bounded action |
| Goal Planner | Bounded hierarchical plans and dependency DAG | Contract/design only in Rust; existing Go/Python reference behavior remains frozen |
| Scorer/Scheduler | Hard gates before deterministic score; recover queued useful jobs; idle success | Weighted v1 scoring and explicit `tick`; no resource broker/leases/fairness yet |
| Policy Broker | Owner-authenticated decisions, default denial, budget/privacy/action scope and signed goal grants | Rust policy crate; accepts only general `local.checksum`; payment without budget denied; no OS peer-credential daemon yet |
| Agent Runner | Execute a typed, bounded action under a scoped grant | Public fixture checksum only, no shell or arbitrary paths |
| Outcome Evaluator | Independently verify artifacts and define success | SHA-256 and exact byte-count verification |
| Strategy Evaluator | Compare measured outcomes with a baseline and propose bounded replanning | Schema and contracts pending; no optimization claimed |
| Memory/Identity Engines | Provenance-aware memory and stable, versioned assistant identity | Provenance status contract only; persistence and identity lifecycle pending |
| Resource Manager | Admit CPU, GPU, memory, disk, USB and model work by mode | Interface plan only; actual measurements require target Linux |
| Local LLM adapter | Use a local llama.cpp OpenAI-compatible server | Real HTTP inference adapter; no server/model configured or inference verified here |
| Native Shell | Owner command center and live capability status | GTK4 shell source present; compile/runtime/Wayland accessibility test blocked on Linux |

The control plane must not inherit authority from prompts, model output or caller-supplied role labels. Owner login authenticates the human; the broker independently evaluates every action. Guest requests cannot create privileged work. Missing, expired, revoked or mismatched grants fail closed. Budgets are zero until explicitly configured. A score never overrides policy. The current development token proves possession only and is not production owner identity, MFA or service isolation.

## UI evaluation: GTK4 selected

GTK4 with gtk-rs is selected for the Linux shell because the product targets Arch/Hyprland and needs a conventional native desktop, Linux accessibility integrations and direct fit with established Linux application conventions. The Rust bindings are mature and GTK4 is already packaged by Arch; the official gtk-rs installation guide documents the Arch GTK4/Libadwaita toolchain requirements. GTK objects are confined to the GLib main thread; Tokio/network work must return results through GLib channels or main-context tasks. [gtk-rs GTK4](https://gtk-rs.org/gtk4-rs/stable/latest/docs/gtk4/), [Linux setup](https://gtk-rs.org/gtk4-rs/stable/latest/book/installation_linux.html).

Slint was evaluated as the alternative: its Rust UI language and configurable Winit/Wayland/software-rendering backends give broad portability and can reduce dependence on GTK packages. GTK is favored because the shipping UI is Linux-first and benefits from native accessibility and GTK integration; Slint would be reconsidered for embedded/kiosk or multi-platform distribution needs. The official Slint docs confirm its Winit backend supports Wayland and can be selected with Cargo features. [Slint backends](https://docs.slint.dev/latest/docs/slint/guide/backends-and-renderers/backends_and_renderers/).

## Core data contracts

Rust `luna-contracts` is the versioned wire model (`contract_version=1`). `CreateGoal` contains action, normalized metrics, privacy class, optional expected cost and provenance. `GoalView` returns opaque ID, action, state, score, policy decision code, provenance state and queue state. `Evidence` carries goal ID, digest, byte count, provenance label and verifier result. Broker requests bind token, principal, action, privacy class and expected cost; decisions have a boolean plus stable reason code. Unknown provenance defaults to `unknown` and cannot be promoted from confidence alone.

Persistent v0 Rust entities are currently goal, queued job, evidence and goal event. Execution attempts, leases, full immutable outbox/inbox, goal dependencies, strategy trials, owner/session revision, policy grant lifecycle, expense ledger, memory assertions, persona revisions, external conversation grants, consent receipts and erasure jobs are required logical entities but are not represented as implemented by this initial Rust slice. Sensitive data, secrets, images, embeddings and private assertion text must never be put into general goal events.

## REST contract

The current loopback API is deliberately small and requires `Authorization: Bearer <owner token>` for goal operations:

| Method | Path | Behavior |
|---|---|---|
| POST | `/v1/goals` | Owner-authenticated, idempotent goal admission; broker gate precedes scoring |
| GET | `/v1/goals/{id}` | Owner-authenticated goal view |
| POST | `/v1/scheduler/tick` | Owner-authenticated single queue advancement |
| GET | `/v1/system/capabilities` | Safe capability states; not an authentication or readiness proof |

The source specifications remain authoritative for the eventual broader API: list/filter, lifecycle patch, bounded plan, independent evaluate, decisions, policy, identity revision, remote session and consent revocation endpoints. Those routes are not implemented in this Rust milestone. The current REST service has no public bind address, CORS or browser integration and is not a production network service.

## Planned Linux adapters

Implement narrow traits for systemd unit status/control, D-Bus, Hyprland Wayland session state, NVIDIA telemetry, ComfyUI workflow submission, filesystem backup, iPhone pairing/sync, NAS, sensors and privacy BLIND control. Each adapter reports explicit readiness and evidence. A missing adapter returns blocked/unavailable, never synthetic success. Before any host mutation, obtain owner SSH authorization and complete the existing read-only baseline. The baseline script must not be executed until authorized.

## Incremental implementation path

1. Preserve the existing Go/Python source as a frozen executable reference; migrate behavior contract-by-contract and keep its tests as compatibility evidence.
2. Complete the Rust Core contract, broker process boundary/peer credentials, goal lifecycle, migrations, DAG, durable event outbox/inbox and restart/crash tests.
3. Add owner identity session lifecycle, protected credential provisioning, policy grant expiry/revocation, production audit redaction and adversarial policy tests.
4. Implement the local llama.cpp service lifecycle/model configuration and verify actual offline inference on target hardware with measured memory, latency and licenses.
5. Build the GTK4 shell against the live API, then test keyboard, accessibility, Wayland startup, recovery and retained Omarchy launcher/notification/clipboard integrations on authorized Linux.
6. Implement real systemd/D-Bus/Linux service and resource adapters; run the existing ComfyUI workflow on NVIDIA and verify returned media plus persistence across reboot.
7. Implement vault, consent, BLIND/sensor isolation, deletion, Intimacy, Discovery, iPhone, NAS, remote control, resource scheduling and all Mont Blanc scope from the feature matrix. Keep every mandatory module visible with actual readiness and evidence.
8. Create pinned Arch package manifests and installer/recovery flows; build and boot the independent archiso image in disposable VMs; produce SBOM, license inventory and signed reproducible artifacts.
9. Run the full target-hardware acceptance matrix, adversarial privacy/security suite, power-loss/recovery tests and release review. Mont Blanc remains blocked until every mandatory acceptance gate is measured.
