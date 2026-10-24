# LunaOS — Task 1 local checkpoint

LunaOS is planned on top of an existing Omarchy installation. This repository now
contains a portable, real goal-engine slice. **Full V0 and Mont Blanc are not finished.**
No Linux host has been contacted; no OS/hardware integration is enabled.

Requires Python 3.11+ and Go 1.20+ already installed. No third-party runtime packages,
Docker or host package changes. Locally verified versions and results: [STATUS](docs/STATUS.md).

```sh
make env                  # optional isolated Python, no pip/network
make test check
.venv/bin/python scripts/dev.py
```

Keep development running in one terminal. In another:

```sh
python3 scripts/client.py POST /v1/goals --idempotency-key checkpoint-1 --body '{"action":"local.checksum"}'
python3 scripts/client.py GET /v1/goals
python3 scripts/client.py GET /v1/system/capabilities
```

The worker scores an authorized goal, checks the independent broker, hashes a fixed
public fixture, verifies the digest/length, and commits evidence. Repeated requests
with the same key return the same job. State stays under ignored `.local/`; restarting
the launcher reuses the database/key. Nothing claims a ComfyUI or GPU result.
Only non-sensitive public fixture metadata is supported. Assertion claim/source text is discarded before hashing or persistence; caller-supplied privacy labels are not a general-purpose DLP control.

To prove restart without racing the automatic worker, the integration suite queues
a goal in one process and completes it in another; it also kills a worker with SIGKILL
inside execution and verifies rollback/retry. `python3 -m luna serve` defaults to
manual ticks; `--auto-run` enables the one-second local worker.

[Architecture](docs/architecture/system.md), [Open decisions](docs/OPEN_DECISIONS.md),
[Incremental plan](docs/IMPLEMENTATION_PLAN.md), [REST contracts](contracts/openapi/local.json).
The supplied start-pack and source specifications are preserved.

`scripts/linux-baseline.sh` is authored for a future authorized Linux baseline.
**Do not execute it until the owner supplies an SSH alias and authorizes connection
and baseline execution.** The bundled cleanup scripts are reference material only.
