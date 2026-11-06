# LunaOS CLI

LunaOS wird normalerweise über die Tastenkürzel und das LunaOS-Menü (`Super + Space`) gesteuert. Aber du kannst es auch über die `lunaos`-CLI steuern. Das ist besonders hilfreich, wenn ein KI-Agent mit dir an Anpassung oder Konfiguration arbeitet.

Die CLI hat Zugriff auf alle internen Werkzeuge, die sowohl über das Menü als auch anderweitig verwendet werden. Du kannst alles sehen, was verfügbar ist, indem du `lunaos` im Terminal ausführst.

Es sieht ungefähr so aus:

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

Und du kannst bei jeder Gruppe tiefer eintauchen:

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

Jeder Befehl akzeptiert auch `--help`, egal ob du eine ganze Gruppe fragst (`lunaos capture --help`) oder einen einzelnen Befehl (`lunaos capture screenshot --help`).

### Das Menü vom Terminal aus öffnen

Das LunaOS-Menü ist ebenfalls skriptbar, was für deine eigenen Tastenzuordnungen praktisch ist. `lunaos menu` öffnet es an der Wurzel, und du kannst direkt zu jedem Punkt im Baum springen, indem du ihn benennst: `lunaos menu summon style.theme` geht direkt zur Design-Auswahl, `lunaos menu toggle system` öffnet das Systemmenü und schließt es wieder, wenn es bereits offen ist, und `lunaos menu close` räumt es weg.
