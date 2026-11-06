# LunaOS CLI

LunaOS는 보통 단축키와 LunaOS 메뉴(`Super + Space`)를 통해 제어합니다. 하지만 `lunaos` CLI를 통해서도 제어할 수 있습니다. 특히 AI 에이전트와 함께 사용자 지정이나 구성을 작업할 때 유용합니다.

CLI는 메뉴 등을 통해 사용되는 모든 내부 도구에 접근할 수 있습니다. 터미널에서 `lunaos`를 실행하면 사용 가능한 모든 것을 볼 수 있습니다.

모습은 다음과 비슷합니다:

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

그리고 모든 그룹에 대해 더 깊이 파고들 수 있습니다:

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

모든 명령은 `--help`도 지원합니다. 전체 그룹에 대해 물어보든(`lunaos capture --help`) 단일 명령에 대해 물어보든(`lunaos capture screenshot --help`) 마찬가지입니다.

### 터미널에서 메뉴 열기

LunaOS 메뉴도 스크립트화할 수 있어, 직접 키바인딩을 만들 때 유용합니다. `lunaos menu`는 루트에서 메뉴를 열고, 이름을 지정하면 트리의 어느 지점으로든 바로 이동할 수 있습니다. `lunaos menu summon style.theme`은 테마 선택기로 바로 가고, `lunaos menu toggle system`은 시스템 메뉴를 열었다가 이미 열려 있으면 다시 닫으며, `lunaos menu close`는 메뉴를 치웁니다.
