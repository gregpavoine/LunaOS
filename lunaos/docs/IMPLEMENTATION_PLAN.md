# Task 1 implementation plan

Scope: Mac-local, non-destructive checkpoint; not full V0. Sources: AGENTS.md,
Goal Engine spec §§0–13, unified master U0–U21; Series 26 controls Omarchy retention.

1. Inventory source documents and Git, preserve all supplied files. Record precedence,
   unanswered decisions and architecture in OPEN_DECISIONS, DECISIONS and ADRs.
2. Write acceptance tests before implementing: independent Go broker authentication,
   fail-closed policy, absent-budget refusal; real-process Python integration tests
   for score ordering, durable restart, idempotency, provenance and idle outcomes.
3. Implement Go HTTP/JSON broker over private Unix socket; owner key never enters DB.
   Issue signed, goal/action-bound expiring grants only after owner authentication.
   Evaluate at admission and immediately before execution. No root/shell/network tools.
4. Implement Python registry, transactional SQLite migration, deterministic scorer,
   serial scheduler, fixed-fixture checksum executor and independent outcome check.
   Persist event outbox + inbox delivery receipt atomically with result and job state.
5. Expose private Unix REST API and CLI with reproducible local startup/test scripts.
   Keep hardware adapters as interfaces with honest unavailable capability state.
6. Author Linux baseline script, inspect source and syntax only; never execute it.
7. Run full tests, race/static checks, live local smoke and restart tests; save actual
   evidence, update STATUS and release matrix, inspect diff and commit checkpoint.

Incremental V0 after this task:
- M1b: bounded dependency/parent DAG, cycle rejection, scheduler gating and v1→v2
  migration are implemented locally. Dependencies must be completed or approved
  with positive utility; already blocked, failed or zero-utility goals are rejected.
  Expired grants become visible blocked goals that require owner replacement.
  Remaining: owner session expiry/revocation, independent OS identities/peer
  credentials, production key store, encrypted private import vault.
- M1c: implement session expiry/revocation before production tool scopes, then
  generated subgoal planning, pause/resume/cancel lifecycle and private-vault design.
- M2 (requires approved SSH alias): run read-only baseline, verify recovery, choose
  scoped deployment; retain Omarchy launcher/notifications/clipboard and all KEEP apps.
- M3: real model router, Linux service adapter, measured GPU broker and existing
  ComfyUI workflow with downloaded/decoded output; keyboard/mouse command center.
- M4: boot/session startup, actual reboot recovery and V0 exit scenario. Full Mont Blanc
  remains gated on every mandatory module and hardware/privacy acceptance suite.
