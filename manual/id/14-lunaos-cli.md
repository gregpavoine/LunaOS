# CLI LunaOS

LunaOS biasanya dikontrol melalui hotkey dan menu LunaOS (`Super + Space`). Namun Anda juga dapat mengontrolnya melalui CLI `lunaos`. Ini sangat membantu ketika Anda memiliki agen AI yang bekerja dengan Anda dalam kustomisasi atau konfigurasi.

CLI memiliki akses ke semua perkakas internal yang digunakan baik melalui menu maupun sebaliknya. Anda dapat melihat semua yang tersedia dengan menjalankan `lunaos` di terminal.

Ia terlihat seperti ini:

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

Dan Anda dapat menyelami lebih dalam pada setiap grup:

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

Setiap perintah juga menerima `--help`, apakah Anda menanyakan seluruh grup (`lunaos capture --help`) atau satu perintah (`lunaos capture screenshot --help`).

### Membuka menu dari terminal

Menu LunaOS juga dapat diskrip, yang berguna untuk keybinding Anda sendiri. `lunaos menu` membukanya di akar, dan Anda dapat langsung melompat ke titik mana pun dalam pohon dengan menamainya: `lunaos menu summon style.theme` langsung menuju pemilih tema, `lunaos menu toggle system` membuka menu sistem dan menutupnya lagi jika sudah terbuka, dan `lunaos menu close` menutupnya.
