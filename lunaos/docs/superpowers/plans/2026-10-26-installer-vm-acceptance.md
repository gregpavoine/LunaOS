# LunaOS Installer VM Acceptance Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Prove that the published LunaOS install ISO can install to a disposable UEFI VM disk, boot the installed target, start its real Core/Broker services and complete one benign goal.

**Architecture:** Generate a test-only Archinstall configuration and credentials in a temporary directory, pass them into an isolated QEMU guest, and exercise Archinstall and the shipped target finalizer against a disposable virtual disk. The test configuration, account, disk selection and any test-only access mechanism stay out of release images.

**Tech Stack:** Arch Linux, Archinstall, Bash, Go, QEMU/OVMF, Expect, SQLite and the existing LunaOS HTTP/Unix-socket contracts.

**Spec:** `lunaos/docs/superpowers/specs/2026-10-26-standalone-installable-iso-design.md`; `lunaos/docs/LUNAOS_MONT_BLANC_1_0_CODEX_UNIFIED_SPEC.md` §§4, U11, U15, U16, U19; `lunaos/docs/STATUS.md`.

## Global Constraints

- First target remains x86_64 UEFI and the test disk is a freshly created QEMU image under a private temporary directory.
- No target path, credentials or acceptance-only service is included in the shipped Archinstall template or release ISO.
- Disk operations run only inside the disposable QEMU guest and only against the explicit virtual target device.
- Do not run against, connect to or modify the owner's Linux installation.
- The installed target must use the actual LunaOS package, finalizer, systemd units, Policy Broker, Core and real bounded checksum action.
- Report image build, live boot, disk install, installed boot, service state and benign-goal execution as separate results.

## Review Focus

- Missing or ambiguous target device: the generated CI config must name only `/dev/vda`; test the guard before starting QEMU.
- Installer or finalizer failure: stop the test, preserve redacted logs and do not count a live boot as install success.
- Disk config or account credentials leaking into release image: test the production Archinstall template remains disk-neutral and credential-free.
- Broker/Core unavailable after installed boot: the acceptance check must fail unless both real systemd services are active and the Core API answers.
- Goal action denied, stuck or missing evidence: the smoke goal must complete through the production owner API and the test must verify its durable evidence.

## Task 1: Generate and validate isolated installer test inputs

**Files:**
- Create: `lunaos/scripts/generate-install-smoke-config.sh`
- Create: `lunaos/tests/install-smoke-config-test.sh`
- Modify: `lunaos/Makefile`

**Interfaces:**
- Consumes: the shipped Archinstall JSON template and an output directory.
- Produces: `user_configuration.json` for a fixed two-partition `/dev/vda` UEFI layout and `user_credentials.json` for a CI-only owner account; the helper must require an explicit disposable-test mode and create credentials with mode `0600`.

- [x] Write tests for output path permissions, `/dev/vda`-only disk selection, bounded UEFI partition layout, test account configuration, and absence of disk/credential fields in the shipped template.
- [x] Run the new test and confirm it fails because the generator is absent.
- [x] Implement the generator using `jq` and a temporary per-run credential hash; never print credential contents.
- [x] Run the focused test, then `make test check`.
- [x] Commit the config generator and its tests (`6b4d031d`).

## Task 2: Install and boot the actual target in QEMU

**Files:**
- Create: `lunaos/scripts/test-iso-install.sh`
- Create: `lunaos/scripts/test-iso-install.expect`
- Create: `lunaos/packaging/arch/install-test/PKGBUILD`
- Create: `lunaos/packaging/arch/install-test/lunaos-install-smoke.service`
- Create: `lunaos/packaging/arch/install-test/lunaos-install-smoke.go`
- Create: `lunaos/tests/install-smoke-script-test.sh`
- Modify: `lunaos/Makefile`

**Interfaces:**
- Consumes: an ISO, the generated configuration and credentials from Task 1, and the production installer launcher.
- Produces: exit status `0` only after the production ISO installs to the fresh QEMU disk, the installed image boots, Core and Broker are active, and a real owner-authorized `local.checksum` goal reaches completed state with SHA-256 evidence.

- [x] Add a host-platform preflight regression; the full install, service and evidence assertions are implemented in the VM smoke package and harness.
- [x] Run the focused preflight test and confirm it fails while the harness is absent.
- [x] Implement the QEMU/OVMF harness with a dedicated seed medium for test-only config; test config and credentials remain outside the release ISO and the test package is injected only in the CI guest.
- [ ] Run the script on the Arch CI runner against the built ISO; require install, reboot, service and goal evidence markers. Blocked until this branch is pushed and Actions runs.
- [x] Run the full portable Go and installer test suite, shell syntax checks and `git diff --check`.
- [x] Commit the VM acceptance harness and test package (`da0d06dc`); the explicit live-to-installed reboot fix is `e503cbe4`.

## Task 3: Gate the ISO artifact and update factual release status

**Files:**
- Modify: `.github/workflows/lunaos-archiso.yml`
- Modify: `lunaos/docs/STATUS.md`

**Interfaces:**
- Consumes: the real ISO and exit status from Task 2.
- Produces: CI evidence separating image build, live boot, install-to-disk, installed boot, service health, and goal completion; no artifact may be described as install-verified unless all Task 2 assertions pass.

- [x] Wire the install acceptance command after ISO creation and live boot, before artifact upload.
- [ ] Confirm the workflow gate against a live GitHub Actions run; blocked until this branch is pushed.
- [x] Add the install-and-reboot test to the Arch workflow and retain only its serial log on failure; private VM inputs and disk are removed by the harness.
- [x] Update status documentation with exact local outcomes and remaining install-verification blockers; no new artifact checksum exists yet.
- [x] Run YAML parsing, `make test check`, and `git diff --check` locally.
- [x] Commit the CI gate and factual status (`da0d06dc`).

