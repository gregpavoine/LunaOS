package discovery

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func dlCandidate(url string) Candidate {
	return Candidate{Item: Item{Provider: "civitai", ExternalID: "42"}, MediaURL: url}
}

func dlTmp(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "dl")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	return dir
}

func partPath(dir string, c Candidate) string {
	h := sha256.Sum256([]byte(c.Provider + "\x00" + c.ExternalID))
	return filepath.Join(dir, hex.EncodeToString(h[:])+".part")
}

func sumHex(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

func dirEntries(t *testing.T, dir string) []string {
	t.Helper()
	es, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, e := range es {
		out = append(out, e.Name())
	}
	return out
}

// rangeServer serves body honouring single "bytes=n-" ranges.
func rangeServer(t *testing.T, body []byte, ranges *[]string) *httptest.Server {
	t.Helper()
	var mu sync.Mutex
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rg := r.Header.Get("Range")
		mu.Lock()
		*ranges = append(*ranges, rg)
		mu.Unlock()
		if rg == "" {
			w.Header().Set("Content-Length", strconv.Itoa(len(body)))
			_, _ = w.Write(body)
			return
		}
		var n int
		if _, err := fmt.Sscanf(rg, "bytes=%d-", &n); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", n, len(body)-1, len(body)))
		w.Header().Set("Content-Length", strconv.Itoa(len(body)-n))
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(body[n:])
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestResumeWithRange(t *testing.T) {
	body := []byte(strings.Repeat("0123456789", 100))
	var ranges []string
	srv := rangeServer(t, body, &ranges)
	dir := dlTmp(t)
	c := dlCandidate(srv.URL + "/m")
	if err := os.WriteFile(partPath(dir, c), body[:300], 0600); err != nil {
		t.Fatal(err)
	}
	path, sha, err := Download(context.Background(), srv.Client(), c, 0, dir)
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if len(ranges) != 1 || ranges[0] != "bytes=300-" {
		t.Fatalf("ranges = %v", ranges)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != string(body) {
		t.Fatalf("final file differs: %v", err)
	}
	if sha != sumHex(body) {
		t.Fatal("sha is not over the whole file")
	}
	if info, _ := os.Stat(path); info.Mode().Perm() != 0600 {
		t.Fatalf("mode = %v", info.Mode().Perm())
	}
	if names := dirEntries(t, dir); len(names) != 1 || !strings.HasSuffix(names[0], ".media") || strings.Contains(path, "42") {
		t.Fatalf("dir = %v", names)
	}
}

func TestServerIgnoringRangeRestartsFromZero(t *testing.T) {
	body := []byte(strings.Repeat("ab", 200))
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(body) }))
	defer srv.Close()
	dir := dlTmp(t)
	c := dlCandidate(srv.URL)
	_ = os.WriteFile(partPath(dir, c), []byte("stale-garbage"), 0600)
	path, sha, err := Download(context.Background(), srv.Client(), c, 0, dir)
	if err != nil || sha != sumHex(body) {
		t.Fatalf("err=%v", err)
	}
	if got, _ := os.ReadFile(path); string(got) != string(body) {
		t.Fatal("content wrong")
	}
}

func TestRangeNotSatisfiableRestartsOnce(t *testing.T) {
	body := []byte("complete-body")
	var calls []string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Header.Get("Range"))
		if r.Header.Get("Range") != "" {
			w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
			return
		}
		_, _ = w.Write(body)
	}))
	defer srv.Close()
	dir := dlTmp(t)
	c := dlCandidate(srv.URL)
	_ = os.WriteFile(partPath(dir, c), []byte("xxxxx"), 0600)
	_, sha, err := Download(context.Background(), srv.Client(), c, 0, dir)
	if err != nil || sha != sumHex(body) || len(calls) != 2 {
		t.Fatalf("err=%v calls=%v", err, calls)
	}
}

func TestMisalignedContentRangeRestarts(t *testing.T) {
	body := []byte(strings.Repeat("z", 100))
	var calls int
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Range") != "" {
			w.Header().Set("Content-Range", "bytes 0-99/100")
			w.WriteHeader(http.StatusPartialContent)
			_, _ = w.Write(body)
			return
		}
		_, _ = w.Write(body)
	}))
	defer srv.Close()
	dir := dlTmp(t)
	c := dlCandidate(srv.URL)
	_ = os.WriteFile(partPath(dir, c), []byte("zzzzz"), 0600)
	path, _, err := Download(context.Background(), srv.Client(), c, 0, dir)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(path); string(got) != string(body) || calls != 2 {
		t.Fatalf("calls=%d len=%d", calls, len(got))
	}
}

