# Nuashonruithe

Coinnítear LunaOS agus do phacáistí cothrom le dáta trí _Nuashonraigh > LunaOS_ sa roghchlár LunaOS (`Super + Space`).

Suiteáiltear LunaOS féin mar phacáistí rialta pacman ó [Stóras Pacáistí LunaOS](https://github.com/omacom-io/omarchy-pkgs), mar sin suiteálann nuashonrú [an t-eisiúint LunaOS is déanaí](https://github.com/gregpavoine/LunaOS/releases), ritheann sé aon imircí ar feitheamh chun do chóras a chur i sioncron leis an gceann is déanaí, agus nuashonraíonn sé pacáistí córais go léir ó [Scáthán Arch LunaOS](https://github.com/omacom-io/omarchy-mirror) agus [AUR](https://aur.archlinux.org/) (má tá aon phacáistí AUR suiteáilte agat).

Nuair a dhéantar eisiúintí nua, taispeánfar deilbhín saighead chiorcail ar thaobh deas do chloig. Cliceáil air agus tosóidh an próiseas nuashonraithe.

![update-available](images/update-available.webp)

### Ceithre bhealach

Nuashonraítear LunaOS feadh ceithre bhealach: stable, RC, edge, agus dev. Tosaíonn suiteálacha nua ar an mbealach cobhsaí, a rianaíonn na [heisiúintí oifigiúla](https://github.com/gregpavoine/LunaOS/releases/), chomh maith le [scáthán Arch LunaOS cobhsaí](https://github.com/omacom-io/omarchy-mirror) atá ag rith mí amháin chun deiridh ar an gceann is déanaí, ionas gur féidir linn aon neamh-chomhoiriúnachtaí nua a éilíonn athruithe cumraíochta a ghabháil sula gcuirfidh siad fadhbanna ar dhaoine.

Ach más mian leat cabhrú leis na fadhbanna féideartha sin a aimsiú, is féidir leat rith ar bhealach edge. Coinneoidh sé sin do phacáistí LunaOS ag rianú na leaganacha forbartha is déanaí, agus ligeann duit nuashonrú chuig na pacáistí Arch is déanaí a luaithe atá siad ar fáil. Níor chóir duit é seo a dhéanamh ach má tá taithí agat le Linux, agus má tá a fhios agat conas córas a bhfuil fadhbanna aige a aisghabháil.

Roimh aon phríomh-eisiúint nua, beimid ag déanamh bailíochtú deiridh ag úsáid an bhealaigh RC. Má tá suim agat cabhrú leis an snasú deiridh, tar croch amach i #lunaos-release-candidates ar an Discord.

Mar fhocal scoir, tá an bealach dev ann, a nascann LunaOS go díreach le seiceáil amach git den chód foinse in `~/lunaos`, in éineacht le pacáistí edge. Níor chóir duit an bealach seo a úsáid ach más úsáideoir Linux le taithí tú, ag obair go díreach ar LunaOS, agus toilteanach briseadh a fhulaingt.

Is féidir leat aistriú idir bealaí trí _Nuashonraigh > Bealach_ a úsáid ón roghchlár LunaOS (nó `lunaos-channel-set` sa teirminéal).

### Nuashonruithe firmearta

Ní hé amháin do phacáistí a théann as dáta. Seolann go leor ríomhairí glúine agus forimeallacha BIOS, SSD, agus firmearta duga trí Sheirbhís Firmearta Soláthraí Linux, agus gheobhaidh _Nuashonraigh > Firmearta_ sa roghchlár LunaOS agus suiteálfaidh sé cibé rud atá ag fanacht ag do chrua-earraí. Suiteálann sé `fwupd` an chéad uair a rithann tú é. Ní féidir a lán firmearta a scríobh ach le linn atosaigh, mar sin ná bíodh iontas ort má iarrtar ceann ort.

### Rabhadh faoi nuashonruithe díreacha pacman/yay

Má tá tú eolach ar Arch cheana, b'fhéidir go dteastaíonn uait `pacman -Syu` nó `yay -Syu` a rith leat féin, ach má dhéanann tú é sin, caillfidh tú an seat, na himircí, agus na nuashonruithe cumraíochta a ritheann LunaOS mar aon le pacáistí nua. Sin an fáth go stopfaidh LunaOS uasghrádú díreach córais i ndáiríre agus go ndíreoidh sé tú chuig `lunaos update` ina ionad. (Má tá a fhios agat i ndáiríre cad atá á dhéanamh agat, inseoidh an garda duit conas é a sheachaint d'idirbheart amháin.)

### Nuashonruithe dona a aischur

Má bhíonn fadhb agat riamh tar éis nuashonrú a dhéanamh, is féidir leat do chóras a aischur chuig an seat a tógadh roimh an nuashonrú. Atosaigh agus pioc an seat sa roghchlár lódála tosaithe ón am sular thosaigh tú an nuashonrú.

![bootloader](images/bootloader.webp)

Má tá do chomhaid chumraíochta truaillithe ar bhealach éigin, is féidir leat athshuiteáil LunaOS a dhéanamh freisin trí `lunaos reinstall` a úsáid sa teirminéal. Athshuiteálfaidh sé seo na pacáistí réamhshocraithe LunaOS go léir, cuirfidh sé tú ar stable agus ísleoidh sé aon phacáistí atá ró-nua, agus athshocróidh sé na comhaid chumraíochta go léir. Tabhair faoi deara go bhforscríobhfar do athruithe cumraíochta úsáideora go léir ar réamhshocruithe LunaOS agus seo á dhéanamh!
