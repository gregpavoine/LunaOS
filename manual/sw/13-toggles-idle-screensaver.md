# Vibadili, Uvivu na Kiokozi cha Skrini

Mengi ya unayobadilisha siku hadi siku si mpangilio kweli. Ni hali unayowasha kwa saa moja na kuizima tena: taa ya usiku unapofanya kazi hadi kuchelewa, usinisumbue unapowasilisha, usilale unapotazama kitu. LunaOS huita hizo vibadili, na zote hufanya kazi kwa njia hiyo hiyo — mkato wa kitufe, ingizo la menyu, na amri, zote zikipiga swichi hiyo hiyo.

### Menyu ya kibadili

`Super + Ctrl + O` hufungua _Anzisha > Badilisha_ moja kwa moja, au unaweza kutembea huko kutoka menyu ya LunaOS (`Super + Space`). Kila kitu katika orodha hiyo ni swichi unayoweza kubadilisha bila kufikiria hali iko wapi. Swichi za jinsi eneo-kazi linavyoonekana hukaa kando yake, chini ya _Anzisha > Muonekano_.

Kutoka terminali, swichi hizo hizo ni `lunaos toggle <thing>`. Endesha `lunaos toggle` pekee kuona kundi zima.

| Kibadili | Mkato wa kitufe | Amri |
| ------ | ------ | ------- |
| Taa ya usiku | `Super + Ctrl + N` | `lunaos toggle nightlight` |
| Nyamazisha arifa | `Super + Ctrl + ,` | `lunaos toggle notification silencing` |
| Usilale (hakuna kufunga kwa uvivu) | `Super + Ctrl + I` | `lunaos toggle idle` |
| Nasa mkwamo | — | `lunaos toggle crash-capture` |
| Kiokozi cha skrini | — | `lunaos toggle screensaver` |
| Usawazishaji wa mandhari wa [Herdr](https://herdr.dev) | — | `lunaos toggle theme sync` |
| Upau wa menyu | `Super + Shift + Space` | `lunaos toggle bar` |
| Kiguso | `XF86TouchpadToggle` | `lunaos toggle touchpad` |
| Skrini ya kugusa | — | `lunaos toggle touchscreen` |
| Sitisha | — | `lunaos toggle suspend` |
| GPU mseto | — | `lunaos toggle hybrid gpu` |

Swichi za kiguso, skrini ya kugusa, na GPU mseto huishi chini ya _Anzisha > Vifaa_ (`Super + Ctrl + H`) badala ya chini ya Badilisha, kwa kuwa huonekana tu ukiwa kweli na vifaa hivyo. Za kiguso na skrini ya kugusa hudumu kupitia upakiaji upya wa Hyprland — jina la kifaa kilichozimwa huhifadhiwa kwenye faili ndogo ya hali ambayo Hyprland husoma wakati wa kuanza kukizima tena.

_Anzisha > Muonekano_ huhifadhi upau wa menyu na uhuishaji, pamoja na mambo machache ambayo si amri za `lunaos toggle` lakini hufanya kazi vivyo hivyo: asilimia ya betri katika upau, mpangilio chaguomsingi wa nafasi ya kazi, mapengo ya madirisha (`Super + Shift + Backspace`), na kipengele cha mraba cha dirisha 1 (`Super + Ctrl + Backspace`).

Nyingi za hizi ni faili tu ya bendera chini ya `~/.local/state/lunaos/toggles/`. Ikiwa unataka kutenda kulingana na moja katika hati, `lunaos-toggle-enabled` hukupa msimbo wa kutoka badala ya kukufanya uende kutafuta:

```bash
lunaos-toggle-enabled screensaver-off && echo "screensaver is off"
```

Bendera zimepewa majina kwa hali ya kuzimwa — `screensaver-off`, `suspend-off`, `bar-off` — hivyo uwepo wao una maana kipengele kimezimwa.

### Viashiria katika upau

Hali ikiwa imewashwa, unapata glifu ndogo katikati ya upau wa juu karibu na saa. Hiyo ndiyo wijeti ya viashiria, na hubeba imla, kurekodi skrini, vikumbusho vinavyongoja, taa ya usiku, usinisumbue, na usilale.

Viashiria visivyotumika vimefichwa. Elea eneo linalovizunguka na vinafifia kuonekana, hivyo unaweza kubofya kimoja kukiwashe bila kujua mkato wake wa kitufe. Kubofya kinachotumika hukizima tena. Ikiwa unapendelea kuviona vyote wakati wote, weka `alwaysShow` kuwa `true` kwenye ingizo la `lunaos.indicators` katika `~/.config/lunaos/shell.json` — tazama [upau wa juu](05-the-top-bar.md) kwa jinsi wijeti za upau zinavyosanidiwa.

### Taa ya usiku

`Super + Ctrl + N` huipa skrini joto hadi 4000K, na kuigonga tena hurejesha 6500K. Huendeshwa na hyprsunset, ambayo kibadili hukuanzishia ikiwa haiendeshwi tayari.

Kwa chaguomsingi hyprsunset haifanyi chochote kwenye skrini yako kabisa. `~/.config/hypr/hyprsunset.conf` hutumwa na wasifu wa utambulisho hasa ili skrini ibaki isiyoguswa hadi uombe joto. Ikiwa unapendelea ibadilike kwa saa, badilisha hiyo na wasifu wa muda:

```
profile {
    time = 20:00
    temperature = 4000
}
```

Kisha anza hyprsunset wakati wa kuingia kwa kuongeza `o.launch_on_start("hyprsunset")` kwenye `~/.config/hypr/autostart.lua`. Jozi ya 4000K/6500K inayotumiwa na kibadili ni thabiti, hivyo faili ya usanidi ndipo unapoenda ukitaka halijoto tofauti.

### Usinisumbue

`Super + Ctrl + ,` hunyamazisha arifa. Hakuna ibukia zinazoonekana wakati ikiweshwa, na kiashiria cha kengele iliyokatwa hukaa kwenye upau kukukumbusha kwa nini eneo-kazi limekuwa kimya.

Hakuna kinachopotea, hata hivyo. Arifa iliyonyamazishwa huandikwa moja kwa moja katika historia yako ya arifa, ambayo ndiyo rekodi hasa unayotaka unaporudi na kujiuliza ulichokosa. Ifungue kwa `Super + Shift + Alt + ,`. Tazama [arifa](10-notices.md) kwa hadithi iliyobaki ya arifa.

Aina mbili za ujumbe bado hupita: ibukia za uthibitisho za LunaOS yenyewe kwa kitu ulichofanya sasa hivi ("Theme changed", "Screenshot saved"), na tahadhari muhimu zinazotumwa kutoka mstari wa amri. Programu za mazungumzo zinazotia kila kitu muhimu kujilazimisha mbele yako hazistahili.

### Uvivu

Shell ya LunaOS humiliki tabia ya uvivu, na nyakati ni kizuizi cha kiwango cha juu `idle` katika `~/.config/lunaos/shell.json`:

```json
{
  "version": 1,
  "idle": {
    "screensaver": 150,
    "lock": 300
  }
}
```

Nambari zote mbili ni sekunde zinazohesabiwa tangu ulipokuwa mvivu — si kutoka kila moja. Kwa hiyo kwa chaguomsingi, kiokozi cha skrini huja baada ya dakika mbili na nusu na skrini ya kufunga huchukua katika dakika tano, iwe kiokozi cha skrini kiliendeshwa au la. Hifadhi faili na shell huchukua nyakati mpya mara moja.

Ikiwa unaondoa kiokozi cha skrini kabla ya tarehe ya mwisho ya kufunga, hiyo huhesabika kama shughuli na kufunga kunakongoja hughairiwa. Hufungiwi kwa kuchungulia mashine yako.

Kusimamisha kufunga kwa uvivu kabisa, `Super + Ctrl + I` — au `lunaos toggle idle` — hubadilisha usilale kuwaka, na kiashiria cha kikombe cha kahawa huonekana kwenye upau. Hiyo ndiyo ya kugonga kabla ya wasilisho refu au ujenzi unaotaka kutazama. Igonge tena kurudi kawaida. `lunaos toggle idle status` huchapisha hali ya sasa kama JSON ukiihitaji kutoka hati.

Hii ni kuhusu kufunga na kiokozi cha skrini, si nguvu. Sitisha na sinzia vina usanidi wao mwenyewe katika [usingizi wa mfumo](36-system-sleep.md).

### Kiokozi cha skrini

Kiokozi cha skrini cha LunaOS ni sanaa ya ASCII inayoendesha kupitia athari za maandishi za nasibu, mfano moja kwa kila skrini. Kitufe chochote au mwendo wa kipanya hutoka ndani yake.

Unaweza kukianzisha kwa mahitaji kutoka _Mfumo > Kiokozi cha skrini_ (`Super + Esc`), ambayo hukilazimisha hata kama umezima kiokozi cha skrini cha uvivu. Hakuna mkato wa kitufe uliofungwa kwake kwa chaguomsingi.

`lunaos toggle screensaver` ndicho kinachozima cha uvivu, ikiwa unapendelea kwenda moja kwa moja kutoka kufanya kazi hadi kufungwa. Kinahitaji terminali inayojua kusanidi — Alacritty, Foot, Ghostty, au Kitty — na kitakuambia hivyo ikiwa terminali yako chaguomsingi ni kitu kingine.

Nembo inayochora ni yako kubadilisha, chini ya _Mtindo > Kiokozi cha skrini_. Pakia png au svg na LunaOS huibadilisha kuwa ASCII. Tazama [utambulisho](41-branding.md).

### Skrini ya kufunga

`Super + Ctrl + L` hufunga mashine. Hiyo huendesha skrini ya kufunga kutoka shell ya LunaOS, huweka skrini tupu, huweka upya umbo lako la kibodi kuwa la kwanza ili usiandike nenosiri lako katika alfabeti isiyo sahihi, na — ikiwa unaiendesha — hufunga 1Password unapotoka.

Skrini ya kufunga huchukua nenosiri, na pia itachukua alama ya kidole ukishaiweka. Hiyo, na njia nyingine za kuthibitisha, zimefunikwa katika [uthibitishaji wa vifaa](37-hardware-authentication.md).
