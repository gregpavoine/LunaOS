echo "Install oh-my-pi (omp) via mise wrapper"

if [[ ! -f $HOME/.local/state/lunaos/preinstalls-removed ]]; then
  lunaos-mise-install github:can1357/oh-my-pi omp
fi
