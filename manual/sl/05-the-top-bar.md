# Menijska vrstica

Trak vzdolž vrha vašega zaslona je vrstica LunaOS. To ni privita vrstica stanja, ampak del lupine LunaOS, enega samega dolgo delujočega procesa Quickshell, ki riše tudi meni, obvestila, pojavna okna OSD in zaklenjeni zaslon. Zato se odlično ujema s temo vsega drugega in zato se plošča odpre v hipu, namesto da bi pognala novo aplikacijo.

To je tudi edini del namizja, ki je vedno na zaslonu, zato je vredno vedeti, kaj počnejo vsi tisti majhni glifi.

## Kaj je na njej privzeto

Vrstica ima tri odseke. Na levi sedita logotip LunaOS (zaganjalnik menija) in kazalniki delovnih prostorov. Na sredini dobite kazalnike stanja, uro, Elsewhen (svetovno uro), vreme in značko za posodobitev LunaOSja. Na desni: razpored tipkovnice (ko je konfiguriranih več), sistemski pladenj, agenti, bluetooth, omrežje, zvok, zaslon in napajanje.

Nekaj od teh se prikaže le, ko imajo kaj za povedati. Razpored tipkovnice se pojavi le, če ste konfigurirali več kot en razpored. Značka za posodobitev se pojavi le, ko čaka posodobitev LunaOSja. Ikona agentov pa se pojavi prvič, ko LunaOS najde uporabo kodiranja z AI na računalniku (glejte [AI](17-ai.md)).

## Klikanje naokoli

Skoraj vsak gradnik naredi nekaj ob levem, desnem in srednjem kliku, več pa se jih odzove na drsenje. To je del, ki ga ljudje zgrešijo — desni in srednji gumb sta tam, kjer se skriva veliko dobrih stvari.

| Gradnik | Levi | Desni | Srednji / drsenje |
| --- | --- | --- | --- |
| Meni | Meni LunaOS | Nov terminal | — |
| Delovni prostori | Osredotoči ta delovni prostor | — | Drsenje: naslednji/prejšnji delovni prostor |
| Ura | Pojavno okno koledarja | Kroženje med formati zapisa | Srednji: izbirnik časovnega pasu |
| Vreme | Pojavno okno napovedi | Celotno vreme kot obvestilo | Srednji: osveži |
| Zvok | Zvočna plošča | Utišaj | Srednji: plošča · drsenje: glasnost |
| Mikrofon | Utišaj mikrofon | — | Srednji: zvočna plošča · drsenje: vhodna glasnost |
| Omrežje | Omrežna plošča | — | — |
| Bluetooth | Plošča Bluetooth | Preklopi radio | — |
| Zaslon | Plošča zaslona | — | Drsenje: svetlost |
| Napajanje | Plošča napajanja | Preklopi odstotek baterije | — |
| Mediji | Predvajaj/premor | Pojavno okno z naslovnico | Srednji: naprej · drsenje: nazaj/naprej |
| Agenti | Plošča agentov | Zaženi svojega agenta | Srednji: naslednja naročnina |
| Pladenj | Lebdenje razkrije predal | Desni klik na ševron za upravljanje | — |
| Posodobitev LunaOS | Zaženi posodobitev | — | — |

Vse v tej tabeli ni na vaši vrstici takoj po namestitvi. Gradnik za medije (MPRIS trenutno predvaja z drsečim zapisom skladbe in izvajalca) in gradnik za mikrofon sta vgrajena, vendar privzeto izklopljena — dodajte ju, če ju želite, kot je opisano spodaj.

## Plošče

Klik ikone v vrstici odpre ploščo, ki je pravo pojavno okno z drsniki, seznami in navigacijo s tipkovnico namesto orodjarnega namiga. Vsaka ima tudi bližnjico, tako da vam nikoli ni treba meriti na glif s 16 slikovnimi pikami:

| Bližnjica | Plošča |
| --- | --- |
| `Super + Ctrl + A` | Zvok |
| `Super + Ctrl + W` | Omrežje |
| `Super + Ctrl + B` | Bluetooth |
| `Super + Ctrl + D` | Zaslon |
| `Super + Ctrl + P` | Napajanje |
| `Super + Ctrl + Alt + D` | Koledar |
| `Super + Ctrl + Alt + E` | Svetovna ura |
| `Super + Ctrl + 1-9` | Preklopi n-to ploščo v desnem odseku |

