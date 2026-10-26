#!/bin/bash
set -euo pipefail

chmod 0755 /usr/bin/luna-core /usr/bin/luna-policy-broker /usr/bin/luna-owner-setup /usr/bin/luna-vault
chmod 0755 /usr/local/bin/lunaos-install
chmod 0440 /etc/sudoers.d/lunaos-install
visudo -c -f /etc/sudoers.d/lunaos-install
systemd-sysusers /usr/lib/sysusers.d/lunaos.conf
useradd --create-home --groups wheel,luna-owner luna
passwd --lock luna
install -d -o luna -g luna -m 0755 /home/luna/.config/hypr
cp /etc/skel/.config/hypr/hyprland.conf /home/luna/.config/hypr/hyprland.conf
chown luna:luna /home/luna/.config/hypr/hyprland.conf
chmod 0644 /home/luna/.config/hypr/hyprland.conf
systemctl enable NetworkManager.service
systemctl set-default graphical.target
systemctl enable greetd.service
systemctl enable luna-owner-setup.service
systemctl enable luna-policy-broker.service
systemctl enable luna-core.service
