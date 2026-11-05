// Package peer reads the UID of the process on the other end of a Unix socket.
package peer

import (
	"context"
	"errors"
)

// ErrUnsupported is returned where Unix-socket peer credentials are not implemented.
var ErrUnsupported = errors.New("Unix-socket peer credentials are only implemented on Linux")

type contextKey struct{}

// WithUID stores a peer UID in ctx, as ConnContext does for a real connection.
func WithUID(ctx context.Context, uid int) context.Context {
	return context.WithValue(ctx, contextKey{}, uid)
}

// FromContext returns the peer UID stored by ConnContext or WithUID.
func FromContext(ctx context.Context) (int, bool) {
	uid, ok := ctx.Value(contextKey{}).(int)
	return uid, ok
}
