# First-party plugins

These plugins ship with LunaOS and are discovered by the shell at startup.
They use the same `manifest.json` contract as third-party plugins; the
only difference is that the shell flags them with `__isFirstParty: true`.
First-party non-bar plugins are enabled unless listed in `disabledPlugins[]`;
`lunaos.bar` is the default bar option and becomes inactive only while another
`kind: "bar"` plugin is selected. Services and keep-loaded panels are mounted
at startup; other panels, overlays, and menus are loaded on demand.

User-installed plugins live alongside these conceptually but on disk under
`~/.config/lunaos/plugins/<plugin-id>/` rather than in this directory.

| Plugin        | id                        | kinds                   | entry point                           |
|---------------|---------------------------|-------------------------|---------------------------------------|
| Bar           | `lunaos.bar`             | `bar`                   | `bar/Bar.qml`                         |
| Image picker  | `lunaos.image-picker`    | `overlay`               | `image-picker/ImagePicker.qml`        |
| Emojis        | `lunaos.emojis`          | `overlay`               | `emojis/Emojis.qml`                   |
| Clipboard mgr | `lunaos.clipboard`       | `overlay`               | `clipboard/Clipboard.qml`             |
| Reminders     | `lunaos.reminders`       | `overlay`               | `reminders/ReminderFlow.qml`          |
| LunaOS menu  | `lunaos.menu`            | `menu`, `bar-widget`    | `menu/Menu.qml`, `menu/BarWidget.qml` |
| Notifications | `lunaos.notifications`   | `service`               | `notifications/Service.qml`           |
| Audio         | `lunaos.audio`           | `bar-widget`            | `panels/audio/Panel.qml`              |
| Bluetooth     | `lunaos.bluetooth`       | `bar-widget`            | `panels/bluetooth/Panel.qml`          |
| Clock         | `lunaos.clock`           | `bar-widget`            | `panels/clock/BarWidget.qml`          |
| Elsewhen      | `lunaos.elsewhen`        | `bar-widget`            | `panels/elsewhen/Panel.qml`           |
| Monitor       | `lunaos.monitor`         | `bar-widget`            | `panels/monitor/Panel.qml`            |
| Network       | `lunaos.network`         | `bar-widget`            | `panels/network/Panel.qml`            |
| Power         | `lunaos.power`           | `bar-widget`            | `panels/power/Panel.qml`              |
| Tailscale     | `lunaos.tailscale`       | `bar-widget`            | `panels/tailscale/Panel.qml`          |
| Agents   | `lunaos.agents`     | `bar-widget`            | `agents/Panel.qml`               |
| Weather       | `lunaos.weather`         | `bar-widget`            | `panels/weather/BarWidget.qml`        |
| Media         | `lunaos.media`           | `service`, `bar-widget` | `services/media/Service.qml`, `services/media/BarWidget.qml` |
| Battery       | `lunaos.battery`         | `service`               | `services/battery/Service.qml`        |
| Idle          | `lunaos.idle`            | `service`               | `services/idle/Service.qml`           |
| Night light   | `lunaos.nightlight`      | `service`               | `services/nightlight/Service.qml`     |
| Remote session | `lunaos.remote-session` | `service`               | `services/remote-session/Service.qml` |
| Lock screen   | `lunaos.lock`            | `service`               | `lock/Service.qml`                    |
| OSD           | `lunaos.osd`             | `panel`                 | `osd/Osd.qml`                         |
| Polkit agent  | `lunaos.polkit`          | `service`               | `polkit/PolkitAgent.qml`              |
| Dock          | `lunaos.dock`            | `service`               | `dock/Dock.qml`                       |

First-party bar-only widgets also carry manifests next to their QML files,
e.g. `bar/widgets/Workspaces.manifest.json`. Rich popup widgets live in their
own plugin directories, each with its own `manifest.json`.

## Bar

The built-in status bar and default full-bar option. Layout lives in the
top-level `bar:` subtree of `~/.config/lunaos/shell.json` (with the shell
providing [`config/lunaos/shell.json`](../../config/lunaos/shell.json) when
the user has no file). See [`bar/README.md`](bar/README.md) for the widget catalogue
and customization schema.

