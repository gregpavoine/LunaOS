# Commutadors, repòs i el salvapantalles

Gran part del que canvieu dia a dia no és realment un paràmetre. És un mode que activeu durant una hora i desactiveu de nou: la llum nocturna mentre treballeu fins tard, no molesteu mentre presenteu, mantenir despert mentre mireu alguna cosa. LunaOS els anomena commutadors, i tots funcionen igual — una drecera, una entrada de menú i una ordre, tots accedint al mateix interruptor.

### El menú de commutadors

`Super + Ctrl + O` obre _Activa > Commuta_ directament, o hi podeu anar des del menú d'LunaOS (`Super + Space`). Tot el que hi ha en aquesta llista és un interruptor que podeu accionar sense pensar en on viu l'estat. Els interruptors sobre com es veu l'escriptori són al costat, a _Activa > Aparença_.

Des del terminal, els mateixos interruptors són `lunaos toggle <thing>`. Executeu `lunaos toggle` pel seu compte per veure tot el grup.

| Commutador | Drecera | Ordre |
| ------ | ------ | ------- |
| Llum nocturna | `Super + Ctrl + N` | `lunaos toggle nightlight` |
| Silencia les notificacions | `Super + Ctrl + ,` | `lunaos toggle notification silencing` |
| Mantén despert (sense bloqueig en repòs) | `Super + Ctrl + I` | `lunaos toggle idle` |
| Captura de fallades | — | `lunaos toggle crash-capture` |
| Salvapantalles | — | `lunaos toggle screensaver` |
| Sincronització de tema de [Herdr](https://herdr.dev) | — | `lunaos toggle theme sync` |
| Barra de menú | `Super + Shift + Space` | `lunaos toggle bar` |
| Panell tàctil | `XF86TouchpadToggle` | `lunaos toggle touchpad` |
| Pantalla tàctil | — | `lunaos toggle touchscreen` |
| Suspèn | — | `lunaos toggle suspend` |
| GPU híbrida | — | `lunaos toggle hybrid gpu` |

Els interruptors del panell tàctil, la pantalla tàctil i la GPU híbrida viuen a _Activa > Maquinari_ (`Super + Ctrl + H`) en lloc de a Commuta, perquè només apareixen quan realment teniu aquest maquinari. Els del panell tàctil i la pantalla tàctil sobreviuen a una recàrrega de Hyprland — el nom del dispositiu desactivat es desa en un petit fitxer d'estat que Hyprland llegeix a l'arrencada per desactivar-lo de nou.

_Activa > Aparença_ conté la barra de menú i les animacions, a més d'algunes coses que no són ordres `lunaos toggle` però que es comporten igual: el percentatge de bateria a la barra, la disposició per defecte dels espais, la separació de finestres (`Super + Shift + Backspace`) i l'aspecte quadrat d'una finestra (`Super + Ctrl + Backspace`).

La majoria són simplement un fitxer de marca a `~/.local/state/lunaos/toggles/`. Si voleu ramificar-ne un en un script, `lunaos-toggle-enabled` us dona un codi de sortida en lloc de fer-vos anar a buscar:

```bash
lunaos-toggle-enabled screensaver-off && echo "screensaver is off"
```

Les marques s'anomenen per l'estat desactivat — `screensaver-off`, `suspend-off`, `bar-off` — de manera que la seva presència vol dir que la funció està desactivada.

### Indicadors a la barra

Quan un mode és activat, rebeu un petit glif al mig de la barra superior al costat del rellotge. És el giny d'indicadors, i porta el dictat, la gravació de pantalla, els recordatoris pendents, la llum nocturna, no molesteu i mantenir despert.

Els indicadors inactius són ocults. Passeu el cursor per la zona que els envolta i apareixen atenuats, de manera que en podeu fer clic en un per activar-lo sense saber-ne la drecera. Fer clic en un d'actiu el torna a desactivar. Si preferiu veure'ls tots sempre, definiu `alwaysShow` a `true` a l'entrada `lunaos.indicators` de `~/.config/lunaos/shell.json` — vegeu [la barra superior](05-the-top-bar.md) per saber com es configuren els ginys de la barra.

### Llum nocturna

`Super + Ctrl + N` escalfa la pantalla a 4000 K, i tornar-ho a prémer la retorna a 6500 K. El mou hyprsunset, que el commutador inicia per vosaltres si encara no s'està executant.

Per defecte hyprsunset no fa res en absolut a la vostra pantalla. `~/.config/hypr/hyprsunset.conf` ve amb un perfil d'identitat precisament perquè la pantalla es mantingui intacta fins que demaneu escalfor. Si preferiu que canvii segons el rellotge, substituïu-lo per un perfil horari:

```
profile {
    time = 20:00
    temperature = 4000
}
```

Després inicieu hyprsunset en iniciar sessió afegint `o.launch_on_start("hyprsunset")` a `~/.config/hypr/autostart.lua`. La parella 4000 K/6500 K que fa servir el commutador és fixa, de manera que el fitxer de configuració és on aneu si voleu una temperatura diferent.

### No molesteu

`Super + Ctrl + ,` silencia les notificacions. Cap avís emergent apareix mentre és activat, i l'indicador de la campana ratllada és a la barra per recordar-vos per què l'escriptori s'ha tornat silenciós.

Però no es perd res. Una notificació silenciada s'escriu directament al vostre historial de notificacions, que és exactament el registre que voleu quan torneu i us pregunteu què us heu perdut. Obriu-lo amb `Super + Shift + Alt + ,`. Vegeu [avisos](10-notices.md) per a la resta de la història de les notificacions.

Dos tipus de missatge encara passen: els avisos de confirmació propis d'LunaOS per a una cosa que acabeu de fer (“Theme changed”, “Screenshot saved”) i les alertes crítiques enviades des de la línia d'ordres. Les aplicacions de xat que ho marquen tot com a crític per obrir-se pas davant vostre no qualifiquen.

### Repòs

El shell d'LunaOS és el propietari del comportament en repòs, i els temps són un bloc `idle` de nivell superior a `~/.config/lunaos/shell.json`:

```json
{
  "version": 1,
  "idle": {
    "screensaver": 150,
    "lock": 300
  }
}
```

Tots dos números són segons comptats des del moment que vau quedar inactius — no l'un des de l'altre. Així, amb els valors per defecte, el salvapantalles apareix al cap de dos minuts i mig i la pantalla de bloqueig pren el relleu als cinc minuts, tant si el salvapantalles s'ha executat com si no. Deseu el fitxer i el shell recull els nous temps immediatament.

Si descarteu el salvapantalles abans del termini de bloqueig, això compta com a activitat i el bloqueig pendent es cancel·la. No us quedeu fora per fer un cop d'ull a la màquina.

Per deixar de bloquejar en repòs completament, `Super + Ctrl + I` — o `lunaos toggle idle` — activa mantenir despert, i l'indicador de la tassa de cafè apareix a la barra. És el que cal prémer abans d'una presentació llarga o d'una compilació que voleu mirar. Torneu-lo a prémer per tornar a la normalitat. `lunaos toggle idle status` imprimeix l'estat actual com a JSON si el necessiteu des d'un script.

Això tracta del bloqueig i del salvapantalles, no de l'energia. La suspensió i la hibernació tenen la seva pròpia configuració a [repòs del sistema](36-system-sleep.md).

### El salvapantalles

El salvapantalles d'LunaOS és art ASCII que recorre efectes de text aleatoris, una instància per monitor. Qualsevol tecla o moviment del ratolí en surt.

El podeu iniciar sota demanda des de _Sistema > Salvapantalles_ (`Super + Esc`), que el força encara que hàgiu desactivat el salvapantalles en repòs. No hi ha cap drecera assignada per defecte.

`lunaos toggle screensaver` és el que desactiva el de repòs, si preferiu anar directament de treballar a bloquejat. Necessita un terminal que sàpiga configurar — Alacritty, Foot, Ghostty o Kitty — i us ho dirà si el vostre terminal per defecte és una altra cosa.

El logotip que dibuixa el podeu canviar, a _Estil > Salvapantalles_. Pugeu un png o un svg i LunaOS el converteix en ASCII. Vegeu [marca](41-branding.md).

### La pantalla de bloqueig

`Super + Ctrl + L` bloqueja la màquina. Això executa la pantalla de bloqueig del shell d'LunaOS, apaga la pantalla, restableix la disposició del teclat a la primera perquè no escriviu la contrasenya en l'alfabet equivocat i — si el teniu en execució — bloqueja 1Password en sortir.

La pantalla de bloqueig accepta una contrasenya, i també acceptarà una empremta digital un cop n'hagueu configurat una. Això, i les altres maneres d'autenticar-se, es tracten a [autenticació de maquinari](37-hardware-authentication.md).
