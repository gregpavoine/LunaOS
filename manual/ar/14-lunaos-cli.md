# LunaOS CLI

يُتحكّم في LunaOS عادة عبر اختصارات المفاتيح وقائمة LunaOS (`Super + Space`). لكن يمكنك أيضًا التحكّم فيه عبر `lunaos` CLI. هذا مفيد بشكل خاص عندما يكون لديك وكيل AI يعمل معك على تخصيص أو تكوين.

لـ CLI وصول إلى كل الأدوات الداخلية المستخدمة عبر القائمة وبخلاف ذلك. يمكنك رؤية كل شيء متاح بتشغيل `lunaos` في الطرفية.

يبدو شيئًا كهذا:

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

ويمكنك الغوص أعمق في كل مجموعة:

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

كل أمر يأخذ `--help` أيضًا، سواء سألت مجموعة كاملة (`lunaos capture --help`) أو أمرًا واحدًا (`lunaos capture screenshot --help`).

### فتح القائمة من الطرفية

قائمة LunaOS قابلة للبرمجة النصية أيضًا، وهو ما هو مفيد لروابط مفاتيحك الخاصة. `lunaos menu` يفتحها في الجذر، ويمكنك القفز مباشرة إلى أي نقطة في الشجرة بتسميتها: `lunaos menu summon style.theme` يذهب مباشرة إلى منتقي السمات، و`lunaos menu toggle system` يفتح قائمة النظام ويغلقها مرة أخرى إذا كانت مفتوحة بالفعل، و`lunaos menu close` يضعها بعيدًا.
