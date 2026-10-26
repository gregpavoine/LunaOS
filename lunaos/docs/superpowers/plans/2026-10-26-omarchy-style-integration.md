# Omarchy-Style LunaOS Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild LunaOS as an Arch/Omarchy-style distribution extension that reuses the current system source and desktop, keeps only Luna-specific AI services, and ships one verified installer ISO from separate package and ISO repositories.

**Architecture:** The current LunaOS source repository remains the system/configuration source. `gregpavoine/LunaOS-pkgs` owns Arch package recipes and a signed pacman repository; `gregpavoine/LunaOS-iso` owns the branded Archiso build and reuses the established Omarchy installation path. The system source installs the Go Core and Policy Broker through its package/lifecycle conventions and integrates LunaOS controls into its existing Quickshell shell. The existing nested installer and shell remain frozen references until the replacement passes end-to-end acceptance.

**Tech Stack:** Arch Linux packages and `makepkg`, Bash lifecycle scripts, systemd, Go Core/Policy Broker, SQLite, Quickshell/QML, Archiso, GitHub Actions, QEMU/OVMF disposable VM acceptance.

**Spec:** `lunaos/docs/superpowers/specs/2026-10-26-omarchy-style-lunaos-design.md`

## Global Constraints

- Reuse the existing Arch/Omarchy-derived system source for boot, users, system configuration, desktop session, updates and recovery.
- Maintain separate LunaOS package-recipe and ISO-build repositories, following the split used by Omarchy.
- The Policy Broker stays in a distinct service identity and accepts only typed, allowlisted requests over a protected Unix socket. Deny by default; check authorization before scoring and immediately before dispatch. No model output grants authority.
- Python remains outside boot, shell, policy and goal-control paths.
- Release one ISO file plus its checksum. If the image exceeds GitHub's single-asset limit, reduce duplicated package content or use a release host that accepts a single file; do not split the image into user-reassembled parts.
- Keep existing LunaOS code until each retained behavior has a mapped destination and passes parity and installation checks.
- Do not push or run inherited GitHub Actions until their destinations and trust boundaries are reviewed and adapted for LunaOS.
- Do not remove Omarchy applications, services or user configuration from any installed machine as part of this source migration.
- Do not claim Mont Blanc 1.0 features, model inference, hardware support or performance without runtime evidence.
- Build and test against disposable Arch environments only. Do not access a remote Linux host or mutate the owner's Mac/Linux installation.

## Review Focus

- **Missing or invalid package-signing key:** package installation must fail closed; exercise a valid signed package and a package with a bad or absent signature in the isolated package-repository test.
- **Core or Broker unavailable at boot:** the desktop must report the service as unavailable and must not bypass the Policy Broker; test missing socket, broker denial and service restart in the Arch VM acceptance task.
- **Shell starts extra background processes while idle:** keep the Luna panel lazy and event-driven; run the existing shell idle-spawn regression and the focused QML checks.
- **Installer sees an existing host disk:** tests and manual review must prove no host disk is passed through and the installer only sees the disposable VM disk.
- **Single-file ISO exceeds the release limit:** the image job must fail before publication and report the exact size; test the size guard with a deliberately oversized fixture without uploading it.

---

### Task 1: Record the approved Omarchy-style architecture

**Files:**
- Modify: `AGENTS.md` LunaOS subproject architecture paragraph
- Modify: `lunaos/AGENTS.md`
- Create: `lunaos/docs/adr/0005-omarchy-style-integration.md`
- Preserve: `lunaos/docs/adr/0004-omarchy-aligned-implementation.md` as historical decision record
- Modify: `lunaos/docs/architecture/omarchy-aligned.md`
- Modify: `lunaos/docs/DECISIONS.md`
- Modify: `lunaos/docs/OPEN_DECISIONS.md`
- Modify: `lunaos/README.md`

**Interfaces:**
- Consumes: approved design at `lunaos/docs/superpowers/specs/2026-10-26-omarchy-style-lunaos-design.md`.
- Produces: one authoritative statement that the main source tree owns the OS lifecycle, while `LunaOS-pkgs` and `LunaOS-iso` own package recipes and the branded image.

