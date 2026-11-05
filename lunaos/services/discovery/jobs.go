package discovery

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"
)

const (
	maxAttempts = 5
	maxWait     = 5 * time.Minute
	// removalScope marks dead letters of a failed RemovalChecker call.
	removalScope = "#removals"
)

// Engine syncs every connector into the store. Providers are isolated: one
// provider's refusal, rate limit or failure never stops another.
type Engine struct {
	Store      *Store
	Connectors []Connector
	Policy     Policy
	Now        func() time.Time
	Sleep      func(context.Context, time.Duration) error
	// Download fetches an admitted candidate's media to a temp file and returns
	// its path and SHA-256. Nil means media candidates are rejected.
	Download func(ctx context.Context, c Candidate) (path, sha string, err error)
}

// Report carries only states and counters, never provider content.
type Report struct {
	PerProvider map[string]ProviderReport
}

type ProviderReport struct {
	State                                 string // ok|disconnected|rate_limited|failed
	Stored, Duplicates, Rejected, Removed int
}

var stateRank = map[string]int{"ok": 0, "failed": 1, "rate_limited": 2, "disconnected": 3}

func (r *ProviderReport) raise(state string) {
	if stateRank[state] > stateRank[r.State] {
		r.State = state
	}
}

// longLimit is a rate limit too long to wait out within this sync.
type longLimit struct{ until time.Time }

func (longLimit) Error() string { return "DISCOVERY_RATE_LIMITED" }

func (e *Engine) now() time.Time {
	if e.Now != nil {
		return e.Now()
	}
	return time.Now()
}

