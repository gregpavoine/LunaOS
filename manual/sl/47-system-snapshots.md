# Posnetki sistema

Posnetke ustvarimo samodejno ob vsaki posodobitvi LunaOSja, če pa želite ustvariti svojega, lahko uporabite `lunaos-snapshot create`.

Za zagon in obnovitev posnetka ga izberete iz zagonskega nalagalnika Limine. (Če trenutno zaganjate naravnost v zaslon za dešifriranje LunaOSja, boste morali najprej prek BIOS-a izbrati Limine kot zagonsko možnost).

Na tem zaslonu izberite posnetek, v katerega želite zagnati, na podlagi datuma in različice. Različico LunaOSja v času posnetka lahko vidite v spodnjem levem kotu.

 ![snapshots-bootloader](images/snapshots-bootloader.webp)

Ko prispete noter, se prikaže obvestilo, ki vas obvešča, da ste v zagonskem posnetku, in če ga kliknete, zažene postopek obnovitve. Druga možnost je, da uporabite `lunaos-snapshot restore`.

 ![snapshots-restore](images/snapshots-restore.webp)

To bo obnovilo vaš korenski datotečni sistem, ne pa vašega `/home`. Deluje torej za povrnitev pokvarjene sistemske posodobitve, ne pa za obnovitev izgubljenih osebnih datotek.

To tudi pomeni, da vaš imenik `~/.config` ostane tak, kot je. Če torej povračate na starejšo različico knjižnice ali aplikacije, ki shranjuje konfiguracijske datoteke v novi obliki, boste morali to urediti ročno.

_Opomba: Ta funkcija je na voljo samo na namestitvah, ki uporabljajo zagonski nalagalnik Limine, ki je privzeti od LunaOS 2.0. Ni na voljo, če ste na GRUB ali systemd-boot._

### Preskakovanje zagonskega menija

Če se zagonskega menija nikoli ne dotaknete in želite samo, da računalnik gre naravnost na zaslon za dešifriranje, v meniju LunaOS zaženite _Nastavitve > Direct Boot_. To doda vnos EFI, ki kaže neposredno na LunaOS, tako da ga vdelana programska oprema zažene, ne da bi se ustavila pri Limine.

Kompromis je tisti, omenjen na vrhu: z vklopljenim neposrednim zagonom pridobitev posnetka pomeni, da najprej izberete Limine iz zagonskega menija BIOS-a. Znova zaženite _Nastavitve > Direct Boot_, da odstranite vnos in se vrnete na zagon prek Limine. Nekatera vdelana programska oprema ne prenaša rad vnosov EFI po meri, zato namestitev noče delovati na vdelani programski opremi American Megatrends in Apple.
