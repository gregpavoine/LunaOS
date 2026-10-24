#!/bin/bash
set -euo pipefail

lunaos_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
binary="$lunaos_dir/rust/target/release/luna-core"
[[ -x $binary ]] || { printf '%s\n' 'Build luna-core in release mode first.' >&2; exit 2; }
for command_name in curl jq; do
  command -v "$command_name" >/dev/null 2>&1 || { printf 'Missing smoke-test tool: %s\n' "$command_name" >&2; exit 2; }
done

temporary_dir=$(mktemp -d "${TMPDIR:-/tmp}/lunaos-native-smoke.XXXXXX")
owner_token='local-smoke-only-owner-token'
base_url='http://127.0.0.1:17766'
server_pid=''
cleanup() {
  if [[ -n $server_pid ]]; then
    kill "$server_pid" 2>/dev/null || true
    wait "$server_pid" 2>/dev/null || true
  fi
  rm -rf -- "$temporary_dir"
}
trap cleanup EXIT

start_core() {
  LUNA_OWNER_TOKEN="$owner_token" \
    LUNA_DATABASE="$temporary_dir/goals.sqlite" \
    LUNA_LISTEN='127.0.0.1:17766' \
    "$binary" >"$temporary_dir/core.log" 2>&1 &
  server_pid=$!
  for _ in {1..50}; do
    if curl --silent --fail "$base_url/v1/system/capabilities" >/dev/null; then return 0; fi
    sleep 0.1
  done
  cat "$temporary_dir/core.log" >&2
  return 1
}

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
kill "$server_pid"
wait "$server_pid" 2>/dev/null || true
server_pid=''
start_core
persisted=$(curl --silent --fail "$base_url/v1/goals/$goal_id" \
  --header "Authorization: Bearer $owner_token")
[[ $(jq -er '.status' <<<"$persisted") == 'completed' ]]
printf '%s\n' 'PASS: unauthorized request denied; owner goal persisted; duplicate idempotent; action verified; completed state survived Core restart.'
