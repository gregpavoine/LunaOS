cat <<EOF | sudo tee /etc/udev/rules.d/99-lunaos.rules >/dev/null
SUBSYSTEM=="power_supply", ATTR{type}=="Mains", RUN+="/usr/bin/lunaos-powerprofiles-set"
EOF
