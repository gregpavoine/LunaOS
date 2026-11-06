# LunaOS CLI

LunaOS зазвичай керують через гарячі клавіші й меню LunaOS (`Super + Space`). Але ви також можете керувати ним через CLI `lunaos`. Це особливо корисно, коли AI-агент працює з вами над налаштуванням або конфігурацією.

CLI має доступ до всіх внутрішніх інструментів, які використовуються як через меню, так і в інших місцях. Ви можете побачити все доступне, запустивши `lunaos` в терміналі.

Це виглядає приблизно так:

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

І ви можете зануритися глибше в кожну групу:

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

Кожна команда також приймає `--help`, чи питаєте ви цілу групу (`lunaos capture --help`), чи одну команду (`lunaos capture screenshot --help`).

### Відкриття меню з терміналу

Меню LunaOS також можна скриптувати, що зручно для ваших власних сполучень клавіш. `lunaos menu` відкриває його в корені, а ви можете стрибнути просто до будь-якої точки дерева, назвавши її: `lunaos menu summon style.theme` переходить просто до вибору теми, `lunaos menu toggle system` відкриває системне меню й закриває його знову, якщо воно вже відкрите, а `lunaos menu close` прибирає його.
