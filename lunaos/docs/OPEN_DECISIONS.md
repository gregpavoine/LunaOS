# LunaOS requirements map and open decisions

Updated 2026-10-24 after the owner's native-OS architecture correction. Requirements are not implementation evidence. Source order: current owner instruction; Goal Engine addendum for goal/authority rules; unified Mont Blanc U0–U21 specification for all other product requirements; Series 26 for Omarchy retention.

## Explicit architecture and constraints

| Requirement | Source | Current treatment |
|---|---|---|
| Independently installable OS based on Arch Linux | Owner correction; unified spec release scope | Accepted; Archiso/installer not built yet |
| Hyprland/Wayland, native GTK4/Slint shell | Owner correction | GTK4 selected in architecture evaluation; source added, Linux compile/runtime blocked |
| Rust/Tokio control plane and systemd services | Owner correction | Rust workspace and Tokio/Axum Core/API added; deployed systemd services remain pending |
| Native D-Bus, Unix socket and Linux API integration | Owner correction | Adapter boundary specified; target implementations not verified |
| Goal Engine, Scheduler, Policy Broker, Agent Runner, Memory, Identity, Resource Manager in Rust | Owner correction | Initial Goal Engine and in-process policy crate; other modules partial/design-only |
| Local LLM with real inference | Owner correction | llama.cpp OpenAI-compatible HTTP adapter; no local server/model configured or exercised |
| SQLite initially and extensible storage | Owner correction; Goal Engine §2 | SQLite WAL/FULL durable store; migration adapter not implemented |
| Omarchy as reuse/reference, never required final runtime | Owner correction | Prior additive-overlay ADR superseded; keep proven components until replacement tests pass |
| Existing Go/Python prototype retained and frozen | Owner correction | Existing source/tests preserved; all new behavior moves to Rust |
| No Omarchy installation on macOS, no remote access without alias + explicit authorization | Goal spec and user instruction | No install, baseline execution or remote connection performed |
| Full Mont Blanc U0–U21 scope remains mandatory | Unified specification | Feature matrix retained; no capability silently cut or mocked |

## Contradictions and resolutions

| Conflict | Resolution |
|---|---|
| Earlier bootstrap: additive app over existing Omarchy vs independent LunaOS OS | Owner's newer explicit correction wins. LunaOS owns distro configuration, installer, shell, services, updates and recovery. Preserve Omarchy components as references and migrate with KEEP tests. |
| Earlier ADR: Go broker/Python core vs Rust/Tokio control plane | Superseded by owner correction. Freeze prototype and port behavior/API/tests to Rust. |
| Prior `docs/STATUS.md`: Python is current runtime and Go broker is current boundary | Historical checkpoint only. This status is replaced by native migration status and must not guide new implementation. |
| Unified/legacy specs say “Omarchy/Arch/Hyprland” vs owner says Omarchy is not mandatory runtime | Use Arch + Hyprland as base; reuse Omarchy components only when compatible with independent ownership. |
| Old from-scratch ISO language vs earlier Task 1 add-on-only scope | Current correction authorizes independent distro and ISO architecture; actual image creation remains unverified and gated on reproducible packaging/recovery. |
| Prior stack discussion allowed Python for planner/ML vs owner excludes Python control plane | Python may only host isolated ML/ComfyUI/third-party workloads that require it. Planner, memory and system controller remain Rust. |
| Product autonomy vs independent policy/consent/owner authority | Hard policy gates, consent, spend=0 until configured, owner-only privileges and revocation always outrank score/model/user prompt content. |
| Zero-budget spending vs spend within explicit configured budget | No configured budget means refusal. A future budget allows only broker-enforced scoped spending after ledger/reservation/reconciliation exists. |
| Historical consent answer on retaining third-party intimate data vs revocable consent/legal deletion | Stop collection immediately; restrict and erase records/derivatives where required. Never infer a right to retain from historical answers. |
| BLIND vs authorized capture | BLIND/manual stop overrides agents and pending jobs. Only separate scoped visible/revocable capture session can run when all subjects and grants pass. No sensors are implemented here. |

## Unresolved owner decisions from source specifications

- Series 29.30: how source provenance should be displayed in the user experience.
- Series 29.42: permanent rule for requiring tests before every code modification.
- Series 29.79: handling already-held private data while BLIND is active.
- Series 29.82: mobile platforms beyond iPhone.
- Series 28: CPU/VRAM limits, GPU concurrency, notification behavior, goal priorities and spending budgets. Current single worker and score weights are engineering defaults, not owner-approved product settings.
- Series 22: disk and dual-boot strategy, filesystem/encryption and unlock, NAS defaults, cloud-GPU policy, recovery media and install target selection.
- Series 24: detailed private preferences. Do not import private history into general memory or this repository.
- Local license choices and model size/context/quantization; target NVIDIA GPU, VRAM and thermal budget.
- LunaOS source/package/asset licensing and the release signing-key lifecycle; the Arch package template must not ship until this is decided and its license metadata/artifact signing are verified.
- Per-feature data retention, NAS replication scopes, iPhone coverage, legal/provider constraints and user-facing visibility where the source spec intentionally leaves product choices open.

These decisions are not needed to implement the safe checksum slice or define the architecture. Keep dependent capabilities blocked until configured.

## Critical risks and constraints

1. **Boot/recovery risk:** this is now an OS distribution project. Installer, disk selection, bootloader, rollback, update and recovery must be proven in disposable VMs and then target hardware; no destructive host action is inferred from this architecture authorization.
2. **Security boundary incomplete:** current Rust API uses a development owner bearer token and an in-process broker crate; it is not production identity, separate-UID broker isolation, D-Bus authorization, MFA or peer-credential enforcement. Do not grant privileged host actions.
3. **Current Core network surface:** Axum binds loopback by default and authenticates goal routes, but has not been hardened or reviewed as a production service. Capability status is a diagnostic, not readiness proof.
4. **Target hardware absent:** no approved SSH alias, Linux baseline, NVIDIA/ComfyUI measurement, actual Hyprland run, reboot, sensor cutoff or recovery test exists. The baseline script remains unexecuted.
5. **LLM adapter vs inference:** the Rust HTTP adapter targets a real llama.cpp server. No server/model is provisioned, licensed or measured; generated text is not yet connected to goals and cannot be treated as execution authority.
6. **Privacy scope is release-blocking:** encrypted private vault, consent receipt/revocation, sensor isolation, BLIND, deletion and model-artifact handling are not implemented. Keep all private data/capture disabled.
7. **Persistence/side effects:** SQLite slice covers one bounded pure action. It is not yet durable leasing/outbox/inbox for multi-process or external effects. Payments/messages/filesystem writes must remain unavailable.
8. **Mont Blanc scale:** the full U0–U21 feature set includes GPU, consent, Intimacy, Discovery, iPhone, NAS, remote administration, orchestration and signed release. Current bootstrap does not satisfy V0 exit or Mont Blanc release.

## Inputs needed before later milestones

No user input is essential for the local Rust migration milestone. Before Linux-specific investigation, obtain the SSH alias and explicit authorization. Before disk installation or changes, obtain an explicit target/disk choice and demonstrated recovery authorization. Before spending, recipient communication, private media or sensor capture, obtain the applicable owner configuration plus independent subject consent where required.
