# Schakelaars, idle & de schermbeveiliging

Veel van wat je dagelijks verandert is niet echt een instelling. Het is een modus die je een uur aanzet en weer uitzet: nachtlicht terwijl je laat werkt, niet storen terwijl je presenteert, wakker blijven terwijl je ergens naar kijkt. LunaOS noemt die schakelaars, en ze werken allemaal hetzelfde — een sneltoets, een menu-item en een commando, allemaal dezelfde schakelaar rakend.

### Het schakelmenu

`Super + Ctrl + O` opent _Activeren > Schakelen_ direct, of je kunt ernaartoe lopen vanuit het LunaOS-menu (`Super + Space`). Alles in die lijst is een schakelaar die je kunt omzetten zonder na te denken over waar de staat leeft. Schakelaars voor hoe het bureaublad eruitziet zitten ernaast, onder _Activeren > Weergave_.

Vanaf de terminal zijn dezelfde schakelaars `lunaos toggle <thing>`. Voer `lunaos toggle` op zichzelf uit om de hele groep te zien.

| Schakelaar | Sneltoets | Commando |
| ------ | ------ | ------- |
| Nachtlicht | `Super + Ctrl + N` | `lunaos toggle nightlight` |
| Meldingen dempen | `Super + Ctrl + ,` | `lunaos toggle notification silencing` |
| Wakker blijven (geen idle-vergrendeling) | `Super + Ctrl + I` | `lunaos toggle idle` |
| Crashregistratie | — | `lunaos toggle crash-capture` |
| Schermbeveiliging | — | `lunaos toggle screensaver` |
| [Herdr](https://herdr.dev)-themasync | — | `lunaos toggle theme sync` |
| Menubalk | `Super + Shift + Space` | `lunaos toggle bar` |
| Touchpad | `XF86TouchpadToggle` | `lunaos toggle touchpad` |
| Aanraakscherm | — | `lunaos toggle touchscreen` |
| Slaapstand | — | `lunaos toggle suspend` |
| Hybride GPU | — | `lunaos toggle hybrid gpu` |

De touchpad-, aanraakscherm- en hybride GPU-schakelaars leven onder _Activeren > Hardware_ (`Super + Ctrl + H`) in plaats van onder Schakelen, omdat ze alleen verschijnen wanneer je die hardware daadwerkelijk hebt. De touchpad- en aanraakscherm-versies overleven een Hyprland-herlaad — de naam van het uitgeschakelde apparaat wordt opgeslagen in een klein statusbestand dat Hyprland bij het opstarten leest om het opnieuw uit te schakelen.

_Activeren > Weergave_ bevat de menubalk en animaties, plus een paar dingen die geen `lunaos toggle`-commando's zijn maar zich hetzelfde gedragen: batterijpercentage in de balk, de standaard werkbladindeling, vensterafstand (`Super + Shift + Backspace`) en de 1-venster vierkante aspectverhouding (`Super + Ctrl + Backspace`).

De meeste hiervan zijn gewoon een vlagbestand onder `~/.local/state/lunaos/toggles/`. Als je er in een script op wilt branchen, geeft `lunaos-toggle-enabled` je een exitcode in plaats van dat je moet zoeken:

```bash
lunaos-toggle-enabled screensaver-off && echo "screensaver is off"
```

De vlaggen zijn vernoemd naar de uit-staat — `screensaver-off`, `suspend-off`, `bar-off` — dus hun aanwezigheid betekent dat de functie is uitgeschakeld.

### Indicatoren in de balk

Wanneer een modus aan staat, krijg je een kleine glyph in het midden van de bovenbalk naast de klok. Dat is de indicatoren-widget, en hij draagt dicteren, schermopname, openstaande herinneringen, nachtlicht, niet storen en wakker blijven.

Inactieve indicatoren zijn verborgen. Zweef over het gebied eromheen en ze vervagen gedimd in beeld, zodat je er een kunt aanklikken om hem in te schakelen zonder zijn sneltoets te kennen. Op een actieve klikken schakelt hem weer uit. Als je ze liever allemaal altijd ziet, stel je `alwaysShow` in op `true` op het `lunaos.indicators`-item in `~/.config/lunaos/shell.json` — zie [de bovenbalk](05-the-top-bar.md) voor hoe balkwidgets worden geconfigureerd.

### Nachtlicht

`Super + Ctrl + N` verwarmt het scherm naar 4000K, en er opnieuw op drukken zet het terug naar 6500K. Het wordt aangestuurd door hyprsunset, dat de schakelaar voor je start als het nog niet draait.

Standaard doet hyprsunset helemaal niets met je scherm. `~/.config/hypr/hyprsunset.conf` wordt geleverd met een identiteitsprofiel precies zodat het scherm onaangeroerd blijft totdat je om warmte vraagt. Als je liever hebt dat het op de klok schakelt, vervang je dat door een tijdprofiel:

```
profile {
    time = 20:00
    temperature = 4000
}
```

Start vervolgens hyprsunset bij het inloggen door `o.launch_on_start("hyprsunset")` toe te voegen aan `~/.config/hypr/autostart.lua`. Het 4000K/6500K-paar dat door de schakelaar wordt gebruikt is vast, dus het configbestand is waar je naartoe gaat als je een andere temperatuur wilt.

### Niet storen

`Super + Ctrl + ,` dempt meldingen. Er verschijnen geen toasts terwijl het aan staat, en de doorgestreepte bel-indicator zit in de balk om je eraan te herinneren waarom het bureaublad stil is geworden.

Er gaat echter niets verloren. Een gedempte melding wordt direct in je meldingsgeschiedenis geschreven, wat precies het record is dat je wilt wanneer je terugkomt en je afvraagt wat je hebt gemist. Open hem met `Super + Shift + Alt + ,`. Zie [kennisgevingen](10-notices.md) voor de rest van het meldingsverhaal.

Twee soorten berichten komen nog steeds door: LunaOS's eigen bevestigingstoasts voor iets wat je net hebt gedaan ("Theme changed", "Screenshot saved"), en kritieke waarschuwingen die vanaf de commandoregel worden verzonden. Chat-apps die alles als kritiek markeren om zich voor je neus te forceren kwalificeren niet.

### Idle

De LunaOS-shell bezit idle-gedrag, en de timings zijn een `idle`-blok op het hoogste niveau in `~/.config/lunaos/shell.json`:

```json
{
  "version": 1,
  "idle": {
    "screensaver": 150,
    "lock": 300
  }
}
```

Beide nummers zijn seconden geteld vanaf het moment dat je idle ging — niet vanaf elkaar. Dus met de standaardwaarden komt de schermbeveiliging op na tweeënhalve minuut en neemt het vergrendelscherm het over bij vijf minuten, of de schermbeveiliging nu draaide of niet. Sla het bestand op en de shell pikt de nieuwe timings meteen op.

Als je de schermbeveiliging sluit vóór de vergrendelingsdeadline, telt dat als activiteit en wordt de openstaande vergrendeling geannuleerd. Je wordt niet buitengesloten omdat je naar je machine gluurde.

Om vergrendelen bij idle helemaal te stoppen, zet `Super + Ctrl + I` — of `lunaos toggle idle` — wakker blijven aan, en de koffiekop-indicator verschijnt in de balk. Dat is degene die je moet indrukken voor een lange presentatie of een build die je wilt bekijken. Druk er opnieuw op om terug te gaan naar normaal. `lunaos toggle idle status` print de huidige staat als JSON als je die nodig hebt vanuit een script.

Dit gaat over vergrendelen en de schermbeveiliging, niet over stroom. Slaapstand en sluimerstand hebben hun eigen setup in [systeemslaap](36-system-sleep.md).

### De schermbeveiliging

LunaOS's schermbeveiliging is ASCII-art die door willekeurige teksteffecten loopt, één instantie per monitor. Elke toets of muisbeweging sluit hem.

Je kunt hem op aanvraag starten vanuit _Systeem > Schermbeveiliging_ (`Super + Esc`), wat hem forceert zelfs als je de idle-schermbeveiliging hebt uitgeschakeld. Er is standaard geen sneltoets aan gebonden.

`lunaos toggle screensaver` is wat de idle-versie uitschakelt, als je liever direct van werken naar vergrendeld gaat. Het heeft een terminal nodig die het weet te configureren — Alacritty, Foot, Ghostty of Kitty — en zal je dat vertellen als je standaardterminal iets anders is.

Het logo dat het tekent is aan jou om te veranderen, onder _Stijl > Schermbeveiliging_. Upload een png of svg en LunaOS converteert hem naar ASCII. Zie [branding](41-branding.md).

### Het vergrendelscherm

`Super + Ctrl + L` vergrendelt de machine. Dat draait het vergrendelscherm vanuit de LunaOS-shell, maakt het scherm leeg, reset je toetsenbordindeling naar de eerste zodat je je wachtwoord niet in het verkeerde alfabet typt, en — als je het draait — vergrendelt 1Password op weg naar buiten.

Het vergrendelscherm accepteert een wachtwoord, en het accepteert ook een vingerafdruk zodra je er een hebt ingesteld. Dat, en de andere manieren om te authenticeren, worden behandeld in [hardware-authenticatie](37-hardware-authentication.md).
