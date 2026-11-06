# Egyéb csomagok

Az Arch fantasztikusan gazdag csomagkínálattal rendelkezik szinte bármilyen szoftverhez a hivatalos tároló és az Arch User Repository (AUR) között.

A használata ennél egyszerűbb már nem is lehetne. Új Arch csomagot úgy telepíthetsz, hogy az LunaOS menüben (`Super + Space`) az _Install > Package_ pontra mész, és beírod a kívánt csomagot. Automatikusan fuzzy szűréssel szűkíti az összes csomag listáját. (Ezt kézzel is elvégezheted a terminálban az `lunaos pkg add [package]` paranccsal.)

Ugyanezt megteheted az AUR-ral is, csak használd az _Install > AUR_ pontot. Csak ne feledd, hogy az AUR-t nem ellenőrzi az Arch csapata. Olyan, mint a RubyGems vagy az npm. Bárki feltölthet.

Ha el szeretnél távolítani egy csomagot, használhatod az LunaOS menü _Remove > Package_ pontját. Ez eltávolítja a csomagot, a konfigurációs fájlokat és a függőségeket. (Ezt kézzel is elvégezheted az `lunaos pkg drop [package]` paranccsal.)
