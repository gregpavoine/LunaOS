# LunaOS CLI

LunaOS di solito si controlla tramite le scorciatoie e il menu LunaOS (`Super + Space`). Ma puoi anche controllarlo tramite la CLI `lunaos`. Questo è particolarmente utile quando hai un agente AI che lavora con te su personalizzazione o configurazione.

La CLI ha accesso a tutti gli strumenti interni usati sia tramite il menu che altrimenti. Puoi vedere tutto ciò che è disponibile eseguendo `lunaos` nel terminale.

Ha un aspetto simile a questo:

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

E puoi approfondire ogni gruppo:

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

Ogni comando accetta anche `--help`, che tu chieda un intero gruppo (`lunaos capture --help`) o un singolo comando (`lunaos capture screenshot --help`).

### Aprire il menu da terminale

Anche il menu LunaOS è scriptabile, il che è comodo per le tue combinazioni. `lunaos menu` lo apre alla radice, e puoi saltare dritto a qualsiasi punto dell'albero nominandolo: `lunaos menu summon style.theme` va dritto al selettore del tema, `lunaos menu toggle system` apre il menu di sistema e lo chiude di nuovo se è già aperto, e `lunaos menu close` lo mette via.
