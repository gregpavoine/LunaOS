#!/bin/bash

set -euo pipefail

source "$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)/base-test.sh"

source "$SHELL_TEST_DIR/fixtures/sudo-boundary-test.sh"
test_tmp="$boundary_tmp"
package_root="$SUDO_TEST_ROOT"
copy_boundary_file bin/lunaos-channel-set
python3 - "$SUDO_TEST_ROOT/bin/lunaos-channel-set" "$package_root" <<'PYTHON'
import sys
from pathlib import Path
p = Path(sys.argv[1])
p.write_text(p.read_text().replace("/usr/share/lunaos", sys.argv[2]))
PYTHON

stub_bin="$SUDO_TEST_ROOT/bin"
log_file="$test_tmp/channel.log"
mkdir -p "$stub_bin" "$test_tmp/home"

write_stub() {
  local name="$1"
  local body="$2"

  rm -f "$stub_bin/$name"
  cat >"$stub_bin/$name" <<<"$body"
  chmod +x "$stub_bin/$name"
}

write_stub lunaos-refresh-pacman '#!/bin/bash
printf "refresh" >>"$LUNAOS_CHANNEL_TEST_LOG"
for arg in "$@"; do printf "\t%s" "$arg" >>"$LUNAOS_CHANNEL_TEST_LOG"; done
printf "\n" >>"$LUNAOS_CHANNEL_TEST_LOG"
'

write_stub sudo '#!/bin/bash
case "${1:-}" in
  -h) echo "usage: sudo [-ABbEHkNnPS] command"; exit 0 ;;
  -k|-K) exit 0 ;;
esac
printf "sudo" >>"$LUNAOS_CHANNEL_TEST_LOG"
for arg in "$@"; do printf "\t%s" "$arg" >>"$LUNAOS_CHANNEL_TEST_LOG"; done
printf "\n" >>"$LUNAOS_CHANNEL_TEST_LOG"
'

cp "$stub_bin/sudo" "$SUDO_TEST_ROOT/mock/sudo"

write_stub lunaos-update-pacman '#!/bin/bash
printf "update-pacman" >>"$LUNAOS_CHANNEL_TEST_LOG"
for arg in "$@"; do printf "\t%s" "$arg" >>"$LUNAOS_CHANNEL_TEST_LOG"; done
printf "\n" >>"$LUNAOS_CHANNEL_TEST_LOG"
'

write_stub lunaos-dev-unlink '#!/bin/bash
printf "unlink" >>"$LUNAOS_CHANNEL_TEST_LOG"
for arg in "$@"; do printf "\t%s" "$arg" >>"$LUNAOS_CHANNEL_TEST_LOG"; done
printf "\n" >>"$LUNAOS_CHANNEL_TEST_LOG"
'

write_stub lunaos-state '#!/bin/bash
printf "state" >>"$LUNAOS_CHANNEL_TEST_LOG"
for arg in "$@"; do printf "\t%s" "$arg" >>"$LUNAOS_CHANNEL_TEST_LOG"; done
printf "\n" >>"$LUNAOS_CHANNEL_TEST_LOG"
'

write_stub lunaos-update '#!/bin/bash
printf "update" >>"$LUNAOS_CHANNEL_TEST_LOG"
for arg in "$@"; do printf "\t%s" "$arg" >>"$LUNAOS_CHANNEL_TEST_LOG"; done
printf "\tLUNAOS_PATH=%s" "$LUNAOS_PATH" >>"$LUNAOS_CHANNEL_TEST_LOG"
printf "\n" >>"$LUNAOS_CHANNEL_TEST_LOG"
'

write_stub gum '#!/bin/bash
printf "gum" >>"$LUNAOS_CHANNEL_TEST_LOG"
for arg in "$@"; do printf "\t%s" "$arg" >>"$LUNAOS_CHANNEL_TEST_LOG"; done
printf "\n" >>"$LUNAOS_CHANNEL_TEST_LOG"
exit 0
'

