# LunaOS CLI

LunaOS thường được điều khiển qua các phím tắt và menu LunaOS (`Super + Space`). Nhưng bạn cũng có thể điều khiển nó qua CLI `lunaos`. Điều này đặc biệt hữu ích khi bạn có một AI agent làm việc cùng bạn về tùy chỉnh hoặc cấu hình.

CLI có quyền truy cập vào mọi công cụ nội bộ được dùng cả qua menu và bằng cách khác. Bạn có thể xem mọi thứ khả dụng bằng cách chạy `lunaos` trong terminal.

Nó trông giống như thế này:

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

Và bạn có thể đi sâu hơn vào mọi nhóm:

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

Mọi lệnh cũng nhận `--help`, dù bạn hỏi một nhóm toàn bộ (`lunaos capture --help`) hay một lệnh duy nhất (`lunaos capture screenshot --help`).

### Mở menu từ terminal

Menu LunaOS cũng có thể viết tập lệnh được, điều này tiện dụng cho các phím gán của riêng bạn. `lunaos menu` mở nó ở gốc, và bạn có thể nhảy thẳng tới bất kỳ điểm nào trong cây bằng cách đặt tên cho nó: `lunaos menu summon style.theme` đi thẳng tới bộ chọn chủ đề, `lunaos menu toggle system` mở menu hệ thống và đóng nó lại nếu nó đã hiển thị, và `lunaos menu close` cất nó đi.
