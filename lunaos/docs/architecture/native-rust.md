# Native LunaOS architecture

**Target:** independently installable Arch Linux distribution with Hyprland on
Wayland. Omarchy contributes reusable configurations and integrations; it is not
required at runtime. The Python/Go checkpoint remains a frozen reference. This
document distinguishes target boundaries from the subset already running locally.

## Deployable service boundaries

| Service | Rust workspace crate | Contract | Current state |
|---|---|---|---|
| LUNA Core API + scheduler | `luna-core` | Authenticated loopback REST/JSON; Tokio periodic scheduler | Runs locally on macOS; not installed as systemd service |
| Goal Engine | `luna-goal-engine` | Internal Rust API; SQLite migration | Local slice: score, idempotent create, queue/tick, restart persistence, evidence |
| Policy Broker | `luna-policy-broker` | Separate process; bounded newline JSON over private Unix socket | Rust child-process test passes; one public checksum action only |
| Shared contracts | `luna-contracts` | Versioned serde request/result types | Compiles; several Mont Blanc types are schema-only |
| Local model adapter | `luna-llm-adapter` | OpenAI-compatible local llama.cpp endpoint | Adapter source and loopback test; no model/server or real inference |
| Native shell | `luna-shell` | GTK4 + Wayland; Core status API | Source exists behind feature; GTK build/runtime unverified |

On Linux, Core and Policy Broker must run as the same dedicated unprivileged
`lunaos` UID for the initial socket ACL and peer-UID check. The broker verifies the
owner token on authorization and revalidation, denies all actions outside its
allowlist, and issues goal/action/expiry-scoped HMAC grants. The Core checks owner
credentials at the REST edge; the broker independently validates those credentials
before a grant is issued or reused. Policy denial occurs before score calculation
and authorization is checked again immediately before execution. No root command,
payment, communication, sensor or general shell adapter is enabled.

The broker socket directory is owner-only; the socket mode is `0600`. The service
refuses to replace an existing socket path. Requests and replies are capped at 8 KiB
and each connection carries one JSON request. Linux peer credentials must match the
broker's effective UID. This is an initial service boundary, not a claim of isolation
from a compromised same-UID process; distinct identities, protected credential
provisioning, revocation and production key lifecycle remain release requirements.

## Goal execution slice

`POST /v1/goals` accepts the typed create contract and an `Idempotency-Key`. Core
authenticates the owner, sends the policy request to the broker, persists an
authorized goal and queued job in SQLite, and computes the documented weighted
score only after authorization. The scheduler revalidates the scoped grant, runs
only `local.checksum` against a fixed public fixture, verifies the independent
SHA-256 and size, then commits the goal outcome and evidence. SQLite uses WAL and
full synchronous durability. Restarting Core leaves jobs readable and resumable.
The initial dependency graph rejects cycles and holds dependent goals until every
prerequisite completes. A queue that contains only blocked dependents reports
`waiting_for_dependencies`, not `idle_success`.

REST routes implemented in this slice:

| Method | Route | Authorization |
|---|---|---|
| `POST` | `/v1/goals` | Owner bearer token; requires idempotency key |
| `GET` | `/v1/goals/{id}` | Owner bearer token |
| `POST` | `/v1/goals/{id}/dependencies` | Owner bearer token; cycle-checked dependency edge |
| `POST` | `/v1/scheduler/tick` | Owner bearer token |
| `GET` | `/v1/system/capabilities` | Public static status; reports unavailable target capabilities |

Goal DAG planning, lifecycle mutations, full CRUD, event subscriptions, durable
leases, resource quotas, external side effects, evidence vault, memory and identity
services are not implemented. Contract structs do not mean those services exist.

## Native UI and inference choices

GTK4 is the first shell choice because LunaOS targets a GTK-heavy Arch/Wayland
ecosystem and benefits from existing accessibility, settings, theming and D-Bus
integration. The current source displays Core capability status only; it is not yet
a desktop shell suitable for release. The GTK feature is built and tested on an Arch
VM in a later milestone.

The LLM interface targets real local llama.cpp inference through a loopback-only
OpenAI-compatible endpoint. The adapter is never an authority source: model output
is an untrusted proposal that must pass the Goal Engine and Policy Broker. No model
weights are committed or downloaded by this repository. Release readiness requires
a model manifest with license and hardware measurements, followed by offline
inference evidence.

## Runtime/development boundary

Native services build from `rust/Cargo.toml`; the default Make targets use Cargo.
The legacy `luna/`, `services/policy/` and Python/Go tests exist only for behavioral
parity while migration proceeds. A LunaOS boot must not require Python or Go. This
has not yet been proven because no standalone image exists.

`systemd/luna-policy-broker.service` and `systemd/luna-core.service` are deployment
templates only. The account, credential provisioning, package, image, install,
update, rollback and recovery paths remain unfinished and untested on systemd.