Plošče niso odčitki. Na njih stvari dejansko naredite:

- **Zvok** ima glavni drsnik glasnosti, izbirnik izhodne naprave in mešalnik po aplikacijah, tako da lahko utišate tisti en zavihek v brskalniku, ne da bi se dotaknili vsega drugega.
- **Omrežje** išče Wi-Fi, prikazuje moč signala, se povezuje in vam omogoča, da izberete ponudnika DNS.
- **Bluetooth** navede vaše naprave s povezovanjem/prekinjanjem in ravmi baterije.
- **Napajanje** prikazuje statistiko baterije, preklaplja profile napajanja (zapomni si ločeno izbiro za baterijo in AC) in izpiše nekaj informacij o sistemu.
- **Zaslon** nosi drsnik svetlosti, velikost besedila, prednastavitve skaliranja zaslona in — ko imate več kot en zaslon — kontrolnike po zaslonu. Za globljo zgodbo glejte [zasloni](33-monitors.md).
- **Ura** odpre mesečno mrežo s številkami tednov ISO in pomikom po mesecih.

Vsaka plošča sprejme tipkovnico prav tako kot miško: puščice se premikajo, Return aktivira, Tab naredi korak do sosednje plošče in Escape zapre.

`Super + Ctrl + 1-9` šteje plošče od leve proti desni v desnem odseku, pri čemer preskoči pladenj, saj nima lastne plošče. Število se torej ujema z ikono, na katero bi pokazali.

### Tailscale in Dropbox

Še dva gradnika se pojavita v vrstici šele, ko ustrezno storitev namestite iz **Install → Service**, in oba sta vredna poznavanja, saj počneta več kot poročata o stanju.

Plošča **Tailscale** povezuje in prekinja tailnet, preklaplja med računi in izbere izhodno vozlišče (v tem seznamu se pojavijo tako vaši lastni stroji kot regije Mullvad). Preiskuje tudi vaše stroje — in ko je en izbran, `s` mu pošlje datoteke prek Taildropa, kar je najhitrejši način, da datoteko premaknete v telefon ali drug prenosnik. `c` kopira IP stroja, `n` njegovo ime in `d` njegovo polno ime DNS. Na vsaki vrstici stroja je tudi gumb za pošiljanje, če raje klikate. Enako iz terminala je `lunaos tailscale send <machine> [file...]`.

Plošča **Dropbox** obravnava prijavo, prikazuje, koliko prostora za shranjevanje ste porabili, in navede nedavno sinhronizirane datoteke.

Odstranitev katere koli storitve njen gradnik vzame iz vrstice.

## Kazalniki

Majhna gruča na sredini je gradnik kazalnikov. To so statusni glifi za načine, ki ste jih vklopili: ne moti, nočna svetloba, [opomnik](09-reminders.md) v vrsti, aktivno snemanje zaslona, ostanek budnega stanja in [narekovanje](11-text-extraction-dictation.md). Svetijo, ko je način aktifven, sicer pa ostanejo stran — z miško pojdite levo od ure (nad njo na navpični vrstici), da pokukate med neaktivne. Klik kazalnika ta način preklopi. Še eden je samo statusni: oddaljena seja sveti rdeče, medtem ko je nekdo povezan s tem računalnikom prek gliffa.

Če želite, da bi bili vedno vidni, nastavite `alwaysShow` na `true` na gradniku. In če vas zanimajo samo nekateri od njih, navedite tiste, ki jih želite, v `items`: `["Dnd", "Reminder", "NightLight"]`. Imate lahko več kot en gradnik kazalnikov, tako da različni odseki prikazujejo različne podmnožice.

## Prerazporejanje vrstice

Vrstica se konfigurira sama. Za premikanje stvari vam ni treba odpirati konfiguracijske datoteke.

Primite prazen košček vrstice okoli sredine in ga povlecite proti drugemu robu zaslona, in vrstica se prestavi tja — levo, desno, zgoraj ali spodaj vse deluje, vsak gradnik pa se prilagodi (navpične vrstice padejo nazaj v kompaktne oblike samo z ikonami). Enako vlečenje začne klik-in-držanje. Dvokliknite z levim na isti prazen prostor, da ozadje vrstice krožite od prosojnega do steklenega in neprosojnega. In povlecite kateri koli gradnik, da ga prerazporedite ali vržete v drug odsek.

