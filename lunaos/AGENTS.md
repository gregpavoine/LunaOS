# LunaOS repository instructions

## Mission and architecture
- Build LunaOS as an independently installable Arch Linux distribution using Omarchy's proven distribution techniques: upstream Arch packages, declarative package/configuration manifests, shell scripts for integration and lifecycle, Hyprland/Wayland, and Quickshell/QML for the desktop shell. Reuse components instead of replacing them without a demonstrated need.
- Use Bash for OS composition, setup and hardware integration; use Go for long-running Goal Engine, Scheduler and Policy Broker services where a compiled service is justified; use QML/Quickshell for desktop UI. Python is limited to isolated ML/ComfyUI and third-party workloads that require it.
- Rust code already in `rust/` is a frozen migration checkpoint/reference while the Go service design is evaluated. Do not expand the Rust control plane. Keep its behavior and tests available until an equivalent Go implementation passes parity tests; do not delete the Rust code in this architecture checkpoint.
- Use systemd, D-Bus, Unix sockets, Linux APIs and SQLite initially. Local inference must use a real native backend such as llama.cpp; never report a missing model as working.
- Use the upstream Arch Linux kernel and maintained NVIDIA/other drivers. Do not write a LunaOS kernel. Build or DKMS-compile an out-of-tree kernel module only when verified target hardware requires it, and record the exact kernel/driver compatibility and recovery path.
- Every LunaOS-owned core service shipped in the image must be built as a native target binary/package; Python is never on the boot or core control path. Assemble upstream system components from pinned Arch packages. Package-time QML compilation/cache generation should be enabled where Quickshell/Qt supports it. Optimize only for verified target CPUs and require reproducible benchmarks for startup, idle CPU/RAM and service/API latency; never claim fastest without measurements.
- Keep the repository's Mont Blanc feature scope mandatory. Docs, schemas, mocks and CI-only results are not runtime capability evidence.

## Platforms and access
- **Mac Apple Silicon:** controller, editor, Git, portable Go tests and docs. Do not install Omarchy, systemd, NVIDIA drivers or CUDA on macOS. Mac builds do not substitute for target-architecture release builds or Linux benchmarks.
- **Linux x86_64 NVIDIA:** target build/run and hardware verification only through the owner-provided SSH alias and explicit authorization. Before any write, perform a read-only baseline and verify a recovery plan.
- Never assume hostnames, users, paths, versions or credentials. Do not connect to a remote Linux system before receiving the alias and authorization.

## Safety and authority
- Owner-only privileged commands. The Policy Broker is a separate compiled service; authorization fails closed and is checked before scoring and again before execution.
- Budgets begin at zero. No payment, external message, remote access, destructive operation or private-data processing without an independent scoped grant.
- Never execute package removals, upgrades, bootloader changes, partitioning, service disablement or cleanup scripts from a spec alone. Preserve required Omarchy components during any explicitly authorized transitional deployment.
- BLIND stops camera, screen and microphone collection without recording an activation frame. No intimate/biometric third-party data without subject-specific applicable consent. Keep secrets and sensitive content out of the repo and logs.
- No mock capability may stand in for hardware, model inference, privacy controls, installation or recovery.

## Development loop
1. Read the relevant specs and decisions; identify one measurable native milestone.
2. Add behavior tests, verify the failure, then implement the smallest Bash/QML/Go slice appropriate to the component.
3. Run shell lint/syntax checks, QML checks on Linux, Go formatting/vet/unit tests and a local process-level smoke test when applicable. Run frozen Python reference and Rust checkpoint tests separately.
4. Record real results and exact blockers in `docs/STATUS.md`; update `docs/OPEN_DECISIONS.md`, `docs/DECISIONS.md`, architecture and release matrix when scope changes.
5. Commit a tested checkpoint. Do not claim release readiness from local portable tests.
6. End each work cycle with implemented, tested, blocked, risks and next milestone.

## First native milestones
- Document and test the Omarchy-aligned component map: Arch package manifests, Bash lifecycle/integration scripts, Quickshell/QML shell, and Go services.
- Keep the existing Rust checkpoint frozen until Go parity coverage proves equivalent owner authentication, policy denial, persistence, idempotency and evidence behavior.
- Implement the Policy Broker as an isolated compiled Go service with owner authentication and Unix-socket peer checks before adding privileged adapters.
- Complete goal lifecycle/planning, persistent scheduler recovery and an evidence-backed evaluator.
- Then validate Quickshell, llama.cpp inference, Arch packaging and boot/recovery in an isolated Arch VM.
- Inspect a real Omarchy target only after SSH alias and explicit connection authorization. Do not modify any current Linux installation without verified recovery and explicit authorization.
