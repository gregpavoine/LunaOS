echo "Theme the Fcitx candidate window with the current LunaOS theme"

# Activation selects the theme; the refresh then renders it, since the current
# theme was staged before its template existed.
lunaos-theme-set-fcitx5 --activate
lunaos-theme-refresh
