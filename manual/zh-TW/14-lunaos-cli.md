# LunaOS CLI

LunaOS 通常透過快捷鍵和 LunaOS 選單（`Super + Space`）來控制。但你也可以透過 `lunaos` CLI 來控制它。當你讓 AI 代理程式和你一起處理自訂或設定時，這特別有幫助。

這個 CLI 可以存取所有透過選單以及其他方式使用的內部工具。在終端機中執行 `lunaos`，你就可以看到所有可用的東西。

它看起來像這樣：

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

而你可以深入每一個群組：

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

每一個指令也都接受 `--help`，無論你問的是一整個群組（`lunaos capture --help`）還是單一個指令（`lunaos capture screenshot --help`）。

### 從終端機開啟選單

LunaOS 選單也可以用腳本控制，這對你自己的快捷鍵很有用。`lunaos menu` 會在根目錄開啟它，而你可以透過命名直接跳到樹狀結構中的任何一點：`lunaos menu summon style.theme` 會直接前往主題選擇器，`lunaos menu toggle system` 會開啟系統選單，若它已經開啟則再次關閉，而 `lunaos menu close` 會把它收起來。
