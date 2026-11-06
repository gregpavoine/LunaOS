#!/bin/bash

cat <<EOF | \
  sudo tee /etc/lunaos/review.conf
ExecStart=$HOME/.local/bin/payload
EOF
