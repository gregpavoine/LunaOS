# LunaOS system architecture

The native distribution architecture and current service contracts are documented in [`native-rust.md`](native-rust.md). It supersedes this file's earlier description of a Python Core/Go broker running additively on Omarchy.

## Frozen behavior reference

The `luna/` Python Core, `services/policy/` Go broker, `migrations/core/`, `contracts/openapi/local.json` and `tests/` remain in the repository as the original portable behavior/API/test reference. They are not the LunaOS runtime architecture and must not receive new Goal Engine features. The Rust equivalent is under `rust/`; it is incomplete and migrates behavior incrementally.

## Current native slice

The Rust workspace contains versioned request/response and logical entity contracts; a deny-by-default policy crate; a SQLite-backed single-goal registry, weighted scorer and bounded checksum executor; a Tokio/Axum Core API; a loopback-only llama.cpp HTTP adapter; and GTK4 shell source. Only the checksum path is an implemented action. The policy library and Core currently share a process, so production policy isolation is still blocked. The REST API uses loopback TCP for the present local smoke test; Unix socket transport and OS peer credentials are required before production deployment.

The Linux system controller is not implemented. systemd/D-Bus/Wayland, NVIDIA, ComfyUI, sensors, consent, BLIND, vault, iPhone, NAS and recovery remain behind typed adapter/service contracts or are not implemented. No unsupported capability is emulated.

See [`STATUS.md`](../STATUS.md) for executed verification and blockers; see [`OPEN_DECISIONS.md`](../OPEN_DECISIONS.md) for source conflicts and requirements awaiting owner input.