## Image picker

Fullscreen image-grid selector overlay. Used by `lunaos-menu-images`
(wallpaper picker) and `lunaos-theme-switcher` (theme picker) and any
other caller that wants to present a directory of images with previews.

Two ways to drive it:

- Shell-level summon: `lunaos-shell shell summon lunaos.image-picker '<jsonPayload>'`.
  The payload can carry `imageDirs`, `imageRows`, `selectedImage`,
  `selectionFile`, `doneFile`, `showLabels`, `filterable`. Best for
  in-shell callers that already speak JSON.
- Direct IPC target: `lunaos-shell image-selector open <imageDirs> <imageRowsB64> <selectedImage> <selectionFile> <doneFile> <showLabels> <filterable>`.
  Positional args; `imageRowsB64` is base64-encoded so embedded newlines /
  tabs survive the bash argv handoff. This is what `lunaos-menu-images`
  uses. Colors come from the central shell theme singleton; there is no
  per-call override surface.

The selection round-trip remains file-based: callers create a
`selection_file` and `done_file` (both `mktemp`), pass the paths, and
poll `done_file` for existence. The plugin writes the chosen path into
`selection_file` and touches `done_file` when it's done. `cancel` IPC
clears it without writing a selection.

The plugin has `keepLoaded: true` so the layer-shell window survives
between summons within a single shell session.

The carousel renders only the slices that fit on the screen plus one prefetch slice per side, capped at 33 cards. It preserves overlapping delegates during navigation, decodes images asynchronously at card size, and loads the selected preview before its neighbors. Lazy thumbnail preparation runs below normal CPU and I/O priority in one shared worker pool per image list: one worker on single/dual-core machines and at most two on larger machines. Opening and navigating the picker does not wait for that queue to finish.

## Lock screen

Session-lock surface using Quickshell's native `WlSessionLock` and two
separate PAM services: `lunaos-lock-password` for password auth and,
only when fingerprints are enrolled, `lunaos-lock-fingerprint` for
fingerprint auth. It mirrors the previous lock screen field dimensions,
colors, blurred wallpaper, placeholder, and Hyprland-driven corners.
The plugin sets `keepLoaded: true` so a plugin hot-reload (for example
an installed bar widget changing on disk) does not destroy the lock
client while Hyprland still holds the session lock.

## Polkit agent

Theme-aware authentication dialog for privileged actions. It uses
Quickshell's native `Quickshell.Services.Polkit.PolkitAgent` backend and
runs inside the long-lived `lunaos-shell` process, replacing the old
`polkit-gnome-authentication-agent-1` autostart.

## Dock

App dock shown on floating workspaces, and hidden on tiled ones and behind a fullscreen window. Floating is not a Hyprland layout, so the dock reads each workspace's mode through `lunaos-hyprland-workspace-layout-current`, caches it, and is told to read again over `lunaos-shell lunaos.dock refresh` by the layout commands. `lunaos-shell lunaos.dock apps` toggles its grid of every app.

Its settings live in the `dock:` subtree of `shell.json`: `position` (`top`, `bottom`, `left` or `right`) and `pinned` (desktop ids). A `shell.json` without pins uses the defaults. `lunaos dock` sets both.

## LunaOS menu

Quickshell-powered LunaOS command menu.
The menu UI lives in `menu/Menu.qml` as a first-party `menu` plugin and is
summoned through the shell (`lunaos-shell shell summon lunaos.menu ...`),
so it shares the long-running `lunaos-shell` process instead of starting a
second Quickshell instance.

The menu definition lives outside the shell host code:

- defaults: `default/lunaos/lunaos-menu.jsonc`
- user extensions: `~/.config/lunaos/extensions/lunaos-menu.jsonc`

The shell parses both JSONC files at startup (with `watchChanges: true`
so edits take effect without a restart), evaluates `when:` / `checked:`
bash expressions in a single batched subprocess, and executes the
selected `action:` string directly via `Quickshell.execDetached`. The
long-running shell process keeps the parsed menu in memory, so the
keybind → IPC → visible path costs ~30ms cold.

## Coming soon

- `lunaos.theme-switcher` — folds theme switching into the shell.
