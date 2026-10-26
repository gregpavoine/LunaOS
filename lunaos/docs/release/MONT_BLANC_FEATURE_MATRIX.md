# Feature status — Mont Blanc NOT RELEASED

Updated 2026-10-26. No full V0, RC or Mont Blanc release claimed.
`VERIFIED_LOCAL` means a bounded local test or process result; it does not imply
Arch installation, systemd, boot, hardware or release verification. All mandatory
later features remain in Mont Blanc scope.

| Feature / acceptance family | State | Evidence or blocker |
|---|---|---|
| Frozen Go/Python goal registry + deterministic score | VERIFIED_LOCAL | `make test-reference`; prototype only |
| Go Policy Broker + Go Core process boundary | VERIFIED_LOCAL | Linux Docker process test: peer UID, owner auth, scoped grant, queued-job recovery, API idempotency and verified fixture evidence; not Arch/systemd |
| Go Core/Broker compiled services | VERIFIED_LOCAL | Linux container build/test; packaged x86_64 Arch binaries built and installed under emulation; target hardware/systemd remain unverified |
| Go owner credential provisioning | PARTIAL | CLI creates separate private owner and Core-to-Broker credentials; first-boot unit is present but not validated under systemd; session delivery/revocation UX remains absent |
| Go service Arch package | VERIFIED_LOCAL | x86_64 Arch package built under emulation; latest package not installed or run under systemd; no VM/target claim |
| Standalone Arch installer source | PARTIAL | `feature/standalone-installable-iso` adds guided archinstall, diskless config, target finalizer, signed local mirrors for Arch packages and a command-scoped live sudo rule; shell boundary regression passes on macOS; no Arch package/ISO build or install/reboot VM acceptance |
| Native Rust goal registry + deterministic score | VERIFIED_LOCAL | Rust tests + process smoke; G-003 bounded subset |
| Independent owner-authenticated Rust policy process | VERIFIED_LOCAL | Child-process UDS test + native smoke; checksum only |
| Native Rust Core API and durable local queue | VERIFIED_LOCAL | Real local REST smoke; no Linux/systemd claim |
| Rust goal list/pause/resume/cancel subset | VERIFIED_LOCAL | Owner HTTP tests, score-sorted bounded list, terminal-state checks, restart persistence; not full goal state machine |
| Durable append-only goal events + local outbox/inbox | VERIFIED_LOCAL | SQLite migration, idempotent local consumer and owner event-history API; no external bus |
| Budget zero, deny-before-score | VERIFIED_LOCAL | Rust policy/engine tests; no privileged system action enabled |
| Pure-job idempotency/restart/lease-expiry recovery | VERIFIED_LOCAL | Local SQLite/process-reopen tests; no SIGKILL or target power-loss proof |
| Unknown provenance hypothesis | VERIFIED_LOCAL | G-010 |
| Idle success/offline local goal feature | VERIFIED_LOCAL | G-021, G-023 local subset |
| Evidence-backed checksum completion | VERIFIED_LOCAL | G-013 pure fixture only |
| Local REST owner transparency | VERIFIED_LOCAL | API round-trip; partial G-030 |
| Bounded dependency edges/cycle rejection/scheduler ordering | VERIFIED_LOCAL | Owner-authenticated Rust and Go Core parent/dependency plans; additive Go SQLite migration, 32-edge bound, combined-graph cycle checks, dependency gate and restart persistence pass local tests; no natural-language planning or Arch/systemd proof |
| Go Core goal decisions history REST endpoint | LOCAL_MERGED_UNPUSHED | Commit `440591b3` is merged locally to `mont-blanc`; owner-authenticated, privacy-safe retrieval of the latest 100 events; focused handler test, `make test check` and Core race suite pass; not pushed or included in an ISO |
| Go Core bounded checksum evidence evaluation | LOCAL_MERGED_UNPUSHED | `feature/go-goal-evaluation` is merged locally to `mont-blanc`; recomputes SHA-256/byte count for the known fixture, reports insufficient evidence or mismatch, and appends an audit event; focused test, `make test check`, Core race suite and local build pass; not pushed or in an ISO; no strategy evaluation |
| Go Core live component readiness | LOCAL_MERGED_UNPUSHED | `feature/go-system-capabilities` is merged locally to `mont-blanc`; probes SQLite and authenticated, peer-checked Policy Broker readiness over the Unix socket; LLM unavailable and hardware unverified; tests, full Go checks, Core race suite and local binaries pass; not pushed or in an ISO, no Linux peer-credential/systemd validation |
| Full goal lifecycle/CRUD/planner/strategy optimizer | PARTIAL | Go Core has owner-authenticated title/description editing before execution, pause/resume/cancel, bounded plan mutation and policy-rechecked reprioritization; action/content lifecycle editing, remaining transitions, natural-language planning and evidence-backed strategy evaluation remain absent |
| Authenticated production sessions/key vault/remote admin | BLOCKED | Device identities/key store and hardened Linux boundary absent |
| Payments, reservations/ledger, external conversation grants | NOT_STARTED | Always denied; no budget or recipients configured |
| Native Wayland shell/session/keyboard command center | PARTIAL | Quickshell/QML dashboard source and Hyprland bindings exist; Arch `qmllint` exits 0 with a `PanelWindow` type-resolution warning; no Wayland/Hyprland runtime |
| Launcher/clipboard/notifications/apps KEEP validation | BLOCKED | No target access; no host changes |
| Native lock/idle/update UI + recovery/rollback | BLOCKED | Target baseline and recovery proof needed |
| Local LLM/router/memory/persona continuity | NOT_STARTED | No inference runtime tested |
| GPU resource broker/thermals/multi-GPU | BLOCKED | Actual inventory/benchmarks needed |
| Existing ComfyUI image/video workflow | BLOCKED | Real path/workflow/licenses/GPU not provided |
| Independent guardian/system-root action sandbox | NOT_STARTED | Broker here has zero privileged actions |
| Privacy controller + BLIND/sensor brokers | NOT_STARTED | No camera/mic/screen access; no protection claim |
| Consent/revocation/erasure/subject identity | NOT_STARTED | Verifiable subject workflow and deletion proofs required |
| Encrypted private vault + body/face/training scopes | PARTIAL | Owner-only AES-GCM vault, scoped consent, keygen, atomic key rotation and encrypted backup/restore CLI exist; no OS key store, key-file rotation, lost-key recovery, third-party scopes, body/face data or model scopes |
| Full native Intimacy Core/preferences/conversation | PARTIAL | Go Core now has encrypted, owner-declared preference CRUD with scoped consent and correction; no conversation model, preference inference, verified age/identity or media features |
| Discovery Reddit/Civitai/source policy/downloads | BLOCKED | Encrypted fail-closed local metadata allowlist is implemented; real provider API grants, reviewed connector, media retrieval/licensing and age/consent checks remain absent |
| Adult media multimodal index/search/recommendations | NOT_STARTED | Licensed fixtures and real models required |
| Voice/STT/TTS, perception, private notifications | NOT_STARTED | Permissions and real integrations required |
| iPhone USB pairing/media/encrypted backup/contacts | BLOCKED | Actual Linux+iPhone/iOS tests required |
| WhatsApp authorized archive | BLOCKED | Capability/backup availability not guaranteed |
| NAS encrypted replication/backup restore | BLOCKED | Actual mount/policy/keys and restore-readiness tests |
| Reproducible image/installer/signing/SBOM/soak | NOT_STARTED | Baseline, recovery and full release gates first |

Test output: `docs/test-evidence/task-1.txt`. Full required inventories remain
Goal G-001..030, unified U-001..040, BL-001..030, Discovery E2E-01..40,
iPhone U20 acceptance and Series26 T01..15. None of the complete hardware suites
is declared passed by local unit/integration tests.
