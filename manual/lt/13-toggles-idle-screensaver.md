# Perjungikliai, tuščioji eiga ir ekrano užsklanda

Daugelis to, ką keičiate kiekvieną dieną, nėra tikra nuostata. Tai režimas, kurį įjungiate valandai ir vėl išjungiate: naktinė šviesa, kai dirbate vėlai, netrukdyti, kai pristatote, nemiegoti, kai ką nors žiūrite. LunaOS juos vadina perjungikliais ir jie visi veikia vienodai — spartusis klavišas, meniu įrašas ir komanda, visi smogiantys į tą patį jungiklį.

### Perjungimo meniu

`Super + Ctrl + O` atidaro _Veiksmai > Perjungti_ tiesiogiai, arba galite ten nueiti iš LunaOS meniu (`Super + Space`). Viskas tame sąraše yra jungiklis, kurį galite perjungti negalvodami, kur gyvena būsena. Jungikliai, kaip darbalaukis atrodo, yra šalia, po _Veiksmai > Išvaizda_.

Iš terminalo tie patys jungikliai yra `lunaos toggle <thing>`. Paleiskite `lunaos toggle` atskirai, kad pamatytumėte visą grupę.

| Perjungiklis | Spartusis klavišas | Komanda |
| ------ | ------ | ------- |
| Naktinė šviesa | `Super + Ctrl + N` | `lunaos toggle nightlight` |
| Nutildyti pranešimus | `Super + Ctrl + ,` | `lunaos toggle notification silencing` |
| Nemiegoti (be užrakinimo tuščiojoje eigoje) | `Super + Ctrl + I` | `lunaos toggle idle` |
| Strigčių fiksavimas | — | `lunaos toggle crash-capture` |
| Ekrano užsklanda | — | `lunaos toggle screensaver` |
| [Herdr](https://herdr.dev) temos sinchronizacija | — | `lunaos toggle theme sync` |
| Meniu juosta | `Super + Shift + Space` | `lunaos toggle bar` |
| Jutiklinis kilimėlis | `XF86TouchpadToggle` | `lunaos toggle touchpad` |
| Lietimui jautrus ekranas | — | `lunaos toggle touchscreen` |
| Pristabdyti | — | `lunaos toggle suspend` |
| Hybrid GPU | — | `lunaos toggle hybrid gpu` |

Jutiklinio kilimėlio, lietimui jautraus ekrano ir Hybrid GPU jungikliai gyvena po _Veiksmai > Aparatinė įranga_ (`Super + Ctrl + H`), o ne po Perjungti, nes jie pasirodo tik kai iš tikrųjų turite tą aparatinę įrangą. Jutiklinio kilimėlio ir lietimui jautraus ekrano jungikliai išlieka po Hyprland perkrovimo — išjungto įrenginio pavadinimas išsaugomas mažame būsenos faile, kurį Hyprland paleidžiant perskaito, kad vėl jį išjungtų.

_Veiksmai > Išvaizda_ laiko meniu juostą ir animacijas, plius keletą dalykų, kurie nėra `lunaos toggle` komandos, bet elgiasi vienodai: akumuliatoriaus procentai juostoje, numatytasis darbo srities išdėstymas, langų tarpai (`Super + Shift + Backspace`) ir 1 lango kvadratinis kraštinių santykis (`Super + Ctrl + Backspace`).

Daugelis jų yra tik žymos failas po `~/.local/state/lunaos/toggles/`. Jei norite scenarijuje nuo jo šakotis, `lunaos-toggle-enabled` jums duoda išėjimo kodą, užuot vertęs jus ieškoti:

```bash
lunaos-toggle-enabled screensaver-off && echo "screensaver is off"
```

Žymos pavadintos pagal išjungtą būseną — `screensaver-off`, `suspend-off`, `bar-off` — todėl jų buvimas reiškia, kad funkcija yra išjungta.

### Indikatoriai juostoje

Kai režimas įjungtas, viršutinės juostos viduryje šalia laikrodžio gaunate mažą glifą. Tai indikatorių valdiklis ir jame yra diktavimas, ekrano įrašas, laukiantys priminimai, naktinė šviesa, netrukdyti ir nemiegoti.

Neaktyvūs indikatoriai yra paslėpti. Užveskite ant srities aplink juos ir jie blankiai pasirodo, todėl galite spustelėti vieną, kad jį įjungtumėte, nežinodami jo sparčiojo klavišo. Spustelėjus aktyvų, jis vėl išjungiamas. Jei mieliau norėtumėte juos visus matyti visą laiką, nustatykite `alwaysShow` į `true` ant `lunaos.indicators` įrašo `~/.config/lunaos/shell.json` — kaip konfigūruojami juostos valdikliai, žr. [viršutinę juostą](05-the-top-bar.md).

### Naktinė šviesa

`Super + Ctrl + N` pašildo ekraną iki 4000K, o paspaudus dar kartą grąžina jį į 6500K. Jį valdo hyprsunset, kurį perjungiklis jums paleidžia, jei jis dar neveikia.

Pagal numatytuosius nustatymus hyprsunset visai nieko nedaro jūsų ekranui. `~/.config/hypr/hyprsunset.conf` pateikiamas su tapatybės profiliu būtent tam, kad ekranas liktų nepaliestas, kol nepaprašysite šilumos. Jei mieliau norėtumėte, kad jis persijungtų pagal laikrodį, pakeiskite tai laiko profiliu:

```
profile {
    time = 20:00
    temperature = 4000
}
```

Tada paleiskite hyprsunset prisijungus pridėdami `o.launch_on_start("hyprsunset")` į `~/.config/hypr/autostart.lua`. Perjungiklio naudojama 4000K/6500K pora yra fiksuota, todėl konfigūracijos failas yra ta vieta, kur einate, jei norite kitokios temperatūros.

### Netrukdyti

`Super + Ctrl + ,` nutildo pranešimus. Kol jis įjungtas, jokie skrebučiai nepasirodo, o perbraukto varpelio indikatorius yra juostoje, kad primintų, kodėl darbalaukis nutilo.

Tačiau niekas neprarandama. Nutildytas pranešimas įrašomas tiesiai į jūsų pranešimų istoriją, o tai yra būtent tas įrašas, kurio norite, kai grįžtate ir galvojate, ką praleidote. Atidarykite ją su `Super + Shift + Alt + ,`. Likusią pranešimų istorijos dalį žr. [pastabose](10-notices.md).

Dviejų rūšių pranešimai vis tiek praeina: pačios LunaOS patvirtinimo skrebučiai dėl to, ką ką tik padarėte („Tema pakeista“, „Ekrano nuotrauka išsaugota“), ir kritiniai perspėjimai, siunčiami iš komandinės eilutės. Pokalbių programos, kurios viską pažymi kritiniu, kad prisiverstų prieš jus, neatitinka reikalavimų.

### Tuščioji eiga

LunaOS apvalkalas valdo tuščiosios eigos elgseną, o laikai yra aukščiausio lygio `idle` blokas `~/.config/lunaos/shell.json`:

```json
{
  "version": 1,
  "idle": {
    "screensaver": 150,
    "lock": 300
  }
}
```

Abu skaičiai yra sekundės, skaičiuojamos nuo momento, kai nuėjote į tuščiąją eigą — ne vienas nuo kito. Taigi su numatytaisiais nustatymais ekrano užsklanda pasirodo po dviejų su puse minučių, o užrakinimo ekranas perima po penkių minučių, nesvarbu, ar ekrano užsklanda veikė. Išsaugokite failą ir apvalkalas iš karto pritaikys naujus laikus.

Jei ekrano užsklandą atmesite prieš užrakinimo terminą, tai skaičiuojama kaip veikla ir laukiantis užraktas atšaukiamas. Jūs neužrakinami už tai, kad pažvelgėte į savo kompiuterį.

Norėdami visiškai sustabdyti užrakinimą tuščiojoje eigoje, `Super + Ctrl + I` — arba `lunaos toggle idle` — įjungia nemiegoti, o juostoje pasirodo kavos puodelio indikatorius. Tai tas, kurį reikia paspausti prieš ilgą pristatymą arba rinkinį, kurį norite stebėti. Paspauskite jį dar kartą, kad grįžtumėte į normalią būseną. `lunaos toggle idle status` atspausdina dabartinę būseną kaip JSON, jei jums jos reikia iš scenarijaus.

Tai apie užrakinimą ir ekrano užsklandą, ne apie maitinimą. Pristabdymas ir užmigdymas turi savo sąranką [sistemos miege](36-system-sleep.md).

### Ekrano užsklanda

LunaOS ekrano užsklanda yra ASCII menas, leidžiamas per atsitiktinius teksto efektus, po vieną egzempliorių kiekvienam monitoriui. Bet kuris klavišas arba pelės judesys ją užbaigia.

Galite ją paleisti pagal pareikalavimą iš _Sistema > Ekrano užsklanda_ (`Super + Esc`), kuri priverčia ją pasirodyti, net jei išjungėte tuščiosios eigos ekrano užsklandą. Pagal numatytuosius nustatymus su ja nėra susietas joks spartusis klavišas.

`lunaos toggle screensaver` išjungia tuščiosios eigos, jei mieliau norėtumėte eiti tiesiai iš darbo į užrakintą. Jam reikia terminalo, kurį jis žino, kaip konfigūruoti — Alacritty, Foot, Ghostty arba Kitty — ir jis jums tai pasakys, jei jūsų numatytasis terminalas yra kažkas kita.

Logotipą, kurį ji piešia, galite keisti po _Stilius > Ekrano užsklanda_. Įkelkite png arba svg ir LunaOS pavers jį ASCII. Žr. [prekės ženklą](41-branding.md).

### Užrakinimo ekranas

`Super + Ctrl + L` užrakina kompiuterį. Tai paleidžia užrakinimo ekraną iš LunaOS apvalkalo, užtušuoja ekraną, atstato jūsų klaviatūros išdėstymą į pirmąjį, kad nerašytumėte slaptažodžio neteisinga abėcėle, ir — jei jis veikia — išeidamas užrakina 1Password.

Užrakinimo ekranas priima slaptažodį ir taip pat priims piršto antspaudą, kai jį nustatysite. Tai ir kiti autentifikavimo būdai aptarti [aparatinės įrangos autentifikavime](37-hardware-authentication.md).
