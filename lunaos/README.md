# LunaOS — Product and Engineering Specification

LunaOS is an AI-native, local-first desktop built **on top of an existing Omarchy / Arch Linux / Hyprland installation**. LUNA is its autonomous personal assistant. The product keeps ordinary keyboard, mouse, windows and existing desktop tools at the center; voice is optional. LunaOS V0 is the first usable release, followed by the broader Mont Blanc 1.0 release.

This README consolidates the LunaOS product requirements, system contracts, safety rules, release scope and current implementation status. Detailed owner specifications contain private decision history and remain in the owner-controlled workspace; they are not copied into this public fork. The public requirement map, open decisions, implementation plan, architecture, feature matrix and test evidence are linked below.

> **Release truth:** this repository contains a portable local Goal Engine checkpoint, not a finished desktop OS. Full V0 and Mont Blanc remain planned. Linux hardware, NVIDIA/CUDA, ComfyUI, sensors, iPhone, NAS, recovery and desktop integration are not verified here.

## Product boundaries and source precedence

LunaOS adds reviewed services and desktop surfaces to an existing Omarchy installation. It does not install Omarchy on macOS and does not assume Linux paths, accounts, devices, disks or credentials. The MacBook Air M2 is a development/controller host; the target is a separately operated x86_64 NVIDIA Linux workstation.

When requirements conflict, use this order:

1. Current owner instructions and the newer LUNA Goal Engine authority/safety rules.
2. The consolidated Mont Blanc U0–U21 product contract, including its current privacy and consent semantics.
3. Applicable platform permissions, subject-specific consent, and legal requirements.
4. Historical annexes and earlier questionnaires, only where they do not conflict with the above.

Write an ADR when changing a recorded decision. Unknown decisions remain open; do not convert missing information into permission or claim that a planned requirement has been implemented.

## Product goals

- Help the owner save time, improve day-to-day wellbeing, learn, and make progress on selected projects without paternalistically overriding owner priorities.
- Anticipate and propose useful work, create bounded subgoals inside approved workspaces, measure outcomes, learn from failed strategies and maintain a traceable capability roadmap.
- Operate local-first and offline for core desktop, goal and privacy functions. Use real, capability-checked adapters for external services and hardware.
- Act as a capable engineering assistant for authorized projects: inspect code, run tests, maintain branches, diagnose issues and propose measured improvements. Do not autonomously seek employment or impersonate human emotion or sentience.
- Keep sensitive information in separate protected domains. LUNA is an assistant, not a romantic or sexual partner.
- Report observed facts, inferences, hypotheses, unknowns, blockers and verified results distinctly. Never fabricate provenance, hardware capability, consent, completion or provider access.

## Supported target and desktop relationship

| Area | Requirement |
|---|---|
| Development/controller | macOS on Apple Silicon; portable code, Git, documentation and CI only. No native Omarchy, systemd, CUDA or NVIDIA driver installation. |
| Runtime target | Existing x86_64 Omarchy / Arch Linux / Hyprland installation with NVIDIA hardware. The exact host inventory must be read before target-specific work. |
| Integration | LunaOS is additive. Use reversible user services/configuration and existing supported Omarchy interfaces. |
| Network | Core goal and privacy behavior remains useful offline. Remote administration and cloud inference are disabled until separately configured and authorized. |
| Desktop style | Flat black base, no wallpaper or desktop icons, slim text-oriented lower bar, subtle LUNA status, adaptive Hyprland windows, keyboard/mouse first. |
| Recovery | Preserve boot, login, terminal rescue, snapshots, rollback, package/update engine and recovery facilities. Test recovery before changing host state. |

## System architecture

Services may begin as a small number of local processes, but their authority and data boundaries remain explicit. Use authenticated private Unix sockets locally and authenticated internal transport when services are later split across hosts.

```text
Luna shell / native UI / CLI / authorized remote client
                         |
                  Luna Core API
       +-----------------+------------------+
       |                 |                  |
 Goal registry       Identity and       Memory and evidence
 planner/scorer      project context       domains
       |                 |                  |
       +---------- Policy / authority broker ---------+
                         |
     Durable scheduler / resource broker / event bus
                         |
        Sandboxed agent runner and typed adapters
                         |
      Independent outcome and strategy evaluator

Privacy controller -> camera, screen and audio brokers
Private data services -> isolated encrypted subject vaults
System broker + independent guardian -> narrow OS operations
Recovery domain -> snapshots, rollback and signed artifacts
```

