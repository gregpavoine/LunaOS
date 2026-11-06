# CLI של LunaOS

LunaOS נשלטת בדרך כלל דרך קיצורי המקלדת ותפריט LunaOS (`Super + Space`). אבל אפשר גם לשלוט בה דרך ה-CLI של `lunaos`. זה שימושי במיוחד כאשר סוכן AI עובד איתכם על התאמה אישית או הגדרה.

ל-CLI יש גישה לכל הכלים הפנימיים שנמצאים בשימוש הן דרך התפריט והן בדרכים אחרות. אפשר לראות את כל מה שזמין על ידי הרצת `lunaos` בטרמינל.

זה נראה בערך כך:

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

ואפשר לצלול עמוק יותר בכל קבוצה:

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

כל פקודה מקבלת גם את `--help`, בין אם שואלים קבוצה שלמה (`lunaos capture --help`) או פקודה בודדת (`lunaos capture screenshot --help`).

### פתיחת התפריט מהטרמינל

תפריט LunaOS ניתן גם לתסריט, מה שנחמד לשיוכי המקשים שלכם. הפקודה `lunaos menu` פותחת אותו בשורש, ואפשר לקפוץ ישירות לכל נקודה בעץ על ידי ציון שמה: `lunaos menu summon style.theme` עובר ישר לבורר ערכות הנושא, `lunaos menu toggle system` פותח את תפריט המערכת וסוגר אותו שוב אם הוא כבר פתוח, ו-`lunaos menu close` מסלק אותו.
