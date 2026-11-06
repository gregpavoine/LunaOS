echo "Install hey (hey-cli) via mise wrapper"

if [[ ! -f $HOME/.local/state/lunaos/preinstalls-removed ]]; then
  lunaos-mise-install github:basecamp/hey-cli hey
fi
