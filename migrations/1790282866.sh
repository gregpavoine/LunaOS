echo "Apply the NVIDIA video driver fix on hybrid laptops"

if lunaos-hw-nvidia && ! lunaos-hw-nvidia-display; then
  # Hyprland applies env at startup. Reloading the updated defaults cannot
  # clear the old overrides inherited by the compositor and its applications.
  lunaos-state set reboot-required
  echo "Reboot to stop forcing NVIDIA VA-API/GLX drivers on the integrated GPU."
fi
