// Package peer reads the UID of the process on the other end of a Unix socket.
package peer

import (
	"context"
	"errors"
)

// ErrUnsupported is returned where Unix-socket peer credentials are not implemented.
var ErrUnsupported = errors.New("Unix-socket peer credentials are only implemented on Linux")

type contextKey struct{}

// WithUID stores a peer UID in ctx, as ConnContext does for a real connection. It is reserved
// for ConnContext and tests: production code must not call it to vouch for a request.
func WithUID(ctx context.Context, uid int) context.Context {
	return context.WithValue(ctx, contextKey{}, uid)
}

// FromContext returns the peer UID stored by ConnContext or WithUID.
func FromContext(ctx context.Context) (int, bool) {
	uid, ok := ctx.Value(contextKey{}).(int)
	return uid, ok
}

// Allowed reports whether the peer recorded in ctx is the UID uid. It fails closed: it is false
// when uid is negative, when ctx holds no peer UID, and when the stored UID is negative (a failed
// credential read), so two "unknown" values never match.
func Allowed(ctx context.Context, uid int) bool {
	if uid < 0 {
		return false
	}
	stored, ok := FromContext(ctx)
	if !ok || stored < 0 {
		return false
	}
	return stored == uid
}
