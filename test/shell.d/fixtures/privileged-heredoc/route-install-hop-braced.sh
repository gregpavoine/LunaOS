tmp=/tmp/lunaos-generated
cat >"$tmp" <<EOF
command=$HOME/.local/share/lunaos/bin/example
EOF
sudo install -m644 "${tmp}" /etc/lunaos/example.conf
