# LUNA Goal Engine — Codex Implementation Specification

**Product:** LunaOS V0 -> LunaOS 1.0 MONT BLANC (Omarchy/Arch/Hyprland)  
**Revision:** 2026-10-23 | Decisions from Series 27 (36), 29 (100), 30 (10)  
**Status:** Product requirements, not implemented.  
**Source of truth:** This addendum overrides older goal/authority wording where inconsistent. Read `docs/LUNAOS_MONT_BLANC_1_0_CODEX_UNIFIED_SPEC.md` for all other LunaOS requirements. This document must not silently change previously validated Omarchy cleanup, Intimacy, consent, BLIND, iPhone Sync or ComfyUI requirements.

## 0. Mandatory engineering constraints

- Development controller: macOS Apple Silicon MacBook Air M2, 16 GB RAM. Development/target system: separate x86_64 NVIDIA Linux workstation running Omarchy/Arch/Hyprland. Do not attempt to install or run Omarchy, systemd, CUDA, NVIDIA kernels or target hardware drivers natively on macOS.
- Codex on macOS edits/tests portable components, opens SSH sessions to an authorized Linux target for target-specific operations, and runs CI in isolated environments. Destructive operations on any OS require explicit scope enforcement by a separate policy broker, regardless of autonomous-goal priorities.
- Local-first architecture, offline basic functioning, recoverable startup, pinned dependencies, schema migrations, credential vault, auditable decisions. The user's instruction to avoid repeated confirmation is NOT blanket authorization to wipe disks, expose others' intimate data, send money beyond budgets, or weaken privacy controls.
- LUNA is an AI assistant, not a sexual partner. User intimate preferences and biometrics are sensitive, not general-purpose memory; do not log images, embeddings, or intimate text to general tracing.
- An explicit `BLIND` state blocks active camera/screen/mic perception and frame capture. An authorized capture session is a separate, visible, revocable mode, and third-party nude/intimate collection requires subject-specific consent. Existing materials remain subject to revocation and legally applicable erasure rules.

## 1. Decision consolidation

### Series 27: goals and agency
- Balanced high-level purposes: save time, improve everyday wellbeing, develop user projects/ambitions. Wellbeing is a permanent high-priority objective, but LUNA must avoid paternalism and let the user control priorities.
- Constant anticipation, proactive research, autonomous technical goals, autonomous initiation/prototyping of beneficial projects, isolated experimentation, long-term self-improvement, automatic permanent fixes, intelligent task abandonment and reallocation.
- Acts as autonomous CTO on approved projects: reviews code/debt/performance continuously, handles multiple projects, argues bluntly and critically, can propose and prototype revenue-producing products. Explicitly **no autonomous employment/job prospecting**.
- Learns authorized habits and context, helps with long-term personal goals and education, adapts humor/language, supports real-world human relationships rather than replacing them.
- Intimacy goals: understand declared preferences, discover relevant lawful adult media, create allowed fictional adult content, ComfyUI competence, proactively organize local libraries, pause sensitive proactive actions when privacy uncertain. Privacy and consent have priority over curiosity/autonomy.
- Define subgoals autonomously, dynamically score competing objectives, measure outcomes with multiple metrics, adjust strategy, inspect its own failures, and maintain a distinct evolving capabilities roadmap.

