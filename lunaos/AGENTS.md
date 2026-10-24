# LunaOS repository instructions

## Mission and architecture
- Build LunaOS as an independently installable Arch Linux distribution. Omarchy is a reference for reusable Arch/Hyprland configuration and integrations, not a required runtime dependency.
- Native OS control plane, Goal Engine, Scheduler, Policy Broker, Agent Runner, Memory/Identity/Resource services and shell are Rust. Use Tokio, systemd, D-Bus, Unix sockets, Linux APIs and SQLite initially.
- The existing Go/Python Task 1 implementation is frozen reference code. Do not expand the Python Goal Engine or add Python to the OS control plane. Python may be used only for isolated ML/third-party workloads that require it.
- Select GTK4 or Slint for the native Wayland shell through a documented technical evaluation. Local inference must use a real native backend such as llama.cpp; never report a missing model as working.
- Keep the repository's Mont Blanc feature scope mandatory. Docs, schemas, mocks and CI-only results are not runtime capability evidence.

## Platforms and access
- **Mac Apple Silicon:** controller, editor, Git, Rust portable tests and docs. Do not install Omarchy, systemd, NVIDIA drivers or CUDA on macOS.
- **Linux x86_64 NVIDIA:** target build/run and hardware verification only through the owner-provided SSH alias and explicit authorization. Before any write, perform a read-only baseline and verify a recovery plan.
- Never assume hostnames, users, paths, versions or credentials. Do not connect to a remote Linux system before receiving the alias and authorization.

## Safety and authority
- Owner-only privileged commands. The Policy Broker is a separate Rust process; authorization fails closed and is checked before scoring and again before execution.
- Budgets begin at zero. No payment, external message, remote access, destructive operation or private-data processing without an independent scoped grant.
- Never execute package removals, upgrades, bootloader changes, partitioning, service disablement or cleanup scripts from a spec alone. Preserve required Omarchy components during any explicitly authorized transitional deployment.
- BLIND stops camera, screen and microphone collection without recording an activation frame. No intimate/biometric third-party data without subject-specific applicable consent. Keep secrets and sensitive content out of the repo and logs.
- No mock capability may stand in for hardware, model inference, privacy controls, installation or recovery.

## Development loop
1. Read the relevant specs and decisions; identify one measurable native milestone.
2. Add behavior tests, verify the failure, then implement the smallest Rust slice.
3. Run Rust formatting, lint, unit/integration tests and a local process-level smoke test when applicable. Run frozen Go/Python reference tests separately.
4. Record real results and exact blockers in `docs/STATUS.md`; update `docs/OPEN_DECISIONS.md`, `docs/DECISIONS.md`, architecture and release matrix when scope changes.
5. Commit a tested checkpoint. Do not claim release readiness from local portable tests.
6. End each work cycle with implemented, tested, blocked, risks and next milestone.

## First native milestones
- Port the verified Task 1 behavior into the Rust workspace; retain Go/Python as reference.
- Split Policy Broker into its own Rust service with owner authentication and Unix-socket peer checks before adding privileged adapters.
- Complete goal lifecycle/planning, persistent scheduler recovery and an evidence-backed evaluator.
- Then validate the GTK shell, llama.cpp inference, Arch packaging and boot/recovery in an isolated Arch VM.
- Inspect a real Omarchy target only after SSH alias and explicit connection authorization. Do not modify any current Linux installation without verified recovery and explicit authorization.
