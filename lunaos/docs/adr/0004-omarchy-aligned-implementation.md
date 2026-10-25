# ADR 0004: Omarchy-aligned distribution and implementation techniques

**Status: Accepted by owner direction, 2026-10-25. Supersedes ADR 0003 for future implementation.**

## Context

The owner has directed LunaOS to follow Omarchy's practical technical approach
instead of making Rust the required implementation language. Omarchy is an Arch
distribution assembled primarily from upstream packages, shell-based system
integration and configuration, Hyprland, and a Quickshell desktop written in
QML/Qt. LunaOS should reuse these mature components and own its manifests,
configuration, release image and recovery path.

The Mont Blanc specification recommends upstream Arch/Hyprland packages,
explicitly says not to write a new kernel, and describes Rust as a preference for
some privileged services rather than a hard requirement. This ADR resolves the
later owner direction against those earlier language preferences.

## Decision

- **Distribution and host integration:** Arch Linux packages and systemd units;
  declarative package/configuration manifests; Bash scripts for provisioning,
  diagnostics, session integration and hardware-specific glue.
- **Desktop:** Hyprland/Wayland and Quickshell/QML, reusing Omarchy-compatible
  patterns where useful. LunaOS owns its shell configuration and brand. Prefer
  QML components and existing Linux desktop protocols over a new GTK or Slint
  application framework.
- **Long-running Luna services:** Go for the Goal Engine, Scheduler and Policy
  Broker where typed contracts, durable state and independent service boundaries
  justify compiled binaries. Keep the broker a separate least-privilege process;
  use systemd, private Unix sockets, peer credentials, owner authentication and
  deny-by-default scoped policy. SQLite remains the initial store.
- **AI workloads:** Python may run in isolated ML/ComfyUI or third-party workload
  environments. It must not own OS authority or become a required boot/control
  plane runtime. llama.cpp remains a separate real local inference backend.
- **Native build and performance:** every LunaOS-owned core daemon/service shipped
  in the image is compiled into a native target binary and packaged reproducibly.
  Upstream OS components are assembled from pinned Arch packages rather than
  rebuilt without reason. Package Quickshell/Qt QML through the supported compiled
  cache/AOT path when available; retain QML as the UI authoring format. Build
  release binaries for a declared CPU baseline. The specified Ryzen 9 5950X target
  may receive an x86-64-v3 Go build, selected only after CPU feature detection;
  ship a generic x86-64-v1 path for other compatible machines. Use measured PGO
  or additional target-specific optimization only when it improves representative
  workloads. No `-march=native` artifacts in a generic image.
- **Existing Rust workspace:** freeze `rust/` as a completed migration checkpoint
  and parity reference. Do not expand it or delete it in this checkpoint. The Go
  services must pass behavioral parity and process-level security tests before
  Rust binaries are removed from any package or image.
- **Kernel and drivers:** use supported upstream Arch kernels and vendor/upstream
  drivers. Do not create a LunaOS kernel. Build a kernel package or DKMS module
  only when a verified hardware requirement cannot be met by supported packages;
  pin its supported kernel range, validate it on target hardware, and document a
  bootable fallback/recovery path before shipping it.

## Consequences

Omarchy's relevant techniques are package composition, shell integration and a
QML desktop shell—not a claim that every Omarchy component uses one language.
LunaOS remains an independent Arch distribution and does not depend on an installed
Omarchy system at runtime. `rust/` still exists, and the current local Core/API
remains Rust until a Go replacement is implemented and verified. This ADR records
the accepted target direction, not a completed migration.

No remote Linux access, package removal, host cleanup, kernel replacement or
installation is authorized by this ADR.

## Performance acceptance

"As fast as possible" means measured performance on the declared target, not
aggressive compiler flags selected by assumption. The first Arch VM and authorized
target runs must record cold/warm boot-to-shell, service startup, scheduler/API
latency (p50/p95), idle CPU/RSS and representative Goal Engine throughput. Set
numeric budgets after capturing an unoptimized baseline; then compare release
builds and keep only changes that improve the relevant metric without correctness,
security or recovery regressions. Linux kernel/module builds are required only for
verified hardware needs; the default is the upstream Arch kernel package.

## Migration gates

1. Inventory Rust/Go/Python runtime behavior and map each behavior to a Go service,
   Bash integration or Quickshell/QML UI component.
2. Port broker contracts and owner/policy tests to Go; demonstrate that denial
   occurs before scoring and again before action execution.
3. Port persistent goal and scheduler behavior to Go/SQLite; demonstrate restart
   recovery, idempotency, evidence, provenance and lifecycle parity.
4. Package and run the services under systemd in an isolated Arch VM with a real
   Quickshell session; do not count local source tests as OS integration.
5. Remove Rust from the image only after the Go implementation passes the parity
   suite and rollback artifacts are verified.