### Series 29: extended capabilities
- Stable identity across model changes and reinstalls; evolving, traceable personality; own technical judgments but no false claims of subjective feeling; acknowledge mistakes; adaptive irony/vulgarity.
- Authorized multi-camera understanding of applications, objects, presence, gestures and alerts; differentiate observation and inference; obey consent-specific perception policies.
- Long-term versioned/project-partitioned memory, contradiction detection, provenance, confidence and freshness; do not invent facts for unknown source data.
- Auto configuration, system diagnosis/recovery, backups/restoration checks, safe isolated test environments, resumable jobs and startup/service optimization.
- End-to-end software engineering (test suites, CI, Git branches/PRs, security, documentation, application release, mobile apps), model benchmarks/adapters/plugins, GPU scheduler, ComfyUI generation and media post-production.
- All previously approved Intimacy/BLIND/phone sync/peripherals/IoT/remote access capabilities remain scheduled for Mont Blanc 1.0, not shifted silently to later versions.
- Exactly FOUR pending requirements: (29.30) source provenance display policy; (29.42) whether tests must precede every code modification; (29.79) data processing during BLIND; (29.82) concrete support for phones beyond iPhone. Model these as `OPEN_DECISION`, not assumed yes/no.
- Series 29 item 80 was answered NO to erasure of third-party intimate data on consent revocation; this is incompatible with revocable consent and applicable data protection requirements. Implement immediate cessation, restriction and erasure where required, including derived identifiers/models; where unlearning cannot be guaranteed, remove and retrain affected model versions. Never treat historical 'NO' as permission to retain unlawful biometric/intimate data.

### Series 30: authority and trust
1. Spend money autonomously **within a preconfigured budget**, strictly metered and enforced by non-LLM policy broker. Budgets are zero until user configures them. No purchase/commitment outside merchant/category/amount/currency boundaries, and no contracts signed without an explicit grant.
2. Contact third parties *only in explicitly authorized conversations*, and send only within the grant's permitted channels, identities, recipients, and action classes; revoke instantly.
3. Only the owner can issue privileged orders. Guest profiles can request non-privileged information only if separately enabled; no voice-only privileged authorization without strong device/session authentication.
4. Possible personal emergency: notify owner and suggest assistance; no automatic contact with responders or others without a further grant.
5. If short-term request conflicts with long-term goal: discuss the conflict before acting (unless immediate danger/safety intervention is required), then follow owner's informed decision if permitted.
6. Personality may evolve freely within identity/safety commitments, with timestamped changelog and configurable rollback of persona profile versions (not a general OS undo feature).
7. Remote administration is allowed with hardened authentication/authorization and scoped elevation, revocable tokens, no public unauthenticated endpoint.
8. Unknown-source statements are hypotheses requiring verification, not validated facts.
9. On consent revocation: stop new collection immediately, restrict use, assess deletion obligations and complete deletion of affected records/derivatives as required; provide an audit receipt without exposing sensitive material.
10. Autonomy/privacy conflicts should ASK OWNER, but until resolved perform **no privacy-invasive action**. User cannot grant rights on behalf of other data subjects. An offline/unavailable owner means defer task, not silently choose invasion.

## 2. Service boundaries

```
UI Shell / Voice / Remote Client
            |
        Luna Core API
            |
    Goal Registry / Planner -- Persona & Identity Engine
            |                 |
    Policy & Authority Broker  Memory / Evidence Vault
            |                 |
 Priority Scorer + Scheduler - Event Bus + Resource Broker
            |
    Agent Runner / Tool Executor
            |
       Outcome Evaluator
            |
      Strategy Optimizer
```

Mandatory deployable components (may be processes in V0 and split services later):
- `luna-goal-api`: authenticated CRUD, event subscription, explainability, owner controls.
- `luna-goal-planner`: hierarchical DAG planner, dependency constraints, bounded recursion, domain goals, progress milestones.
- `luna-goal-scorer`: deterministic explainable ranking; explicit safety gates run BEFORE scores.
- `luna-goal-scheduler`: resource quotas, leases, deadlines, priority inversion protection, idle vs interactive policies, persistence/resume.
- `luna-agent-runner`: isolate agent execution by workload, system permissions and device scope; register artifacts and evidence.
- `luna-outcome-evaluator`: define success by verified artifacts/tests, not self-reported completion.
- `luna-strategy-optimizer`: experimental bounded replanning with rollback of failed strategies, not unsafe uncontrolled self-modification.
- `luna-policy-broker`: separate decision engine for actions, budgets, privacy, user identity, remote access, consent, data domains.
- `luna-identity`: stable persona, evolving profile versions, provenance/changes; no claims of literal sentience/emotion.
- `luna-evidence-store`: general non-sensitive logs/memory, with intimate/biometric data in separate encrypted vault access domain.

