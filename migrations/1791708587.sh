echo "Install Signed, the PDF signing app"

if [[ ! -f $HOME/.local/state/lunaos/preinstalls-removed ]]; then
  lunaos-pkg-add signed
fi
