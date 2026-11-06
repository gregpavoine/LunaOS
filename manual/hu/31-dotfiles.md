# Dotfile-ok

Az LunaOSt elsődlegesen az úgynevezett dotfile-okból konfigurálod, amelyek a `~/.config` mappában élnek. Ezek a saját módosításaidhoz a saját fájljaidnak tekintendők. A `/usr/share/lunaos` mappában élő fájlok magához az LunaOShoz tartoznak, és nem kellene belekotnod. Ha bármit módosítani szeretnél a `/usr/share/lunaos` mappában, helyette a `~/.config` mappában lévő értéket kell felülírnia.

A legfontosabb konfigurációk közvetlenül az LunaOS menüből (`Super + Space`) szerkeszthetők, mint például _Setup > Monitors_, _Setup > Keybindings_, _Setup > Input_ és _Setup > Config > [file]_. Amikor így szerkesztesz, minden folyamat, amelynek a konfiguráció módosítása után szüksége van az újraindításra, automatikusan újraindul, miután kilépsz a szerkesztőből (alapértelmezetten a Neovim — `:wq`, emlékszel! —, de ezt a _Setup > Defaults > Editor_ pontban módosíthatod).

Íme a legfontosabb fájlok listája a `~/.config` mappában és az, amiket vezérelnek:

