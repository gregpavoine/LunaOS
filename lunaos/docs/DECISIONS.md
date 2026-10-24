# Decisions

- Source order: current task > AGENTS/new Goal addendum for goals and authority >
  unified U0–U21 > legacy annexes; Series26 controls Omarchy retention.
- ADR0001: additive existing Omarchy; ADR0002: native Go broker/Python SQLite core.
- Development owner key is generated locally with private permissions and never
  committed or logged. No caller-supplied role/owner field establishes identity.
- Only action `local.checksum` on the compiled public fixture is executable.
  Explicit owner submission creates a signed one-hour goal-specific grant.
- Zero utility is idle success. Hard denial precedes scoring. Unknown-source
  assertions remain hypotheses even when submitted by the owner.
- Milestone 1b: parent hierarchy is separate from `blocks` dependencies; each is
  acyclic, one plan has at most 32 prerequisites, and queued execution waits until
  all blocking goals are completed. Blocked, failed and zero-utility prerequisites
  are rejected so the queue cannot stall on a known terminal goal. SQLite enforces
  the cap on inserts and updates. Plan input references existing goals only; it
  does not synthesize subgoals or expand execution authority. An expired grant
  blocks its goal and records `GRANT_RENEWAL_REQUIRED`; the owner replaces the goal
  and updates dependent plans. Automatic renewal is not implemented.
- One SQL write transaction encloses policy recheck, pure execution, verification,
  evidence, event delivery and completion. No external side effects supported.
- No real hardware capability replaced by a fake provider; interfaces only.
