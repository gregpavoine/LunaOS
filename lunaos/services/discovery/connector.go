package discovery

import (
	"context"
	"errors"
	"time"
)

// Candidate is one provider result before policy admission. MediaURL is empty
// for metadata-only results; nothing is downloaded for them.
type Candidate struct {
	Item
	MediaURL       string
	DeclaredSHA256 string
	MinorFlag      bool
}

// Page is one response of a scoped, cursor-paginated provider query. Removed
// lists external IDs the provider reports as taken down.
type Page struct {
	Items      []Candidate
	NextCursor string
	Removed    []string
}

// Connector queries one provider's official API. Scopes are the allowlist
// entries it will query; no scopes means no request at all.
type Connector interface {
	Name() string
	Scopes() []string
	Fetch(ctx context.Context, scope, cursor string) (Page, error)
}

// RemovalChecker is implemented by connectors that can tell which already
// stored external IDs the provider has taken down.
type RemovalChecker interface {
	Removed(ctx context.Context, externalIDs []string) ([]string, error)
}

// ErrRateLimited asks the engine to wait RetryAfter before the next request.
type ErrRateLimited struct{ RetryAfter time.Duration }

func (ErrRateLimited) Error() string { return "DISCOVERY_RATE_LIMITED" }

var (
	// ErrUnauthorized means the credentials were refused; it is never retried.
	ErrUnauthorized = errors.New("DISCOVERY_UNAUTHORIZED")
	// ErrTransient is a retryable failure (network, 5xx).
	ErrTransient = errors.New("DISCOVERY_TRANSIENT")
	// ErrCursorExpired means the provider refuses to continue a saved cursor;
	// the scope restarts from the beginning on the next sync.
	ErrCursorExpired = errors.New("DISCOVERY_CURSOR_EXPIRED")

	// errTakedownGuard refuses a removal check that would delete most of the
	// provider's stored items in one sync.
	errTakedownGuard = errors.New("DISCOVERY_TAKEDOWN_GUARD")
)

func asRateLimited(err error) (ErrRateLimited, bool) {
	var v ErrRateLimited
	if errors.As(err, &v) {
		return v, true
	}
	var p *ErrRateLimited
	if errors.As(err, &p) && p != nil {
		return *p, true
	}
	return ErrRateLimited{}, false
}

// errorCode maps an error to a short code. Unclassified errors never surface
// their text, which may carry a URL or other provider data.
func errorCode(err error) string {
	switch {
	case errors.Is(err, ErrUnauthorized):
		return "unauthorized"
	case errors.Is(err, ErrHashMismatch):
		return "hash_mismatch"
	case errors.Is(err, ErrTransient):
		return "transient"
	case errors.Is(err, ErrCursorExpired):
		return "cursor_expired"
	case errors.Is(err, errTakedownGuard):
		return "takedown_guard"
	}
	if _, ok := asRateLimited(err); ok {
		return "rate_limited"
	}
	return "error"
}
