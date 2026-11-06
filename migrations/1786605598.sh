echo "Rebuild the initramfs so NVIDIA-only systems shed nouveau's unused GSP firmware"

# lunaos_hooks.conf now filters the kms hook out of HOOKS when the proprietary
# NVIDIA driver handles early KMS and NVIDIA owns every display controller
# (#6790). The settings package deploys that conditional, but nothing rebuilds
# the initramfs for a mkinitcpio drop-in change, so affected machines would
# carry ~100 MB of dead nouveau firmware until their next kernel update.
# Rebuild once, and only where the conditional actually changes the outcome:
# evaluate the installed drop-ins the way mkinitcpio does and check that kms
# dropped out. A user-edited lunaos_hooks.conf (pacman leaves the packaged
# update as a .pacnew) keeps kms and correctly skips the rebuild.

hooks_conf="${LUNAOS_MKINITCPIO_HOOKS_CONF:-/etc/mkinitcpio.conf.d/lunaos_hooks.conf}"
baseline_conf="${hooks_conf%/*}/00-lunaos-hooks.conf"
nvidia_conf="${LUNAOS_MKINITCPIO_NVIDIA_CONF:-/etc/mkinitcpio.conf.d/nvidia.conf}"
rebuild_marker="${LUNAOS_KMS_REBUILD_MARKER:-/var/lib/lunaos/migrations/1786605598}"

lunaos-cmd-present limine-mkinitcpio || exit 0
[[ -f $hooks_conf && -f $nvidia_conf ]] || exit 0

# The rebuild is machine-wide, but migrations run once per user: a marker
# records completion so another user's run does not repeat it, while a missing
# marker still retries an interrupted rebuild.
[[ ! -e $rebuild_marker ]] || exit 0

# Include the baseline when installed, then the hardware and filter drop-ins.
# Skip conservatively if evaluation fails.
hooks=$(bash -c 'if [[ -f $1 ]]; then source "$1" || exit 1; fi; source "$2" && source "$3" && echo " ${HOOKS[*]} "' -- "$baseline_conf" "$nvidia_conf" "$hooks_conf") || exit 0

[[ $hooks != *" kms "* ]] || exit 0

echo "This machine no longer uses the kms hook; rebuilding the initramfs without nouveau"
sudo limine-mkinitcpio
sudo install -Dm644 /dev/null "$rebuild_marker"
