#!/bin/bash
iso_name="lunaos"
iso_label="LUNAOS_0_1_1"
iso_publisher="LunaOS"
iso_application="LunaOS Arch Linux Installable Preview"
iso_version="0.1.1"
install_dir="lunaos"
buildmodes=('iso')
bootmodes=('uefi.systemd-boot')
arch=('x86_64')
pacman_conf='pacman.conf'
airootfs_image_type="squashfs"
file_permissions=(
  ["/root"]="0:0:750"
  ["/etc/shadow"]="0:0:400"
)
