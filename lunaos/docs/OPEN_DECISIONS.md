# Open decisions and requirement map

## Architecture corrections (owner direction, 2026-10-24 and 2026-10-25)

The 2026-10-24 correction established an independently installable Arch-based
distribution and froze the initial Go/Python implementation. On 2026-10-25 the
owner directed LunaOS to follow Omarchy's implementation techniques and not require
Rust. ADR 0004 now selects upstream Arch packages, Bash integration, Hyprland,
Quickshell/QML, and Go for long-running Luna services; Python remains isolated to
ML/third-party workloads. The existing Rust checkpoint is frozen until Go parity
is verified. Omarchy remains a reference, not a runtime dependency. The kernel is
upstream Arch; no LunaOS kernel is planned by default.

No host cleanup or system modification is authorized. The Go Core/Broker and
optional private-memory vault have a portable local slice; Quickshell/QML and all
Linux/Arch execution remain unimplemented or unverified.

2026-10-24. Product specs describe requirements, not working software.

## Explicit constraints

| Requirement | Source | Task 1 treatment |
|---|---|---|
| Mac controller; no remote before authorization | Current task; AGENTS Platforms; Series 26 | Local-only; no installation or remote connection |
| Independent native broker; owner-only privileges; budgets zero | Goal §§0,1,2,6 plus owner architecture corrections | Current Rust checkpoint has a separate process and scoped grants; target Go broker migration remains pending; payments denied |
| Durable registry, deny before score, verified outcomes, idle success | Goal §§3–5,11 | SQLite transaction; fixed public checksum action; executable tests |
| No invented facts from unknown provenance | Goal §1 Series30.8; G-010 | Unknown assertion always hypothesis, confidence does not verify it |
| Local-first, protected secrets, no sensitive general logs | Goal §§0,3,7 | Private UDS/key; public-fixture metadata only; no sensitive ingestion |
| Preserve Omarchy KEEP components | Series 26; G-028 | No host mutation; Linux adapters unavailable |
| Full V0 requires real GPU workflow and reboot | Goal §9 | Release blocked; local tests do not count as hardware proof |
| All named Mont Blanc modules remain mandatory | Unified U0,U19,U20 | Retained in feature matrix, not silently deferred to Fuji/Everest |

## Contradictions resolved by precedence

| Conflict | Resolution |
|---|---|
| Additive Omarchy overlay vs independently installable Arch OS | Later owner architecture correction supersedes the older Task 1 and Series26 deployment recommendation: LunaOS owns its image; Omarchy is a reference. Any installed-host transition still requires separate authorization and recovery proof. |
| Old romantic partner behavior vs assistant identity | Goal addendum and U0 win: assistant, no claims of subjective feeling |
| Old unlimited autonomy/no confirmation vs independent safety | Goal §§0,1,6 win; no implicit privilege or financial grants |
| Series29.80 retain after revocation vs revocable consent | New Goal §§1,7 require cessation, restriction, required erasure/derivative handling |
| Annex blind always preempts deliberate capture vs U3 session exception | U3 allows specifically scoped visible session; manual BLIND always overrides; no sensors implemented here |
| Annex discovery sync allowed during BLIND vs unresolved Q79/current U1 | Conservative no sensitive processing; exact pre-existing-data policy OPEN_DECISION |
| 2027/year-plus roadmap vs U21 8/12-week target | Current target is conditional and high risk, never a completion guarantee |
| Claude Code 24/7 vs Codex implementation lead | Current task authorizes Codex; no external agent subscription, unattended spending or scheduler provisioned |
| Rust checkpoint vs Omarchy-aligned implementation | Later 2026-10-25 owner direction and ADR 0004 select Bash + Quickshell/QML + Go services for future work. Keep Rust frozen until parity; Go parity and Rust removal are not complete. |

## Unresolved (no fabricated approvals)

- OPEN_DECISION 29.30: provenance visibility. Always expose safe status in local API;
  broader UI policy undecided.
- OPEN_DECISION 29.42: tests before every modification. Task 1 uses test-first for
  runtime behavior; this does not settle the permanent product decision.
