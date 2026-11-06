# CLI d'LunaOS

LunaOS normalment es controla mitjançant les dreceres i el menú d'LunaOS (`Super + Space`). Però també el podeu controlar mitjançant la CLI `lunaos`. Això és especialment útil quan teniu un agent d'IA treballant amb vosaltres en la personalització o la configuració.

La CLI té accés a totes les eines internes que es fan servir tant des del menú com d'altres maneres. Podeu veure tot el que hi ha disponible executant `lunaos` al terminal.

És una cosa així:

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

I podeu aprofundir en cada grup:

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

Cada ordre també accepta `--help`, tant si demaneu a tot un grup (`lunaos capture --help`) com a una sola ordre (`lunaos capture screenshot --help`).

### Obrir el menú des del terminal

El menú d'LunaOS també es pot programar, cosa útil per a les vostres pròpies dreceres. `lunaos menu` l'obre a l'arrel, i podeu saltar directament a qualsevol punt de l'arbre anomenant-lo: `lunaos menu summon style.theme` va directament al selector de temes, `lunaos menu toggle system` obre el menú del sistema i el torna a tancar si ja és obert, i `lunaos menu close` el desa.
