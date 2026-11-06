package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	testRedditUA      = "linux:local.lunaos.discovery:v0.1 (by /u/owner)"
	testRedditID      = "client-id"
	testRedditSecret  = "client-secret-value"
	testRedditRefresh = "refresh-token-value"
)

// redditAPI imitates the documented OAuth2 token endpoint (POST
// /api/v1/access_token, HTTP Basic client auth, grant_type=refresh_token) and
// the oauth.reddit.com listings GET /r/<sub>/new and GET /api/info.
type redditAPI struct {
	t       *testing.T
	mu      sync.Mutex
	tokens  int
	issued  []string
	revoked map[string]bool
	reqs    []*http.Request
	// token answers a token request; nil issues "tok-<n>" valid one hour.
	token func(n int) (int, http.Header, any)
	// listing answers an authenticated API request.
	listing func(r *http.Request) (int, http.Header, any)
}

func newRedditAPI(t *testing.T, listing func(r *http.Request) (int, http.Header, any)) (*redditAPI, *httptest.Server) {
	api := &redditAPI{t: t, listing: listing, revoked: map[string]bool{}}
	srv := httptest.NewServer(api)
	t.Cleanup(srv.Close)
	return api, srv
}

func (a *redditAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if got := r.Header.Get("User-Agent"); got != testRedditUA {
		a.t.Errorf("User-Agent = %q", got)
	}
	a.mu.Lock()
	if r.URL.Path == "/api/v1/access_token" {
		a.tokens++
		n := a.tokens
		a.mu.Unlock()
		if r.Method != http.MethodPost {
			a.t.Errorf("token method %s", r.Method)
		}
		if id, secret, ok := r.BasicAuth(); !ok || id != testRedditID || secret != testRedditSecret {
			a.t.Errorf("token request without the client's HTTP Basic auth")
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
			a.t.Errorf("token Content-Type = %q", ct)
		}
		if err := r.ParseForm(); err != nil || r.PostForm.Get("grant_type") != "refresh_token" || r.PostForm.Get("refresh_token") != testRedditRefresh {
			a.t.Errorf("token form = %v", r.PostForm)
		}
		status, header, body := http.StatusOK, http.Header(nil), any(nil)
		if a.token != nil {
			status, header, body = a.token(n)
		}
		if body == nil {
			tok := fmt.Sprintf("tok-%d", n)
			a.mu.Lock()
			a.issued = append(a.issued, tok)
			a.mu.Unlock()
			body = map[string]any{"access_token": tok, "token_type": "bearer", "expires_in": 3600, "scope": "read"}
		}
		writeJSON(w, status, header, body)
		return
	}
	a.reqs = append(a.reqs, r)
	auth := r.Header.Get("Authorization")
	tok, _ := strings.CutPrefix(auth, "bearer ")
	known := false
	for _, v := range a.issued {
		known = known || v == tok
	}
	revoked := a.revoked[tok]
	a.mu.Unlock()
	if r.Method != http.MethodGet {
		a.t.Errorf("API method %s", r.Method)
	}
	if !known || revoked {
		writeJSON(w, http.StatusUnauthorized, nil, map[string]any{"message": "Unauthorized", "error": 401})
		return
	}
	status, header, body := a.listing(r)
	writeJSON(w, status, header, body)
}

func writeJSON(w http.ResponseWriter, status int, header http.Header, body any) {
	for k, v := range header {
		w.Header()[k] = v
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func (a *redditAPI) Requests() []*http.Request {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]*http.Request(nil), a.reqs...)
}

func (a *redditAPI) Tokens() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.tokens
}

func (a *redditAPI) Revoke(tok string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.revoked[tok] = true
}

// post returns one documented t3 thing of a Listing.
func post(id, sub string, over18 bool, hint, link string) map[string]any {
	return map[string]any{"kind": "t3", "data": map[string]any{
		"id": id, "name": "t3_" + id, "subreddit": sub, "subreddit_name_prefixed": "r/" + sub,
		"title": "title " + id, "author": "author_" + id, "over_18": over18, "post_hint": hint,
		"url": link, "permalink": "/r/" + sub + "/comments/" + id + "/title_" + id + "/",
		"is_self": hint == "self", "removed_by_category": nil, "created_utc": 1.7e9,
	}}
}

