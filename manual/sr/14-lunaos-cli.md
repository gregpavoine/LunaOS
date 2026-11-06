# LunaOS CLI

LunaOS се обично контролише преко пречица и LunaOS менија (`Super + Space`). Али можете га контролисати и преко `lunaos` CLI-ја. Ово је посебно корисно кад AI агент ради с вама на прилагођавању или конфигурацији.

CLI има приступ свим интерним алатима који се користе и преко менија и иначе. Све што је доступно можете видети покретањем `lunaos` у терминалу.

Изгледа отприлике овако:

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

И можете заронити дубље у сваку групу:

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

Свака команда такође прихвата `--help`, било да питате целу групу (`lunaos capture --help`) или једну команду (`lunaos capture screenshot --help`).

### Отварање менија из терминала

LunaOS мени је такође скриптабилан, што је згодно за ваше сопствене пречице тастатуре. `lunaos menu` га отвара у корену, а можете скочити право на било коју тачку у стаблу именовањем: `lunaos menu summon style.theme` иде право на бирач тема, `lunaos menu toggle system` отвара системски мени и поново га затвара ако је већ отворен, а `lunaos menu close` га склања.
