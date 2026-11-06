# Visasisho

LunaOS na vifurushi vyako huhifadhiwa vikiwa vya kisasa kupitia _Sasisha > LunaOS_ katika menyu ya LunaOS (`Super + Space`).

LunaOS yenyewe husakinishwa kama vifurushi vya kawaida vya pacman kutoka [Hifadhi ya Vifurushi ya LunaOS](https://github.com/omacom-io/omarchy-pkgs), hivyo sasisho husakinisha [toleo jipya zaidi la LunaOS](https://github.com/gregpavoine/LunaOS/releases), huendesha uhamaji wowote unaosubiri kuweka mfumo wako sawia na jipya zaidi, na husasisha vifurushi vyote vya mfumo kutoka [Kioo cha Arch cha LunaOS](https://github.com/omacom-io/omarchy-mirror) na [AUR](https://aur.archlinux.org/) (kama umesakinisha vifurushi vyovyote vya AUR).

Wakati matoleo mapya yanapofanywa, aikoni ya mshale wa duara itaonekana kulia kwa saa yako. Ibofye na mchakato wa kusasisha utaanza.

![update-available](images/update-available.webp)

### Vituo vinne

LunaOS husasishwa kupitia vituo vinne: stable, RC, edge, na dev. Usakinishaji mpya huanza kwenye kituo cha stable, ambacho hufuata [matoleo rasmi](https://github.com/gregpavoine/LunaOS/releases/), pamoja na [kioo thabiti cha Arch cha LunaOS](https://github.com/omacom-io/omarchy-mirror) kinachoendesha mwezi mmoja nyuma ya jipya zaidi, hivyo tunaweza kunasa kutokuingiliana kokote kupya kunakohitaji mabadiliko ya usanidi kabla hayajasababisha matatizo kwa watu.

Lakini kama ungependa kusaidia kuona masuala hayo yanayowezekana, unaweza kuendesha kwenye kituo cha edge. Hiyo itaweka vifurushi vyako vya LunaOS kufuata majenzi mapya zaidi ya uendelezaji, na inakuwezesha kusasisha hadi vifurushi vya hivi karibuni vya Arch mara vinapopatikana. Unapaswa kufanya hili tu kama una uzoefu na Linux, na unajua jinsi ya kurejesha mfumo wenye matatizo.

Kabla ya toleo lolote kubwa jipya, tutakuwa tunafanya uhakiki wa mwisho kwa kutumia kituo cha RC. Kama una nia ya kusaidia na ung'arishaji wa mwisho, njoo uhangae katika #lunaos-release-candidates kwenye Discord.

Mwishowe, kuna kituo cha dev, ambacho huunganisha LunaOS moja kwa moja na nakala ya git ya msimbo wa chanzo katika `~/lunaos`, pamoja na vifurushi vya edge. Unapaswa kutumia kituo hiki tu kama wewe ni mtumiaji mwenye uzoefu wa Linux, unafanya kazi moja kwa moja kwenye LunaOS, na uko tayari kuvumilia kuvunjika.

Unaweza kubadilisha kati ya vituo kwa kutumia _Sasisha > Kituo_ kutoka menyu ya LunaOS (au `lunaos-channel-set` katika terminali).

### Visasisho vya firmware

Vifurushi vyako si kitu pekee kinachochakaa. Kompyuta mpakato nyingi na vifaa vya pembeni husafirisha firmware ya BIOS, SSD, na doki kupitia Huduma ya Firmware ya Muuzaji wa Linux, na _Sasisha > Firmware_ katika menyu ya LunaOS itachukua na kusakinisha chochote ambacho vifaa vyako vinasubiri. Husakinisha `fwupd` mara ya kwanza unapoiendesha. Firmware nyingi zinaweza kuandikwa tu wakati wa kuwashwa upya, hivyo usishangae kuombwa kimoja.

### Onyo kuhusu visasisho vya moja kwa moja vya pacman/yay

Kama tayari unaifahamu Arch, unaweza kujaribiwa kuendesha tu `pacman -Syu` au `yay -Syu` mwenyewe, lakini kama utafanya hivyo, utakosa picha ya haraka, uhamaji, na visasisho vya usanidi ambavyo LunaOS huendesha pamoja na vifurushi vipya. Ndiyo maana LunaOS kwa kweli itasimamisha uboreshaji wa moja kwa moja wa mfumo na kukuelekeza kwenye `lunaos update` badala yake. (Kama kwa kweli unajua unachofanya, mlinzi atakuambia jinsi ya kuikwepa kwa muamala mmoja.)

### Kurudisha nyuma visasisho vibaya

Kama utawahi kuwa na tatizo baada ya kufanya sasisho, unaweza kurudisha mfumo wako nyuma hadi picha ya haraka iliyochukuliwa kabla ya sasisho. Anzisha upya tu na uchague picha ya haraka katika menyu ya upakiaji wa kuwasha kutoka kabla hujanza sasisho.

![bootloader](images/bootloader.webp)

Kama kwa namna fulani faili zako za usanidi zimeharibika, unaweza pia kufanya usakinishaji upya wa LunaOS kwa kutumia `lunaos reinstall` katika terminali. Hii itasakinisha upya vifurushi vyote chaguomsingi vya LunaOS, kukuweka kwenye stable na kushusha vifurushi vyovyote ambavyo ni vipya sana, na kuweka upya faili zote za usanidi. Kumbuka kwamba mabadiliko yako yote ya usanidi wa mtumiaji kwa chaguomsingi za LunaOS yataandikwa juu kwa kufanya hivi!
