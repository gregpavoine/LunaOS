#!/bin/bash
iso_name="lunaos"
iso_label="LUNAOS_0_1_0"
iso_publisher="LunaOS"
iso_application="LunaOS Arch Linux Live Image"
iso_version="0.1.0"
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
