# Risoluzione dei problemi

### Ho rotto il sistema con un aggiornamento!

Prova prima a [ripristinare il sistema](47-system-snapshots.md) alla versione precedente al tuo aggiornamento recente. Se non funziona, usa `lunaos-debug` per condividere il tuo problema su #lunaos-help nel Discord. E se tutto questo fallisce, puoi reinstallare configurazioni e pacchetti predefiniti usando `lunaos-reinstall`.

### Perché alcune app sono così grandi sul mio schermo?

LunaOS presuppone uno schermo ad alta risoluzione 2x, il che richiede di impostare `GDK_SCALE` a 2 in `~/.config/hypr/monitors.lua`. Ma se hai uno schermo 1x, puoi cambiare `local lunaos_gdk_scale = 2` in 1 (e poi riavviare qualsiasi app sovradimensionata). Vedi [il manuale sui monitor](33-monitors.md).

Per Spotify, puoi usare `Ctrl + Minus` per rimpicciolire l'interfaccia (e `Ctrl + Plus` per ingrandirla).

### Wi-Fi, Bluetooth, audio o trackpad hanno smesso di funzionare

Prima di riavviare, prova a riavviare da solo il sottosistema incriminato. _Aggiorna > Hardware_ nel menu di LunaOS ha Wi-Fi, Bluetooth, Audio e Trackpad, e ricaricare uno di questi risolve la maggior parte delle situazioni "cinque minuti fa funzionava": cuffie Bluetooth che non si ricollegano, un trackpad morto dopo una sospensione, l'audio sparito quando hai scollegato un monitor.

### Perché i miei altoparlanti esterni non suonano?

Probabilmente perché non sono impostati come uscita principale. Fai clic sull'icona dell'altoparlante sul lato destro della barra e si aprirà il popup del volume, dove puoi scegliere il dispositivo di uscita (e anche regolare i volumi per singola app).

### Gli altoparlanti del mio portatile suonano male

Su alcuni portatili, LunaOS applica automaticamente una regolazione degli altoparlanti che corregge la risposta in frequenza degli altoparlanti integrati. `lunaos audio tuning status` ti dice se ce n'è una attiva sulla tua macchina e `lunaos audio tuning off` la disattiva se preferisci sentire gli altoparlanti al naturale.

### Perché non riesco ad accedere o usare sudo con la mia password?

Probabilmente l'hai digitata male troppe volte e sei rimasto bloccato fuori. Se sta succedendo nella schermata di blocco, puoi premere `CTRL + ALT + F2` per avviare una nuova TTY in cui accedere come root, poi eseguire `faillock --reset --user [your-username]`. Questo azzererà il blocco e sarai a posto.

### Perché le richieste di autorizzazione di 1Password per 1Password SSH Agent / CLI non compaiono?

Può succedere per 2 motivi:

Affinché compaia la richiesta di approvazione completa, Settings > Advanced > Use Hardware Acceleration deve essere attivato. _Nota: serve un riavvio perché inizi a funzionare._

 ![troubleshooting-1password](images/troubleshooting-1password.webp)

Oppure, se non hai avviato 1Password da quando hai acceso il computer, la richiesta non comparirà.
