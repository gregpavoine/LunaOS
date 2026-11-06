# Atnaujinimai

LunaOS ir tavo paketai išlaikomi naujausi per _Atnaujinti > LunaOS_ LunaOS meniu (`Super + Space`).

Pati LunaOS įdiegiama kaip įprastiniai pacman paketai iš [LunaOS Package Repository](https://github.com/omacom-io/omarchy-pkgs), tad atnaujinimas įdiegia [naujausią LunaOS leidimą](https://github.com/gregpavoine/LunaOS/releases), paleidžia visas laukiančias migracijas, kad tavo sistema būtų sinchronizuota su naujausia, ir atnaujina visus sistemos paketus iš [LunaOS Arch Mirror](https://github.com/omacom-io/omarchy-mirror) ir [AUR](https://aur.archlinux.org/) (jei įdiegei kokius nors AUR paketus).

Kai išleidžiami nauji leidimai, dešinėje nuo tavo laikrodžio pasirodys apskritimo rodyklės piktograma. Spustelėk ją ir atnaujinimo procesas prasidės.

![update-available](images/update-available.webp)

### Keturi kanalai

LunaOS atnaujinama pagal keturis kanalus: stable, RC, edge ir dev. Nauji diegimai prasideda stable kanale, kuris seka [oficialius leidimus](https://github.com/gregpavoine/LunaOS/releases/), taip pat [stable LunaOS Arch mirror](https://github.com/omacom-io/omarchy-mirror), kuris atsilieka mėnesiu nuo naujausio, kad galėtume pagauti visus naujus nesuderinamumus, reikalaujančius konfigūracijos pakeitimų, kol jie nesukėlė problemų žmonėms.

Bet jei norėtum padėti pastebėti tas galimas problemas, gali veikti edge kanale. Tai leis tavo LunaOS paketams sekti naujausias kūrimo versijas ir leis atnaujinti iki naujausių Arch paketų, kai tik jie prieinami. Tai turėtum daryti tik jei esi patyręs su Linux ir žinai, kaip atkurti sistemą, kuri turi problemų.

Prieš kiekvieną naują didelį leidimą, galutinį patvirtinimą atliksime naudodami RC kanalą. Jei domiesi pagalba su galutiniu blizginimu, apsilankyk #lunaos-release-candidates Discord'e.

Galiausiai yra dev kanalas, kuris susieja LunaOS tiesiogiai su šaltinio kodo git kopija `~/lunaos`, kartu su edge paketais. Šį kanalą turėtum naudoti tik jei esi patyręs Linux naudotojas, dirbantis tiesiai su LunaOS, ir pasiruošęs toleruoti lūžimus.

Tarp kanalų gali perjungti naudodamas _Atnaujinti > Channel_ iš LunaOS meniu (arba `lunaos-channel-set` terminale).

### Programinės aparatinės įrangos (firmware) atnaujinimai

Tavo paketai nėra vienintelis dalykas, kuris pasensta. Daugelis nešiojamųjų kompiuterių ir periferinių įrenginių BIOS, SSD ir dock programinę aparatinę įrangą pristato per Linux Vendor Firmware Service, o _Atnaujinti > Firmware_ LunaOS meniu paims ir įdiegs visa, kas laukia tavo aparatūrai. Jis įdiegia `fwupd` pirmą kartą jį paleidus. Daugelį programinės aparatinės įrangos galima įrašyti tik perkrovimo metu, tad nenustebk, jei paprašys.

### Įspėjimas dėl tiesioginių pacman/yay atnaujinimų

Jei jau esi susipažinęs su Arch, gali kilti pagunda tiesiog pačiam paleisti `pacman -Syu` arba `yay -Syu`, bet jei tai padarysi, praleisi momentinę kopiją, migracijas ir konfigūracijos atnaujinimus, kuriuos LunaOS paleidžia kartu su naujais paketais. Todėl LunaOS iš tikrųjų sustabdys tiesioginį sistemos atnaujinimą ir nukreips tave į `lunaos update`. (Jei tikrai žinai, ką darai, apsauga pasakys, kaip ją apeiti vienai operacijai.)

### Blogų atnaujinimų atšaukimas

Jei po atnaujinimo kils problemų, gali atšaukti savo sistemą į momentinę kopiją, paimtą prieš atnaujinimą. Tiesiog paleisk iš naujo ir paleidimo meniu pasirink momentinę kopiją iš laikų prieš pradedant atnaujinimą.

![bootloader](images/bootloader.webp)

Jei kaip nors tavo konfigūracijos failai buvo sugadinti, taip pat gali atlikti LunaOS pakartotinį diegimą naudodamas `lunaos reinstall` terminale. Tai iš naujo įdiegs visus numatytuosius LunaOS paketus, nustatys tave į stable ir sugrąžins paketus, kurie per nauji, bei atkurs visus konfigūracijos failus. Atmink, kad atlikus tai visi tavo naudotojo konfigūracijos pakeitimai LunaOS numatytiesiems bus perrašyti!
