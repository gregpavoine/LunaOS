echo "Give the N1x swap on zram, the memory tuning and realtime audio scheduling"

# lunaos-settings now ships the x86_64 memory stack on N1x too (see
# default/settings-runtime-profile): zram, zswap off, the reclaim sysctls and
# systemd-oomd's slice policy. Installs from the earlier N1x images ran with no
# swap at all, and without rtkit, which nothing pulls in on aarch64, so
# PipeWire got no realtime scheduling.
if ! lunaos-hw-aarch64-n1x; then
  exit 0
fi

lunaos-pkg-add rtkit zram-generator

sudo sysctl --system >/dev/null
sudo systemctl try-restart systemd-oomd.service

# Installing the generator's configuration while swap.target is already active
# leaves dev-zram0.swap generated but stopped until the next boot.
if [[ -f /usr/lib/systemd/zram-generator.conf.d/90-lunaos.conf ]] &&
  ! systemctl is-active --quiet dev-zram0.swap; then
  sudo systemctl daemon-reload
  sudo systemctl start dev-zram0.swap
fi
