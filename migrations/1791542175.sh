echo "Install Gliff, the Hyprland remote desktop over SSH"

# Gliff has no aarch64 build (install/lunaos-x86_64-only.packages).
if lunaos-hw-x86; then
  lunaos-pkg-add gliff
fi
