# LunaOS CLI

LunaOS styres vanligvis gjennom snarveiene og LunaOS-menyen (`Super + Space`). Men du kan også styre den gjennom `lunaos` CLI-et. Dette er særlig nyttig når du har en AI-agent som jobber med deg på tilpasning eller konfigurasjon.

CLI-et har tilgang til alt det interne verktøyet som brukes både via menyen og ellers. Du kan se alt som er tilgjengelig ved å kjøre `lunaos` i terminalen.

Det ser omtrent slik ut:

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

Og du kan gå dypere på hver gruppe:

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

Hver kommando tar også `--help`, enten du spør en hel gruppe (`lunaos capture --help`) eller én enkelt kommando (`lunaos capture screenshot --help`).

### Åpne menyen fra terminalen

LunaOS-menyen er også skriptbar, noe som er hendig for dine egne tastebindinger. `lunaos menu` åpner den ved roten, og du kan hoppe rett til ethvert punkt i treet ved å navngi det: `lunaos menu summon style.theme` går rett til temavelgeren, `lunaos menu toggle system` åpner systemmenyen og lukker den igjen hvis den allerede er oppe, og `lunaos menu close` legger den bort.
