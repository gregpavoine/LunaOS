# Actualizări

LunaOS și pachetele tale sunt ținute la zi prin _Actualizează > LunaOS_ în meniul LunaOS (`Super + Space`).

LunaOS însuși este instalat ca pachete pacman obișnuite din [depozitul de pachete LunaOS](https://github.com/omacom-io/omarchy-pkgs), așa că o actualizare instalează [cea mai recentă versiune LunaOS](https://github.com/gregpavoine/LunaOS/releases), rulează toate migrările în așteptare ca să-ți aducă sistemul în sincronizare cu cea mai recentă versiune și actualizează toate pachetele de sistem din [oglinda Arch a LunaOS](https://github.com/omacom-io/omarchy-mirror) și [AUR](https://aur.archlinux.org/) (dacă ai instalat pachete AUR).

Când sunt făcute versiuni noi, o pictogramă cu o săgeată în cerc va apărea în dreapta ceasului tău. Fă clic pe ea și procesul de actualizare va începe.

![update-available](images/update-available.webp)

### Patru canale

LunaOS este actualizat pe patru canale: stable, RC, edge și dev. Instalările noi pornesc pe canalul stable, care urmărește [versiunile oficiale](https://github.com/gregpavoine/LunaOS/releases/), precum și [oglinda Arch LunaOS stabilă](https://github.com/omacom-io/omarchy-mirror) care rulează cu o lună în spatele celei mai recente, astfel încât să putem prinde orice incompatibilități noi care cer schimbări de configurare înainte ca ele să provoace probleme oamenilor.

Dar dacă ai vrea să ajuți la depistarea acestor probleme potențiale, poți rula pe canalul edge. Asta îți va ține pachetele LunaOS urmărind cele mai recente versiuni de dezvoltare și îți permite să actualizezi la cele mai recente pachete Arch de îndată ce sunt disponibile. Ar trebui să faci asta numai dacă ai experiență cu Linux și știi cum să recuperezi un sistem care are probleme.

Înainte de orice versiune majoră nouă, vom face validarea finală folosind canalul RC. Dacă ești interesat să ajuți la finisarea finală, vino pe #lunaos-release-candidates pe Discord.

În sfârșit, există canalul dev, care leagă LunaOS direct de o copie git a codului sursă din `~/lunaos`, combinată cu pachetele edge. Ar trebui să folosești acest canal numai dacă ești un utilizator Linux experimentat, lucrezi direct la LunaOS și ești dispus să tolerezi stricăciuni.

Poți comuta între canale folosind _Actualizează > Channel_ din meniul LunaOS (sau `lunaos-channel-set` în terminal).

### Actualizări de firmware

Pachetele tale nu sunt singurul lucru care învechește. Multe laptopuri și periferice livrează firmware BIOS, SSD și dock prin Linux Vendor Firmware Service, iar _Actualizează > Firmware_ în meniul LunaOS va aduce și instala orice are hardware-ul tău în așteptare. Instalează `fwupd` prima dată când îl rulezi. Mult firmware poate fi scris numai în timpul unei reporniri, așa că nu te mira dacă ți se cere una.

### Avertisment despre actualizările directe cu pacman/yay

Dacă ești deja familiarizat cu Arch, ai putea fi tentat să rulezi pur și simplu `pacman -Syu` sau `yay -Syu` tu însuți, dar dacă faci asta, vei rata instantaneul, migrările și actualizările de configurare pe care LunaOS le rulează împreună cu pachetele noi. De aceea LunaOS va opri de fapt o actualizare directă de sistem și te va trimite în schimb la `lunaos update`. (Dacă știi cu adevărat ce faci, gardianul îți va spune cum să-l ocolești pentru o singură tranzacție.)

### Revenirea de la actualizări proaste

Dacă ai vreodată o problemă după ce ai făcut o actualizare, îți poți reveni sistemul la instantaneul luat înainte de actualizare. Repornește pur și simplu și alege în meniul de pornire instantaneul de dinainte să începi actualizarea.

![bootloader](images/bootloader.webp)

Dacă cumva fișierele tale de configurare au fost corupte, poți efectua și o reinstalare LunaOS folosind `lunaos reinstall` în terminal. Asta va reinstala toate pachetele LunaOS implicite, te va pune pe stable și va retrograda orice pachete care sunt prea noi și va reseta toate fișierele de configurare. Reține că toate schimbările tale de utilizator la valorile implicite LunaOS vor fi suprascrise făcând asta!
