# Omarchy-aligned LunaOS architecture

LunaOS is an independent Arch-based distribution assembled with the same broad
techniques used by Omarchy: upstream packages, configuration and Bash integration,
Hyprland on Wayland, and a Quickshell/QML desktop. LunaOS owns its image, package
manifest, branding, services, updates and recovery. An installed Omarchy system is
not a runtime dependency.

## Component and language boundaries

| Layer | LunaOS choice | Why | Current state |
|---|---|---|---|
| Base and boot | Arch packages, upstream Linux kernel, systemd, boot/recovery tools | Reuse maintained Linux components instead of writing an OS kernel | Arch ISO, boot and recovery not built or tested |
| Desktop session | Hyprland/Wayland plus Omarchy-compatible config and protocols | Reuse the target desktop and its mature integrations | Config/source only; no real Linux session test |
| Shell, bar and panels | Quickshell/QML/Qt | Same shell technology used by current Omarchy; supports a long-running compositor-integrated shell | Luna shell not implemented; Rust GTK source is a frozen experiment |
| OS composition | Bash scripts, declarative package lists, systemd units | Fits Arch installation and Omarchy's shell-first integration approach | Existing scripts are partial and not release-ready |
| Goal/API/scheduler | Go services with SQLite | Compiled Linux binaries, typed service boundary and existing Go reference behavior | Core/Broker process slice and owner credential CLI pass Linux-container tests; Arch package builds/installs in an emulated x86_64 Arch container; full service parity and systemd boot remain |
| Policy broker | Separate Go service over private Unix socket | Independent owner-authenticated policy boundary; avoids a scripting process holding broad system authority | Go owner-token, scoped-grant and Linux peer-UID checks passed in real Broker/Core processes in Debian and an emulated Arch container; systemd sandboxing and separate service-UID validation remain; Rust broker is frozen |
| Owner credential setup | Go `luna-owner-setup` CLI | Creates the credential format consumed by Go Core and Broker without replacing an existing file | Packaged, tested for private permissions and no-overwrite; automated first boot and owner login/session remain |
| ML and external tools | Isolated Python workloads only where required; real llama.cpp backend for local inference | Keep Python out of the OS control plane while supporting established ML tooling | No model inference or workload sandbox verified |
| Hardware extensions | Upstream drivers/modules first; DKMS or kernel package only for verified need | Avoid maintaining an unnecessary kernel fork | No target hardware inspected |

## Service and authority boundaries

The desktop shell communicates with unprivileged Go services over authenticated
loopback HTTP or a user-scoped Unix socket. The Goal Engine stores goals, plans,
events, leases and evidence in SQLite. The scheduler requests an action only after
the owner-authenticated Policy Broker permits it; the broker revalidates the scoped
grant immediately before dispatch. The broker exposes typed, allowlisted operations
and never accepts model-generated shell strings. Linux mutations require a separate
narrow adapter and explicit owner grant. Initial budgets remain zero.

Bash scripts compose packages and perform deterministic host integration. Scripts
must be idempotent, auditable and split into read-only diagnostics versus explicit
mutations. A QML shell or model process receives no direct root authority. systemd
units run services with dedicated users and filesystem/socket restrictions where
the service supports them.

## Kernel and packaging policy

Use an Arch-supported upstream kernel and matching driver packages. LunaOS will not
write a new kernel. If target hardware requires an out-of-tree module, first prove
that requirement from a read-only hardware inventory; package the module with DKMS
or a pinned kernel package, record the supported kernel/driver matrix, build it in
the Arch toolchain, and test both boot and fallback recovery in an isolated target.
Never replace the kernel or drivers on an existing Omarchy host as part of a source
development task.

Build LunaOS applications as Arch packages with `makepkg`/clean build environments;
build the install image with `archiso`. Pin package inputs and record package
checksums/SBOMs. A source build or package manifest is not evidence that the image
boots. Every LunaOS-owned core daemon must ship as a native binary for the declared
target CPU; scripts serve as deterministic glue, not as an interpreted core service.
Build Go release binaries for the target architecture with reproducible inputs.
Keep x86-64-v1 for generic Arch compatibility; the specified Ryzen 9 5950X can
use the optional GOAMD64=v3 binary after CPU feature detection confirms support.
Never ship the v3 binary as the only boot/runtime variant.
Use Quickshell's native Qt runtime and enable the supported QML cache/AOT packaging
path. Use upstream optimized package builds for the kernel, compositor, Qt and
drivers unless a verified LunaOS requirement justifies maintaining a custom build.

## Performance policy

Optimize from measurements on the declared hardware. Record cold and warm
boot-to-shell time, daemon startup, scheduler and API p50/p95 latency, idle CPU/RSS,
and representative goal throughput. Establish budgets from the first real Arch
baseline, then compare optimized release builds against it. Use profile-guided or
additional CPU-specific flags only when a target measurement shows a repeatable
gain and the image's CPU compatibility is explicit. Do not call a build “fastest” based on
compiler flags alone; do not use generic `-march=native` output in a portable ISO.

## Migration sequence

1. Keep the Rust workspace frozen and inventory the exact behaviors and REST/socket
   contracts currently verified there.
2. Port the Policy Broker contract and security tests into a separate Go process.
3. Port the Goal Engine, SQLite lifecycle and scheduler with parity tests for
   authorization-before-score, idempotency, restart recovery, evidence, provenance,
   dependencies, pause/resume/cancel and idle success.
4. Implement a Quickshell/QML Luna shell that displays real service state and uses
   authenticated service contracts.
5. Integrate systemd, D-Bus and only the required Linux adapters, behind package
   boundaries and owner-scoped policy.
6. Build optimized native service packages and an Arch ISO in an isolated Arch VM;
   collect baseline and release performance results, then verify boot, install,
   update, rollback and recovery before target-hardware work.
7. Remove Rust artifacts from the runtime image only after the Go/QML equivalent
   passes the full parity suite and recovery is verified.

## Evidence limits

The repository currently contains a Rust implementation checkpoint, a frozen
Go/Python reference implementation, partial Arch packaging and a GTK shell source.
It does not yet contain a working Quickshell desktop, complete Go Goal Engine, real
local model inference, bootable image or verified kernel/driver integration. Refer
to [STATUS](../STATUS.md) for executed checks and exact blockers.

