# Decisions

- Source order: later owner architecture correction > AGENTS/new Goal addendum for
  goals and authority > unified U0–U21 > legacy annexes. Series26 KEEP items remain
  relevant before any separately authorized migration from an existing Omarchy host.
- ADR0001 now selects an independently installable Arch-based LunaOS; Omarchy is
  a reference implementation, not a runtime dependency.
- ADR0002's Go/Python selection is superseded. ADR0003 selects the Rust/Tokio OS
  control plane; the initial Go/Python behavior remains frozen as test reference.
- Development owner key is generated locally with private permissions and never
  committed or logged. No caller-supplied role/owner field establishes identity.
- Only action `local.checksum` on the compiled public fixture is executable.
  Explicit owner submission creates a signed one-hour goal-specific grant.
- Zero utility is idle success. Hard denial precedes scoring. Unknown-source
  assertions remain hypotheses even when submitted by the owner.
- One SQL write transaction encloses policy recheck, pure execution, verification,
  evidence, event delivery and completion. No external side effects supported.
- No real hardware capability replaced by a fake provider; interfaces only.
