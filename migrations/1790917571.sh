echo "Route ASUS ProArt P14 (NVIDIA N1x) speakers, headphones and microphones"

# lunaos-settings now ships a UCM profile for this model's MediaTek MT8901
# SoundWire card, which until now only gave PipeWire a dummy output. PipeWire
# reads UCM profiles when it opens the card, so restart it to pick it up.
if ! lunaos-hw-match "H7407BA"; then
  exit 0
fi

systemctl --user restart wireplumber pipewire pipewire-pulse 2>/dev/null || true
