# LunaOS implementation plan — Omarchy-aligned target

The accepted target follows Omarchy's Arch composition model: upstream packages,
declarative configuration, Bash integration, Hyprland/Wayland, and Quickshell/QML.
Long-running Luna services use Go. The Rust workspace is a frozen tested
checkpoint; Python remains limited to isolated ML/third-party workloads. Mont Blanc
scope stays mandatory. This plan is a dependency order, not a completion claim.

## Dependency order

1. **Behavior inventory and Go broker.** Map each verified Rust behavior and
   contract to Go service boundaries. Port owner authentication, deny-by-default
   decisions, scoped grants and private Unix-socket peer checks into a separate Go
   Policy Broker. Exit: Go tests prove owner-only privileges, absent-budget payment
   refusal, policy denial before scoring, grant expiry and peer UID enforcement.
2. **Go Goal Engine and store.** Implemented: goal title/description persistence
   and owner-authenticated metadata edits before execution; goal persistence, weighted
   scoring after authorization, provenance, idempotent creation and one checksum
   action over SQLite. Tests cover duplicate idempotency, unknown-source hypotheses,
   owner-only operations, payment refusal without budget and persisted evidence.
   Full lifecycle and strategy evaluation remain.
3. **Go scheduler and lifecycle.** Single-process scheduler, durable claim and
   queued-job recovery after reopening the database are implemented. Multi-process
   worker safety, dependency ordering, configurable retries, pause/resume/cancel,
   and stale lease tests remain.
4. **Go Core API and contracts.** Loopback-only authenticated HTTP routes for
   create/list/read/tick/evidence and a Unix-socket broker client are implemented.
   Strict input and handler tests pass. A process-level Linux test now starts the
   Go Broker and Core, verifies queued SQLite recovery, owner-authenticated API
   access, duplicate idempotency and checksum evidence in a Debian container.
   Arch package installation and systemd validation remain.
5. **Quickshell/QML desktop.** Build owner login/session surfaces, goal list and
   status/evidence screens, command center and persistent privacy status. Exit: real
   service state renders under Hyprland, keyboard and mouse controls work, and no
   placeholder reports unavailable hardware as active.
6. **Arch package composition.** Define LunaOS package groups, systemd units,
   sysusers, file permissions and first-run credential setup. The native Go package
   now builds and installs in a disposable x86_64 Arch container; Broker/Core also
   run as the `lunaos` account there. Exit remains a non-emulated Arch VM that boots
   systemd, loads credentials, validates unit restrictions and verifies restart.
7. **Linux adapters.** Add one typed broker adapter at a time for systemd/D-Bus,
   Hyprland, storage and GPU only after read-only discovery. Exit: each adapter has
   scope tests, owner-grant tests and a real Linux VM integration check.
8. **Local model and strategy evaluation.** Connect llama.cpp with a licensed model
   manifest; treat outputs as proposals. Exit: reproducible offline inference and
   strategy baselines with measured latency/resource use; no model output bypasses
   policy.
9. **Mont Blanc capabilities.** Implement privacy/consent, memory, identity,
   resource/GPU scheduling, ComfyUI/media workflows, Intimacy/Discovery, iPhone
   sync, NAS/backup, voice and remote administration as separately scoped services.
   Exit: every applicable acceptance row has real evidence from the actual adapter
   or device; mocks and schema-only contracts do not count.
   The initial Intimacy control milestone now persists STOP and blocks requests under
   unknown privacy context. Next dependencies are a scoped owner session/consent,
   separate encrypted memory, a real local conversation backend, then adult media
   indexing and ComfyUI workflows with safety and hardware verification.
10. **Independent image and release.** Build Arch packages and an `archiso` image;
    compile every LunaOS-owned core service for the declared target CPU, producing
    generic x86-64-v1 and CPU-gated x86-64-v3 variants where supported; enable
    supported Quickshell QML cache/AOT packaging; record startup, latency, idle
    resource and throughput benchmarks. Verify signing, installation preview,
    updates, rollback and recovery. Exit: a
    clean VM boots, installs and recovers; target NVIDIA and 72-hour soak checks pass
    before release. No existing Linux host is modified without its authorized
    baseline, verified recovery and explicit mutation approval.

## Kernel and driver work

Use the supported Arch kernel and existing driver packages by default. Do not write
a LunaOS kernel. Only when a read-only hardware baseline proves a required driver
is unavailable in maintained packages may we build a kernel package or DKMS module.
That work must specify kernel compatibility, package reproducibility, target boot
verification and a known-good fallback before entering an image.

## Development commands

`make test`, `make check` and `make build` verify the portable Go Core/Broker slice.
`make test-rust-checkpoint`, `make check-rust-checkpoint` and
`make smoke-rust-checkpoint` are explicit frozen-checkpoint checks. None validate
the complete Go/QML target. Build and run Quickshell, systemd, packages and images
in an isolated Arch VM; do not infer Linux correctness or performance from macOS.
Never run the baseline against a real host until the owner provides an SSH alias
and explicit authorization.
