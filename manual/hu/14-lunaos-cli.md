# LunaOS CLI

Az LunaOSt általában a billentyűparancsokkal és az LunaOS menüvel (`Super + Space`) vezérlik. De az `lunaos` CLI-n keresztül is vezérelheted. Ez különösen hasznos, amikor egy AI-agenttel dolgozol együtt testreszabási vagy konfigurációs feladatokon.

A CLI hozzáfér az összes belső eszközhöz, amelyet mind a menün keresztül, mind máshol használnak. Mindent láthatsz, ami rendelkezésre áll, ha terminálban futtatod az `lunaos` parancsot.

Valahogy így néz ki:

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

És minden csoportban mélyebbre merülhetsz:

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

Minden parancs elfogadja a `--help` kapcsolót is, akár egy egész csoporttól kérdezed (`lunaos capture --help`), akár egyetlen parancstól (`lunaos capture screenshot --help`).

### A menü megnyitása a terminálból

Az LunaOS menü szkriptelhető is, ami jól jön a saját billentyűhozzárendeléseidhez. Az `lunaos menu` a gyökérnél nyitja meg, és a fában egyenesen bármely pontra ugorhatsz, ha megnevezed: az `lunaos menu summon style.theme` egyenesen a témaválasztóhoz megy, az `lunaos menu toggle system` megnyitja a rendszermenüt, és ismét bezárja, ha már nyitva van, az `lunaos menu close` pedig elteszi.