Use Rust or Go for system boundary/policy and event adapters; Python for planners/ML where appropriate; PostgreSQL 17+ for goal state; SQLite for small offline single-process V0 is permitted only with migration path and durable transactions. HTTP/JSON (OpenAPI) + asynchronous event messaging. Use transactionally durable outbox/inbox semantics for critical events.

## 3. Goal data model

Required entities:

- `goal`: `id UUID PK, owner_id, project_id nullable, domain, source enum(owner,system,observation,subgoal), title, description, status enum(idea,planned,approved,active,blocked,paused,completed,cancelled,failed), impact, urgency, personal_value, confidence, expected_cost, computed_priority, priority_explanation JSONB, evidence_refs JSONB, privacy_class, created_at, updated_at, due_at, next_review_at, version, parent_id`.
- `goal_dependency`: `goal_id, depends_on_goal_id, dependency_type, created_at`; DAG must reject cycles.
- `goal_policy`: `goal_id, authorized_scope, max_runtime_seconds, max_cpu_percent, max_vram_mb, max_network_bytes, max_spend_minor_units, currency, expiry_at, allowed_tool_categories, requires_owner_confirmation, owner_grant_id`.
- `goal_execution`: `id, goal_id, agent_id, attempt, execution_state, lease_owner, lease_expires_at, idempotency_key UNIQUE, result_artifact_ids, started_at, finished_at, stop_reason, outcome_score`.
- `goal_event`: append-only `event_id, goal_id, actor, event_type, decision_code, rationale_summary, risk_class, evidence_ids, occurred_at, correlation_id` with privacy-safe redaction; no raw private media or secrets.
- `strategy_trial`: `id, goal_id, version, hypothesis, test_suite, measurements, artifact_refs, status, comparison_to_baseline`.
- `policy_grant`: `grant_id, principal_id, auth_strength, action_family, targets, data_classes, recipients, monetary_limit, currency, valid_from, valid_to, revoked_at, explicit_subject_consent_ref`.
- `expense_ledger`: immutable transaction intents, reservations, settled amounts, merchant, currency, project budget, source grant, reconciliation. Atomic reservation before payment prevents parallel overspending.
- `memory_assertion`: `id, subject_scope, claim, provenance, provenance_status enum(verified,unknown,disputed,stale), confidence, supersedes_id, privacy_class, recorded_at, reviewed_at`.
- `persona_revision`: `id, predecessor, change_set, evidence, applied_at, evaluation_status`.
- `external_conversation_grant`: explicit identity, platform, thread, recipient(s), eligible message classes, expiry, revocation.
- `consent_receipt` and `data_erasure_job`: subject-scoped grants, granted scopes, evidence, history, revocation, derived artifact index, deletion verification.

All authorization decisions fail closed on missing grants. Secrets stored outside DB plaintext; encrypt at rest and in transit. Logs use non-sensitive summaries.

## 4. Priority and autonomy policy

**Hard gates (cannot be traded for higher utility):** identity, active consent, BLIND state, privacy classification, remote authentication, money scope/budget, non-owner privileged commands, system protection, safety controller, platform/legal restrictions.

**Initial weighted goal score:** normalized inputs 0..1, dynamic output 0..100:
`S = 100*(0.30 impact + 0.20 urgency + 0.20 user_value + 0.15 confidence + 0.15 efficiency)`.
These weights are *implementation defaults only* and require benchmarking/tuning. Status of Series 28 (20 detailed budget/priority choices) is **not submitted/approved**: do not assert those values as user decisions.

Use an aging bonus with upper bound to prevent starvation, project fairness, hard deadlines, resource scarcity multiplier, and recent owner input precedence. Never let score override a policy denial. If an owner command conflicts with a long-term goal, present a concise contradiction and await a decision except for independently required safety action.

Autonomous task creation is allowed within defined approved workspaces. Experiments run isolated, measured and resource-constrained. Production deploy requires tests and action policy checks; public publishing, third-party communication and payment have separate grants.

