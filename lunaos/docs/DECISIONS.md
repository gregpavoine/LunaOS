# Decisions

- Source order: later owner architecture correction > AGENTS/new Goal addendum for
  goals and authority > unified U0–U21 > legacy annexes. Series26 KEEP items remain
  relevant before any separately authorized migration from an existing Omarchy host.
- ADR0001 and ADR0004 are historical. ADR0005 is the current distribution boundary: this repository is the Omarchy-derived LunaOS system source; package recipes and the branded ISO are separate sibling repositories. Do not build a parallel installer, shell or base package closure.
- ADR0002 records the initial Go/Python checkpoint. ADR0003 records the Rust checkpoint. ADR0005 keeps Go for active long-running Luna services, Bash for distribution integration, Quickshell/QML for desktop integration and Python only for isolated ML/third-party workloads. Keep the Rust checkpoint frozen until Go parity passes; no Rust removal is implemented yet.
- Development owner key is generated locally with private permissions and never
  committed or logged. No caller-supplied role/owner field establishes identity.
- Only action `local.checksum` on the compiled public fixture is executable.
  Explicit owner submission creates a signed one-hour goal-specific grant.
- Zero utility is idle success. Hard denial precedes scoring. Unknown-source
  assertions remain hypotheses even when submitted by the owner.
- Job claim and result completion use separate immediate SQLite transactions. A
  random expiring lease token fences stale workers; the only action remains a pure
  checksum, so this does not establish exactly-once behavior for external effects.
- Goal lifecycle currently exposes owner pause, resume and cancel. Completion and
  cancellation are terminal; pause clears any lease; resume queues the goal again.
- Goal events are append-only and atomically enter a local SQLite outbox; an
  idempotent local consumer records inbox receipt. This is not external bus delivery.
- No real hardware capability replaced by a fake provider; interfaces only.
