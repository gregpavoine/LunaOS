# CLI ya LunaOS

LunaOS kwa kawaida hudhibitiwa kupitia mikato ya vitufe na menyu ya LunaOS (`Super + Space`). Lakini unaweza pia kuidhibiti kupitia CLI ya `lunaos`. Hii inasaidia hasa unapokuwa na ajenti wa AI anayefanya kazi nawe katika kubinafsisha au kusanidi.

CLI ina ufikiaji wa zana zote za ndani zinazotumiwa kupitia menyu na vinginevyo. Unaweza kuona kila kinachopatikana kwa kuendesha `lunaos` katika terminali.

Inaonekana kama hivi:

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

Na unaweza kuchimba zaidi katika kila kundi:

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

Kila amri huchukua `--help` pia, iwe unauliza kundi zima (`lunaos capture --help`) au amri moja (`lunaos capture screenshot --help`).

### Kufungua menyu kutoka terminali

Menyu ya LunaOS inaweza pia kuandikwa kwa hati, ambayo ni muhimu kwa mikato yako mwenyewe ya vitufe. `lunaos menu` huifungua kwenye mzizi, na unaweza kuruka moja kwa moja hadi hatua yoyote katika mti kwa kuitaja: `lunaos menu summon style.theme` huenda moja kwa moja hadi kichaguzi cha mandhari, `lunaos menu toggle system` hufungua menyu ya mfumo na kuifunga tena ikiwa tayari iko wazi, na `lunaos menu close` huiweka mbali.