func (e *Engine) sleep(ctx context.Context, d time.Duration) error {
	if e.Sleep != nil {
		return e.Sleep(ctx, d)
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func backoff(attempt int) time.Duration {
	if attempt > 9 {
		return maxWait
	}
	return min(time.Second<<(attempt-1), maxWait)
}

// retry runs one request up to maxAttempts times. It returns nil, the context
// error, ErrUnauthorized (never retried), a longLimit, or the last error once
// the attempts are spent. A rate limit counts as an attempt.
func (e *Engine) retry(ctx context.Context, call func() error) error {
	for attempt := 1; ; attempt++ {
		err := call()
		if err == nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if errors.Is(err, ErrUnauthorized) {
			return ErrUnauthorized
		}
		wait := backoff(attempt)
		if rl, ok := asRateLimited(err); ok {
			if rl.RetryAfter > maxWait {
				return longLimit{until: e.now().Add(rl.RetryAfter)}
			}
			wait = max(rl.RetryAfter, 0)
		}
		if attempt == maxAttempts {
			return err
		}
		if err = e.sleep(ctx, wait); err != nil {
			return err
		}
	}
}

// Sync runs every connector once and reports per provider.
func (e *Engine) Sync(ctx context.Context) Report {
	rep := Report{PerProvider: map[string]ProviderReport{}}
	for _, c := range e.Connectors {
		if ctx.Err() != nil {
			break
		}
		rep.PerProvider[c.Name()] = e.syncProvider(ctx, c)
	}
	return rep
}

type providerRun struct {
	e    *Engine
	c    Connector
	name string
	pr   ProviderReport
}

func (e *Engine) syncProvider(ctx context.Context, c Connector) ProviderReport {
	p := &providerRun{e: e, c: c, name: c.Name(), pr: ProviderReport{State: "ok"}}
	scopes := c.Scopes()
	if len(scopes) == 0 {
		return p.pr
	}
	notBefore, err := e.Store.NotBefore(ctx, p.name)
	if err != nil {
		p.pr.raise("failed")
		return p.pr
	}
	if e.now().Before(notBefore) {
		p.pr.raise("rate_limited")
		return p.pr
	}
	if rc, ok := c.(RemovalChecker); ok && !p.checkRemovals(ctx, rc) {
		return p.pr
	}
	for _, scope := range scopes {
		if !p.syncScope(ctx, scope) {
			break
		}
	}
	return p.pr
}

// fail settles a request that could not complete. It returns false when the
// whole provider must stop for this sync, true when the run may go on after
// recording a dead letter.
func (p *providerRun) fail(ctx context.Context, err error, dl DeadLetter) bool {
	if ctx.Err() != nil {
		p.pr.raise("failed")
		return false
	}
	if errors.Is(err, ErrUnauthorized) {
		p.pr.raise("disconnected")
		return false
	}
	var ll longLimit
	if errors.As(err, &ll) {
		if setErr := p.e.Store.SetNotBefore(ctx, p.name, ll.until); setErr != nil {
			p.pr.raise("failed")
		}
		p.pr.raise("rate_limited")
		return false
	}
	p.pr.raise("failed")
	dl.Provider, dl.Code, dl.At = p.name, errorCode(err), p.e.now()
	return p.e.Store.AddDeadLetter(ctx, dl) == nil
}

func (p *providerRun) checkRemovals(ctx context.Context, rc RemovalChecker) bool {
	ids, err := p.e.Store.ExternalIDs(ctx, p.name)
	if err != nil {
		p.pr.raise("failed")
		return false
	}
	if len(ids) == 0 {
		return true
	}
	var gone []string
	err = p.e.retry(ctx, func() (err error) {
		gone, err = rc.Removed(ctx, ids)
		return err
	})
	if err != nil {
		return p.fail(ctx, err, DeadLetter{Scope: removalScope})
	}
	for _, id := range gone {
		if !p.remove(ctx, id) {
			return false
		}
	}
	return true
}

func (p *providerRun) remove(ctx context.Context, externalID string) bool {
	existed, err := p.e.Store.deleteItem(ctx, p.name, externalID)
	if err != nil {
		p.pr.raise("failed")
		return false
	}
	if existed {
		p.pr.Removed++
	}
	return true
}

// syncScope pages through one scope from its saved cursor. It returns false
// when the provider must stop.
func (p *providerRun) syncScope(ctx context.Context, scope string) bool {
	cursor, err := p.e.Store.Cursor(ctx, p.name, scope)
	if err != nil {
		p.pr.raise("failed")
		return false
	}
	seen := map[string]bool{cursor: true}
	for {
		if ctx.Err() != nil {
			p.pr.raise("failed")
			return false
		}
		var page Page
		err = p.e.retry(ctx, func() (err error) {
			page, err = p.c.Fetch(ctx, scope, cursor)
			return err
		})
		if err != nil {
			return p.fail(ctx, err, DeadLetter{Scope: scope, Cursor: cursor})
		}
		for _, cand := range page.Items {
			if !p.process(ctx, cand) {
				return false
			}
		}
		for _, id := range page.Removed {
			if !p.remove(ctx, id) {
				return false
			}
		}
		if err = p.e.Store.SaveCursor(ctx, p.name, scope, page.NextCursor); err != nil {
			p.pr.raise("failed")
			return false
		}
		// An empty, unchanged or already visited cursor ends the scope.
		if page.NextCursor == "" || seen[page.NextCursor] {
			return true
		}
		seen[page.NextCursor] = true
		cursor = page.NextCursor
	}
}

// process admits, downloads and stores one candidate. It returns false when the
// provider must stop.
func (p *providerRun) process(ctx context.Context, cand Candidate) bool {
	if ctx.Err() != nil {
		p.pr.raise("failed")
		return false
	}
	// A connector only ever writes into its own provider's namespace.
	cand.Provider = p.name
	if ok, _ := p.e.Policy.Admit(cand); !ok {
		p.pr.Rejected++
		return true
	}
	item := cand.Item
	if cand.MediaURL == "" {
		stored, err := p.e.Store.Put(ctx, item, nil)
		if err != nil {
			return p.fail(ctx, err, DeadLetter{ExternalID: cand.ExternalID})
		}
		p.count(stored)
		return true
	}
	if p.e.Download == nil {
		p.pr.Rejected++
		return true
	}
	var path, sha string
	err := p.e.retry(ctx, func() (err error) {
		path, sha, err = p.e.Download(ctx, cand)
		if err != nil && path != "" {
			_ = os.Remove(path)
		}
		return err
	})
	if err != nil {
		return p.fail(ctx, err, DeadLetter{ExternalID: cand.ExternalID})
	}
	defer os.Remove(path)
	if cand.DeclaredSHA256 != "" && !strings.EqualFold(cand.DeclaredSHA256, sha) {
		return p.fail(ctx, ErrHashMismatch, DeadLetter{ExternalID: cand.ExternalID})
	}
	f, err := os.Open(path)
	if err != nil {
		return p.fail(ctx, err, DeadLetter{ExternalID: cand.ExternalID})
	}
	item.SHA256 = sha
	stored, err := p.e.Store.Put(ctx, item, f)
	_ = f.Close()
	if err != nil {
		return p.fail(ctx, err, DeadLetter{ExternalID: cand.ExternalID})
	}
	p.count(stored)
	return true
}

func (p *providerRun) count(stored bool) {
	if stored {
		p.pr.Stored++
	} else {
		p.pr.Duplicates++
	}
}
