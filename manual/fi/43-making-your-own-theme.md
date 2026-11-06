# Oman teeman tekeminen

Voit lisätä omia teemojasi kansioon `~/.config/lunaos/themes`. Kopioi vain yksi olemassa olevista pohjaksi (katso kansiota `/usr/share/lunaos/themes`) ja viilaa sitten mielesi mukaan. Niin kauan kuin teemasi on kyseisessä kansiossa, se sisältyy teemanvalintavalikkoon.

Tärkein viilattava tiedosto on `colors.toml`. Se määrittelee värisarjan, jota käytetään sitten määritysten luomiseen terminaalille (Foot/Alacritty/Ghostty/Kitty), btopille, Chromiumille, Hyprlandille, Neovimille, Helixille, VS Codelle, Obsidianille ja koko LunaOS-käyttöliittymälle (yläpalkki, valikko, ilmoitukset, OSD ja lukitusruutu).

Voit myös käyttää mukana tulevaa Aether-sovellusta uuden teeman luomiseen ihanalla graafisella käyttöliittymällä värien kanssa leikkimiseen ja taustojen etsimiseen. Käynnistä se vain sovellusruudukosta näppäimillä `Super + Alt + Space`.

### Mitä asennettu teema voi sisältää

Itse kansioon `~/.config/lunaos/themes` kirjoittamasi teema voi sisältää mitä haluat — se on sinun koneesi ja sinun tiedostosi, ja LunaOS soveltaa kaikkea sitä.

Teema, jonka asennat jonkun toisen reposta komennolla `lunaos theme install`, säilyttää kaiken, mikä on väriä, ja menettää kourallisen tiedostoja, jotka ajaisivat koodia koneellasi: kaikki `.lua`-tiedostot, terminaalimääritykset (`alacritty.toml`, `foot.ini`, `ghostty.conf`, `kitty.conf`) ja tiedoston `vscode.json`. Teeman `hyprland.lua` on Luaa, jota sommittelijasi ajaa kirjauduttaessa, terminaalimääritys nimeää ohjelman, jonka terminaalisi käynnistää, ja `vscode.json` nimeää asennettavan VS Code -laajennuksen. Jonkun teeman asentamisen pitäisi muuttaa sitä, miltä työpöytäsi näyttää, ei koskaan sitä, mitä se ajaa.

Kaikki muu toimii edelleen täsmälleen niin kuin teeman tekijä kirjoitti sen — `btop.theme`, `chromium.theme`, `helix.toml`, `icons.theme`, `shell.toml`, taustat ja esikatselut säilyvät kaikki. Vain se, mikä pudotettiin, luodaan uudelleen tiedostosta `colors.toml` koneellasi.

LunaOS erottaa nämä kaksi siitä, onko teemalla oma git-reponsa sisällään, minkä `lunaos theme install` jättää jälkeensä kloonatessaan. Joten itse kirjoittamasi teema pysyy sinun, ja internetistä haettu pysyy väreinä.

### Vaalea tila

Jos teet vaalean tilan teemaa, aseta `mode = "light"` tiedostosi `colors.toml` alkuun. Sitten se paritetaan automaattisesti vaalean tilan kanssa kaikille sovelluksille. (Vanha tapa pudottaa tyhjä tiedosto nimeltä `light.mode` teemasi juureen toimii myös edelleen.)

### Kuvakevärit

Jos haluat väriyhteensovittaa tiedostonhallinnan kuvakkeet teemaasi, lisää tiedosto nimeltä `icons.theme`, jossa on haluamasi kuvakesarjan nimi. Oletuksena vaihtoehdot ovat: `Yaru Yaru-blue Yaru-dark Yaru-magenta Yaru-olive Yaru-prussiangreen Yaru-purple Yaru-red Yaru-sage Yaru-wartybrown Yaru-yellow`.

### Lukituksen avauskuva

Teemat, jotka toimitetaan kuvien `unlock.png` ja `preview-unlock.png` kanssa, luetellaan kohdassa _Tyyli > Avaa lukitus_. Tiedostosi `unlock.png` olisi mieluiten läpinäkyvä png. Ja voit luoda esikatselukuvan komennolla `lunaos plymouth preview`.

### Sellaisten sovellusten teemoitus, joita LunaOS ei kata

Jos käytät sovellusta, joka ei ole kyseisessä luettelossa, voit opettaa LunaOSn teemoittamaan sen itse mallipohjan avulla. Pudota tiedosto kansioon `~/.config/lunaos/themed/` nimettynä sen luoman määrityksen mukaan plus `.tpl`-päätteellä ja kirjoita määritys käyttäen paikkamerkkejä `{{ background }}`, `{{ foreground }}`, `{{ accent }}`, `{{ red }}`, `{{ color0 }}` - `{{ color15 }}` ja muuta palettia. Joka kerta kun vaihdat teemaa, tiedosto luodaan uudelleen kyseisen teeman väreillä.

Kansiossa on täysin kommentoitu `alacritty.toml.tpl.sample`, josta kopioida — se luettelee jokaisen käyttämäsi muuttujan sekä muuntimet `_strip` ja `_rgb` sovelluksille, jotka haluavat värinsä ilman `#`-merkkiä tai desimaali-RGB:nä. Mallipohjillasi on etusija LunaOSn omiin nähden, joten voit käyttää tätä myös ohittamaan, miten sisäänrakennettu sovellus teemoitetaan.

### Teemasi jakelu

Jos haluat jakaa teemaasi, jotta muut voivat käyttää sitä, sinun täytyy laittaa se julkiselle git-palvelimelle, kuten GitHubiin. Sitten ihmiset voivat asentaa sen LunaOS-valikosta kohdasta _Asenna > Tyyli > Teema_ käyttäen kyseistä URL-osoitetta. On suositeltavaa noudattaa nimeämiskäytäntöä `lunaos-[themename]-theme`, koska teema näkyy oikein pelkkänä `[themename]` teemanvalintavalikossa asennuksen jälkeen.

Jäljelle jäävästä `[themename]`-osasta tulee teeman hakemiston nimi, joten sen täytyy olla sellainen, jota LunaOS voi käsitellä turvallisesti: sen täytyy alkaa kirjaimella, numerolla tai alaviivalla, ja loppuosa voi sisältää kirjaimia, numeroita, merkkejä `.`, `_`, `+` ja `-`. Isot kirjaimet muutetaan pieniksi puolestasi, mutta kaikki muu — välilyönti, lainausmerkki, ei-englanninkielinen merkki — hylätään asennuksen yhteydessä sen sijaan, että siitä tehtäisiin hakemiston nimi. Joten `lunaos-tokyo-night-theme`, `lunaos-flexoki_light-theme` ja `lunaos-c++-theme` asentuvat kaikki hyvin.

Muista, että kun se asennetaan reposta, kaikki sen toimittamat `.lua`-tiedostot, terminaalimääritykset tai `vscode.json` pudotetaan, joten älä rakenna teemaa niiden varaan.

Voit saada teemasi lisättyä [lisäteemojen sivulle](https://omarchy.org/themes/) lähettämällä pull requestin [omarchy-site-repoon](https://github.com/omacom-io/omarchy-site).
