#!/bin/bash

# Builds lunaos-control-plane, signs it and its repository database with the
# LunaOS package key, then proves pacman accepts only what that key signed.
# Runs as root inside a disposable Arch Linux (ARM) container:
#
#   build-signed-repo.sh <source-root> <output-dir>
#
# <source-root> is this repository; the private key is read from
# $LUNAOS_SIGNING_KEY_FILE (default /signing.asc) and must be the key published
# in lunaos/packaging/keys/lunaos-packages.asc. The output holds the package for
# an image bundle and a signed repository under repo/<arch>/.

set -euo pipefail

src=${1:?source root}
out=${2:?output directory}
key_file=${LUNAOS_SIGNING_KEY_FILE:-/signing.asc}
public_key="$src/lunaos/packaging/keys/lunaos-packages.asc"
arch=$(uname -m)

fail() {
  echo "ERROR: $*" >&2
  exit 1
}

fingerprint_of() {
  gpg --batch --with-colons "$@" 2>/dev/null | awk -F: '/^fpr:/ { print $10; exit }'
}

# pacman 7 downloads as the alpm user inside a Landlock sandbox, which an
# unprivileged container cannot apply ("restricting filesystem access failed").
# The container is disposable and every package is still signature-checked.
grep -qx DisableSandbox /etc/pacman.conf || sed -i '/^\[options\]/a DisableSandbox' /etc/pacman.conf

pacman-key --init >/dev/null
if [[ $arch == "aarch64" ]]; then
  pacman-key --populate archlinuxarm >/dev/null
else
  pacman-key --populate archlinux >/dev/null
fi
pacman -Syu --noconfirm --needed base-devel go git >/dev/null

# The secret must be the published key, or every image built from it would
# carry packages that a correctly configured pacman refuses.
GNUPGHOME=$(mktemp -d)
export GNUPGHOME
chmod 700 "$GNUPGHOME"
gpg --batch --quiet --import "$key_file"
signing_fpr=$(fingerprint_of --list-secret-keys)
published_fpr=$(fingerprint_of --show-keys "$public_key")
[[ -n $signing_fpr ]] || fail "no secret key in $key_file"
[[ $signing_fpr == "$published_fpr" ]] ||
  fail "signing key $signing_fpr is not the published LunaOS key $published_fpr"

# makepkg never runs as root. Build from a writable copy of lunaos/.
id builder &>/dev/null || useradd -m builder
work=/home/builder/lunaos
rm -rf "$work"
cp -a "$src/lunaos" "$work"
chown -R builder:builder /home/builder
# Arch Linux ARM's makepkg.conf writes .pkg.tar.xz; pin the destination and
# the zstd format so both architectures produce the same artifact names.
packages=/home/builder/packages
mkdir -p "$packages"
chown -R builder:builder /home/builder
su builder -c "cd '$work/packaging/control-plane' && PKGDEST='$packages' PKGEXT=.pkg.tar.zst makepkg --noconfirm --cleanbuild"

package=$(find "$packages" -maxdepth 1 -name "lunaos-control-plane-[0-9]*-$arch.pkg.tar.zst" -print -quit)
[[ -n $package ]] || fail "makepkg produced no lunaos-control-plane package"
package_name=${package##*/}

repo="$out/repo/$arch"
mkdir -p "$repo"
cp "$package" "$repo/"
gpg --batch --yes --local-user "$signing_fpr" --detach-sign "$repo/$package_name"
repo-add --sign --key "$signing_fpr" "$repo/lunaos.db.tar.gz" "$repo/$package_name"
cp "$package" "$out/"

# pacman must trust the published key and nothing else: a correctly signed
# repository installs, an unsigned one and one signed by another key do not.
pacman-key --add "$public_key" >/dev/null
pacman-key --lsign-key "$published_fpr" >/dev/null

try_repo() {
  local dir=$1 root
  root=$(mktemp -d)
  mkdir -p "$root/db" "$root/cache"
  cat >"$root/pacman.conf" <<CONF
[options]
Architecture = $arch
SigLevel = Required DatabaseRequired
DisableSandbox

[lunaos]
Server = file://$dir
CONF
  pacman --config "$root/pacman.conf" --dbpath "$root/db" --cachedir "$root/cache" --noconfirm -Sy >/dev/null 2>&1 &&
    pacman --config "$root/pacman.conf" --dbpath "$root/db" --cachedir "$root/cache" --noconfirm -Sw lunaos-control-plane >/dev/null 2>&1
}

try_repo "$repo" || fail "pacman refused the signed LunaOS repository"
echo "ok - pacman accepts the repository signed by $published_fpr"

unsigned=$(mktemp -d)
cp "$package" "$unsigned/"
repo-add "$unsigned/lunaos.db.tar.gz" "$unsigned/$package_name" >/dev/null
if try_repo "$unsigned"; then
  fail "pacman accepted an unsigned LunaOS repository"
fi
echo "ok - pacman refuses an unsigned repository"

foreign=$(mktemp -d)
cp "$package" "$foreign/"
cat >"$GNUPGHOME/foreign" <<'PARAMS'
%no-protection
Key-Type: eddsa
Key-Curve: ed25519
Key-Usage: sign
Name-Real: Not LunaOS
Expire-Date: 1d
%commit
PARAMS
gpg --batch --quiet --gen-key "$GNUPGHOME/foreign"
foreign_fpr=$(gpg --batch --with-colons --list-secret-keys "Not LunaOS" | awk -F: '/^fpr:/ { print $10; exit }')
gpg --batch --yes --local-user "$foreign_fpr" --detach-sign "$foreign/$package_name"
repo-add --sign --key "$foreign_fpr" "$foreign/lunaos.db.tar.gz" "$foreign/$package_name" >/dev/null
if try_repo "$foreign"; then
  fail "pacman accepted a repository signed by a key other than LunaOS's"
fi
echo "ok - pacman refuses a repository signed by another key"

rm -rf "$GNUPGHOME"
(cd "$out" && sha256sum "$package_name" >"$package_name.sha256")
ls -lR "$out"
