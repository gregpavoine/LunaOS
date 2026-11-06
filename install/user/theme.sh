# Setup user theme folder and seed the default only when no theme exists yet.
mkdir -p ~/.config/lunaos/themes

if [[ ! -s $HOME/.local/state/lunaos/current/theme.name ]]; then
  # iso-chroot and provision-owner both run without a live session to notify.
  if [[ ${LUNAOS_SETUP_CONTEXT:-runtime} != "runtime" ]]; then
    LUNAOS_THEME_HEADLESS=1 lunaos-theme-set "Tokyo Night"
    rm -f ~/.config/chromium/SingletonLock # otherwise archiso owns the Chromium singleton
  else
    lunaos-theme-set "Tokyo Night"
  fi
fi
lunaos-theme-set-pi --activate
lunaos-theme-set-fcitx5 --activate

mkdir -p ~/.config/btop/themes
ln -snf "$HOME/.local/state/lunaos/current/theme/btop.theme" ~/.config/btop/themes/current.theme
