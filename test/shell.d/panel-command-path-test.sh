#!/bin/bash

set -euo pipefail

source "$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)/base-test.sh"

if matches=$(rg -n 'root\.bar\.lunaosPath|/bin/lunaos-' "$ROOT/shell/plugins/panels" -g '*.qml'); then
  fail "panels do not resolve lunaos helpers through bar paths" "$matches"
fi

pass "panels avoid bar path resolution for lunaos helpers"
