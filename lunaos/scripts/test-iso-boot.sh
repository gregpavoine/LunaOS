#!/bin/bash
set -euo pipefail

if [[ $(uname -s) != "Linux" || $(uname -m) != "x86_64" ]]; then
  printf '%s\n' 'ISO boot smoke requires an x86_64 Linux QEMU host.' >&2
  exit 2
fi

if (($# != 1)) || [[ ! -f $1 ]]; then
  printf 'Usage: %s /path/to/lunaos.iso\n' "$0" >&2
  exit 2
fi

if ! command -v qemu-system-x86_64 >/dev/null 2>&1; then
  printf '%s\n' 'Missing required command: qemu-system-x86_64' >&2
  exit 2
fi

firmware_code=""
firmware_vars=""
for pair in \
  "/usr/share/OVMF/OVMF_CODE_4M.fd /usr/share/OVMF/OVMF_VARS_4M.fd" \
  "/usr/share/OVMF/OVMF_CODE.fd /usr/share/OVMF/OVMF_VARS.fd" \
  "/usr/share/edk2/x64/OVMF_CODE.4m.fd /usr/share/edk2/x64/OVMF_VARS.4m.fd"; do
  read -r code_candidate vars_candidate <<<"$pair"
  if [[ -f $code_candidate && -f $vars_candidate ]]; then
    firmware_code=$code_candidate
    firmware_vars=$vars_candidate
    break
  fi
done
if [[ -z $firmware_code ]]; then
  printf '%s\n' 'Missing OVMF UEFI firmware; install the ovmf package.' >&2
  exit 2
fi

log_file=${LUNAOS_BOOT_LOG:-"${TMPDIR:-/tmp}/lunaos-iso-boot.log"}
timeout_seconds=${LUNAOS_BOOT_TIMEOUT_SECONDS:-600}
firmware_work=$(mktemp -d "${TMPDIR:-/tmp}/lunaos-ovmf.XXXXXX")
cp "$firmware_vars" "$firmware_work/OVMF_VARS.fd"
timeout --signal=TERM "$timeout_seconds" qemu-system-x86_64 \
  -machine q35,accel=tcg \
  -cpu max \
  -smp 2 \
  -m 4096 \
  -drive "if=pflash,format=raw,unit=0,readonly=on,file=$firmware_code" \
  -drive "if=pflash,format=raw,unit=1,file=$firmware_work/OVMF_VARS.fd" \
  -cdrom "$1" \
  -boot order=d \
  -display none \
  -serial stdio \
  -monitor none \
  -net none \
  -no-reboot >"$log_file" 2>&1 &
qemu_pid=$!

cleanup() {
  if kill -0 "$qemu_pid" 2>/dev/null; then
    kill -TERM "$qemu_pid" 2>/dev/null || true
    wait "$qemu_pid" 2>/dev/null || true
  fi
  rm -rf -- "$firmware_work"
}
trap cleanup EXIT

deadline=$((SECONDS + timeout_seconds))
while ((SECONDS < deadline)); do
  # systemd writes ANSI color escapes between words on a serial console, so
  # match the stable target name instead of the formatted log line.
  if grep -q 'Graphical Interface' "$log_file"; then
    for marker in \
      'Owner and policy credentials provisioned' \
      'Started LunaOS owner-authenticated policy broker' \
      'Started LunaOS Core API and durable goal scheduler'; do
      if ! grep -q "$marker" "$log_file"; then
        printf 'FAIL: LunaOS boot did not confirm: %s\n' "$marker" >&2
        cat "$log_file" >&2
        exit 1
      fi
    done
    printf 'PASS: LunaOS reached the graphical target and started first-boot credentials, Policy Broker, and Core in UEFI QEMU.\n'
    cat "$log_file"
    exit 0
  fi
  process_state=$(ps -o stat= -p "$qemu_pid" 2>/dev/null || true)
  if [[ -z $process_state || $process_state == Z* ]]; then
    printf '%s\n' 'QEMU exited before LunaOS reached the systemd graphical target.' >&2
    cat "$log_file" >&2
    exit 1
  fi
  sleep 5
done

printf 'FAIL: LunaOS did not reach the systemd graphical target within %s seconds.\n' "$timeout_seconds" >&2
cat "$log_file" >&2
exit 1
