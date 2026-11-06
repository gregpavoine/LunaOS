# The platform packages' own setup, last so it builds on every other leaf
# (lunaos-lifecycle-dispatch setup-boot and setup-system, see
# docs/lifecycle-dispatch.md): the boot package's first, then the runtime
# package's. A no-op where the platform registers none. An image's first boot
# says so: there the network may be down, and a boot-file rebuild is asked for
# rather than built, since the deferred setup rebuilds once after its last step.
if [[ ${LUNAOS_IMAGE_DEFERRED_HARDWARE:-} == "1" ]]; then
  lunaos-lifecycle-dispatch setup-boot image-first-boot
  lunaos-lifecycle-dispatch setup-system image-first-boot
else
  lunaos-lifecycle-dispatch setup-boot
  lunaos-lifecycle-dispatch setup-system
fi
