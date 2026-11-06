# CLI LunaOS

LunaOS se običajno nadzira prek bližnjic in menija LunaOS (`Super + Space`). Lahko pa ga nadzirate tudi prek CLI `lunaos`. To je še posebej koristno, kadar vam agent AI pomaga pri prilagajanju ali konfiguraciji.

CLI ima dostop do vsega notranjega orodjarjenja, ki se uporablja tako prek menija kot sicer. Vse, kar je na voljo, si lahko ogledate tako, da v terminalu zaženete `lunaos`.

Videti je nekako takole:

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

V vsako skupino se lahko poglobite:

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

Vsak ukaz sprejme tudi `--help`, ne glede na to, ali prosite za celotno skupino (`lunaos capture --help`) ali en sam ukaz (`lunaos capture screenshot --help`).

### Odpiranje menija iz terminala

Meni LunaOS je mogoče skriptati tudi, kar je priročno za lastne vezave tipk. `lunaos menu` ga odpre v korenu, skočite pa lahko naravnost na kateri koli točki drevesa, če ga poimenujete: `lunaos menu summon style.theme` gre naravnost do izbirnika teme, `lunaos menu toggle system` odpre sistemski meni in ga znova zapre, če je že odprt, `lunaos menu close` pa ga spravi stran.
