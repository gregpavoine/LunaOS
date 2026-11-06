tmp=$(mktemp)

cat >"$tmp" <<EOF
#!/bin/bash
exec "$HOME/.local/share/lunaos/bin/lunaos-agent" "$@"
EOF

sudo install -m 0755 "$tmp" /usr/local/bin/lunaos-agent-shim
