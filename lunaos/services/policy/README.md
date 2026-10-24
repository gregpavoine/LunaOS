# Local policy broker

Go standard library HTTP over private Unix socket; no privileged tool implementation.
Build: `go build -buildvcs=false -o .local/policy ./services/policy` after creating a
private `.local` directory, or use `make dev`. Test: `go test -buildvcs=false ./services/policy`.

Requires private owner key (64 hex chars) and 0700 socket parent. Refuses to replace
an existing socket. API: POST `/v1/grants`, POST `/v1/policy/evaluate`; both independently
authenticate bearer key. Goal/action-bound HMAC grant expires after one hour.
Only `local.checksum` under general privacy may pass; payment budget is always zero.
Stopping process prevents new work. Production OS privilege separation, MFA, grant
revocation registry and secure key custody are not implemented by this dev service.
