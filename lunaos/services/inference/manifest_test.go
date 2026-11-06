package inference

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func sum(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func writeManifest(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "manifest.json")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

const goodHash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func entry(id, url, hash, lic string, size int) string {
	return `{"id":"` + id + `","url":"` + url + `","sha256":"` + hash + `","license":"` + lic + `","size_bytes":` + itoa(size) + `,"vram_mib":1024}`
}

func itoa(n int) string { return strconv.Itoa(n) }

func TestLoadManifestAcceptsValid(t *testing.T) {
	p := writeManifest(t, "["+entry("m1", "https://example.com/m.gguf", goodHash, "MIT", 5)+"]")
	ms, err := LoadManifest(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(ms) != 1 || ms[0].ID != "m1" || ms[0].SizeBytes != 5 || ms[0].VRAMMiB != 1024 || ms[0].License != "MIT" {
		t.Fatalf("unexpected: %+v", ms)
	}
}

func TestLoadManifestRejectsInvalidEntries(t *testing.T) {
	good := "https://example.com/m.gguf"
	cases := map[string]string{
		"duplicate id":   "[" + entry("m1", good, goodHash, "MIT", 5) + "," + entry("m1", good, goodHash, "MIT", 5) + "]",
		"short hash":     "[" + entry("m1", good, "abcd", "MIT", 5) + "]",
		"non-hex hash":   "[" + entry("m1", good, strings.Repeat("z", 64), "MIT", 5) + "]",
		"uppercase hash": "[" + entry("m1", good, strings.ToUpper(goodHash), "MIT", 5) + "]",
		"http url":       "[" + entry("m1", "http://example.com/m.gguf", goodHash, "MIT", 5) + "]",
		"empty url":      "[" + entry("m1", "", goodHash, "MIT", 5) + "]",
		"empty license":  "[" + entry("m1", good, goodHash, "", 5) + "]",
		"zero size":      "[" + entry("m1", good, goodHash, "MIT", 0) + "]",
		"empty id":       "[" + entry("", good, goodHash, "MIT", 5) + "]",
		"path id":        "[" + entry("../evil", good, goodHash, "MIT", 5) + "]",
		"slash id":       "[" + entry("a/b", good, goodHash, "MIT", 5) + "]",
		"dot id":         "[" + entry(".hidden", good, goodHash, "MIT", 5) + "]",
		"upper id":       "[" + entry("Model", good, goodHash, "MIT", 5) + "]",
		"not json":       "nope",
		"unknown field":  `[{"id":"m1","bogus":1}]`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := LoadManifest(writeManifest(t, body)); err == nil {
				t.Fatalf("expected error for %s", name)
			}
		})
	}
	t.Run("negative size", func(t *testing.T) {
		body := `[{"id":"m1","url":"` + good + `","sha256":"` + goodHash + `","license":"MIT","size_bytes":-3,"vram_mib":1}]`
		if _, err := LoadManifest(writeManifest(t, body)); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("missing file", func(t *testing.T) {
		if _, err := LoadManifest(filepath.Join(t.TempDir(), "none.json")); err == nil {
			t.Fatal("expected error")
		}
	})
}

func serve(t *testing.T, h http.HandlerFunc) (*httptest.Server, *int32) {
	t.Helper()
	var n int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&n, 1)
		h(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv, &n
}

func noLeftovers(t *testing.T, dir string, names ...string) {
	t.Helper()
	for _, n := range names {
		if _, err := os.Stat(filepath.Join(dir, n)); !os.IsNotExist(err) {
			t.Errorf("%s should not exist (err=%v)", n, err)
		}
	}
}

func TestFetchVerifiesHash(t *testing.T) {
	data := []byte("hello model")
	srv, _ := serve(t, func(w http.ResponseWriter, r *http.Request) { w.Write(data) })
	dir := t.TempDir()
	m := Model{ID: "m1", URL: srv.URL, SHA256: sum(data), SizeBytes: int64(len(data))}
	p, err := Fetch(context.Background(), m, dir, srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	if p != filepath.Join(dir, "m1.gguf") {
		t.Fatalf("path = %s", p)
	}
	got, _ := os.ReadFile(p)
	if string(got) != string(data) {
		t.Fatal("content mismatch")
	}
	noLeftovers(t, dir, ".m1.part")
}

func TestFetchHashMismatchLeavesNoFile(t *testing.T) {
	data := []byte("hello model")
	srv, _ := serve(t, func(w http.ResponseWriter, r *http.Request) { w.Write(data) })
	dir := t.TempDir()
	m := Model{ID: "m1", URL: srv.URL, SHA256: sum([]byte("other")), SizeBytes: int64(len(data))}
	if _, err := Fetch(context.Background(), m, dir, srv.Client()); err == nil {
		t.Fatal("expected error")
	}
	noLeftovers(t, dir, ".m1.part", "m1.gguf")
}

func TestFetchOversizeAborted(t *testing.T) {
	data := []byte("0123456789")
	srv, _ := serve(t, func(w http.ResponseWriter, r *http.Request) { w.Write(data) })
	dir := t.TempDir()
	m := Model{ID: "m1", URL: srv.URL, SHA256: sum(data), SizeBytes: 4}
	if _, err := Fetch(context.Background(), m, dir, srv.Client()); err == nil {
		t.Fatal("expected error")
	}
	noLeftovers(t, dir, ".m1.part", "m1.gguf")
}

func TestFetchOversizeStreamedWithoutContentLength(t *testing.T) {
	data := []byte("0123456789")
	srv, _ := serve(t, func(w http.ResponseWriter, r *http.Request) {
		// Flush between writes so the response is chunked, with no Content-Length.
		for i := 0; i < len(data); i += 2 {
			w.Write(data[i : i+2])
			w.(http.Flusher).Flush()
		}
	})
	dir := t.TempDir()
	m := Model{ID: "m1", URL: srv.URL, SHA256: sum(data), SizeBytes: 4}
	if _, err := Fetch(context.Background(), m, dir, srv.Client()); err == nil {
		t.Fatal("expected error")
	}
	noLeftovers(t, dir, ".m1.part", "m1.gguf")
}

func TestFetchSkipsVerifiedFile(t *testing.T) {
	data := []byte("cached")
	srv, n := serve(t, func(w http.ResponseWriter, r *http.Request) { w.Write(data) })
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "m1.gguf"), data, 0o644)
	m := Model{ID: "m1", URL: srv.URL, SHA256: sum(data), SizeBytes: int64(len(data))}
	if _, err := Fetch(context.Background(), m, dir, srv.Client()); err != nil {
		t.Fatal(err)
	}
	if atomic.LoadInt32(n) != 0 {
		t.Fatal("made an HTTP request for a verified file")
	}
}

func TestFetchReplacesCorruptFinalFile(t *testing.T) {
	data := []byte("good bytes")
	srv, n := serve(t, func(w http.ResponseWriter, r *http.Request) { w.Write(data) })
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "m1.gguf"), []byte("corrupt"), 0o644)
	m := Model{ID: "m1", URL: srv.URL, SHA256: sum(data), SizeBytes: int64(len(data))}
	p, err := Fetch(context.Background(), m, dir, srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	if atomic.LoadInt32(n) != 1 {
		t.Fatal("expected one download")
	}
	got, _ := os.ReadFile(p)
	if string(got) != string(data) {
		t.Fatal("not replaced")
	}
}

