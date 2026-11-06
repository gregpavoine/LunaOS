#!/bin/bash

set -euo pipefail

source "$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)/base-test.sh"

# Browser and Steam installs, and Steam's removal, call the platform's own
# hooks through the real dispatcher: a Mac fixture with omarchy-mac's
# entrypoints, and x86.

require_platform_fixtures "app-install hooks on platform fixtures"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

fake_platform "$tmp/apple" aarch64-apple
fake_platform "$tmp/x86" x86

# The hooks record what they were called with and which flags files the
# browser install had written by then.
lifecycle=$tmp/lifecycle
mkdir -p "$lifecycle/usr/lib/omarchy/mac"
for operation in post-install pre-remove; do
  cat >"$lifecycle/usr/lib/omarchy/mac/$operation" <<SH
#!/bin/bash
printf '%s %s\n' "$operation" "\$*" >>"$tmp/calls"
ls "\$HOME/.config" 2>/dev/null | grep -- '-flags.conf$' >"$tmp/flags-at-hook" || true
[[ ! -e $tmp/fail ]]
SH
  chmod 755 "$lifecycle/usr/lib/omarchy/mac/$operation"
done
chmod -R go-w "$lifecycle"

stub_bin=$tmp/bin
lunaos_path=$tmp/lunaos
mkdir -p "$stub_bin" "$lunaos_path/install/helpers" "$lunaos_path/config"
printf 'browser_policy_setup_dir() { :; }\nbrowser_policy_setup_firefox_distribution() { :; }\nas_root() { "$@"; }\n' \
  >"$lunaos_path/install/helpers/browser-policy.sh"
cp "$ROOT/config/chromium-flags.conf" "$lunaos_path/config/"
for command in lunaos-pkg-add lunaos-pkg-aur-add lunaos-pkg-drop lunaos-install-gaming-gpu-lib32; do
  printf '#!/bin/bash\nprintf "%%s %%s\\n" "%s" "$*" >>"%s"\n' "$command" "$tmp/calls" >"$stub_bin/$command"
done
for command in lunaos-install-chromium-copy-url lunaos-install-chromium-ytdlp lunaos-theme-set-browser setsid; do
  printf '#!/bin/bash\nexit 0\n' >"$stub_bin/$command"
