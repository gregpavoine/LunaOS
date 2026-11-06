package discovery

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"
)

// fakeClock drives Engine.Now and Engine.Sleep: sleeping advances the clock.
type fakeClock struct {
	mu     sync.Mutex
	t      time.Time
	sleeps []time.Duration
}

func newFakeClock() *fakeClock { return &fakeClock{t: time.Unix(1_800_000_000, 0)} }

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

func (c *fakeClock) Sleep(ctx context.Context, d time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sleeps = append(c.sleeps, d)
	c.t = c.t.Add(d)
	return nil
}

func (c *fakeClock) Sleeps() []time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]time.Duration(nil), c.sleeps...)
}

type fetchCall struct {
	Scope, Cursor string
	At            time.Time
}

// fakeConnector is scripted by the test through fetch; the engine is under test.
type fakeConnector struct {
	name   string
	scopes []string
	clock  *fakeClock
	fetch  func(scope, cursor string, n int) (Page, error)
	mu     sync.Mutex
	calls  []fetchCall
}

func (f *fakeConnector) Name() string     { return f.name }
func (f *fakeConnector) Scopes() []string { return f.scopes }

func (f *fakeConnector) Fetch(ctx context.Context, scope, cursor string) (Page, error) {
	f.mu.Lock()
	f.calls = append(f.calls, fetchCall{scope, cursor, f.clock.Now()})
	n := len(f.calls)
	f.mu.Unlock()
	return f.fetch(scope, cursor, n)
}

func (f *fakeConnector) Calls() []fetchCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]fetchCall(nil), f.calls...)
}

type removalConnector struct {
	*fakeConnector
	removed func(ids []string) []string
	asked   [][]string
}

func (r *removalConnector) Removed(ctx context.Context, ids []string) ([]string, error) {
	r.asked = append(r.asked, append([]string(nil), ids...))
	return r.removed(ids), nil
}

func meta(provider, id string) Candidate {
	return Candidate{Item: Item{Provider: provider, ExternalID: id, ContentLevel: "adult"}}
}

func media(provider, id string) Candidate {
	c := meta(provider, id)
	c.MediaURL = "https://media.example.test/" + id
	c.License = "cc-by"
	return c
}

// downloader writes the candidate's external id as its media and records every
// temp path so tests can check none is left behind.
type downloader struct {
	t     *testing.T
	dir   string
	mu    sync.Mutex
	paths []string
	calls map[string]int
	fail  func(c Candidate, n int) error
}

func newDownloader(t *testing.T) *downloader {
	return &downloader{t: t, dir: t.TempDir(), calls: map[string]int{}}
}

func (d *downloader) Download(ctx context.Context, c Candidate) (string, string, error) {
	d.mu.Lock()
	d.calls[c.ExternalID]++
	n := d.calls[c.ExternalID]
	d.mu.Unlock()
	if d.fail != nil {
		if err := d.fail(c, n); err != nil {
			return "", "", err
		}
	}
	body := []byte("media-" + c.ExternalID)
	p := filepath.Join(d.dir, c.ExternalID+"-"+time.Now().Format("150405.000000000"))
	if err := os.WriteFile(p, body, 0600); err != nil {
		return "", "", err
	}
	d.mu.Lock()
	d.paths = append(d.paths, p)
	d.mu.Unlock()
	return p, shaHex(body), nil
}

func (d *downloader) assertCleaned() {
	d.t.Helper()
	for _, p := range d.paths {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			d.t.Errorf("downloaded temp file left behind")
		}
	}
}

func newTestEngine(t *testing.T, s *Store, clock *fakeClock, d *downloader, conns ...Connector) *Engine {
	t.Helper()
	e := &Engine{
		Store:      s,
		Connectors: conns,
		Policy:     Policy{Civitai: CivitaiPolicy{AllowedLicenses: []string{"cc-by"}}},
		Now:        clock.Now,
		Sleep:      clock.Sleep,
	}
	if d != nil {
		e.Download = d.Download
	}
	return e
}

