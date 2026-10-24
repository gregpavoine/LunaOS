# ADR 0003: native Rust LunaOS control plane

**Status: Accepted by owner direction, 2026-10-24.**

LunaOS is an independently installable Arch Linux OS, not an application layered
on Omarchy. The native OS control plane uses Rust and Tokio, systemd, D-Bus, Unix
sockets and Linux APIs. SQLite is the initial persistent store. The existing Go and
Python Task 1 implementation is frozen reference behavior and remains available for
parity tests while native services replace it incrementally.

The first native slice uses versioned Rust contracts, a SQLite-backed Goal Engine,
a deny-by-default Policy Broker running as its own process, and an Axum Core API.
The broker speaks a bounded newline-delimited JSON protocol only over an owner-only
Unix socket. Core and broker authenticate the owner using the same protected
credential; the broker also checks the connecting process UID against its own UID.
The only enabled action is a pure checksum of a public fixture. No privileged Linux
adapter is enabled.

GTK4 is selected over Slint for the first native shell because LunaOS targets
Hyprland/Wayland on Arch and needs direct access to mature GTK accessibility,
settings, theming and D-Bus integrations already common in that environment. This
decision reduces integration work; it does not claim the shell has been built or
tested on Linux. Revisit only if a measured startup, resource or portability test
shows GTK4 cannot meet the target requirements.

The local inference boundary is an OpenAI-compatible loopback llama.cpp adapter.
It remains unavailable until a real model with a documented license is configured
and inference is demonstrated. No Python process may own system authority.
