# Feature status — release PLANNED

As of Task 1, 2026-10-24. No full V0, RC or Mont Blanc release claimed.
VERIFIED_LOCAL means only the bounded Mac slice has executed evidence.
All mandatory later features remain in Mont Blanc scope.

| Feature / acceptance family | State | Evidence or blocker |
|---|---|---|
| Local goal registry + deterministic score | VERIFIED_LOCAL | test_slice; G-003 bounded subset |
| Independent owner-authenticated policy process | VERIFIED_LOCAL | Go + real UDS tests; G-008 local subset |
| Budget zero, deny-before-score | VERIFIED_LOCAL | payment/root denial tests; G-005 |
| Pure-job idempotency/restart/SIGKILL recovery | VERIFIED_LOCAL | G-012 local subset only; no target power-loss proof |
| Unknown provenance hypothesis | VERIFIED_LOCAL | G-010 |
| Idle success/offline local goal feature | VERIFIED_LOCAL | G-021, G-023 local subset |
| Evidence-backed checksum completion | VERIFIED_LOCAL | G-013 pure fixture only |
| Local REST owner transparency | VERIFIED_LOCAL | API round-trip; partial G-030 |
| Full planner DAG/goal CRUD/strategy optimizer | NOT_STARTED | Basic single-job slice only |
| Authenticated production sessions/key vault/remote admin | BLOCKED | Device identities/key store and hardened Linux boundary absent |
| Payments, reservations/ledger, external conversation grants | NOT_STARTED | Always denied; no budget or recipients configured |
| Omarchy additive shell/session/keyboard command center | BLOCKED | Linux inventory + authorized deployment needed |
| Launcher/clipboard/notifications/apps KEEP validation | BLOCKED | No target access; no host changes |
| Native lock/idle/update UI + recovery/rollback | BLOCKED | Target baseline and recovery proof needed |
| Local LLM/router/memory/persona continuity | NOT_STARTED | No inference runtime tested |
| GPU resource broker/thermals/multi-GPU | BLOCKED | Actual inventory/benchmarks needed |
| Existing ComfyUI image/video workflow | BLOCKED | Real path/workflow/licenses/GPU not provided |
| Independent guardian/system-root action sandbox | NOT_STARTED | Broker here has zero privileged actions |
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
| Reproducible image/installer/signing/SBOM/soak | NOT_STARTED | Baseline, recovery and full release gates first |

Test output: `docs/test-evidence/task-1.txt`. Full required inventories remain
Goal G-001..030, unified U-001..040, BL-001..030, Discovery E2E-01..40,
iPhone U20 acceptance and Series26 T01..15. None of the complete hardware suites
is declared passed by local unit/integration tests.
