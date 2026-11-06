# LunaOS CLI

LunaOS biasanya dikawal melalui kekunci panas dan menu LunaOS (`Super + Space`). Tetapi anda juga boleh mengawalnya melalui CLI `lunaos`. Ini amat membantu apabila anda meminta ejen AI bekerja dengan anda mengenai penyesuaian atau konfigurasi.

CLI mempunyai akses kepada semua perkakasan dalaman yang digunakan melalui menu dan sebaliknya. Anda boleh melihat semua yang tersedia dengan menjalankan `lunaos` dalam terminal.

Ia kelihatan seperti ini:

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

Dan anda boleh menyelami lebih mendalam pada setiap kumpulan:

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

Setiap arahan juga mengambil `--help`, sama ada anda bertanya keseluruhan kumpulan (`lunaos capture --help`) atau satu arahan (`lunaos capture screenshot --help`).

### Membuka menu dari terminal

Menu LunaOS juga boleh diskrip, yang berguna untuk ikatan kekunci anda sendiri. `lunaos menu` membukanya pada akar, dan anda boleh melompat terus ke mana-mana titik dalam pokok dengan menamainya: `lunaos menu summon style.theme` pergi terus ke pemilih Tema, `lunaos menu toggle system` membuka menu Sistem dan menutupnya semula jika ia sudah dipaparkan, dan `lunaos menu close` menyimpannya.
