#!/bin/bash
set -euo pipefail

if [[ $(uname -s) != "Linux" || $(uname -m) != "x86_64" ]]; then
  printf '%s\n' 'LunaOS archiso builds require an Arch Linux x86_64 builder.' >&2
  exit 2
fi
for command_name in go mkarchiso; do
  if ! command -v "$command_name" >/dev/null 2>&1; then
    printf 'Missing required Arch build tool: %s\n' "$command_name" >&2
    exit 2
  fi
done

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
lunaos_dir=$(cd -- "$script_dir/.." && pwd)
repo_dir=$(cd -- "$lunaos_dir/.." && pwd)
output_dir=${LUNAOS_ISO_OUTPUT:-"$repo_dir/out/archiso"}
work_dir=${LUNAOS_ISO_WORK:-"$repo_dir/.archiso-work"}
profile_source="$lunaos_dir/packaging/arch"
staging=$(mktemp -d "${TMPDIR:-/tmp}/lunaos-archiso.XXXXXX")
trap 'rm -rf -- "$staging"' EXIT
mkdir -p "$staging/airootfs/usr/bin"

cd "$lunaos_dir"
CGO_ENABLED=1 go build -buildvcs=false -trimpath -o "$staging/airootfs/usr/bin/luna-core" "$lunaos_dir/services/core/cmd/luna-core"
CGO_ENABLED=1 go build -buildvcs=false -trimpath -o "$staging/airootfs/usr/bin/luna-policy-broker" "$lunaos_dir/services/policy"
CGO_ENABLED=1 go build -buildvcs=false -trimpath -o "$staging/airootfs/usr/bin/luna-owner-setup" "$lunaos_dir/services/identity/cmd/luna-owner-setup"
CGO_ENABLED=1 go build -buildvcs=false -trimpath -o "$staging/airootfs/usr/bin/luna-vault" "$lunaos_dir/services/intimacy/cmd/luna-vault"
cp -a "$profile_source/." "$staging/"
install -Dm644 "$lunaos_dir/systemd/luna-core.service" "$staging/airootfs/usr/lib/systemd/system/luna-core.service"
install -Dm644 "$lunaos_dir/systemd/luna-owner-setup.service" "$staging/airootfs/usr/lib/systemd/system/luna-owner-setup.service"
install -Dm644 "$lunaos_dir/systemd/luna-policy-broker.service" "$staging/airootfs/usr/lib/systemd/system/luna-policy-broker.service"
install -Dm644 "$lunaos_dir/systemd-sysusers/lunaos.conf" "$staging/airootfs/usr/lib/sysusers.d/lunaos.conf"
install -Dm644 "$lunaos_dir/packaging/arch/lunaos.desktop" "$staging/airootfs/usr/share/applications/lunaos.desktop"
install -Dm644 "$lunaos_dir/README.md" "$staging/airootfs/usr/share/doc/lunaos/README.md"
install -d "$staging/airootfs/usr/share/lunaos/shell"
cp -a "$lunaos_dir/shell/." "$staging/airootfs/usr/share/lunaos/shell/"

mkdir -p "$output_dir" "$work_dir"
sudo mkarchiso -v -w "$work_dir" -o "$output_dir" "$staging"
printf 'ISO output directory: %s\n' "$output_dir"