done
chmod +x "$stub_bin"/*

run() {
  local platform=$1
  shift
  rm -f "$tmp/calls" "$tmp/flags-at-hook"
  rm -rf "$tmp/home"
  mkdir -p "$tmp/home"
  HOME="$tmp/home" LUNAOS_PATH="$lunaos_path" LUNAOS_OPT_PATH="$tmp/opt" \
    LUNAOS_PROC_ROOT="$tmp/$platform/proc" LUNAOS_LIFECYCLE_ROOT="$lifecycle" \
    PATH="$stub_bin:$tmp/$platform/bin:$ROOT/bin:$PATH" bash "$@" >"$tmp/output" 2>&1
}

# ── browsers ─────────────────────────────────────────────────────────────────

declare -A flags=(
  [chromium]=chromium-flags.conf [chrome]=chrome-flags.conf [edge]=microsoft-edge-stable-flags.conf
  [brave]=brave-flags.conf [brave-origin]=brave-origin-flags.conf [firefox]="" [zen]=""
)
for browser in chromium chrome edge brave brave-origin firefox zen; do
  run apple "$ROOT/bin/lunaos-install-browser" "$browser" || fail "apple: $browser installs" "$(cat "$tmp/output")"
  [[ $(grep -c '^post-install ' "$tmp/calls") == 1 ]] && grep -qx "post-install $browser" "$tmp/calls" ||
    fail "apple: a $browser install runs post-install $browser once" "$(cat "$tmp/calls")"
  [[ $(cat "$tmp/flags-at-hook") == "${flags[$browser]}" ]] ||
    fail "apple: the $browser hook runs once its flags file is written" "$(cat "$tmp/flags-at-hook")"
  [[ $(grep -n . "$tmp/calls" | grep -m1 'post-install' | cut -d: -f1) -gt 1 ]] ||
    fail "apple: the $browser hook runs after its package step" "$(cat "$tmp/calls")"
done
pass "apple: every browser install runs post-install <browser> after its package and flags"

touch "$tmp/fail"
for browser in chromium brave; do
  if run apple "$ROOT/bin/lunaos-install-browser" "$browser"; then
    fail "apple: a failed $browser hook fails the install"
  fi
  ! grep -q 'browser installed' "$tmp/output" || fail "apple: a failed $browser hook is not announced as installed" "$(cat "$tmp/output")"
done
pass "apple: a failed browser hook fails the install"

for browser in chromium brave-origin zen; do
  run x86 "$ROOT/bin/lunaos-install-browser" "$browser" || fail "x86: $browser installs" "$(cat "$tmp/output")"
  ! grep -q '^post-install' "$tmp/calls" || fail "x86: a $browser install runs no hook" "$(cat "$tmp/calls")"
done
rm -f "$tmp/fail"
pass "x86: browser installs run no hook, even with Mac entrypoints on disk"

# ── Steam ────────────────────────────────────────────────────────────────────

run apple "$ROOT/bin/lunaos-install-gaming-steam" || fail "apple: Steam installs" "$(cat "$tmp/output")"
[[ $(cat "$tmp/calls") == $'lunaos-install-gaming-gpu-lib32 \nlunaos-pkg-add steam\npost-install steam' ]] ||
  fail "apple: Steam's hook runs after the 32-bit drivers and the package" "$(cat "$tmp/calls")"
run apple "$ROOT/bin/lunaos-remove-gaming-steam" || fail "apple: Steam is removed" "$(cat "$tmp/output")"
[[ $(head -n 2 "$tmp/calls") == $'pre-remove steam\nlunaos-pkg-drop steam' ]] ||
  fail "apple: Steam's pre-remove hook runs before the package goes" "$(cat "$tmp/calls")"
touch "$tmp/fail"
if run apple "$ROOT/bin/lunaos-install-gaming-steam"; then
  fail "apple: a failed Steam hook fails the install"
fi
if run apple "$ROOT/bin/lunaos-remove-gaming-steam"; then
  fail "apple: a failed pre-remove hook stops the removal"
fi
! grep -q '^lunaos-pkg-drop' "$tmp/calls" || fail "apple: a failed pre-remove hook removes nothing" "$(cat "$tmp/calls")"
pass "apple: Steam's install and removal run their hooks in order, and fail with them"

run x86 "$ROOT/bin/lunaos-install-gaming-steam" || fail "x86: Steam installs" "$(cat "$tmp/output")"
[[ $(cat "$tmp/calls") == $'lunaos-install-gaming-gpu-lib32 \nlunaos-pkg-add steam' ]] ||
  fail "x86: Steam installs as before" "$(cat "$tmp/calls")"
run x86 "$ROOT/bin/lunaos-remove-gaming-steam" || fail "x86: Steam is removed" "$(cat "$tmp/output")"
[[ $(cat "$tmp/calls") == "lunaos-pkg-drop steam" ]] || fail "x86: Steam is removed as before" "$(cat "$tmp/calls")"
rm -f "$tmp/fail"
pass "x86: Steam installs and is removed as before, running no hook"

# ── 32-bit drivers ───────────────────────────────────────────────────────────

# Without an Intel, AMD or NVIDIA GPU there is nothing to add, which must not
# fail the Steam install that asked.
gpu_bin=$tmp/gpu-bin
mkdir -p "$gpu_bin"
printf '#!/bin/bash\necho "00:02.0 VGA compatible controller: %s"\n' '${GPU:-Apple Inc. AGX}' >"$gpu_bin/lspci"
printf '#!/bin/bash\nexit "${ARCH_X86:-0}"\n' >"$gpu_bin/lunaos-hw-x86"
printf '#!/bin/bash\nexit 1\n' >"$gpu_bin/lunaos-hw-nvidia-gsp"
printf '#!/bin/bash\nexit 1\n' >"$gpu_bin/lunaos-hw-nvidia-without-gsp"
printf '#!/bin/bash\necho "lunaos-pkg-add $*" >>"%s"\n' "$tmp/calls" >"$gpu_bin/lunaos-pkg-add"
chmod +x "$gpu_bin"/*
rm -f "$tmp/calls"
PATH="$gpu_bin:$PATH" bash "$ROOT/bin/lunaos-install-gaming-gpu-lib32" >"$tmp/output" 2>&1 ||
  fail "no supported GPU is not a failure" "$(cat "$tmp/output")"
[[ ! -e $tmp/calls ]] || fail "no supported GPU installs nothing" "$(cat "$tmp/calls")"
GPU="Advanced Micro Devices, Inc. [AMD/ATI] Navi 31" PATH="$gpu_bin:$PATH" bash "$ROOT/bin/lunaos-install-gaming-gpu-lib32" >/dev/null 2>&1 ||
  fail "an AMD GPU installs its 32-bit driver"
[[ $(cat "$tmp/calls") == "lunaos-pkg-add lib32-vulkan-radeon" ]] || fail "an AMD GPU gets lib32-vulkan-radeon" "$(cat "$tmp/calls")"
pass "32-bit drivers: a machine with no supported GPU adds nothing without failing"

# Multilib only exists on x86_64; an ARM machine runs the native arm64 Steam.
rm -f "$tmp/calls"
ARCH_X86=1 GPU="Advanced Micro Devices, Inc. [AMD/ATI] Navi 31" PATH="$gpu_bin:$PATH" bash "$ROOT/bin/lunaos-install-gaming-gpu-lib32" >/dev/null 2>&1 ||
  fail "aarch64 skips the 32-bit drivers without failing"
[[ ! -e $tmp/calls ]] || fail "aarch64 installs no 32-bit driver" "$(cat "$tmp/calls")"
pass "32-bit drivers: aarch64 adds none, even with a supported GPU"
