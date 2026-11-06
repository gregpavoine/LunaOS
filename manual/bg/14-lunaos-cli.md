# LunaOS CLI

LunaOS обикновено се управлява чрез клавишните комбинации и менюто на LunaOS (`Super + Space`). Но можеш да го управляваш и чрез CLI-то `lunaos`. Това е особено полезно, когато AI агент работи с теб върху персонализиране или конфигурация.

CLI-то има достъп до всички вътрешни инструменти, които се използват както чрез менюто, така и иначе. Можеш да видиш всичко налично, като изпълниш `lunaos` в терминала.

Изглежда нещо подобно:

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

И можеш да навлезеш по-дълбоко във всяка група:

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

Всяка команда също приема `--help`, независимо дали питаш цяла група (`lunaos capture --help`) или единична команда (`lunaos capture screenshot --help`).

### Отваряне на менюто от терминала

Менюто на LunaOS също може да се скриптира, което е удобно за собствените ти клавишни комбинации. `lunaos menu` го отваря в корена и можеш да скочиш директно до всяка точка в дървото, като я назовеш: `lunaos menu summon style.theme` отива директно на селектора за теми, `lunaos menu toggle system` отваря системното меню и го затваря отново, ако вече е отворено, а `lunaos menu close` го прибира.