- [ ] **Step 1: Update the subproject guidance and decision record** so they no longer call LunaOS an independently implemented parallel distribution or describe Omarchy as only a reference. Preserve historical ADRs and record the new decision as a superseding ADR.
- [ ] **Step 2: Review all architecture claims** in the listed docs and label current source, planned work and verified runtime separately.
- [ ] **Step 3: Validate documentation consistency** with `git diff --check` and cross-check every repository/path reference against the design.
- [ ] **Step 4: Commit** `docs: define Omarchy-style LunaOS integration` on `feature/omarchy-style-integration`.

### Task 2: Create the LunaOS package repository and native packages

**Files in separate repository `gregpavoine/LunaOS-pkgs`:**
- Create: `pkgbuilds/lunaos-control-plane/PKGBUILD`
- Create: `pkgbuilds/lunaos-control-plane/.omarchy/package.json`
- Create: pinned vendored Go source archive and package documentation
- Modify: `README.md` and create: `UPSTREAM.md`
- Preserve inherited workflows until their trust boundaries and destinations are reviewed; do not push or run them yet
- Create: `scripts/build-repository.sh` after signing-key setup
- Create: `.github/workflows/packages.yml` after reviewing and replacing inherited automation
- Preserve: `README.md` LunaOS package status and `UPSTREAM.md` provenance

**Interfaces:**
- Consumes: Go source from the LunaOS source repository at an immutable commit; existing service entrypoints in `lunaos/services/core/cmd/luna-core`, `lunaos/services/policy`, and `lunaos/services/identity/cmd/luna-owner-setup`; systemd templates in `lunaos/systemd/`.
- Produces: an x86_64 package containing the native Core, Broker and owner-setup binaries and systemd units. Package/repository signing is blocked until LunaOS signing-key ownership and CI secret provisioning are explicitly established; do not install the package with signature checks weakened.

- [ ] **Step 1: Add package-content tests** that inspect each built package and assert expected binaries, systemd units, sysusers entries, file ownership and non-writable service paths.
- [ ] **Step 2: Build each package in a clean Arch container** using `makepkg --cleanbuild --syncdeps --noconfirm`; verify the expected package files and metadata with Pacman tooling.
- [ ] **Step 3: Implement the PKGBUILDs** with pinned source revisions, native Go compilation, explicit runtime dependencies and no build toolchain installed in the target image.
- [ ] **Step 4: Build a signed repository** from the produced packages; publish only after the repository database and package signatures verify with the checked-in public key.
- [ ] **Step 5: Add CI negative checks** proving an invalid or missing signature is rejected by the test pacman configuration; keep the signing private key in Actions secrets, never in Git.
- [ ] **Step 6: Document package build, local repository setup and key rotation**; verify the clean container can install all three packages from the signed repository.
- [ ] **Step 7: Commit and merge locally** the completed package-repository feature after its clean-build and signature checks pass. Do not push.

### Task 3: Install the LunaOS packages through the base system lifecycle

**Files in the main LunaOS source repository:**
- Modify: `default/pacman/pacman-stable.conf`
- Modify: `default/pacman/pacman-rc.conf`
- Modify: `default/pacman/pacman-edge.conf`
- Modify: `install/omarchy-x86_64-only.packages`
- Create: `install/config/lunaos-services.sh`
- Modify: `install/config/all.sh`
- Extend: the closest existing install/config shell test, after locating its current coverage

**Interfaces:**
- Consumes: signed package repository and public key from Task 2; system lifecycle contracts in `bin/omarchy-apply-system` and `install/`.
- Produces: supported package sources and idempotent systemd enablement for LunaOS Core, Broker and owner setup on a fresh LunaOS installation.

