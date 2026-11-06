# LunaOS CLI

De ghnáth rialaítear LunaOS trí na heochracha gasta agus roghchlár LunaOS (`Super + Space`). Ach is féidir leat é a rialú freisin trí CLI `lunaos`. Tá sé seo an-chabhrach go háirithe nuair a bhíonn gníomhaire AI ag obair leat ar shaincheapadh nó ar chumraíocht.

Tá rochtain ag an CLI ar an uirlisiú inmheánach ar fad a úsáidtear tríd an roghchlár agus ar shlí eile araon. Is féidir leat gach rud atá ar fáil a fheiceáil trí `lunaos` a rith sa teirminéal.

Tá cuma air rud éigin mar seo:

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

Agus is féidir leat dul níos doimhne ar gach grúpa:

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

Glacann gach ordú le `--help` freisin, cibé acu a fhiafraíonn tú de ghrúpa iomlán (`lunaos capture --help`) nó d'ordú aonair (`lunaos capture screenshot --help`).

### An roghchlár a oscailt ón teirminéal

Is féidir roghchlár LunaOS a scriptiú freisin, atá áisiúil do do cheangail eochracha féin. Osclaíonn `lunaos menu` ag an bhfréamh é, agus is féidir leat léim go díreach chuig aon phointe sa chrann trí é a ainmniú: téann `lunaos menu summon style.theme` go díreach chuig an roghnóir téama, osclaíonn `lunaos menu toggle system` roghchlár an chórais agus dúnann sé arís é má tá sé suas cheana, agus cuireann `lunaos menu close` ar shiúl é.