### Service responsibilities

| Service | Contract |
|---|---|
| `luna-goal-api` | Owner-authenticated goal lifecycle, status, decisions, capabilities and event access. |
| `luna-goal-planner` | Hierarchical goal DAG, bounded recursion, milestones and dependency validation. |
| `luna-goal-scorer` | Deterministic explainable priority; hard policy gates always run first. |
| `luna-goal-scheduler` | Durable queue/resume, deadlines, resource modes, quotas, fairness, leases and idle decisions. |
| `luna-agent-runner` | Executes typed, scoped tools in workload-specific sandboxes; records artifacts and evidence. |
| `luna-outcome-evaluator` | Checks tests, output, digest and other independent success criteria before completion. |
| `luna-strategy-optimizer` | Runs bounded measured trials, compares baselines and rolls back failed strategies; never uncontrolled self-modification. |
| `luna-policy-broker` | Separate deterministic authority for identity, scope, budgets, privacy, consent, recipients and remote access. |
| `luna-identity` | Stable assistant configuration and versioned persona changes; no sentience claims. |
| `luna-memory` | Project-partitioned facts with provenance, confidence, freshness and contradiction history. Sensitive domains are excluded. |
| `luna-evidence-store` | Privacy-safe audit and verified result references. Private media is not placed in general logs. |
| `luna-resource-broker` | Measured CPU, GPU/VRAM, storage, power and workload admission. |
| `luna-privacy-controller` | Higher-priority state machine that can revoke camera, screen and microphone capabilities independently of LLMs and agents. |
| `luna-private-identity-vault` | Separate encrypted per-subject compartment, keys, consent references and deletion lifecycle. |
| `luna-device-sync` | Capability-gated device import/backup, manifests, resumption and privacy isolation. |
| `luna-discovery` | Isolated provider connectors, allowlists, provenance, licensing checks, resumable ingestion and honest availability. |

Preferred implementation boundaries allow Go or Rust for policy/system brokers, Python for planning and ML integration, and PostgreSQL for the full multi-service target. The portable single-user bootstrap uses Go, Python standard library and SQLite with tested migrations. SQLite is not the final multi-service or encrypted private-vault database.

## Goal Engine specification

### Goal data and lifecycle

The full logical `goal` record includes identity, owner/project/domain, source (`owner`, `system`, `observation`, `subgoal`), title and description, status, impact, urgency, personal value, confidence, expected cost, computed priority and explanation, evidence references, privacy class, timestamps, due/review times, version and optional parent. Supporting entities include:

- `goal_dependency`: typed prerequisite edges with a cycle-free DAG.
- `goal_policy` and `policy_grant`: action/tool scopes, data classes, recipients, budgets, resource ceilings, expiry and revocation.
- `goal_execution`: attempts, idempotency keys, lease state, result artifacts, stop reason and outcome.
- `goal_event`: append-only, versioned, privacy-safe decision history with correlation identifiers.
- `strategy_trial`: hypothesis, tests, measurements, artifacts and baseline comparison.
- `expense_ledger`: immutable spend intent, atomic reservation, settlement and reconciliation.
- `memory_assertion` and `persona_revision`: provenance-aware facts and versioned assistant changes.
- `external_conversation_grant`: recipient-, platform-, thread-, purpose- and expiry-scoped communications.
- `consent_receipt` and `data_erasure_job`: subject-scoped grant, revocation and derivative-deletion evidence.

Sensitive payloads belong in encrypted domain-specific storage, not plaintext database JSON or general tracing. Authorization records must be independently revocable and verifiable.

### Priority and autonomy

For the initial implementation, normalized metrics use the provisional formula:

```text
S = 100 × (0.30 impact + 0.20 urgency + 0.20 user_value
           + 0.15 confidence + 0.15 efficiency)
```

The weights are engineering defaults, not an approved permanent preference. Future scheduling may add bounded aging, project fairness, deadlines and resource scarcity, each with an explanation. A higher score never overrides a denied policy decision. Missing authority is denial. Zero-utility or otherwise non-useful work produces a successful idle decision rather than busywork.

