#!/usr/bin/env bash
# LunaOS V0 / Series 26 — conservative Omarchy package cleaner.
# Default: read-only audit. Explicit --apply + typed confirmation required.
set -Eeuo pipefail
export LC_ALL=C
MODE=audit
if [[ ${1:-} == --apply && $# -eq 1 ]]; then MODE=apply; elif [[ $# -ne 0 && ${1:-} != --audit ]]; then echo 'Usage: ./lunaos-v0-cleanup.sh [--audit|--apply]' >&2; exit 2; fi
for command in pacman uname sed grep awk date; do command -v "$command" >/dev/null || { echo "Missing: $command" >&2; exit 2; }; done
[[ $(uname -s) == Linux ]] || { echo 'Linux required' >&2; exit 2; }
[[ -r /etc/arch-release ]] || { echo 'Not Arch-based; refusing' >&2; exit 2; }
if ! command -v omarchy >/dev/null 2>&1 && [[ ! -d /usr/share/omarchy && ! -d "$HOME/.local/share/omarchy" ]]; then
  echo 'Omarchy signature not found; refusing to alter generic Arch.' >&2; exit 2
fi
# Exact package identifiers only. Never remove whole groups or recursive orphans.
# Alternate names are checked against local pacman DB; only installed names are selected.
CANDIDATES=(
  libreoffice-fresh libreoffice-still libreoffice-fresh-fr libreoffice-still-fr
  kdenlive omacut pinta xournalpp moonlight-qt moonlight omacalc
)
# Meta packages may have dependencies; pacman -R (not -Rs/-Rns) intentionally blocks unsafe removal.
PROTECTED=(hyprland linux linux-lts linux-zen linux-firmware systemd pacman sudo networkmanager pipewire wireplumber openssh limine grub efibootmgr btrfs-progs snapper omarchy omarchy-shell quickshell foot chromium obs-studio obsidian omawrite cliamp localsend neovim docker lazydocker gnome-disk-utility)
contains() { local needle=$1 x; shift; for x in "$@"; do [[ "$x" == "$needle" ]] && return 0; done; return 1; }
installed=(); for pkg in "${CANDIDATES[@]}"; do if pacman -Qq "$pkg" >/dev/null 2>&1; then if contains "$pkg" "${PROTECTED[@]}"; then echo "PROTECTED $pkg" >&2; exit 3; fi; installed+=("$pkg"); fi; done
printf 'Omarchy identity: %s\n' "$(command -v omarchy || echo detected-via-directory)"
printf 'Kernel: %s\n' "$(uname -r)"
if command -v omarchy >/dev/null 2>&1; then omarchy --version 2>/dev/null || true; fi
printf 'Mode: %s\n' "$MODE"
printf 'Removal candidates found (%d):\n' "${#installed[@]}"
for pkg in "${installed[@]}"; do printf '  - %s\n' "$pkg"; done
cat <<'NOTE'
PRESERVED: Omarchy menu/launcher, all existing web apps, OBS Studio, Obsidian,
Omawrite, Chromium, LocalSend, Cliamp, Neovim, Foot, Docker tools, disk tools,
AI launchers, clipboard, notifications, screen capture/recording/dictation,
Hyprland, network, GPU drivers, boot/recovery and Omarchy updater back end.
NOT AUTOMATICALLY MODIFIED: wallpapers, theming, Quickshell bar, screen lock,
key bindings and updater UI. Claude Code must replace those AFTER smoke tests.
Do not use Omarchy 'Remove > Preinstalls': this can remove retained programs
and custom web app/TUI launchers.
NOTE
if ((${#installed[@]} == 0)); then echo 'No matching installed packages; done.'; exit 0; fi
# Use pacman's own transaction resolver to reject removals that would break dependencies.
if ! pacman -R --print -- "${installed[@]}"; then
  echo 'Pacman rejects this exact transaction. No changes made. Resolve manually; do not force or cascade.' >&2
  exit 4
fi
if [[ $MODE == audit ]]; then echo 'AUDIT ONLY — to apply: ./lunaos-v0-cleanup.sh --apply'; exit 0; fi
[[ -t 0 ]] || { echo 'Interactive terminal required to apply.' >&2; exit 5; }
[[ $EUID -ne 0 ]] || { echo 'Run as normal user, not directly as root.' >&2; exit 5; }
read -r -p 'Type REMOVE OPTIONAL APPS to confirm (anything else cancels): ' confirm
[[ $confirm == 'REMOVE OPTIONAL APPS' ]] || { echo 'Cancelled.'; exit 0; }
# No --noconfirm; pacman still shows its own final transaction confirmation.
sudo pacman -R -- "${installed[@]}"
echo 'Package removal finished. Verify app launchers and reboot recovery before replacing UI components.'
