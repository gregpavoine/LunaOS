# LunaOS native migration threat model

This is the current threat model for the Rust migration slice, not a security certification. The old Go/Python reference has separate tests, but its security results do not automatically carry over.

## Protected assets and boundaries

- **Owner credential:** read by `luna-core` from a systemd credential file when deployed that way; local development can provide an environment variable. There is no enrollment/session lifecycle, secure key store, rotation or owner-facing revoke flow yet.
- **Goal state and scoped action grants:** stored in local SQLite. The database file is set to mode 0600 on Unix; WAL/FULL synchronous persistence is used. General DB encryption and tamper-proof external audit are absent. No sensitive goal text is persisted in this slice.
- **Policy grants:** HMAC-bound to a specific goal and action with one-hour expiry; the owner bearer credential is the development HMAC key. Broker crate and Core are currently in one process, so a Core compromise can issue grants. No privileged host action exists.
- **Local model prompt:** the llama.cpp adapter only accepts HTTP loopback endpoints. This limits accidental network egress from this adapter; the model/server is not installed, and prompt isolation/secret filtering are not yet implemented.
- **System and private data:** no arbitrary shell, file, root, camera, microphone, screen, GPU, iPhone, NAS, external messaging or payment adapter is enabled. Such features remain unavailable until independent authorization and privacy gates are built.

## Threats and current controls

| Threat | Current control | Remaining gap |
|---|---|---|
| Guest or unauthenticated caller creates goal | REST requires bearer token; broker also requires owner token and owner principal | No production owner identity, MFA or role/session service |
| Caller claims verified provenance | Request schema has no source-status input; persisted status is forced to unknown | No source verification workflow |
| Prompt or model output invents authority | No model-to-broker path; only fixed action is allowlisted | Full agent runner and tool-policy boundary remain future work |
| Score bypasses policy | Broker evaluation occurs before score; denied actions never enter queue | Need independent broker process and explicit test instrumentation at service boundary |
| Replay changes action or target | HMAC grant binds goal ID and action, expires after one hour; unique idempotency key | Revocation/session replacement, multi-process execution and policy outbox not implemented |
| Payment without budget | Broker denies if configured budget is absent; payments are never allowlisted | Ledger/reservations/provider reconciliation absent |
| LLM leaks prompt to cloud | HTTP adapter rejects non-loopback endpoints | Direct third-party adapters do not exist; data-class routing/redaction not integrated |
| API exposes data remotely | Core binds to 127.0.0.1 by default and goal routes require bearer auth | Production bind validation, request limits, Unix sockets and peer-credential checks remain required |
| Process crash loses queued goal | SQLite transaction and reopen test preserve queued work | No target reboot, disk fault, multi-process lease or power-loss evidence |
| Sensitive capture or private data enters logs | No such adapter/ingestion path exists; general events omit user prose | No BLIND/sensor broker/vault; private-data operation must remain disabled |

## Required before privileged or sensitive operations

Implement a separately deployed broker with least-privilege service identities, authenticated Unix-socket peer credentials, OS-managed owner/session credentials, grant expiry/revocation, bounded requests and independent policy tests. Bind typed system adapters to signed action/target/data grants. Add audit redaction, secure backups, service sandboxing, offline and crash tests. Camera/screen/audio paths additionally require independent sensor arbitration, a trusted BLIND latch, visible subject-specific consent, immediate revocation and verified derivative deletion. Payments and outbound communications require atomic ledgers and recipient-scoped grants. None of these operations is approved by a score or LLM output.

The frozen Python/Go prototype retains its own historical risk notes and tests. It must not be used to claim Rust service isolation or target Linux protection.
