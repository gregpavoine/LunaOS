# ADR 0001: independently installable Arch-based LunaOS

**Status: Accepted; supersedes the Task 1 additive-overlay decision.**

LunaOS owns its distribution configuration, packages, installer, desktop shell,
system services, update path and recovery. Arch Linux is the base, Hyprland is the
Wayland compositor, and Omarchy is a source of reusable configurations and proven
integrations rather than a mandatory runtime dependency.

The old Task 1 work remains a local behavior reference. No current Omarchy
installation has been inspected or changed. Any eventual transition on an existing
installation requires an authorized read-only baseline, a verified recovery plan,
and explicit authorization for each system mutation. A standalone image must boot
and install in an isolated VM before being considered a candidate release.