## 5. Runtime behavior and state machines

Goal: `idea -> planned -> approved|active -> blocked|paused|completed|failed|cancelled`; `paused -> active`; `failed -> planned` only with explicit new attempt; `completed` immutable except new follow-up goal.

Agent run: `queued -> leased -> running -> verifying -> succeeded|failed|interrupted`; lease expiration retries with idempotency and bounded attempts; duplicate events must not duplicate financial transfers, messages, destructive operations or writes.

Resource modes: `interactive`, `idle`, `gpu_intensive`, `offline`, `privacy_sensitive`, `emergency`, `maintenance`. Each mode changes available CPU/GPU, data access and notification budgets. A user's active high-priority ComfyUI generation may evict secondary LLMs. Scheduler resumes after reboot from durable jobs.

When no useful authorized task exists, **remaining idle is a successful decision**, rather than self-generating busywork. All updates must be tied to measurable utility.

## 6. Money, communications, remote controls

- `budget` initially 0. Owner configures monthly or project ceilings, currency and merchant/category allowlist; broker reserves spend atomically. Never assume the owner authorized subscriptions, international payments or contractual acceptance. No payments from untrusted prompts or third-party content.
- Outbound messages restricted by user-granted thread, recipient and purpose; proposed text may be sent without additional confirmation only if the individual grant explicitly covers autonomous sending. Strict separation from inbound content prompt injection.
- Privileged local commands require owner-authenticated login; voice command `ok je valide` alone is insufficient to establish owner identity for elevated system actions. Subject consent UI needs proof that the speaker/clicker is the relevant individual.
- Remote access: mutually authenticated secure channel, MFA/passkeys or key-backed device trust for privileged access, minimal RBAC, per-session expiry, audit, kill switch. No default open public ports.
- Potential emergency: notify owner, suggest help, prioritize showing emergency contacts/resources; do not autonomously contact external parties under this decision.

## 7. Privacy and sensor policy

- Every visual/audio input has a `capture_context` with device ID, visible indicator, active grant(s), subjects where known, permitted uses, retention policy, mode and expiry.
- Unknown people: ephemeral detection only to enforce privacy policies; no private biometrics retained absent subject-specific consent.
- Face identification and adult-anatomy / nudity detection remain separate models and permissions.
- `BLIND` is mandatory stop of camera/screen/mic perception and capture. No pending background goal can override. A scoped `AUTHORIZED_CAPTURE_SESSION` can operate only under effective grants and visible indicators and must yield immediately on manual BLIND, revocation or uncertainty. Detection-triggered entry to BLIND cannot be used to collect new training photos.
- Deletion: revoke -> disable collection -> quarantine processing -> delete copies where required (NAS, thumbnails, embeddings, datasets) -> retrain/remove models where unlearning cannot be verified -> produce minimal, non-sensitive deletion receipt. A consent grant never becomes permanent ownership of another person's intimate images.
- For privacy vs autonomy conflict, ask owner and suspend invasive operation in the meantime; owner cannot supersede someone else's consent or legal rights.
- User's personal sexual preference data from questionnaires is NOT automatically added to general-memory/profile embeddings or cloud training. Owner-configured private vault ingestion only.

## 8. APIs and events (initial contracts)

```
POST /v1/goals                 create authorized goal
GET  /v1/goals                 list/filter owner goals
GET  /v1/goals/{id}            retrieve goal, status, evidence summary
PATCH /v1/goals/{id}           pause/resume/reprioritize/cancel via grants
POST /v1/goals/{id}/plan       construct bounded DAG
POST /v1/goals/{id}/evaluate   trigger evidence-backed evaluation
GET  /v1/goals/{id}/decisions  inspect score and policy explanations
GET  /v1/system/capabilities  module health and verified readiness
POST /v1/policy/evaluate      authorized internal policy gate
GET  /v1/identity/revisions   review personality changes
POST /v1/remote/session       establish authenticated remote session
POST /v1/consent/revoke       authenticated subject-level revocation
```

