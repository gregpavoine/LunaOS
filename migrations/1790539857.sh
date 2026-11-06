echo "Install Monologue, the webcam recorder"

if [[ ! -f $HOME/.local/state/lunaos/preinstalls-removed ]]; then
  lunaos-pkg-add monologue
fi
