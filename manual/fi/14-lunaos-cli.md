# LunaOS CLI

LunaOSä ohjataan yleensä pikanäppäinten ja LunaOS-valikon kautta (`Super + Space`). Mutta voit ohjata sitä myös `lunaos` CLI:n kautta. Tämä on erityisen hyödyllistä, kun sinulla on tekoälyagentti työstämässä kanssasi mukautuksia tai määrityksiä.

CLI:llä on pääsy kaikkiin sisäisiin työkaluihin, joita käytetään sekä valikon kautta että muuten. Näet kaiken saatavilla olevan ajamalla pääätteessä komennon `lunaos`.

Se näyttää suunnilleen tältä:

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

Ja voit sukeltaa syvemmälle jokaiseen ryhmään:

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

Jokainen komento hyväksyy myös valitsimen `--help`, kysyitpä kokonaisen ryhmän (`lunaos capture --help`) tai yhden komennon (`lunaos capture screenshot --help`) ohjetta.

### Valikon avaaminen pääätteestä

Myös LunaOS-valikko on skriptattavissa, mikä on kätevää omia pikanäppäimiäsi varten. `lunaos menu` avaa sen juuresta, ja voit hypätä suoraan mihin tahansa puun kohtaan nimeämällä sen: `lunaos menu summon style.theme` menee suoraan teeman valitsimeen, `lunaos menu toggle system` avaa järjestelmävalikon ja sulkee sen uudelleen, jos se on jo avoinna, ja `lunaos menu close` sulkee sen kokonaan.