write_stub git '#!/bin/bash
printf "git" >>"$LUNAOS_CHANNEL_TEST_LOG"
for arg in "$@"; do printf "\t%s" "$arg" >>"$LUNAOS_CHANNEL_TEST_LOG"; done
printf "\n" >>"$LUNAOS_CHANNEL_TEST_LOG"
if [[ $1 == "clone" ]]; then
  dest="${@: -1}"
  /usr/bin/cp -a "$SUDO_TEST_ROOT" "$dest"
  mkdir -p "$dest/.git" "$dest/shell"
fi
'

write_stub lunaos-dev-link '#!/bin/bash
printf "link" >>"$LUNAOS_CHANNEL_TEST_LOG"
for arg in "$@"; do printf "\t%s" "$arg" >>"$LUNAOS_CHANNEL_TEST_LOG"; done
printf "\n" >>"$LUNAOS_CHANNEL_TEST_LOG"
'

write_stub lunaos-version-channel '#!/bin/bash
printf "%s\n" "${LUNAOS_TEST_VERSION_CHANNEL:-unknown}"
'

write_stub pacman '#!/bin/bash
[[ $1 == "-Q" ]] || exit 1
shift
case "${LUNAOS_TEST_PACKAGES:-}" in
  stable) [[ $* == "lunaos lunaos-settings" ]] ;;
  dev) [[ $* == "lunaos-dev lunaos-settings-dev" ]] ;;
  *) exit 1 ;;
esac
'

run_channel() {
  : >"$log_file"
  LUNAOS_CHANNEL_TEST_LOG="$log_file" \
    LUNAOS_PATH="${LUNAOS_TEST_PATH:-$package_root}" \
    LUNAOS_REGION_FILE="${LUNAOS_TEST_REGION_FILE:-$test_tmp/no-region}" \
    HOME="$test_tmp/home" \
    PATH="$stub_bin:$ROOT/bin:$PATH" \
    "${LUNAOS_TEST_PATH:-$package_root}/bin/lunaos-channel-set" "$@"
}

assert_log_line() {
  local expected="$1"
  local description="$2"

  grep -Fx -- "$expected" "$log_file" >/dev/null || fail "$description" "$(cat "$log_file")"
  pass "$description"
}

run_channel stable
assert_log_line $'refresh\tstable' "stable refreshes the stable pacman channel"
assert_log_line $'update-pacman\t-S\t--needed\t--noconfirm\t--ask\t4\tlunaos\tlunaos-settings' "stable installs stable LunaOS packages"
assert_log_line $'unlink\t--no-reboot' "stable restores the package-backed LunaOS path without an early reboot prompt"
assert_log_line $'update\t-y\tLUNAOS_PATH='"$package_root" "stable runs the normal update pipeline from the package-backed path"
if grep -q $'^state\tset\treboot-required$' "$log_file"; then
  fail "stable does not require reboot when already package-backed" "$(cat "$log_file")"
fi
pass "stable does not require reboot when already package-backed"

run_channel rc
assert_log_line $'refresh\trc' "rc refreshes the rc pacman channel"
assert_log_line $'update-pacman\t-S\t--needed\t--noconfirm\t--ask\t4\tlunaos\tlunaos-settings' "rc installs rc LunaOS packages"
assert_log_line $'unlink\t--no-reboot' "rc restores the package-backed LunaOS path without an early reboot prompt"
assert_log_line $'update\t-y\tLUNAOS_PATH='"$package_root" "rc runs the normal update pipeline from the package-backed path"

