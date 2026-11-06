echo "Install cf (Cloudflare CLI) via mise wrapper"

if [[ ! -f $HOME/.local/state/lunaos/preinstalls-removed ]]; then
  lunaos-mise-install npm:cf cf
fi
