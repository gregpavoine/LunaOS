# Dotfiler

LunaOS konfigureres primært gennem de såkaldte dotfiler, der lever i `~/.config`. De betragtes som dine filer til dine ændringer. Filerne, der lever i `/usr/share/lunaos`, tilhører selve LunaOS, og du bør ikke rode med dem. Hvis du har brug for at ændre noget i `/usr/share/lunaos`, bør du i stedet overskrive værdien i `~/.config`.

De vigtigste konfigurationer kan redigeres direkte fra LunaOS-menuen (`Super + Space`), som _Opsætning > Monitors_, _Opsætning > Keybindings_, _Opsætning > Input_ og _Opsætning > Config > [file]_. Når du gør det på denne måde, vil enhver proces, der skal genstartes efter konfigurationsredigeringer, automatisk blive det, efter du afslutter editoren (Neovim som standard — `:wq`, husk det! — men du kan ændre det via _Opsætning > Defaults > Editor_).

Her er en liste over de vigtigste filer i `~/.config`, og hvad de styrer:

| File                  | Purpose              |
| ----------------------- | --------------------- |
| `~/.config/hypr/hyprland.lua` | Den vigtigste Hyprland-konfiguration. Indlæser LunaOS-standarderne plus dine tilsidesættelsesfiler nedenfor. [Lær mere om Hyprland-konfigurationer](https://wiki.hypr.land/Configuring/).  |
| `~/.config/hypr/bindings.lua` | Dine egne tastaturgenveje og tilsidesættelser af standarderne. |
| `~/.config/hypr/monitors.lua` | Styrer dine skærme, opløsning og position. |
| `~/.config/hypr/input.lua` | Styrer dit tastaturlayout, mus og pegefelt-indstillinger. |
| `~/.config/hypr/looknfeel.lua` | Styrer mellemrum, kanter, animationer og resten af udseendet. |
| `~/.config/hypr/autostart.lua` | Styrer ekstra processer startet med sessionen. |
| `~/.config/lunaos/shell.json` | Styrer LunaOS-skallen: linjeposition, layout og widgets plus pausaskærm, lås og inaktivitetstidspunkter. |
| `~/.config/foot/foot.ini` | Styrer din terminal (foot er standarden). |
| `~/.XCompose` | Definerer dine hurtigadgangs-emojier og navn/e-mail-autofuldførelse. Sørg for at køre `lunaos-restart-xcompose` efter at have lavet ændringer. |

Hvis du ender med at lave mange ændringer for at tilpasse din egen opsætning, er det en god idé at sikkerhedskopiere alle disse dotfiler. [Stow er en fantastisk måde at gøre det på](https://www.youtube.com/watch?v=NoFiYOqnC4o).

### Starte dine egne apps med sessionen

Hvis du vil have noget til at køre, hver gang du logger ind — en synkroniseringsdaemon, en chatapp, dit eget script — så læg det i `~/.config/hypr/autostart.lua`:

```lua
o.launch_on_start("my-service")
```

Det starter kommandoen som en del af sessionen, så den bliver ryddet ordentligt op, når du logger ud igen.

### Køre scripts ved systemhændelser

LunaOS affyrer hooks på en håndfuld tidspunkter, og du kan hænge dine egne scripts på dem. De lever i `~/.config/lunaos/hooks/<event>.d/`, en mappe pr. hændelse, og hver eksekverbar fil derinde kører, når hændelsen sker:

| Event | When it runs |
| ----- | ------------ |
| `post-boot` | Lige efter skrivebordet er startet |
| `post-update` | Nær slutningen af `lunaos update`, efter pakker, migreringer og genstart af tjenester, før mise-værktøjer opdateres |
| `pre-refresh-pacman` | Efter `lunaos refresh pacman` gensynkroniserer pakkekonfigurationen, før den opdaterer pakker; et kanalskifte kører den under det samme genopfriskningstrin |
| `theme-set` | Efter et temaskift (temanavn i `$1`) |
| `font-set` | Efter et skrifttypeskift (skrifttypenavn i `$1`) |
| `battery-low` | Når batteriet bliver lavt (procentdel i `$1`) |

Hooken `pre-refresh-pacman` er, hvor brugerdefinerede repositories eller `IgnorePkg`-linjer hører hjemme, da den kører før pakketransaktionen. Begge opdateringsrelaterede hooks kører som din bruger, efter LunaOS rydder sin cachelagrede sudo-autorisation, så en hook, der bruger `sudo`, skal bruge sin egen autorisation og kan bede om din adgangskode.

Hver af de mapper indeholder allerede en `.sample`-fil, der viser formen på en hook — fjern `.sample` fra navnet for at sætte den i arbejde. For at installere et script, du har skrevet et andet sted, skal du bruge `lunaos hook install post-boot ~/my-hook`, som kopierer det ind og gør det eksekverbart.

### Tilføje dine egne menuposter

LunaOS-menuen (`Super + Space`) kan udvides med dine egne rækker ved at redigere `~/.config/lunaos/extensions/lunaos-menu.jsonc`. Poster er nøglet efter et prikket id, og id'et er det, der placerer dem i træet, så `personal` dukker op på rodmenuen, og `personal.notes` dukker op inde i den:

```jsonc
"personal": {"icon":"","label":"Personal"},
"personal.notes": {"icon":"󰎞","label":"Notes","action":"lunaos-launch-editor ~/notes"},
```

Genbrug et eksisterende id, og du tilsidesætter den række i stedet for at tilføje en ny. Filen leveres med alle de tilgængelige felter dokumenteret som kommentarer.

### Tilføje dine egne shell-eksporter, funktioner og aliasser

LunaOS leveres med en masse ergonomiske aliasser og nyttige funktioner, men det er meget almindeligt at ville tilføje dine egne. Du bør tilføje både aliasser, funktioner og eksporter i `~/.bashrc`. Denne fil vil ikke blive overskrevet ved opdateringer. Hvis du vil ændre nogen af LunaOS-standarderne, kan du også trygt tilføje dem her.

### Ændre interne LunaOS-filer

Hør her, dette er din computer. Du kan gøre, hvad du vil med den, men jeg vil råde mod at lave ændringer direkte i filerne i `/usr/share/lunaos`. De tilhører LunaOS pacman-pakken, så dine ændringer vil simpelthen blive overskrevet ved næste opdatering. Du har det bedre ved bare at overskrive alle standardværdier, du ikke kan lide, i mappen `~/.config/*` i stedet.

Du kan ændre stort set alt på den måde, som standardtastaturgenvejene. Rediger bare `~/.config/hypr/bindings.lua` for for eksempel at erstatte [Obsidian](https://obsidian.md/) med [Joplin](https://joplinapp.org/) (installer med `lunaos-pkg-add joplin-bin`):

```lua
o.rebind("SUPER + SHIFT + O", "Joplin", "joplin-desktop")
```

`o.rebind` fjerner den eksisterende binding, før den tilføjer sin erstatning. Den tager de samme argumenter som `o.bind`, inklusive starthjælpere og bindingsmuligheder. Brug `o.bind` til at tilføje en binding eller `hl.unbind` til at fjerne en uden at erstatte den.

Hvis du insisterer på at hacke på de interne LunaOS-filer, skal du skifte til dev-kanalen via _Opdater > Channel > Dev_. Det forbinder LunaOS til et git-checkout af kildekoden i `~/lunaos`, som du frit kan ændre til dit hjertes lyst. Der er ingen her til at fortælle dig, hvad du skal gøre!

### Nulstille alle ændringer

Hvis du ender med at lave et rod i konfigurationerne, kan du altid vende dem tilbage til standarderne via _Opdater > Config_ i LunaOS-menuen. Eller ved at køre `lunaos reinstall configs` for at nulstille alt.
