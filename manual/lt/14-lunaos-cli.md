# LunaOS CLI

LunaOS paprastai valdoma per sparčiuosius klavišus ir LunaOS meniu (`Super + Space`). Bet taip pat galite ją valdyti per `lunaos` CLI. Tai ypač naudinga, kai AI agentas dirba su jumis prie pritaikymo ar konfigūracijos.

CLI turi prieigą prie visų vidinių įrankių, kurie naudojami tiek per meniu, tiek kitaip. Viską, kas prieinama, galite pamatyti terminale paleisdami `lunaos`.

Tai atrodo maždaug taip:

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

Ir galite gilintis į kiekvieną grupę:

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

Kiekviena komanda taip pat priima `--help`, nesvarbu, ar klausiate visos grupės (`lunaos capture --help`), ar vienos komandos (`lunaos capture screenshot --help`).

### Meniu atidarymas iš terminalo

LunaOS meniu taip pat galima užprogramuoti scenarijais, o tai patogu jūsų pačių klavišų susiejimams. `lunaos menu` atidaro jį šaknyje, ir galite šokti tiesiai į bet kurį medžio tašką jį įvardydami: `lunaos menu summon style.theme` eina tiesiai į temų parinkiklį, `lunaos menu toggle system` atidaro sistemos meniu ir vėl jį uždaro, jei jis jau atidarytas, o `lunaos menu close` jį padeda.
