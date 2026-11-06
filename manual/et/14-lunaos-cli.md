# LunaOS CLI

LunaOSt juhitakse tavaliselt kiirklahvide ja LunaOS menüü kaudu (`Super + Space`). Kuid saad seda juhtida ka `lunaos` CLI kaudu. See on eriti abiks, kui sul on AI agent, kes töötab sinuga kohandamise või seadistamise kallal.

CLI-l on juurdepääs kogu sisemisele tööriistakastile, mida kasutatakse nii menüü kui ka muul viisil. Saad näha kõike saadaolevat, käivitades terminalis `lunaos`.

See näeb välja umbes nii:

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

Ja saad igasse rühma sügavamale sukelduda:

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

Iga käsk võtab ka `--help`, olenemata sellest, kas küsid tervet rühma (`lunaos capture --help`) või ühte käsku (`lunaos capture screenshot --help`).

### Menüü avamine terminalist

LunaOS menüü on samuti skriptitav, mis on sinu enda klahviseoste jaoks käepärane. `lunaos menu` avab selle juure juures ja saad hüpata otse puu mis tahes punkti, nimetades selle: `lunaos menu summon style.theme` läheb otse teemavalija juurde, `lunaos menu toggle system` avab süsteemi menüü ja sulgeb selle uuesti, kui see juba üleval on, ja `lunaos menu close` paneb selle ära.