func listing(after string, children ...map[string]any) map[string]any {
	if children == nil {
		children = []map[string]any{}
	}
	var a any
	if after != "" {
		a = after
	}
	return map[string]any{"kind": "Listing", "data": map[string]any{"after": a, "before": nil, "dist": len(children), "children": children}}
}

func redditConfig(srv *httptest.Server) RedditConfig {
	return RedditConfig{
		ClientID: testRedditID, ClientSecret: testRedditSecret, RefreshToken: testRedditRefresh, UserAgent: testRedditUA,
		TokenURL: srv.URL + "/api/v1/access_token", APIBase: srv.URL,
	}
}

func newTestReddit(t *testing.T, srv *httptest.Server, p RedditPolicy, clock *fakeClock) Connector {
	t.Helper()
	conn := NewReddit(redditConfig(srv), p, srv.Client())
	conn.(*reddit).now = clock.Now
	return conn
}

func redditEngine(t *testing.T, s *Store, clock *fakeClock, d *downloader, conn Connector, p RedditPolicy) *Engine {
	e := newTestEngine(t, s, clock, d, conn)
	e.Policy = Policy{Reddit: p}
	return e
}

func assertNoLeak(t *testing.T, err error, needles ...string) {
	t.Helper()
	if err == nil {
		return
	}
	for _, n := range append(needles, "127.0.0.1", testRedditSecret, testRedditRefresh, "tok-") {
		if strings.Contains(err.Error(), n) {
			t.Errorf("error %q leaks %q", err, n)
		}
	}
}

func TestRedditCandidatesAreMetadataOnly(t *testing.T) {
	var mediaHits atomic.Int32
	mediaSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mediaHits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(mediaSrv.Close)
	removed := post("gone1", "Allowed_Sub", true, "image", mediaSrv.URL+"/gone1.jpg")
	removed["data"].(map[string]any)["removed_by_category"] = "moderator"
	deleted := post("gone2", "Allowed_Sub", true, "image", mediaSrv.URL+"/gone2.jpg")
	deleted["data"].(map[string]any)["author"] = "[deleted]"
	_, srv := newRedditAPI(t, func(r *http.Request) (int, http.Header, any) {
		return ok(listing("",
			post("img1", "Allowed_Sub", true, "image", mediaSrv.URL+"/img1.jpg"),
			post("vid1", "Allowed_Sub", false, "hosted:video", "https://v.redd.it/vid1"),
			post("ired", "Allowed_Sub", true, "", "https://i.redd.it/ired.png"),
			post("self1", "Allowed_Sub", true, "self", "https://www.reddit.com/r/Allowed_Sub/comments/self1/"),
			post("link1", "Allowed_Sub", true, "link", "https://example.test/article"),
			removed, deleted,
		))
	})
	p := RedditPolicy{Subreddits: []string{"Allowed_Sub"}}
	clock := newFakeClock()
	conn := newTestReddit(t, srv, p, clock)

	pg, err := conn.Fetch(context.Background(), "r/Allowed_Sub", "")
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, c := range pg.Items {
		ids = append(ids, c.ExternalID)
		if c.MediaURL != "" || c.DeclaredSHA256 != "" || c.License != "" || c.MinorFlag {
			t.Errorf("%s is not metadata-only: %+v", c.ExternalID, c)
		}
	}
	if !reflect.DeepEqual(ids, []string{"t3_img1", "t3_vid1", "t3_ired"}) {
		t.Fatalf("candidates = %v (only direct media posts, removed ones excluded)", ids)
	}
	want := Item{
		Provider: "reddit", ExternalID: "t3_img1", Title: "title img1", Creator: "author_img1",
		SourceURL: "https://www.reddit.com/r/Allowed_Sub/comments/img1/title_img1/", LinkURL: mediaSrv.URL + "/img1.jpg",
		Tags: []string{"r/Allowed_Sub"}, ContentLevel: "X",
	}
	if got := pg.Items[0].Item; !reflect.DeepEqual(got, want) {
		t.Fatalf("item = %+v\nwant %+v", got, want)
	}
	if pg.Items[1].ContentLevel != "None" {
		t.Fatalf("not over_18 → %q", pg.Items[1].ContentLevel)
	}
	if !reflect.DeepEqual(pg.Removed, []string{"t3_gone1", "t3_gone2"}) {
		t.Fatalf("removed = %v", pg.Removed)
	}

	// Through a real Engine and Store: rows are stored, no blob, no download.
	s, _ := newTestStore(t)
	d := newDownloader(t)
	rep := redditEngine(t, s, clock, d, conn, p).Sync(context.Background())
	if pr := rep.PerProvider["reddit"]; pr.State != "ok" || pr.Stored != 3 {
		t.Fatalf("report = %+v", pr)
	}
	if got := externalIDs(t, s, "reddit"); !reflect.DeepEqual(got, []string{"t3_img1", "t3_ired", "t3_vid1"}) {
		t.Fatalf("stored = %v", got)
	}
	if len(blobFiles(t, s)) != 0 || len(d.calls) != 0 {
		t.Fatalf("blobs=%d downloads=%d", len(blobFiles(t, s)), len(d.calls))
	}
	if n := mediaHits.Load(); n != 0 {
		t.Fatalf("%d requests reached the media host", n)
	}
	var sealed []byte
	if err := s.db.QueryRow(`SELECT meta FROM items WHERE external_id='t3_img1'`).Scan(&sealed); err != nil {
		t.Fatal(err)
	}
	m, err := s.openMeta(sealed, "reddit", "t3_img1")
	if err != nil || m.Title != want.Title || m.LinkURL != want.LinkURL {
		t.Fatalf("sealed meta = %+v, %v", m, err)
	}
}

