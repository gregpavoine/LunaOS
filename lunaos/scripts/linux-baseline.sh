#!/bin/bash
# READ-ONLY TARGET INVENTORY. AUTHORED BUT NOT EXECUTED BY TASK 1.
# Execute only on the authorized Linux host after explicit owner authorization.
# Output may contain machine names/mount paths. Review locally before committing.
set -uo pipefail
if [[ "$(uname -s)" != Linux ]]; then
  echo 'Linux target required; no operations performed.' >&2
  exit 2
fi
probe() {
  local label="$1"
  shift
  printf '\n## %s\n' "$label"
  if command -v "$1" >/dev/null 2>&1; then
    "$@" || printf 'UNAVAILABLE: probe failed (%s)\n' "$label"
  else
    printf 'UNAVAILABLE: command missing (%s)\n' "$1"
  fi
}
probe 'Host (review identifiers before sharing)' hostnamectl
probe 'Installed packages' pacman -Q
probe 'GPU inventory' nvidia-smi --query-gpu=name,driver_version,memory.total --format=csv,noheader
probe 'Running system services' systemctl --no-pager --plain list-units --type=service --state=running
probe 'Running user services' systemctl --user --no-pager --plain list-units --type=service --state=running
probe 'Mounted filesystems (review paths)' findmnt --list --output TARGET,FSTYPE,OPTIONS
probe 'Network links' ip -brief link
probe 'Kernel' uname -r
if [[ -n "${LUNA_COMFYUI_ROOT:-}" ]]; then
  if [[ -d "$LUNA_COMFYUI_ROOT" ]]; then
    printf '\n## Explicitly supplied ComfyUI directory\n'
    for entry in main.py launch.sh requirements.txt; do
      [[ ! -f "$LUNA_COMFYUI_ROOT/$entry" ]] || printf 'PRESENT: %s\n' "$entry"
    done
  else
    printf '\nComfyUI: supplied path unavailable\n'
  fi
else
  printf '\nComfyUI: path not supplied; not guessed, scanned or launched\n'
fi
# No sudo, installs, remote calls, service control, mounts or file writes.
