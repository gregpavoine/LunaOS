# ADR 0003: native LunaOS distribution and Rust control plane

Accepted 2026-10-24 from the owner's explicit architecture correction. LunaOS ships as an independently installable Arch Linux OS, with Hyprland/Wayland and systemd. Rust/Tokio owns Core, Goal Engine, Scheduler, Policy Broker, Agent Runner, Memory/Identity, Resource Manager and native shell. GTK4 is the selected UI (evaluation in `docs/architecture/native-rust.md`). SQLite is the initial persistent store. llama.cpp is the first local inference adapter. D-Bus, Unix sockets and Linux APIs are the operating-system integration boundary.

Omarchy remains a source of reusable configuration, packages and proven integrations, but LunaOS owns its package manifests, installer, desktop shell, system services, updates and recovery. Keep upstream Omarchy components marked KEEP until their LunaOS replacements pass equivalent tests. Do not install Omarchy on macOS. Target-host inspection and mutation require a supplied SSH alias and explicit owner authorization; the baseline is read-only and must remain unexecuted until that authorization.

The existing Go/Python prototype is frozen as a migration reference. No new Goal Engine behavior is added there. The current Rust workspace is an initial implementation slice, not proof of independent boot, real inference, target integration or Mont Blanc completion.
