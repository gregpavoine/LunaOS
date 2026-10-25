# ADR 0003: native Rust LunaOS control plane

**Status: Superseded for future implementation by [ADR 0004](0004-omarchy-aligned-implementation.md), 2026-10-25. Retained as historical record of the Rust checkpoint.**

LunaOS is an independently installable Arch Linux OS, not an application layered
on Omarchy. The native OS control plane uses Rust and Tokio, systemd, D-Bus, Unix
sockets and Linux APIs. SQLite is the initial persistent store. The existing Go and
Python Task 1 implementation is frozen reference behavior and remains available for
parity tests while native services replace it incrementally.

The first native slice used versioned Rust contracts, a SQLite-backed Goal Engine,
a deny-by-default Policy Broker running as its own process, and an Axum Core API.
The broker speaks a bounded newline-delimited JSON protocol only over an owner-only
Unix socket. Core and broker authenticate the owner using the same protected
credential; the broker also checks the connecting process UID against its own UID.
The only enabled action is a pure checksum of a public fixture. No privileged Linux
adapter is enabled.

GTK4 is selected over Slint for the first native shell because LunaOS targets
Hyprland/Wayland on Arch and needs direct access to mature GTK accessibility,
settings, theming and D-Bus integrations already common in that environment. This
was the selection at the time; ADR 0004 now selects the Omarchy-aligned
Quickshell/QML desktop. The GTK shell source remains a frozen experiment and has
not been built or tested on Linux.

The local inference boundary was an OpenAI-compatible loopback llama.cpp adapter.
It remains unavailable until a real model with a documented license is configured
and inference is demonstrated. No Python process may own system authority.