Gradnik, ki ga želite imeti pri roki, a ne vedno na očeh, lahko spravite v predal pladnja: povlecite ga na pladenj in predal se odpre, da ga sprejme. Tam sedi za ikonami aplikacij, pokaže se ob vsakem odprtju predala in predal drži odprt, dokler je odprta njegova plošča. Povlecite ga nazaj na vrstico, da ga vrnete.

Če raje izberete iz menija, ima **Trigger → Appearance → Menu Bar** položaj, **Bar Background** pa izbere, kako se vrstica nariše: **Opaque Bar**, **Glassy Bar** (prosojna, s tem, kar je za njo, zamegljenim) ali **Clear Bar**.

Enake stvari imajo ukaze, kar je tisto, kar želite za namestitev z [dotfiles](31-dotfiles.md):

```bash
lunaos bar position bottom
lunaos bar background glassy  # opaque, glassy, none, or cycle
lunaos bar move lunaos.clock --section center --index 0
lunaos bar move lunaos.bluetooth --section tray  # into the tray's drawer
lunaos bar set lunaos.clock format "HH:mm"
lunaos bar defaults          # back to the shipped layout
```

Za dodajanje ali odstranjevanje gradnika v celoti uporabite ukaze za vtičnike. `lunaos plugin list` izpiše vsak gradnik, ki ga lupina pozna, z njegovim id-jem, nato pa:

```bash
lunaos plugin enable lunaos.media --section center
lunaos plugin disable lunaos.weather
```

## Skrivanje vrstice

`Super + Shift + Space` vrstico izklopi in znova vklopi, ne da bi ubil lupino — plošče in bližnjice še naprej delujejo, samo slikovne pike dobite nazaj. Je tudi v meniju pod **Trigger → Appearance → Menu Bar → Hide Bar**.

## Konfiguracijska datoteka

Vse se shrani v `~/.config/lunaos/shell.json` pod tipko `bar`. Tu je okrajšana različica:

```json
{
  "version": 1,
  "bar": {
    "position": "top",
    "background": "opaque",
    "centerAnchor": "lunaos.clock",
    "layout": {
      "left": [{ "id": "lunaos.menu" }, { "id": "lunaos.workspaces" }],
      "center": [{ "id": "lunaos.clock", "format": "HH:mm" }],
      "right": [{ "id": "lunaos.audio" }, { "id": "lunaos.power" }]
    }
  }
}
```

Vsak gradnik je en vnos v eni od treh tabel postavitve, njegove nastavitve pa sedijo vdelano na tem vnosu — ni ločene datoteke z nastavitvami in ni podobiekta `config`. Ure `format`, `formatAlt` (v kar kroži desni klik) in `verticalFormat` vsi živijo prav tam na `{ "id": "lunaos.clock" }`.

`centerAnchor` imenuje edini sredinski gradnik, ki se pripne na natančno sredino zaslona, drugi pa ga obdajo. Tako ura ostane točno na sredini, tudi ko vreme in značka za posodobitev prihajata in odhajata. Če ga nastavite na prazen niz, je seznam na sredini samo usredinjen kot skupina.

Eno pravilo, ki si ga je vredno zapomniti: **ko imate svoj `shell.json`, je kanoničen**. Dokler ničesar ne prilagodite, lupina bere LunaOSjevo privzeto datoteko. Trenutek, ko povlečete gradnik, zaženete `lunaos bar` ali datoteko uredite sami, postane vaša — ni globokega združevanja, zato novi privzeti gradniki v prihodnjih izdajah LunaOSja ne bodo samodejno prikazani v vaši vrstici. `lunaos bar defaults` dobavljeno postavitev vrne, kadar koli želite čist list.

Ista datoteka hrani tudi vaše timingse nedejavnosti na zgornji ravni, zunaj tipke `bar`: `idle.screensaver` in `idle.lock`, oba v sekundah od tistega, ko ste postali nedejavni. Privzeti ohranjevalnik torej začne pri 150 sekundah, zaklepanje pa pri 300.
