# ADR 0002: standard-library Go broker + Python/SQLite local core (superseded)

This was the Task 1 reference-slice decision and is superseded by the owner's
2026-10-24 architecture correction. The Go/Python implementation remains frozen
behavioral reference material. New OS control-plane code is Rust/Tokio; SQLite is
the initial store; GTK4 is the selected native desktop toolkit; a local llama.cpp
server is the first inference target. Python is permitted only in isolated ML,
ComfyUI and necessary third-party workloads.

The previous rationale remains historical context only; it does not authorize new
Go/Python control-plane features. The migration is in progress, not complete.

Migration: preserve current prototype and tests until Rust parity is verified. The
SQLite storage boundary must remain replaceable; PostgreSQL migration is not a
current Linux runtime requirement and has not been executed.
