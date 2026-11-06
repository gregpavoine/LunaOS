if true; then
  cat <<-EOF | sudo tee /etc/lunaos/indented.conf >/dev/null
	helper=$HOME/.local/share/lunaos/bin/lunaos-agent
	EOF
fi
