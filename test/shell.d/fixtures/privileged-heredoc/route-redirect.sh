# A plain redirect into /etc, no sudo: the command re-execs itself as root.
cat >/etc/lunaos/agent.conf <<EOF
helper=$HOME/.local/share/lunaos/bin/lunaos-agent
EOF
