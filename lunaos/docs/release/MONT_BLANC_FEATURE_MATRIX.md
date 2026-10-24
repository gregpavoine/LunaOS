# Feature status — release PLANNED

Updated 2026-10-24 after native architecture correction. The prior Python/Go checkpoint is a frozen reference. The new Rust migration has portable unit-test evidence only; it is not a deployed OS or a full V0. All mandatory Mont Blanc requirements remain in scope.

| Feature / acceptance family | State | Evidence or blocker |
|---|---|---|
| Rust goal registry + deterministic score | IMPLEMENTED; VERIFIED_LOCAL_UNIT | SQLite WAL/FULL and five Rust engine tests; no target service/reboot test |
| Rust owner policy crate + scoped signed grant | IMPLEMENTED; VERIFIED_LOCAL_UNIT | Owner-only, payment-without-budget, deny-before-score and scope tests; no separate UID/peer-credential service boundary |
| Durable queued work + restart resume | IMPLEMENTED; VERIFIED_LOCAL_UNIT | Reopen SQLite then execute fixed action in process test; no host reboot/power-loss evidence |
| Unknown provenance remains unknown | IMPLEMENTED; VERIFIED_LOCAL_UNIT | Contract excludes caller-controlled verification status; test records unknown |
| Idle success and exact action evidence | IMPLEMENTED; VERIFIED_LOCAL_UNIT | Fixed public fixture SHA-256 and byte count; fixture-only behavior |
| Rust Core REST routes | IMPLEMENTED; NOT_RUNTIME_VERIFIED | Tokio/Axum loopback source; no live API smoke test yet |
| Local llama.cpp inference | BLOCKED_EXTERNAL | Real OpenAI-compatible adapter exists; no server or model configured, no inference run |
| GTK4 native shell | SOURCE_PRESENT; BLOCKED_TARGET | Requires GTK4 development libraries and Arch/Hyprland runtime; not built here |
| Independent Arch package and ISO | PREPARATION_ONLY | PKGBUILD/profile/build script authored; no Arch build, installer, signatures or boot verification |
| Bounded explicit parent/dependency DAG + cycle-safe scheduler gate | REFERENCE_ONLY | Existing Go/Python behavior remains frozen; Rust port not yet done |
| Full goal CRUD/generated planner/strategy optimizer | NOT_STARTED | Only explicit dependency plans and one measured checksum strategy |
| Authenticated production sessions/key vault/remote admin | BLOCKED | Device identities/key store and hardened Linux boundary absent |
| Payments, reservations/ledger, external conversation grants | NOT_STARTED | Always denied; no budget or recipients configured |
| Independent LunaOS shell/session/keyboard command center | BLOCKED | GTK binary unbuilt; Linux inventory and authorized deployment needed |
| Omarchy component reuse and KEEP validation | BLOCKED | No target access; no host changes |
| Native lock/idle/update UI + recovery/rollback | BLOCKED | Target baseline and recovery proof needed |
| Local LLM/router/memory/persona continuity | PARTIAL / BLOCKED | HTTP adapter source only; no real inference, persistent memory engine or identity lifecycle |
| GPU resource broker/thermals/multi-GPU | BLOCKED | Actual inventory/benchmarks needed |
| Existing ComfyUI image/video workflow | BLOCKED | Real path/workflow/licenses/GPU not provided |
| Independent guardian/system-root action sandbox | NOT_STARTED | Rust broker crate has no privileged action; process/UID isolation and systemd integration remain absent |
| Privacy controller + BLIND/sensor brokers | NOT_STARTED | No camera/mic/screen access; no protection claim |
| Consent/revocation/erasure/subject identity | NOT_STARTED | Verifiable subject workflow and deletion proofs required |
| Encrypted private vault + body/face/training scopes | NOT_STARTED | No sensitive data ingestion supported |
| Full native Intimacy Core/preferences/conversation | NOT_STARTED | Mandatory Mont Blanc scope preserved |
| Discovery Reddit/Civitai/source policy/downloads | BLOCKED | Actual API grants/provider review; no simulated connector |
| Adult media multimodal index/search/recommendations | NOT_STARTED | Licensed fixtures and real models required |
| Voice/STT/TTS, perception, private notifications | NOT_STARTED | Permissions and real integrations required |
| iPhone USB pairing/media/encrypted backup/contacts | BLOCKED | Actual Linux+iPhone/iOS tests required |
| WhatsApp authorized archive | BLOCKED | Capability/backup availability not guaranteed |
| NAS encrypted replication/backup restore | BLOCKED | Actual mount/policy/keys and restore-readiness tests |
| Reproducible image/installer/signing/SBOM/soak | PREPARATION_ONLY | Skeleton only; source pin, license decision, installer/recovery and Arch image build/boot evidence absent |

Historical Go/Python test output: `docs/test-evidence/task-1.txt`. Native Rust test output is captured in the current turn evidence log after final verification. Full required inventories remain
Goal G-001..030, unified U-001..040, BL-001..030, Discovery E2E-01..40,
iPhone U20 acceptance and Series26 T01..15. None of the complete hardware suites
is declared passed by local unit/integration tests.
