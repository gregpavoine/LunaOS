# Merkingar

LunaOS gerir þér kleift að stilla fyrirtækjamerki þitt eða persónulega mynd fyrir bæði aflæsingu við ræsingu, skjáverjann og Um-skjáinn.

### Aflæsing við ræsingu

Þú getur notað `lunaos plymouth preview` til að sjá hvernig sérsniðið merki þitt og litir myndu líta út. Það tekur bakgrunnslit, textalit, merki á png-sniði og slóð fyrir forskoðunarmyndina:

```
lunaos plymouth preview '#1d2021' '#ebdbb2' logo.png preview.png
```

Beittu svo uppsetningunni með `lunaos plymouth set '#1d2021' '#ebdbb2' logo.png`, sem mun líka gefa SDDM innskráningarskjánum sömu liti og merki. Ef þú vilt snúa aftur geturðu notað `lunaos plymouth reset`.

 ![branding-plymouth-shopify](images/branding-plymouth-shopify.webp)

### Skjáverji

Þú getur breytt merkinu sem notað er fyrir skjáverjann undir _Stíll > Skjáverji_. Það er ASCII-merki, svo þú getur breytt textanum beint, en þú getur líka gefið því png- eða svg-mynd og við umbreytum henni í ASCII. Það lítur ansi vel út.

 ![branding-screensaver](images/branding-screensaver.webp)

Það eru þrír færslur í þeirri valmynd:

- **Breyta texta** opnar `~/.config/lunaos/branding/screensaver.txt` í ritlinum þínum. Sláðu inn eða límdu hvað sem þér lystir — ASCII-list, nafn þitt, blótsyrði. Vistaðu og hættu, og skjáverjinn fer strax í gang svo þú getir séð það.
- **Stilla frá mynd** opnar skráaveljara fyrir png eða svg, umbreytir því í ASCII og sýnir þér útkomuna. Merki með greinilega skugamynd virka mun betur en ljósmyndir.
- **Endurheimta sjálfgildi** setur LunaOS-merkið til baka.

### Um-skjár

Sömu þrjá valkostir eru undir _Stíll > Um_ fyrir _Um_-skjáinn sem þú færð úr LunaOS-valmyndinni, og þeir virka eins — skráin er `~/.config/lunaos/branding/about.txt`, og Um-glugginn birtist eftir hverja breytingu. Um-listin er umbreytt í minni stærð en skjáverjans, þar sem hún þarf að passa í glugga frekar en að fylla skjáinn þinn.

Á meðan glugginn er opinn hallar grænn glampi yfir listina á nokkurra sekúndna fresti og skilur hana svo aftur kyrrstæða. Þín eigin list fær það líka, svo framarlega sem sérhver stafur í henni er einn dálkur á breidd — allt sem _Stilla frá mynd_ framleiðir er það. List byggð úr emoji eða tvíbreiðum stöfum helst kyrr í staðinn, og skjárinn gerir það líka ef þú geymir þína eigin fastfetch-stillingu: kyrrstætt merki í þessum tilvikum er hreyfingin að halda sig frá frekar en að mistakast, þar sem að renna glampa yfir þau myndi lenda afgangi línunnar á röngum stað.

 ![branding-about](images/branding-about.webp)

### Að umbreyta myndum sjálfur

Báðir _Stilla frá mynd_ valkostirnir kalla bara á `lunaos transcode ascii`, sem þú getur keyrt beint ef þú vilt stjórn á umbreytingunni:

```
lunaos transcode ascii ~/logo.svg ~/.config/lunaos/branding/screensaver.txt --width 100
```

Það tekur `--width` og `--height` í terminal-dálkum og röðum, `--mode` sem er annaðhvort `braille` (sjálfgildið, og mun fínna) eða `block`, `--threshold` prósentu til að ákveða hvaða pixlar teljast hluti af merkinu, og `--invert` fyrir þegar merkið þitt er ljóst á dökkum bakgrunni. Ef umbreyting kemur út sem klumpur er þröskuldurinn venjulega hnappurinn til að snúa.

### Orð í stað merkis

`lunaos ascii` teiknar texta í Delta Corps Priest 1, FIGlet-leturgerðinni sem LunaOS-orðmerkið sjálft er teiknað í, svo skjáverji geti sagt eitthvað frekar en sýnt mynd:

```
lunaos ascii "Back in five" > ~/.config/lunaos/branding/screensaver.txt
```

Það tekur textann sem viðföng, eða les hann úr röri þegar engin eru gefin. Leturgerðin ber aðeins stafi og bil — hún var teiknuð án tölustafa eða greinarmerkja — svo öllu öðru er sleppt og það nefnt á stderr frekar en það gleypt hljóðlega.