func TestRefreshesTokenOnce(t *testing.T) {
	api, srv := newRedditAPI(t, func(r *http.Request) (int, http.Header, any) {
		return ok(listing("", post("a", "sub_one", true, "image", "https://i.redd.it/a.jpg")))
	})
	clock := newFakeClock()
	conn := newTestReddit(t, srv, RedditPolicy{Subreddits: []string{"sub_one"}}, clock)
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if _, err := conn.Fetch(ctx, "r/sub_one", ""); err != nil {
			t.Fatal(err)
		}
	}
	if n := api.Tokens(); n != 1 {
		t.Fatalf("token requests = %d, want the cached token reused", n)
	}
	for _, r := range api.Requests() {
		if got := r.Header.Get("Authorization"); got != "bearer tok-1" {
			t.Fatalf("Authorization = %q", got)
		}
	}
	// Close to expiry the token is refreshed before use.
	clock.Advance(59 * time.Minute)
	if _, err := conn.Fetch(ctx, "r/sub_one", ""); err != nil {
		t.Fatal(err)
	}
	if n := api.Tokens(); n != 2 {
		t.Fatalf("token requests after expiry = %d", n)
	}
	// A token revoked server-side: one refresh, one retry, success.
	api.Revoke("tok-2")
	before := len(api.Requests())
	if _, err := conn.Fetch(ctx, "r/sub_one", ""); err != nil {
		t.Fatalf("after revocation: %v", err)
	}
	if n := api.Tokens(); n != 3 {
		t.Fatalf("token requests after a 401 = %d, want exactly one refresh", n)
	}
	if n := len(api.Requests()) - before; n != 2 {
		t.Fatalf("listing requests after a 401 = %d, want one retry", n)
	}
}

