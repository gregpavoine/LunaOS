tmp=/tmp/lunaos-generated
copy=$tmp
cat >"$tmp" <<EOF
command=$HOME/.local/share/lunaos/bin/example
EOF
sudo install -m644 "$copy" /etc/lunaos/example.conf
