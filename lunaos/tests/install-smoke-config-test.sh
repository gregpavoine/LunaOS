#!/bin/bash
set -euo pipefail

repo_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
generator="$repo_dir/scripts/generate-install-smoke-config.sh"
template="$repo_dir/packaging/arch/airootfs/usr/share/lunaos/archinstall.json"
temporary_dir=$(mktemp -d "${TMPDIR:-/tmp}/lunaos-install-smoke-config-test.XXXXXX")
trap 'rm -rf -- "$temporary_dir"' EXIT

fail() {
  printf 'FAIL: %s\n' "$1" >&2
  exit 1
}

output_dir="$temporary_dir/generated"
if ! bash "$generator" "$template" "$output_dir"; then
  fail 'the install-smoke configuration generator did not complete'
fi

config="$output_dir/user_configuration.json"
credentials="$output_dir/user_credentials.json"
password_file="$output_dir/root-password"
[[ -s $config && -s $credentials && -s $password_file ]] || fail 'generated configuration files are missing or empty'
jq -e '.disk_config.device_modifications | length == 1 and .[0].device == "/dev/vda" and .[0].wipe == true' "$config" >/dev/null || fail 'test config does not target only the disposable QEMU disk'
jq -e '.disk_config.device_modifications[0].partitions | length == 2 and any(.[]; .mountpoint == "/boot" and .fs_type == "fat32") and any(.[]; .mountpoint == "/" and .fs_type == "ext4")' "$config" >/dev/null || fail 'test config does not define the UEFI and root partitions'
jq -e '.silent == true and .offline == true and .no_pkg_lookups == true' "$config" >/dev/null || fail 'test config must install silently from the bundled repositories'
jq -e '(.users | length == 1) and (.users[0].username == "lunaos-ci") and (.users[0].sudo == true) and (.users[0].enc_password | startswith("$6$")) and (.root_enc_password | startswith("$6$"))' "$credentials" >/dev/null || fail 'test credentials do not contain the disposable owner and root accounts'
if jq -e 'has("users") or has("auth_config")' "$config" >/dev/null; then
  fail 'test account credentials leaked into the public configuration file'
fi
if jq -e 'has("disk_config") or has("users") or has("auth_config")' "$template" >/dev/null; then
  fail 'the shipped Archinstall template contains a target disk or account credentials'
fi
if rg -F -q -- "$(cat "$password_file")" "$credentials"; then
  fail 'the generated password was not hashed before being stored'
fi

for path in "$output_dir" "$config" "$credentials" "$password_file"; do
  mode=$(stat -f '%Lp' "$path" 2>/dev/null || stat -c '%a' "$path")
  expected=700
  if [[ $path != "$output_dir" ]]; then
    expected=600
  fi
  [[ $mode == "$expected" ]] || fail "unexpected permissions $mode on $path"
done

printf '%s\n' 'PASS: generated Archinstall config is isolated to the disposable UEFI test disk and credentials stay private.'
