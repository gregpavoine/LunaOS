# LunaOS CLI

LunaOS styrs vanligtvis via snabbtangenterna och LunaOS-menyn (`Super + Space`). Men du kan också styra det via `lunaos` CLI. Detta är särskilt hjälpsamt när du låter en AI-agent arbeta med dig på anpassning eller konfiguration.

CLI:t har åtkomst till alla interna verktyg som används både via menyn och annars. Du kan se allt som är tillgängligt genom att köra `lunaos` i terminalen.

Det ser ut ungefär så här:

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

Och du kan dyka djupare i varje grupp:

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

Varje kommando tar också `--help`, oavsett om du frågar en hel grupp (`lunaos capture --help`) eller ett enskilt kommando (`lunaos capture screenshot --help`).

### Öppna menyn från terminalen

LunaOS-menyn är också skriptbar, vilket är praktiskt för dina egna tangentbindningar. `lunaos menu` öppnar den vid roten, och du kan hoppa direkt till valfri punkt i trädet genom att namnge den: `lunaos menu summon style.theme` går rakt till temaväljaren, `lunaos menu toggle system` öppnar systemmenyn och stänger den igen om den redan är uppe, och `lunaos menu close` lägger undan den.