func TestOnlyListedSubredditsRequested(t *testing.T) {
	api, srv := newRedditAPI(t, func(r *http.Request) (int, http.Header, any) {
		if r.URL.Query().Get("after") == "" && strings.HasPrefix(r.URL.Path, "/r/Pics_ok/") {
			return ok(listing("t3_p1", post("p1", "Pics_ok", true, "image", "https://i.redd.it/p1.jpg")))
		}
		return ok(listing(""))
	})
	clock := newFakeClock()
	p := RedditPolicy{Subreddits: []string{"Pics_ok", "bad name!", "a", "ab", "../etc", "pics_OK", "abcdefghijklmnopqrstuv", "x/new?y"}}
	conn := newTestReddit(t, srv, p, clock)
	if got := conn.Scopes(); !reflect.DeepEqual(got, []string{"r/Pics_ok", "r/ab"}) {
		t.Fatalf("scopes = %v", got)
	}
	ctx := context.Background()
	for _, scope := range []string{"", "r/", "r/a", "r/other", "Pics_ok", "r/../etc", "r/bad name!", "user:Pics_ok", "r/Pics_ok/../x"} {
		if _, err := conn.Fetch(ctx, scope, ""); err == nil {
			t.Errorf("scope %q fetched", scope)
		}
	}
	if n := len(api.Requests()) + api.Tokens(); n != 0 {
		t.Fatalf("%d requests sent for refused scopes", n)
	}
	s, _ := newTestStore(t)
	rep := redditEngine(t, s, clock, nil, conn, p).Sync(ctx)
	if pr := rep.PerProvider["reddit"]; pr.State != "ok" || pr.Stored != 1 {
		t.Fatalf("report = %+v", pr)
	}
	var got []string
	for _, r := range api.Requests() {
		q := r.URL.Query()
		if q.Get("limit") != "100" || q.Get("raw_json") != "1" {
			t.Errorf("query %v", q)
		}
		got = append(got, r.URL.Path+"?after="+q.Get("after"))
	}
	if want := []string{"/r/Pics_ok/new?after=", "/r/Pics_ok/new?after=t3_p1", "/r/ab/new?after="}; !reflect.DeepEqual(got, want) {
		t.Fatalf("requests = %v", got)
	}
	// No listed subreddit: no scope, no request at all.
	empty := newTestReddit(t, srv, RedditPolicy{}, clock)
	before := len(api.Requests()) + api.Tokens()
	if len(empty.Scopes()) != 0 {
		t.Fatal("scopes without a list")
	}
	if pr := redditEngine(t, s, clock, nil, empty, RedditPolicy{}).Sync(ctx).PerProvider["reddit"]; pr != (ProviderReport{State: "ok"}) {
		t.Fatalf("report = %+v", pr)
	}
	if len(api.Requests())+api.Tokens() != before {
		t.Fatal("requests sent without a subreddit list")
	}
}

func TestRateLimitHeadersBecomeRetryAfter(t *testing.T) {
	var mu sync.Mutex
	var status int
	var header http.Header
	api, srv := newRedditAPI(t, func(r *http.Request) (int, http.Header, any) {
		mu.Lock()
		defer mu.Unlock()
		if status != http.StatusOK {
			return status, header, map[string]any{"message": "Too Many Requests", "error": 429}
		}
		return http.StatusOK, header, listing("", post("a", "sub_one", true, "image", "https://i.redd.it/a.jpg"))
	})
	clock := newFakeClock()
	conn := newTestReddit(t, srv, RedditPolicy{Subreddits: []string{"sub_one"}}, clock)
	ctx := context.Background()
	set := func(s int, h http.Header) { mu.Lock(); status, header = s, h; mu.Unlock() }
	for _, c := range []struct {
		header http.Header
		want   time.Duration
	}{
		{http.Header{"Retry-After": {"7"}, "X-Ratelimit-Reset": {"300"}}, 7 * time.Second},
		{http.Header{"X-Ratelimit-Remaining": {"0.0"}, "X-Ratelimit-Reset": {"42"}}, 42 * time.Second},
		{nil, 60 * time.Second},
	} {
		set(http.StatusTooManyRequests, c.header)
		_, err := conn.Fetch(ctx, "r/sub_one", "")
		if rl, ok := asRateLimited(err); !ok || rl.RetryAfter != c.want {
			t.Errorf("429 %v: err = %v (%+v)", c.header, err, rl)
		}
		assertNoLeak(t, err, "sub_one")
		clock.Advance(10 * time.Minute)
	}
	// An exhausted budget on a successful response: the page is used, and
	// the next call waits for X-Ratelimit-Reset without sending a request.
	set(http.StatusOK, http.Header{"X-Ratelimit-Used": {"600"}, "X-Ratelimit-Remaining": {"0.0"}, "X-Ratelimit-Reset": {"120"}})
	pg, err := conn.Fetch(ctx, "r/sub_one", "")
	if err != nil || len(pg.Items) != 1 {
		t.Fatalf("page with an exhausted budget: %v %+v", err, pg)
	}
	sent := len(api.Requests())
	clock.Advance(20 * time.Second)
	if _, err := conn.Fetch(ctx, "r/sub_one", ""); err == nil {
		t.Fatal("fetch sent with an exhausted budget")
	} else if rl, ok := asRateLimited(err); !ok || rl.RetryAfter != 100*time.Second {
		t.Fatalf("err = %v (%+v), want the rest of the reset window", err, rl)
	}
	if _, err := conn.(RemovalChecker).Removed(ctx, []string{"t3_a"}); err == nil {
		t.Fatal("removal check sent with an exhausted budget")
	}
	if len(api.Requests()) != sent {
		t.Fatal("a request was sent while the budget was exhausted")
	}
	set(http.StatusOK, http.Header{"X-Ratelimit-Remaining": {"599.0"}, "X-Ratelimit-Reset": {"600"}})
	clock.Advance(100 * time.Second)
	for i := 0; i < 2; i++ {
		if _, err := conn.Fetch(ctx, "r/sub_one", ""); err != nil {
			t.Fatalf("after the reset: %v", err)
		}
	}
	// Through the engine: a short wait is slept and the same cursor retried.
	set(http.StatusOK, http.Header{"X-Ratelimit-Remaining": {"0"}, "X-Ratelimit-Reset": {"30"}})
	s, _ := newTestStore(t)
	if _, err := conn.Fetch(ctx, "r/sub_one", ""); err != nil {
		t.Fatal(err)
	}
	set(http.StatusOK, http.Header{"X-Ratelimit-Remaining": {"10"}, "X-Ratelimit-Reset": {"30"}})
	rep := redditEngine(t, s, clock, nil, conn, RedditPolicy{Subreddits: []string{"sub_one"}}).Sync(ctx)
	if pr := rep.PerProvider["reddit"]; pr.State != "ok" || pr.Stored != 1 {
		t.Fatalf("report = %+v", pr)
	}
	if got := clock.Sleeps(); !reflect.DeepEqual(got, []time.Duration{30 * time.Second}) {
		t.Fatalf("sleeps = %v", got)
	}
}

