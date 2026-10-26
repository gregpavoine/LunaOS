#!/bin/bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
script="$repo_root/lunaos/scripts/test-iso-install.sh"
temp_root=$(mktemp -d "${TMPDIR:-/tmp}/lunaos-install-script-test.XXXXXX")
trap 'rm -rf -- "$temp_root"' EXIT

if [[ ! -f $script ]]; then
  printf '%s\n' 'FAIL: install VM acceptance script is missing.' >&2
  exit 1
fi

if [[ $(uname -s) != Linux ]]; then
  output=$(TMPDIR="$temp_root" "$script" /dev/null 2>&1) && {
    printf '%s\n' 'FAIL: install VM acceptance script unexpectedly accepted a non-Linux host.' >&2
    exit 1
  }
  status=$?
  if (( status != 2 )) || [[ $output != *'requires an x86_64 Linux QEMU host'* ]]; then
    printf 'FAIL: expected a clear safe preflight failure (exit 2); got status %s and output: %s\n' "$status" "$output" >&2
    exit 1
  fi
  if [[ -n $(find "$temp_root" -mindepth 1 -print -quit) ]]; then
    printf '%s\n' 'FAIL: unsupported host created VM artifacts before preflight rejection.' >&2
    exit 1
  fi
fi

printf '%s\n' 'PASS: install VM acceptance rejects unsupported hosts before creating artifacts.'
