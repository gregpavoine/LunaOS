# Virtuálny stroj s Windowsom

LunaOS ponúka jednoduchý spôsob, ako prevádzkovať Windows prostredníctvom virtuálneho stroja v Dockeri. Môžeš ho nainštalovať pomocou _Inštalovať > Windows_ z menu LunaOS (`Super + Space`).

Tvoj počítač na to potrebuje virtualizáciu KVM, ktorú väčšina má — ale niekedy je v BIOSe vypnutá a inštalátor ti povie, ak je to tak. Budeš chcieť aj miesto na disku: čokoľvek dáš Windowsu plus asi 10 GB na samotný obraz.

Inštalátor sa opýta, koľko RAM, koľko jadier procesora a koľko miesta na disku chceš odovzdať (64 GB alebo viac je rozumné minimum), a potom na používateľské meno a heslo do Windowsu. Ak ich necháš prázdne, dostaneš `docker` / `admin`. Sťahovanie chvíľu trvá — 10 – 15 minút je bežné — a priebeh môžeš sledovať v prehliadači na adrese `http://127.0.0.1:8006`. Prehliadač si pred otvorením konzoly vyžiada to isté používateľské meno a heslo.

 ![windows-vm](images/windows-vm.webp)

## Používanie

Keď je nainštalované, spusti _Windows_ zo spúšťača aplikácií. Tým sa virtuálny stroj spustí, ak ešte nebeží, a pripojí ťa cez RDP na celú obrazovku. Pri studenom štarte mu daj 15 – 30 sekúnd.

Relácia RDP prenáša zvuk, tvoj mikrofón a zdieľanú schránku, takže kopírovanie textu medzi Linuxom a Windowsom jednoducho funguje. Rozlíšenie sa riadi tvojím oknom a LunaOS prenáša tvoje škálovanie displeja, takže na obrazovke HiDPI to nie je rozmazaný neporiadok.

Keď okno RDP zatvoríš, virtuálny stroj sa automaticky vypne. Ak by si ho radšej nechal bežať — povedzme, že tam máš niečo, čo pracuje — spusti ho namiesto toho pomocou `lunaos windows vm launch --keep-alive`.

Ostatné ovládacie prvky sú na tom istom príkaze:

```bash
lunaos windows vm status    # is it running?
lunaos windows vm stop      # shut it down
lunaos windows vm launch    # start and connect
```

## Zdieľanie súborov

Adresár `~/Windows` v tvojom domovskom adresári sa automaticky zdieľa s virtuálnym strojom. Ak chceš, aby boli súbory dostupné pre Windows, umiestni ich tam. Virtuálny stroj nemá prístup k žiadnej inej časti tvojho súborového systému, takže si v bezpečí pred všetkým nepríjemným na strane Windowsu. Jeho vlastný virtuálny disk je dostupný na `~/.windows`.

Tieto známe cesty v domovskom adresári zostávajú vo svojich vlastných súborových systémoch. Môžu to byť aj symbolické odkazy na adresáre, ktoré vlastníš, čo je užitočné, keď virtuálny disk žije na väčšej jednotke. Inštalátor meria voľné miesto na súborovom systéme, ktorý skutočne obsahuje `~/.windows`, nie nevyhnutne na súborovom systéme obsahujúcom tvoj domovský adresár.

Cesty disku a zdieľané cesty udržiavaj ako samostatné, neprekrývajúce sa adresáre. Odstránenie zámerne vyprázdni adresár disku, ale zachová zdieľaný adresár. Bezprostredne pred odstránením LunaOS vykoná ohraničenú kontrolu uzavretia (containment) a odmietne čokoľvek odstrániť, ak táto kontrola vyprší alebo nedokáže preukázať, že tieto dva stromy sú oddelené.

Pred spustením virtuálneho stroja LunaOS otvorí a pripne tieto dva adresáre a potom pomocou bind-mount pripojí presné inódy adresárov na súkromné kotvy pre jednotlivých používateľov pod `/var/lib/lunaos/windows/mounts`. Docker vidí len tieto kotvy chránené rootom. Tým sa zachovávajú vlastné umiestnenia disku a zároveň sa zabraňuje tomu, aby iný proces bežiaci pod tvojím používateľom vymenil skontrolovanú cestu skôr, než ju privilegovaný kontajner skonzumuje. Existujúce adresáre disku a zdieľané adresáre sa počas migrácie sprísnia na režim `0700`, aby iné lokálne účty nemohli prezerať ich obsah.

Porty virtuálneho stroja sú viazané len na localhost, takže nič v tvojej sieti sa nedostane k počítaču s Windowsom. Webová konzola tiež vyžaduje nakonfigurované používateľské meno a heslo do Windowsu, čo zabraňuje inému lokálnemu účtu ovládať virtuálny stroj prostredníctvom portu 8006.

## Obmedzenia a licencie

Pri tomto nastavení nie je k dispozícii prenechanie GPU (GPU passthrough), takže nie je vhodné na hranie hier ani na úpravu videa. Je to skvelý spôsob, ako prevádzkovať Microsoft Office alebo čokoľvek iné, čo bezpodmienečne musíš mať.

Nainštalovaná verzia je Windows 11 Pro, neaktivovaná. Na používanie funkcií s obmedzeným prístupom budeš potrebovať vlastný licenčný kľúč.

Ak sa tento počítač dodával s Windowsom, kľúč OEM je stále vo firmvéri aj po nainštalovaní LunaOS. Vypíš ho pomocou `lunaos windows key`. Tento kľúč je viazaný na tento počítač — aktivuje Windows preinštalovaný na tomto hardvéri, ale zvyčajne neaktivuje virtuálny stroj.

Prideľovanie prostriedkov môžeš neskôr zmeniť opätovným spustením `lunaos-windows-vm install`, čo prepíše konfiguráciu virtuálneho stroja z tvojich odpovedí. Samotný súbor compose teraz žije na `/var/lib/lunaos/windows/docker-compose.yml` a vlastní ho root — to je zámerné, aby proces bežiaci pod tvojím používateľom nemohol tento súbor prepísať a prinútiť privilegované spúšťanie pripojiť celý tvoj disk do kontajnera. Ak ho potrebuješ upraviť ručne (napríklad pripojiť zariadenie USB), uprav ho pomocou `sudo` a pozri si všetky možnosti v [projekte Dockur Windows](https://github.com/dockur/windows).

Ak sa chceš zbaviť celej veci, použi _Odstrániť > Windows_ z menu LunaOS. Tým sa odstráni disk virtuálneho stroja a všetky jeho údaje, takže sa najprv uisti, že všetko, na čom ti záleží, je vonku z `~/Windows`.
