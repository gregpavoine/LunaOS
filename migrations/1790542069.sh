echo "Install Hype, the Markdown presentation app"

if [[ ! -f $HOME/.local/state/lunaos/preinstalls-removed ]]; then
  lunaos-pkg-add hype
fi
