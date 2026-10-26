#!/bin/bash
set -euo pipefail

if [[ $(uname -s) != Linux || $(uname -m) != x86_64 ]]; then
  printf '%s\n' 'Install VM acceptance requires an x86_64 Linux QEMU host.' >&2
  exit 2
fi
if (($# != 1)) || [[ ! -f $1 ]]; then
  printf 'Usage: %s /path/to/lunaos.iso\n' "$0" >&2
  exit 2
fi

script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
repo_root=$(cd "$script_dir/../.." && pwd)
iso_path=$(cd "$(dirname "$1")" && pwd)/$(basename "$1")
for command_name in qemu-system-x86_64 qemu-img expect xorriso jq openssl makepkg runuser repo-add timeout; do
  if ! command -v "$command_name" >/dev/null 2>&1; then
    printf 'Missing install VM acceptance tool: %s\n' "$command_name" >&2
    exit 2
  fi
done

firmware_code=''
firmware_vars=''
for pair in \
  '/usr/share/OVMF/OVMF_CODE_4M.fd /usr/share/OVMF/OVMF_VARS_4M.fd' \
  '/usr/share/OVMF/OVMF_CODE.fd /usr/share/OVMF/OVMF_VARS.fd' \
  '/usr/share/edk2/x64/OVMF_CODE.4m.fd /usr/share/edk2/x64/OVMF_VARS.4m.fd'; do
  read -r code_candidate vars_candidate <<<"$pair"
  if [[ -f $code_candidate && -f $vars_candidate ]]; then
    firmware_code=$code_candidate
    firmware_vars=$vars_candidate
    break
  fi
done
if [[ -z $firmware_code ]]; then
  printf '%s\n' 'Missing OVMF UEFI firmware; install the edk2-ovmf package.' >&2
  exit 2
fi

work_dir=$(mktemp -d "${TMPDIR:-/tmp}/lunaos-install-vm.XXXXXX")
chmod 0700 "$work_dir"
keep_logs=${LUNAOS_KEEP_INSTALL_LOGS:-0}
cleanup() {
  if [[ $keep_logs == 1 ]]; then
    log_destination=${LUNAOS_INSTALL_LOG_DEST:-}
    if [[ -n $log_destination && -f $work_dir/install.log ]]; then
      install -D -m 0600 "$work_dir/install.log" "$log_destination"
      printf 'Sanitized serial diagnostics retained at %s\n' "$log_destination" >&2
    else
      printf '%s\n' 'Install VM diagnostics destination missing; private VM artifacts will be removed.' >&2
    fi
  fi
  rm -rf -- "$work_dir"
}
trap cleanup EXIT

"$script_dir/generate-install-smoke-config.sh" \
  "$repo_root/lunaos/packaging/arch/airootfs/usr/share/lunaos/archinstall.json" \
  "$work_dir/inputs"
jq '.packages += ["lunaos-install-smoke"] | .mirror_config.custom_repositories += [{name:"lunaos-ci",url:"file:///mnt/lunaos-ci/repo/$arch",sign_check:"Never",sign_option:"TrustAll"}]' \
  "$work_dir/inputs/user_configuration.json" > "$work_dir/user_configuration.json"
install -m 0600 "$work_dir/inputs/user_credentials.json" "$work_dir/user_credentials.json"

pkg_dir="$work_dir/pkgbuild"
mkdir -m 0755 "$pkg_dir"
cp "$repo_root/lunaos/packaging/arch/install-test/PKGBUILD" \
  "$repo_root/lunaos/packaging/arch/install-test/lunaos-install-smoke.go" \
  "$repo_root/lunaos/packaging/arch/install-test/lunaos-install-smoke.service" "$pkg_dir/"
chown -R nobody:nobody "$pkg_dir"
(cd "$pkg_dir" && runuser -u nobody -- makepkg --nodeps --noconfirm --clean)
smoke_package=$(find "$pkg_dir" -maxdepth 1 -type f -name 'lunaos-install-smoke-*.pkg.tar.*' -print -quit)
if [[ -z $smoke_package ]]; then
  printf '%s\n' 'CI-only install smoke package was not built.' >&2
  exit 1
fi

mkdir "$work_dir/seed"
install -m 0600 "$work_dir/user_configuration.json" "$work_dir/seed/user_configuration.json"
install -m 0600 "$work_dir/user_credentials.json" "$work_dir/seed/user_credentials.json"
mkdir -p "$work_dir/seed/repo/x86_64"
install -m 0644 "$smoke_package" "$work_dir/seed/repo/x86_64/$(basename "$smoke_package")"
repo-add "$work_dir/seed/repo/x86_64/lunaos-ci.db.tar.gz" "$work_dir/seed/repo/x86_64/$(basename "$smoke_package")"
xorriso -as mkisofs -quiet -V LUNAOS_CI -o "$work_dir/lunaos-ci-seed.iso" "$work_dir/seed"
qemu-img create -q -f qcow2 "$work_dir/target.qcow2" 32G
cp "$firmware_vars" "$work_dir/OVMF_VARS.fd"

printf '%s\n' 'Starting isolated UEFI QEMU install; only the newly created target.qcow2 is writable.'
LUNAOS_ISO_PATH="$iso_path" \
LUNAOS_SEED_ISO="$work_dir/lunaos-ci-seed.iso" \
LUNAOS_TARGET_DISK="$work_dir/target.qcow2" \
LUNAOS_OVMF_CODE="$firmware_code" \
LUNAOS_OVMF_VARS="$work_dir/OVMF_VARS.fd" \
LUNAOS_INSTALL_LOG="$work_dir/install.log" \
timeout --signal=TERM "${LUNAOS_INSTALL_TIMEOUT_SECONDS:-7200}" \
  expect "$repo_root/lunaos/scripts/test-iso-install.expect"

if ! grep -q 'LUNAOS_INSTALL_SMOKE=PASS installed_boot=1 core=active broker=active goal=completed evidence=verified' "$work_dir/install.log"; then
  printf '%s\n' 'FAIL: QEMU exited without the installed-system, service and evidence acceptance marker.' >&2
  cat "$work_dir/install.log" >&2
  exit 1
fi
printf '%s\n' 'PASS: production installer installed and booted LunaOS; real Core/Broker services completed the owner-authorized checksum goal with verified evidence.'
