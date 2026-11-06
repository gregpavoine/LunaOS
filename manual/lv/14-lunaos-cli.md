# LunaOS CLI

LunaOS parasti kontrolē, izmantojot karstās taustiņu kombinācijas un LunaOS izvēlni (`Super + Space`). Bet tu vari to kontrolēt arī, izmantojot `lunaos` CLI. Tas ir īpaši noderīgi, kad tev ir AI aģents, kas strādā ar tevi pie pielāgošanas vai konfigurēšanas.

CLI ir piekļuve visiem iekšējiem rīkiem, ko izmanto gan caur izvēlni, gan citādi. Tu vari redzēt visu, kas ir pieejams, palaižot `lunaos` terminālī.

Tas izskatās apmēram šādi:

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

Un tu vari iedziļināties katrā grupā:

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

Katru komandu var izsaukt arī ar `--help`, neatkarīgi no tā, vai tu jautā visai grupai (`lunaos capture --help`) vai vienai komandai (`lunaos capture screenshot --help`).

### Izvēlnes atvēršana no termināla

LunaOS izvēlne ir arī skriptējama, kas ir ērti tavām taustiņu piesaistēm. `lunaos menu` atver to saknē, un tu vari lekt tieši uz jebkuru punktu kokā, nosaucot to: `lunaos menu summon style.theme` dodas tieši uz tēmu atlasītāju, `lunaos menu toggle system` atver sistēmas izvēlni un atkal to aizver, ja tā jau ir augšā, un `lunaos menu close` to novāc.
