echo "Repair fingerprint support left without a libfprint"

# An earlier version of this migration swapped libfprint-git for stock
# libfprint in two steps. A run that failed between them left fprintd with
# no library; finish with the driver the fingerprint setup installs now.
if lunaos-pkg-present fprintd && lunaos-pkg-missing libfprint && lunaos-pkg-missing libfprint-git; then
  lunaos-pkg-add libfprint-git
fi
