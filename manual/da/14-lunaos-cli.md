# LunaOS CLI

LunaOS styres som regel via genvejstasterne og LunaOS-menuen (`Super + Space`). Men du kan også styre den via `lunaos` CLI'en. Det er særligt nyttigt, når du har en AI-agent til at arbejde sammen med dig om tilpasning eller konfiguration.

CLI'en har adgang til alle de interne værktøjer, der bruges både via menuen og på anden vis. Du kan se alt, hvad der er tilgængeligt, ved at køre `lunaos` i terminalen.

Det ser nogenlunde sådan her ud:

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

Og du kan dykke dybere ned i hver gruppe:

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

Hver kommando tager også `--help`, uanset om du spørger en hel gruppe (`lunaos capture --help`) eller en enkelt kommando (`lunaos capture screenshot --help`).

### Åbning af menuen fra terminalen

LunaOS-menuen kan også scriptes, hvilket er praktisk til dine egne tastebindinger. `lunaos menu` åbner den ved roden, og du kan hoppe direkte til et hvilket som helst punkt i træet ved at navngive det: `lunaos menu summon style.theme` går direkte til temavælgeren, `lunaos menu toggle system` åbner systemmenuen og lukker den igen, hvis den allerede er oppe, og `lunaos menu close` lægger den væk.
