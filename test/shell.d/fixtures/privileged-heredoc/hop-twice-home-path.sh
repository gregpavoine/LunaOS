#!/bin/bash

# Two hops. The scan resolves lunaos_bin into helper, so the value it ends up
# judging still carries an unresolved $HOME rather than a literal path.
lunaos_bin="$HOME/.local/share/lunaos/bin"
helper="$lunaos_bin/lunaos-agent"

# lunaos:heredoc-expands paths=none -- helper names the agent, no path is baked in
cat <<EOF | sudo tee /etc/udev/rules.d/99-lunaos-agent.rules >/dev/null
SUBSYSTEM=="power_supply", RUN+="$helper"
EOF
