echo "Install ori (OpenRouter's agent harness) via mise wrapper"

if [[ ! -f $HOME/.local/state/lunaos/preinstalls-removed ]]; then
  lunaos-mise-install github:OpenRouterLabs/ori-releases ori
fi
