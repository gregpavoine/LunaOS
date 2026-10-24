# LunaOS decisions

- Source precedence: latest explicit owner instruction; Goal Engine addendum for goal/authority rules; unified Mont Blanc U0–U21 for all other LunaOS requirements; Series 26 for Omarchy preservation.
- ADR0001 (additive Omarchy V0) and ADR0002 (Go/Python local stack) are superseded. ADR0003 defines independent Arch LunaOS with Rust/Tokio control plane; ADR0004 selects GTK4.
- Existing Go/Python prototype is preserved and frozen as a behavioral reference. No more control-plane features are added there; migrate its behaviors, API and meaningful tests to Rust.
- Rust workspace edition 2024; `luna-contracts`, `luna-policy-broker`, `luna-goal-engine`, `luna-core`, `luna-llm-adapter` and `luna-shell` crates. Tokio/Axum, SQLite WAL/FULL, GTK4, and llama.cpp OpenAI-compatible endpoint are initial choices.
- Broker policy is default-deny; owner token required; guest/system caller labels cannot obtain an owner grant; broker grants are HMAC-bound to goal ID and action. Only public-fixture checksum is allowlisted. Missing budget refuses payment. Policy admission occurs before score computation.
- Weighted score is the Goal Engine spec's provisional formula: 0.30 impact, 0.20 urgency, 0.20 user value, 0.15 confidence, 0.15 efficiency. Series 28 values remain unapproved.
- Unknown-source provenance remains unknown/hypothesis; confidence does not validate a source. Idle/zero-useful-work returns success.
- The Rust SQLite slice is only for pure deterministic fixture hashing. It does not establish safety for external side effects or production multi-process execution.
- Linux APIs are traits/adapters and must report unavailable until real target implementations are verified. No mock provider may claim actual GPU, systemd, D-Bus, Wayland, ComfyUI, sensors, phone or NAS capability.
- No Omarchy installation on macOS, Linux baseline execution or remote Linux connection has occurred. An owner-provided SSH alias plus explicit authorization is required before connection.
- LunaOS owns its system manifest, installer, shell, services, updates and recovery. Reuse Omarchy components that pass KEEP and replacement acceptance requirements; Omarchy is not a final runtime dependency.
