# Brendiranje

LunaOS vam omogućuje postavljanje logotipa vaše tvrtke ili osobne slike za otključavanje pri pokretanju, čuvara zaslona i zaslon O programu.

### Otključavanje pri pokretanju

Možete koristiti `lunaos plymouth preview` da vidite kako bi izgledali vaš prilagođeni logotip i boje. Prima boju pozadine, boju teksta, png logotip i putanju za sliku pregleda:

```
lunaos plymouth preview '#1d2021' '#ebdbb2' logo.png preview.png
```

Zatim primijenite postavke s `lunaos plymouth set '#1d2021' '#ebdbb2' logo.png`, što će i SDDM zaslonu za prijavu dati iste boje i logotip. Ako želite vratiti, možete koristiti `lunaos plymouth reset`.

 ![branding-plymouth-shopify](images/branding-plymouth-shopify.webp)

### Čuvar zaslona

Logotip koji se koristi za čuvara zaslona možete promijeniti pod _Stil > Screensaver_. To je ASCII logotip, pa tekst možete uređivati izravno, ali mu možete dati i png ili svg sliku, a mi ćemo je pretvoriti u ASCII. Izgleda prilično sjajno.

 ![branding-screensaver](images/branding-screensaver.webp)

U tom izborniku nalaze se tri stavke:

- **Edit Text** otvara `~/.config/lunaos/branding/screensaver.txt` u vašem uređivaču. Upišite ili zalijepite što god želite — ASCII umjetnost, svoje ime, prosta riječ. Spremite i izađite, a čuvar zaslona odmah se pokrene da ga možete vidjeti.
- **Set From Image** otvara birač datoteka za png ili svg, pretvara ga u ASCII i prikazuje vam rezultat. Logotipi s jasnom siluetom rade daleko bolje od fotografija.
- **Restore Default** vraća LunaOS logotip.

### Zaslon O programu

Iste tri opcije nalaze se pod _Stil > About_ za zaslon _O programu_ koji dobivate iz LunaOS izbornika i rade identično — datoteka je `~/.config/lunaos/branding/about.txt`, a prozor O programu pojavi se nakon svake promjene. Umjetnost za O programu pretvara se u manju veličinu od one za čuvara zaslona, jer mora stati u prozor umjesto da ispuni vaš zaslon.

Dok je prozor otvoren, odsjaj zelene svako malo se nasloni preko umjetnosti svako nekoliko sekundi, a zatim je opet ostavi mirnom. Vaša vlastita umjetnost također ga dobiva, sve dok je svaki znak u njoj širok jedan stupac — sve što _Set From Image_ proizvede jest. Umjetnost izrađena od emojija ili znakova dvostruke širine umjesto toga ostaje mirna, a tako i zaslon ako zadržite vlastiti fastfetch konfiguracijski zapis: miran logotip u tim je slučajevima animacija koja se drži podalje, a ne kvara, jer bi klizanje odsjaja preko njih ostatak retka ostavilo na pogrešnom mjestu.

 ![branding-about](images/branding-about.webp)

### Samostalno pretvaranje slika

Obje opcije _Set From Image_ samo pozivaju `lunaos transcode ascii`, što možete pokrenuti izravno ako želite kontrolu nad pretvorbom:

```
lunaos transcode ascii ~/logo.svg ~/.config/lunaos/branding/screensaver.txt --width 100
```

Prima `--width` i `--height` u stupcima i recima terminala, `--mode` koji je `braille` (zadano i mnogo finije) ili `block`, postotak `--threshold` za odlučivanje koji se pikseli računaju kao dio logotipa te `--invert` za slučaj kad je vaš logotip svijetao na tamnoj pozadini. Ako pretvorba ispadne kao mrlja, prag je obično gumb koji treba okrenuti.

### Riječi umjesto logotipa

`lunaos ascii` crta tekst u Delta Corps Priest 1, FIGlet fontu u kojem je nacrtan i sam LunaOS žig riječi, pa čuvar zaslona može nešto reći umjesto da prikazuje sliku:

```
lunaos ascii "Back in five" > ~/.config/lunaos/branding/screensaver.txt
```

Prima tekst kao argumente ili ga čita iz cijevi kad nijedan nije zadan. Font nosi samo slova i razmake — nacrtan je bez znamenki ili interpunkcije — pa se sve drugo odbacuje i imenuje na stderr umjesto da se tiho proguta.