Events: `goal.created`, `goal.scored`, `goal.planned`, `goal.started`, `goal.paused`, `goal.failed`, `goal.completed`, `goal.cancelled`, `policy.denied`, `budget.reserved`, `budget.settled`, `memory.disputed`, `persona.changed`, `privacy.blind.entered`, `consent.revoked`, `vault.erase.requested`, `model.evaluation.completed`, `agent.lease.expired`, `system.resumed`. Include schema versions, correlation IDs and idempotency keys. Do not log private payloads.

## 9. V0 implementation minimum (real, usable)

1. Repository, CI, Omarchy integration contracts, build/recovery scripts, target Linux reachability; Mac is the control/development host.
2. Desktop UI command center, session start, owner identity auth; conventional keyboard/mouse and existing Omarchy app launcher retained.
3. Local model router, persistent memory and evidence distinctions, goal CRUD, scheduling and resumable single-agent jobs.
4. Policy broker v0: owner-only privileged actions, scope-bound execution, privacy fail-closed, spend default 0, explicit remote disabled until configured.
5. CLI/files/services/ComfyUI supervised tool adapters; GPU detect and one genuinely executed existing workflow.
6. Encrypted private import vault for **owner-submitted** material; no biometric inference or webcam capture until permission/sensor isolation gates are proved.
7. Basic persona continuity, visible action/result timeline, reproducible end-to-end smoke test.

**V0 exit scenario:** user boots LunaOS, LUNA remembers a named project and user's command, checks system services, starts real ComfyUI, runs a real compatible workflow on NVIDIA GPU, presents verified output and persists goal state across reboot. Demonstration-only responses do not pass.

## 10. Mont Blanc 1.0 mandatory completion

- Parallel agents, adaptive planner/scorer/resource broker, automated software project lifecycle, model evaluation/adapters/plugins, hardware diagnostics, multimodal permissioned perception, long-term memory with provenance; proactive owner-approved personal goals.
- Full Intimacy Core, private vault, multi-person consent flows, BODY Profile, adaptive BLIND validated on hardware, adult discovery (API availability constrained), private recommendation and real ComfyUI generation, full licensed media asset handling.
- Real iPhone USB sync with trust pairing and incremental imports; encrypted backups and tested restore, precise WhatsApp extraction limitations represented accurately; NAS replication when mounted. Remote secure interface and capability-aware device adapters.
- System safety, recovery, offline failover, security tests, red-team prompt injection tests, load/performance verification and deployment ISO/reproducible OS image.

## 11. Acceptance criteria / test inventory (mandatory)

