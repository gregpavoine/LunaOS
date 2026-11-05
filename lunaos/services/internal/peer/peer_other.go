//go:build !linux

package peer

import (
	"context"
	"net"
)

// UID is unsupported off Linux.
func UID(net.Conn) (int, error) { return -1, ErrUnsupported }

// ConnContext records -1 as the peer UID, so callers fail closed.
func ConnContext(ctx context.Context, _ net.Conn) context.Context { return WithUID(ctx, -1) }
