# CLI do LunaOS

O LunaOS geralmente é controlado pelos atalhos e pelo menu do LunaOS (`Super + Space`). Mas você também pode controlá-lo pela CLI `lunaos`. Isso é particularmente útil quando você tem um agente de IA trabalhando com você em personalização ou configuração.

A CLI tem acesso a todas as ferramentas internas usadas tanto pelo menu quanto de outras formas. Você pode ver tudo o que está disponível executando `lunaos` no terminal.

Ela se parece com algo assim:

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

E você pode mergulhar mais fundo em cada grupo:

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

Todo comando também aceita `--help`, quer você pergunte a um grupo inteiro (`lunaos capture --help`) ou a um único comando (`lunaos capture screenshot --help`).

### Abrindo o menu pelo terminal

O menu do LunaOS também é programável por script, o que é útil para seus próprios atalhos. `lunaos menu` o abre na raiz, e você pode pular direto para qualquer ponto da árvore nomeando-o: `lunaos menu summon style.theme` vai direto ao seletor de temas, `lunaos menu toggle system` abre o menu do sistema e o fecha de novo se já estiver aberto, e `lunaos menu close` o guarda.
