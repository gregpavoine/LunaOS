# LunaOS Omarchy-Style Integration Design

## Intent

Make LunaOS an Arch-based desktop product by extending the Omarchy-derived system already present in the main LunaOS source repository. Reuse its Bash lifecycle, system setup, Hyprland session, Quickshell desktop, update flow and recovery mechanisms. Add LunaOS-specific AI capabilities as small, packaged services instead of maintaining a second copy of the operating-system layer. Follow Omarchy's repository boundaries by maintaining separate LunaOS package and ISO repositories.

Omarchy is more than a collection of scripts: it composes Arch packages, system configuration, shell integration, desktop components, an installer image and lifecycle tooling. The useful pattern for LunaOS is to build on those maintained layers rather than recreate them in `lunaos/`.

## Goals

- Keep the existing Arch/Omarchy-derived system as the single source of truth for boot, users, system configuration, desktop session, updates and recovery.
- Maintain separate LunaOS package-recipe and ISO-build repositories, following the split used by Omarchy. Proposed names are `LunaOS-pkgs` and `LunaOS-iso`; the exact GitHub repository names and visibility are deployment details to settle before creating them.
- Keep the requested Mont Blanc product capabilities as LunaOS product scope, but ship them incrementally as independently testable packages or features.
- Preserve real, implemented Goal Engine behavior and the owner-authenticated, deny-by-default Policy Broker as LunaOS-specific services. They remain least-privilege Go binaries under systemd, with SQLite state and no model authority.
- Integrate LunaOS UI into the existing Quickshell desktop and command router rather than starting a parallel desktop shell or duplicate shell service.
- Produce a single downloadable LunaOS ISO through the existing Arch/Omarchy distribution and image path. The ISO must install in a disposable VM before being presented as ready; a split ISO is not the release target.
- Keep existing LunaOS code until each retained behavior has a mapped destination and passes parity and installation checks.

## Non-goals

- Writing or maintaining a LunaOS kernel, init system, package manager, compositor, bootloader, desktop framework or general-purpose installer.
- Running a separate full-system product from the nested `lunaos/packaging/arch` profile when the existing distribution image path can provide the same result.
- Rewriting the Goal Engine or Policy Broker in Rust, or using Python as a boot-time or privileged control-plane dependency.
- Removing Omarchy applications, services or user configuration from any installed machine as part of this source migration.
- Claiming Mont Blanc 1.0 features, model inference, hardware support or performance without runtime evidence.

## Architecture

### Distribution layer

The main LunaOS source repository's Arch/Omarchy-derived code owns the system lifecycle. Reuse its system configuration, setup and migration scripts, systemd integration, Hyprland session, Quickshell shell, update machinery, and boot/recovery flow. LunaOS configuration should use the same lifecycle and ownership rules as existing components. Do not create a runtime dependency on a separately installed upstream Omarchy package; this repository's distribution tree is the base being extended.

The branded LunaOS image remains a supported deliverable. A separate `LunaOS-iso` repository should follow the maintained Omarchy ISO structure and consume LunaOS packages from `LunaOS-pkgs` plus the base source/package versions from the main LunaOS repository. A separate `LunaOS-pkgs` repository should own Arch PKGBUILDs and package metadata, matching the upstream package-repository pattern. Avoid maintaining a second installer implementation, a second package closure, or duplicate boot/recovery policy. Until the new repositories build and install the same package set, the current nested profile is a migration reference and must not be deleted or treated as the release source of truth.

### LunaOS services

Keep the Go Core and Policy Broker as the boundary for AI-specific behavior that needs durable state, scheduling or authorization. Package their compiled target binaries and systemd units using the repository's existing Arch conventions. The broker stays in a distinct service identity and accepts only typed, allowlisted requests over a protected Unix socket. Deny by default; check authorization before scoring and immediately before dispatch. No model output grants authority. Use SQLite for the initial local registry and evidence store.

