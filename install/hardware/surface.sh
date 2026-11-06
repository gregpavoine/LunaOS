# The Marvell Wi-Fi firmware is for x86 Surfaces only; Snapdragon Surfaces use Qualcomm Wi-Fi.
if lunaos-hw-surface && lunaos-hw-x86; then
  lunaos-pkg-add linux-firmware-marvell
fi