- OPEN_DECISION 29.79: processing already-held private data during BLIND. Sensitive
  ingestion/processing unavailable; no inference that existing consent resolves it.
- OPEN_DECISION 29.82: mobile platforms beyond iPhone. No additional platform claimed.
- Series28: CPU/VRAM/concurrency/notification/budget preferences unanswered. One
  serialized local worker, no GPU/network/spending; weighted formula is provisional.
- Series22: disk strategy, dual boot, encryption unlock, NAS defaults, cloud GPU,
  immutable safety/recovery and release packaging; decide before those integrations.
- Performance budgets: define boot-to-shell, service startup, scheduler/API p50/p95,
  idle CPU/RSS and throughput targets only after an Arch VM and target baseline.
  The current Mac test/build results do not establish Linux performance.
- Series24: detailed private preferences unanswered; never import historic answers
  into generic memory. No private profile created by this checkpoint.
- SSH alias/authorization, actual Linux inventory, existing ComfyUI path, model
  licenses/workflows, consent proofs, external APIs, device identities and money
  settings absent. Needed only for later milestones, not this local slice.
- Intimacy adult eligibility: current API records only an owner self-attestation;
  decide whether independent age assurance is required before enabling adult
  conversation, media indexing or generation. Those capabilities remain disabled.
- Intimate vault key lifecycle: an in-library atomic rotation primitive and
  separate-key encrypted backup/verified restore format are implemented locally.
  A `luna-vault` keygen/backup/restore CLI exists, but OS key-store integration,
  key-file replacement during rotation, lost-key recovery and
  cryptographic erasure remain undecided/unimplemented. The optional vault accepts
  externally supplied private 32-byte key files and must not be treated as
  production key management.
- Conversation model and safeguards: no local model is installed or licensed, and
  no adult-content moderation/evaluation stack is configured. Select an actual
  model, license, resource profile and safety evaluation before exposing chat.
- Media/generation policy: approved source allowlists, rights/takedown workflows,
  classifier models, ComfyUI workflow/model manifests, age/consent gates and GPU
  budgets require concrete sources and target hardware. No crawling or generation
  capability is enabled.

## Critical risks and release constraints

1. Local owner key authenticates possession, not production MFA. Same-UID compromise
   can access development keys. Go Broker source checks Linux peer UID but has not
   been executed on Linux; broker and Core run under the same account in templates
   and are not ready to hold system authority.
2. The Core SQLite DB is not encrypted. A separate optional vault encrypts memory
   payloads and now has local rotation/backup/restore commands, but key
   provisioning, service coordination, lost-key recovery and Linux
   at-rest verification are absent; sensitive operation must remain opt-in and
   disabled without an externally supplied private key.
3. Transactional serialization is safe for this bounded pure action only. It is
   not a distributed scheduler or general exactly-once external execution.
   Lease renewal and a default three-expiry retry ceiling are local only; an
   independent heartbeat worker, configurable retries and external event delivery
   are not implemented. Payments/messages require durable intents, reservation and
   provider reconciliation.
4. Power-loss hardware durability, sensors/BLIND cutoff, GPU contention, iPhone,
   NAS, recovery and all Linux claims require real authorized target tests.
5. Signed grants expire after one hour; expired/invalid/unavailable policy defers
   work. No automatic renewal on behalf of an absent owner.
6. Existing duplicate start-pack specs are preserved; root docs are canonical.
   Bundled cleanup script is reference material and must not be auto-executed.
7. A local owner-token setup command now creates a private credential file and the
   systemd-sysusers configuration declares the service account. Linux first-boot
   orchestration, token rotation/revocation and an install/recovery workflow are not
   implemented. The units must not be enabled by an image until this flow is tested
   in an isolated Arch VM.
8. The Intimacy vault's `adult_attested` value is a self-declaration, not verified
   age. Third-party consent, intimacy-read/write session capabilities beyond the
   memory scopes, local model safeguards, STOP wiring to voice/UI, and automatic
   deletion of derivatives are absent; no adult conversation, indexing or
   generation is available in this checkpoint.
