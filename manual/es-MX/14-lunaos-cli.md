# CLI de LunaOS

LunaOS normalmente se controla con los atajos de teclado y el menú de LunaOS (`Super + Space`). Pero también puedes controlarlo con la CLI `lunaos`. Esto es especialmente útil cuando tienes a un agente de IA trabajando contigo en personalización o configuración.

La CLI tiene acceso a todas las herramientas internas que se usan tanto desde el menú como por otros medios. Puedes ver todo lo disponible ejecutando `lunaos` en la terminal.

Se ve más o menos así:

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

Y puedes profundizar en cada grupo:

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

Cada comando también acepta `--help`, ya sea que le preguntes a un grupo completo (`lunaos capture --help`) o a un solo comando (`lunaos capture screenshot --help`).

### Abrir el menú desde la terminal

El menú de LunaOS también se puede controlar con scripts, lo cual es práctico para tus propios atajos. `lunaos menu` lo abre en la raíz, y puedes saltar directo a cualquier punto del árbol nombrándolo: `lunaos menu summon style.theme` va directo al selector de temas, `lunaos menu toggle system` abre el menú del sistema y lo cierra si ya está abierto, y `lunaos menu close` lo esconde.
