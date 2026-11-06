# LunaOS CLI

LunaOS se obično kontrolira prečacima i LunaOS izbornikom (`Super + Space`). Ali možete ga kontrolirati i putem CLI-ja `lunaos`. To je osobito korisno kada vam AI agent pomaže u prilagodbi ili konfiguraciji.

CLI ima pristup svim internim alatima koji se koriste i putem izbornika i na drugi način. Sve što je dostupno možete vidjeti pokretanjem `lunaos` u terminalu.

Izgleda otprilike ovako:

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

I možete kopati dublje po svakoj grupi:

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

Svaka naredba također prima `--help`, bilo da pitate cijelu grupu (`lunaos capture --help`) ili jednu naredbu (`lunaos capture screenshot --help`).

### Otvaranje izbornika iz terminala

LunaOS izbornik također se može skriptirati, što je zgodno za vaše vlastite poveznice tipki. `lunaos menu` otvara ga u korijenu, a izravno možete skočiti na bilo koju točku u stablu imenovanjem: `lunaos menu summon style.theme` ide ravno na birač tema, `lunaos menu toggle system` otvara izbornik sustava i ponovno ga zatvara ako je već otvoren, a `lunaos menu close` ga sprema.
