# Kemas Kini

LunaOS dan pakej anda sentiasa dikemas kini melalui _Kemas Kini > LunaOS_ dalam menu LunaOS (`Super + Space`).

LunaOS itu sendiri dipasang sebagai pakej pacman biasa daripada [Repositori Pakej LunaOS](https://github.com/omacom-io/omarchy-pkgs), jadi kemas kini memasang [keluaran LunaOS terkini](https://github.com/gregpavoine/LunaOS/releases), menjalankan mana-mana migrasi tertangguh untuk menyelaraskan sistem anda dengan yang terkini, dan mengemas kini semua pakej sistem daripada [Cermin Arch LunaOS](https://github.com/omacom-io/omarchy-mirror) dan [AUR](https://aur.archlinux.org/) (jika anda telah memasang mana-mana pakej AUR).

Apabila keluaran baharu dibuat, ikon anak panah bulatan akan muncul di kanan jam anda. Klik dan proses kemas kini akan bermula.

![update-available](images/update-available.webp)

### Empat saluran

LunaOS dikemas kini sepanjang empat saluran: stabil, RC, edge, dan dev. Pemasangan baharu bermula pada saluran stabil, yang menjejak [keluaran rasmi](https://github.com/gregpavoine/LunaOS/releases/), serta [cermin Arch LunaOS stabil](https://github.com/omacom-io/omarchy-mirror) yang berjalan sebulan di belakang yang terkini, supaya kami boleh menangkap sebarang ketidakserasian baharu yang memerlukan perubahan konfigurasi sebelum ia menyebabkan masalah kepada orang.

Tetapi jika anda ingin membantu mengesan isu berpotensi itu, anda boleh berjalan pada saluran edge. Itu akan memastikan pakej LunaOS anda menjejak binaan pembangunan terkini, dan membolehkan anda mengemas kini ke pakej Arch terkini sebaik sahaja tersedia. Anda hanya patut melakukan ini jika anda berpengalaman dengan Linux, dan tahu cara memulihkan sistem yang bermasalah.

Sebelum mana-mana keluaran utama baharu, kami akan melakukan pengesahan akhir menggunakan saluran RC. Jika anda berminat untuk membantu dengan penggilapan akhir, datang berjumpa di #lunaos-release-candidates di Discord.

Akhirnya, ada saluran dev, yang menghubungkan LunaOS terus ke checkout git kod sumber di `~/lunaos`, digabungkan dengan pakej edge. Anda hanya patut menggunakan saluran ini jika anda pengguna Linux yang berpengalaman, bekerja terus pada LunaOS, dan sanggup menoleransi kerosakan.

Anda boleh beralih antara saluran menggunakan _Kemas Kini > Saluran_ daripada menu LunaOS (atau `lunaos-channel-set` di terminal).

### Kemas kini perisian tegar

Pakej anda bukan satu-satunya benda yang menjadi usang. Banyak komputer riba dan persisi menghantar perisian tegar BIOS, SSD, dan dok melalui Perkhidmatan Perisian Tegar Vendor Linux, dan _Kemas Kini > Perisian Tegar_ dalam menu LunaOS akan mengambil dan memasang apa-apa yang menunggu perkakasan anda. Ia memasang `fwupd` kali pertama anda menjalankannya. Banyak perisian tegar hanya boleh ditulis semasa but semula, jadi jangan terkejut jika diminta berbuat demikian.

### Amaran mengenai kemas kini pacman/yay terus

Jika anda sudah biasa dengan Arch, anda mungkin tergoda untuk menjalankan `pacman -Syu` atau `yay -Syu` sendiri, tetapi jika anda berbuat demikian, anda akan terlepas snapshot, migrasi, dan kemas kini konfigurasi yang dijalankan LunaOS bersama-sama pakej baharu. Itulah sebabnya LunaOS sebenarnya akan menghentikan peningkatan sistem terus dan menunjukkan anda ke `lunaos update` sebaliknya. (Jika anda benar-benar tahu apa yang anda lakukan, pengawal akan memberitahu anda cara memintasnya untuk satu transaksi.)

### Menggulung semula kemas kini yang buruk

Jika anda pernah menghadapi masalah selepas melakukan kemas kini, anda boleh menggulung semula sistem anda ke snapshot yang diambil sebelum kemas kini. Mulakan semula sahaja dan pilih snapshot dalam menu pemuatan but daripada sebelum anda memulakan kemas kini.

![bootloader](images/bootloader.webp)

Jika entah bagaimana fail konfigurasi anda telah rosak, anda juga boleh melakukan pemasangan semula LunaOS menggunakan `lunaos reinstall` di terminal. Ini akan memasang semula semua pakej LunaOS lalai, meletakkan anda pada stabil dan menurunkan mana-mana pakej yang terlalu baharu, dan menetapkan semula semua fail konfigurasi. Perhatikan bahawa semua perubahan konfigurasi pengguna anda kepada lalai LunaOS akan ditimpa dengan melakukan ini!
