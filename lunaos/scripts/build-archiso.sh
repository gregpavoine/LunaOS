#!/bin/bash
set -euo pipefail

if [[ $(uname -s) != "Linux" || $(uname -m) != "x86_64" ]]; then
  printf '%s\n' 'LunaOS archiso builds require an Arch Linux x86_64 builder.' >&2
  exit 2
fi
for command_name in go mkarchiso makepkg runuser pacman repo-add jq; do
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
package_build=$(mktemp -d "${TMPDIR:-/tmp}/lunaos-package.XXXXXX")
trap 'rm -rf -- "$staging" "$package_build"' EXIT
mkdir -p "$staging/airootfs/usr/bin"
mkdir -p "$staging/airootfs/etc/sudoers.d" "$staging/airootfs/usr/local/bin" \
  "$staging/airootfs/usr/share/applications" "$staging/airootfs/usr/share/lunaos"

# Build the target package and its complete dependency closure into an
# ISO-local Pacman repository. makepkg must never run as root.
mkdir -p "$package_build/lunaos/packaging/arch"
mkdir -p "$package_build/home" "$package_build/gocache" "$package_build/gomodcache"
cp "$profile_source/PKGBUILD" "$package_build/lunaos/packaging/arch/PKGBUILD"
cp -a "$profile_source/install" "$package_build/lunaos/packaging/arch/install"
cp -a "$lunaos_dir/go.mod" "$lunaos_dir/go.sum" "$lunaos_dir/services" "$lunaos_dir/systemd" \
  "$lunaos_dir/systemd-sysusers" "$lunaos_dir/shell" "$package_build/lunaos/"
cp "$lunaos_dir/README.md" "$package_build/lunaos/README.md"
chown -R nobody:nobody "$package_build"
runuser -u nobody -- env HOME="$package_build/home" GOCACHE="$package_build/gocache" \
  GOMODCACHE="$package_build/gomodcache" \
  bash -euo pipefail -c 'cd "$1/packaging/arch" && makepkg --noconfirm' _ "$package_build/lunaos"
package_file=$(find "$package_build/lunaos/packaging/arch" -maxdepth 1 -type f -name 'lunaos-native-*.pkg.tar.*' -print -quit)
if [[ -z $package_file ]]; then
  printf '%s\n' 'makepkg did not produce lunaos-native.' >&2
  exit 1
fi

local_repo="$staging/lunaos-repo/lunaos-local/os/x86_64"
package_cache="$staging/package-cache"
package_db="$staging/package-db"
mkdir -p "$local_repo" "$package_cache" "$package_db"
cp "$package_file" "$local_repo/"
repo-add "$local_repo/lunaos-local.db.tar.gz" "$local_repo"/lunaos-native-*.pkg.tar.*
mapfile -t target_packages < <(awk 'NF && $1 !~ /^#/ { print $1 }' "$profile_source/packages.install.x86_64")
package_json=$(printf '%s\n' "${target_packages[@]}" | jq -Rsc 'split("\n") | map(select(length > 0))')
jq --argjson packages "$package_json" '.packages = $packages' \
  "$profile_source/airootfs/usr/share/lunaos/archinstall.json" \
  > "$staging/archinstall.json"
install -Dm644 "$staging/archinstall.json" "$staging/airootfs/usr/share/lunaos/archinstall.json"
install -Dm644 "$target_packages" "$staging/airootfs/usr/share/lunaos/target-packages.txt"

cat /etc/pacman.conf > "$staging/pacman-install.conf"
cat >> "$staging/pacman-install.conf" <<EOF

[lunaos-local]
SigLevel = Optional TrustAll
Server = file://$local_repo
EOF
pacman -Syw --noconfirm --needed --cachedir "$package_cache" --dbpath "$package_db" \
  --config "$staging/pacman-install.conf" "${target_packages[@]}"
target_core="$staging/airootfs/usr/share/lunaos/repo/core/os/x86_64"
target_extra="$staging/airootfs/usr/share/lunaos/repo/extra/os/x86_64"
target_local="$staging/airootfs/usr/share/lunaos/repo/lunaos-local/os/x86_64"
mkdir -p "$target_core" "$target_extra" "$target_local"
mapfile -d '' cached_packages < <(find "$package_cache" -maxdepth 1 -type f -name '*.pkg.tar.*' ! -name '*.sig' -print0)
if ((${#cached_packages[@]} == 0)); then
  printf '%s\n' 'Pacman did not download any target packages.' >&2
  exit 1
fi
for cached_package in "${cached_packages[@]}"; do
  package_basename=${cached_package##*/}
  package_name=$(pacman -Qp --qf '%n' "$cached_package")
  source_repo=$(pacman -Si --config "$staging/pacman-install.conf" "$package_name" | awk -F ': ' '/^Repository/ { print $2; exit }')
  case $source_repo in
    core) package_destination=$target_core ;;
    extra) package_destination=$target_extra ;;
    lunaos-local) package_destination=$target_local ;;
    *)
      printf 'Refusing to stage package %s from unexpected repository %s.\n' "$package_basename" "$source_repo" >&2
      exit 1
      ;;
  esac
  cp "$cached_package" "$package_destination/"
  if [[ -r $cached_package.sig ]]; then
    cp "$cached_package.sig" "$package_destination/"
  fi
done
cp "$package_file" "$target_local/"
mapfile -d '' core_packages < <(find "$target_core" -maxdepth 1 -type f -name '*.pkg.tar.*' ! -name '*.sig' -print0)
mapfile -d '' extra_packages < <(find "$target_extra" -maxdepth 1 -type f -name '*.pkg.tar.*' ! -name '*.sig' -print0)
if ((${#core_packages[@]} == 0 || ${#extra_packages[@]} == 0)); then
  printf '%s\n' 'No signed Arch target packages were staged in the local mirrors.' >&2
  exit 1
fi
repo-add "$target_core/core.db.tar.gz" "${core_packages[@]}"
repo-add "$target_extra/extra.db.tar.gz" "${extra_packages[@]}"
repo-add "$target_local/lunaos-local.db.tar.gz" "$target_local"/*.pkg.tar.*

install -Dm755 "$profile_source/airootfs/usr/local/bin/lunaos-install" "$staging/airootfs/usr/local/bin/lunaos-install"
install -Dm644 "$profile_source/airootfs/usr/share/applications/lunaos-install.desktop" "$staging/airootfs/usr/share/applications/lunaos-install.desktop"
install -Dm440 "$profile_source/airootfs/etc/sudoers.d/lunaos-install" "$staging/airootfs/etc/sudoers.d/lunaos-install"

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
install -Dm644 "$staging/archinstall.json" "$staging/airootfs/usr/share/lunaos/archinstall.json"
install -Dm644 "$lunaos_dir/packaging/arch/lunaos.desktop" "$staging/airootfs/usr/share/applications/lunaos.desktop"
install -Dm644 "$lunaos_dir/README.md" "$staging/airootfs/usr/share/doc/lunaos/README.md"
install -d "$staging/airootfs/usr/share/lunaos/shell"
cp -a "$lunaos_dir/shell/." "$staging/airootfs/usr/share/lunaos/shell/"

mkdir -p "$output_dir" "$work_dir"
sudo mkarchiso -v -w "$work_dir" -o "$output_dir" "$staging"
printf 'ISO output directory: %s\n' "$output_dir"