func externalIDs(t *testing.T, s *Store, provider string) []string {
	t.Helper()
	ids, err := s.ExternalIDs(context.Background(), provider)
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(ids)
	return ids
}

func deadLetters(t *testing.T, s *Store, provider string) []DeadLetter {
	t.Helper()
	dl, err := s.DeadLetters(context.Background(), provider)
	if err != nil {
		t.Fatal(err)
	}
	return dl
}

func TestRetryAfterIsHonoured(t *testing.T) {
	t.Run("short wait retries the same cursor", func(t *testing.T) {
		s, _ := newTestStore(t)
		clock := newFakeClock()
		start := clock.Now()
		conn := &fakeConnector{name: "civitai", scopes: []string{"tag:a"}, clock: clock, fetch: func(scope, cursor string, n int) (Page, error) {
			if n == 1 {
				return Page{}, ErrRateLimited{RetryAfter: 120 * time.Second}
			}
			return Page{Items: []Candidate{meta("civitai", "1")}}, nil
		}}
		rep := newTestEngine(t, s, clock, nil, conn).Sync(context.Background())
		calls := conn.Calls()
		if len(calls) != 2 {
			t.Fatalf("want 2 fetches, got %d", len(calls))
		}
		if calls[1].At.Before(start.Add(120 * time.Second)) {
			t.Fatalf("retried %v after the limit, before 120 s", calls[1].At.Sub(start))
		}
		if calls[1].Cursor != calls[0].Cursor {
			t.Fatal("retry must reuse the same cursor")
		}
		if got := clock.Sleeps(); !reflect.DeepEqual(got, []time.Duration{120 * time.Second}) {
			t.Fatalf("sleeps = %v", got)
		}
		if pr := rep.PerProvider["civitai"]; pr.State != "ok" || pr.Stored != 1 {
			t.Fatalf("report = %+v", pr)
		}
	})

	t.Run("long wait persists and skips the provider", func(t *testing.T) {
		s, _ := newTestStore(t)
		clock := newFakeClock()
		limited := true
		conn := &fakeConnector{name: "civitai", scopes: []string{"tag:a", "tag:b"}, clock: clock, fetch: func(scope, cursor string, n int) (Page, error) {
			if limited {
				return Page{}, ErrRateLimited{RetryAfter: 10 * time.Minute}
			}
			return Page{Items: []Candidate{meta("civitai", scope)}}, nil
		}}
		other := &fakeConnector{name: "reddit", scopes: []string{"r/x"}, clock: clock, fetch: func(scope, cursor string, n int) (Page, error) {
			return Page{Items: []Candidate{meta("reddit", "r1")}}, nil
		}}
		e := newTestEngine(t, s, clock, nil, conn, other)
		rep := e.Sync(context.Background())
		if pr := rep.PerProvider["civitai"]; pr.State != "rate_limited" {
			t.Fatalf("report = %+v", pr)
		}
		if rep.PerProvider["reddit"].State != "ok" {
			t.Fatal("another provider must keep syncing")
		}
		if len(conn.Calls()) != 1 || len(clock.Sleeps()) != 0 {
			t.Fatalf("calls=%d sleeps=%v: a long limit must stop the provider without waiting", len(conn.Calls()), clock.Sleeps())
		}
		limited = false
		clock.Advance(9*time.Minute + 59*time.Second)
		// A fresh engine over the same store: the limit is persisted, not in memory.
		rep = newTestEngine(t, s, clock, nil, conn).Sync(context.Background())
		if len(conn.Calls()) != 1 || rep.PerProvider["civitai"].State != "rate_limited" {
			t.Fatalf("called before not_before: calls=%d report=%+v", len(conn.Calls()), rep.PerProvider["civitai"])
		}
		clock.Advance(time.Second)
		rep = e.Sync(context.Background())
		if pr := rep.PerProvider["civitai"]; pr.State != "ok" || pr.Stored != 2 {
			t.Fatalf("report after the limit = %+v", pr)
		}
	})
}

