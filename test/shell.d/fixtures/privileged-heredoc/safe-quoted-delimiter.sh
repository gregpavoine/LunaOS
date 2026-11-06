cat <<'EOF' | sudo tee /etc/udev/rules.d/99-lunaos.rules >/dev/null
SUBSYSTEM=="power_supply", RUN+="/usr/bin/lunaos-powerprofiles-set $HOME"
EOF

cat <<"XML" | sudo tee /etc/lunaos/agent.xml >/dev/null
<config path="$HOME/.local/share/lunaos" />
XML
