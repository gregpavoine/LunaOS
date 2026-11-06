# Horná lišta

Pás na vrchu vašej obrazovky je lišta LunaOS. Nie je to pripevnená stavová lišta, ale súčasť shellu LunaOS, jediného dlhodobo bežiaceho procesu Quickshell, ktorý tiež kreslí menu, upozornenia, vyskakovacie okná OSD a uzamykaciu obrazovku. Preto sa perfektne ladí so všetkým ostatným a preto sa panel otvára okamžite namiesto spúšťania novej aplikácie.

Je to tiež jediný kus pracovnej plochy, ktorý je vždy na obrazovke, takže stojí za to vedieť, čo všetky tie malé glyfy robia.

## Čo je na nej predvolene

Lišta má tri sekcie. Naľavo sedí logo LunaOS (spúšťač menu) a indikátory pracovných plôch. V strede získate indikátory stavu, hodiny, Elsewhen (svetový čas), počasie a odznak aktualizácie LunaOS. Napravo: rozloženie klávesnice (keď je nakonfigurovaných viacero), systémová lišta, agenti, bluetooth, sieť, zvuk, displej a napájanie.

Niektoré z nich sa objavia, len keď majú čo povedať. Rozloženie klávesnice sa objaví, len ak ste nakonfigurovali viac než jedno rozloženie. Odznak aktualizácie sa objaví, len keď čaká aktualizácia LunaOS. A ikona agentov sa objaví pri prvom nájdení využívania AI kódovania na stroji (pozrite si [AI](17-ai.md)).

## Klikanie okolo

Takmer každý widget niečo robí na ľavé, pravé a stredné kliknutie a viaceré reagujú na rolovanie. Toto je časť, ktorú ľudia prehliadnu — v pravom a strednom tlačidle sa skrýva veľa dobrých vecí.

| Widget | Ľavé | Pravé | Stredné / rolovanie |
| --- | --- | --- | --- |
| Menu | Menu LunaOS | Nový terminál | — |
| Pracovné plochy | Zamerať tú pracovnú plochu | — | Rolovanie: ďalšia/predchádzajúca pracovná plocha |
| Hodiny | Vyskakovací kalendár | Cyklovať formát štítku | Stredné: výber časového pásma |
| Počasie | Vyskakovacia predpoveď | Úplné počasie ako upozornenie | Stredné: obnoviť |
| Zvuk | Panel zvuku | Stlmiť | Stredné: panel · rolovanie: hlasitosť |
| Mikrofón | Stlmiť mikrofón | — | Stredné: panel zvuku · rolovanie: vstupná hlasitosť |
| Sieť | Panel siete | — | — |
| Bluetooth | Panel Bluetooth | Prepnúť rádio | — |
| Displej | Panel displeja | — | Rolovanie: jas |
| Napájanie | Panel napájania | Prepnúť percento batérie | — |
| Médiá | Prehrať/pozastaviť | Vyskakovací obal | Stredné: ďalšia · rolovanie: predchádzajúca/ďalšia |
| Agenti | Panel agentov | Spustiť vášho agenta | Stredné: ďalšie predplatné |
| Lišta | Prejsť kurzorom na odhalenie zásuvky | Pravé na šípke na správu | — |
| Aktualizácia LunaOS | Spustiť aktualizáciu | — | — |

Nie všetko z tejto tabuľky je na vašej lište hneď po vybalení. Widget médií (MPRIS práve prehrávané, s rolujúcou skladbou a interpretom) a widget mikrofónu sú obe vstavané, ale predvolene vypnuté — pridajte si ich, ak ich chcete, ako je opísané nižšie.

## Panely

Kliknutie na ikonu lišty otvorí panel, ktorý je poriadnym vyskakovacím oknom s posúvačmi, zoznamami a navigáciou klávesnicou, a nie popisek. Každý má aj klávesovú skratku, takže nikdy nemusíte mieriť na 16-pixelový glyf:

| Skratka | Panel |
| --- | --- |
| `Super + Ctrl + A` | Zvuk |
| `Super + Ctrl + W` | Sieť |
| `Super + Ctrl + B` | Bluetooth |
| `Super + Ctrl + D` | Displej |
| `Super + Ctrl + P` | Napájanie |
| `Super + Ctrl + Alt + D` | Kalendár |
| `Super + Ctrl + Alt + E` | Svetový čas |
| `Super + Ctrl + 1-9` | Prepnúť n-tý panel v pravej sekcii |