func TestResumesFromSavedCursor(t *testing.T) {
	s, _ := newTestStore(t)
	clock := newFakeClock()
	ctx := context.Background()
	revoked := true
	conn := &fakeConnector{name: "civitai", scopes: []string{"tag:a"}, clock: clock, fetch: func(scope, cursor string, n int) (Page, error) {
		switch cursor {
		case "":
			return Page{Items: []Candidate{meta("civitai", "1")}, NextCursor: "c1"}, nil
		case "c1":
			return Page{Items: []Candidate{meta("civitai", "2")}, NextCursor: "c2"}, nil
		case "c2":
			if revoked {
				return Page{}, ErrUnauthorized
			}
			return Page{Items: []Candidate{meta("civitai", "3")}, NextCursor: "c2"}, nil
		}
		t.Fatalf("unexpected cursor %q", cursor)
		return Page{}, nil
	}}
	rep := newTestEngine(t, s, clock, nil, conn).Sync(ctx)
	if pr := rep.PerProvider["civitai"]; pr.State != "disconnected" || pr.Stored != 2 {
		t.Fatalf("first sync = %+v", pr)
	}
	if c, err := s.Cursor(ctx, "civitai", "tag:a"); err != nil || c != "c2" {
		t.Fatalf("saved cursor = %q %v", c, err)
	}
	revoked = false
	before := len(conn.Calls())
	rep = newTestEngine(t, s, clock, nil, conn).Sync(ctx)
	calls := conn.Calls()[before:]
	if len(calls) != 1 || calls[0].Cursor != "c2" {
		t.Fatalf("second sync must start from the saved cursor and stop on an unchanged one: %+v", calls)
	}
	if pr := rep.PerProvider["civitai"]; pr.State != "ok" || pr.Stored != 1 {
		t.Fatalf("second sync = %+v", pr)
	}
	if got := externalIDs(t, s, "civitai"); !reflect.DeepEqual(got, []string{"1", "2", "3"}) {
		t.Fatalf("ids = %v", got)
	}
}

func TestUnauthorizedDisconnectsOnlyThatProvider(t *testing.T) {
	s, _ := newTestStore(t)
	clock := newFakeClock()
	d := newDownloader(t)
	bad := &fakeConnector{name: "civitai", scopes: []string{"tag:a", "tag:b"}, clock: clock, fetch: func(scope, cursor string, n int) (Page, error) {
		return Page{}, ErrUnauthorized
	}}
	good := &fakeConnector{name: "reddit", scopes: []string{"r/x"}, clock: clock, fetch: func(scope, cursor string, n int) (Page, error) {
		return Page{Items: []Candidate{meta("reddit", "a"), meta("reddit", "b")}}, nil
	}}
	rep := newTestEngine(t, s, clock, d, bad, good).Sync(context.Background())
	if pr := rep.PerProvider["civitai"]; pr.State != "disconnected" {
		t.Fatalf("civitai = %+v", pr)
	}
	if len(bad.Calls()) != 1 || len(clock.Sleeps()) != 0 {
		t.Fatalf("unauthorized must not be retried nor reach other scopes: calls=%d sleeps=%v", len(bad.Calls()), clock.Sleeps())
	}
	if pr := rep.PerProvider["reddit"]; pr.State != "ok" || pr.Stored != 2 {
		t.Fatalf("reddit = %+v", pr)
	}
	if len(deadLetters(t, s, "civitai")) != 0 {
		t.Fatal("a disconnect is not a dead letter")
	}
}