active_checkout="$test_tmp/active-checkout"
cp -a "$package_root" "$active_checkout"
LUNAOS_TEST_PATH="$active_checkout" run_channel edge
assert_log_line $'refresh\tedge' "edge refreshes the edge pacman channel"
assert_log_line $'update-pacman\t-S\t--needed\t--noconfirm\t--ask\t4\tlunaos-dev\tlunaos-settings-dev' "edge installs development LunaOS packages"
assert_log_line $'unlink\t--no-reboot' "edge unlinks dev without an early reboot prompt"
assert_log_line $'state\tset\treboot-required' "edge marks reboot required when leaving dev"
assert_log_line $'update\t-y\tLUNAOS_PATH='"$package_root" "edge runs the normal update pipeline from the package-backed path"
[[ $(grep -E $'^(unlink|state|update)\t' "$log_file") == $'unlink\t--no-reboot\nstate\tset\treboot-required\nupdate\t-y\tLUNAOS_PATH='"$package_root" ]] ||
  fail "edge defers the reboot prompt until the update restart stage" "$(cat "$log_file")"
pass "edge defers the reboot prompt until the update restart stage"

checkout="$test_tmp/home/lunaos"
mkdir -p "$checkout"
if run_channel dev >"$test_tmp/occupied.out" 2>"$test_tmp/occupied.err"; then
  fail "dev refuses to use an occupied non-checkout path"
fi

grep -q "already exists and is not a git checkout" "$test_tmp/occupied.err" || fail "dev explains occupied checkout paths" "$(cat "$test_tmp/occupied.err")"
if grep -Fx $'refresh\tedge' "$log_file" >/dev/null; then
  fail "dev validates checkout path before changing packages" "$(cat "$log_file")"
fi
pass "dev refuses occupied non-checkout paths before package changes"

rmdir "$checkout"
run_channel dev
assert_log_line $'gum\tconfirm\t--default=false\tSwitch to dev channel?' "dev asks for confirmation"
assert_log_line $'refresh\tedge' "dev refreshes the edge pacman channel"
assert_log_line $'update-pacman\t-S\t--needed\t--noconfirm\t--ask\t4\tlunaos-dev\tlunaos-settings-dev' "dev installs development LunaOS packages"
assert_log_line $'git\tclone\t--filter=blob:none\thttps://github.com/gregpavoine/LunaOS.git\t'"$checkout" "dev clones the source checkout to ~/lunaos"
assert_log_line $'link\t'"$checkout"$'\t--no-reboot' "dev links ~/lunaos without an early reboot prompt"
assert_log_line $'state\tset\treboot-required' "dev defers the reboot prompt to the update pipeline"
assert_log_line $'update\t-y\tLUNAOS_PATH='"$checkout" "dev runs the normal update pipeline from the source checkout"
[[ $(grep -E '^(git|link|state|refresh|sudo|update)' "$log_file") == $'git\tclone\t--filter=blob:none\thttps://github.com/gregpavoine/LunaOS.git\t'"$checkout"$'\nlink\t'"$checkout"$'\t--no-reboot\nstate\tset\treboot-required\nrefresh\tedge\nupdate-pacman\t-S\t--needed\t--noconfirm\t--ask\t4\tlunaos-dev\tlunaos-settings-dev\nupdate\t-y\tLUNAOS_PATH='"$checkout" ]] ||
  fail "dev activates the checkout before changing or updating packages" "$(cat "$log_file")"
pass "dev activates the checkout before changing or updating packages"

LUNAOS_TEST_PATH="$checkout" run_channel stable
assert_log_line $'unlink\t--no-reboot' "switching from dev to stable unlinks without an early reboot prompt"
assert_log_line $'state\tset\treboot-required' "switching from dev to stable marks reboot required"

run_channel dev
if grep -q $'^git\tclone\t' "$log_file"; then
  fail "dev reuses an existing checkout" "$(cat "$log_file")"
fi
assert_log_line $'link\t'"$checkout"$'\t--no-reboot' "switching back to dev links ~/lunaos"
pass "switching back to dev reuses the existing ~/lunaos checkout"

printf 'cn\n' >"$test_tmp/region"
if LUNAOS_TEST_REGION_FILE="$test_tmp/region" run_channel dev >"$test_tmp/region.out" 2>"$test_tmp/region.err"; then
  fail "dev refuses a checkout without this machine's region"
