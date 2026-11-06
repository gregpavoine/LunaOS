# Ažuriranja

LunaOS i tvoji paketi održavaju se ažurnima putem _Ažuriraj > LunaOS_ u izborniku LunaOS (`Super + Space`).

Sam LunaOS instaliran je kao redovni pacman paketi iz [LunaOS Package Repository](https://github.com/omacom-io/omarchy-pkgs), pa ažuriranje instalira [najnovije LunaOS izdanje](https://github.com/gregpavoine/LunaOS/releases), pokreće sve migracije na čekanju kako bi tvoj sustav uskladio s najnovijim i ažurira sve pakete sustava iz [LunaOS Arch Mirror](https://github.com/omacom-io/omarchy-mirror) i [AUR](https://aur.archlinux.org/) (ako imaš instalirane bilo kakve AUR pakete).

Kad se objave nova izdanja, ikona kružne strelice pojavit će se desno od tvog sata. Klikni je i proces ažuriranja će započeti.

![update-available](images/update-available.webp)

### Četiri kanala

LunaOS se ažurira duž četiri kanala: stable, RC, edge i dev. Nove instalacije počinju na kanalu stable, koji prati [službena izdanja](https://github.com/gregpavoine/LunaOS/releases/), kao i [stabilno LunaOS Arch zrcalo](https://github.com/omacom-io/omarchy-mirror) koje zaostaje mjesec dana za najnovijim, kako bismo mogli uhvatiti sve nove nekompatibilnosti koje zahtijevaju promjene konfiguracije prije nego što ljudima uzrokuju probleme.

Ali ako želiš pomoći u uočavanju tih potencijalnih problema, možeš koristiti kanal edge. To će tvoje LunaOS pakete držati u praćenju najnovijih razvojnih buildova i omogućuje ti ažuriranje na najnovije Arch pakete čim budu dostupni. Ovo bi trebao raditi samo ako si iskusan s Linuxom i znaš kako oporaviti sustav koji ima probleme.

Prije svakog novog glavnog izdanja, radit ćemo završnu validaciju koristeći kanal RC. Ako te zanima pomoć sa završnim poliranjem, dođi visjeti u #lunaos-release-candidates na Discordu.

Na kraju, tu je kanal dev, koji LunaOS povezuje izravno s git checkoutom izvornog koda u `~/lunaos`, u kombinaciji s paketima edge. Ovaj kanal trebao bi koristiti samo ako si iskusan korisnik Linuxa, radiš izravno na LunaOSju i spreman si tolerirati kvarove.

Između kanala možeš prebacivati koristeći _Ažuriraj > Channel_ iz izbornika LunaOS (ili `lunaos-channel-set` u terminalu).

### Ažuriranja firmwarea

Tvoji paketi nisu jedina stvar koja zastarijeva. Mnogi prijenosnici i periferni uređaji isporučuju BIOS, SSD i firmware docka putem Linux Vendor Firmware Service, a _Ažuriraj > Firmware_ u izborniku LunaOS dohvatit će i instalirati sve što tvoj hardver čeka. Instalira `fwupd` prvi put kad ga pokreneš. Mnogo firmwarea može se zapisati samo tijekom ponovnog pokretanja, pa se nemoj iznenaditi ako te zatraži jedno.

### Upozorenje o izravnim ažuriranjima pacman/yay

Ako si već upoznat s Archom, možda ćeš biti u iskušenju da samo sam pokreneš `pacman -Syu` ili `yay -Syu`, ali ako to učiniš, propustit ćeš snapshot, migracije i ažuriranja konfiguracije koje LunaOS pokreće zajedno s novim paketima. Zato će LunaOS zapravo zaustaviti izravnu nadogradnju sustava i uputiti te na `lunaos update` umjesto toga. (Ako stvarno znaš što radiš, stražar će ti reći kako ga zaobići za jednu transakciju.)

### Vraćanje loših ažuriranja

Ako ikada imaš problem nakon ažuriranja, možeš vratiti sustav na snapshot snimljen prije ažuriranja. Samo ponovno pokreni i odaberi snapshot u izborniku za pokretanje od prije početka ažuriranja.

![bootloader](images/bootloader.webp)

Ako su nekako tvoje konfiguracijske datoteke oštećene, možeš izvršiti i ponovnu instalaciju LunaOSja koristeći `lunaos reinstall` u terminalu. To će ponovno instalirati sve zadane LunaOS pakete, postaviti te na stable i downgradirati sve pakete koji su previše novi te resetirati sve konfiguracijske datoteke. Imaj na umu da će sve tvoje korisničke promjene konfiguracije u odnosu na LunaOS zadane postavke ovime biti prebrisane!
