# Task 1 threat boundaries

Protected assets: owner key, policy grant integrity, goal state, idempotency, truthful
evidence. Untrusted: request JSON, claimed role, external assertion/source, unknown
actions, replayed requests. There is no privileged or sensor tool to target here.

- Unauthorized/guest/voice-only request: reject at API and independently at broker.
- Prompt-injected claim: stored only as hypothesis; never evaluated as code, SQL,
  grant or shell. Unknown source cannot become verified through confidence=1.
- Payload/grant replay: request hash + unique key; signed goal/action/expiry binding.
- Missing budget: all payment families denied before score, irrespective of utility.
- Broker outage/expiry/forgery: fail closed, queued job retained without evidence.
- Crash/concurrency: SQLite write serialization, FULL synchronous commit and unique
  evidence/execution; pure action can safely repeat after transaction rollback.
- Logs: generic codes only; no request logging; secrets excluded from Git and DB.
- Endpoint: private UDS, 0700 directory, 0600 files/sockets, body limits/timeouts;
  no TCP/public remote path, unsupported routes 404, browser Origin rejected.

Limits: same-UID or root compromise can steal dev key/change core/DB; DB triggers are
not a tamper-proof external audit. Key is plaintext in protected file, not a secure
enclave/credential vault. General DB is not encrypted. Store only public-fixture,
non-sensitive assertions; no DLP claim based on caller privacy labels. Independent
production service users, key custody, revocation/expiry sessions, rate limits,
secure backup, signed broker deployment and OS sandbox tests are later gates.
No privacy/sensor safety, encrypted intimate storage, payment execution or external
message delivery is certified by these local tests. A fully privileged host attacker
is outside this checkpoint's assurance boundary.
