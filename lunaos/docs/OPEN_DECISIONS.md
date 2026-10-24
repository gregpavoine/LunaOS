# Open decisions and requirement map

## Architecture correction (owner direction, 2026-10-24)

The later owner instruction requires a native, independently installable Arch
distribution: Rust/Tokio control plane, systemd, D-Bus/Unix sockets/Linux APIs,
Hyprland/Wayland, SQLite initial persistence, a native UI, and real local inference.
It supersedes this document's earlier Task 1 recommendation to ship LunaOS as a
Go/Python runtime overlay on Omarchy. The Go/Python code is frozen reference code.
Omarchy remains a reusable configuration and integration source, not a required
runtime dependency. No host cleanup or system modification is authorized by this
architecture decision.

GTK4 is selected for the first shell in ADR 0003 based on the target's GTK/Wayland
ecosystem. The shell and target integration still require Linux build/run evidence.

2026-10-24. Product specs describe requirements, not working software.

## Explicit constraints

| Requirement | Source | Task 1 treatment |
|---|---|---|
| Mac controller; no remote before authorization | Current task; AGENTS Platforms; Series 26 | Local-only; no installation or remote connection |
| Independent native broker; owner-only privileges; budgets zero | Goal §§0,1,2,6 plus owner architecture correction | Separate Rust process with scoped grants; payments denied |
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
| Earlier Go/Python implementation vs native Rust control plane | Later owner architecture correction wins; Go/Python remains frozen reference only and ADR 0002 is superseded. |

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
- Series24: detailed private preferences unanswered; never import historic answers
  into generic memory. No private profile created by this checkpoint.
- SSH alias/authorization, actual Linux inventory, existing ComfyUI path, model
  licenses/workflows, consent proofs, external APIs, device identities and money
  settings absent. Needed only for later milestones, not this local slice.

## Critical risks and release constraints

1. Local owner key authenticates possession, not biometric identity or production
   MFA. Same-UID compromise can access development keys; separate Linux UIDs,
   peer credentials and OS key store must precede privileged/untrusted execution.
2. SQLite development DB is not encrypted. Only non-sensitive fixture metadata is
   supported; private vault and encrypted backup are blocked, not simulated.
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
7. The systemd units and Arch packaging include separate Core and Broker binaries,
   but first-boot user creation, owner-token provisioning/rotation and an install or
   recovery workflow are not implemented. The units must not be enabled by an image
   until that credential bootstrap is designed and exercised in an isolated Arch VM.
