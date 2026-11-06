# LunaOS CLI

LunaOS er venjulega stjórnað með flýtilyklunum og LunaOS-valmyndinni (`Super + Space`). En þú getur líka stjórnað því í gegnum `lunaos` CLI. Þetta er sérstaklega gagnlegt þegar þú lætur gervigreindar-agent vinna með þér að aðlögun eða stillingum.

CLI-ið hefur aðgang að öllum innri verkfærunum sem eru notuð bæði í gegnum valmyndina og annars staðar. Þú getur séð allt sem er í boði með því að keyra `lunaos` í terminalnum.

Það lítur svona út:

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

Og þú getur kafað dýpra í hvern hóp:

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

Sérhver skipun tekur líka við `--help`, hvort sem þú biður allan hópinn (`lunaos capture --help`) eða eina skipun (`lunaos capture screenshot --help`).

### Að opna valmyndina úr terminalnum

LunaOS-valmyndin er einnig forskriftahæf, sem er þægilegt fyrir þínar eigin lyklatengingar. `lunaos menu` opnar hana við rótina, og þú getur hoppað beint á hvaða punkt sem er í trénu með því að nefna hann: `lunaos menu summon style.theme` fer beint í þemavalinn, `lunaos menu toggle system` opnar kerfisvalmyndina og lokar henni aftur ef hún er þegar uppi, og `lunaos menu close` setur hana í burtu.
