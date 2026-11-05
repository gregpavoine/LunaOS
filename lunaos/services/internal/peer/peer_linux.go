//go:build linux

package peer

import (
	"context"
	"fmt"
	"net"
	"syscall"
)

// UID returns the SO_PEERCRED UID of the peer of a Unix-socket connection.
func UID(conn net.Conn) (int, error) {
	unixConn, ok := conn.(*net.UnixConn)
	if !ok {
		return -1, fmt.Errorf("Unix socket connection required")
	}
	raw, err := unixConn.SyscallConn()
	if err != nil {
		return -1, err
	}
	uid := -1
	var controlErr error
	if err := raw.Control(func(fd uintptr) {
		credential, err := syscall.GetsockoptUcred(int(fd), syscall.SOL_SOCKET, syscall.SO_PEERCRED)
		if err != nil {
			controlErr = err
			return
		}
		uid = int(credential.Uid)
	}); err != nil {
		return -1, err
	}
	if controlErr != nil {
		return -1, controlErr
	}
	return uid, nil
}

// ConnContext is an http.Server ConnContext hook. It records the peer UID, or -1 when it
// cannot be read, so callers fail closed.
func ConnContext(ctx context.Context, c net.Conn) context.Context {
	uid, err := UID(c)
	if err != nil {
		return WithUID(ctx, -1)
	}
	return WithUID(ctx, uid)
}
