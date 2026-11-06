echo "Normalize Snapper snapshot services"

LUNAOS_PATH="${LUNAOS_PATH:-/usr/share/lunaos}"
snapper_config_script=/usr/share/lunaos/install/config/snapper.sh
if [[ ! -f $snapper_config_script ]]; then
  snapper_config_script="$LUNAOS_PATH/install/config/snapper.sh"
fi

as_root() {
  if (( EUID == 0 )); then
    "$@"
  else
    sudo "$@"
  fi
}

unit_enabled() {
  systemctl is-enabled --quiet "$1" >/dev/null 2>&1
}

unit_active() {
  systemctl is-active --quiet "$1" >/dev/null 2>&1
}

needs_repair=0

[[ -f /etc/snapper/configs/root ]] || needs_repair=1

if ! unit_enabled snapper-cleanup.timer || ! unit_active snapper-cleanup.timer; then
  needs_repair=1
fi

if ! unit_enabled limine-snapper-sync.service || ! unit_active limine-snapper-sync.service; then
  needs_repair=1
fi

(( needs_repair )) || exit 0

as_root env LUNAOS_PATH="$LUNAOS_PATH" bash -euo pipefail "$snapper_config_script"