func TestUnauthorizedAfterRefreshDisconnects(t *testing.T) {
	t.Run("listing still 401 after one refresh", func(t *testing.T) {
		api, srv := newRedditAPI(t, func(r *http.Request) (int, http.Header, any) {
			return http.StatusUnauthorized, nil, map[string]any{"message": "Unauthorized", "error": 401}
		})
		clock := newFakeClock()
		p := RedditPolicy{Subreddits: []string{"sub_one", "sub_two"}}
		conn := newTestReddit(t, srv, p, clock)
		_, err := conn.Fetch(context.Background(), "r/sub_one", "")
		if !errors.Is(err, ErrUnauthorized) {
			t.Fatalf("err = %v", err)
		}
		assertNoLeak(t, err, "sub_one")
		if api.Tokens() != 2 || len(api.Requests()) != 2 {
			t.Fatalf("tokens=%d listings=%d, want one refresh and one retry", api.Tokens(), len(api.Requests()))
		}
		s, _ := newTestStore(t)
		rep := redditEngine(t, s, clock, nil, conn, p).Sync(context.Background())
		if pr := rep.PerProvider["reddit"]; pr.State != "disconnected" {
			t.Fatalf("report = %+v", pr)
		}
		if api.Tokens() != 4 || len(api.Requests()) != 4 || len(clock.Sleeps()) != 0 {
			t.Fatalf("tokens=%d listings=%d sleeps=%v: a disconnect stops the provider unretried", api.Tokens(), len(api.Requests()), clock.Sleeps())
		}
	})
	t.Run("refresh token refused", func(t *testing.T) {
		for _, c := range []struct {
			status int
			body   any
		}{
			{http.StatusBadRequest, map[string]any{"error": "invalid_grant"}},
			{http.StatusUnauthorized, map[string]any{"message": "Unauthorized", "error": 401}},
			{http.StatusOK, map[string]any{"error": "invalid_grant"}},
		} {
			api, srv := newRedditAPI(t, func(r *http.Request) (int, http.Header, any) {
				t.Error("listing requested without a token")
				return ok(listing(""))
			})
			api.token = func(n int) (int, http.Header, any) { return c.status, nil, c.body }
			clock := newFakeClock()
			p := RedditPolicy{Subreddits: []string{"sub_one"}}
			conn := newTestReddit(t, srv, p, clock)
			_, err := conn.Fetch(context.Background(), "r/sub_one", "")
			if !errors.Is(err, ErrUnauthorized) {
				t.Errorf("token %d: err = %v", c.status, err)
			}
			assertNoLeak(t, err, "invalid_grant")
			s, _ := newTestStore(t)
			if pr := redditEngine(t, s, clock, nil, conn, p).Sync(context.Background()).PerProvider["reddit"]; pr.State != "disconnected" {
				t.Errorf("token %d: report = %+v", c.status, pr)
			}
		}
	})
}

