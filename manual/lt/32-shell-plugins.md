# Shell įskiepiai

LunaOS darbalaukis veikia kaip vienas ilgai gyvuojantis Quickshell procesas pavadinimu `lunaos-shell`, ir beveik viskas, ką matai ekrane, yra įskiepis jame. Juosta yra įskiepis. Taip pat skydeliai, krentantys iš jos, viso ekrano perteikliai, pavyzdžiui, emoji rinkiklis ir iškarpinės tvarkyklė, pats LunaOS meniu, užrakinimo ekranas, polkit dialogas ir nematomas paslaugos, kurios stebiavo akumuliatorių ir šildo ekraną nakčiai.

Tai ne vien įgyvendinimo detalė. Tai reiškia, kad gali išjungti darbalaukio dalis, jas pakeisti ar parašyti savo, neliesdamas nė vienos LunaOS šaltinio eilutės.

Pirmosios šalies įskiepiai pristatomi su LunaOS ir gyvena `$LUNAOS_PATH/shell/plugins/`. Viskas, ką pridedi pats — savo eksperimentai ar kažkas rasta GitHub — gyvena `~/.config/lunaos/plugins/`. Abu atrandami tuo pačiu būdu paleidimo metu, bet įtaisyti gauna patikimas shell sąsajas, o trečiųjų šalių įskiepiai gauna ribotą sąsają, apribotą jų pačių paslauga ir gyvavimo ciklu. Įtaisytųjų klonai išlaiko tik su šaltiniu susijusią konfigūraciją ir UI iškvietimus, reikalingus originaliam veikimui.

## Pamatyti, ką turi

```
lunaos plugin list
```

Tai išspausdina kiekvieną atrastą įskiepį su jo id, ar jis įjungtas, ar pirmosios, ar trečiosios šalies, jo tipus ir rodomą pavadinimą. Pridėk `--json`, jei tai perduodi kam nors kitam.

Įskiepių id yra pavadintiniai (namespaced). Įtaisyti visi prasideda `lunaos.` — `lunaos.clock`, `lunaos.network`, `lunaos.notifications` — ir tas pavadinimų erdve yra rezervuota, tad trečiosios šalies įskiepis niekada negali jos pretenduoti.

## Įjungti ir išjungti

```
lunaos plugin enable lunaos.tailscale
lunaos plugin disable lunaos.weather
```

Arba naudok meniu: _Nuostatos > Plugins_ turi Enable, Disable, Add, Clone ir Remove, kiekvienas su rinkikliu, kuris siūlo tik tuos įskiepius, kurie turi prasmę tam veiksmui.

Įjungta būsena saugoma `~/.config/lunaos/shell.json`, ir taisyklė šiek tiek skiriasi dviem įskiepių tipams. Trečiosios šalies įskiepis įjungtas būtent tada, kai jo id pasirodo kur nors tame faile — kaip juostos išdėstymo įrašas, kaip įrašas `plugins[]` arba kaip `bar.id`. Pirmosios šalies įskiepiai, kurie nėra juostos valdikliai, yra atvirkščiai: jie įjungti pagal nutylėjimą ir išsijungia tik būdami išvardyti `disabledPlugins[]`.

Pilnas juostos įskiepis iš viso neturi išjungtos būsenos. Visada yra būtent viena juosta, tad ją pakeiti įjungdamas kitą. Juostos valdiklių išdėstymas aptariamas [meniu juostoje](05-the-top-bar.md).

## Įskiepio pridėjimas iš git

Trečiosios šalies įskiepis yra tiesiog git saugykla su `manifest.json` jo šaknyje.

```
lunaos plugin add https://github.com/acme/lunaos-weather.git --enable
```

Prieš ką nors darant, jis paprastai tau pasako, kad įskiepiai veikia kaip savavališkas, nesmėliadėžytas kodas tavo ilgai gyvuojančiame shell procese, parodo tau URL ir paprašo patvirtinti. Priimk tai rimtai. Trečiųjų šalių įskiepių sąsaja tiesiogiai neatskleidžia autentifikavimo paslaugų, o pakaitinė juosta gauna tik ribotas galimybes sukonfigūruotiems neautentifikavimo UI. Vaizdinio įskiepiai vis dar dalijasi shell QML scena ir gali vaikščioti įprastais tėviniais objektais, o visas įskiepių kodas veikia su viskuo, ką tavo naudotojo paskyra gali pasiekti. Autentifikavimo būsena saugoma atskirai, laikant tas paslaugas už pasiekiamo šeimininko objektų grafo ribų. Pridėk tik tas saugyklas, kurias esi pasiruošęs paleisti, ir perskaityk jas prieš įjungdamas.

Pakaitinė juosta gali atvaizduoti įdiegtus valdiklius, bet paslaugomis paremti trečiųjų šalių valdikliai ten gali turėti sumažintą funkcionalumą, nes juostai neleidžiama prašyti kito įskiepio gyvos paslaugos objekto. Grįžk prie įtaisytos `lunaos.bar`, jei toks valdiklis reikia jo papildančios paslaugos.

Tada jis klonuoja saugyklą į pakopos katalogą, patvirtina manifestą, atsisako diegimo, jei kitas įskiepis jau pretenduoja į tą id, ir perkelia jį į `~/.config/lunaos/plugins/<id>/`. Be `--enable` jis paklausia, ar nori jį įjungti dabar, ir gali pasakyti ne ir pirmiausia eiti perskaityti kodą. Jis niekada nieko nepaleidžia iš įskiepio, niekada nevykdo diegimo kablio ir niekada neprašo sudo — jis klonuoja failus, patikrina manifestą ir perjungia bitą per IPC.

