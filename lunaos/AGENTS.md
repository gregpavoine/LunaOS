# LunaOS subproject instructions

The parent `../AGENTS.md` defines Omarchy repository preservation, privilege, style, and test rules. Keep LunaOS-specific changes under `lunaos/`; never replace or remove parent Omarchy features while extracting reusable components.

## Product architecture

LunaOS is an independently installable Arch Linux distribution. Omarchy is a reference implementation and reusable source, not a final runtime dependency. Hyprland/Wayland is the desktop. The OS control plane is Rust 2024 with Tokio, systemd, D-Bus, Unix sockets and Linux APIs. GTK4 is the selected native Rust UI. SQLite is the initial local persistent store with an explicit storage abstraction/migration path. llama.cpp or another native inference server must provide real local inference. Python is restricted to isolated ML, ComfyUI and third-party tools that require it.

The current Go/Python prototype is frozen as a behavior/API/test reference. Do not expand it and do not delete it until Rust parity is implemented, tested and reviewed. Read `docs/architecture/native-rust.md`, the Goal Engine specification and Mont Blanc unified specification before broad architecture changes. Keep all U0–U21 Mont Blanc capabilities mandatory and visible in the feature matrix; mark true blockers rather than replacing missing hardware with mock behavior.

## Safety and scope

Never install Omarchy on macOS. Do not connect to a remote Linux host, run `scripts/linux-baseline.sh`, or perform Linux mutations until an SSH alias and explicit authorization are supplied. Before any later mutation, perform the authorized read-only inventory first. Never run destructive package, disk, bootloader, service-disable or data-removal commands based only on a specification.

The initial real action remains a deterministic public-fixture checksum. The native policy broker denies by default, authenticates owner authority independently of goal text/model output, checks policy before scoring, scopes grants to goal/action and refuses spending when budget is absent. No arbitrary shell, root, external messages, private media, biometrics, sensor capture, payments or real hardware claims without separately reviewed contracts and real verification. `BLIND` must stop all collection and processing; consent is subject-specific and revocable. Do not place secrets, private images, biometric data or assertion text into the repository, tests, logs, or general memory.

## Development and evidence

- Native Rust: run `cargo test --workspace --locked` in `rust/`.
- Frozen reference: run `make test check` in the LunaOS root.
- Keep GTK/Linux-only dependencies behind an explicit feature when the Mac host cannot supply them; verify the actual shell on Arch/Wayland before declaring it working.
- Record exact executed commands and their real outcomes in `docs/STATUS.md`; source, syntax, or contract checks are not target-hardware acceptance.
- Update `docs/DECISIONS.md`, `docs/OPEN_DECISIONS.md`, and the Mont Blanc matrix with each architectural milestone.
- Commit coherent tested checkpoints. Do not call the V0 exit scenario or Mont Blanc release complete until the real NVIDIA/ComfyUI, reboot, privacy, recovery, and release gates in the source specifications pass.
