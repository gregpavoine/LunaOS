# `>|` is a plain redirect with noclobber overridden, not a redirect into a pipe.
cat >|/etc/lunaos/agent.conf <<EOF
helper=$HOME/.local/share/lunaos/bin/lunaos-agent
EOF
