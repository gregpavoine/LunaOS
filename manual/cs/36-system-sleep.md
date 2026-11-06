# Spánek systému

LunaOS ve výchozím nastavení povoluje uspání i hibernaci, ale pokud vám na vašem počítači některá z nich dělá potíže, můžete je vypnout.

### Profily napájení

U notebooku si LunaOS pamatuje profil napájení zvlášť pro provoz ze sítě a pro provoz z baterie a při zapojování a odpojování napájení mezi oběma přepíná. Hned po instalaci to znamená výkon při provozu ze sítě a vyvážený režim při provozu z baterie.

Nabídku svého počítače si můžete prohlédnout příkazem `lunaos powerprofiles list` a požadovaný profil pro stav, ve kterém právě jste, nastavit příkazem `lunaos powerprofiles set autodetect power-saver`. Chcete-li nastavit druhý stav, aniž byste něco odpojovali, pojmenujte ho přímo: `lunaos powerprofiles set battery power-saver`. Co si vyberete, to dostanete zpět, až se příště v daném stavu ocitnete.

### Přepnutí uspání

Uspání přepnete spuštěním příkazu `lunaos toggle suspend` z terminálu. Tím se pouze zobrazí/skryje volba v části _Systém_ (nebo po stisku `Super + Esc`), takže pak můžete vyzkoušet, zda na vašem systému funguje spolehlivě. Pokud ne, stejným příkazem ji opět skryjete.

### Přepnutí hibernace

Hibernaci nastavíte spuštěním příkazu `lunaos hibernation setup` z terminálu. Hibernace vytvoří na spouštěcí jednotce podsvazek /swap o velikosti přidělené fyzické paměti RAM, takže si ověřte, že máte dostatek volného místa. U počítače s 32 GB budete pro tento svazek vždy potřebovat více než 32 GB volného místa. Hibernace vyžaduje také výchozí zavaděč Limine.

Po nastavení uvidíte volbu hibernace v části _Systém_ (nebo po stisku `Super + Esc`), takže pak můžete vyzkoušet, zda na vašem systému funguje spolehlivě. Pokud ne, odstraníte ji opět spuštěním příkazu `lunaos hibernation remove`.
