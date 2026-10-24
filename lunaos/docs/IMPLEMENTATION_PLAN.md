# Native LunaOS implementation plan

The Rust migration replaces the frozen Task 1 Go/Python runtime incrementally.
Each milestone is a working vertical slice with tests and evidence. The full Mont
Blanc feature matrix remains mandatory; this plan does not reclassify missing work
as optional. Current milestone labels are updated in `STATUS.md`.

## Dependency order

1. **Native workspace and process boundary.** Establish locked Rust workspace,
   versioned contracts, separate owner-authenticated Policy Broker, Core API,
   SQLite Goal Engine, persistent local smoke and systemd unit sources. Keep the
   prior implementation unchanged as reference. Exit: Rust tests, process-level
   broker test and real local API smoke pass. Basic dependency edges, cycle
   rejection and scheduler waiting are also verified.
2. **Owner identity and grants.** Replace development bearer setup with OS-backed
   credential provisioning, session expiry/rotation/revocation and policy grant
   registry. Exit: unauthorized, guest, revoked, expired and owner flows pass on
   Linux with peer-credential checks and filesystem ACL evidence.
3. **Goal model and planner.** List/create/read/tick, dependency edges and the
   pause/resume/cancel subset and bounded owner-managed parent/dependency DAG
   are implemented. Port remaining CRUD/state machine, natural-language planning,
   and richer provenance. Exit: Rust parity
   tests cover every prior behavior plus DAG and transition invariants; migrations
   preserve stored records.
4. **Durable scheduler and agent runner.** Atomic SQLite leases, expiry recovery,
   local transactional outbox/inbox and a data-preserving schema migration are
   implemented. Add lease renewal, heartbeats, bounded retries, quotas, interruption handling and tool
   isolation. Exit: process-kill, duplicate-delivery and concurrent-worker tests
   prove no duplicated effect for every enabled adapter.
5. **Evidence and memory.** Implement source/confidence/freshness assertions,
   project-partitioned memory, sensitive encrypted vault, key rotation and verified
   backup/restore. Exit: provenance, access-control, erasure and restore suites pass;
   sensitive canaries are absent from general logs and network egress.
6. **Native local inference and strategy evaluation.** Configure a real llama.cpp
   backend/model manifest, connect proposal-only planning, measure inference and
   compare bounded strategy trials to a baseline. Exit: offline model inference is
   reproducible, licensed and benchmarked on the target; model text cannot bypass
   the broker.
7. **Linux integration.** Implement typed D-Bus, systemd, Hyprland, filesystem,
   GPU and ComfyUI adapters behind broker-enforced scopes. Exit: read-only adapter
   checks pass in a disposable Arch VM; a real compatible ComfyUI image workflow
   runs on the authorized NVIDIA target and its output is verified.
8. **Native desktop shell.** Complete GTK4 owner login, goals, status, evidence,
   consent, BLIND state and stop controls. Exit: keyboard/mouse and accessibility
   tests pass in Hyprland; owner actions use the same broker and session authority.
9. **Mont Blanc modules.** Implement privacy/sensor controller, consent and
   revocation/erasure, Intimacy Core, Discovery/media index, authorized messaging,
   iPhone sync, NAS/backup, remote admin, voice and multi-agent coordination as
   independently tested services. Exit: all U0–U21, BL-001..030, Discovery and
   iPhone acceptance rows have real target evidence; no mock capability is accepted.
10. **Independent distribution and release.** Build signed Arch packages, installer,
    live/recovery mode, updates, rollback, SBOM and reproducible ISO. Exit: clean
    Arch VM installs, boots and recovers; hardware smoke and 72-hour soak pass; all
    release artifacts and checksums are archived. No existing host is modified
    without its separately authorized baseline, verified recovery and explicit
    mutation approval.

## Development commands

Use `make test`, `make check`, `make build` and `make native-smoke` for the Rust
runtime. Use `make test-reference` and `make check-reference` only to preserve
behavioral reference coverage. Linux baseline and archiso build commands are never
run against a real host without explicit authorization and an independently
verified recovery plan.
