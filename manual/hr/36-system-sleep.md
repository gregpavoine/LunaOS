# Spavanje sustava

LunaOS prema zadanim postavkama omogućuje obustavu i hibernaciju, ali ako imate problema s bilo čime od toga na svojem računalu, možete ih isključiti.

### Profili napajanja

Na prijenosnom računalu LunaOS pamti vaš profil napajanja odvojeno za priključeno napajanje i rad na bateriji te se prebacuje između njih dok priključujete i isključujete. Odmah po instalaciji to znači performanse na struji i uravnoteženo na bateriji.

Što vaše računalo nudi možete vidjeti s `lunaos powerprofiles list`, a onaj koji želite za trenutačno stanje postaviti s `lunaos powerprofiles set autodetect power-saver`. Za postavljanje drugog stanja bez isključivanja ičega, imenujte ga izravno: `lunaos powerprofiles set battery power-saver`. Što god odaberete, to ćete dobiti natrag sljedeći put kad budete u tom stanju.

### Prebacivanje obustave

Obustavu prebacujete pokretanjem `lunaos toggle suspend` iz terminala. Time se samo otkriva/sakriva opcija pod _Sustav_ (ili `Super + Esc`), a zatim možete vidjeti radi li dosljedno na vašem sustavu. Ako ne, možete je ponovno sakriti istom naredbom.

### Prebacivanje hibernacije

Hibernaciju postavljate pokretanjem `lunaos hibernation setup` iz terminala. Hibernacija stvara podvolumen /swap na vašem pogonu za pokretanje veličine vaše fizičke alokacije RAM-a, pa se pobrinite da imate dovoljno slobodnog prostora. Na računalu s 32GB uvijek će vam trebati 32GB+ slobodno za ovaj volumen. Hibernacija također zahtijeva zadani Limine bootloader.

Kad je postavljena, vidjet ćete opciju hibernacije pod _Sustav_ (ili `Super + Esc`), a zatim možete vidjeti radi li dosljedno na vašem sustavu. Ako ne, možete je ponovno ukloniti pokretanjem `lunaos hibernation remove`.
