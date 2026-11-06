# Windows VM

LunaOS ponuja enostaven način za zagon Windows prek Docker VM. Namestite ga lahko z _Namesti > Windows_ iz menija LunaOS (`Super + Space`).

Vaš računalnik za to potrebuje virtualizacijo KVM, kar večina podpira — včasih pa je izklopljena v BIOS-u in namestitveni program vam bo povedal, če je tako. Potrebovali boste tudi prostor na disku: kolikor ga daste Windowsu, plus približno 10 GB za samo sliko.

Namestitveni program vpraša, koliko RAM-a, koliko jeder CPU in koliko diska naj da (64 GB ali več je smiselna spodnja meja), nato za uporabniško ime in geslo Windows. Če pustite prazno, dobite `docker` / `admin`. Prenos traja — 10-15 minut je običajno — in napredek lahko spremljate v brskalniku na `http://127.0.0.1:8006`. Brskalnik pred odpiranjem konzole zahteva isto uporabniško ime in geslo.

 ![windows-vm](images/windows-vm.webp)

## Uporaba

Ko je nameščen, zaženite _Windows_ iz zaganjalnika aplikacij. To zažene VM, če še ne teče, in vas poveže prek RDP, čez cel zaslon. Pri hladnem zagonu mu dajte 15-30 sekund.

Seja RDP prenaša zvok, vaš mikrofon in deljeno odložišče, tako da kopiranje besedila med Linuxom in Windows preprosto deluje. Ločljivost sledi vašemu oknu, LunaOS pa posreduje vaše povečevanje zaslona, tako da na zaslonu HiDPI ni zamegljena zmešnjava.

Ko zaprete okno RDP, se VM samodejno zaustavi. Če ga želite pustiti teči — recimo, da kaj dela v notranjosti — ga namesto tega zaženite z `lunaos windows vm launch --keep-alive`.

Preostali ukazi so na istem ukazu:

```bash
lunaos windows vm status    # is it running?
lunaos windows vm stop      # shut it down
lunaos windows vm launch    # start and connect
```

## Deljenje datotek

Mapa `~/Windows` v vaši domači mapi se samodejno deli z VM. Datoteke postavite tja, če želite, da so dostopne Windowsu. VM nima dostopa do nobenega drugega dela vašega datotečnega sistema, tako da ste varni pred vsem zlobnim na strani Windows. Njegov lastni virtualni disk je na voljo na `~/.windows`.

Te znane domače poti ostanejo na svojih lastnih datotečnih sistemih. Lahko so tudi simbolne povezave do map, ki jih posedujete, kar je uporabno, kadar virtualni disk živi na večjem pogonu. Namestitveni program meri prosti prostor na datotečnem sistemu, ki dejansko vsebuje `~/.windows`, ne nujno na datotečnem sistemu z vašo domačo mapo.

Disk in deljene poti hranite kot ločeni, neprekrivajoči se mapi. Odstranitev namerno izprazni mapo diska, a ohrani deljeno mapo. Tik pred izbrisom LunaOS izvede omejen pregled obsežnosti in zavrne odstranitev česar koli, če pregled poteče ali ne more dokazati, da sta drevesi ločeni.

Pred zagonom VM LunaOS odpre in pripne ti dve mapi, nato natančne inode map priklopi z vezavo na zasebna sidra na uporabnika pod `/var/lib/lunaos/windows/mounts`. Docker vidi samo ta sidra, zaščitena z rootom. To ohranja lokacije diskov po meri, hkrati pa preprečuje, da bi drug proces, ki teče kot vi, zamenjal preverjeno pot, preden jo privilegirani vsebnik porabi. Obstoječe mape diska in deljenja se med migracijo zožijo na način `0700`, tako da drugi lokalni računi ne morejo brskati po vsebini.

Vrata VM so vezana samo na localhost, tako da nič v vašem omrežju ne more doseči računalnika Windows. Spletna konzola prav tako zahteva konfigurirano uporabniško ime in geslo Windows, kar drugemu lokalnemu računu preprečuje poganjanje VM prek vrat 8006.

## Omejitve in licenciranje

Pri tej namestitvi ni prehoda GPU, zato ni primerno za igranje iger ali urejanje videa. Je odličen način za zagon Microsoft Office ali česar koli drugega, kar nujno morate imeti.

Nameščena različica je Windows 11 Pro, neaktivirana. Za uporabo zaklenjenih funkcij boste potrebovali lastni licenčni ključ.

Če je ta računalnik dobavil Windows, je ključ OEM po namestitvi LunaOSja še vedno v vdelani programski opremi. Izpišite ga z `lunaos windows key`. Ta ključ je vezan na ta računalnik — aktiviral bo Windows, ponovno nameščen na tej strojni opremi, običajno pa ne bo aktiviral VM.

Razporeditev virov lahko pozneje spremenite s ponovnim zagonom `lunaos-windows-vm install`, kar iz vaših odgovorov prepiše konfiguracijo VM. Datoteka compose sama zdaj živi na `/var/lib/lunaos/windows/docker-compose.yml` in je v lasti roota — to je namerno, da proces, ki teče kot vi, je ne more prepisati in bi privilegirani zagon priklopil vaš celoten disk v vsebnik. Če jo morate ročno urediti (na primer za priklop naprave USB), jo uredite s `sudo` in si oglejte vse možnosti na [projektu Dockur Windows](https://github.com/dockur/windows).

Če se želite znebiti vsega, uporabite _Odstrani > Windows_ iz menija LunaOS. To izbriše disk VM in vse svoje podatke, zato najprej poskrbite, da je vse, kar vam je mar, iz `~/Windows`.