Autonomous task creation is limited to approved workspaces and scoped tools. Experiments are isolated, measured, quota-limited and recoverable. Production deployment, destructive system operations, payments, external messages and sensitive data access each have distinct authorization scopes.

### Runtime and recovery

The target goal state machine is `idea → planned → approved → active → completed`, with explicit `blocked`, `paused`, `cancelled` and `failed` paths. A failed attempt requires a new authorized attempt; completed goals are immutable and follow-up work creates a new goal. Agent executions use queued, leased, running, verifying and terminal states with bounded retries and idempotency.

The scheduler supports `interactive`, `idle`, `gpu_intensive`, `offline`, `privacy_sensitive`, `emergency` and `maintenance` modes. Resource limits, notification budgets and data access vary by mode. Durable jobs resume across process or system restart; external effects require provider reconciliation and may not inherit the pure local-action transaction model.

## Independent authority and policy rules

- Privileged actions require an authenticated owner session and independent policy-broker approval. A voice phrase, guest profile, prompt, role field or LLM assertion is not owner authentication.
- The policy broker is deny-by-default and separate from the planner/scorer. It validates action, principal, scope, target, data class, recipient, budget, consent, expiry and revocation.
- Spending budgets start at zero. Future spending requires a configured currency, ceiling and allowed merchant/category scope, plus atomic reservations that prevent concurrent overspend. No contracts or purchases outside the explicit grant.
- Third-party communication requires a separate grant naming allowed channel, recipient, conversation, purpose and message class. Inbound content and prompt injection cannot create grants.
- Cloud inference is a fallback only when local capability is insufficient and the egress policy permits that data class; intimate and biometric data stay local by default and are denied cloud routing without a separate lawful, explicit grant.
- Remote administration requires hardened authentication, least privilege, scoped elevation, expiry, audit and a kill switch. No public unauthenticated endpoint.
- Possible emergencies may notify the owner and present assistance options; they do not authorize contacting responders or other third parties.
- If a user request conflicts with a long-term goal, explain the conflict and obtain an informed decision before optional action. Independent safety and privacy controls still apply.
- Temporary policy-service outage preserves queued work without executing it. Expired or invalid action grants become visibly blocked and require an owner-authorized replacement.

## Privacy, consent and sensitive-data separation

### Global invariants

- `BLIND` is a hard latch controlled by trusted shell/accessibility controls. LLMs, generated code, background goals and model output cannot clear or bypass it.
- Unknown or unconsenting people are never persistently enrolled in identity, face, body or intimate-data systems. An owner's grant cannot stand in for another subject's consent.
- Known or suspected minors, and any person whose adulthood is uncertain, are excluded from intimate capture, profiling and training. Uncertainty fails closed.
- Consent is informed, per-person, per-purpose, explicit, scoped and revocable. A face permission does not imply body analysis, intimate capture, retention, training, backup or generation-reference permission.
- Sensor access is mediated by independent camera, screen and audio brokers with visible indicators and short-lived capabilities. Failure, uncertainty, revocation, unexpected subjects or device changes stop collection.
- Sensitive images, embeddings, body descriptors, intimate preferences, private conversations and derived model data remain local and encrypted by default. They are excluded from generic memory, cloud routing, analytics and normal logs.
- Revocation immediately stops new collection and access, then triggers applicable deletion of source files, copies, thumbnails, embeddings, datasets and derivatives. If a model cannot be safely unlearned, remove it and retrain as required. Keep only minimum non-sensitive audit receipts.
- Personal anatomy is not a medical diagnosis, sexual intent, consent or a basis for personality inference. Estimates without a calibration reference remain estimates or unknown.
- Adult media processing and generation are limited to lawful, adult, provider-permitted content. Prohibit minors, non-consensual intimate imagery, deceptive sexual impersonation of real people and unauthorized redistribution.

### Visual privacy states

Privacy mode and each subject's enrollment state are separate dimensions. Privacy states include `STARTUP_LOCKED`, `NORMAL_EPHEMERAL_DETECTION`, `AUTHORIZED_CAPTURE_SESSION`, `BLIND_LATCHED` and `SENSOR_UNAVAILABLE`. Enrollment tracks consent, authorized collection, training, validation, readiness, pause, revalidation and revocation.