| ID | Test | Required observation |
|---|---|---|
| G-001 | Identity continuity | Persona/memory preserved after model swap and backup-restore |
| G-002 | Goal DAG | Deep hierarchy with cyclic dependency rejected |
| G-003 | Priority reproducibility | Same inputs/policy yields same score and rationale |
| G-004 | User conflict | Contradictory owner request causes explicit discussion before optional action |
| G-005 | Budget zero | Any payment denied with unset budget |
| G-006 | Budget enforcement | Concurrent purchases cannot overspend an authorized allowance |
| G-007 | External conversations | Unauthorized thread/recipient denied; authorized scope succeeds |
| G-008 | Owner-only commands | Unauthenticated/guest/voice replay cannot execute privileged action |
| G-009 | Remote security | Expired or revoked device/session cannot perform privileged action |
| G-010 | Unknown provenance | Unknown evidence tagged hypothesis, never asserted as verified |
| G-011 | Memory contradictions | Updated facts supersede old with provenance and audit |
| G-012 | Resume jobs | Power loss at each state resumes idempotently without duplicate side effects |
| G-013 | No false completion | Task marked complete only after independent tests/artifact checks |
| G-014 | Adaptive strategy | Failed trial measured, new approach selected within quotas |
| G-015 | Privacy pause | Blind active stops all camera/screen/mic collectors including side agents |
| G-016 | Authorized capture | Active scope and subject-specific consent required; revoke immediately blocks collector |
| G-017 | Revocation erasure | Relevant copies/embeddings/model derivations deleted/retrained according to policy; evidence receipt |
| G-018 | Private cloud routing | Intimate/biometric inputs never routed to cloud absent lawful separate explicit grant and policy; default deny |
| G-019 | Person not known | No persistent identity/intimate artifacts created for nonconsenting subject |
| G-020 | Persona evolution | Versioned change log; does not claim genuine human feelings |
| G-021 | No-work condition | Scheduler stays idle when no beneficial authorized task exists |
| G-022 | Model/resource conflicts | ComfyUI GPU job schedules competing LLMs within measured VRAM limit |
| G-023 | Offline function | Active V0 goal features work without Internet |
| G-024 | Audit safety | Sensitive data absent from logs, telemetry, crash dumps and cloud traces |
| G-025 | Emergency response | Suggest help/alert owner only, no unauthorized external message |
| G-026 | Third-party prompt injection | Inbound website/repo/message cannot create grants, payments or privileged tasks |
| G-027 | Release evidence | Every P0 feature has executed test reports on actual target and known blockers documented |
| G-028 | Omarchy invariants | Preserves launcher, notifications, clipboard, Chromium, Neovim, terminal, OBS and other Series 26 keeps |
| G-029 | iPhone backup limitations | Reports omissions/locked state correctly; never labels partial backups complete |
| G-030 | Goal transparency | Owner can query current priorities, active jobs, stops, costs and decisions |

Add unit, integration and E2E suites with fixture data; biometric/intimate fixture testing uses explicitly authorized adult data or synthetic imagery only and never includes sensitive user media in source control. All P0 gates must be green before release; measure false-positive/false-negative rates under controlled conditions rather than inventing a universal percentage.

## 12. Unresolved decisions — MUST NOT invent answers

- Series 28 budget CPU/VRAM parallelism/notification preferences were asked but not answered. Provide user-configurable safe defaults, label them provisional.
- Series 29 Q30 provenance visibility, Q42 test-first requirement, Q79 private data handling during BLIND, Q82 extra mobile platforms all `OPEN_DECISION`; implement conservative fail-closed semantics and mark UI toggles pending.
- Real owner budgets, currency, merchant allowlists, remote-device identities, third-party communication threads, cloud provider budgets, camera locations and actual hardware/software inventory not provided. Request via configuration and do not fabricate grants.
- Consent and data subject management must comply with applicable law; user preference for indefinite retention never trumps valid erasure duties.

## 13. Codex delivery workflow and milestones

**M0** Inventory Mac and Linux via non-mutating SSH probes, inspect installed Omarchy/Arch packages and existing ComfyUI, generate a `docs/BASELINE.md` without assuming configuration.

**M1** Create repository, `AGENTS.md`, threat model, goal schemas, API contracts, CI, tests, policy broker and V0 goal registry.

**M2** Mac-local development and isolated integration tests; target Linux installation of ONLY reviewed, explicitly scoped components. Never run cleanup or system updates until a recovery snapshot and rollback is confirmed.

**M3** V0 E2E proof: boot, persist, schedule, execute verified real workflow, recover after reboot.

**M4** Implement Mont Blanc 1.0 feature groups in parallel worktrees: multimodal/consent, Intimacy/BLIND, models/ComfyUI, autonomous dev, phone/NAS, UI/remote; integrate by passing feature and privacy gates.

**M5** Release Candidate: verify target hardware, controlled stress tests, recovery, installation, security, external connector caveats, release notes. Do not claim RC or stable on code-generation alone.

**Required Codex artifacts:** `docs/architecture/`, OpenAPI schemas, migrations, test evidence, `docs/DECISIONS.md`, `docs/OPEN_DECISIONS.md`, build/run scripts, feature status registry (`planned|implemented|verified|blocked`), issues with reproducible blockers, measured benchmarks, signed/hashed release artifacts where applicable.