Panely nie sú len výpisy. Práve tam tu vec skutočne robíte:

- **Zvuk** má hlavný posúvač hlasitosti, výber výstupného zariadenia a mixér pre jednotlivé aplikácie, takže môžete stíšiť tú jednu kartu prehliadača bez toho, aby ste sa dotkli všetkého ostatného.
- **Sieť** skenuje Wi-Fi, ukazuje silu signálu, pripája sa a umožňuje vám vybrať poskytovateľa DNS.
- **Bluetooth** zobrazuje vaše zariadenia s pripojením/odpojením a úrovňami batérie.
- **Napájanie** ukazuje štatistiky batérie, prepína profily napájania (pamätá si samostatnú voľbu pre batériu a AC) a zobrazuje niektoré informácie o systéme.
- **Displej** nesie posúvač jasu, veľkosť textu, predvoľby škálovania monitora a — keď máte viac než jednu obrazovku — ovládacie prvky pre jednotlivé monitory. Pozrite si [monitory](33-monitors.md) pre hlbší príbeh.
- **Hodiny** otvárajú mriežku mesiaca s číslami týždňov podľa ISO a krokovaním mesiacov.

Každý panel prijíma klávesnicu rovnako ako myš: šípky sa pohybujú, Return aktivuje, Tab prechádza na susedný panel a Escape zatvára.

`Super + Ctrl + 1-9` počíta panely zľava doprava v pravej sekcii, pričom preskakuje lištu, keďže nemá vlastný panel. Takže číslo zodpovedá ikone, na ktorú by ste ukázali.

### Tailscale a Dropbox

Ďalšie dva widgety sa na lište objavia, len keď nainštalujete zodpovedajúcu službu z **Inštalovať → Služba**, a oba stojí za to poznať, pretože robia viac než len hlásia stav.

Panel **Tailscale** pripája a odpája tailnet, prepína medzi účtami a vyberá výstupný uzol (v tomto zozname sa objavia vaše vlastné stroje aj regióny Mullvad). Prehliada aj vaše stroje — a keď je jeden vybraný, `s` mu odošle súbory cez Taildrop, čo je najrýchlejší spôsob, ako presunúť súbor do telefónu alebo iného notebooku. `c` kopíruje IP adresu stroja, `n` jeho názov a `d` jeho úplný názov DNS. Na každom riadku stroja je aj tlačidlo na odoslanie, ak radšej klikáte. To isté z terminálu je `lunaos tailscale send <machine> [file...]`.

Panel **Dropbox** spracúva prihlásenie, ukazuje, koľko úložiska ste použili, a zobrazuje nedávno synchronizované súbory.

Odstránenie ktorejkoľvek služby vezme jej widget späť z lišty.

## Indikátory

Malý klaster v strede je widget indikátorov. Sú to glyfy stavu pre režimy, ktoré ste zapli: nerušiť, nočné svetlo, [pripomienka](09-reminders.md) v poradí, aktívne nahrávanie obrazovky, zostať hore a [diktovanie](11-text-extraction-dictation.md). Rozsvietia sa, keď je režim aktívny, a inak zostávajú mimo cesty — prejdite kurzorom doľava od hodín (nad ne na zvislej lište), aby ste nahliadli na neaktívne. Kliknutie na indikátor prepne tento režim. Jeden ďalší je len stavový: vzdialená relácia svieti načerveno, kým je niekto pripojený k tomuto stroju cez gliff.

Ak by ste radšej mali vždy viditeľné, nastavte `alwaysShow` na `true` na widgete. A ak vám záleží len na niektorých z nich, uveďte tie, ktoré chcete, v `items`: `["Dnd", "Reminder", "NightLight"]`. Môžete mať viac než jeden widget indikátorov, takže rôzne sekcie môžu zobrazovať rôzne podmnožiny.

## Preusporadúvanie lišty

Lišta sa konfiguruje sama. Na presúvanie vecí nemusíte otvárať konfiguračný súbor.

