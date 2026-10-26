# LunaOS Standalone Install ISO Design

## Intent

Turn the existing Arch live image into an interactive, independently installable LunaOS ISO. The live environment remains useful for recovery and includes a visible Install LunaOS entry. The installed system receives the LunaOS Go services, Hyprland session defaults, owner-group access, and systemd units from the ISO itself.

## User flow

1. Boot the x86_64 UEFI ISO into the existing live session.
2. Start LunaOS Installer from the desktop or terminal. The installer prints the detected disks and installation package plan without writing to them.
3. The owner confirms the exact phrase `INSTALL LUNAOS`, then uses Arch Linux's maintained `archinstall` guided TUI to select the target disk/layout, account, bootloader, and optional LUKS encryption. The config contains no disk path and does not select a disk automatically.
4. Archinstall installs from a local package repository bundled on the ISO, runs the LunaOS package's target-side finalizer, and returns to the live session. The finalizer configures greetd/Hyprland, adds interactive users to `luna-owner`, removes the temporary local repository entry, and enables the LunaOS services.
5. Reboot is an explicit user action after removing the installation media.

## Defaults and scope

- First installer target: x86_64 UEFI, consistent with the current ISO profile.
- Use upstream Arch `archinstall` for disk operations and package installation rather than reimplementing partitioning in Bash.
- Disk selection, partition scheme, encryption, username, and bootloader remain visible operator choices; the wrapper never passes a preselected disk or unattended credentials.
- Bundle the exact target package set and LunaOS package in a local Pacman repository on the ISO so installation does not require Internet access. The installed system uses standard Arch repositories after the finalizer removes the ISO-only repository entry.
- Archinstall/Python is an installer-only upstream tool in the live ISO; it is not installed as a LunaOS service or included in the Go control plane.
- Do not claim dual-boot preservation, snapshots, recovery, Secure Boot, or physical-hardware support until each is exercised in isolated VM/hardware acceptance.
- This design does not change adult-content capabilities or edit public documentation to hide them. Any GitHub compliance response must be accurate about shipped functionality.

## Boundaries and safety

The ISO build and package assembly are non-destructive. The installer itself can partition the operator-selected target disk only after interactive selection and archinstall's disk confirmation. The installer must refuse non-interactive execution, empty target lists, missing local package repository, unsupported boot mode, and non-root execution. No remote host or current user installation is touched.

## Acceptance

- The ISO contains `lunaos-install`, a desktop entry, archinstall, the target package repository, and a LunaOS package.
- The guided flow presents no default disk and requires the owner to confirm before archinstall starts.
- The generated installed system contains the actual Go binaries, systemd service units, first-boot credentials, greetd/Hyprland configuration, and working `luna-owner` socket permissions.
- A disposable x86_64 UEFI VM completes a disk installation, reboots into the installed system, reaches greetd/Hyprland, starts Core and Policy Broker, and verifies one benign checksum goal.
- No acceptance claim is made until the built ISO completes that disposable VM installation. macOS source checks and ISO boot-only smoke are insufficient.
