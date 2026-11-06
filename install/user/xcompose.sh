# Set default XCompose that is triggered with Shift + CapsLock
tee ~/.XCompose >/dev/null <<EOF
# Run lunaos-restart-xcompose to apply changes

# Include fast emoji access
include "/usr/share/lunaos/default/xcompose"

# Identification
<Multi_key> <space> <n> : "$LUNAOS_USER_NAME"
<Multi_key> <space> <e> : "$LUNAOS_USER_EMAIL"
EOF
