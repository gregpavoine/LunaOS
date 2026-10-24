# LunaOS subproject instructions

The parent repository `../AGENTS.md` defines Omarchy commands, privilege escalation, style, testing, and preservation rules. Read and follow it together with these narrower LunaOS constraints.

LunaOS is a subproject of this Omarchy fork. Keep additions inside `lunaos/` unless an integration explicitly requires a small, reviewable change to a parent-owned file. Never replace the parent AGENTS, README, launcher, notification, clipboard, browser, or recovery implementation. LunaOS V0 is additive to the existing installation.

The authoritative detailed product specifications remain in the owner’s local LunaOS workspace and are not copied into this public fork because they contain private preference history. This repository contains only the non-sensitive requirement map and implementation status. Never add personal, intimate, biometric, consent-evidence, secret, or real-device inventory data here.

Mac work may build and test portable code. Do not install Omarchy on macOS or run the Linux baseline. Linux hardware work requires an owner-provided SSH alias, explicit connection authorization, and a read-only baseline before any mutation.

The initial executable action is fixed-fixture checksum only. Do not expose arbitrary shell, privileged actions, payment, external messages, sensor capture, or private-data processing until independent policy, consent and target-hardware contracts exist and are verified. Preserve deny-by-default behavior and truthful capability status.

Use `make test check` from `lunaos/`. Add regression tests only for meaningful behavior changes; all test evidence must distinguish local CI from target-hardware validation. Keep Python and Go dependencies at the standard library for this bootstrap.