fi
grep -q "does not support the 'cn' region" "$test_tmp/region.err" || fail "dev explains the missing region" "$(cat "$test_tmp/region.err")"
if grep -Eq $'^(link|refresh)\t' "$log_file"; then
  fail "dev refuses before linking or refreshing pacman" "$(cat "$log_file")"
fi
pass "dev refuses a checkout without the region before touching pacman"

# A directory with the region's name, or a helper that cannot run, is not
# support: the refresh after linking would fail with dev already linked.
mkdir -p "$checkout/default/regions/cn"
touch "$checkout/bin/lunaos-apply-pacman"
chmod +x "$checkout/bin/lunaos-apply-pacman"
if LUNAOS_TEST_REGION_FILE="$test_tmp/region" run_channel dev >"$test_tmp/region.out" 2>"$test_tmp/region.err"; then
  fail "dev refuses an empty region profile"
fi
if grep -Eq $'^(link|refresh)\t' "$log_file"; then
  fail "dev refuses an empty region profile before linking or refreshing pacman" "$(cat "$log_file")"
fi
mkdir -p "$checkout/default/regions/cn/pacman"
cn_fragments=(lunaos-stable.prepend lunaos-rc.prepend lunaos-edge.prepend mirrorlist-stable.prepend mirrorlist-rc.prepend mirrorlist-edge.prepend)
for missing in "${cn_fragments[@]}"; do
  for fragment in "${cn_fragments[@]}"; do
    if [[ $fragment == "$missing" ]]; then
      rm -f "$checkout/default/regions/cn/pacman/$fragment"
    else
      touch "$checkout/default/regions/cn/pacman/$fragment"
    fi
  done
  if LUNAOS_TEST_REGION_FILE="$test_tmp/region" run_channel dev >"$test_tmp/region.out" 2>"$test_tmp/region.err"; then
    fail "dev refuses a region profile without $missing"
  fi
  if grep -Eq $'^(link|refresh)\t' "$log_file"; then
    fail "dev refuses a profile without $missing before linking or refreshing pacman" "$(cat "$log_file")"
  fi
done
touch "${cn_fragments[@]/#/$checkout/default/regions/cn/pacman/}"
chmod -x "$checkout/bin/lunaos-apply-pacman"
if LUNAOS_TEST_REGION_FILE="$test_tmp/region" run_channel dev >"$test_tmp/region.out" 2>"$test_tmp/region.err"; then
  fail "dev refuses a checkout whose pacman helper cannot run"
fi
if grep -Eq $'^(link|refresh)\t' "$log_file"; then
  fail "dev refuses incomplete region support before linking or refreshing pacman" "$(cat "$log_file")"
fi
pass "dev refuses a region profile missing any fragment or a helper that cannot run"

chmod +x "$checkout/bin/lunaos-apply-pacman"
LUNAOS_TEST_REGION_FILE="$test_tmp/region" run_channel dev
assert_log_line $'refresh\tedge' "dev proceeds with a checkout that supports the region"

current_channel() {
  LUNAOS_TEST_VERSION_CHANNEL="$1" \
    LUNAOS_TEST_PACKAGES="$2" \
    LUNAOS_PATH="$3" \
    PATH="$stub_bin:$ROOT/bin:$PATH" \
    "$ROOT/bin/lunaos-channel-current"
}

[[ $(current_channel stable stable /usr/share/lunaos) == "stable" ]] || fail "current channel detects stable"
pass "current channel detects stable"

[[ $(current_channel rc stable /usr/share/lunaos) == "rc" ]] || fail "current channel detects rc"
pass "current channel detects rc"

[[ $(current_channel edge dev /usr/share/lunaos) == "edge" ]] || fail "current channel detects package-backed edge"
pass "current channel detects package-backed edge"

[[ $(current_channel edge dev "$test_tmp/dev-checkout") == "dev" ]] || fail "current channel detects dev from LUNAOS_PATH"
pass "current channel honors a dev link outside ~/lunaos"
