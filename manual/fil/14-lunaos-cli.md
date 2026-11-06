# LunaOS CLI

Karaniwang kinokontrol ang LunaOS sa pamamagitan ng mga hotkey at ng menu ng LunaOS (`Super + Space`). Ngunit maaari mo rin itong kontrolin sa pamamagitan ng `lunaos` CLI. Partikular itong nakakatulong kapag mayroon kang isang AI agent na nakikipagtulungan sa iyo sa pag-customize o configuration.

Ang CLI ay may access sa lahat ng internal tooling na ginagamit parehong sa pamamagitan ng menu at kung hindi man. Makikita mo ang lahat ng available sa pamamagitan ng pagpapatakbo ng `lunaos` sa terminal.

Mukha itong ganito:

```
~ ❯ lunaos
LunaOS command center

Usage:
  lunaos <command> [args...]
  lunaos commands [--all] [--json] [--check]
  lunaos <group> --help
  lunaos <group> <command> --help

Common commands:
  lunaos update              Update LunaOS and system packages
  lunaos theme list          List available themes
  lunaos theme set <name>    Apply a theme
  lunaos font list           List available fonts
  lunaos screenshot          Take a screenshot
  lunaos debug               Print debugging information

Groups:
  agent          AI coding agents, their usage, and subscription accounts
  audio          Audio input and output controls
  bar            LunaOS shell bar layout and settings
  battery        Battery status helpers
  bluetooth      Bluetooth device controls
  branch         LunaOS git branch management
  branding       About and screensaver branding
  brightness     Display and keyboard brightness
  capture        Screenshots and screen recording
  channel        LunaOS release channel management
  clipboard      Clipboard helpers
  cmd            Command and shortcut helpers
  config         System configuration helpers
  debug          Diagnostics and support logs
  ...
```

At maaari kang sumisid nang mas malalim sa bawat grupo:

```
~ ❯ lunaos capture
Capture commands — Screenshots and screen recording:
  lunaos capture qr                                                                                                                                                                                                       Decode a QR code from a screenshot region
  lunaos capture screenrecording [--fullscreen] [--with-desktop-audio] [--with-microphone-audio] [--with-webcam] [--webcam-device=<device>] [--webcam-size=<small|medium|large>] [--resolution=<size>] [--stop-recording]  Start or stop screen recording
  lunaos capture screenrecording with webcam                                                                                                                                                                              Pick a webcam and start a screen recording with it
  lunaos capture screenshot [smart|region|windows|fullscreen|scroll] [copy|save]                                                                                                                                          Take a screenshot
  lunaos capture text                                                                                                                                                                                                     Extract text from a screenshot region with OCR
  lunaos capture webcam resize <smaller|larger|reset|small|medium|large>                                                                                                                                                  Resize the active webcam recording overlay
```

Tumatanggap din ang bawat command ng `--help`, magtanong ka man sa isang buong grupo (`lunaos capture --help`) o sa isang nag-iisang command (`lunaos capture screenshot --help`).

### Pagbubukas ng menu mula sa terminal

Ang menu ng LunaOS ay scriptable din, na madaling gamitin para sa sarili mong mga keybinding. Binubuksan ito ng `lunaos menu` sa root, at maaari kang tumalon nang diretso sa anumang punto sa tree sa pamamagitan ng pagpapangalan dito: ang `lunaos menu summon style.theme` ay dumidiretso sa tagapili ng tema, ang `lunaos menu toggle system` ay binubuksan ang menu ng sistema at isinasara ito ulit kung nakataas na ito, at ang `lunaos menu close` ay inaalis ito.
