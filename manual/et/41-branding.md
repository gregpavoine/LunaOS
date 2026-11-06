# Bränding

LunaOS võimaldab sul seada oma ettevõtte logo või isikliku pildi nii alglaadimise avamise, ekraanisäästja kui ka teabekuva jaoks.

### Alglaadimise avamine

Sa saad käsuga `lunaos plymouth preview` vaadata, milline sinu kohandatud logo ja värvid välja näeksid. See võtab taustavärvi, tekstivärvi, logo png-faili ja eelvaate pildi tee:

```
lunaos plymouth preview '#1d2021' '#ebdbb2' logo.png preview.png
```

Seejärel rakenda seadistus käsuga `lunaos plymouth set '#1d2021' '#ebdbb2' logo.png`, mis annab ka SDDM sisselogimiskraanile samad värvid ja logo. Kui tahad taastada, saad kasutada käsku `lunaos plymouth reset`.

 ![branding-plymouth-shopify](images/branding-plymouth-shopify.webp)

### Ekraanisäästja

Ekraanisäästja logo saad muuta jaotises _Stiil > Ekraanisäästja_. See on ASCII logo, nii et saad teksti otse muuta, aga võid sellele anda ka png- või svg-pildi ja me teisendame selle ASCII-ks. See näeb üsna lahe välja.

 ![branding-screensaver](images/branding-screensaver.webp)

Selles menüüs on kolm kirjet:

- **Muuda teksti** avab faili `~/.config/lunaos/branding/screensaver.txt` sinu redaktoris. Kirjuta või kleebi mida tahad — ASCII kunst, oma nimi, roppus. Salvesta ja välju ning ekraanisäästja käivitub kohe, nii et saad seda näha.
- **Sea pildist** avab failivalija png- või svg-faili jaoks, teisendab selle ASCII-ks ja näitab sulle tulemust. Selge siluetiga logod töötavad palju paremini kui fotod.
- **Taasta vaikimisi** paneb LunaOS logo tagasi.

### Teabekuva

Samad kolm valikut on jaotises _Stiil > Teave_ LunaOS menüüst avatava _Teave_ kuva jaoks ja need töötavad identselt — fail on `~/.config/lunaos/branding/about.txt` ja Teave aken ilmub pärast iga muudatust. Teave kunst teisendatakse väiksemaks kui ekraanisäästja oma, sest see peab mahtuma aknasse, mitte täitma kogu ekraani.

Kuni aken on avatud, libiseb iga paari sekundi tagant üle kunsti roheline helk ja jätab selle seejärel jälle liikumatuks. Sinu oma kunst saab selle samuti, kui iga märk selles on ühe tulba laiune — kõik, mida _Sea pildist_ toodab, on. Emoji-dest või kahelaiustest märkidest üles ehitatud kunst jääb selle asemel liikumatuks ja sama kehtib kuva kohta, kui hoiad enda fastfetch seadistust: liikumatuks jäänud logo tähendab neil juhtudel seda, et animatsioon hoiab eemale, mitte et see ebaõnnestuks, sest helgi üle nende libistamine paneks ülejäänud rea valesse kohta.

 ![branding-about](images/branding-about.webp)

### Piltide teisendamine ise

Mõlemad valikud _Sea pildist_ kutsuvad lihtsalt välja käsu `lunaos transcode ascii`, mida saad otse käivitada, kui tahad teisendust kontrollida:

```
lunaos transcode ascii ~/logo.svg ~/.config/lunaos/branding/screensaver.txt --width 100
```

See võtab `--width` ja `--height` terminali tulpades ja ridades, `--mode` valikuks kas `braille` (vaikimisi ja palju peenem) või `block`, `--threshold` protsendi otsustamiseks, millised pikslid logo osaks loetakse, ja `--invert` juhuks, kui sinu logo on heledal tume taust. Kui teisendus tuleb välja plönnina, on lävi tavaliselt see nupp, mida keerata.

### Sõnad logo asemel

`lunaos ascii` joonistab teksti Delta Corps Priest 1 kirjatüübis, FIGlet kirjatüübis, milles on joonistatud ka LunaOS sõnamärk ise, nii et ekraanisäästja saab midagi ütelda, mitte pilti näidata:

```
lunaos ascii "Back in five" > ~/.config/lunaos/branding/screensaver.txt
```

See võtab teksti argumentidena või loeb selle torust, kui argumente pole antud. Kirjatüüp sisaldab ainult tähti ja tühikuid — see joonistati ilma numbrite ja kirjavahemärkideta — nii et kõik muu jäetakse välja ja nimetatakse stderr-is, mitte ei neelata vaikselt alla.
