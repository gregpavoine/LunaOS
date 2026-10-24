# ADR 0001: additive V0 on existing Omarchy (superseded)

This was the initial bootstrap decision and is superseded by the owner's 2026-10-24
architecture correction. LunaOS will be an independently installable Arch Linux
distribution and Omarchy is a reference/reuse source, not a required runtime. Keep
all Omarchy KEEP behavior during migration; do not install Omarchy on macOS or
mutate a Linux target without a read-only baseline and explicit authorization.
Independent image, installer, update and recovery work remains required and
unverified.
