# ADR 0002: standard-library Go broker + Python/SQLite local core

Accepted for limited Task 1. Goal §2 requires Go or Rust for policy and permits
SQLite single-process V0. Use available Go 1.20.6 and CPython 3.11.7 for local
verification; Go module language floor 1.20, Python floor 3.11. Production toolchain
security support/pinning must be reviewed before Linux deployment. No ML service
is implemented, so the Discovery Python 3.12 recommendation is not its runtime.

Go standard library supplies native separate broker, HMAC and UDS HTTP. Python
standard library supplies SQLite, HTTP, tests and hashing. Zero third-party runtime
packages, no Docker, Redis, PostgreSQL daemon or package download needed.
Alternatives: Go + SQLite requires external SQLite binding/driver; all Python
violates native policy boundary; PostgreSQL + queue adds services without value
for a one-worker pure action. Revisit when concurrency/external effects arrive.

Migration: versioned SQL; UUID text keys, UTC timestamps, explicit JSON serialization,
FKs and unique constraints map to PostgreSQL UUID/timestamptz/JSONB. Export/import
under maintenance, reconcile row/event counts and hashes, replay integration tests,
then switch repository adapter. This migration has not been executed.
