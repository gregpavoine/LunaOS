# Image builds and the hardware setup they defer to the machine's first boot.
# Sourced by lunaos-apply-hardware and lunaos-provision-hardware; no shebang.
#
# An image is built away from the machine it will run on, so hardware setup
# during the build would describe the builder, not the target. The builder
# declares the build in a root-owned manifest before it runs
# lunaos-apply-system in the image root:
#
#   /var/lib/lunaos/image/target
#     format=1
#     platform=x86|aarch64|aarch64-apple|aarch64-qualcomm|aarch64-n1x|aarch64-gb10
#
# While it exists, lunaos-apply-hardware queues each hardware leaf in
# /var/lib/lunaos/image/deferred-steps instead of running it, and arms
# lunaos-provision-hardware.service. On the machine's first boot that service,
# or the platform's own first boot, retires the manifest (to target.booted) and
# runs the queue on the real hardware. Unknown manifest keys are ignored so the
# format can grow.
#
# An image carries no pacman master key of its own, or every machine flashed
# from it would share one: where install finalization would make the keyring
# (install/post-install/pacman.sh), a build requests it in
# /var/lib/lunaos/image/pacman-keyring instead, and the first boot makes it
# before any deferred step runs.
#
# The manifest is also the only source of the image's platform: until that boot,
# lunaos-hw-platform reports its platform instead of the build host's, so the
# initramfs, services and packages the build sets up are the target's. It reads
# the manifest by the same rules as lunaos_image_read_manifest below.
#
# Root always uses the fixed paths. Only a non-root test may move them under
# LUNAOS_IMAGE_ROOT, so no environment variable can make a live system defer,
# skip or replay its hardware setup.

lunaos_image_init() {
  if (( EUID == 0 )); then
    lunaos_image_root=""
  elif [[ -n ${LUNAOS_IMAGE_ROOT:-} ]]; then
    lunaos_image_root=$LUNAOS_IMAGE_ROOT
  else
    echo "Error: image setup state is root-owned; run as root" >&2
    return 1
  fi

  lunaos_image_dir=$lunaos_image_root/var/lib/lunaos/image
  lunaos_image_manifest=$lunaos_image_dir/target
  lunaos_image_queue=$lunaos_image_dir/deferred-steps
  lunaos_image_initramfs_baseline=$lunaos_image_dir/initramfs-inputs
  lunaos_image_boot_rebuild=$lunaos_image_dir/boot-rebuild
  lunaos_image_keyring_request=$lunaos_image_dir/pacman-keyring
  lunaos_image_unit=lunaos-provision-hardware.service
  lunaos_image_systemd_dir=$lunaos_image_root/etc/systemd/system
}

# Owned by whoever runs this (root on a live system), not a symlink, and not
# writable by anyone else.
lunaos_image_trusted() {
  local path=$1 owner mode

  [[ -e $path && ! -L $path ]] || return 1
  read -r owner mode < <(stat -c '%u %a' -- "$path") || return 1
  (( owner == EUID && (8#$mode & 8#022) == 0 ))
}

lunaos_image_manifest_present() {
  [[ -e $lunaos_image_manifest || -L $lunaos_image_manifest ]]
}

# Sets lunaos_image_platform, or fails without guessing whether this is a build.
lunaos_image_read_manifest() {
  local line key value format="" platform=""

  if [[ ! -d $lunaos_image_dir || ! -f $lunaos_image_manifest ]] || ! lunaos_image_trusted "$lunaos_image_dir" ||
    ! lunaos_image_trusted "$lunaos_image_manifest"; then
    echo "Error: $lunaos_image_manifest is not a root-owned regular file" >&2
    return 1
  fi

  while IFS= read -r line || [[ -n $line ]]; do
    [[ -n $line && $line != \#* ]] || continue
    if [[ $line != *=* ]]; then
      echo "Error: $lunaos_image_manifest is malformed: $line" >&2
      return 1
    fi
    key=${line%%=*}
    value=${line#*=}
    case $key in
      format) format=$value ;;
      platform) platform=$value ;;
    esac
  done <"$lunaos_image_manifest"

  if [[ $format != "1" ]]; then
    echo "Error: $lunaos_image_manifest is not format=1" >&2
    return 1
  fi
  # Builders written before the platform names settled say apple-silicon,
  # generic-aarch64, generic or qualcomm (as lunaos-hw-platform reads them too).
  case $platform in
    apple-silicon) platform=aarch64-apple ;;
    generic-aarch64) platform=aarch64 ;;
    generic) platform=x86 ;;
    qualcomm) platform=aarch64-qualcomm ;;
  esac
  case $platform in
    x86 | aarch64 | aarch64-apple | aarch64-qualcomm | aarch64-n1x | aarch64-gb10) ;;
    *)
      echo "Error: $lunaos_image_manifest names no known platform: ${platform:-none}" >&2
      return 1
      ;;
  esac

  lunaos_image_platform=$platform
}

# Holds the image state still for the rest of the caller, so hardware setup and
# a first-boot run never overlap or decide from a manifest the other retires.
# A machine that was never an image has no state and nothing to lock.
lunaos_image_lock() {
  [[ -e $lunaos_image_dir || -L $lunaos_image_dir ]] || return 0
  if [[ ! -d $lunaos_image_dir ]] || ! lunaos_image_trusted "$lunaos_image_dir"; then
    echo "Error: $lunaos_image_dir is not a root-owned directory" >&2
    return 1
  fi
  exec {lunaos_image_lock_fd}>>"$lunaos_image_dir/lock" || return 1
  flock "$lunaos_image_lock_fd"
}

lunaos_image_write_queue() {
  local tmp

  tmp=$(mktemp "$lunaos_image_dir/.deferred-steps.XXXXXX") || return 1
  if (( $# )); then
    printf '%s\n' "$@" >"$tmp" || { rm -f "$tmp"; return 1; }
  fi
  chmod 0644 "$tmp" && mv -f "$tmp" "$lunaos_image_queue"
}

# Queue every leaf install/hardware/all.sh would run, in its order, running
# none of them, then arm the first-boot service. Rerunning rewrites the same queue.
lunaos_image_defer_hardware() {
  local listing
  local -a steps=()

  listing=$(
    run_logged() {
      [[ $1 == "$LUNAOS_INSTALL"/hardware/*.sh && $1 != *..* ]] || {
        echo "Error: cannot defer a step outside the hardware setup: $1" >&2
        exit 1
      }
      printf 'install/%s\n' "${1#"$LUNAOS_INSTALL"/}"
    }
    source "$LUNAOS_INSTALL/hardware/all.sh"
  ) || return 1

  [[ -z $listing ]] || mapfile -t steps <<<"$listing"
  lunaos_image_write_queue "${steps[@]}" || return 1

  install -d -m 0755 "$lunaos_image_systemd_dir/multi-user.target.wants" &&
    install -m 0644 "$LUNAOS_INSTALL/provisioning/$lunaos_image_unit" "$lunaos_image_systemd_dir/$lunaos_image_unit" &&
    ln -sfn "/etc/systemd/system/$lunaos_image_unit" "$lunaos_image_systemd_dir/multi-user.target.wants/$lunaos_image_unit" ||
    return 1

  echo "Image build for $lunaos_image_platform: deferred ${#steps[@]} hardware steps to first boot"
}
