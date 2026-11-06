echo "Put Elsewhen, the world clock, on the bar"

# Best-effort, like the put below: an update whose shell cannot be asked still
# finishes, and restarts the shell once the migrations are through.
lunaos-shell -q shell rescanPlugins

# A legacy omacom.elsewhen entry is renamed in place by a later migration.
config_file="$HOME/.config/lunaos/shell.json"
if [[ -s $config_file ]] && jq -e '[.bar.layout[]?[]? | if type == "object" then .id else . end] | index("omacom.elsewhen")' "$config_file" >/dev/null 2>&1; then
  echo "Elsewhen is already on the bar"
else
  lunaos-bar put lunaos.elsewhen --after lunaos.clock
fi