func TestUserAgentRequired(t *testing.T) {
	api, srv := newRedditAPI(t, func(r *http.Request) (int, http.Header, any) {
		t.Error("request sent with an incomplete configuration")
		return ok(listing(""))
	})
	p := RedditPolicy{Subreddits: []string{"sub_one"}}
	for _, mutate := range []func(*RedditConfig){
		func(c *RedditConfig) { c.UserAgent = "" },
		func(c *RedditConfig) { c.UserAgent = "  " },
		func(c *RedditConfig) { c.ClientID = "" },
		func(c *RedditConfig) { c.RefreshToken = "" },
	} {
		cfg := redditConfig(srv)
		mutate(&cfg)
		conn := NewReddit(cfg, p, srv.Client())
		if _, err := conn.Fetch(context.Background(), "r/sub_one", ""); !errors.Is(err, ErrConfig) {
			t.Errorf("Fetch err = %v", err)
		}
		if _, err := conn.(RemovalChecker).Removed(context.Background(), []string{"t3_a"}); !errors.Is(err, ErrConfig) {
			t.Errorf("Removed err = %v", err)
		}
		s, _ := newTestStore(t)
		if _, err := s.Put(context.Background(), Item{Provider: "reddit", ExternalID: "t3_a"}, nil); err != nil {
			t.Fatal(err)
		}
		clock := newFakeClock()
		rep := redditEngine(t, s, clock, nil, conn, p).Sync(context.Background())
		if pr := rep.PerProvider["reddit"]; pr.State != "failed" {
			t.Errorf("report = %+v", pr)
		}
		if dl := deadLetters(t, s, "reddit"); len(dl) != 1 || dl[0].Code != "config" {
			t.Errorf("dead letters = %+v, want one config letter", dl)
		}
		if len(clock.Sleeps()) != 0 {
			t.Errorf("a configuration error is retried: %v", clock.Sleeps())
		}
	}
	if api.Tokens() != 0 || len(api.Requests()) != 0 {
		t.Fatalf("tokens=%d requests=%d", api.Tokens(), len(api.Requests()))
	}
	if errorCode(ErrConfig) != "config" || !isFinal(ErrConfig) {
		t.Fatal("ErrConfig must be final with code config")
	}
}

func TestRedditStatusMapping(t *testing.T) {
	var mu sync.Mutex
	status := map[string]int{}
	_, srv := newRedditAPI(t, func(r *http.Request) (int, http.Header, any) {
		mu.Lock()
		defer mu.Unlock()
		sub := strings.Split(r.URL.Path, "/")[2]
		if s := status[sub]; s != 0 {
			return s, nil, map[string]any{"reason": "private secret_body", "message": "Forbidden", "error": s}
		}
		return ok(listing("", post("ok1", sub, true, "image", "https://i.redd.it/ok1.jpg")))
	})
	clock := newFakeClock()
	p := RedditPolicy{Subreddits: []string{"sub_private", "sub_ok"}}
	conn := newTestReddit(t, srv, p, clock)
	for _, c := range []struct {
		status int
		check  func(error) bool
	}{
		{403, func(err error) bool { return errors.Is(err, ErrScopeUnavailable) && isFinal(err) }},
		{404, func(err error) bool { return errors.Is(err, ErrScopeUnavailable) && isFinal(err) }},
		{500, func(err error) bool { return errors.Is(err, ErrTransient) }},
		{503, func(err error) bool { return errors.Is(err, ErrTransient) }},
		{418, func(err error) bool { return err != nil && !isFinal(err) && !errors.Is(err, ErrTransient) }},
	} {
		mu.Lock()
		status["sub_private"] = c.status
		mu.Unlock()
		_, err := conn.Fetch(context.Background(), "r/sub_private", "")
		if !c.check(err) {
			t.Errorf("status %d: err = %v", c.status, err)
		}
		assertNoLeak(t, err, "sub_private", "secret_body")
	}
	// A forbidden subreddit is dead-lettered; the provider stays connected
	// and the next subreddit syncs.
	mu.Lock()
	status["sub_private"] = 403
	mu.Unlock()
	s, _ := newTestStore(t)
	rep := redditEngine(t, s, clock, nil, conn, p).Sync(context.Background())
	if pr := rep.PerProvider["reddit"]; pr.State != "failed" || pr.Stored != 1 {
		t.Fatalf("report = %+v", pr)
	}
	if dl := deadLetters(t, s, "reddit"); len(dl) != 1 || dl[0].Code != "scope_unavailable" || dl[0].Scope != "r/sub_private" {
		t.Fatalf("dead letters = %+v", dl)
	}
	if len(clock.Sleeps()) != 0 {
		t.Fatalf("a final scope error is retried: %v", clock.Sleeps())
	}
	srv.Close()
	_, err := conn.Fetch(context.Background(), "r/sub_ok", "")
	if !errors.Is(err, ErrTransient) {
		t.Fatalf("network error = %v", err)
	}
	assertNoLeak(t, err)
}