func TestTransientErrorsBackOffThenDeadLetter(t *testing.T) {
	backoff := []time.Duration{time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second}

	t.Run("fetch", func(t *testing.T) {
		s, _ := newTestStore(t)
		clock := newFakeClock()
		ctx := context.Background()
		if err := s.SaveCursor(ctx, "civitai", "tag:a", "c7"); err != nil {
			t.Fatal(err)
		}
		conn := &fakeConnector{name: "civitai", scopes: []string{"tag:a", "tag:b"}, clock: clock, fetch: func(scope, cursor string, n int) (Page, error) {
			if scope == "tag:a" {
				if n%2 == 0 {
					return Page{}, errors.New("unclassified https://secret.example.test")
				}
				return Page{}, ErrTransient
			}
			return Page{Items: []Candidate{meta("civitai", "b1")}}, nil
		}}
		rep := newTestEngine(t, s, clock, nil, conn).Sync(ctx)
		var onA int
		for _, c := range conn.Calls() {
			if c.Scope == "tag:a" {
				onA++
				if c.Cursor != "c7" {
					t.Fatalf("retry used cursor %q", c.Cursor)
				}
			}
		}
		if onA != 5 {
			t.Fatalf("want 5 attempts, got %d", onA)
		}
		if got := clock.Sleeps(); !reflect.DeepEqual(got, backoff) {
			t.Fatalf("sleeps = %v", got)
		}
		if pr := rep.PerProvider["civitai"]; pr.State != "failed" || pr.Stored != 1 {
			t.Fatalf("report = %+v (the next scope must still sync)", pr)
		}
		if c, _ := s.Cursor(ctx, "civitai", "tag:a"); c != "c7" {
			t.Fatalf("cursor moved to %q", c)
		}
		dl := deadLetters(t, s, "civitai")
		if len(dl) != 1 || dl[0].Scope != "tag:a" || dl[0].Cursor != "c7" || dl[0].ExternalID != "" || dl[0].Code != "transient" || !dl[0].At.Equal(clock.Now()) {
			t.Fatalf("dead letters = %+v", dl)
		}
	})

	t.Run("download", func(t *testing.T) {
		s, _ := newTestStore(t)
		clock := newFakeClock()
		d := newDownloader(t)
		d.fail = func(c Candidate, n int) error {
			if c.ExternalID == "bad" {
				return ErrTransient
			}
			return nil
		}
		conn := &fakeConnector{name: "civitai", scopes: []string{"tag:a"}, clock: clock, fetch: func(scope, cursor string, n int) (Page, error) {
			return Page{Items: []Candidate{media("civitai", "bad"), media("civitai", "good")}}, nil
		}}
		rep := newTestEngine(t, s, clock, d, conn).Sync(context.Background())
		if d.calls["bad"] != 5 || d.calls["good"] != 1 {
			t.Fatalf("download attempts = %v", d.calls)
		}
		if got := clock.Sleeps(); !reflect.DeepEqual(got, backoff) {
			t.Fatalf("sleeps = %v", got)
		}
		if pr := rep.PerProvider["civitai"]; pr.Stored != 1 {
			t.Fatalf("report = %+v", pr)
		}
		dl := deadLetters(t, s, "civitai")
		if len(dl) != 1 || dl[0].ExternalID != "bad" || dl[0].Code != "transient" || dl[0].Scope != "" {
			t.Fatalf("dead letters = %+v", dl)
		}
		if got := externalIDs(t, s, "civitai"); !reflect.DeepEqual(got, []string{"good"}) {
			t.Fatalf("ids = %v", got)
		}
		d.assertCleaned()
	})

	t.Run("cancel stops promptly", func(t *testing.T) {
		s, _ := newTestStore(t)
		clock := newFakeClock()
		ctx, cancel := context.WithCancel(context.Background())
		conn := &fakeConnector{name: "civitai", scopes: []string{"tag:a", "tag:b"}, clock: clock, fetch: func(scope, cursor string, n int) (Page, error) {
			cancel()
			return Page{}, ErrTransient
		}}
		newTestEngine(t, s, clock, nil, conn).Sync(ctx)
		if len(conn.Calls()) != 1 {
			t.Fatalf("cancelled sync kept calling: %d", len(conn.Calls()))
		}
	})
}

