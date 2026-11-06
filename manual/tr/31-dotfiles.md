# Dotfiles

LunaOS, öncelikle `~/.config` içinde yaşayan sözde dotfiles aracılığıyla yapılandırılır. Bunlar sizin değişiklikleriniz için sizin dosyalarınız olarak kabul edilir. `/usr/share/lunaos` içinde yaşayan dosyalar LunaOS'nin kendisine aittir ve bunlarla uğraşmamalısınız. `/usr/share/lunaos` içinde herhangi bir şeyi değiştirmeniz gerekiyorsa, bunun yerine değeri `~/.config` içinde geçersiz kılmalısınız.

Temel yapılandırmalar, LunaOS menüsünden (`Super + Space`) doğrudan düzenlenebilir; örneğin _Ayarlar > Monitors_, _Ayarlar > Keybindings_, _Ayarlar > Input_ ve _Ayarlar > Config > [file]_. Bunu bu şekilde yaptığınızda, yapılandırma düzenlemelerinden sonra yeniden başlatılması gereken tüm işlemler, editörden çıktıktan sonra otomatik olarak yeniden başlatılacaktır (varsayılan olarak Neovim — `:wq`, hatırlayın! — ancak bunu _Ayarlar > Defaults > Editor_ aracılığıyla değiştirebilirsiniz).

İşte `~/.config` içindeki temel dosyaların ve neyi kontrol ettiklerinin bir listesi:

