// Package inference supervises llama.cpp and only serves hash-verified models.
//
// # HTTP contract
//
// luna-inference serves HTTP over a Unix socket (mode 0660). Every route first
// checks the socket peer's UID; anyone but the allowed UID (the Core) gets 403
// SOCKET_PEER_DENIED, whatever the path or method. Errors are JSON
// {"code":"<CODE>"}.
//
//	GET  /v1/health    200 {"state":"stopped|starting|ready|degraded","model":"<id>"}
//	                   Never starts or waits for the model; model is "" when
//	                   none is verified.
//	POST /v1/generate  {"prompt":"<1..32768 bytes>","max_tokens":<1..1024>}
//	                   200 {"text":"...","model":"<id>","untrusted":true}
//	                   Starts the model lazily. Unknown fields and trailing
//	                   data are refused.
//	POST /v1/unload    204, after generations in flight finish; the next
//	                   generate starts the model again.
//
// Error codes:
//
//	403 SOCKET_PEER_DENIED     peer is not the allowed UID
//	400 INVALID_REQUEST        body, prompt or max_tokens out of bounds
//	404 NOT_FOUND              unknown path
//	405 METHOD_NOT_ALLOWED     wrong method (Allow names the right one)
//	503 INFERENCE_NO_MODEL     no verified model configured; fix the setup
//	503 INFERENCE_DEGRADED     too many start failures, or llama-server or its
//	                           key directory unusable; no start is tried until
//	                           the service restarts (see Recovery)
//	503 INFERENCE_UNAVAILABLE  one start failed within the budget; retryable
//	503 INFERENCE_TIMEOUT      the model is still loading; retry later, the
//	                           load goes on (bounded by --start-timeout)
//	502 INFERENCE_UPSTREAM     llama-server failed or gave no content
//
// # Untrusted output
//
// Generated text is model output: untrusted data, flagged "untrusted":true on
// every reply. It can never authorize an action, lift BLIND or widen a scope.
// Client refuses a reply that is not flagged.
//
// # Client
//
// Client returns a non-200 reply as *APIError (Status and Code). A missing
// socket or ECONNREFUSED comes back as a plain net error, not *APIError: the
// service is not running, which callers treat as unavailable.
//
// # Recovery
//
// Three start failures within 60 s, or an unusable llama-server binary or key
// directory, make the supervisor degraded. It stays so, with no cooldown, until
// Supervisor.Reset. The service does not call Reset, so the recovery is
// `systemctl restart luna-inference`. Who may trigger a Reset is decided at
// integration.
package inference
