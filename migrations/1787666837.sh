echo "Enable Dell XPS 13 sidecar speaker amplifiers"

if lunaos-hw-dell-xps13-sidecar-amps; then
  source "$LUNAOS_PATH/install/hardware/dell-xps13-sidecar-amps.sh"
  lunaos-state set reboot-required
fi
