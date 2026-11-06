# Aktualizácie

LunaOS a tvoje balíčky sa udržiavajú aktuálne prostredníctvom časti _Aktualizovať > LunaOS_ v menu LunaOS (`Super + Space`).

Samotný LunaOS je nainštalovaný ako bežné balíčky pacman z [LunaOS Package Repository](https://github.com/omacom-io/omarchy-pkgs), takže aktualizácia nainštaluje [najnovšie vydanie LunaOS](https://github.com/gregpavoine/LunaOS/releases), spustí všetky čakajúce migrácie, aby bol tvoj systém zosynchronizovaný s najnovším vydaním, a aktualizuje všetky systémové balíčky z [LunaOS Arch Mirror](https://github.com/omacom-io/omarchy-mirror) a [AUR](https://aur.archlinux.org/) (ak máš nainštalované nejaké balíčky AUR).

Keď vyjdú nové vydania, napravo od tvojich hodín sa objaví ikona kruhovej šípky. Klikni na ňu a proces aktualizácie sa začne.

![update-available](images/update-available.webp)

### Štyri kanály

LunaOS sa aktualizuje po štyroch kanáloch: stabilný (stable), RC, edge a dev. Nové inštalácie začínajú na stabilnom kanáli, ktorý sleduje [oficiálne vydania](https://github.com/gregpavoine/LunaOS/releases/), ako aj [stabilné zrkadlo LunaOS Arch](https://github.com/omacom-io/omarchy-mirror), ktoré beží jeden mesiac pozadu za najnovším, aby sme mohli zachytiť všetky nové nekompatibility, ktoré si vyžadujú zmeny konfigurácie, skôr než spôsobia ľuďom problémy.

Ak by si však chcel pomôcť spozorovať tieto potenciálne problémy, môžeš bežať na kanáli edge. Tým tvoje balíčky LunaOS udržia krok s najnovšími vývojovými zostavami a umožnia ti aktualizovať na najnovšie balíčky Archu hneď, ako budú dostupné. Mal by si to robiť len vtedy, ak máš skúsenosti s Linuxom a vieš, ako obnoviť systém, ktorý má problémy.

Pred každým novým hlavným vydaním budeme robiť finálnu validáciu pomocou kanála RC. Ak máš záujem pomôcť s finálnym leštením, príď sa porozprávať do #lunaos-release-candidates na Discorde.

Nakoniec je tu kanál dev, ktorý prepája LunaOS priamo s git checkoutom zdrojového kódu v `~/lunaos` v kombinácii s balíčkami edge. Tento kanál by si mal používať len vtedy, ak si skúsený používateľ Linuxu, pracuješ priamo na LunaOS a si ochotný tolerovať poruchy.

Medzi kanálmi môžeš prepínať pomocou _Aktualizovať > Kanál_ z menu LunaOS (alebo `lunaos-channel-set` v termináli).

### Aktualizácie firmvéru

Nielen tvoje balíčky zostarnú. Mnoho notebookov a periférií dodáva BIOS, SSD a dokovací firmvér prostredníctvom Linux Vendor Firmware Service a časť _Aktualizovať > Firmvér_ v menu LunaOS načíta a nainštaluje všetko, čo tvoj hardvér čaká. Pri prvom spustení nainštaluje `fwupd`. Veľa firmvéru sa dá zapísať len počas reštartu, takže sa nečuduj, ak ťa oň požiada.

### Upozornenie na priame aktualizácie pomocou pacman/yay

Ak už poznáš Arch, môžeš byť v pokušení jednoducho spustiť `pacman -Syu` alebo `yay -Syu` sám, ale ak to urobíš, zmeškáš snímku, migrácie a aktualizácie konfigurácie, ktoré LunaOS spúšťa spolu s novými balíčkami. To je dôvod, prečo LunaOS priamu aktualizáciu systému skutočne zastaví a namiesto toho ťa nasmeruje na `lunaos update`. (Ak naozaj vieš, čo robíš, stráž ti povie, ako ho obísť pre jednu transakciu.)

### Vrátenie zlých aktualizácií

Ak budeš mať niekedy problém po vykonaní aktualizácie, môžeš svoj systém vrátiť späť na snímku zhotovenú pred aktualizáciou. Stačí reštartovať a v menu zavádzania pri štarte vybrať snímku spred začatia aktualizácie.

![bootloader](images/bootloader.webp)

Ak sa nejakým spôsobom poškodia tvoje konfiguračné súbory, môžeš tiež vykonať preinštalovanie LunaOS pomocou `lunaos reinstall` v termináli. Tým sa preinštalujú všetky predvolené balíčky LunaOS, dostaneš sa na stabilný kanál a znížia sa verzie všetkých balíčkov, ktoré sú príliš nové, a obnovia sa všetky konfiguračné súbory. Maj na pamäti, že tým sa prepíšu všetky tvoje používateľské zmeny konfigurácie oproti predvoleným nastaveniam LunaOS!
