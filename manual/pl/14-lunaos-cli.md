# CLI LunaOS

LunaOS zwykle steruje się przez skróty oraz menu LunaOS (`Super + Space`). Ale możesz też sterować nim przez CLI `lunaos`. Jest to szczególnie pomocne, gdy pracujesz z agentem AI nad dostosowywaniem lub konfiguracją.

CLI ma dostęp do wszystkich wewnętrznych narzędzi używanych zarówno z menu, jak i w inny sposób. Wszystko, co jest dostępne, zobaczysz, uruchamiając `lunaos` w terminalu.

Wygląda to mniej więcej tak:

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

A do każdej grupy możesz wejść głębiej:

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

Każde polecenie przyjmuje też `--help`, niezależnie od tego, czy zapytasz całą grupę (`lunaos capture --help`), czy pojedyncze polecenie (`lunaos capture screenshot --help`).

### Otwieranie menu z terminala

Menu LunaOS jest też skryptowalne, co jest przydatne do własnych przypisań klawiszy. `lunaos menu` otwiera je u nasady, a możesz wskoczyć wprost do dowolnego punktu drzewa, nazywając go: `lunaos menu summon style.theme` idzie od razu do wyboru motywu, `lunaos menu toggle system` otwiera menu systemowe i zamyka je ponownie, jeśli jest już otwarte, a `lunaos menu close` je chowa.
