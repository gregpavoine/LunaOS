# First task for Codex (paste as-is)

You are the implementation lead for **LunaOS V0**, an autonomous AI-native desktop built ON TOP OF an existing Omarchy Linux installation, controlled initially from a MacBook Air M2. Read `AGENTS.md`, `docs/LUNA_GOAL_ENGINE_CODEX_SPEC_EN.md` and `docs/LUNAOS_MONT_BLANC_1_0_CODEX_UNIFIED_SPEC.md` before editing.

Task 1: perform a strictly **non-destructive repository and architecture bootstrap**. Do not install Omarchy on macOS. Do not connect to any remote Linux host until I provide an SSH alias and authorize the connection.

1. Map explicit, unresolved and contradictory requirements; report the most critical risks and constraints in `docs/OPEN_DECISIONS.md`.
2. Define service boundaries, secure owner-authenticated policy broker, persistent goal registry, scored scheduler, agent executor, strategy evaluator, data schemas and REST contracts. Provide an incremental V0 implementation plan.
3. Create a portable local development environment (choose the least complex technically justified stack) and implement a runnable vertical slice: persist a goal, score it, enforce a deny-by-default authorization policy, execute a benign whitelisted action, record evidence, resume a queued goal after process restart, and expose tests.
4. Add tests for owner-only privileges, absent-budget payment refusal, policy deny-before-score, duplicate job idempotency, unknown-source provenance and inactive-idle success.
5. Keep Linux integration behind adapter interfaces. Create a read-only Linux baseline script, but **do not execute it** until authorized.
6. Create `docs/STATUS.md` with REAL results (implemented / verified / blocked); no mock replacements for unimplemented hardware capability. Commit a working tested checkpoint.

Only ask me for data that is essential to proceed. Work independently inside the repository with reasonable safe technical defaults. Report failures precisely; do not claim the full V0 is finished after this first task.
