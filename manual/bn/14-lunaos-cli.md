# LunaOS CLI

LunaOS সাধারণত হটকি ও LunaOS মেনু (`Super + Space`) দিয়ে নিয়ন্ত্রিত হয়। কিন্তু আপনি `lunaos` CLI দিয়েও এটি নিয়ন্ত্রণ করতে পারেন। কাস্টমাইজেশন বা কনফিগারেশনে কোনো AI এজেন্টকে আপনার সঙ্গে কাজ করানোর সময় এটি বিশেষভাবে সহায়ক।

CLI-এর মেনু দিয়ে ও অন্যভাবে ব্যবহৃত সমস্ত অভ্যন্তরীণ সরঞ্জামে অ্যাক্সেস রয়েছে। টার্মিনালে `lunaos` চালিয়ে উপলব্ধ সবকিছু দেখতে পারেন।

এটি দেখতে অনেকটা এমন:

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

আর আপনি প্রতিটি গোষ্ঠীতে আরও গভীরে যেতে পারেন:

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

প্রতিটি কমান্ড `--help`-ও নেয়, আপনি পুরো একটি গোষ্ঠীকে জিজ্ঞেস করুন (`lunaos capture --help`) বা একটি মাত্র কমান্ডকে (`lunaos capture screenshot --help`)।

### টার্মিনাল থেকে মেনু খোলা

LunaOS মেনু স্ক্রিপ্টযোগ্যও, যা আপনার নিজের কিবাইন্ডিংয়ের জন্য সুবিধাজনক। `lunaos menu` এটি মূলে খোলে, এবং নাম ধরে আপনি গাছের যেকোনো জায়গায় সরাসরি যেতে পারেন: `lunaos menu summon style.theme` সরাসরি থিম বাছাইকারে যায়, `lunaos menu toggle system` সিস্টেম মেনু খোলে এবং ইতিমধ্যে খোলা থাকলে আবার বন্ধ করে, এবং `lunaos menu close` এটি সরিয়ে দেয়।
