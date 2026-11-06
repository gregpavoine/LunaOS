# Yangilanishlar

LunaOS va paketlaringiz LunaOS menyusidagi (`Super + Space`) _Yangilash > LunaOS_ orqali yangilab turiladi.

LunaOS ning oʻzi [LunaOS Package Repository](https://github.com/omacom-io/omarchy-pkgs) dan oddiy pacman paketlari sifatida oʻrnatilgan, shuning uchun yangilash [eng soʻnggi LunaOS relizi](https://github.com/gregpavoine/LunaOS/releases) ni oʻrnatadi, tizimingizni eng soʻnggisi bilan sinxron qilish uchun kutilayotgan har qanday migratsiyalarni ishga tushiradi va [LunaOS Arch Mirror](https://github.com/omacom-io/omarchy-mirror) va [AUR](https://aur.archlinux.org/) dan (agar biror AUR paketlarini oʻrnatgan boʻlsangiz) barcha tizim paketlarini yangilaydi.

Yangi relizlar chiqarilganda, soatingizning oʻng tomonida doira strelka piktogrammasi paydo boʻladi. Uni bosing va yangilash jarayoni boshlanadi.

![update-available](images/update-available.webp)

### Toʻrtta kanal

LunaOS toʻrtta kanal boʻylab yangilanadi: stable, RC, edge va dev. Yangi oʻrnatishlar [rasmiy relizlar](https://github.com/gregpavoine/LunaOS/releases/) ni, shuningdek eng soʻnggisidan bir oy orqada ishlaydigan [barqaror LunaOS Arch oynasi](https://github.com/omacom-io/omarchy-mirror) ni kuzatuvchi stable kanalda boshlanadi, shunday qilib biz konfiguratsiya oʻzgarishlarini talab qiluvchi har qanday yangi mos kelmasliklarni ular odamlarga muammo tugʻdirishidan oldin tuta olamiz.

Lekin oʻsha mumkin boʻlgan muammolarni aniqlashga yordam bermoqchi boʻlsangiz, edge kanalda ishlashingiz mumkin. Bu LunaOS paketlaringizni eng soʻnggi dasturlash yigʻmalarini kuzatib turishga majbur qiladi va eng soʻnggi Arch paketlari mavjud boʻlishi bilanoq yangilash imkonini beradi. Buni faqat Linux bilan tajribangiz boʻlsa va muammolari bor tizimni qanday tiklashni bilsangiz qilishingiz kerak.

Har qanday yangi yirik relizdan oldin biz RC kanali yordamida yakuniy tekshiruv qilamiz. Agar yakuniy sayqallashga yordam berish bilan qiziqsangiz, Discord dagi #lunaos-release-candidates ga keling.

Nihoyat, dev kanali bor, u LunaOS ni toʻgʻridan-toʻgʻri `~/lunaos` dagi manba kodining git checkout iga, edge paketlari bilan birgalikda bogʻlaydi. Bu kanaldan faqat tajribali Linux foydalanuvchisi boʻlsangiz, toʻgʻridan-toʻgʻri LunaOS da ishlasangiz va buzilishga chidashga tayyor boʻlsangiz foydalanishingiz kerak.

Kanallar orasida LunaOS menyusidagi _Yangilash > Channel_ yordamida (yoki terminalda `lunaos-channel-set` bilan) almashishingiz mumkin.

### Proshivka yangilanishlari

Paketlaringiz eskiradigan yagona narsa emas. Koʻplab noutbuklar va periferik qurilmalar BIOS, SSD va dock proshivkasini Linux Vendor Firmware Service orqali yetkazadi va LunaOS menyusidagi _Yangilash > Firmware_ uskunangizda kutib turgan har qanday narsani olib kelib oʻrnatadi. U birinchi marta ishga tushirganingizda `fwupd` ni oʻrnatadi. Koʻp proshivka faqat qayta yuklash davomida yozilishi mumkin, shuning uchun bittasi soʻralishiga hayron boʻlmang.

### Toʻgʻridan-toʻgʻri pacman/yay yangilanishlari haqida ogohlantirish

Agar Arch bilan allaqachon tanish boʻlsangiz, shunchaki oʻzingiz `pacman -Syu` yoki `yay -Syu` ni ishga tushirish vasvasasiga tushishingiz mumkin, lekin agar shunday qilsangiz, LunaOS yangi paketlar bilan birga ishga tushiradigan snapshot, migratsiyalar va konfiguratsiya yangilanishlarini oʻtkazib yuborasiz. Shuning uchun LunaOS aslida toʻgʻridan-toʻgʻri tizim yangilashini toʻxtatadi va uning oʻrniga sizni `lunaos update` ga yoʻnaltiradi. (Agar haqiqatan ham nima qilayotganingizni bilsangiz, himoya sizga uni bitta tranzaksiya uchun qanday chetlab oʻtishni aytadi.)

### Yomon yangilanishlarni qaytarish

Yangilanishni qilgandan keyin qachonlardir muammoga duch kelsangiz, tizimingizni yangilanishdan oldin olingan snapshotga qaytarishingiz mumkin. Shunchaki qayta ishga tushiring va yangilanishni boshlashingizdan oldingi yuklash menyusida snapshotni tanlang.

![bootloader](images/bootloader.webp)

Agar qandaydir tarzda konfiguratsiya fayllaringiz buzilgan boʻlsa, terminalda `lunaos reinstall` yordamida LunaOS qayta oʻrnatishini ham bajarishingiz mumkin. Bu barcha standart LunaOS paketlarini qayta oʻrnatadi, sizni stable ga qoʻyadi va juda yangi boʻlgan har qanday paketlarni pasaytiradi hamda barcha konfiguratsiya fayllarini tiklaydi. Eʼtibor bering, buni qilish orqali LunaOS standartlariga barcha foydalanuvchi konfiguratsiya oʻzgarishlaringiz ustiga yoziladi!
