# CLI LunaOS

LunaOS обычно управляется через сочетания клавиш и меню LunaOS (`Super + Space`). Но ей также можно управлять через CLI `lunaos`. Это особенно полезно, когда ИИ-агент работает с вами над настройкой или конфигурацией.

CLI имеет доступ ко всем внутренним инструментам, используемым как через меню, так и иным образом. Всё доступное можно увидеть, выполнив `lunaos` в терминале.

Это выглядит примерно так:

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

И можно углубиться в каждую группу:

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

Каждая команда также принимает `--help`, спрашиваете ли вы целую группу (`lunaos capture --help`) или одну команду (`lunaos capture screenshot --help`).

### Открытие меню из терминала

Меню LunaOS также поддаётся скриптингу, что удобно для ваших собственных привязок клавиш. `lunaos menu` открывает его в корне, а вы можете перейти прямо к любой точке дерева, назвав её: `lunaos menu summon style.theme` переходит прямо к выбору темы, `lunaos menu toggle system` открывает системное меню и снова закрывает его, если оно уже открыто, а `lunaos menu close` убирает его.
