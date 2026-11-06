# LunaOS CLI

LunaOS عام طور پر ہاٹ کیز اور LunaOS مینیُو (`Super + Space`) کے ذریعے کنٹرول کیا جاتا ہے۔ لیکن آپ اسے `lunaos` CLI کے ذریعے بھی کنٹرول کر سکتے ہیں۔ یہ خاص طور پر اس وقت مددگار ہوتا ہے جب آپ AI ایجنٹ کو کسٹمائزیشن یا کنفیگریشن پر اپنے ساتھ کام کرا رہے ہوں۔

CLI کو ان تمام داخلی ٹولز تک رسائی حاصل ہے جو مینیُو کے ذریعے بھی استعمال ہوتے ہیں اور دوسری جگہ بھی۔ آپ ٹرمینل میں `lunaos` چلا کر دستیاب سب کچھ دیکھ سکتے ہیں۔

یہ کچھ اس طرح لگتا ہے:

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

اور آپ ہر گروپ پر مزید گہرائی میں جا سکتے ہیں:

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

ہر کمانڈ `--help` بھی لیتی ہے، چاہے آپ پورا گروپ پوچھیں (`lunaos capture --help`) یا کوئی ایک کمانڈ (`lunaos capture screenshot --help`)۔

### ٹرمینل سے مینیُو کھولنا

LunaOS مینیُو بھی اسکرپٹ کے قابل ہے، جو آپ کی اپنی کی بائنڈنگز کے لیے سودمند ہے۔ `lunaos menu` اسے جڑ پر کھولتا ہے، اور آپ اس کا نام دے کر درخت میں کسی بھی مقام پر براہ راست جا سکتے ہیں: `lunaos menu summon style.theme` سیدھا تھیم پِکر پر جاتا ہے، `lunaos menu toggle system` سسٹم مینیُو کھولتا ہے اور اگر یہ پہلے سے اوپر ہے تو اسے دوبارہ بند کر دیتا ہے، اور `lunaos menu close` اسے ہٹا دیتا ہے۔
