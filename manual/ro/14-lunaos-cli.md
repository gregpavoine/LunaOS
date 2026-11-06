# CLI-ul LunaOS

LunaOS este controlat de obicei prin scurtături și meniul LunaOS (`Super + Space`). Dar îl poți controla și prin CLI-ul `lunaos`. Este deosebit de util atunci când ai un agent AI care lucrează cu tine la personalizare sau configurare.

CLI-ul are acces la toate instrumentele interne folosite atât prin meniu, cât și altfel. Poți vedea tot ce este disponibil rulând `lunaos` în terminal.

Arată cam așa:

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

Și poți aprofunda fiecare grup:

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

Fiecare comandă acceptă și `--help`, fie că întrebi un întreg grup (`lunaos capture --help`), fie o singură comandă (`lunaos capture screenshot --help`).

### Deschiderea meniului din terminal

Meniul LunaOS este și el programabil, ceea ce este util pentru propriile tale mapări. `lunaos menu` îl deschide la rădăcină, și poți sări direct la orice punct din arbore numindu-l: `lunaos menu summon style.theme` merge direct la selectorul de teme, `lunaos menu toggle system` deschide meniul de sistem și îl închide din nou dacă este deja afișat, iar `lunaos menu close` îl pune deoparte.
