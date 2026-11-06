#!/bin/bash

# lunaos:heredoc-expands paths=none -- review regression fixture
sudo tee /etc/lunaos/review.conf >/dev/null <<EOF
ExecStart=${target:-$HOME/.local/bin/payload}
EOF
