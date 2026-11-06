# Togol, Terbiar & Tabir Skrin

Banyak perkara yang anda tukar hari demi hari sebenarnya bukanlah tetapan. Ia ialah mod yang anda hidupkan selama sejam dan matikan semula: Cahaya Malam semasa anda bekerja lewat, jangan ganggu semasa anda membentangkan, Kekal Jaga semasa anda menonton sesuatu. LunaOS memanggilnya Togol, dan semuanya berfungsi dengan cara yang sama — kekunci panas, entri menu, dan arahan, semuanya menekan suis yang sama.

### Menu Togol

`Super + Ctrl + O` membuka _Pencetus > Togol_ terus, atau anda boleh berjalan ke sana dari menu LunaOS (`Super + Space`). Segala-galanya dalam senarai itu ialah suis yang anda boleh balikkannya tanpa memikirkan di mana keadaan berada. Suis untuk cara desktop kelihatan duduk di sebelahnya, di bawah _Pencetus > Penampilan_.

Dari terminal, suis yang sama ialah `lunaos toggle <thing>`. Jalankan `lunaos toggle` dengan sendirinya untuk melihat keseluruhan kumpulan.

| Togol | Kekunci Panas | Arahan |
| ------ | ------ | ------- |
| Cahaya Malam | `Super + Ctrl + N` | `lunaos toggle nightlight` |
| Senyapkan Pemberitahuan | `Super + Ctrl + ,` | `lunaos toggle notification silencing` |
| Kekal Jaga (tiada kunci terbiar) | `Super + Ctrl + I` | `lunaos toggle idle` |
| Tangkap ranap | — | `lunaos toggle crash-capture` |
| Tabir Skrin | — | `lunaos toggle screensaver` |
| Segerak Tema [Herdr](https://herdr.dev) | — | `lunaos toggle theme sync` |
| Bar Menu | `Super + Shift + Space` | `lunaos toggle bar` |
| Pad Sentuh | `XF86TouchpadToggle` | `lunaos toggle touchpad` |
| Skrin Sentuh | — | `lunaos toggle touchscreen` |
| Tidur | — | `lunaos toggle suspend` |
| GPU Hibrid | — | `lunaos toggle hybrid gpu` |

Suis Pad Sentuh, Skrin Sentuh, dan GPU Hibrid berada di bawah _Pencetus > Perkakasan_ (`Super + Ctrl + H`) dan bukannya di bawah Togol, kerana ia hanya muncul apabila anda benar-benar mempunyai Perkakasan itu. Suis Pad Sentuh dan Skrin Sentuh kekal selepas muat semula Hyprland — nama peranti yang dilumpuhkan disimpan ke fail keadaan kecil yang dibaca oleh Hyprland semasa permulaan untuk melumpuhkannya semula.

_Pencetus > Penampilan_ memegang Bar Menu dan animasi, ditambah beberapa perkara yang bukan arahan `lunaos toggle` tetapi berkelakuan sama: peratus bateri dalam bar, Susun Atur Ruang Kerja lalai, Jurang Tetingkap (`Super + Shift + Backspace`), dan aspek segi empat sama 1-tetingkap (`Super + Ctrl + Backspace`).

Kebanyakan daripadanya hanyalah fail bendera di bawah `~/.local/state/lunaos/toggles/`. Jika anda mahu bercabang padanya dalam skrip, `lunaos-toggle-enabled` memberi anda kod keluar dan bukannya membuat anda mencarinya:

```bash
lunaos-toggle-enabled screensaver-off && echo "screensaver is off"
```

Bendera dinamakan untuk keadaan mati — `screensaver-off`, `suspend-off`, `bar-off` — jadi kehadirannya bermakna ciri itu dilumpuhkan.

### Penunjuk dalam bar

Apabila mod dihidupkan, anda mendapat glif kecil di tengah bar atas di sebelah jam. Itulah widget Penunjuk, dan ia membawa pendiktean, rakaman skrin, Peringatan yang belum selesai, Cahaya Malam, jangan ganggu, dan Kekal Jaga.

Penunjuk tidak aktif disembunyikan. Layangkan kawasan di sekelilingnya dan ia pudar masuk dalam keadaan malap, jadi anda boleh mengklik satu untuk menghidupkannya tanpa mengetahui kekunci panasnya. Mengklik yang aktif mematikannya kembali. Jika anda lebih suka melihat kesemuanya pada setiap masa, tetapkan `alwaysShow` kepada `true` pada entri `lunaos.indicators` dalam `~/.config/lunaos/shell.json` — lihat [bar atas](05-the-top-bar.md) untuk cara widget bar dikonfigurasikan.

### Cahaya Malam

`Super + Ctrl + N` menghangatkan skrin kepada 4000K, dan menekannya sekali lagi mengembalikannya kepada 6500K. Ia dipacu oleh hyprsunset, yang Togol mulakan untuk anda jika ia belum berjalan.

Secara lalai hyprsunset tidak melakukan apa-apa langsung pada skrin anda. `~/.config/hypr/hyprsunset.conf` dihantar dengan profil identiti tepatnya supaya paparan kekal tidak disentuh sehingga anda meminta kehangatan. Jika anda lebih suka ia bertukar mengikut jam, gantikannya dengan profil masa:

```
profile {
    time = 20:00
    temperature = 4000
}
```

Kemudian mulakan hyprsunset semasa log masuk dengan menambah `o.launch_on_start("hyprsunset")` kepada `~/.config/hypr/autostart.lua`. Pasangan 4000K/6500K yang digunakan oleh Togol adalah tetap, jadi fail konfigurasi ialah tempat anda pergi jika anda mahukan suhu yang berbeza.

### Jangan ganggu

`Super + Ctrl + ,` menyenyapkan Pemberitahuan. Tiada roti bakar muncul semasa ia dihidupkan, dan penunjuk loceng yang dipalang duduk dalam bar untuk mengingatkan anda mengapa desktop menjadi senyap.

Walau bagaimanapun, tiada apa-apa yang hilang. Pemberitahuan yang disenyapkan ditulis terus ke dalam sejarah pemberitahuan anda, iaitu rekod tepat yang anda mahukan apabila anda kembali dan tertanya-tanya apa yang anda terlepas. Bukanya dengan `Super + Shift + Alt + ,`. Lihat [notis](10-notices.md) untuk selebihnya kisah pemberitahuan.

Dua jenis mesej masih sampai: roti bakar pengesahan LunaOS sendiri untuk sesuatu yang baru anda lakukan ("Tema ditukar", "Tangkap Skrin disimpan"), dan amaran kritikal yang dihantar dari baris arahan. Aplikasi sembang yang menandakan segala-galanya kritikal untuk memaksa laluannya di hadapan anda tidak layak.

### Terbiar

Shell LunaOS memiliki tingkah laku terbiar, dan pemasaan ialah blok `idle` peringkat atas dalam `~/.config/lunaos/shell.json`:

```json
{
  "version": 1,
  "idle": {
    "screensaver": 150,
    "lock": 300
  }
}
```

Kedua-dua nombor ialah saat yang dikira dari saat anda menjadi terbiar — bukan antara satu sama lain. Jadi dengan lalai, Tabir Skrin muncul selepas dua minit setengah dan skrin kunci mengambil alih pada lima minit, sama ada Tabir Skrin berjalan atau tidak. Simpan fail dan shell mengambil pemasaan baharu serta-merta.

Jika anda menolak Tabir Skrin sebelum tarikh akhir kunci, itu dikira sebagai aktiviti dan kunci yang belum selesai dibatalkan. Anda tidak terkunci keluar kerana menjeling mesin anda.

Untuk berhenti mengunci semasa terbiar sepenuhnya, `Super + Ctrl + I` — atau `lunaos toggle idle` — menghidupkan Kekal Jaga, dan penunjuk cawan kopi muncul dalam bar. Itulah yang perlu ditekan sebelum pembentangan panjang atau binaan yang anda mahu tonton. Tekannya sekali lagi untuk kembali kepada normal. `lunaos toggle idle status` mencetak keadaan semasa sebagai JSON jika anda memerlukannya dari skrip.

Ini adalah mengenai mengunci dan Tabir Skrin, bukan kuasa. Tidur dan hibernasi mempunyai persediaan sendiri dalam [tidur sistem](36-system-sleep.md).

### Tabir Skrin

Tabir Skrin LunaOS ialah seni ASCII yang berjalan melalui kesan teks rawak, satu tika bagi setiap monitor. Sebarang kekunci atau pergerakan tetikus keluar daripadanya.

Anda boleh memulakannya atas permintaan dari _Sistem > Tabir Skrin_ (`Super + Esc`), yang memaksanya muncul walaupun anda telah mematikan Tabir Skrin terbiar. Tiada kekunci panas yang diikat kepadanya secara lalai.

`lunaos toggle screensaver` ialah apa yang mematikan yang terbiar, jika anda lebih suka pergi terus dari bekerja kepada terkunci. Ia memerlukan terminal yang ia tahu cara mengkonfigurasikannya — Alacritty, Foot, Ghostty, atau Kitty — dan akan memberitahu anda begitu jika terminal lalai anda adalah sesuatu yang lain.

Logo yang dilukisnya adalah milik anda untuk ditukar, di bawah _Gaya > Tabir Skrin_. Muat naik png atau svg dan LunaOS menukarnya kepada ASCII. Lihat [penjenamaan](41-branding.md).

### Skrin kunci

`Super + Ctrl + L` mengunci mesin. Itu menjalankan skrin kunci dari shell LunaOS, mengosongkan paparan, menetapkan semula Susun Atur Papan Kekunci anda kepada yang pertama supaya anda tidak menaip kata laluan anda dalam abjad yang salah, dan — jika anda menjalankannya — mengunci 1Password semasa keluar.

Skrin kunci mengambil kata laluan, dan ia juga akan mengambil cap jari sebaik sahaja anda menyediakannya. Itu, dan cara lain untuk mengesahkan, dibincangkan dalam [pengesahan perkakasan](37-hardware-authentication.md).