func TestRemovedItemsAreDeleted(t *testing.T) {
	s, _ := newTestStore(t)
	clock := newFakeClock()
	ctx := context.Background()
	mustPut(t, s, Item{Provider: "civitai", ExternalID: "x"}, []byte("media-x"))
	mustPut(t, s, Item{Provider: "civitai", ExternalID: "y"}, []byte("media-y"))
	mustPut(t, s, Item{Provider: "civitai", ExternalID: "z"}, []byte("media-z"))
	conn := &removalConnector{
		fakeConnector: &fakeConnector{name: "civitai", scopes: []string{"tag:a"}, clock: clock, fetch: func(scope, cursor string, n int) (Page, error) {
			return Page{Removed: []string{"y", "never-stored"}}, nil
		}},
		removed: func(ids []string) []string { return []string{"x"} },
	}
	rep := newTestEngine(t, s, clock, nil, conn).Sync(ctx)
	if pr := rep.PerProvider["civitai"]; pr.State != "ok" || pr.Removed != 2 {
		t.Fatalf("report = %+v", pr)
	}
	if len(conn.asked) != 1 {
		t.Fatalf("removal check called %d times", len(conn.asked))
	}
	sort.Strings(conn.asked[0])
	if !reflect.DeepEqual(conn.asked[0], []string{"x", "y", "z"}) {
		t.Fatalf("removal check got %v", conn.asked[0])
	}
	if got := externalIDs(t, s, "civitai"); !reflect.DeepEqual(got, []string{"z"}) {
		t.Fatalf("ids = %v", got)
	}
	for _, gone := range []string{"media-x", "media-y"} {
		if _, err := s.Open(ctx, shaHex([]byte(gone))); !errors.Is(err, ErrNotFound) {
			t.Fatalf("blob of a removed item still opens: %v", err)
		}
	}
}

func TestRejectedCandidatesAreNeverDownloaded(t *testing.T) {
	s, _ := newTestStore(t)
	clock := newFakeClock()
	d := newDownloader(t)
	minor := media("civitai", "minor")
	minor.MinorFlag = true
	unlicensed := media("civitai", "unlicensed")
	unlicensed.License = ""
	otherLicence := media("civitai", "other")
	otherLicence.License = "all-rights-reserved"
	redditMedia := media("reddit", "rm")
	conns := []Connector{
		&fakeConnector{name: "civitai", scopes: []string{"tag:a"}, clock: clock, fetch: func(scope, cursor string, n int) (Page, error) {
			return Page{Items: []Candidate{minor, unlicensed, otherLicence, media("civitai", "ok")}}, nil
		}},
		&fakeConnector{name: "reddit", scopes: []string{"r/x"}, clock: clock, fetch: func(scope, cursor string, n int) (Page, error) {
			return Page{Items: []Candidate{redditMedia, meta("reddit", "m")}}, nil
		}},
	}
	rep := newTestEngine(t, s, clock, d, conns...).Sync(context.Background())
	if pr := rep.PerProvider["civitai"]; pr.Rejected != 3 || pr.Stored != 1 {
		t.Fatalf("civitai = %+v", pr)
	}
	if pr := rep.PerProvider["reddit"]; pr.Rejected != 1 || pr.Stored != 1 {
		t.Fatalf("reddit = %+v", pr)
	}
	if !reflect.DeepEqual(d.calls, map[string]int{"ok": 1}) {
		t.Fatalf("downloaded = %v", d.calls)
	}
	d.assertCleaned()

	// Without a downloader, media candidates fail closed; zero scopes make no calls.
	s2, _ := newTestStore(t)
	none := &fakeConnector{name: "reddit", clock: clock, fetch: func(scope, cursor string, n int) (Page, error) {
		t.Fatal("fetched with no scopes")
		return Page{}, nil
	}}
	nodl := &fakeConnector{name: "civitai", scopes: []string{"tag:a"}, clock: clock, fetch: func(scope, cursor string, n int) (Page, error) {
		return Page{Items: []Candidate{media("civitai", "1")}}, nil
	}}
	rep = newTestEngine(t, s2, clock, nil, none, nodl).Sync(context.Background())
	if pr := rep.PerProvider["reddit"]; pr != (ProviderReport{State: "ok"}) {
		t.Fatalf("no scopes = %+v", pr)
	}
	if pr := rep.PerProvider["civitai"]; pr.Rejected != 1 || pr.Stored != 0 {
		t.Fatalf("no downloader = %+v", pr)
	}
}
