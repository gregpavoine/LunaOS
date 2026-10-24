# LunaOS V0 — Omarchy Cleanup and Migration Contract

**Status:** APPROVED PRODUCT DECISIONS / NOT EXECUTED  
**Owner:** Claude Code implementation agents  
**Inputs:** Series 26 (30 user decisions), Omarchy installed state as source of truth  
**Safety:** Always audit first; maintain boot and desktop recovery.  
**Priority:** V0, before shipping LunaOS shell replacement.

## Objective and boundaries

Convert an **existing Omarchy installation** into a minimal black LunaOS workstation while preserving the existing Omarchy application launcher, all social/professional web apps, productivity and development tools requested by the user, Hyprland, network and boot/recovery stack. Replace only the designated desktop surfaces when tested LunaOS counterparts exist. The user's selections have higher priority than Omarchy defaults, but cannot justify destroying system dependencies.

**Important release version caveat:** Omarchy 3.x / 4.0 / 4.5 can have different packages, terminals, shell UI and screenshot utilities. Query the actual installed package list, display entries, installed Omarchy branch/version and configuration before applying changes. Never use a hardcoded removal list without inventory. Especially do not interpret "replace" as "uninstall its dependency tree".

## Decision matrix — exactly 30 selections

| # | Feature / program | User decision | Claude Code implementation outcome |
|--:|---|---|---|
| 1 | Omarchy wallpapers | Replace with LUNA | Solid black or LunaOS-controlled visual state; no wallpaper process needed; keep rollback |
| 2 | Omarchy themes / switcher | Replace with LUNA | LunaOS theme tokens and settings; disable old switcher only after replacement works |
| 3 | Quickshell/Omarchy status bar | Replace with LUNA | Native LunaOS minimal bottom bar; retain running shell until functional replacement; do not remove quickshell if other Omarchy UI depends on it |
| 4 | Main menu / app launcher | **Keep** | Preserve launcher, existing Super+Space behavior and search |
| 5 | Lock screen / idle animations | Replace with LUNA | Secure LunaOS lock/idle UI; **never disable secure locking first** |
| 6 | LibreOffice Writer/Calc/Impress | **Remove** | Remove only installed packages clearly belonging to LibreOffice; inspect dependent packages |
| 7 | Obsidian / Omawrite | **Keep** | Preserve data, launchers and configuration |
| 8 | Kdenlive / Omacut | **Remove** | Remove installed application packages, not shared ffmpeg codecs/libraries |
| 9 | OBS Studio | **Keep** | Preserve capture/recording and all OBS profiles |
| 10 | Pinta / Xournal++ | **Remove** | Remove installed application packages only |
| 11 | Chromium | **Keep** | Preserve browser profiles, extensions and web-app integration |
| 12 | LocalSend | **Keep** | Preserve transfer configuration |
| 13 | Cliamp | **Keep** | Preserve player and command/launcher |
| 14 | Moonlight | **Remove** | Remove exact installed client if present; never touch NVIDIA drivers |
| 15 | Omacalc | **Remove** | Remove only calculator application package |
| 16 | Neovim + Omarchy development configuration | **Keep** | Preserve editor plugins/settings |
| 17 | Foot terminal | **Keep** | Keep Foot and selected default terminal; when on Omarchy 3.x retain its existing default too |
| 18 | Docker interfaces and container tools | **Keep** | Preserve container runtime, images, volumes, compose and Docker TUIs |
| 19 | Disks and Disk Usage tools | **Keep** | Retain tools appropriate to installed Omarchy version |
| 20 | Codex, Gemini, OpenCode, Claude and other AI entrypoints | **Keep all** | Preserve wrappers, agent selector, lazy installers and configurations |
| 21 | HEY, Basecamp, professional web apps | **Keep** | Preserve desktop entries and Chromium profiles |
| 22 | Discord, X and social web apps | **Keep** | Preserve desktop entries |
| 23 | Google Maps and YouTube | **Keep** | Preserve entries |
| 24 | Google Contacts, Messages, Photos | **Keep** | Preserve entries |
| 25 | WhatsApp Web and Zoom | **Keep** | Preserve entries |
| 26 | Omarchy clipboard history | **Keep** | Preserve shortcut, backend and history policy |
| 27 | Omarchy notifications and notification center | **Keep** | Preserve backend/UI, coexist with LUNA notifications rather than replacing |
| 28 | Omarchy keyboard shortcuts | Reorganize for LUNA | Inventory and reconcile conflicts; keep terminal, launcher, rescue, lock, screenshot, clipboard controls |
| 29 | Omarchy update/packages UI | Replace UI, **keep engine** | LUNA frontend for existing pacman/Omarchy update channels, rollback/snapshot facilities |
| 30 | Screenshot/screen-recording/dictation | **Keep** | Preserve existing utilities, hotkeys and privacy implications |

## Removal policy

