# LunaOS CLI

LunaOS معمولاً از طریق کلیدهای میانبر و منوی LunaOS (`Super + Space`) کنترل می‌شود. اما همچنین می‌توانید آن را از طریق CLI ـی `lunaos` کنترل کنید. این به‌ویژه وقتی یک عامل هوش مصنوعی در سفارشی‌سازی یا پیکربندی با شما کار می‌کند مفید است.

CLI به همهٔ ابزارهای درونی که هم از طریق منو و هم در غیر این صورت استفاده می‌شوند دسترسی دارد. می‌توانید با اجرای `lunaos` در ترمینال همهٔ آنچه در دسترس است را ببینید.

تقریباً این‌گونه به نظر می‌رسد:

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

و می‌توانید در هر گروه عمیق‌تر شوید:

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

هر فرمان همچنین `--help` را می‌پذیرد، چه یک گروه کامل را بپرسید (`lunaos capture --help`) یا یک فرمان واحد (`lunaos capture screenshot --help`).

### باز کردن منو از ترمینال

منوی LunaOS همچنین قابل اسکریپت‌نویسی است، که برای نگاشت‌های کلیدی خودتان مفید است. `lunaos menu` آن را در ریشه باز می‌کند و می‌توانید با نام‌بردن مستقیماً به هر نقطه‌ای در درخت بروید: `lunaos menu summon style.theme` مستقیماً به انتخاب‌گر تم می‌رود، `lunaos menu toggle system` منوی سیستم را باز می‌کند و اگر از قبل باز باشد دوباره آن را می‌بندد و `lunaos menu close` آن را کنار می‌گذارد.