| Dosya | Amaç |
| ----------------------- | --------------------- |
| `~/.config/hypr/hyprland.lua` | Ana Hyprland yapılandırması. LunaOS varsayılanlarını artı aşağıdaki geçersiz kılma dosyalarınızı yükler. [Hyprland yapılandırmaları hakkında daha fazla bilgi edinin](https://wiki.hypr.land/Configuring/).  |
| `~/.config/hypr/bindings.lua` | Kendi tuş atamalarınız ve varsayılanların geçersiz kılmaları. |
| `~/.config/hypr/monitors.lua` | Monitörlerinizi, çözünürlüğü ve konumu kontrol eder. |
| `~/.config/hypr/input.lua` | Klavye düzeninizi, fare ve dokunmatik yüzey ayarlarınızı kontrol eder. |
| `~/.config/hypr/looknfeel.lua` | Boşlukları, kenarlıkları, animasyonları ve görünümün geri kalanını kontrol eder. |
| `~/.config/hypr/autostart.lua` | Oturumla başlatılan ek işlemleri kontrol eder. |
| `~/.config/lunaos/shell.json` | LunaOS kabuğunu kontrol eder: çubuk konumu, düzen ve widget'lar, artı ekran koruyucu, kilit ve boşta kalma süreleri. |
| `~/.config/foot/foot.ini` | Terminalinizi kontrol eder (foot varsayılandır). |
| `~/.XCompose` | Hızlı erişim emojinizi ve ad/e-posta otomatik tamamlamanızı tanımlar. Değişiklik yaptıktan sonra `lunaos-restart-xcompose` komutunu çalıştırdığınızdan emin olun. |

Kendi kurulumunuzu ince ayarlamak için çok fazla değişiklik yaparsanız, tüm bu dotfiles'ı yedeklemek iyi bir fikirdir. [Stow bunu yapmanın harika bir yoludur](https://www.youtube.com/watch?v=NoFiYOqnC4o).

### Kendi uygulamalarınızı oturumla başlatmak

Her giriş yaptığınızda bir şeyin çalışmasını isterseniz — bir eşitleme arka plan süreci, bir sohbet uygulaması, kendi betiğiniz — onu `~/.config/hypr/autostart.lua` içine koyun:

```lua
o.launch_on_start("my-service")
```

Bu, komutu oturumun bir parçası olarak başlatır, böylece yeniden çıkış yaptığınızda düzgün şekilde temizlenir.

### Sistem olaylarında betik çalıştırmak

LunaOS bir avuç anda kancalar tetikler ve kendi betiklerinizi bunlara asabilirsiniz. Bunlar `~/.config/lunaos/hooks/<event>.d/` içinde yaşar, olay başına bir dizin, ve oradaki çalıştırılabilir her dosya olay gerçekleştiğinde çalışır:

| Olay | Ne zaman çalışır |
| ----- | ------------ |
| `post-boot` | Masaüstü başladıktan hemen sonra |
| `post-update` | `lunaos update` komutunun sonuna doğru, paketler, taşımalar ve hizmet yeniden başlatmalarından sonra, mise araçları güncellenmeden önce |
| `pre-refresh-pacman` | `lunaos refresh pacman` paket yapılandırmasını yeniden eşitledikten sonra, paketleri güncellemeden önce; bir kanal değişikliği, aynı yenileme adımı sırasında bunu çalıştırır |
| `theme-set` | Bir tema değişikliğinden sonra (`$1` içinde tema adı) |
| `font-set` | Bir yazı tipi değişikliğinden sonra (`$1` içinde yazı tipi adı) |
| `battery-low` | Pil azaldığında (`$1` içinde yüzde) |

`pre-refresh-pacman` kancası, paket işleminden önce çalıştığı için özel depoların veya `IgnorePkg` satırlarının ait olduğu yerdir. Güncellemeyle ilgili her iki kanca da LunaOS, önbelleğe alınmış sudo yetkilendirmesini temizledikten sonra sizin kullanıcınız olarak çalışır; bu yüzden `sudo` kullanan bir kancanın kendi yetkilendirmesine ihtiyacı vardır ve parolanızı isteyebilir.

Bu dizinlerin her biri zaten bir kancanın şeklini gösteren bir `.sample` dosyası içerir — onu işe koymak için adından `.sample` ekini çıkarın. Başka yerde yazdığınız bir betiği kurmak için `lunaos hook install post-boot ~/my-hook` komutunu kullanın; bu onu kopyalar ve çalıştırılabilir yapar.

### Kendi menü girişlerinizi eklemek

LunaOS menüsü (`Super + Space`), `~/.config/lunaos/extensions/lunaos-menu.jsonc` dosyasını düzenleyerek kendi satırlarınızla genişletilebilir. Girişler noktalı bir kimlik ile anahtarlanır ve kimlik onları ağaca yerleştirir; bu yüzden `personal` kök menüde, `personal.notes` ise onun içinde görünür:

```jsonc
"personal": {"icon":"","label":"Personal"},
"personal.notes": {"icon":"󰎞","label":"Notes","action":"lunaos-launch-editor ~/notes"},
```

Mevcut bir kimliği yeniden kullanırsanız, yeni bir tane eklemek yerine o satırı geçersiz kılarsınız. Dosya, kullanılabilir tüm alanlar yorum olarak belgelenmiş olarak gelir.

### Kendi kabuk dışa aktarımlarınızı, işlevlerinizi ve takma adlarınızı eklemek

LunaOS, bir sürü ergonomik takma ad ve yararlı işlevle birlikte gelir, ancak kendinizinkini eklemek istemeniz çok yaygındır. Takma adları, işlevleri ve dışa aktarımları `~/.bashrc` içine eklemelisiniz. Bu dosyanın üzerine güncellemelerde yazılmayacaktır. LunaOS varsayılanlarından herhangi birini değiştirmek isterseniz, bunları buraya da güvenle ekleyebilirsiniz.

### İç LunaOS dosyalarını değiştirmek

Bakın, bu sizin bilgisayarınız. Onunla istediğinizi yapabilirsiniz, ancak `/usr/share/lunaos` içindeki dosyalara doğrudan değişiklik yapmamanızı tavsiye ederim. Bunlar LunaOS pacman paketine aittir, bu yüzden değişiklikleriniz bir sonraki güncellemede basitçe üzerine yazılacaktır. Bunun yerine sevmediğiniz varsayılan değerleri `~/.config/*` klasöründe geçersiz kılmanız daha iyidir.

Varsayılan tuş atamaları gibi hemen her şeyi bu şekilde değiştirebilirsiniz. Örneğin [Obsidian](https://obsidian.md/) yerine [Joplin](https://joplinapp.org/) koymak için `~/.config/hypr/bindings.lua` dosyasını düzenleyin (`lunaos-pkg-add joplin-bin` ile kurun):

```lua
o.rebind("SUPER + SHIFT + O", "Joplin", "joplin-desktop")
```

`o.rebind`, yerine geçenini eklemeden önce mevcut atamayı kaldırır. Başlatma yardımcıları ve atama seçenekleri dahil `o.bind` ile aynı argümanları alır. Bir atama eklemek için `o.bind`, yerine bir şey koymadan birini kaldırmak için `hl.unbind` kullanın.

İç LunaOS dosyaları üzerinde uğraşmakta ısrar ederseniz, _Güncelle > Channel > Dev_ aracılığıyla dev kanalına geçin. Bu, LunaOS'yi `~/lunaos` içindeki kaynak kodun bir git kopyasına bağlar; bunu gönlünüzce değiştirmekte özgürsünüz. Size ne yapacağınızı söyleyecek kimse yok burada!

### Herhangi bir değişikliği sıfırlamak

Yapılandırmaları karmakarışık ederseniz, bunları her zaman LunaOS menüsündeki _Güncelle > Config_ aracılığıyla varsayılanlara döndürebilirsiniz. Veya her şeyi sıfırlamak için `lunaos reinstall configs` komutunu çalıştırarak.