Chyťte prázdne miesto lišty okolo stredu a potiahnite ho k inému okraju obrazovky a lišta sa tam presunie — ľavo, pravo, hore alebo dolu, všetko funguje a každý widget sa prispôsobí (zvislé lišty sa vrátia ku kompaktným formám len s ikonami). Kliknutie a podržanie spustí rovnaké ťahanie. Dvakrát kliknite ľavým tlačidlom na to isté prázdne miesto, aby ste cyklovali pozadie lišty z priehľadného na sklenené a na neprehľadné. A potiahnite ľubovoľný widget, aby ste ho preusporiadali alebo hodili do inej sekcie.

Widget, ktorý chcete mať poruke, ale nie stále na očiach, môže ísť do zásuvky systémovej lišty: potiahnite ho na systémovú lištu a zásuvka sa otvorí, aby ho prijala. Sedí tam za ikonami aplikácií, objaví sa vždy, keď sa objaví zásuvka, a kým je jeho panel otvorený, drží svoju zásuvku otvorenú. Potiahnite ho späť na lištu, aby ste ho vrátili na miesto.

Ak radšej vyberáte z menu, **Akcie → Vzhľad → Lišta menu** má polohu a **Pozadie lišty** vyberá, ako sa lišta kreslí: **Neprehľadná lišta**, **Sklenená lišta** (priesvitná, s tým, čo je za ňou, rozmazaným) alebo **Priehľadná lišta**.

Tie isté veci majú príkazy, čo je to, čo chcete pre nastavenie [dotfiles](31-dotfiles.md):

```bash
lunaos bar position bottom
lunaos bar background glassy  # opaque, glassy, none, or cycle
lunaos bar move lunaos.clock --section center --index 0
lunaos bar move lunaos.bluetooth --section tray  # into the tray's drawer
lunaos bar set lunaos.clock format "HH:mm"
lunaos bar defaults          # back to the shipped layout
```

Ak chcete widget úplne pridať alebo odstrániť, použite príkazy pluginov. `lunaos plugin list` vytlačí každý widget, ktorý shell pozná, s jeho id, a potom:

```bash
lunaos plugin enable lunaos.media --section center
lunaos plugin disable lunaos.weather
```

## Skrytie lišty

`Super + Shift + Space` prepína lištu vypnúť a zapnúť bez zabitia shellu — panely a klávesové skratky naďalej fungujú, len dostanete späť pixely. Je to aj v menu pod **Akcie → Vzhľad → Lišta menu → Skryť lištu**.

## Konfiguračný súbor

Všetko je uložené v `~/.config/lunaos/shell.json` pod kľúčom `bar`. Tu je skrátená verzia:

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

Každý widget je jednou položkou v jednom z troch polí rozloženia a jeho nastavenia sedia priamo na tej položke — neexistuje žiadny samostatný súbor nastavení a žiadny podobjekt `config`. Formát `format` hodín, `formatAlt` (na čo cykluje pravé kliknutie) a `verticalFormat` žijú práve tam na `{ "id": "lunaos.clock" }`.

`centerAnchor` menuje jeden centrálny widget, ktorý sa pripne na presný stred obrazovky, pričom ostatné ho flankujú. Tak zostanú hodiny mŕtvo v strede, aj keď počasie a odznak aktualizácie prichádzajú a odchádzajú. Nastavte ho na prázdny reťazec a centrálny zoznam sa namiesto toho len centruje ako skupina.

Jedno pravidlo, ktoré stojí za osvojenie: **keď raz máte vlastný `shell.json`, je kanonický**. Kým si nič neprispôsobíte, shell číta predvolený súbor LunaOS. V momente, keď potiahnete widget, spustíte `lunaos bar` alebo sami upravíte súbor, vlastníte ho — neexistuje žiadne hlboké zlúčenie, takže nové predvolené widgety v budúcich vydaniach LunaOS sa na vašej lište automaticky neobjavia. `lunaos bar defaults` vráti dodávané rozloženie, kedykoľvek chcete čistý štít.

Rovnaký súbor tiež drží vaše načasovania nečinnosti na najvyššej úrovni, mimo kľúča `bar`: `idle.screensaver` a `idle.lock`, obe v sekundách odvtedy, čo ste išli do nečinnosti. Takže predvolený šetrič obrazovky naskočí po 150 sekundách a zámok po 300.
