package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

const testAPIKey = "test-key"

// civitaiAPI imitates GET /api/v1/images as documented in
// civitai-developer-docs/site/reference/images.md.
type civitaiAPI struct {
	t       *testing.T
	mu      sync.Mutex
	queries []url.Values
	respond func(q url.Values) (status int, header http.Header, body any)
}

func newCivitaiAPI(t *testing.T, respond func(q url.Values) (int, http.Header, any)) (*civitaiAPI, *httptest.Server) {
	api := &civitaiAPI{t: t, respond: respond}
	srv := httptest.NewServer(api)
	t.Cleanup(srv.Close)
	return api, srv
}

func (a *civitaiAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || r.URL.Path != "/api/v1/images" {
		a.t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if got := r.Header.Get("Authorization"); got != "Bearer "+testAPIKey {
		a.t.Errorf("Authorization = %q", got)
	}
	q := r.URL.Query()
	a.mu.Lock()
	a.queries = append(a.queries, q)
	a.mu.Unlock()
	status, header, body := a.respond(q)
	for k, v := range header {
		w.Header()[k] = v
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func (a *civitaiAPI) Queries() []url.Values {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]url.Values(nil), a.queries...)
}

// image returns one documented items[] entry.
func image(id int, level, user, baseModel string) map[string]any {
	return map[string]any{
		"id": id, "url": fmt.Sprintf("https://image.civitai.test/%d.jpeg", id),
		"hash": "UA8N5},:Ioni~C#laKxaoznNwvx]XmRkVstR", "width": 832, "height": 1216, "type": "image",
		"nsfw": level != "None", "nsfwLevel": level, "browsingLevel": 2, "createdAt": "2025-04-17T21:28:57.225Z",
		"postId": 1981754, "thumbnail": nil, "username": user, "baseModel": baseModel,
		"modelVersionIds": []int{9208}, "stats": map[string]int{"likeCount": 1},
	}
}

func page(next string, items ...map[string]any) map[string]any {
	md := map[string]any{}
	if next != "" {
		md["nextCursor"] = next
		md["nextPage"] = "https://civitai.com/api/v1/images?cursor=" + next
	}
	if items == nil {
		items = []map[string]any{}
	}
	return map[string]any{"items": items, "metadata": md}
}

func ok(body any) (int, http.Header, any) { return http.StatusOK, nil, body }

func civitaiEngine(t *testing.T, s *Store, d *downloader, conn Connector, p CivitaiPolicy) *Engine {
	e := newTestEngine(t, s, newFakeClock(), d, conn)
	e.Policy = Policy{Civitai: p}
	return e
}

func TestEmptyAllowlistMakesNoRequest(t *testing.T) {
	api, srv := newCivitaiAPI(t, func(q url.Values) (int, http.Header, any) {
		t.Error("no request may be sent without an allowlist")
		return ok(page(""))
	})
	// Model families only narrow tag and creator scopes; alone they query nothing.
	p := CivitaiPolicy{ModelFamilies: []string{"SDXL 1.0"}, MaxContentLevel: "X", AllowedLicenses: []string{civitaiLicense}}
	conn := NewCivitai(srv.URL, testAPIKey, p, srv.Client())
	if got := conn.Scopes(); len(got) != 0 {
		t.Fatalf("scopes = %v", got)
	}
	s, _ := newTestStore(t)
	mustPut(t, s, Item{Provider: "civitai", ExternalID: "1"}, []byte("media-1"))
	rep := civitaiEngine(t, s, newDownloader(t), conn, p).Sync(context.Background())
	if pr := rep.PerProvider["civitai"]; pr != (ProviderReport{State: "ok"}) {
		t.Fatalf("report = %+v", pr)
	}
	// A scope outside the allowlist is refused without a request.
	for _, scope := range []string{"", "tag:", "user:", "tag:4", "nsfw:X", "user:Ajuro", "model_family:SDXL 1.0"} {
		if _, err := conn.Fetch(context.Background(), scope, ""); err == nil {
			t.Errorf("scope %q fetched", scope)
		}
	}
	if n := len(api.Queries()); n != 0 {
		t.Fatalf("%d requests sent", n)
	}
}

func TestEachRequestCarriesAllowlistFilter(t *testing.T) {
	api, srv := newCivitaiAPI(t, func(q url.Values) (int, http.Header, any) {
		switch {
		case q.Get("tags") == "4":
			// Item 3 ignores the baseModels filter and must be rejected.
			return ok(page("", image(1, "Mature", "someone", "SDXL 1.0"), image(3, "Soft", "other", "Pony")))
		case q.Get("username") == "Ajuro":
			return ok(page("", image(2, "X", "Ajuro", "SDXL 1.0")))
		}
		return ok(page(""))
	})
	p := CivitaiPolicy{Tags: []string{"4"}, Creators: []string{"Ajuro"}, ModelFamilies: []string{"SDXL 1.0", "Flux.1 D"}, MaxContentLevel: "X", AllowedLicenses: []string{civitaiLicense}}
	conn := NewCivitai(srv.URL, testAPIKey, p, srv.Client())
	if got, want := conn.Scopes(), []string{"tag:4", "user:Ajuro"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("scopes = %v: model families are never a scope of their own", got)
	}
	s, _ := newTestStore(t)
	d := newDownloader(t)
	rep := civitaiEngine(t, s, d, conn, p).Sync(context.Background())
	if pr := rep.PerProvider["civitai"]; pr.State != "ok" || pr.Stored != 2 || pr.Rejected != 1 {
		t.Fatalf("report = %+v", pr)
	}
	if got := externalIDs(t, s, "civitai"); !reflect.DeepEqual(got, []string{"1", "2"}) {
		t.Fatalf("ids = %v", got)
	}
	qs := api.Queries()
	if len(qs) != 2 {
		t.Fatalf("%d requests, want one per allowlist entry", len(qs))
	}
	for _, q := range qs {
		if (q.Get("tags") == "") == (q.Get("username") == "") || q.Get("baseModels") != "SDXL 1.0,Flux.1 D" || q.Get("nsfw") != "X" || q.Has("cursor") {
			t.Errorf("query %v lacks exactly one tag or username filter, the base models or the level cap", q)
		}
	}
	// Candidates carry the documented fields and the provider-terms licence.
	cands := map[string]Candidate{}
	for _, scope := range conn.Scopes() {
		pg, err := conn.Fetch(context.Background(), scope, "")
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range pg.Items {
			cands[c.ExternalID] = c
		}
	}
	want := map[string]Candidate{
		"1": {Item: Item{Provider: "civitai", ExternalID: "1", Creator: "someone", License: civitaiLicense, ContentLevel: "Mature", Tags: []string{"4", baseModelTag + "SDXL 1.0"}}, MediaURL: "https://image.civitai.test/1.jpeg"},
		"2": {Item: Item{Provider: "civitai", ExternalID: "2", Creator: "Ajuro", License: civitaiLicense, ContentLevel: "X", Tags: []string{baseModelTag + "SDXL 1.0"}}, MediaURL: "https://image.civitai.test/2.jpeg"},
		"3": {Item: Item{Provider: "civitai", ExternalID: "3", Creator: "other", License: civitaiLicense, ContentLevel: "Soft", Tags: []string{"4", baseModelTag + "Pony"}}, MediaURL: "https://image.civitai.test/3.jpeg"},
	}
	if !reflect.DeepEqual(cands, want) {
		t.Fatalf("candidates = %+v", cands)
	}
}

func TestCursorPagination(t *testing.T) {
	api, srv := newCivitaiAPI(t, func(q url.Values) (int, http.Header, any) {
		switch q.Get("cursor") {
		case "":
			return ok(page("feed:1:10", image(10, "Soft", "Ajuro", "SDXL 1.0"), image(11, "X", "Ajuro", "SDXL 1.0")))
		case "feed:1:10":
			// An empty page with a nextCursor is not the end of the results.
			return ok(page("feed:2:11"))
		case "feed:2:11":
			return ok(page("", image(12, "None", "Ajuro", "SDXL 1.0")))
		}
		t.Errorf("unexpected cursor %q", q.Get("cursor"))
		return ok(page(""))
	})
	p := CivitaiPolicy{Creators: []string{"Ajuro"}, MaxContentLevel: "Mature", AllowedLicenses: []string{civitaiLicense}}
	conn := NewCivitai(srv.URL, testAPIKey, p, srv.Client())
	s, _ := newTestStore(t)
	d := newDownloader(t)
	rep := civitaiEngine(t, s, d, conn, p).Sync(context.Background())
	if pr := rep.PerProvider["civitai"]; pr.State != "ok" || pr.Stored != 2 || pr.Rejected != 1 {
		t.Fatalf("report = %+v", pr)
	}
	if !reflect.DeepEqual(d.calls, map[string]int{"10": 1, "12": 1}) {
		t.Fatalf("downloaded = %v: the X item above the Mature cap must never be downloaded", d.calls)
	}
	var cursors []string
	for _, q := range api.Queries() {
		cursors = append(cursors, q.Get("cursor"))
		if q.Get("username") != "Ajuro" || q.Get("nsfw") != "Mature" || q.Has("baseModels") {
			t.Errorf("query %v", q)
		}
	}
	if !reflect.DeepEqual(cursors, []string{"", "feed:1:10", "feed:2:11"}) {
		t.Fatalf("cursors = %q", cursors)
	}
}

func TestCivitaiStatusMapping(t *testing.T) {
	var status int
	var header http.Header
	var body any = map[string]string{"error": "secret body Ajuro"}
	_, srv := newCivitaiAPI(t, func(q url.Values) (int, http.Header, any) {
		return status, header, body
	})
	conn := NewCivitai(srv.URL, testAPIKey, CivitaiPolicy{Creators: []string{"Ajuro"}, MaxContentLevel: "X"}, srv.Client())
	at := time.Now().Add(90 * time.Second).UTC().Format(http.TimeFormat)
	for _, c := range []struct {
		status int
		header http.Header
		check  func(error) bool
	}{
		{429, http.Header{"Retry-After": {"7"}}, func(err error) bool { rl, ok := asRateLimited(err); return ok && rl.RetryAfter == 7*time.Second }},
		{429, http.Header{"Retry-After": {at}}, func(err error) bool {
			rl, ok := asRateLimited(err)
			return ok && rl.RetryAfter > 60*time.Second && rl.RetryAfter <= 90*time.Second
		}},
		{429, nil, func(err error) bool { rl, ok := asRateLimited(err); return ok && rl.RetryAfter == 60*time.Second }},
		{401, nil, func(err error) bool { return errors.Is(err, ErrUnauthorized) }},
		{403, nil, func(err error) bool { return errors.Is(err, ErrUnauthorized) }},
		{500, nil, func(err error) bool { return errors.Is(err, ErrTransient) }},
		{503, http.Header{"Retry-After": {"2"}}, func(err error) bool { return errors.Is(err, ErrTransient) }},
		{400, nil, func(err error) bool {
			_, rl := asRateLimited(err)
			return err != nil && !rl && !errors.Is(err, ErrTransient) && !errors.Is(err, ErrUnauthorized)
		}},
	} {
		status, header = c.status, c.header
		_, err := conn.Fetch(context.Background(), "user:Ajuro", "")
		if !c.check(err) {
			t.Errorf("status %d: err = %v", c.status, err)
		}
		if err != nil && (strings.Contains(err.Error(), "Ajuro") || strings.Contains(err.Error(), "127.0.0.1") || strings.Contains(err.Error(), "secret")) {
			t.Errorf("status %d: error leaks content: %q", c.status, err)
		}
	}
	// The documented cursor refusals reset the cursor; any other 400 does not.
	status, header = 400, nil
	for _, msg := range []string{"This cursor can no longer be continued; start again without it", "This cursor cannot be continued with these filters"} {
		body = map[string]string{"message": msg}
		if _, err := conn.Fetch(context.Background(), "user:Ajuro", "feed:1:2"); !errors.Is(err, ErrCursorExpired) {
			t.Errorf("%q: err = %v", msg, err)
		}
		if _, err := conn.Fetch(context.Background(), "user:Ajuro", ""); err == nil || errors.Is(err, ErrCursorExpired) {
			t.Errorf("%q without a cursor: err = %v", msg, err)
		}
	}
	body = map[string]string{"message": "Invalid tags"}
	if _, err := conn.Fetch(context.Background(), "user:Ajuro", "feed:1:2"); err == nil || errors.Is(err, ErrCursorExpired) {
		t.Errorf("other 400: err = %v", err)
	}
	// Network errors are transient and never carry the URL.
	srv.Close()
	_, err := conn.Fetch(context.Background(), "user:Ajuro", "")
	if !errors.Is(err, ErrTransient) || strings.Contains(err.Error(), "127.0.0.1") {
		t.Fatalf("network error = %v", err)
	}
}

func TestCivitaiTakedownDeletesItem(t *testing.T) {
	var mu sync.Mutex
	removed := map[string]bool{}
	api, srv := newCivitaiAPI(t, func(q url.Values) (int, http.Header, any) {
		mu.Lock()
		defer mu.Unlock()
		if id := q.Get("imageId"); id != "" {
			if removed[id] {
				return ok(page(""))
			}
			var n int
			fmt.Sscan(id, &n)
			return ok(page("", image(n, "Soft", "Ajuro", "SDXL 1.0")))
		}
		var items []map[string]any
		for _, n := range []int{21, 22} {
			if !removed[fmt.Sprint(n)] {
				items = append(items, image(n, "Soft", "Ajuro", "SDXL 1.0"))
			}
		}
		return ok(page("", items...))
	})
	p := CivitaiPolicy{Creators: []string{"Ajuro"}, MaxContentLevel: "X", AllowedLicenses: []string{civitaiLicense}}
	conn := NewCivitai(srv.URL, testAPIKey, p, srv.Client())
	s, _ := newTestStore(t)
	ctx := context.Background()
	e := civitaiEngine(t, s, newDownloader(t), conn, p)
	if pr := e.Sync(ctx).PerProvider["civitai"]; pr.Stored != 2 {
		t.Fatalf("sync 1 = %+v", pr)
	}
	if got := externalIDs(t, s, "civitai"); !reflect.DeepEqual(got, []string{"21", "22"}) {
		t.Fatalf("ids = %v", got)
	}
	mu.Lock()
	removed["21"] = true
	mu.Unlock()
	if pr := e.Sync(ctx).PerProvider["civitai"]; pr.State != "ok" || pr.Removed != 1 {
		t.Fatalf("sync 2 = %+v", pr)
	}
	if got := externalIDs(t, s, "civitai"); !reflect.DeepEqual(got, []string{"22"}) {
		t.Fatalf("ids after takedown = %v", got)
	}
	if _, err := s.Open(ctx, shaHex([]byte("media-21"))); !errors.Is(err, ErrNotFound) {
		t.Fatalf("blob of the removed item still opens: %v", err)
	}
	if _, err := s.Open(ctx, shaHex([]byte("media-22"))); err != nil {
		t.Fatalf("kept item's blob: %v", err)
	}
	lookups := 0
	for _, q := range api.Queries() {
		if q.Has("imageId") {
			lookups++
			if q.Get("nsfw") != "X" || q.Has("username") || q.Has("tags") || q.Has("cursor") {
				t.Errorf("lookup query %v", q)
			}
		}
	}
	if lookups != 2 {
		t.Fatalf("lookups = %d, want one per stored item in sync 2", lookups)
	}
}
