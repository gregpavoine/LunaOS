# Frissítések

Az LunaOS és a csomagjaid naprakészek maradnak az LunaOS menü (`Super + Space`) _Update > LunaOS_ pontján keresztül.

Maga az LunaOS szokásos pacman-csomagokként települ az [LunaOS Package Repositoryból](https://github.com/omacom-io/omarchy-pkgs), így egy frissítés telepíti [a legújabb LunaOS-kiadást](https://github.com/gregpavoine/LunaOS/releases), lefuttat minden függőben lévő migrációt, hogy a rendszered szinkronba kerüljön a legújabbal, és frissíti az összes rendszercsomagot az [LunaOS Arch Mirrorról](https://github.com/omacom-io/omarchy-mirror) és az [AUR-ból](https://aur.archlinux.org/) (ha telepítettél AUR-csomagokat).

Amikor új kiadások jelennek meg, egy környíl ikon jelenik meg az órádtól jobbra. Kattints rá, és elindul a frissítési folyamat.

![update-available](images/update-available.webp)

### Négy csatorna

Az LunaOS négy csatornán frissül: stable, RC, edge és dev. Az új telepítések a stable csatornán indulnak, amely a [hivatalos kiadásokat](https://github.com/gregpavoine/LunaOS/releases/) követi, valamint a [stable LunaOS Arch mirrort](https://github.com/omacom-io/omarchy-mirror), amely egy hónappal lemarad a legújabbról, hogy kiszúrjunk minden új inkompatibilitást, amely konfigurációs módosítást igényel, mielőtt problémákat okozna az embereknek.

De ha segíteni szeretnél ezeknek a lehetséges problémáknak a megtalálásában, futtathatod az edge csatornát. Ez az LunaOS-csomagoknak a legújabb fejlesztői buildeket követve tartja, és lehetővé teszi, hogy a legújabb Arch csomagokra frissíts, amint elérhetők. Csak akkor tedd ezt, ha tapasztalt Linux-használó vagy, és tudod, hogyan állíts helyre egy problémákkal küzdő rendszert.

Minden új nagyobb kiadás előtt végső validációt végzünk az RC csatornán. Ha érdekel a végső simításokban való segítség, gyere és lógj az #lunaos-release-candidates csatornán a Discordon.

Végül ott van a dev csatorna, amely közvetlenül egy, a forráskód `~/lunaos` alatti git-checkoutjához kapcsolja az LunaOSt, kombinálva az edge csomagokkal. Csak akkor használd ezt a csatornát, ha tapasztalt Linux-használó vagy, közvetlenül az LunaOSn dolgozol, és hajlandó vagy elviselni a töréseket.

A csatornák között az LunaOS menü _Update > Channel_ pontjával válthatsz (vagy terminálban az `lunaos-channel-set` paranccsal).

### Firmware-frissítések

Nem csak a csomagjaid avulnak el. Sok laptop és periféria BIOS-, SSD- és dokkolófirmware-t szállít a Linux Vendor Firmware Service-en keresztül, és az LunaOS menü _Update > Firmware_ pontja lekéri és telepíti, ami a hardveredre vár. Az első futtatáskor telepíti a `fwupd`-t. Rengeteg firmware csak újraindítás során írható, úgyhogy ne lepődj meg, ha arra kérnek.

### Figyelmeztetés a közvetlen pacman/yay frissítésekről

Ha már ismered az Archot, kísértés lehet, hogy magad futtass `pacman -Syu` vagy `yay -Syu` parancsot, de ha ezt teszed, hiányozni fog a pillanatkép, a migrációk és a konfigurációs frissítések, amelyeket az LunaOS az új csomagokkal együtt futtat. Ezért az LunaOS ténylegesen leállít egy közvetlen rendszerfrissítést, és az `lunaos update` parancsra mutat. (Ha tényleg tudod, mit csinálsz, az őrszem megmondja, hogyan kerülheted meg egyetlen tranzakció erejéig.)

### Rossz frissítések visszaállítása

Ha frissítés után valaha problémád adódik, visszaállíthatod a rendszeredet a frissítés előtt készült pillanatképre. Csak indíts újra, és válaszd ki a rendszerbetöltő menüben azt a pillanatképet, amely a frissítés indítása előtt készült.

![bootloader](images/bootloader.webp)

Ha a konfigurációs fájljaid valahogy megsérültek, LunaOS-újratelepítést is végezhetsz a terminálban az `lunaos reinstall` paranccsal. Ez újratelepíti az összes alapértelmezett LunaOS-csomagot, stable csatornára tesz, és visszaminősít minden túl új csomagot, valamint visszaállít minden konfigurációs fájlt. Ne feledd, hogy ezzel minden felhasználói konfigurációs módosításod az LunaOS alapértelmezéseihez felülíródik!
