#!/bin/bash
set -euo pipefail

if [[ $(uname -s) != Linux ]]; then
  printf '%s\n' 'The Go Broker uses Linux Unix-socket peer credentials. Run this launcher on Linux; on macOS use make test-process-linux in Docker.' >&2
  exit 2
fi
for command_name in go curl; do
  if ! command -v "$command_name" >/dev/null 2>&1; then
    printf 'Missing development tool: %s\n' "$command_name" >&2
    exit 2
  fi
done

repo_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
cd "$repo_dir"
temporary_dir=$(mktemp -d /tmp/lunadev.XXXXXX)
chmod 700 "$temporary_dir"
umask 077
broker_pid=''
core_pid=''
cleanup() {
  for pid in "$core_pid" "$broker_pid"; do
    if [[ -n $pid ]]; then
      kill "$pid" 2>/dev/null || true
      wait "$pid" 2>/dev/null || true
    fi
  done
  rm -rf -- "$temporary_dir"
}
trap cleanup EXIT INT TERM

go build -buildvcs=false -trimpath -o "$temporary_dir/luna-owner-setup" "$repo_dir/services/identity/cmd/luna-owner-setup"
go build -buildvcs=false -trimpath -o "$temporary_dir/luna-policy-broker" "$repo_dir/services/policy"
go build -buildvcs=false -trimpath -o "$temporary_dir/luna-core" "$repo_dir/services/core/cmd/luna-core"
"$temporary_dir/luna-owner-setup" --path "$temporary_dir/owner-token" --policy-path "$temporary_dir/policy-token" --reveal-token

"$temporary_dir/luna-policy-broker" \
  --socket="$temporary_dir/policy.sock" \
  --key-file="$temporary_dir/policy-token" --client-user="$(id -un)" &
broker_pid=$!
for _ in {1..50}; do
  if [[ -S $temporary_dir/policy.sock ]]; then
    break
  fi
  if ! kill -0 "$broker_pid" 2>/dev/null; then
    printf '%s\n' 'Policy Broker exited during startup.' >&2
    exit 1
  fi
  sleep 0.1
done
if [[ ! -S $temporary_dir/policy.sock ]]; then
  printf '%s\n' 'Policy Broker socket did not appear.' >&2
  exit 1
fi

listen_address=${LUNA_LISTEN:-127.0.0.1:7766}
LUNA_OWNER_TOKEN_FILE="$temporary_dir/owner-token" \
  LUNA_POLICY_SOCKET="$temporary_dir/policy.sock" \
  LUNA_POLICY_TOKEN_FILE="$temporary_dir/policy-token" \
  LUNA_DATABASE="$temporary_dir/goals.sqlite" \
  LUNA_OWNER_SOCKET="$temporary_dir/core.sock" \
  LUNA_LISTEN="$listen_address" \
  "$temporary_dir/luna-core" &
core_pid=$!

for _ in {1..50}; do
  status=$(curl --silent --output /dev/null --write-out '%{http_code}' --max-time 0.2 "http://$listen_address/v1/goals" || true)
  if [[ $status == 401 ]]; then
    printf 'LunaOS Go Core is ready at http://%s\n' "$listen_address"
    wait "$core_pid"
    exit $?
  fi
  if ! kill -0 "$core_pid" 2>/dev/null; then
    printf '%s\n' 'LUNA Core exited during startup.' >&2
    exit 1
  fi
  sleep 0.1
done
printf '%s\n' 'LUNA Core API did not become ready.' >&2
exit 1
