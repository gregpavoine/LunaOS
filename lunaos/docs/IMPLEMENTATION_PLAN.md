# LunaOS native implementation plan

This plan supersedes the prior additive Python-overlay plan. Keep the full Mont Blanc feature matrix mandatory and preserve the current Go/Python prototype as a frozen migration reference.

## Milestone N0 — architecture and Rust bootstrap

Create native Rust workspace and contracts; port policy-first scoring, owner-only action grants, durable local goal, safe whitelisted action, evidence, restart resume and idle success. Add Core REST routes, initial llama.cpp HTTP adapter, GTK4 shell source, truthful capability states and initial Arch packaging/profile structure. Record compile/runtime blockers accurately. **Current state: in progress; see `STATUS.md`.**

## Milestone N1 — Core and broker parity

Port remaining verified reference behaviors: goal lifecycle and CRUD; bounded parent/dependency DAG and cycle rejection; score explanation; execution attempts; independent evaluator; immutable versioned event log; transactional outbox/inbox; grant expiry/revocation; broker process isolation with Unix peer credentials; protected credential provisioning; owner session lifecycle; policy-before-score and before-execution; API compatibility contracts. Add migrations preserving exported reference records where necessary. No shell/root or external action until broker boundary is independently reviewed.

## Milestone N2 — independent Arch boot path

Own the package manifest, system config, installer, desktop/session defaults, systemd services, update channel and recovery. Reuse Omarchy components only after recording provenance/licensing and demonstrating KEEP behavior or tested LunaOS replacement. Implement reproducible PKGBUILD/lockfile and archiso profile; build and boot images in disposable VMs; test install, rollback and recovery before target hardware. No host disk changes without specific authorization.

## Milestone N3 — native shell, local inference and Linux adapters

Run Core and policy services under systemd with least privilege, owner credential provisioning and private Unix sockets/D-Bus. Build GTK4 shell and test Hyprland/Wayland, keyboard, accessibility, login, recovery and capability/action timeline. Install a licensed local llama.cpp model, measure RAM/latency/offline inference on the authorized target; keep LLM output untrusted and policy-independent. Implement systemd/D-Bus/NVIDIA/ComfyUI adapters with actual output verification. Execute V0 reboot-resume and the source V0 exit scenario on real target hardware.

## Milestone N4 — privacy and Mont Blanc capabilities

Implement encrypted domain-separated memory/vault; identity/persona revisions; consent receipts and revocation/erasure jobs; sensor brokers and BLIND interlock; scoped perception and adult-media controls; adaptive planner/scorer/resource manager; agent orchestration and project lifecycle; model evaluation/adapters; Intimacy Core, Discovery, iPhone sync/backup, NAS replication, remote access and user-facing controls. Maintain per-feature evidence and subject/platform constraints. No Python control-plane fallback.

## Milestone N5 — release hardening

Pass all mandatory U0–U21 and source G/U/T suites on real target hardware. Run adversarial security/privacy, offline, concurrency, GPU pressure, disk-full, power loss, service crash, backup/restore and consent revocation tests. Produce reproducible signed ISO/package artifacts, SBOM/license inventory, update/rollback proof, recovery media, release notes and known-blocker report. Tag `mont-blanc` release only after all release gates pass; no test fixture or emulator is accepted as NVIDIA/phone/sensor evidence.