- [ ] **Step 1: Add a failing package-configuration test** that renders each supported pacman channel and proves the LunaOS repository is signed, ordered after the base Arch repositories, and unavailable when its signature check fails.
- [ ] **Step 2: Add a failing lifecycle test** that checks service enablement is idempotent, does not start an absent package, and never changes an existing user's data.
- [ ] **Step 3: Add the LunaOS package repository and keyring** through the existing pacman template/rendering path; initially add LunaOS packages only to `install/omarchy-x86_64-only.packages`; do not advertise AArch64 support until package build and install checks pass on that architecture.
- [ ] **Step 4: Implement the idempotent lifecycle leaf** to enable the packaged services using the same `install/config/` conventions as adjacent setup leaves; do not start services before credentials and protected directories are provisioned.
- [ ] **Step 5: Run the focused shell tests** with complete logs, then run Go unit/process tests using `make -C lunaos test check`.
- [ ] **Step 6: Verify clean package installation** in an Arch container and confirm no Go compiler or Python runtime is installed as a LunaOS control-plane dependency.
- [ ] **Step 7: Commit and merge locally** this system-package integration after the focused checks pass. Do not push.

### Task 4: Integrate LunaOS controls into the existing Quickshell desktop

**Files in the main LunaOS source repository:**
- Create: `shell/plugins/lunaos/manifest.json`
- Create: `shell/plugins/lunaos/Main.qml`
- Create: `shell/plugins/lunaos/Panel.qml`
- Create: `shell/plugins/lunaos/Client.qml`
- Create: `bin/omarchy-lunaos`
- Modify: `default/omarchy/omarchy-menu.jsonc` only if the existing menu is the correct discoverability surface after UI review
- Modify: `bin/omarchy` `GROUP_DESCRIPTIONS` only if new routed commands are introduced
- Test: `test/shell.d/lunaos-panel-test.sh` or the nearest existing plugin test

**Interfaces:**
- Consumes: owner-scoped local API and service status from the Go Core; canonical shell IPC via `bin/omarchy-shell`; first-party plugin contract in `shell/plugins/`.
- Produces: one lazily-created LunaOS panel that can read service/goal status and submit owner-authorized typed goal requests without embedding credentials or holding root authority.

- [ ] **Step 1: Add a failing client contract test** for unavailable Core, authenticated status retrieval, denied action, and benign goal submission through the local owner socket.
- [ ] **Step 2: Implement `bin/omarchy-lunaos`** as a narrow CLI client that calls the owner-scoped socket; do not place tokens in QML, environment variables or process arguments.
- [ ] **Step 3: Implement the panel and manifest** using the first-party plugin contract, with UI content constructed only when the panel opens and no independent Quickshell process.
- [ ] **Step 4: Run the shell plugin and command tests**, `qmllint` in an Arch/Qt environment, and the existing idle-spawn regression.
- [ ] **Step 5: Verify the panel in a disposable Wayland VM**: show service unavailable state when the broker is down; confirm a denied operation does not score or dispatch; confirm the permitted benign goal displays recorded evidence.
- [ ] **Step 6: Commit and merge locally** the desktop integration after graphical verification. Do not push.

### Task 5: Create the LunaOS ISO repository using the maintained Omarchy image path

**Files in separate repository `gregpavoine/LunaOS-iso`:**
- Reuse: pinned upstream `omacom/omarchy-iso` checkout and its existing Archiso profile
- Create: `bin/lunaos-iso-make` local-source wrapper
- Modify: `README.md` and create: `UPSTREAM.md`
- Preserve inherited workflows until their trust boundaries and destinations are reviewed; add no LunaOS workflow before source pinning, branding and VM gates are ready
- Reuse: upstream Omarchy installer/configurator, base packages, boot and recovery flow; add only LunaOS branding and package selection.

**Interfaces:**
- Consumes: pinned LunaOS source revision, signed packages from Task 2, and the integrated base package/configuration path from Task 3.
- Produces: exactly one x86_64 UEFI installer ISO and `SHA256SUMS`; installation delegates partitioning and boot setup to the maintained Omarchy installation path.

