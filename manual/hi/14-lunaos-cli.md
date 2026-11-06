# LunaOS CLI

LunaOS को आमतौर पर हॉटकी और LunaOS मेनू (`Super + Space`) के ज़रिए नियंत्रित किया जाता है। लेकिन आप इसे `lunaos` CLI के ज़रिए भी नियंत्रित कर सकते हैं। यह विशेष रूप से तब मददगार होता है जब कोई AI एजेंट अनुकूलन या कॉन्फ़िगरेशन पर आपके साथ काम कर रहा हो।

CLI के पास सभी आंतरिक उपकरणों तक पहुँच है जिनका इस्तेमाल मेनू के ज़रिए भी किया जाता है और अन्यथा भी। आप टर्मिनल में `lunaos` चलाकर उपलब्ध सभी चीज़ों को देख सकते हैं।

यह कुछ इस तरह दिखता है:

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

और आप हर समूह में और गहराई से जा सकते हैं:

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

हर कमांड `--help` भी लेता है, चाहे आप पूरा समूह पूछें (`lunaos capture --help`) या कोई एक कमांड (`lunaos capture screenshot --help`)।

### टर्मिनल से मेनू खोलना

LunaOS मेनू भी स्क्रिप्ट करने योग्य है, जो आपकी अपनी कीबाइंडिंग के लिए उपयोगी है। `lunaos menu` उसे जड़ पर खोलता है, और आप उसका नाम लेकर वृक्ष के किसी भी बिंदु पर सीधे जा सकते हैं: `lunaos menu summon style.theme` सीधे थीम पिकर पर जाता है, `lunaos menu toggle system` सिस्टम मेनू खोलता है और अगर वह पहले से खुला है तो उसे फिर से बंद कर देता है, और `lunaos menu close` उसे हटा देता है।
