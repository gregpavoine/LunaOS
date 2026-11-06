# Asbhaint Téacs & Deachtú

### Asbhaint Téacs

Buail `Super + Ctrl + PrtScr` chun réigiún ar an scáileán a roghnú d'asbhaint téacs. Ansin tiontóidh samhail OCR foinse oscailte tesseract an roghnú sin ina théacs go tapa agus cuirfidh sí ar an ngearrthaisce é. Ansin ní gá duit ach `Super + V` a bhualadh chun greamú.

Tá sé seo an-chabhrach chun seoltaí a ghlacadh as buntásca íomhánna nó uimhreacha gutháin atá leabaithe i gceannlínte suíomh gréasáin.

 ![text-extraction](images/text-extraction.webp)

### Deachtú

Ar shuiteálacha úra, is é [Superwhisper](https://superwhisper.com/) an cúlra deachtúcháin réamhshocraithe, ag baint úsáide as próiseáil néil gan íoslódálacha samhla as líne. Roghnaigh Superwhisper nó [Voxtype](https://voxtype.io/) trí _Socrú > Réamhshocruithe > Deachtú_ i roghchlár LunaOS. Suiteálann roghnú cúil é más gá agus déanann sé an réamhshocrú é don deachtú. Lódálann Voxtype bunmhúnla Béarla a thógann 150MB; athraigh an tsamhail le `voxtype setup model` agus socruithe eile trí `~/.config/voxtype/config.toml`. Osclaíonn Superwhisper a shocruithe chun próiseáil néil nó áitiúil a roghnú, agus coinníonn OPR an feidhmchlár cothrom le dáta.

Deachtaíonn tú trí `Right Alt` nó `F9` a choinneáil síos, nó trí scoránú le `Super + Ctrl + X`, agus beidh an téacs deachtaithe le feiceáil sa limistéar ionchuir dírithe. Úsáideann na haicearraí seo an cúl roghnaithe. Tosaíonn Right Alt deachtú freisin nuair a úsáidtear é i gcorda modhnóra; úsáid Left Alt do na cordaí sin. Ar leaganacha méarchláir ina bhfuil Right Alt ina AltGr, úsáid `F9` chun AltGr a choinneáil ar fáil don chlóscríobh.

Oibríonn na haicearraí céanna le ceachtar cúl. Roghnaigh do chúl trí Socrú > Réamhshocruithe > Deachtú. Cumraíonn suiteálaí an chúil é agus sábhálann sé an roghnú. Rith `lunaos default dictation` chun a fheiceáil cé acu ceann atá roghnaithe. Sábháiltear an roghnú i `~/.config/lunaos/defaults/dictation`. Roghnaíonn suiteálacha úra Superwhisper go huathoibríoch.

Lódálann LunaOS comhtháthú deisce an chúil roghnaithe go huathoibríoch. Níl cumraíocht Superwhisper ag teastáil i `~/.config/hypr/`.

Is féidir le ceangail saincheaptha agus scripteanna `lunaos dictation start`, `lunaos dictation stop`, agus `lunaos dictation toggle` a úsáid. Láimhseálann an cúl roghnaithe an taifeadadh agus an trascríobh. Díchumasaíonn roghnú Superwhisper a aicearra coinneála dhúchasaí agus socraíonn sé a scorán dúchasach go `Alt + Space`, á choinneáil ar leithligh ó eochracha taifeadta LunaOS. Teastaíonn aicearra taifeadta dúchasach amháin ar a laghad ó Superwhisper, agus fanann a aicearra cealaithe ar fáil.

Is féidir le cúil bhreise comhtháthú gan LunaOS a athrú. Soláthair inrite darb ainm `lunaos-dictation-<backend>` ar PATH a ghlacann le `start`, `stop`, agus `toggle` agus a fhilleann go pras, le stádas neamhnialasach ar theip. Cumraíonn a shuiteálaí an soláthraí agus glaonn sé `lunaos-dictation-use <backend>` tar éis socrú rathúil. Stopann sé seo an cúl roimhe agus sábhálann sé an roghnú go hadamhach; caithfidh `stop` éirí leis nuair a bhíonn an soláthraí díomhaoin nó ar iarraidh. Úsáideann ainmneacha cúil litreacha beaga, digití, agus fleiscíní. Soláthair `lunaos-install-dictation-<backend>` don socrú; seolann `lunaos default dictation <backend>` é. Cuir iontráil roghchláir Réamhshocruithe leis trí shíneadh roghchláir LunaOS chun an roghnóir sin a ghlaoch. Is féidir le comhtháthú deisce roghnach a bheith i `${XDG_CONFIG_HOME:-$HOME/.config}/<backend>/shortcuts.lua`.
