# LunaOS CLI

LunaOS sa zvyčajne ovláda cez klávesové skratky a menu LunaOS (`Super + Space`). Ale môžete ho ovládať aj cez `lunaos` CLI. To je obzvlášť užitočné, keď s vami AI agent pracuje na prispôsobení alebo konfigurácii.

CLI má prístup ku všetkým interným nástrojom, ktoré sa používajú cez menu aj inak. Všetko, čo je k dispozícii, uvidíte spustením `lunaos` v termináli.

Vyzerá to asi takto:

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

A môžete sa ponoriť hlbšie do každej skupiny:

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

Každý príkaz tiež prijíma `--help`, či už sa pýtate celej skupiny (`lunaos capture --help`) alebo jedného príkazu (`lunaos capture screenshot --help`).

### Otvorenie menu z terminálu

Menu LunaOS je tiež skriptovateľné, čo je šikovné pre vaše vlastné klávesové väzby. `lunaos menu` ho otvára v koreni a môžete skočiť priamo do ktoréhokoľvek bodu v strome tým, že ho pomenujete: `lunaos menu summon style.theme` ide priamo na výber tém, `lunaos menu toggle system` otvára systémové menu a znova ho zatvorí, ak je už otvorené, a `lunaos menu close` ho odloží.
