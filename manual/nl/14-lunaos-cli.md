# LunaOS CLI

LunaOS wordt meestal bediend via de sneltoetsen en het LunaOS-menu (`Super + Space`). Maar je kunt het ook bedienen via de `lunaos` CLI. Dit is bijzonder handig wanneer je een AI-agent met je laat werken aan aanpassing of configuratie.

De CLI heeft toegang tot alle interne tooling die zowel via het menu als anderszins wordt gebruikt. Je kunt alles zien wat beschikbaar is door `lunaos` uit te voeren in de terminal.

Het ziet er ongeveer zo uit:

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

En je kunt dieper duiken in elke groep:

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

Elk commando accepteert ook `--help`, of je nu een hele groep vraagt (`lunaos capture --help`) of een enkel commando (`lunaos capture screenshot --help`).

### Het menu openen vanaf de terminal

Het LunaOS-menu is ook scriptbaar, wat handig is voor je eigen keybindings. `lunaos menu` opent het bij de root, en je kunt rechtstreeks naar elk punt in de boom springen door het te benoemen: `lunaos menu summon style.theme` gaat direct naar de themakiezer, `lunaos menu toggle system` opent het systeemmenu en sluit het weer als het al open is, en `lunaos menu close` bergt het op.
