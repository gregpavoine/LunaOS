# LunaOS CLI

LunaOS は通常、ホットキーと LunaOS メニュー（`Super + Space`）から制御します。しかし `lunaos` CLI からも制御できます。これは、AI エージェントにカスタマイズや設定を一緒にやってもらうときに特に役立ちます。

CLI は、メニュー経由やそれ以外で使われるすべての内部ツールにアクセスできます。ターミナルで `lunaos` を実行すると、利用可能なもののすべてが見えます。

こんな感じです：

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

そしてすべてのグループをより深く掘り下げることもできます：

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

すべてのコマンドは `--help` も受け付けます。グループ全体でも（`lunaos capture --help`）、単一のコマンドでも（`lunaos capture screenshot --help`）です。

### ターミナルからメニューを開く

LunaOS メニューはスクリプト化もでき、自分のキーバインドに便利です。`lunaos menu` はルートで開き、名前を指定するとツリーの任意のポイントへ直接ジャンプできます：`lunaos menu summon style.theme` はテーマピッカーへ直接行き、`lunaos menu toggle system` はシステムメニューを開き、すでに開いていれば再び閉じ、`lunaos menu close` は片付けます。
