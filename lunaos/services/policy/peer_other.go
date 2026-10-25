//go:build !linux

package main

import (
	"errors"
	"net"
)

func unixPeerUID(net.Conn) (int, error) {
	return -1, errors.New("Unix-socket peer credentials are only implemented on Linux")
}