func TestOversizeAborted(t *testing.T) {
	// Declared length over the cap: nothing is read.
	var served int
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		served++
		w.Header().Set("Content-Length", "2000")
		_, _ = w.Write(make([]byte, 2000))
	}))
	defer srv.Close()
	dir := dlTmp(t)
	c := dlCandidate(srv.URL)
	if _, _, err := Download(context.Background(), srv.Client(), c, 1000, dir); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("err = %v", err)
	}
	if names := dirEntries(t, dir); len(names) != 0 {
		t.Fatalf("left behind: %v", names)
	}

	// Chunked body with no length: stopped at the cap.
	chunked := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fl := w.(http.Flusher)
		for range 50 {
			_, _ = w.Write(make([]byte, 100))
			fl.Flush()
		}
	}))
	defer chunked.Close()
	_, _, err := Download(context.Background(), chunked.Client(), dlCandidate(chunked.URL), 1000, dir)
	if !errors.Is(err, ErrTooLarge) {
		t.Fatalf("err = %v", err)
	}
	if errorCode(err) != "too_large" {
		t.Fatalf("code = %s", errorCode(err))
	}
	if names := dirEntries(t, dir); len(names) != 0 {
		t.Fatalf("left behind: %v", names)
	}
}

func TestHashMismatchLeavesNothing(t *testing.T) {
	body := []byte("the real bytes")
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(body) }))
	defer srv.Close()
	dir := dlTmp(t)
	c := dlCandidate(srv.URL)
	c.DeclaredSHA256 = sumHex([]byte("something else"))
	path, _, err := Download(context.Background(), srv.Client(), c, 0, dir)
	if !errors.Is(err, ErrHashMismatch) || path != "" {
		t.Fatalf("path=%q err=%v", path, err)
	}
	if names := dirEntries(t, dir); len(names) != 0 {
		t.Fatalf("left behind: %v", names)
	}
	// A matching (case-insensitive) declared hash is accepted.
	c.DeclaredSHA256 = strings.ToUpper(sumHex(body))
	if _, _, err := Download(context.Background(), srv.Client(), c, 0, dir); err != nil {
		t.Fatalf("matching hash: %v", err)
	}
}

func TestCancelledContextLeavesNoFinalFile(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("first-chunk"))
		w.(http.Flusher).Flush()
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer srv.Close()
	defer close(release)
	dir := dlTmp(t)
	ctx, cancel := context.WithCancel(context.Background())
	go func() { time.Sleep(100 * time.Millisecond); cancel() }()
	path, _, err := Download(ctx, srv.Client(), dlCandidate(srv.URL), 0, dir)
	if !errors.Is(err, context.Canceled) || path != "" {
		t.Fatalf("path=%q err=%v", path, err)
	}
	for _, n := range dirEntries(t, dir) {
		if strings.HasSuffix(n, ".media") {
			t.Fatalf("final file exists: %s", n)
		}
	}
}

func TestDownloadRefusesInsecureAndRedirects(t *testing.T) {
	dir := dlTmp(t)
	if _, _, err := Download(context.Background(), http.DefaultClient, dlCandidate("http://media.example.test/x"), 0, dir); err == nil || strings.Contains(err.Error(), "media.example") {
		t.Fatalf("err = %v", err)
	}
	plain := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("x")) }))
	defer plain.Close()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, plain.URL, http.StatusFound)
	}))
	defer srv.Close()
	hc := srv.Client()
	_, _, err := Download(context.Background(), hc, dlCandidate(srv.URL), 0, dir)
	if err == nil || strings.Contains(err.Error(), "127.0.0.1") {
		t.Fatalf("err = %v", err)
	}
	if hc.CheckRedirect != nil {
		t.Fatal("caller's client was mutated")
	}
}