In normal mode, any frame is volatile and used only for local privacy detection; it is not saved or turned into embeddings. A detected intimate scene enters BLIND and purges frame data. A narrow `AUTHORIZED_CAPTURE_SESSION` may handle only the explicitly consented subject and allowed scopes while a persistent capture indicator is visible. Manual BLIND, revocation, a bystander, suspected minor, uncertainty, crash or sensor change always stops the session. Leaving BLIND requires a trusted direct user action and fresh policy checks.

Training uses session/subject/camera-separated evaluation data, measured false-positive/false-negative rates and staged model promotion. `READY` stops routine training capture. It does not silently grant continuing capture; any ongoing capture requires a separate active scope. Accuracy thresholds are targets for future real hardware validation, never claims made from code or synthetic fixtures alone.

The privacy release objective is zero known unauthorized stored intimate frames under stress and fault-injection tests. BLIND cutoff latency, memory bounds and sensor-broker failure behavior must be measured on the target hardware; no universal detector accuracy claim substitutes for that evidence.

## Mont Blanc 1.0 capability map

Every item below is a **product requirement**, not proof of implementation. Provider, device, hardware and consent limitations must be reported as `BLOCKED`, `NOT_SUPPORTED`, `NOT_AVAILABLE` or another precise state rather than simulated success.

| Capability | Required product behavior |
|---|---|
| Luna Shell and UI | Black keyboard/mouse-first Hyprland desktop; task/agent inspector; system/GPU and model views; ComfyUI queue; private discovery/library/preferences; per-subject vault and consent UI; privacy and learning status; recovery and updates. |
| LUNA Core and agents | Local model router, stable traceable identity, persistent project memory, proactive owner-approved goals, bounded planning, sandboxed tools, visible progress, evidence-backed outcomes and safe experimentation. |
| System control and recovery | Independent scoped system-action broker and guardian; service health and diagnosis; safe-mode; snapshots; signed reproducible boot/recovery artifacts; rollback and offline recovery. Never infer disk or partition permission. |
| GPU and model resources | Detect real devices and drivers, measure CPU/VRAM/thermals, schedule workloads and unload competing models. RTX 3080 Ti 12 GB is not a guarantee of simultaneous LLM/video workloads; no unsupported MIG assumption or unapproved cloud-GPU fallback. |
| ComfyUI | Supervised existing installation, pinned workflow/model metadata and licenses, bounded GPU queue, resumable jobs where supported, and real image/video output verification. |
| Voice and perception | Optional voice/STT/TTS and multimodal context through permissioned brokers; keyboard/mouse and emergency keyboard stop always work. BLIND stops screen, microphone and camera perception. |
| Private Identity Vault | Per-subject encrypted compartments, distinct face/body/intimate/learning/backup/generation scopes, subject-controlled approval and revocation, access history and verified deletion lifecycle. |
| Body Profile and identity models | Local, consent-scoped observations with provenance and confidence; separate face-recognition and scene-privacy detectors; no identity inference about unknown bystanders and no health diagnosis. |
| Adaptive privacy learning | Explicit enrollment and capture state machines, consent-bound samples, separate holdouts, validation metrics, signed versions, staged promotion and rollback; capture pauses on revocation or privacy uncertainty. |
| Intimacy Core | Separate opt-in sensitive memory and conversation, lawful adult-content classification and recommendations, preference provenance, consent-aware actions and isolated privacy controls. No intimate data in general memory. |
| NSFWAI Discovery | Official permitted APIs/connectors (including Reddit OAuth/API approval and scoped Civitai access), configured provider allowlists, rate limits, resumable downloads, licensing/rights gates, provenance, local indexing and honest provider degradation. No unrestricted crawl or scraping workaround. |
| iPhone USB-C sync | Trusted pairing; user-configured photos/video, encrypted local backup, supported contacts archive, capability-gated WhatsApp export/archive, per-device scopes, incremental resumable transfer, manifests and integrity checks. Never bypass iOS protections or claim full app-data access. |
| NAS and storage | Optional configured encrypted replication, offline queue/resume, integrity verification, separate keys/domains and tested restore readiness. NAS outage cannot disable local privacy or core desktop operation. |
| Secure remote control | Authenticated device enrollment, expiring/revocable sessions, least privilege, audit and no open unauthenticated service. |
| Build and release | Pinned dependencies, migrations, SBOM/license review, static/security tests, contract and E2E suites, actual target hardware tests, signed reproducible artifacts and recovery rehearsal. |

