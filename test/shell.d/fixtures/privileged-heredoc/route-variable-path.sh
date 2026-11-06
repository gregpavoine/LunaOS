DROP_IN=/etc/systemd/system/lunaos-agent.service.d/override.conf

cat <<EOF | sudo tee "$DROP_IN" >/dev/null
[Service]
ExecStart=$LUNAOS_PATH/bin/lunaos-agent
EOF
