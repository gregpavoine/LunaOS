# LunaOS CLI

LunaOS සාමාන්‍යයෙන් පාලනය වන්නේ උණුසුම් යතුරු සහ LunaOS මෙනුව (`Super + Space`) හරහායි. නමුත් ඔබට එය `lunaos` CLI හරහා ද පාලනය කළ හැකිය. ඔබ අභිරුචිකරණය හෝ වින්‍යාසය පිළිබඳව AI නියෝජිතයෙකු ඔබ සමඟ වැඩ කරන විට මෙය විශේෂයෙන් ප්‍රයෝජනවත් වේ.

CLI එකට මෙනුව හරහා සහ වෙනත් ආකාරයකින් භාවිතා කරන සියලුම අභ්‍යන්තර මෙවලම් වෙත ප්‍රවේශය ඇත. ටර්මිනලයේ `lunaos` ධාවනය කිරීමෙන් ලබා ගත හැකි සියල්ල ඔබට දැකිය හැකිය.

එය මේ වගේ දෙයක් පෙනේ:

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

ඔබට සෑම කණ්ඩායමක් පිළිබඳවම ගැඹුරට කිමිදිය හැකිය:

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

ඔබ මුළු කණ්ඩායමක් (`lunaos capture --help`) හෝ තනි විධානයක් (`lunaos capture screenshot --help`) ඇසුවත් සෑම විධානයක්ම `--help` ද ගනී.

### ටර්මිනලයෙන් මෙනුව විවෘත කිරීම

LunaOS මෙනුව ස්ක්‍රිප්ට් කළ හැකි අතර, එය ඔබේම යතුරු බන්ධන සඳහා ප්‍රයෝජනවත් වේ. `lunaos menu` එය මූලයේ විවෘත කරන අතර, ඔබට එය නම් කිරීමෙන් ගසේ ඕනෑම ස්ථානයකට කෙලින්ම පැනිය හැකිය: `lunaos menu summon style.theme` කෙලින්ම තේමා තේරීමට යයි, `lunaos menu toggle system` පද්ධති මෙනුව විවෘත කරන අතර එය දැනටමත් ඉහළට පැමිණ ඇත්නම් එය නැවත වසයි, සහ `lunaos menu close` එය ඉවත් කරයි.