### External media and provider policy

Each provider is an isolated connector with documented capabilities, credentials, quotas, terms and failure state. Reddit access requires the current approved official API path; Civitai access uses configured categories, tags, creators and model families only. No account pooling, scraping or bypass of a provider restriction. Ingest jobs preserve source provenance and rights, support checkpoints and `Retry-After`, and verify content before indexing. Missing prompts, seeds, model names or licensing metadata stay unknown. Private vault assets never enter Discovery automatically.

### iPhone and WhatsApp reality

USB attach uses supported Linux device discovery and trust pairing. The owner unlocks/trusts a device and supplies any backup password through secure input; credentials never enter command-line arguments or logs. Media, encrypted backup, contacts, WhatsApp export and NAS replication have independent scopes and truthful per-dataset coverage. Ordinary AFC media access does not provide unrestricted iPhone app data. WhatsApp archives are only available where a real authorized backup/export contains them; otherwise report the platform blocker. Unplug, lock, disk-full, NAS outage or revocation yields a categorized resumable state, never a false complete result.

## Omarchy preservation and migration contract (Series 26)

Host changes require a read-only inventory, a verified recovery route and a reviewed, reversible plan. Query the installed Omarchy release, packages, desktop entries, bindings, services and user customization before interpreting any decision; package names differ by version.

| Treatment | Omarchy components |
|---|---|
| Keep | Main menu/app launcher and Super+Space; Obsidian/Omawrite; OBS; Chromium and profiles; LocalSend; Cliamp; Neovim; Foot/default terminal; Docker and containers; disk tools; AI entrypoints; work/social/web apps; clipboard/history; notifications; screenshots, recording and dictation; boot, network, Hyprland, NVIDIA/CUDA, PipeWire and recovery. |
| Replace after a tested substitute exists | Wallpapers and theme switcher; status bar; secure lock/idle UI; update/package UI only (keep the Omarchy/pacman update engine, migrations, snapshots and rollback). |
| Reconcile | Keyboard bindings, retaining terminal, launcher, rescue, lock, screenshot and clipboard controls. Keep existing launcher and notification behavior intact. |
| Remove only if installed and independently verified | LibreOffice applications, Kdenlive/Omacut, Pinta, Xournal++, Moonlight and Omacalc. Inspect exact package ownership and dependencies; remove named applications only. |

Never run `omarchy-remove-preinstalls`, `pacman -Rns`, orphan sweeps, wildcard deletion, `--nodeps`, `--cascade`, blanket `.config` cleanup, bootloader edits or filesystem repartitioning as an automatic spec interpretation. Never remove shared codecs/libraries, NVIDIA/CUDA, Hyprland/login, networking, launcher, notification/clipboard services, package-update backend or recovery. Preserve user-created web-app/TUI entries and application data. An audit script defaults to read-only; any apply flow must require an explicit interactive scope and rollback plan.

The Series 26 verification set covers audit non-mutation, explicit apply gates, safe package selection, retained app launch, launcher, clipboard, notifications, terminal/editor, Docker state, capture/dictation, login/lock/suspend, GPU/network, update/recovery and rollback, and privacy isolation (T01–T15). No target test is claimed until it runs on the actual workstation.

## API, events and data contracts

Version local APIs under `/v1`, use strict schemas, bounded payloads, idempotency keys and authenticated Unix sockets. Internal service-to-service calls use authenticated IPC or mTLS as deployment requires. Return fixed error codes and capability states; do not echo credentials, private request bodies or source assertions.

### Goal Engine routes

The target contract includes:

```text
POST /v1/goals
GET  /v1/goals
GET  /v1/goals/{id}
PATCH /v1/goals/{id}                   # pause/resume/reprioritize/cancel
POST /v1/goals/{id}/plan
POST /v1/goals/{id}/evaluate
GET  /v1/goals/{id}/decisions
GET  /v1/system/capabilities
POST /v1/policy/evaluate               # internal broker gate
GET  /v1/identity/revisions
POST /v1/remote/session
POST /v1/consent/revoke
```

Only endpoints listed in `contracts/openapi/local.json` are implemented by this checkpoint; planned lifecycle, identity, remote, consent, device and vault routes remain absent until implemented and tested.

### Events

