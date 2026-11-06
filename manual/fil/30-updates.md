# Mga Update

Pinapanatiling up to date ang LunaOS at ang iyong mga package sa pamamagitan ng _Update > LunaOS_ sa LunaOS menu (`Super + Space`).

Ang LunaOS mismo ay naka-install bilang mga regular na pacman package mula sa [LunaOS Package Repository](https://github.com/omacom-io/omarchy-pkgs), kaya ang isang update ay nag-i-install ng [pinakabagong release ng LunaOS](https://github.com/gregpavoine/LunaOS/releases), nagpapatakbo ng anumang nakabinbing migration para i-sync ang sistema mo sa pinakabago, at nag-a-update ng lahat ng system package mula sa [LunaOS Arch Mirror](https://github.com/omacom-io/omarchy-mirror) at [AUR](https://aur.archlinux.org/) (kung nag-install ka ng anumang AUR package).

Kapag may mga bagong release, lalabas ang isang circle arrow icon sa kanan ng iyong orasan. I-click ito at magsisimula ang proseso ng update.

![update-available](images/update-available.webp)

### Apat na channel

Ina-update ang LunaOS sa apat na channel: stable, RC, edge, at dev. Nagsisimula ang mga bagong install sa stable channel, na sumusubaybay sa mga [opisyal na release](https://github.com/gregpavoine/LunaOS/releases/), pati na rin sa [stable LunaOS Arch mirror](https://github.com/omacom-io/omarchy-mirror) na tumatakbo nang isang buwan sa likod ng pinakabago, kaya mahuhuli namin ang anumang bagong incompatibility na nangangailangan ng mga pagbabago sa config bago ito magdulot ng problema sa mga tao.

Pero kung gusto mong tumulong na makita ang mga potensyal na isyung iyon, maaari kang tumakbo sa edge channel. Panatilihin niyon ang iyong mga LunaOS package na sumusubaybay sa pinakabagong development build, at nagbibigay-daan sa iyong mag-update sa pinakabagong mga Arch package sa sandaling available na sila. Dapat mo lang itong gawin kung may karanasan ka sa Linux, at alam mo kung paano i-recover ang isang sistemang may problema.

Bago ang anumang bagong major release, gagawa kami ng final validation gamit ang RC channel. Kung interesado kang tumulong sa final polishing, pumunta sa #lunaos-release-candidates sa Discord.

Sa wakas, mayroong dev channel, na nagli-link sa LunaOS nang direkta sa isang git checkout ng source code sa `~/lunaos`, kasama ang mga edge package. Dapat mo lang gamitin ang channel na ito kung isa kang bihasang Linux user, direktang nagtatrabaho sa LunaOS, at handang tiisin ang pagkasira.

Maaari kang lumipat sa pagitan ng mga channel gamit ang _Update > Channel_ mula sa LunaOS menu (o `lunaos-channel-set` sa terminal).

### Mga update ng firmware

Hindi lang ang iyong mga package ang naluluma. Maraming laptop at peripheral ang nagpapadala ng BIOS, SSD, at dock firmware sa pamamagitan ng Linux Vendor Firmware Service, at kukunin at i-i-install ng _Update > Firmware_ sa LunaOS menu ang anumang naghihintay para sa iyong hardware. Ini-install nito ang `fwupd` sa unang pagkakataong patakbuhin mo ito. Maraming firmware ang maisusulat lang sa panahon ng reboot, kaya huwag magulat kung hingan ka nito.

### Babala tungkol sa mga direktang update ng pacman/yay

Kung pamilyar ka na sa Arch, maaari kang matukso na patakbuhin na lang ang `pacman -Syu` o `yay -Syu` mismo, pero kung gagawin mo iyon, makakaligtaan mo ang snapshot, mga migration, at mga update sa configuration na pinapatakbo ng LunaOS kasama ng mga bagong package. Iyon ang dahilan kung bakit pipigilin talaga ng LunaOS ang isang direktang system upgrade at ituturo ka sa `lunaos update` sa halip. (Kung talagang alam mo ang ginagawa mo, sasabihin sa iyo ng guard kung paano ito lampasan para sa isang solong transaksyon.)

### Pag-rollback ng masasamang update

Kung magkakaroon ka ng problema pagkatapos gumawa ng update, maaari mong i-rollback ang sistema mo sa snapshot na kinuha bago ang update. I-restart lang at piliin ang snapshot sa boot loading menu mula bago mo sinimulan ang update.

![bootloader](images/bootloader.webp)

Kung sakaling nasira ang iyong mga configuration file, maaari ka ring magsagawa ng LunaOS reinstall gamit ang `lunaos reinstall` sa terminal. Muling i-i-install nito ang lahat ng default na LunaOS package, ilalagay ka sa stable at ida-downgrade ang anumang package na masyadong bago, at ire-reset ang lahat ng configuration file. Tandaan na ang lahat ng iyong user config change sa mga LunaOS default ay mapapatungan sa paggawa nito!
