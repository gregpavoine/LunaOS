sudo dd status=none of=/etc/lunaos/boot.conf <<EOF
cmdline=$boot_params
EOF
