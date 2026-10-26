#!/bin/bash
set -euo pipefail

repo_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
installer="$repo_dir/packaging/arch/airootfs/usr/local/bin/lunaos-install"
test_dir=$(mktemp -d "${TMPDIR:-/tmp}/lunaos-installer-boundary.XXXXXX")
trap 'rm -rf -- "$test_dir"' EXIT
marker="$test_dir/startup-ran"
decoy="$test_dir/bash-env"
cat > "$decoy" <<'EOF'
: "${LUNAOS_TEST_MARKER:?}"
: > "$LUNAOS_TEST_MARKER"
EOF
chmod 0600 "$decoy"

set +e
env BASH_ENV="$decoy" LUNAOS_TEST_MARKER="$marker" "$installer" -p > "$test_dir/output" 2>&1
installer_status=$?
set -e

if [[ -e $marker ]]; then
  printf '%s\n' 'FAIL: installer evaluated BASH_ENV before rejecting a decoy -p argument.' >&2
  cat "$test_dir/output" >&2
  exit 1
fi
if (( installer_status == 0 )); then
  printf '%s\n' 'FAIL: installer accepted a command-line argument.' >&2
  exit 1
fi
printf '%s\n' 'PASS: protected interpreter ignored BASH_ENV and the installer rejected a decoy -p argument.'