- [ ] **Step 1: Add profile checks** asserting the LunaOS package set is included once, the source/repository revisions are pinned, and no second installer or duplicate base package closure is introduced.
- [ ] **Step 2: Port the Omarchy ISO profile into the new repository** at a recorded upstream commit; keep changes limited to branding and LunaOS package selection.
- [ ] **Step 3: Build one ISO in an Arch container**; enforce a 1,900,000,000-byte maximum for a GitHub release asset and generate a SHA-256 file.
- [ ] **Step 4: Add the oversize regression** and verify it fails before artifact upload without creating split parts.
- [ ] **Step 5: Run UEFI live boot** in QEMU/OVMF and confirm first-boot setup and the existing desktop reach their actual boot markers.
- [ ] **Step 6: Commit and merge locally** the ISO repository feature after build and live-boot checks pass. Do not push.

### Task 6: Verify install, first boot, services and goal evidence end to end

**Files:**
- Modify: the existing Omarchy ISO acceptance entrypoint in `LunaOS-iso`, retaining upstream naming and behavior
- Modify: `LunaOS-iso/.github/workflows/lunaos-iso.yml`
- Modify: `lunaos/docs/STATUS.md`

**Interfaces:**
- Consumes: the one-file ISO from Task 5, tested Go services from Task 2, and installed system path from Task 3.
- Produces: evidence that a disposable x86_64 UEFI VM installs LunaOS, boots the installed system, starts the actual systemd services, and completes the real benign `local.checksum` owner-authorized goal with persisted evidence; no mocked evidence fixture may satisfy acceptance.

- [ ] **Step 1: Add failing install acceptance assertions** for installed binaries/units, separate service identities, owner-only socket access, service readiness, and completed evidence after reboot.
- [ ] **Step 2: Keep only the VM target disk attached**; assert no host block device or host filesystem mount is passed into QEMU.
- [ ] **Step 3: Run the full install test** with the ISO attached first and the target disk first on reboot; require the explicit installed-target success marker before upload.
- [ ] **Step 4: Run acceptance from a clean build** and retain exact logs, image size and checksum; do not accept source-only or mocked service results.
- [ ] **Step 5: Update STATUS and release notes** with only the steps actually passed, including remaining hardware, update, rollback and Mont Blanc gaps.
- [ ] **Step 6: Commit and merge locally** the acceptance workflow after every gate is green. Do not push or publish.

### Task 7: Retire the duplicate nested distribution profile after parity

**Files:**
- Remove only after Task 6 passes: `lunaos/packaging/arch/` duplicate Archiso profile and installer implementation
- Remove only after Task 4 passes: `lunaos/shell/shell.qml` as a standalone shell entrypoint
- Retain: `lunaos/services/`, Go tests, API contracts, migrations, frozen Rust/Python reference implementations, and relevant build/acceptance fixtures
- Modify: `lunaos/docs/STATUS.md`, `lunaos/docs/DECISIONS.md`, `lunaos/docs/OPEN_DECISIONS.md`, `lunaos/README.md`

**Interfaces:**
- Consumes: verified package, desktop and install parity from Tasks 2–6.
- Produces: one authoritative LunaOS system/image path with historical source preserved where the feature scope still needs it.

- [ ] **Step 1: Prove every removed path has a passing replacement acceptance** by mapping it in the feature matrix; stop if any required Mont Blanc behavior has no replacement.
- [ ] **Step 2: Remove only the duplicate installer/profile and parallel shell entrypoint**; keep service code and test/reference assets.
- [ ] **Step 3: Run all focused LunaOS checks and the image install acceptance** after the removals.
- [ ] **Step 4: Update current-state documents** and verify they contain no claim that an unimplemented feature is complete.
- [ ] **Step 5: Commit the cleanup as its own local feature commit**; merge only after review and passing acceptance. Do not push.

## Branch and execution sequence

Implement natively and sequentially. Use `feature/omarchy-style-integration` for the system-source migration, and separate feature branches in `LunaOS-pkgs` and `LunaOS-iso` for their repositories. Complete and verify one feature before merging it locally into `mont-blanc`. Do not push or publish; the owner handles remote pushes. Do not merge a task that has not passed its stated checks.
