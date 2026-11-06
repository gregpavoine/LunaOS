echo "Let Chromium use the cameras of Snapdragon laptops through PipeWire"

if lunaos-hw-aarch64-qualcomm; then
  source "$LUNAOS_PATH/install/hardware/qualcomm/camera.sh"
  lunaos-install-camera

  echo "Log out and back in to load the PipeWire camera backend."
fi
