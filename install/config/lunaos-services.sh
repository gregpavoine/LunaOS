# LunaOS control plane. The lunaos-control-plane package ships the Core, the
# Policy Broker, the first-boot credential unit and their sysusers accounts.
# Nothing here applies until that package is installed: it has no aarch64
# build yet, and a missing package must not leave dangling unit enablement.
#
# Units are enabled, not started. Installs reboot, and luna-owner-setup must
# create the owner and Broker credentials before the Broker and Core start;
# that ordering lives in the units themselves.
#
# The install user joins luna-owner, the group allowed to open the Core's owner
# socket. Membership is the desktop's authority over LUNA, so it goes to the
# machine's owner only. Deferred-provisioning installs have no user yet and
# record the group for omarchy-provision-owner instead.
if omarchy-pkg-present lunaos-control-plane; then
  systemctl enable luna-owner-setup.service luna-policy-broker.service luna-core.service

  if [[ -n ${OMARCHY_INSTALL_USER:-} ]]; then
    usermod -aG luna-owner "$OMARCHY_INSTALL_USER"
  else
    lunaos_provisioning_dir=${OMARCHY_PROVISIONING_DIR:-/var/lib/omarchy/provisioning}
    mkdir -p "$lunaos_provisioning_dir"
    if ! grep -qx luna-owner "$lunaos_provisioning_dir/groups" 2>/dev/null; then
      echo luna-owner >>"$lunaos_provisioning_dir/groups"
    fi
  fi
fi
