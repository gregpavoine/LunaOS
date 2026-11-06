echo "Install Gliff on aarch64, now that LunaOS's repositories build it there"

if lunaos-hw-aarch64; then
  lunaos-pkg-add gliff
fi