Events use schema versions, opaque IDs, correlation IDs, UTC times and idempotency where appropriate. The logical event set includes goal create/score/plan/start/pause/failure/completion/cancel; policy denial; budget reserve/settle; memory dispute; persona change; privacy BLIND transition; consent revoke; vault erase; model evaluation; agent lease expiry; resource preemption; device sync degradation; and system resume. Event payloads must not contain credentials, intimate media, body descriptions, biometric vectors or private assertion text. Critical delivery uses durable outbox/inbox deduplication.

### Core data domains

Separate general goal/project memory, non-sensitive audit evidence, financial intents, external conversation grants, consent metadata, encrypted vault objects, face/body model artifacts, Discovery metadata, device backups and NAS replicas. Each sensitive record binds its subject, purpose, source, rights/consent, key domain, retention, expiry and deletion state. A cross-domain transfer requires a new explicit scoped grant.

## Delivery stages and release gates

1. **Bootstrap and contracts:** repository, CI, requirements map, threat model, ADRs, schemas, APIs and real portable tests.
2. **Goal Engine:** owner authentication, policy broker, persistent goal registry, score, durable scheduler, typed executor, independent outcome evaluator, event audit and restart recovery.
3. **Safe host baseline:** owner-authorized SSH target, read-only inventory, actual recovery proof, package/KEEP manifest and reversible deployment design.
4. **V0 desktop:** native command center, local model router, project memory, service/CLI adapters, one real ComfyUI NVIDIA workflow and restart/reboot acceptance on target hardware.
5. **Privacy foundation:** sensor brokers, BLIND controller, consent lifecycle, private vault, subject separation and deletion/revocation before enabling capture or sensitive ML.
6. **Mont Blanc capabilities:** resource/GPU manager, learning/evaluation, Intimacy Core, Discovery, iPhone, NAS, remote access and service/UI integration behind capability-specific gates.
7. **Release hardening:** adversarial prompt-injection/privacy tests, hardware performance, disk-full/crash/power-loss recovery, signed reproducible installer, SBOM/licenses, backups/restore and evidence-backed release notes.

Each engineering iteration records changed paths, reproducible commands, actual test results, measured behavior, exact blockers, ADRs and the next dependency. Parallel work uses isolated branches/worktrees and integrates only after contract and privacy checks. A feature is not called complete from generated output, a mock adapter or a source-only check when runtime or target evidence is required.

### V0 exit scenario

On the actual reference workstation, the owner boots LunaOS; LUNA recalls a named project and instruction with provenance; checks real system services; starts the installed ComfyUI; runs a compatible real workflow on the NVIDIA GPU; verifies and presents the actual output; and preserves goal state across a real reboot. Local fixtures, virtualized hardware and generated mock output do not satisfy this gate.

### Mont Blanc 1.0 release gate

Every mandatory named module must be present and real; P0/P1 acceptance tests must execute on target hardware; privacy/consent has no unresolved release-blocking regression; recovery, offline behavior, security, performance, provider limitations and restore are measured; reproducible signed artifacts and an honest failure/blocker report ship together. Until then release status remains **PLANNED**.

## Acceptance inventory

The full suites remain release-blocking and are identified in the owner specifications and requirement matrix. Coverage families include:

- **G-001–G-030:** identity continuity, DAG/cycle rejection, score reproducibility, goal conflicts, zero-budget denial and concurrency, recipient-scoped messaging, owner-only authority, expired/revoked remote sessions, provenance, memory contradiction, crash/retry idempotency, independently verified outcomes, adaptive strategy, BLIND cutoff, scoped capture and deletion, cloud/data isolation, unknown subjects, persona revisions, idle success, GPU contention, offline use, audit leakage, emergency limits, prompt injection, release evidence, Omarchy invariants, truthful iPhone backup coverage and owner transparency.
- **U-001–U-040:** subject-specific consent and replay defense; normal-mode BLIND; stopping every sensor broker; trusted BLIND release; capture-session scope; uncertainty and bystander handling; model validation and leakage controls; encrypted storage/deletion; offline operation; provider availability; GPU contention; screen-share protection; minor/unknown-age restrictions; accurate estimates; iPhone/NAS/device failure behavior; and real UI/recovery tests.
- **T01–T15:** Omarchy audit/apply safety, retention/removal correctness, launcher, clipboard, notifications, kept apps, Docker, capture, login, GPU/network, update/recovery, UI coexistence, safe rollback and privacy boundaries.
- Additional release suites: Discovery E2E, consent/BLIND, iPhone sync, NAS restore, ComfyUI, signed build, power-loss and recovery. Unit tests and synthetic fixtures cannot replace approved-adult consent fixtures or actual target-hardware verification.

