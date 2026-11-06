mkdir -p ~/.config/lunaos

cat >~/.config/lunaos/agent.conf <<EOF
helper=$HOME/.local/share/lunaos/bin/lunaos-agent
EOF

cat >"$HOME/.local/bin/lunaos-shim" <<EOF
exec "$LUNAOS_PATH/bin/lunaos-agent" "$@"
EOF
