#!/usr/bin/env bash
set -euo pipefail

repo_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
broker_bin="$repo_dir/rust/target/release/luna-policy-broker"
core_bin="$repo_dir/rust/target/release/luna-core"
[[ -x $broker_bin && -x $core_bin ]] || { printf '%s\n' 'Run make build before the native smoke test.' >&2; exit 2; }
for command_name in curl jq; do
  command -v "$command_name" >/dev/null 2>&1 || { printf 'Missing smoke-test tool: %s\n' "$command_name" >&2; exit 2; }
done

temporary_dir=$(mktemp -d /tmp/losm.XXXXXX)
chmod 700 "$temporary_dir"
owner_token=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
printf '%s\n' "$owner_token" > "$temporary_dir/owner-token"
chmod 600 "$temporary_dir/owner-token"
base_url=${LUNA_SMOKE_URL:-http://127.0.0.1:17766}
broker_pid=''
core_pid=''
cleanup() {
  for pid in "$core_pid" "$broker_pid"; do
    if [[ -n $pid ]]; then kill "$pid" 2>/dev/null || true; wait "$pid" 2>/dev/null || true; fi
  done
  rm -rf -- "$temporary_dir"
}
trap cleanup EXIT

start_broker() {
  LUNA_OWNER_TOKEN_FILE="$temporary_dir/owner-token" \
    LUNA_POLICY_SOCKET="$temporary_dir/policy.sock" \
    "$broker_bin" >"$temporary_dir/broker.log" 2>&1 &
  broker_pid=$!
  for _ in {1..50}; do
    [[ -S $temporary_dir/policy.sock ]] && return 0
    kill -0 "$broker_pid" 2>/dev/null || break
    sleep 0.1
  done
  cat "$temporary_dir/broker.log" >&2
  return 1
}

start_core() {
  LUNA_OWNER_TOKEN_FILE="$temporary_dir/owner-token" \
    LUNA_POLICY_SOCKET="$temporary_dir/policy.sock" \
    LUNA_DATABASE="$temporary_dir/goals.sqlite" \
    LUNA_LISTEN='127.0.0.1:17766' \
    "$core_bin" >"$temporary_dir/core.log" 2>&1 &
  core_pid=$!
  for _ in {1..50}; do
    if curl --silent --fail "$base_url/v1/system/capabilities" >/dev/null; then return 0; fi
    kill -0 "$core_pid" 2>/dev/null || break
    sleep 0.1
  done
  cat "$temporary_dir/core.log" >&2
  return 1
}

start_broker
start_core
unauthorized_status=$(curl --silent --output /dev/null --write-out '%{http_code}' \
  --request POST "$base_url/v1/goals" \
  --header 'Content-Type: application/json' \
  --header 'Idempotency-Key: smoke-unauthorized' \
  --data '{"action":"local.checksum"}')
[[ $unauthorized_status == '401' ]]

goal=$(curl --silent --fail --request POST "$base_url/v1/goals" \
  --header "Authorization: Bearer $owner_token" \
  --header 'Content-Type: application/json' \
  --header 'Idempotency-Key: smoke-native-goal' \
  --data '{"action":"local.checksum"}')
goal_id=$(jq -er '.id' <<<"$goal")
duplicate=$(curl --silent --fail --request POST "$base_url/v1/goals" \
  --header "Authorization: Bearer $owner_token" \
  --header 'Content-Type: application/json' \
  --header 'Idempotency-Key: smoke-native-goal' \
  --data '{"action":"local.checksum"}')
[[ $(jq -er '.id' <<<"$duplicate") == "$goal_id" ]]

result=$(curl --silent --fail --request POST "$base_url/v1/scheduler/tick" \
  --header "Authorization: Bearer $owner_token")
[[ $(jq -er '.outcome' <<<"$result") == 'succeeded' ]]
kill "$core_pid"
wait "$core_pid" 2>/dev/null || true
core_pid=''
start_core
persisted=$(curl --silent --fail "$base_url/v1/goals/$goal_id" \
  --header "Authorization: Bearer $owner_token")
[[ $(jq -er '.status' <<<"$persisted") == 'completed' ]]
printf '%s\n' 'PASS: Rust Core authenticated owner requests through a separate Rust policy-broker process; idempotent checksum action produced evidence; persisted state survived Core restart.'
