# Pengekstrakan Teks & Pendiktean

### Pengekstrakan Teks

Tekan `Super + Ctrl + PrtScr` untuk memilih kawasan pada skrin untuk pengekstrakan teks. Model OCR sumber terbuka tesseract kemudian akan menukar pemilihan itu kepada teks dengan cepat dan meletakkannya pada papan keratan. Kemudian anda hanya tekan `Super + V` untuk menampal.

Ini sangat membantu untuk mengambil alamat daripada pengaki imej atau nombor telefon yang terbenam dalam tajuk laman web.

 ![text-extraction](images/text-extraction.webp)

### Pendiktean

Pada pemasangan baharu, [Superwhisper](https://superwhisper.com/) ialah bahagian belakang pendiktean lalai, menggunakan pemprosesan awan tanpa muat turun model luar talian. Pilih Superwhisper atau [Voxtype](https://voxtype.io/) melalui _Persediaan > Lalai > Pendiktean_ dalam menu LunaOS. Memilih bahagian belakang memasangnya jika diperlukan dan menjadikannya lalai untuk pendiktean. Voxtype memuatkan model Bahasa Inggeris asas yang mengambil ruang 150MB; tukar model dengan `voxtype setup model` dan tetapan lain melalui `~/.config/voxtype/config.toml`. Superwhisper membuka tetapannya untuk memilih pemprosesan awan atau setempat, dan OPR mengekalkan aplikasi dikemas kini.

Anda mendikte dengan menahan `Right Alt` atau `F9`, atau dengan menogol dengan `Super + Ctrl + X`, dan teks yang didikte akan muncul dalam kawasan input yang difokuskan. Pintasan ini menggunakan bahagian belakang yang dipilih. Right Alt juga memulakan pendiktean apabila digunakan dalam kord pengubah; gunakan Left Alt untuk kord tersebut. Pada susun atur papan kekunci di mana Right Alt ialah AltGr, gunakan `F9` untuk mengekalkan AltGr tersedia untuk menaip.

Pintasan yang sama berfungsi dengan mana-mana bahagian belakang. Pilih bahagian belakang anda melalui Persediaan > Lalai > Pendiktean. Pemasang bahagian belakang mengkonfigurasikannya dan menyimpan pemilihan. Jalankan `lunaos default dictation` untuk melihat yang mana dipilih. Pemilihan disimpan dalam `~/.config/lunaos/defaults/dictation`. Pemasangan baharu memilih Superwhisper secara automatik.

LunaOS memuatkan integrasi desktop bahagian belakang yang dipilih secara automatik. Tiada konfigurasi Superwhisper diperlukan dalam `~/.config/hypr/`.

Ikatan dan skrip tersuai boleh menggunakan `lunaos dictation start`, `lunaos dictation stop`, dan `lunaos dictation toggle`. Bahagian belakang yang dipilih mengendalikan rakaman dan transkripsi. Memilih Superwhisper melumpuhkan pintasan tahan asalnya dan menetapkan togol asalnya kepada `Alt + Space`, mengekalkannya berasingan daripada kekunci rakaman LunaOS. Superwhisper memerlukan sekurang-kurangnya satu pintasan rakaman asal, dan pintasan pembatalannya kekal tersedia.

Bahagian belakang tambahan boleh disepadukan tanpa mengubah LunaOS. Sediakan boleh laksana bernama `lunaos-dictation-<backend>` pada PATH yang menerima `start`, `stop`, dan `toggle` dan kembali dengan segera, dengan status bukan sifar semasa kegagalan. Pemasangnya mengkonfigurasi pembekal dan memanggil `lunaos-dictation-use <backend>` selepas persediaan berjaya. Ini menghentikan bahagian belakang sebelumnya dan menyimpan pemilihan secara atomik; `stop` mesti berjaya apabila pembekal terbiar atau tidak tersedia. Nama bahagian belakang menggunakan huruf kecil, digit, dan tanda sempang. Sediakan `lunaos-install-dictation-<backend>` untuk persediaan; `lunaos default dictation <backend>` melancarkannya. Tambah entri menu Lalai melalui sambungan menu LunaOS untuk memanggil pemilih itu. Integrasi desktop pilihan boleh berada dalam `${XDG_CONFIG_HOME:-$HOME/.config}/<backend>/shortcuts.lua`.
