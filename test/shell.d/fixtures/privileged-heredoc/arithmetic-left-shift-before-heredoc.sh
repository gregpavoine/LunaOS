mask=$((1 << bits))

cat >/etc/lunaos/agent.conf <<EOF
helper=$HOME/.local/share/lunaos/bin/lunaos-agent
EOF
