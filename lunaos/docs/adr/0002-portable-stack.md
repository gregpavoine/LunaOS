# ADR 0002: portable Go/Python checkpoint

**Status: Superseded.** This records the limited Task 1 checkpoint and is retained
only to explain the frozen Go/Python reference implementation.

The first repository slice used a Go policy service, Python Core and SQLite. That
choice is not the LunaOS runtime architecture. New OS control-plane behavior belongs
in the Rust/Tokio workspace. Keep `services/policy/` and `luna/` unchanged except
for critical reference-test fixes; do not expand the Python Goal Engine. See
[ADR 0003](0003-native-rust-distribution.md).
