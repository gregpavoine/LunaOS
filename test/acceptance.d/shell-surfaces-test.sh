#!/bin/bash

set -euo pipefail

source "$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)/base-test.sh"

open_and_close() {
  local name="$1" plugin="$2" namespace="$3" payload="${4:-}"

  if [[ -n $payload ]]; then
    lunaos-shell shell summon "$plugin" "$payload" >/dev/null
  else
    lunaos-shell shell summon "$plugin" >/dev/null
  fi

  wait_until "$name opens" 15 layer_present "$namespace"
  sleep 1
  screenshot "success-$name-open"

  lunaos-shell shell hide "$plugin" >/dev/null
  wait_until "$name closes" 15 layer_absent "$namespace"
}

# Search and select an emoji. The host harness separately proves the shortcut
# with a QMP hardware key chord, while this test focuses on UI behavior.
lunaos-shell shell summon lunaos.emojis >/dev/null
wait_until "emoji picker opens" 15 layer_on_overlay "lunaos-emojis"
wtype "rocket"
sleep 1
screenshot "success-emoji-picker-search"
wtype -k Return
wait_until "emoji picker selection closes" 15 layer_absent "lunaos-emojis"

# Seed two clipboard entries, search for the older one, and copy it back out.
clipboard_token="LunaOS acceptance clipboard $(date +%s)"
printf '%s' "$clipboard_token" | wl-copy
wait_until "clipboard history captures test text" 15 grep -Fq "$clipboard_token" "$HOME/.local/state/lunaos/clipboard-history.json"
printf '%s' "clipboard decoy" | wl-copy
sleep 1

lunaos-shell shell summon lunaos.clipboard >/dev/null
wait_until "clipboard opens" 15 layer_on_overlay "lunaos-clipboard"
wtype "$clipboard_token"
wait_until "clipboard search finds test text" 15 screen_contains "LunaOS acceptance clipboard"
screenshot "success-clipboard-search"
wtype -M shift -k Return -m shift
wait_until "clipboard selection closes" 15 layer_absent "lunaos-clipboard"
wait_until "clipboard selection restores test text" 15 bash -c '[[ $(wl-paste --no-newline) == "$1" ]]' _ "$clipboard_token"

# Exercise the system branch without invoking any destructive action.
lunaos-shell shell summon lunaos.menu '{"menu":"system"}' >/dev/null
wait_until "system menu opens" 15 layer_on_overlay "lunaos-menu"
wait_until "system menu content is visible" 15 screen_contains "Shutdown"
screenshot "success-system-menu"
wtype -k Escape
wait_until "system menu closes" 15 layer_absent "lunaos-menu"

# Preview both visual selectors and cancel without changing user state. These
# cover thumbnail generation, the image-grid overlay, and current selection.
launch_app "lunaos-theme-bg-switcher"
wait_until "background selector opens" 30 layer_on_overlay "lunaos-image-selector"
sleep 1
screenshot "success-background-selector"
wtype -k Escape
wait_until "background selector closes" 15 layer_absent "lunaos-image-selector"

launch_app "lunaos-theme-switcher"
wait_until "theme selector opens" 30 layer_on_overlay "lunaos-image-selector"
sleep 1
screenshot "success-theme-selector"
wtype -k Escape
wait_until "theme selector closes" 15 layer_absent "lunaos-image-selector"

# Walk the reminder flow through each input screen, but dismiss before it
# schedules a real timer in the test user's session.
lunaos-shell shell summon lunaos.reminders >/dev/null
wait_until "reminder flow opens" 15 layer_on_overlay "lunaos-reminders"
screenshot "success-reminder-01-minutes-prompt"
wtype "5"
sleep 1
screenshot "success-reminder-02-minutes-entered"
wtype -k Return
wait_until "reminder message prompt opens" 15 screen_contains "Reminder message"
screenshot "success-reminder-03-message-prompt"
wtype -k Escape
wait_until "reminder flow closes" 15 layer_absent "lunaos-reminders"

# Render a real shell notification and clear it through the notification IPC.
lunaos-shell notifications dismissAll >/dev/null
lunaos-notification-send "Acceptance notification" "Shell notification rendering" --expire-time=15000
wait_until "notification popup opens" 15 layer_present "lunaos-notifications"
wait_until "notification content is visible" 15 screen_contains "Acceptance notification"
screenshot "success-notification-popup"
lunaos-shell notifications dismissAll >/dev/null
# The notification surface stays mapped while idle, so wait for the count.
notifications_cleared() { [[ $(lunaos-shell notifications count) == "0" ]]; }
wait_until "notification popup closes" 15 notifications_cleared

# The menu's Apps submenu does the full launcher loop: open, search by
# typing, launch the top hit.
if window_present "(?i)omawrite" >/dev/null 2>&1; then
  fail "app launch test starts with no Omawrite window" "an Omawrite window is already open"
fi

lunaos-menu summon apps >/dev/null
wait_until "apps menu opens" 15 layer_on_overlay "lunaos-menu"
sleep 1
screenshot "success-apps-menu-open"

wtype "omawrite"
sleep 1
screenshot "success-apps-menu-search"
wtype -k Return

wait_until "apps menu launches the top search hit" 60 window_present "(?i)omawrite"
wait_until "apps menu closes after launching" 15 layer_absent "lunaos-menu"

close_windows "(?i)omawrite"
wait_until "Omawrite window closes" 30 window_absent "(?i)omawrite"
