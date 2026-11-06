echo "Install Kimi coding agent mise wrapper"

if [[ ! -f $HOME/.local/state/lunaos/preinstalls-removed ]] && lunaos-cmd-missing kimi; then
  lunaos-mise-install npm:@moonshot-ai/kimi-code kimi
fi