func TestFetchCorruptFinalFileDeletedWhenDownloadFails(t *testing.T) {
	srv, _ := serve(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) })
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "m1.gguf"), []byte("corrupt"), 0o644)
	m := Model{ID: "m1", URL: srv.URL, SHA256: sum([]byte("x")), SizeBytes: 10}
	if _, err := Fetch(context.Background(), m, dir, srv.Client()); err == nil {
		t.Fatal("expected error")
	}
	noLeftovers(t, dir, ".m1.part", "m1.gguf")
}

func TestFetchHTTPErrorStatus(t *testing.T) {
	srv, _ := serve(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(404) })
	dir := t.TempDir()
	m := Model{ID: "m1", URL: srv.URL, SHA256: sum([]byte("x")), SizeBytes: 10}
	if _, err := Fetch(context.Background(), m, dir, srv.Client()); err == nil {
		t.Fatal("expected error")
	}
	noLeftovers(t, dir, ".m1.part", "m1.gguf")
}

func TestFetchInterruptedLeavesNoFinalFile(t *testing.T) {
	data := []byte("0123456789")
	srv, _ := serve(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "10")
		w.Write(data[:5])
		w.(http.Flusher).Flush()
		panic(http.ErrAbortHandler) // drop the connection mid-body
	})
	dir := t.TempDir()
	m := Model{ID: "m1", URL: srv.URL, SHA256: sum(data), SizeBytes: 10}
	if _, err := Fetch(context.Background(), m, dir, srv.Client()); err == nil {
		t.Fatal("expected error")
	}
	noLeftovers(t, dir, ".m1.part", "m1.gguf")
}

func TestFetchContextCancelCleansUp(t *testing.T) {
	release := make(chan struct{})
	srv, _ := serve(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "10")
		w.Write([]byte("01234"))
		w.(http.Flusher).Flush()
		select {
		case <-release:
		case <-r.Context().Done():
		}
	})
	defer close(release)
	dir := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	m := Model{ID: "m1", URL: srv.URL, SHA256: sum([]byte("0123456789")), SizeBytes: 10}
	if _, err := Fetch(ctx, m, dir, srv.Client()); err == nil {
		t.Fatal("expected error")
	}
	noLeftovers(t, dir, ".m1.part", "m1.gguf")
}

func TestShippedManifestIsValid(t *testing.T) {
	ms, err := LoadManifest(filepath.Join("..", "..", "models", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(ms) == 0 {
		t.Fatal("shipped manifest is empty")
	}
}
