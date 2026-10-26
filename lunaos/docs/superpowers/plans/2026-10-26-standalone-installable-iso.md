# Standalone Installable ISO Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement the plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the current live Arch image into an interactive, offline-capable LunaOS installation ISO modeled on Omarchy's self-contained installer flow.

**Architecture:** Bundle Archinstall and the exact target package closure in an ISO-local Pacman repository. Build the existing LunaOS Arch package into that repository, provide a root-gated interactive launcher with no disk preselection, and run an installed-package finalizer through archinstall's documented target chroot command hook.

**Tech Stack:** Archiso, Arch packages and pacman local repository, Bash, Go PKGBUILD, archinstall, systemd, Hyprland and Quickshell.

**Spec:** `lunaos/docs/superpowers/specs/2026-10-26-standalone-installable-iso-design.md`; `lunaos/docs/LUNAOS_MONT_BLANC_1_0_CODEX_UNIFIED_SPEC.md`, §4 Boot, installer, updates and recovery.

## Global Constraints

- First target is x86_64 UEFI; the current host is macOS arm64 and is never modified.
- The disk path is selected interactively by the owner; configuration must not preselect a target or use unattended credentials.
- Partitioning and bootloader changes run only inside the installer against the explicitly selected target.
- Python is an upstream installer-only dependency, not a boot-time LunaOS control-plane dependency.
- The install repository must not remain configured on the installed system.
- Do not modify GitHub-facing NSFW references to evade moderation; keep public documentation accurate.

## Review Focus

- The installer must fail closed when run without an interactive TTY, as a non-root user, or without the bundled repository.
- The generated archinstall config must contain no target disk or stored credentials.
- The local repository must contain every configured target package and the LunaOS package, with checksums generated.
- The finalizer must remove the ISO-only repository, establish owner-group socket access, and enable only declared services.
- The installer desktop launcher must obtain privilege only for the root-owned installer entrypoint and require typed confirmation before launching archinstall.

---

### Task 1: Build target packages into the ISO

**Files:**
- Create: `lunaos/packaging/arch/packages.install.x86_64`
- Modify: `lunaos/packaging/arch/PKGBUILD`
- Modify: `lunaos/packaging/arch/packages.x86_64`
- Modify: `lunaos/scripts/build-archiso.sh`
- Modify: `.github/workflows/lunaos-archiso.yml`

- [x] Define the installed-system package set separately from packages used only by the live ISO.
- [ ] Build `lunaos-native` from the current source tree in the Arch builder and include it plus the target package dependency closure in the ISO-local repository.
- [ ] Add the local repository to the live ISO's pacman config only; keep target package installation online-free and remove the temporary repo in finalization.

### Task 2: Interactive installer and installed session setup

**Files:**
- Create: `lunaos/packaging/arch/airootfs/usr/local/bin/lunaos-install`
- Create: `lunaos/packaging/arch/airootfs/usr/share/lunaos/archinstall.json`
- Create: `lunaos/packaging/arch/airootfs/usr/share/applications/lunaos-install.desktop`
- Create: `lunaos/packaging/arch/install/lunaos-install-finalize`
- Create: `lunaos/packaging/arch/config/greetd/config.toml`
- Create: `lunaos/packaging/arch/config/hypr/hyprland.conf`
- Modify: `lunaos/packaging/arch/PKGBUILD`
- Modify: `lunaos/packaging/arch/airootfs/root/customize_airootfs.sh`

- [x] Add a read-only package/disk summary and interactive, root-only typed confirmation before starting guided archinstall; set no disk or credentials in the config.
- [x] Add a target-side finalizer for service enablement, owner-group membership, greetd login and Hyprland/Quickshell defaults.
- [x] Make the Install LunaOS entry reachable from the live desktop and terminal without granting a general root shell.

### Task 3: Contract, workflow and factual release status

**Files:**
- Modify: `lunaos/docs/STATUS.md`
- Modify: `lunaos/docs/release/MONT_BLANC_FEATURE_MATRIX.md`
- Modify: `lunaos/README.md`
- Modify: `.github/workflows/lunaos-archiso.yml`

- [x] Name the artifact as an installable preview version distinct from published v0.1.0; this workflow builds an artifact only and does not publish a release.
- [x] Record that source/package assembly is implemented separately from VM install acceptance; do not claim installable until that acceptance is run.
- [x] Preserve package checksums and publish a build artifact on Mont Blanc updates.

### Task 4: Verification boundary

- [x] Run non-destructive shell syntax, JSON template expansion, diff checks and the installer privileged-interpreter boundary regression; do not execute the installer on a host.
- [ ] Do not report a successful installed OS unless a disposable UEFI VM install, reboot and service smoke completes.
- [ ] If no Arch VM/runner is available, report the exact external blocker and leave the install acceptance state blocked.