**Remove when locally installed and pacman allows the transaction:** LibreOffice package variant(s), Kdenlive, Omacut, Pinta, Xournal++, Moonlight, Omacalc. Resolve exact names using `pacman -Q`, `pacman -Qs`, `pacman -Ql`, `.desktop` Exec and `pacman -Qo`. Examples vary by distribution release. A program may be a local script rather than a pacman package: investigate before touching it.

**Never remove:** Hyprland, login/session manager, required Quickshell modules, Omarchy menu/launcher, notifications, clipboard, NVIDIA/CUDA stack, PipeWire, networking, bootloader, snapshot/recovery utilities, pacman/Omarchy update back end, user-created web app/TUI launchers, kept application data.

**Crucial:** DO NOT run `omarchy-remove-preinstalls` / `Remove > Preinstalls`: it removes apps explicitly selected for retention and can remove user-created web app/TUI launchers. DO NOT use `pacman -Rns`, orphan sweeps, blanket wildcard deletions, or delete `.config` wholesale. Treat removing wallpaper background settings and other shell visuals as reversible configuration migrations.

## Audit and execution

1. Create a clean Git branch for LunaOS changes and export a manifest: kernel, distro, version, `pacman -Qqe`, `pacman -Qq`, `systemctl --user --state=running`, installed desktop entries, Omarchy channel and configured bindings.
2. Verify usable rescue shell, boot entry and **actual restorable snapshot** (not just successful snapshot creation). Store baseline of `~/.config/hypr`, Omarchy config, Quickshell config, user-local `.desktop` entries and keyboard mappings.
3. Run `./lunaos-v0-cleanup.sh` (audit-only default); review its exact candidate list and `pacman -R --print` dependency resolver result.
4. Review other installations / custom dependencies, then run `./lunaos-v0-cleanup.sh --apply` interactively. Never execute remotely or unattended without a separately designed authorization mechanism and explicit operator approval.
5. Remove stale shortcuts for removed applications **only after** taking a diff of config files. Leave all selected retained hotkeys and customized bindings intact.
6. Implement LunaOS wallpaper/theme/bottom bar/lock/updater frontend as reversible user services/config overlays. Keep Omarchy launcher and notifications unchanged.
7. Verify first boot, login/auto-login behavior, terminal rescue, Hyprland workspaces, RTX GPU and ComfyUI unaffected; ship a rollback instruction.

## Shell UI implementation

- Black background with no decorative wallpaper and no desktop icons; LUNA presence remains unobtrusive.
- LunaOS bottom bar must not break Omarchy menu launcher (Super+Space) or notification center.
- LunaOS lock screen must enforce real lock; idle modes cannot expose private vault data.
- Keyboard control takes precedence over voice-only navigation. Ensure urgent stop for LUNA remains accessible via keyboard.
- Omarchy native notifications, clipboard, screenshots, recording and dictation remain. Integrate LUNA using existing interfaces and explicit permission scoping.
- Updates UI should call the supported Omarchy / package manager mechanism; preserve snapshot and safe rollback. Never replace pacman or Omarchy migration logic.
- No unsafe forced removal of shared graphical stack on NVIDIA.

## Verification and release gates (mandatory)

- T01: `--audit` performs no mutation; its output identifies installed candidate packages and every retained category.
- T02: `--apply` requires both an interactive TTY and phrase confirmation, then pacman confirmation.
- T03: Never uses `pacman -Rns`, orphan deletion, `--nodeps` or `--cascade`.
- T04: Every retained application opens, or any preexisting missing app is reported and not represented as newly broken.
- T05: Omarchy launcher is available via existing shortcut; web apps still launch with original browser profiles.
- T06: Clipboard and notification history operate, including shortcuts and UI.
- T07: Terminal (Foot, or existing 3.x terminal) and Neovim launch.
- T08: Docker commands, containers, images and volumes unchanged.
- T09: Screenshot, screen recording, OBS and dictation continue to work.
- T10: Login/lock/unlock, suspend/resume, GPU and networking work.
- T11: LunaOS bar, dark desktop, and LUNA agent coexist with menu and kept notifications.
- T12: Omarchy update (including its release migrations) and recovery mechanism remain functional.
- T13: All removed application categories absent; no unrelated pacman package changed.
- T14: Simulated failed UI migration can be rolled back to working Omarchy shell from TTY.
- T15: Ensure no recorded or webcam content is accessible to the wrong privacy context (LUNA Private Vault/BLIND MODE remain independent P0 requirements).

## Relation to master specification

This Series 26 addendum **overrides all earlier conflicting LunaOS V0 app-removal and desktop-retention recommendations**. It does not supersede LUNA Core, Intimacy Core, BLIND MODE or iPhone Sync requirements for Mont Blanc 1.0. V0 cleanup may run before those services are complete. The master spec should reference this addendum as the definitive Omarchy keep/remove/replace matrix.

## Definition of done

All 30 choices are reflected in effective configuration; removed apps are uninstalled if previously present and safely removable; kept apps and web apps remain accessible; replacement UI has smoke tests and tested rollback; native Omarchy package update and recovery still work. No production-readiness claim until tested on the target machine.