Do not add separate daemons for features that can be libraries or Core modules without a concrete isolation need. Optional workloads such as llama.cpp, ComfyUI and third-party models remain separately packaged and sandboxed. Python remains outside boot, shell, policy and goal-control paths.

### Desktop and commands

Extend the existing Quickshell/QML components and existing command routing. Move useful Luna dashboard interactions into that shell and keep UI-to-service communication on the established owner-scoped local API boundary. Do not ship `lunaos/shell/shell.qml` as an independent session shell. New commands must follow the repository's command naming and privilege conventions; avoid introducing a second CLI router.

### Packaging and release

Use the separate `LunaOS-pkgs` and `LunaOS-iso` repositories as the canonical package and image route. Package manifests declare only LunaOS additions; upstream Arch/Omarchy components remain upstream-built packages. Build LunaOS Go binaries for declared target architectures and test packages in a clean Arch environment. Release one ISO file plus its checksum. If the image exceeds GitHub's single-asset limit, reduce duplicated package content or use a release host that accepts a single file; do not split the image into user-reassembled parts.

## Migration plan

1. Inventory the nested `lunaos/` tree and map each real behavior, package, unit, script and UI component to an existing root convention, `LunaOS-pkgs` recipe or an explicitly retained LunaOS package. Mark docs, CI fixtures, prototypes and unimplemented capabilities separately.
2. Create `LunaOS-pkgs` and `LunaOS-iso` from the corresponding maintained Omarchy repository structures, then establish a reproducible build from the LunaOS source revision without changing a host installation.
3. Integrate the smallest vertical slice first: build/install the actual Go Core and Broker packages from `LunaOS-pkgs`, start them from the base system lifecycle, and prove owner-authenticated benign-goal execution in an isolated Arch VM.
4. Port the useful Luna dashboard into the existing Quickshell shell and route it through the owner-scoped API. Verify the real Wayland session in a disposable VM.
5. Integrate goal, memory, evaluator, strategy, resource and optional model features as separate measurable work, retaining the Policy Broker boundary and feature-specific test evidence.
6. Build the LunaOS branded ISO from the separate ISO repository and unified package/configuration path. Verify UEFI live boot, guided install to a disposable disk, installed-session boot, systemd services, goal/evidence path, and update/recovery before publishing.
7. Only after parity and acceptance, archive or remove redundant nested installer, shell and distribution profile code. Preserve historical Go/Python/Rust references and test fixtures where they remain useful; do not delete behavior based only on a document.

## Acceptance criteria

- The main source, package recipes and ISO profile live in separate repositories with reproducible revision references and one authoritative build path.
- One package/configuration/lifecycle source powers both the normal development system and the LunaOS release image.
- No duplicate LunaOS login shell, installer lifecycle, update framework or base package closure runs alongside the existing distribution implementation.
- The Luna Core and Broker are native Go binaries in Arch packages with systemd service boundaries, tested owner authentication and deny-before-score/dispatch behavior.
- The existing Quickshell desktop hosts LunaOS controls and does not start a parallel desktop shell.
- The release provides one ISO file and a checksum, and an isolated x86_64 UEFI VM installs it, boots the installed target and completes a benign goal with persisted evidence.
- No host cleanup or destructive operation runs without a separate read-only baseline, verified recovery plan and explicit authorization.
- Mont Blanc features remain clearly labeled implemented, verified, blocked or not implemented; specifications and CI-only mocks do not count as OS behavior.

## Builder decisions

- Reuse this repository's existing distribution source as LunaOS's base; do not add a separate dependency on an independently installed Omarchy system.
- Follow Omarchy's separate source, package-recipe and ISO repository layout, with proposed repos `LunaOS-pkgs` and `LunaOS-iso`.
- Keep the current installable ISO objective, but converge on one image/package lifecycle rather than preserving the nested profile as a second product implementation.
- Preserve current LunaOS code until the mapped replacement is installed and verified; implementation cleanup is a later gated migration step.
- Keep existing compatibility-sensitive internal command names during the first migration slice. Any user-facing rebrand or command namespace change should be a separate decision after the system integration works.
