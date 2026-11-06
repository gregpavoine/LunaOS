# Sospensione del sistema

LunaOS attiva sospensione e ibernazione per impostazione predefinita, ma se hai problemi con una delle due sulla tua macchina, puoi disattivarle.

### Profili energetici

Su un portatile, LunaOS ricorda il tuo profilo energetico separatamente per quando è collegato alla corrente e quando va a batteria, e passa dall'uno all'altro quando colleghi e scolleghi l'alimentazione. Di serie significa prestazioni con la corrente e bilanciato a batteria.

Puoi vedere cosa offre la tua macchina con `lunaos powerprofiles list` e impostare quello che vuoi per lo stato in cui ti trovi con `lunaos powerprofiles set autodetect power-saver`. Per impostare l'altro stato senza scollegare nulla, nominalo direttamente: `lunaos powerprofiles set battery power-saver`. Qualunque cosa scegli, è quello che ritroverai la prossima volta che ti troverai in quello stato.

### Attivare/disattivare la sospensione

Attivi o disattivi la sospensione eseguendo `lunaos toggle suspend` dal terminale. Questo mostra o nasconde semplicemente l'opzione in _Sistema_ (o `Super + Esc`), e così puoi vedere se sul tuo sistema funziona in modo affidabile. In caso contrario, puoi nasconderla di nuovo con lo stesso comando.

### Attivare/disattivare l'ibernazione

Configuri l'ibernazione eseguendo `lunaos hibernation setup` dal terminale. L'ibernazione crea un sottovolume /swap sul disco di avvio grande quanto la tua RAM fisica, quindi assicurati di avere spazio in abbondanza. Su una macchina da 32GB, ti serviranno sempre più di 32GB liberi per questo volume. L'ibernazione richiede anche il bootloader Limine predefinito.

Una volta configurata, vedrai l'opzione di ibernazione in _Sistema_ (o `Super + Esc`), e così puoi vedere se sul tuo sistema funziona in modo affidabile. In caso contrario, puoi rimuoverla di nuovo eseguendo `lunaos hibernation remove`.
