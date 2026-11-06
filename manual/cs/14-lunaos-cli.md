# CLI LunaOS

LunaOS se obvykle ovládá klávesovými zkratkami a nabídkou LunaOS (`Super + Space`). Můžete ho ale ovládat také prostřednictvím CLI `lunaos`. To je obzvlášť užitečné, když s vámi na přizpůsobení nebo konfiguraci pracuje AI agent.

CLI má přístup ke všem interním nástrojům, které se používají jak přes nabídku, tak jinak. Vše, co je k dispozici, uvidíte spuštěním `lunaos` v terminálu.

Vypadá to asi takto:

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

A u každé skupiny se můžete ponořit hlouběji:

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

Každý příkaz přijímá také `--help`, ať už se ptáte celé skupiny (`lunaos capture --help`) nebo jednotlivého příkazu (`lunaos capture screenshot --help`).

### Otevření nabídky z terminálu

Nabídka LunaOS je také skriptovatelná, což se hodí pro vaše vlastní klávesové zkratky. `lunaos menu` ji otevře v kořenu a můžete skočit přímo na libovolný bod ve stromu jeho pojmenováním: `lunaos menu summon style.theme` jde přímo do výběru motivu, `lunaos menu toggle system` otevře systémovou nabídku a zase ji zavře, pokud už je nahoře, a `lunaos menu close` ji uklidí.