func TestDownloadStatusErrors(t *testing.T) {
	for _, tc := range []struct {
		status int
		check  func(error) bool
		code   string
	}{
		{429, func(err error) bool { rl, ok := asRateLimited(err); return ok && rl.RetryAfter == 7*time.Second }, "rate_limited"},
		{503, func(err error) bool { return errors.Is(err, ErrTransient) }, "transient"},
		{404, func(err error) bool { return errors.Is(err, ErrDownloadGone) }, "gone"},
		{403, func(err error) bool { return errors.Is(err, ErrDownloadRefused) }, "refused"},
	} {
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Retry-After", "7")
			w.WriteHeader(tc.status)
			_, _ = w.Write([]byte("secret body"))
		}))
		_, _, err := Download(context.Background(), srv.Client(), dlCandidate(srv.URL), 0, dlTmp(t))
		srv.Close()
		if err == nil || !tc.check(err) || errorCode(err) != tc.code || strings.Contains(err.Error(), "secret") {
			t.Fatalf("status %d: err=%v code=%s", tc.status, err, errorCode(err))
		}
	}
}

func TestDownloadRejectsBadTmpDirAndSweepsStale(t *testing.T) {
	loose := filepath.Join(t.TempDir(), "loose")
	_ = os.Mkdir(loose, 0755)
	_ = os.Chmod(loose, 0755)
	if _, _, err := Download(context.Background(), http.DefaultClient, dlCandidate("https://x.test/a"), 0, loose); err == nil {
		t.Fatal("accepted a non-private tmp dir")
	}

	body := []byte("fresh")
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(body) }))
	defer srv.Close()
	dir := dlTmp(t)
	stale := filepath.Join(dir, "old.part")
	young := filepath.Join(dir, "young.part")
	_ = os.WriteFile(stale, []byte("x"), 0600)
	_ = os.WriteFile(young, []byte("x"), 0600)
	old := time.Now().Add(-25 * time.Hour)
	_ = os.Chtimes(stale, old, old)
	if _, _, err := Download(context.Background(), srv.Client(), dlCandidate(srv.URL), 0, dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatal("stale part survived")
	}
	if _, err := os.Stat(young); err != nil {
		t.Fatal("young part removed")
	}
}

func TestNewDownloaderFeedsEngine(t *testing.T) {
	body := []byte(strings.Repeat("media-bytes-", 50))
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(body) }))
	defer srv.Close()
	s, _ := newTestStore(t)
	clock := newFakeClock()
	dir := dlTmp(t)
	cand := media("civitai", "dl1")
	cand.MediaURL = srv.URL + "/m"
	cand.DeclaredSHA256 = sumHex(body)
	conn := &fakeConnector{name: "civitai", scopes: []string{"tag:a"}, clock: clock, fetch: func(scope, cursor string, n int) (Page, error) {
		return Page{Items: []Candidate{cand}}, nil
	}}
	e := newTestEngine(t, s, clock, nil, conn)
	e.Download = NewDownloader(srv.Client(), 0, dir)
	rep := e.Sync(context.Background())
	if pr := rep.PerProvider["civitai"]; pr.Stored != 1 || pr.State != "ok" {
		t.Fatalf("report = %+v", pr)
	}
	rc, err := s.Open(context.Background(), sumHex(body))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer rc.Close()
	got, _ := io.ReadAll(rc)
	if string(got) != string(body) {
		t.Fatal("stored media differs")
	}
	if names := dirEntries(t, dir); len(names) != 0 {
		t.Fatalf("tmp dir not empty: %v", names)
	}
}

func TestTooLargeIsNotRetried(t *testing.T) {
	var hits int
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Length", "5000")
		_, _ = w.Write(make([]byte, 5000))
	}))
	defer srv.Close()
	s, _ := newTestStore(t)
	clock := newFakeClock()
	cand := media("civitai", "big")
	cand.MediaURL = srv.URL
	conn := &fakeConnector{name: "civitai", scopes: []string{"tag:a"}, clock: clock, fetch: func(scope, cursor string, n int) (Page, error) {
		return Page{Items: []Candidate{cand}}, nil
	}}
	e := newTestEngine(t, s, clock, nil, conn)
	e.Download = NewDownloader(srv.Client(), 1000, dlTmp(t))
	e.Sync(context.Background())
	if hits != 1 {
		t.Fatalf("server hit %d times", hits)
	}
	if dl := deadLetters(t, s, "civitai"); len(dl) != 1 || dl[0].Code != "too_large" {
		t.Fatalf("dead letters = %+v", dl)
	}
}
