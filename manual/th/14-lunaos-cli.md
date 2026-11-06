# LunaOS CLI

โดยปกติ LunaOS ควบคุมผ่านปุ่มลัดและเมนู LunaOS (`Super + Space`) แต่คุณยังควบคุมมันผ่าน CLI `lunaos` ได้เช่นกัน สิ่งนี้ช่วยได้เป็นพิเศษเมื่อคุณให้เอเจนต์ AI ทำงานร่วมกับคุณในการปรับแต่งหรือการกำหนดค่า

CLI เข้าถึงเครื่องมือภายในทั้งหมดที่ใช้ทั้งผ่านเมนูและวิธีอื่น ๆ ได้ คุณสามารถดูทุกสิ่งที่มีได้โดยการรัน `lunaos` ในเทอร์มินัล

มันมีหน้าตาอะไรแบบนี้:

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

และคุณลงลึกในทุกกลุ่มได้:

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

ทุกคำสั่งรับ `--help` ด้วย ไม่ว่าคุณจะถามทั้งกลุ่ม (`lunaos capture --help`) หรือคำสั่งเดียว (`lunaos capture screenshot --help`)

### การเปิดเมนูจากเทอร์มินัล

เมนู LunaOS ยังเขียนสคริปต์ได้เช่นกัน ซึ่งมีประโยชน์สำหรับการผูกปุ่มของคุณเอง `lunaos menu` เปิดมันที่ราก และคุณกระโดดตรงไปยังจุดใด ๆ ในโครงสร้างโดยการตั้งชื่อมัน: `lunaos menu summon style.theme` ตรงไปยังตัวเลือกธีม `lunaos menu toggle system` เปิดเมนูระบบและปิดอีกครั้งหากมันเปิดอยู่แล้ว และ `lunaos menu close` เก็บมันไว้
