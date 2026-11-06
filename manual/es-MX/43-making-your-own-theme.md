# Crear tu propio tema

Puedes agregar tus propios temas en `~/.config/lunaos/themes`. Solo copia uno de los existentes como base (busca en `/usr/share/lunaos/themes`) y luego ajústalo a tu gusto. Mientras tu tema esté dentro de esa carpeta, estará incluido en el menú de selección de temas.

El archivo principal que tienes que ajustar es `colors.toml`. Eso define el conjunto de colores que luego se usa para generar configuraciones para la terminal (Foot/Alacritty/Ghostty/Kitty), btop, Chromium, Hyprland, Neovim, Helix, VSCode, Obsidian y todo el shell de LunaOS (barra superior, menú, notificaciones, OSD y pantalla de bloqueo).

También puedes usar la aplicación incluida Aether para crear un tema nuevo usando una encantadora interfaz gráfica para jugar con los colores y buscar fondos. Solo iníciala desde la cuadrícula de aplicaciones con `Super + Alt + Space`.

### Lo que puede contener un tema instalado

Un tema que escribes tú mismo en `~/.config/lunaos/themes` puede contener lo que quieras: es tu computadora y tu archivo, y LunaOS lo aplica todo.

Un tema que instalas desde el repositorio de alguien más con `lunaos theme install` conserva todo lo que es color y pierde el puñado de archivos que ejecutarían código en tu computadora: cualquier archivo `.lua`, las configuraciones de la terminal (`alacritty.toml`, `foot.ini`, `ghostty.conf`, `kitty.conf`) y `vscode.json`. El `hyprland.lua` de un tema es Lua que tu compositor ejecuta al iniciar sesión, una configuración de terminal nombra el programa que tu terminal inicia, y `vscode.json` nombra una extensión de VSCode para instalar. Instalar el tema de alguien debería cambiar cómo se ve tu escritorio, nunca qué ejecuta.

Todo lo demás sigue funcionando exactamente como lo escribió el autor del tema: `btop.theme`, `chromium.theme`, `helix.toml`, `icons.theme`, `shell.toml`, los fondos y las vistas previas se conservan todos. Solo lo que se eliminó se regenera desde `colors.toml` en tu computadora.

LunaOS distingue a los dos por si el tema tiene su propio repositorio git dentro de él, que es lo que `lunaos theme install` deja atrás cuando clona. Así que un tema que escribiste sigue siendo tuyo, y uno que descargaste de internet se queda solo en colores.

### Modo claro

Si estás haciendo un tema en modo claro, pon `mode = "light"` en la parte superior de tu `colors.toml`. Entonces se emparejará automáticamente con el modo claro en todas las apps. (La forma anterior de dejar un archivo vacío llamado `light.mode` en la raíz de tu tema también sigue funcionando.)

### Colores de íconos

Si quieres que los íconos del administrador de archivos combinen en color con tu tema, agrega un archivo llamado `icons.theme` con el nombre del conjunto de íconos que quieres usar. Por defecto, las opciones son: `Yaru Yaru-blue Yaru-dark Yaru-magenta Yaru-olive Yaru-prussiangreen Yaru-purple Yaru-red Yaru-sage Yaru-wartybrown Yaru-yellow`.

### Imagen de desbloqueo

Los temas que traen las imágenes `unlock.png` y `preview-unlock.png` aparecerán en _Estilo > Desbloquear_. Tu `unlock.png` debería ser preferiblemente un png transparente. Y puedes crear la imagen de vista previa usando `lunaos plymouth preview`.

### Dar tema a apps que LunaOS no cubre

Si usas una app que no está en esa lista, puedes enseñarle a LunaOS a darle tema tú mismo con una plantilla. Deja un archivo en `~/.config/lunaos/themed/` con el nombre de la configuración que genera más una extensión `.tpl`, y escribe la configuración con `{{ background }}`, `{{ foreground }}`, `{{ accent }}`, `{{ red }}`, `{{ color0 }}` a `{{ color15 }}` y el resto de la paleta como marcadores. Cada vez que cambias de tema, el archivo se regenera con los colores de ese tema.

Hay un `alacritty.toml.tpl.sample` totalmente comentado en esa carpeta para copiar: enumera cada variable que puedes usar, más los modificadores `_strip` y `_rgb` para apps que quieren sus colores sin el `#` o como RGB decimal. Tus plantillas tienen prioridad sobre las propias de LunaOS, así que también puedes usar esto para invalidar cómo se le da tema a una app integrada.

### Distribuir tu tema

Si quieres distribuir tu tema para que otros puedan usarlo, tienes que ponerlo en un servidor git público, como GitHub. Entonces la gente podrá instalarlo usando _Instalar > Estilo > Tema_ en el menú de LunaOS con esa URL. Se recomienda que sigas la convención de nombres de `lunaos-[themename]-theme`, ya que el tema se mostrará correctamente solo como `[themename]` en el menú de selección de temas después de la instalación.

Ese `[themename]` restante se convierte en el nombre del directorio del tema, así que tiene que ser uno que LunaOS pueda pasar de forma segura: debe empezar con una letra, un dígito o un guion bajo, y el resto puede llevar letras, dígitos, `.`, `_`, `+` y `-`. Las mayúsculas se convierten a minúsculas por ti, pero cualquier otra cosa — un espacio, una comilla, un carácter que no sea del inglés — se rechaza al instalar en lugar de convertirse en un nombre de directorio. Así que `lunaos-tokyo-night-theme`, `lunaos-flexoki_light-theme` y `lunaos-c++-theme` se instalan sin problema.

Recuerda que una vez instalado desde un repositorio, cualquier `.lua`, configuración de terminal o `vscode.json` que traiga se descarta, así que no construyas el tema alrededor de eso.

Puedes hacer que tu tema se agregue a [la página de temas extra](https://omarchy.org/themes/) enviando un pull request al [repositorio omarchy-site](https://github.com/omacom-io/omarchy-site).
