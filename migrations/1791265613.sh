echo "Install the Dell XPS 13 Panther Lake speaker firmware"

if lunaos-hw-dell-xps13-dx13260-ptl; then
  source "$LUNAOS_PATH/install/hardware/dell-xps13-ptl-speaker-firmware.sh"
  lunaos-state set reboot-required
fi