| Fájl                  | Cél              |
| ----------------------- | --------------------- |
| `~/.config/hypr/hyprland.lua` | A fő Hyprland-konfiguráció. Betölti az LunaOS alapértelmezéseit és a saját felülíró fájljaidat. [Tudjon meg többet a Hyprland-konfigurációkról](https://wiki.hypr.land/Configuring/).  |
| `~/.config/hypr/bindings.lua` | A saját billentyűkiosztásaid és az alapértelmezéseket felülíró kiosztások. |
| `~/.config/hypr/monitors.lua` | A monitorokat, a felbontást és a pozíciót vezérli. |
| `~/.config/hypr/input.lua` | A billentyűzetkiosztást, az egeret és az érintőpad beállításait vezérli. |
| `~/.config/hypr/looknfeel.lua` | A térközöket, a szegélyeket, az animációkat és a megjelenés többi részét vezérli. |
| `~/.config/hypr/autostart.lua` | A munkamenettel együtt indított extra folyamatokat vezérli. |
| `~/.config/lunaos/shell.json` | Az LunaOS shellt vezérli: a sáv helyzetét, elrendezését és moduljait, valamint a képernyővédő, a zárolás és az üresjárat időzítését. |
| `~/.config/foot/foot.ini` | A terminálodat vezérli (a Foot az alapértelmezett). |
| `~/.XCompose` | Definiálja a gyorselérésű emojidat és a név/e-mail automatikus kiegészítésedet. Gondoskodj arról, hogy a módosítások után futtatod az `lunaos-restart-xcompose` parancsot. |

Ha végül sok módosítást végezsz a saját beállításaidon, jó ötlet biztonsági másolatot készíteni az összes dotfile-ról. [A Stow nagyszerű módja ennek](https://www.youtube.com/watch?v=NoFiYOqnC4o).

### Saját alkalmazások indítása a munkamenettel

Ha valami minden bejelentkezéskor futni szeretnél — egy szinkronizációs folyamatot, egy csevegőalkalmazást, a saját parancsszkriptedet — tedd a `~/.config/hypr/autostart.lua` fájlba:

```lua
o.launch_on_start("my-service")
```

Ez a parancsot a munkamenet részeként indítja, így megfelelően tisztánlódik, amikor ismét kijelentkezel.

### Szkriptek futtatása rendszereseményekhez

Az LunaOS néhány pillanatban hookokat lő ki, és saját szkriptjeidet aggathatod rájuk. Ezek a `~/.config/lunaos/hooks/<event>.d/` helyen találhatók, egy könyvtár eseményenként, és az ott lévő minden végrehajtható fájl fut, amikor az esemény bekövetkezik:

| Esemény | Mikor fut |
| ----- | ------------ |
| `post-boot` | Rögtön az asztal elindulása után |
| `post-update` | Az `lunaos update` vége felé, a csomagok, migrációk és szolgáltatás-újraindítások után, mielőtt a mise-eszközök frissülnének |
| `pre-refresh-pacman` | Miután az `lunaos refresh pacman` újraszinkronizálja a csomagkonfigurációt, mielőtt a csomagokat frissítené; egy csatornaváltás ugyanabban a frissítési lépésben futtatja |
| `theme-set` | Egy témaváltás után (a téma neve a `$1`-ben) |
| `font-set` | Betűtípus-váltás után (a betűtípus neve a `$1`-ben) |
| `battery-low` | Amikor az akkumulátor alacsonnyá válik (a százalék a `$1`-ben) |

A `pre-refresh-pacman` hook az, ahová az egyéni tárolók vagy az `IgnorePkg`-sorok tartoznak, mivel a csomagtranzakció előtt fut. Mindkét frissítéssel kapcsolatos hook a felhasználód nevében fut, miután az LunaOS törli a tárolt sudo-engedélyét, így egy olyan hooknak, amelyik `sudo`-t használ, saját engedélyre van szüksége, és kérheti a jelszavadat.

Ezen könyvtárak mindegyike már tartalmaz egy `.sample` fájlt, amely bemutatja, hogy milyen formaú egy hook — távolítsd el a `.sample` részt a névből, hogy használatba vedd. Egy máshol megírt szkript telepítéséhez használd az `lunaos hook install post-boot ~/my-hook` parancsot, amely indítási végrehajtási engedélyt állít be.

### Saját menüelemek hozzáadása

Az LunaOS menü (`Super + Space`) kiterjeszthető a saját soraival a `~/.config/lunaos/extensions/lunaos-menu.jsonc` szerkesztésével. A bejegyzések pontozott azonosítóval vannak felvéve, és az azonosító az, ami elhelyezi őket a fában, így a `personal` megjelenhet a gyökérmenüben, a `personal.notes` pedig azon belül jelenhet meg:

```jsonc
"personal": {"icon":"","label":"Personal"},
"personal.notes": {"icon":"󰎞","label":"Notes","action":"lunaos-launch-editor ~/notes"},
```

Ha meglévő azonosítót használsz újra, azt a sort váltod fel, ahelyett, hogy újat adnál hozzá. A fájl az összes elérhető mezőt megjegyzésekként dokumentálva érkezik.

### Saját shell-eksportok, függvények és aliaszok hozzáadása

Az LunaOS egy sor ergonómiai aliást és hasznos funkciót szállít, de gyakori, hogy sajátokat akarsz hozzáadni. Mind az aliasokat, funkciókat, mind az exportokat a `~/.bashrc` fájlba kell tenned. Ez a fájl nem íródik felül frissítéseken. Ha az LunaOS alapértelmezéseiből bármit meg szeretnél változtatni, itt is biztonságosan hozzáadhatod őket.

### Belső LunaOS-fájlok módosítása

Nézzük, ez az ön számítógépe. Bármit megtehet vele, de azt javaslom, ne módosítsd közvetlenül a `/usr/share/lunaos` mappában lévő fájlokat. Az LunaOS pacman-csomagjához tartoznak, így a módosításaidat egyszerűen felülírja a következő frissítés. Jobban jársz, ha a `~/.config/*` mappában írsz felül minden alapértelmezett értéket, amit nem szeretsz.

Szinte mindent megváltoztathatsz így, például az alapértelmezett billentyűkiosztásokat. Csak szerkeszd a `~/.config/hypr/bindings.lua` fájlt, hogy például az [Obsidiant](https://obsidian.md/) a [Joplinnal](https://joplinapp.org/) helyettesítsd (telepítse az `lunaos-pkg-add joplin-bin` paranccsal):

```lua
o.rebind("SUPER + SHIFT + O", "Joplin", "joplin-desktop")
```

Az `o.rebind` eltávolítja a meglévő kötést, mielőtt hozzáadná a helyettesítőjét. Ugyanazokat az argumentumokat veszi át, mint az `o.bind`, beleértve az indító segítőket és a kapcsolási beállításokat is. Használd az `o.bind` parancsot egy kötés hozzáadásához, vagy a `hl.unbind` parancsot, ha csak meg akarod szüntetni a helyettesítő beállítása nélkül.

Ha ragaszkodsz a belső LunaOS-fájlok feltöréséhez, válts a dev csatornára az _Update > Channel > Dev_ ponton keresztül. Ez az LunaOSt a forráskód egy git-checkoutjához köti a `~/lunaos` helyén, és szabadon megváltoztathatod a lelked szerint. Senki sem mondja meg, mit csinálj!

### Módosítások visszaállítása

Ha végül összezavarod a konfigurációkat, mindig visszaállíthatod azokat az alapértelmezésekre az LunaOS menüben lévő _Update > Config_ ponton keresztül. Vagy az `lunaos reinstall configs` parancs futtatásával minden visszaállítására.
