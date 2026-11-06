#!/bin/bash

# lunaos-luna is the desktop's only route to the LUNA Core. Without a Core
# it must say so with its own exit code instead of hanging or guessing, and it
# must refuse a malformed goal id before building a request path from it. When
# real luna-core and luna-policy-broker binaries are supplied (build them with
# `make -C lunaos build-local` and point LUNAOS_TEST_BIN at build/local), the
# second half drives the actual services through the owner socket: no stub
# answers for the Core.

set -euo pipefail

source "$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)/base-test.sh"

WORKDIR=$(mktemp -d)
PIDS=()
cleanup() {
  local pid
  for pid in ${PIDS[@]+"${PIDS[@]}"}; do kill "$pid" 2>/dev/null || true; done
  for pid in ${PIDS[@]+"${PIDS[@]}"}; do wait "$pid" 2>/dev/null || true; done
  rm -rf "$WORKDIR"
}
trap cleanup EXIT

cli="$ROOT/bin/lunaos-luna"

status=0
LUNAOS_CORE_SOCKET="$WORKDIR/absent.sock" "$cli" status >/dev/null 2>"$WORKDIR/err" || status=$?
(( status == 3 )) || fail "lunaos reports an absent Core as unavailable" "status=$status $(<"$WORKDIR/err")"
pass "lunaos reports an absent Core as unavailable"

status=0
LUNAOS_CORE_SOCKET="$WORKDIR/absent.sock" "$cli" pause '../intimacy/stop' >/dev/null 2>&1 || status=$?
(( status == 2 )) || fail "lunaos rejects a goal id that is not a UUID" "status=$status"
pass "lunaos rejects a goal id that is not a UUID"

# The install leaf touches nothing until the control-plane package exists, and a
# deferred-provisioning install records the owner group once however often it runs.
stubs="$WORKDIR/stubs"
mkdir -p "$stubs"
printf '#!/bin/bash\n[[ -e %q ]]\n' "$WORKDIR/package-installed" >"$stubs/lunaos-pkg-present"
printf '#!/bin/bash\necho "$0 $*" >>%q\n' "$WORKDIR/calls" | tee "$stubs/systemctl" >"$stubs/usermod"
chmod +x "$stubs"/*
leaf() { PATH="$stubs:$PATH" PROVISIONING_DIR="$WORKDIR/provisioning" LUNAOS_INSTALL_USER="$1" bash -eE "$ROOT/install/config/lunaos-services.sh"; }

leaf ""
[[ ! -e $WORKDIR/calls && ! -e $WORKDIR/provisioning/groups ]] || fail "lunaos install leaf does nothing without the package"
pass "lunaos install leaf does nothing without the package"

touch "$WORKDIR/package-installed"
leaf ""
leaf ""
[[ $(grep -cx luna-owner "$WORKDIR/provisioning/groups") == 1 ]] || fail "deferred provisioning records luna-owner once" "$(cat "$WORKDIR/provisioning/groups")"
leaf owner
grep -q 'usermod -aG luna-owner owner' "$WORKDIR/calls" || fail "the install user joins luna-owner" "$(cat "$WORKDIR/calls")"
grep -q "enable.*luna-inference.service" "$WORKDIR/calls" || fail "the install leaf enables the inference service" "$(cat "$WORKDIR/calls")"
pass "the install leaf grants luna-owner to the owner only, once"

bin=${LUNAOS_TEST_BIN:-}
if [[ -z $bin || ! -x $bin/luna-core || ! -x $bin/luna-policy-broker ]] || [[ $(uname -s) != "Linux" ]]; then
  skip "lunaos drives the real Core and Policy Broker (set LUNAOS_TEST_BIN on Linux)"
  exit 0
fi

run="$WORKDIR/run"
mkdir -m 700 "$run"
head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n' >"$run/owner-token"
head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n' >"$run/policy-token"
chmod 600 "$run/owner-token" "$run/policy-token"

"$bin/luna-policy-broker" --socket="$run/policy.sock" --key-file="$run/policy-token" --client-user="$(id -un)" >"$WORKDIR/broker.log" 2>&1 &
PIDS+=($!)
for _ in {1..50}; do [[ -S $run/policy.sock ]] && break; sleep 0.1; done

LUNA_OWNER_TOKEN_FILE="$run/owner-token" LUNA_DATABASE="$run/goals.sqlite" LUNA_OWNER_SOCKET="$run/core.sock" \
  LUNA_POLICY_SOCKET="$run/policy.sock" LUNA_POLICY_TOKEN_FILE="$run/policy-token" LUNA_LISTEN=127.0.0.1:0 \
  "$bin/luna-core" >"$WORKDIR/core.log" 2>&1 &
PIDS+=($!)
for _ in {1..50}; do [[ -S $run/core.sock ]] && break; sleep 0.1; done
[[ -S $run/core.sock ]] || fail "real Core opened its owner socket" "$(cat "$WORKDIR/core.log")"

export LUNAOS_CORE_SOCKET="$run/core.sock"

broker=$("$cli" status | jq -r '.policy_broker.status' 2>/dev/null || true)
[[ $broker == "ready" ]] || fail "status reports the Policy Broker ready through the owner socket" "$("$cli" status 2>&1)"
pass "status reports the Policy Broker ready through the owner socket"

goal=$("$cli" checksum)
id=$(jq -r '.id' <<<"$goal")
[[ $(jq -r '.status' <<<"$goal") == "approved" ]] || fail "checksum goal is approved by the real Policy Broker" "$goal"
pass "checksum goal is approved by the real Policy Broker"

for _ in {1..60}; do
  [[ $("$cli" goal "$id" | jq -r '.status') == "completed" ]] && break
  sleep 0.2
done
evidence=$("$cli" goal "$id")
[[ $(jq -r '.status' <<<"$evidence") == "completed" && $(jq -r '.evidence.verified' <<<"$evidence") == "true" ]] ||
  fail "the scheduler completes the checksum with verified evidence" "$evidence"
pass "the scheduler completes the checksum with verified evidence"

[[ $("$cli" privacy | jq -r '.mode') == "STARTUP_LOCKED" ]] || fail "LUNA starts with its eyes closed" "$("$cli" privacy)"
[[ $("$cli" open-eyes | jq -r '.mode') == "NORMAL_EPHEMERAL_DETECTION" ]] || fail "the owner socket opens LUNA's eyes"
[[ $("$cli" blind | jq -r '.capture_allowed') == "false" ]] || fail "blind closes LUNA's eyes"
pass "privacy starts locked, opens from the owner socket and closes on blind"

status=0
"$cli" pause "$id" >/dev/null 2>&1 || status=$?
(( status == 5 )) || fail "the Core's refusal to pause a completed goal surfaces as a rejection" "status=$status"
pass "the Core's refusal to pause a completed goal surfaces as a rejection"
