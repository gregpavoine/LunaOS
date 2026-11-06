# LunaOS CLI

LunaOS genellikle kısayollar ve LunaOS menüsü (`Super + Space`) yoluyla denetlenir. Ancak onu `lunaos` CLI yoluyla da denetleyebilirsiniz. Bu, özellikle bir AI ajanını özelleştirme veya yapılandırma konusunda sizinle çalıştırırken yararlıdır.

CLI, hem menü yoluyla hem de başka şekilde kullanılan tüm dahili araçlara erişime sahiptir. Terminalde `lunaos` çalıştırarak kullanılabilir her şeyi görebilirsiniz.

Şöyle bir şeye benzer:

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

Ve her gruba daha derine inebilirsiniz:

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

Her komut `--help` de alır; ister tüm bir grubu sorun (`lunaos capture --help`) ister tek bir komutu sorun (`lunaos capture screenshot --help`).

### Menüyü terminalden açmak

LunaOS menüsü de betiklenebilir; bu kendi tuş bağlamalarınız için kullanışlıdır. `lunaos menu` onu kökte açar ve adını vererek ağaçtaki herhangi bir noktaya doğrudan atlayabilirsiniz: `lunaos menu summon style.theme` doğrudan tema seçiciye gider, `lunaos menu toggle system` sistem menüsünü açar ve zaten açıksa tekrar kapatır ve `lunaos menu close` onu kaldırır.
