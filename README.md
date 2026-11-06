# LunaOS

LunaOS is a personal fork of [Omarchy](https://omarchy.org), DHH's beautiful, fun & agentic Linux distribution. It keeps the upstream base (Arch, Hyprland, the command set, the Quickshell desktop) and adds LUNA: a local control plane that decides what the assistant may do, keeps sensitive data on the machine, and runs its language model locally.

Read more about the upstream project at [omarchy.org](https://omarchy.org).

## What the fork adds

- **LUNA brain (Go)**: `luna-core` runs goals, `luna-policy-broker` grants or refuses each action, and `luna-owner-setup` handles identity. The broker denies by default, keeps every budget at zero, refuses any `payment.*` action, and only accepts the `general` privacy class.
- **Privacy**: a BLIND mode that stops capture, a per-person consent model with 13 scopes, and an encrypted vault (AES-GCM) for sensitive records. Consent is never given on someone else's behalf.
- **Local inference**: `luna-inference` supervises `llama-server` on loopback only, loads only models whose SHA-256 matches the manifest in [`lunaos/models/manifest.json`](lunaos/models/manifest.json), and marks every output as untrusted. There is no cloud fallback.
- **Intimacy and discovery**: the relationship and library features, plus discovery connectors for Reddit and Civitai that honour `Retry-After`. Nothing is fetched from those services in tests.
- **Series 26 defaults and LUNA visuals**: the Series 26 choices are the defaults, with a black theme and a bottom bar.
- **Systemd units**: the LUNA services run with hardened units (`NoNewPrivileges`, `ProtectSystem=strict`, dedicated users) and a restart budget.

## Repository layout

- [`lunaos/`](lunaos/): the Go control plane (`services/`), shared contracts (`contracts/`), systemd units (`systemd/`, `services/inference/systemd/`), and the rebrand tooling (`rebrand/`).
- [`lunaos/docs/design/`](lunaos/docs/design/): the fork design.
- [`lunaos/docs/plans/`](lunaos/docs/plans/): the implementation plans per track, and [`OWNERSHIP.md`](lunaos/docs/plans/OWNERSHIP.md), which says who owns which files.
- [`lunaos/contracts/`](lunaos/contracts/README.md): the frozen action and consent-scope registries, and the Core OpenAPI description.
- [`lunaos/rebrand/`](lunaos/rebrand/README.md): the replayable Omarchy-to-LunaOS rename and the upstream sync procedure.
- Everything else is the Omarchy base, including `bin/`, `config/`, `default/`, `install/`, and `shell/`.

## Prerequisites

There are two ways to use this repository, and they need different things.

**To build and test the control plane** (any Linux or macOS machine):

- Go 1.24.13. `go.mod` pins `toolchain go1.24.13`, so a Go 1.24 install downloads it automatically when the network is available.
- A C compiler (`gcc` or `clang`) with `CGO_ENABLED=1`, because `github.com/mattn/go-sqlite3` is built with cgo.
- `make`, `bash`, `python3` (used by the JSON check), `gofmt` (ships with Go).
- `curl`, for the Linux development launcher `scripts/dev-native.sh`.

**To run the services as LunaOS** (Linux only, on an Arch-based system with systemd):

- The upstream base packages from the Arch repositories, plus the `install/lunaos-*.packages` lists.
- The LUNA binaries installed as `/usr/bin/luna-core`, `/usr/bin/luna-policy-broker`, `/usr/bin/luna-inference`, `/usr/bin/luna-discovery`, `/usr/bin/luna-owner-setup`, `/usr/bin/luna-vault`. `make build` produces them under `lunaos/build/local/`.
- The system users and groups from `lunaos/systemd-sysusers/lunaos.conf` and `lunaos/services/inference/systemd/luna-inference.sysusers`, applied with `systemd-sysusers`.
- The unit files from `lunaos/systemd/` and `lunaos/services/inference/systemd/`, installed under `/usr/lib/systemd/system/`.
- For local inference: a `llama-server` binary at `/usr/bin/llama-server` (llama.cpp), the manifest at `/usr/share/luna-inference/manifest.json`, and a model downloaded with `luna-inference fetch` and matching its SHA-256. The optional environment file `/etc/luna-inference/luna-inference.env` sets `LUNA_INFERENCE_MODEL`, `LUNA_INFERENCE_GPU_LAYERS` and `LUNA_INFERENCE_CTX_SIZE`.

**Not yet in the repository:** the step that installs those binaries, manifest and units on a system is part of integration track Z and does not exist yet. There is no package, no installer, and no ISO build for LunaOS. Until that step lands, the services can be built and run by hand, but they are not installed as a system.

## Building and testing the control plane

The Go module is `lunaos.local/lunaos`, under `lunaos/`. It pins Go 1.24.13 (`toolchain go1.24.13` in `go.mod`).

```bash
make -C lunaos test check
GOOS=linux go vet ./...   # run from lunaos/
```

`test-process-linux` runs the process-level test and needs Linux. `systemd-analyze verify` on the units needs a disposable Linux container; the local check reports it as skipped.

## Status

Verified on macOS: the Go test suite, `go vet` for Linux, and the rebrand check (`lunaos/rebrand/rebrand.sh --check .`).

Still open:

- The integration track is not finished: the `lunaos-luna` CLI, the STOP LUNA hotkey, and the Core route that sends generation to the inference service are not in the tree yet.
- The Linux-only suites, including the process test and `systemd-analyze verify`.
- The real `llama-server` test, which needs a model from the manifest and a llama.cpp package for the platform.
- Live Reddit and Civitai checks. The connectors are tested against fakes only.

## The Omarchy manual

The manual lives in [`manual/`](manual/). It is inherited from Omarchy and is being rewritten for LunaOS; where it still describes upstream behaviour, the code is the reference.

- [Welcome to LunaOS!](manual/01-welcome-to-lunaos.md)

**The Basics**

- [Getting Started](manual/02-getting-started.md)
- [Coming From Mac or Windows](manual/03-coming-from-mac-or-windows.md)
- [Navigation](manual/04-navigation.md)
- [The top bar](manual/05-the-top-bar.md)
- [Themes](manual/06-themes.md)
- [Hotkeys](manual/07-hotkeys.md)
- [Unified Clipboard & History](manual/08-unified-clipboard-history.md)
- [Reminders](manual/09-reminders.md)
- [Notices](manual/10-notices.md)
- [Text Extraction & Dictation](manual/11-text-extraction-dictation.md)
- [Screenshots & Recording](manual/12-screenshots-recording.md)
- [Toggles, idle & screensaver](manual/13-toggles-idle-screensaver.md)
- [LunaOS CLI](manual/14-lunaos-cli.md)

**The Applications**

- [Terminal](manual/15-terminal.md)
- [Neovim](manual/16-neovim.md)
- [AI](manual/17-ai.md)
- [Development Tools](manual/18-development-tools.md)
- [Shell Tools](manual/19-shell-tools.md)
- [Shell Functions](manual/20-shell-functions.md)
- [TUIs](manual/21-tuis.md)
- [GUIs](manual/22-guis.md)
- [Browsers](manual/23-browsers.md)
- [Commercial apps/services](manual/24-commercial-apps-services.md)
- [Web Apps](manual/25-web-apps.md)
- [Gaming](manual/26-gaming.md)
- [Filling out PDFs](manual/27-filling-out-pdfs.md)
- [Windows VM](manual/28-windows-vm.md)
- [Other Packages](manual/29-other-packages.md)

**Configuration**

- [Updates](manual/30-updates.md)
- [Dotfiles](manual/31-dotfiles.md)
- [Shell plugins](manual/32-shell-plugins.md)
- [Monitors](manual/33-monitors.md)
- [Keyboard, Mouse, Trackpad](manual/34-keyboard-mouse-trackpad.md)
- [Networking](manual/35-networking.md)
- [System sleep](manual/36-system-sleep.md)
- [Hardware authentication](manual/37-hardware-authentication.md)
- [Fonts](manual/38-fonts.md)
- [Backgrounds](manual/39-backgrounds.md)
- [Prompt](manual/40-prompt.md)
- [Branding](manual/41-branding.md)
- [Common tweaks](manual/42-common-tweaks.md)
- [Making your own theme](manual/43-making-your-own-theme.md)

**The Rest**

- [Mac support](manual/44-mac-support.md)
- [Troubleshooting](manual/45-troubleshooting.md)
- [FAQ](manual/46-faq.md)
- [System snapshots](manual/47-system-snapshots.md)
- [Security](manual/48-security.md)
- [LunaOS on...](manual/49-lunaos-on.md)
- [Dual Boot Install](manual/50-dual-boot-install.md)
- [Unattended Installs](manual/51-unattended-installs.md)

## License

LunaOS is released under the [MIT License](https://opensource.org/licenses/MIT).
