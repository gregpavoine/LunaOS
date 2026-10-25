//go:build !linux

package core

import (
	"fmt"
	"net"
)

type OwnerSocketListener struct{ net.Listener }

func ListenOwnerSocket(string) (*OwnerSocketListener, error) {
	return nil, fmt.Errorf("owner API Unix socket requires Linux")
}
