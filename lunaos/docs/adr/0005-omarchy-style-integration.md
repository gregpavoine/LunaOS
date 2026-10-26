# ADR 0005: Build LunaOS using Omarchy's repository and integration model

**Status: Accepted by owner direction, 2026-10-26. Supersedes ADR 0001's and ADR 0004's distribution-boundary statements where they describe LunaOS as a parallel distribution assembled separately from the Omarchy-derived source.**

## Context

The owner clarified that LunaOS must follow Omarchy's practical implementation model. The checked-out LunaOS repository already contains the Omarchy-derived Arch system source: declarative package manifests, Bash integration and lifecycle scripts, Hyprland configuration, and the Quickshell/QML desktop. Reimplementing these pieces under `lunaos/` creates duplicate installers, shells, package closures and lifecycle policy.

Omarchy maintains the system source, package recipes and ISO builder in separate repositories. LunaOS will keep the system source here and create sibling repositories for package recipes (`LunaOS-pkgs`) and the branded installer (`LunaOS-iso`).

## Decision

- This repository is the canonical LunaOS Arch distribution source and owns base configuration, system integration, desktop session and update/recovery behavior.
- Reuse the existing Omarchy-derived Bash, package, Hyprland and Quickshell systems. Implement only LunaOS-specific service and UI behavior here.
- Keep LunaOS Arch package recipes and the branded Archiso build in separate sibling repositories, based on the maintained Omarchy package and ISO repository structures.
- The ISO repository consumes pinned source and package revisions. It must not maintain a second installer, system lifecycle or duplicate base package closure.
- Preserve the existing `lunaos/` implementation as behavior reference until each required feature has a real replacement and passes install acceptance. Do not remove Omarchy components during this source migration.
- Continue using Go for the active Luna Core and Policy Broker path, Bash for OS integration, and Quickshell/QML for desktop integration. Python remains limited to isolated ML or third-party workloads. Keep the Rust checkpoint frozen until Go parity is established.
- Begin package and image support with x86_64. Add architectures only after native package builds and install acceptance pass for them.

## Consequences

An installed Omarchy system is not required at runtime: LunaOS builds from the system source in this repository. Omarchy is the upstream implementation and packaging model. Package publishing credentials and signing keys must remain outside source control; the package repository bootstraps trusted signing keys before installing LunaOS packages.

This decision does not authorize host cleanup, package removal, remote Linux access, disk changes, publication or release. The repositories and local implementation must pass their build and isolated-VM gates before a release claim.
