# LunaOS CLI

LunaOS odatda hotkeylar va LunaOS menyusi (`Super + Space`) orqali boshqariladi. Lekin uni `lunaos` CLI orqali ham boshqarishingiz mumkin. Bu, ayniqsa, AI agent siz bilan moslashuv yoki konfiguratsiya ustida ishlayotganda foydali.

CLI menyu orqali ham, boshqacha ham ishlatiladigan barcha ichki vositalarga kira oladi. Mavjud boʻlgan hamma narsani terminalda `lunaos` ni ishga tushirish orqali koʻrishingiz mumkin.

U taxminan mana bunday koʻrinadi:

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

Va har bir guruhga chuqurroq kirishingiz mumkin:

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

Har bir buyruq `--help` ni ham oladi, siz butun guruhdan (`lunaos capture --help`) yoki bitta buyruqdan (`lunaos capture screenshot --help`) soʻrasangiz ham.

### Menyuni terminaldan ochish

LunaOS menyusi ham skriptlashtirilishi mumkin, bu oʻz tugmalar bogʻlanishlaringiz uchun qulay. `lunaos menu` uni ildizda ochadi va siz nom berish orqali daraxtning istalgan nuqtasiga toʻgʻridan-toʻgʻri sakrashingiz mumkin: `lunaos menu summon style.theme` toʻgʻridan-toʻgʻri mavzu tanlovchiga boradi, `lunaos menu toggle system` tizim menyusini ochadi va agar u allaqachon ochiq boʻlsa yana yopadi va `lunaos menu close` uni yigʻishtiradi.
