# Güvenlik

LunaOS güvenliği son derece ciddiye alır. Bu, _Gerçek Dünya_'da _Gerçek İş_ yapmak için kullanabileceğiniz bir işletim sistemi olarak tasarlanmıştır. Bir dizüstü bilgisayarı kaybetmenin bir güvenlik acil durumuna yol açamayacağı bir yer. Yani işte yaptıklarımız:

1. *Tam disk şifrelemesi zorunludur*: Bu, verilerinizin fiziksel korumasını güvence altına almanın en önemli adımıdır. Bilgisayarınız kaybolur veya çalınırsa, veriler standart LUKS (Linux Unified Key Setup) kullanılarak tamamen şifrelenir.
2. *Güvenlik duvarı varsayılan olarak etkindir*: [LocalSend](https://localsend.org/) için 53317 numaralı bağlantı noktası dışında gelen tüm trafik varsayılan olarak engellenir. SSH bile, kurulumun bir parçası olarak 22 numaralı bağlantı noktasını açan (kaba kuvvete karşı hız sınırlı) _Ayarlar > Güvenlik > SSHD_ aracılığıyla açana kadar kapalıdır. Kapsayıcılarınızın yanlışlıkla dünyaya açılmasını önlemek için [ufw-docker](https://github.com/chaifeng/ufw-docker) kurulumunu kullanarak Docker erişimini bile kilitliyoruz.
3. *Arch her zaman en son güncelleştirmelere sahiptir*: LunaOS'nin üzerine kurulduğu temel dağıtım olan Arch, yuvarlanan bir dağıtımdır. Bu, herhangi bir pakette keşfedilen ve yamalanan herhangi bir güvenlik açığının hızla `lunaos-update` kullanılarak kurulmaya hazır olduğu anlamına gelir. Bu şekilde her zaman her şeyin en son ve en güvenli sürümlerini çalıştırırsınız.
4. *LunaOS kendi paketlerini ve aynasını tutar*: LunaOS varsayılan olarak yalnızca Arch'in kendi core/extra/multilib depolarından ve kendi LunaOS Paket Deposu'ndan gelen paketlere dayanır. Yazılımı doğrudan AUR'dan kurabilirsiniz, ancak temel kurulum bunu yapmaz — yalnızca üçüncü taraf tarayıcılar gibi birkaç isteğe bağlı kurulum AUR'dan çeker.
5. *Cloudflare bizi DDoS'tan korur*: Tüm LunaOS dağıtım altyapısı — ISO'lar, LunaOS paketleri, Arch aynası — Cloudflare'in müthiş DDoS kalkanının arkasında korunur ve onların CDN'inde barındırılır. Bu, üstün kullanılabilirlik sağlar.

## Parolalarınızı değiştirme

Şifreli bir kurulumda iki parolanız vardır: önyüklemede sürücünün kilidini açan parola ve giriş yapmak ve `sudo` kullanmak için kullandığınız parola. Kurulum bunları aynı yapar. İkisi de LunaOS menüsünde _Güncelle > Parola_ altında değiştirilebilir — birincisi için _Sürücü Şifrelemesi_, ikincisi için _Kullanıcı_. Sürücü parolasını değiştirmek önce geçerli olanı sorar, bu yüzden onu el altında bulundurun. Bir değişiklik yarıda kesilirse, _Sürücü Şifrelemesi_ yolunu tekrar çalıştırın: sürücünün kilidini şu anda açan parolayı sorar ve değişikliği tamamlar.

## Zaten kullandığınız bir makineyi devretme

Makinenizi başkasına devrediyorsanız, onu yeniden kurmak zorunda değilsiniz. LunaOS menüsünde _Ayarlar > Bilgisayarı Sıfırla_ yolunu çalıştırın, onaylamak için `reset` yazın ve yeniden başlatın. Bu, her kullanıcı hesabını ve `/home` içindeki her şeyi siler, kurulumdan beri yaptığınız tüm paketleri ve sistem değişikliklerini atar ve makinenin kimliğini — ağ bağlantıları, ana bilgisayar anahtarları ve hepsi — temizler. Geri gelen şey, yeni sahibinin kendi adını, parolasını ve şifreleme parolasını girmesi için hazır olan ilk önyüklemedeki kurulum sihirbazıdır.

Bu, kurulum programının aldığı temel anlık görüntüyü geri yükleyerek çalışır, bu yüzden yalnızca LunaOS ISO'sundan kurulmuş makinelerde kullanılabilir. Ve şifrelemesiz bir sürücüde, sıfırlama güvenli bir silme yerine silmedir; bu yüzden veriler hassassa bunun yerine yeni bir kurulum yapın.

## Parolasız sudo

Bazen `sudo` komutunun sormasını durdurmak istersiniz; en sık olarak bir yapay zekâ ajanı sizin için uzun bir sistem çalışması yaptığında. _Ayarlar > Güvenlik > Parolasız Sudo_ erişime ne kadar süre izin verileceğini sorar: **15 dakika**, **1 Saat**, **1 Gün** veya **Kalıcı olarak**. Erişim etkinken diğer menü çubuğu göstergelerinin yanında kırmızı bir uyarı simgesi görünür. Erişimi kapatmak için ona tıklayın veya komutu tekrar çalıştırın. Ayrıca `lunaos-sudo-passwordless 30` ile kendi dakika sayınızı (1 ile 1440 arasında) verebilir veya `lunaos-sudo-passwordless permanent` kullanabilirsiniz.

Zamanlanmış erişim, son tarihi geçen bir askıya almadan devam ettikten hemen sonra dahil olmak üzere, otomatik olarak sona erer. Bilgisayarı yeniden başlatmak onu erken bitirir. Kalıcı erişim yeniden başlatmalarda da sürer ve siz devre dışı bırakana kadar etkin kalır.

LunaOS'nin ayarlar paketini güncelleştirme veya kaldırma, sona erme desteği değişmeden önce herhangi bir geçici yetkiyi sonlandırır. Komut bir yetkilendirme veya temizleme hatası bildirirse, başka bir yetki etkinleştirmeyi denemeden önce bunu çözün; bir hata, parolasız erişimin etkin olmadığı anlamına gelmez.

Bu konuda açık göz olun: açıkken, kullanıcınız olarak çalışan her şey, sorulmadan root olarak her şeyi yapabilir. Bütün nokta budur ve bütün risk de budur.

## İmzalama Anahtarları

Tüm ISO imzaları ve LunaOS depo paketi için genel anahtar `40DFB630FF42BCFFB047046CF0134EE680CAC571` anahtarıdır ([openpgp.org adresinde doğrulayın](https://keys.openpgp.org/search?q=pkgs%40omarchy.org)). `omarchy/omarchy-keyring` paketi bunu da içerir ve olası güncelleştirmeleri sorunsuz şekilde dağıtmak için kullanılacaktır.

Herhangi bir ISO sürümünün imzasını URL'ye .sig ekleyerek bulabilirsiniz. Örneğin https://iso.omarchy.org/omarchy-x.x.x.iso.sig.