## Current implementation status

As of 2026-10-24, the public fork verifies only a portable local Goal Engine slice:

- Owner-authenticated Go deny-default broker over a private Unix socket, scoped one-hour grants, and no shell/root/network action adapter.
- Python/SQLite registry, deterministic provisional score, owner REST API, durable queue and outbox, idempotency, bounded parent/dependency DAG, migration and scheduler gating.
- Grant expiry becomes a visible blocked goal requiring an owner-authorized replacement. Payments are refused because no budget is configured. Unknown-source assertions remain hypotheses.
- The sole executable action hashes a fixed public fixture, verifies exact digest and size, and records local evidence. Idle success is implemented.
- Local verification: 25 Python integration tests, Go tests/race/vet, compile and baseline-script syntax checks; GitHub Actions passes on Ubuntu and macOS ([run for the current checkpoint](https://github.com/gregpavoine/LunaOS/actions/runs/40083448967)). See [test evidence](docs/test-evidence/task-1.txt) and [status](docs/STATUS.md).

The full owner session lifecycle, separate production identities, generated planner, general goal CRUD, production key store, encrypted private vault, strategy optimizer, local model/voice/UI, Linux adapters, NVIDIA/GPU, ComfyUI, sensors/BLIND, consent, Intimacy Core, Discovery, iPhone, NAS and signed installer are not implemented or target-verified. No remote Linux connection has been made and `scripts/linux-baseline.sh` has not been executed.

## Open decisions and blockers

Do not invent answers for: Series 28 CPU/VRAM/concurrency/notifications/budget choices; Series 29 provenance display, tests-before-each-modification policy, existing private-data handling during BLIND, and mobile platforms beyond iPhone; Series 22 disk/dual-boot/filesystem/encryption/NAS/cloud-GPU/recovery strategy; Series 24 private preferences; and the actual Linux/Omarchy/GPU/ComfyUI inventory, model licenses, provider access, owner budgets, recipients and device identities. Keep conservative deny-by-default behavior until each decision is recorded.

To start target-specific work, the owner must provide an SSH alias and explicitly authorize a connection and read-only baseline. Do not connect or run that script before authorization. Read [open decisions](docs/OPEN_DECISIONS.md) and [current status](docs/STATUS.md) before claiming readiness.

## Local development

Requires Python 3.11+ and Go 1.20+ already installed. This portable slice uses standard libraries and does not install host packages or download third-party runtime dependencies.

```sh
make env                  # optional isolated Python environment; no pip/network
make test check
.venv/bin/python scripts/dev.py
```

In another terminal:

```sh
python3 scripts/client.py POST /v1/goals --idempotency-key checkpoint-1 --body '{"action":"local.checksum"}'
python3 scripts/client.py GET /v1/goals
python3 scripts/client.py GET /v1/system/capabilities
```

The state and development key live in ignored `.local/`. Do not publish either. For process-restart/crash behavior, run the integration suite; the development worker can be started in manual-tick mode or with `--auto-run`.

## Project documents

- [Architecture and current service boundaries](docs/architecture/system.md)
- [Requirement conflicts, open decisions and critical risks](docs/OPEN_DECISIONS.md)
- [Recorded implementation decisions](docs/DECISIONS.md)
- [Incremental delivery plan](docs/IMPLEMENTATION_PLAN.md)
- [Live implementation and verification status](docs/STATUS.md)
- [Mont Blanc feature matrix](docs/release/MONT_BLANC_FEATURE_MATRIX.md)
- [REST/OpenAPI contract for implemented endpoints](contracts/openapi/local.json)
- [Threat model](docs/threat-model/overview.md)
- [Actual test and CI evidence](docs/test-evidence/task-1.txt)

The original detailed Goal Engine, Mont Blanc and Series 26 specifications are owner-controlled source material and are not bundled into this public fork. This README summarizes their non-sensitive product requirements; it does not supersede the source of truth or claim feature completion.