Atnaujinimas yra tos pačios kopijos spartusis persiuntimas (fast-forward pull):

```
lunaos plugin update acme.weather
lunaos plugin update
```

Be id jis atnaujina kiekvieną tavo turimą git valdomą įskiepį. Jis parodo tau diff prieš pritaikydamas, atsisako atnaujinti, jei turi vietinių pakeitimų, kurių jis negali greitai persiųsti, ir atšaukia, jei nauja versija nepavyksta patvirtinti.

```
lunaos plugin remove acme.weather
```

Pašalinimas pirmiausia išjungia įskiepį, tada jį ištrina, jei tai git kopija (saugykla vis dar prieš srovę), arba atsieja, jei tai simbolinė nuoroda. Rankų darbo įskiepio aplankas be git saugyklos perkeliamas į atsarginę kopiją su laiko antspaudu įskiepių kataloge, užuot ištrynęs jį visiškai.

## Įtaisyto klonavimas, kad jį pakeistum

Tai mano mėgstamiausia dalis. Jei nori pakeisti, kaip elgiasi įtaisytas valdiklis, neredaguok failų pagal `$LUNAOS_PATH` — jie priklauso paketui ir kitas atnaujinimas juos perrašys. Vietoj to jį klonuok:

```
lunaos plugin clone lunaos.clock
```

Tai nukopijuoją visą įskiepį į `~/.config/lunaos/plugins/dhh.clock` (tavo naudotojo vardas, ne mano), pervadina jį į "My Clock", įjungia jį ir perjungia shell nuo įtaisyto prie tavo kopijos — išlaikydamas esamo juostos valdiklio padėtį ir nuostatas. Pridėk `--edit`, kad naują katalogą iškart atvertum savo `$EDITOR`, ką meniu _Nuostatos > Plugins > Clone Plugin_ daro už tave.

Naudotojo vardo priešdėlis išlaiko tavo klono id tavuoju, tad juo dalintis nesusiduria su niekieno kito. Iškvietimai originaliam įtaisytam id nukreipiami į tavo kloną, tad nieko, kas nurodžiau į `lunaos.clock`, nereikia atnaujinti. O jei jį sujauksi, `lunaos plugin remove dhh.clock` grąžina įtaisytą atgal.

Failo išsaugojimas bet kur `~/.config/lunaos/plugins/` automatiškai perkrauna įskiepio kodą, tad gali palikti redaktorių atvirą ir stebėti, kaip atsiunčia tavo pakeitimai.

## Savo rašymas

Įskiepis yra katalogas su `manifest.json` ir šiek tiek QML. Manifestas deklaruoja `schemaVersion: 1`, `id`, `name`, `version`, vieną ar kelis `kinds` ir `entryPoints` objektą, nurodantį į QML failą kiekvienam tipui:

| Tipas | Kas tai |
|------|------------|
| `bar-widget` | Komponentas, kurį aktyvi juosta gali įmesti į skiltį |
| `panel` | Nuolatinis arba iškviestas plūduriuojantis langas |
| `overlay` | Viso ekrano perteiklis |
| `menu` | Iškviečiamas meniu paviršius |
| `service` | Nematomas singleton be UI |
| `bar` | Pilna juosta, pakeičianti įtaisytą |

Įskiepis gali deklaruoti kelis tipus vienu metu — media įskiepis yra ir `service`, ir `bar-widget`. Juostos valdikliai gauna papildomą `barWidget` bloką su rodomu pavadinimu, kategorija, neprivalomu `defaultSection` ir `allowMultiple`, kuris sako, ar prasminga turėti daugiau nei vieną jį juostoje. Dauguma valdiklių jį nustato į `false`; skirtukai ir indikatoriai nustato jį į `true`.

Prieš ką nors skelbdamas, patikrink tai:

```
lunaos plugin validate ./my-plugin
```

Tai paleidžia tuos pačius patikrinimus, kuriuos shell atlieka įkėlimo metu: schemos versiją, privalomus laukus, id, kuris nėra rezervuotas, įėjimo taškus, kurie yra saugūs santykiniai keliai ir iš tikrųjų egzistuoja, įėjimo tašką kiekvienam tipui, kurį pretendei, ir be simbolinių nuorodų niekur aplanko viduje.

Įskiepių kūrimui žiūrėk [shell nuorodą](https://github.com/omacom/omarchy/blob/quattro/docs/omarchy-shell.md) ir [pirmosios šalies įskiepių katalogą](https://github.com/omacom/omarchy/blob/quattro/shell/plugins/README.md).

## Dalinimasis savuoju su pasauliu

Kai sukūrei kažką, kas tau patinka, įdėk tai į viešą git saugyklą. Tai visas platinimo mechanizmas — tada bet kas gali paleisti `lunaos plugin add` prieš tavo URL ir turėti jį veikiantį per kelias sekundes.

Kad padėtum žmonėms jį iš tikrųjų rasti, išvardyk jį [plugins.omarchy.org](https://plugins.omarchy.org). Tai LunaOS shell įskiepių turgus (marketplace) ir pirmoji vieta, kurios žiūrėti, kai domiesi, ar kas nors jau sukūrę valdiklį, kurį ruošiesi rašyti. Peržiūrėk jį prieš pradedant!