func TestRedditTakedownViaInfo(t *testing.T) {
	var mu sync.Mutex
	gone := map[string]string{} // name → "removed" | "deleted" | "absent"
	api, srv := newRedditAPI(t, func(r *http.Request) (int, http.Header, any) {
		mu.Lock()
		defer mu.Unlock()
		if r.URL.Path != "/api/info" {
			return ok(listing(""))
		}
		var children []map[string]any
		for _, name := range strings.Split(r.URL.Query().Get("id"), ",") {
			id := strings.TrimPrefix(name, "t3_")
			p := post(id, "sub_one", true, "image", "https://i.redd.it/"+id+".jpg")
			switch gone[name] {
			case "absent":
				continue
			case "removed":
				p["data"].(map[string]any)["removed_by_category"] = "moderator"
			case "deleted":
				p["data"].(map[string]any)["author"] = "[deleted]"
				p["data"].(map[string]any)["removed_by_category"] = "deleted"
			}
			children = append(children, p)
		}
		return ok(listing("", children...))
	})
	clock := newFakeClock()
	conn := newTestReddit(t, srv, RedditPolicy{Subreddits: []string{"sub_one"}}, clock)
	rc := conn.(RemovalChecker)
	ids := make([]string, 0, 152)
	for i := 0; i < 150; i++ {
		ids = append(ids, fmt.Sprintf("t3_p%d", i))
	}
	ids = append(ids, "not-a-name", "t3_BAD,ID")
	mu.Lock()
	gone["t3_p3"], gone["t3_p120"], gone["t3_p149"] = "absent", "removed", "deleted"
	mu.Unlock()
	got, err := rc.Removed(context.Background(), ids)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"t3_p3", "t3_p120", "t3_p149"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("removed = %v", got)
	}
	var sizes []int
	for _, r := range api.Requests() {
		if r.URL.Path != "/api/info" || r.URL.Query().Get("raw_json") != "1" {
			t.Errorf("request %s %v", r.URL.Path, r.URL.Query())
		}
		sizes = append(sizes, len(strings.Split(r.URL.Query().Get("id"), ",")))
	}
	if !reflect.DeepEqual(sizes, []int{100, 50}) {
		t.Fatalf("batch sizes = %v", sizes)
	}
	// Through the engine: the removed stored item is deleted.
	s, _ := newTestStore(t)
	for _, id := range []string{"t3_p1", "t3_p2", "t3_p3"} {
		if _, err := s.Put(context.Background(), Item{Provider: "reddit", ExternalID: id}, nil); err != nil {
			t.Fatal(err)
		}
	}
	rep := redditEngine(t, s, clock, nil, conn, RedditPolicy{Subreddits: []string{"sub_one"}}).Sync(context.Background())
	if pr := rep.PerProvider["reddit"]; pr.State != "ok" || pr.Removed != 1 {
		t.Fatalf("report = %+v", pr)
	}
	if got := externalIDs(t, s, "reddit"); !reflect.DeepEqual(got, []string{"t3_p1", "t3_p2"}) {
		t.Fatalf("ids = %v", got)
	}
}
