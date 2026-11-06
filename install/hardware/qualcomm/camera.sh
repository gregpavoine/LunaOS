# Snapdragon cameras run through libcamera, which apps reach through
# PipeWire's libcamera backend.

if lunaos-hw-aarch64-qualcomm; then
  lunaos-pkg-add pipewire-libcamera
fi
