#!/bin/bash
set -euo pipefail

if (( $# != 2 )); then
  printf 'Usage: %s /path/to/release-archinstall.json /new/private/output-directory\n' "$0" >&2
  exit 2
fi

template=$1
output_dir=$2
for command_name in jq openssl; do
  if ! command -v "$command_name" >/dev/null 2>&1; then
    printf 'Missing install-smoke config tool: %s\n' "$command_name" >&2
    exit 2
  fi
done
if [[ ! -r $template ]]; then
  printf 'Archinstall template is not readable: %s\n' "$template" >&2
  exit 2
fi
if [[ -e $output_dir ]]; then
  printf 'Refusing to overwrite existing install-smoke output: %s\n' "$output_dir" >&2
  exit 2
fi
jq empty "$template"
mkdir -m 0700 -- "$output_dir"

password_file="$output_dir/root-password"
openssl rand -hex 32 > "$password_file"
chmod 0600 "$password_file"
password_hash=$(openssl passwd -6 -stdin < "$password_file")
if [[ $password_hash != '$6$'* ]]; then
  printf '%s\n' 'OpenSSL did not produce a SHA-512 crypt password hash.' >&2
  exit 1
fi

jq \
  --arg password_hash "$password_hash" \
  '
  .disk_config = {
    config_type: "default_layout",
    device_modifications: [
      {
        device: "/dev/vda",
        wipe: true,
        partitions: [
          {
            btrfs: [],
            dev_path: null,
            flags: ["boot"],
            fs_type: "fat32",
            mount_options: [],
            mountpoint: "/boot",
            obj_id: "2c3fa2d5-2c79-4fab-86ec-22d0ea1543c0",
            size: {
              sector_size: {unit: "B", value: 512},
              unit: "MiB",
              value: 512
            },
            start: {
              sector_size: {unit: "B", value: 512},
              unit: "MiB",
              value: 1
            },
            status: "create",
            type: "primary"
          },
          {
            btrfs: [],
            dev_path: null,
            flags: [],
            fs_type: "ext4",
            mount_options: [],
            mountpoint: "/",
            obj_id: "3e7018a0-363b-4d05-ab83-8e82d13db208",
            size: {
              sector_size: {unit: "B", value: 512},
              unit: "GiB",
              value: 20
            },
            start: {
              sector_size: {unit: "B", value: 512},
              unit: "MiB",
              value: 513
            },
            status: "create",
            type: "primary"
          }
        ]
      }
    ]
  }
  | .silent = true
  | .offline = true
  | .no_pkg_lookups = true
  | del(.auth_config, .users, .disk_encryption)
  ' "$template" > "$output_dir/user_configuration.json"
jq -n \
  --arg password_hash "$password_hash" \
  '{root_enc_password: $password_hash, users: [{username: "lunaos-ci", enc_password: $password_hash, sudo: true}]}' \
  > "$output_dir/user_credentials.json"
chmod 0600 "$output_dir/user_configuration.json" "$output_dir/user_credentials.json"
unset password_hash

printf 'Generated disposable Archinstall inputs in %s\n' "$output_dir"
