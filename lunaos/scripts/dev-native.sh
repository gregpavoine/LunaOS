#!/usr/bin/env bash
set -euo pipefail

repo_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
manifest="$repo_dir/rust/Cargo.toml"
temporary_dir=$(mktemp -d /tmp/lunadev.XXXXXX)
chmod 700 "$temporary_dir"
umask 077
owner_token=$(openssl rand -hex 32)
printf '%s\n' "$owner_token" > "$temporary_dir/owner-token"
broker_pid=''
core_pid=''
cleanup() {
  for pid in "$core_pid" "$broker_pid"; do
    if [[ -n $pid ]]; then kill "$pid" 2>/dev/null || true; wait "$pid" 2>/dev/null || true; fi
  done
  rm -rf -- "$temporary_dir"
}
trap cleanup EXIT INT TERM

cargo build --manifest-path "$manifest" -p luna-policy-broker -p luna-core
LUNA_OWNER_TOKEN_FILE="$temporary_dir/owner-token" \
  LUNA_POLICY_SOCKET="$temporary_dir/policy.sock" \
  "$repo_dir/rust/target/debug/luna-policy-broker" &
broker_pid=$!
for _ in {1..50}; do
  [[ -S $temporary_dir/policy.sock ]] && break
  kill -0 "$broker_pid" 2>/dev/null || { printf '%s\n' 'Policy broker exited during startup.' >&2; exit 1; }
  sleep 0.1
done
[[ -S $temporary_dir/policy.sock ]] || { printf '%s\n' 'Policy broker socket did not appear.' >&2; exit 1; }

printf 'Owner API token: %s\n' "$owner_token"
LUNA_OWNER_TOKEN_FILE="$temporary_dir/owner-token" \
  LUNA_POLICY_SOCKET="$temporary_dir/policy.sock" \
  LUNA_DATABASE="$temporary_dir/goals.sqlite" \
  LUNA_LISTEN='127.0.0.1:7766' \
  "$repo_dir/rust/target/debug/luna-core" &
core_pid=$!
wait "$core_pid"
