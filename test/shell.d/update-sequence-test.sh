#!/bin/bash

set -euo pipefail

source "$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)/base-test.sh"
source "$SHELL_TEST_DIR/fixtures/sudo-boundary-test.sh"
copy_boundary_file bin/lunaos-update
test_tmp="$boundary_tmp"
stub_bin="$SUDO_TEST_ROOT/bin"

# Every step lunaos-update runs, recorded in order with the unattended flag it
# was handed. One of them can be told to fail. lunaos-update-available is here
# so a recheck after the update would show up as an unexpected step.
steps=(
  lunaos-update-lock
  lunaos-update-requires-free-space
  lunaos-update-confirm
  lunaos-update-pkg-prune
  lunaos-snapshot
  lunaos-update-stay-awake
  lunaos-update-dev
  lunaos-update-boot
  lunaos-update-keyring
  lunaos-update-system-pkgs
  lunaos-migrate
  lunaos-hook
  lunaos-update-aur-pkgs
  lunaos-update-mise
  lunaos-update-orphan-pkgs
  lunaos-update-analyze-logs
  lunaos-shell
  lunaos-update-available
  lunaos-update-restart
)

for step in "${steps[@]}"; do
  rm -f "$stub_bin/$step"
  cat >"$stub_bin/$step" <<'STUB'
#!/bin/bash
printf '%s unattended=%s\n' "${0##*/}" "${LUNAOS_UPDATE_UNATTENDED:-}" >>"$STEP_LOG"
if [[ ${0##*/} == "lunaos-update-orphan-pkgs" ]]; then
  printf '%s\n' "$@" >"$STEP_LOG.orphan-args"
fi
if [[ ${0##*/} == "lunaos-shell" ]]; then
  printf '%s\n' "$*" >"$STEP_LOG.shell-args"
fi
[[ ${FAILING_STEP:-} != "${0##*/}" ]] || exit 1
STUB
  chmod +x "$stub_bin/$step"
done

# LUNAOS_UPDATE_LOGGED stands in for the script(1) wrapper the update re-execs
# itself under; the stubbed lock reports itself already held.
run_update() {
  : >"$test_tmp/steps"
  STEP_LOG="$test_tmp/steps" \
    FAILING_STEP="${FAILING_STEP:-}" \
    LUNAOS_UPDATE_LOGGED=1 \
    PATH="$stub_bin:$PATH" \
    "$SUDO_TEST_ROOT/bin/lunaos-update" "$@" >"$test_tmp/out" 2>"$test_tmp/err"
}

steps_run() {
  cut -d' ' -f1 "$test_tmp/steps"
}

# Every step of a whole update, in order. $1 asks for the one a person confirms.
# Stay Awake bookends the work, so it is here twice.
expected_steps() {
  printf '%s\n' \
    lunaos-update-lock \
    lunaos-update-requires-free-space \
    ${1:+lunaos-update-confirm} \
    lunaos-update-pkg-prune \
    lunaos-snapshot \
    lunaos-update-stay-awake \
    lunaos-update-dev \
    lunaos-update-keyring \
    lunaos-update-system-pkgs \
    lunaos-migrate \
    lunaos-update-orphan-pkgs \
    lunaos-update-analyze-logs \
    lunaos-shell \
    lunaos-update-restart \
    lunaos-hook \
    lunaos-update-mise \
    lunaos-update-aur-pkgs \
    lunaos-update-boot \
    lunaos-update-stay-awake \
    lunaos-update-restart
}

run_update -y || fail "an update where everything works reports a failure"
diff <(expected_steps) <(steps_run) >"$test_tmp/order" ||
  fail "an update where everything works does not run every step in order" "$(cat "$test_tmp/order")"
pass "an update where every step works runs all of them, in order"
[[ $(cat "$test_tmp/steps.orphan-args") == "-y" ]] || fail "an unattended update approves orphan removal"
pass "an unattended update approves orphan removal"
[[ $(cat "$test_tmp/steps.shell-args") == "-q lunaos.system-update clear" ]] ||
  fail "a finished update clears the shell update indicator" "$(cat "$test_tmp/steps.shell-args")"
pass "a finished update clears the shell update indicator without rechecking"

grep -q '^lunaos-update-system-pkgs unattended=1$' "$test_tmp/steps" ||
  fail "-y does not mark the update unattended"
run_update </dev/null || fail "a confirmed update reports a failure"
diff <(expected_steps confirmed) <(steps_run) >"$test_tmp/order" ||
  fail "a confirmed update runs a different set of steps" "$(cat "$test_tmp/order")"
grep -q '^lunaos-update-system-pkgs unattended=$' "$test_tmp/steps" ||
  fail "an update a person confirmed is treated as unattended"
pass "-y is what marks an update unattended, not the update itself"
[[ $(cat "$test_tmp/steps.orphan-args") == "-y" ]] || fail "a confirmed update approves orphan removal"
pass "a confirmed update approves orphan removal"

# Migrations ship with the packages the upgrade installs and are written against
# them. Running them against what is still on disk is the failure this ordering
# exists to prevent, so the update stops where the packages did.
if FAILING_STEP=lunaos-update-system-pkgs run_update -y; then
  fail "an update whose packages did not upgrade passes for a whole one"
fi
for step in lunaos-migrate lunaos-hook lunaos-update-aur-pkgs lunaos-update-restart; do
  if grep -q "^$step " "$test_tmp/steps"; then
    fail "a blocked package upgrade still runs $step"
  fi
done
pass "a blocked package upgrade stops the update before it migrates"
