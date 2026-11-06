package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

type env struct {
	dir  string
	args []string
}

func newEnv(t *testing.T, policy, secrets string) env {
	t.Helper()
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	e := env{dir: dir}
	writeFile(t, filepath.Join(dir, "policy.json"), policy)
	e.args = []string{"sync",
		"--policy", filepath.Join(dir, "policy.json"),
		"--secrets", filepath.Join(dir, "secrets.json"),
		"--db", filepath.Join(dir, "index.db"),
		"--blobs", filepath.Join(dir, "blobs"),
		"--key", filepath.Join(dir, "key"),
		"--tmp", filepath.Join(dir, "tmp"),
	}
	if secrets != "" {
		writeFile(t, filepath.Join(dir, "secrets.json"), secrets)
	}
	var out bytes.Buffer
	if code := run([]string{"keygen", "--key", filepath.Join(dir, "key")}, &out); code != 0 {
		t.Fatalf("keygen code=%d", code)
	}
	return e
}

func TestSyncWithNoProvidersConfiguredIsNoop(t *testing.T) {
	// Allowlists are non-empty but there is no secrets file: nothing is
	// configured, so no connector exists and no request can be made.
	e := newEnv(t, `{"civitai":{"tags":["4"],"max_content_level":"X","allowed_licenses":["provider-terms:civitai"]},"reddit":{"subreddits":["pics"]}}`, "")
	hit := false
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hit = true }))
	defer srv.Close()
	var out bytes.Buffer
	code := runWith(e.args, &out, deps{civitaiBase: srv.URL, hc: srv.Client()})
	if code != 0 {
		t.Fatalf("code=%d out=%s", code, out.String())
	}
	if hit {
		t.Fatal("a request was made with no provider configured")
	}
	var rep struct {
		Providers map[string]any `json:"providers"`
	}
	if err := json.Unmarshal(out.Bytes(), &rep); err != nil || len(rep.Providers) != 0 {
		t.Fatalf("report=%s err=%v", out.String(), err)
	}
	// A provider with a secrets block but an empty allowlist stays unconfigured.
	writeFile(t, filepath.Join(e.dir, "secrets.json"), `{"civitai":{"api_key":"k"}}`)
	writeFile(t, filepath.Join(e.dir, "policy.json"), `{}`)
	out.Reset()
	if code = runWith(e.args, &out, deps{civitaiBase: srv.URL, hc: srv.Client()}); code != 0 || hit {
		t.Fatalf("code=%d hit=%v", code, hit)
	}
}

func TestKeygenRefusesOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "key")
	var out bytes.Buffer
	if code := run([]string{"keygen", "--key", path}, &out); code != 0 {
		t.Fatalf("first keygen code=%d", code)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0600 || info.Size() != 32 {
		t.Fatalf("key info=%v err=%v", info, err)
	}
	first, _ := os.ReadFile(path)
	if strings.Contains(out.String(), string(first)) {
		t.Fatal("key printed")
	}
	if code := run([]string{"keygen", "--key", path}, &out); code != 1 {
		t.Fatalf("second keygen code=%d, want 1", code)
	}
	if again, _ := os.ReadFile(path); !bytes.Equal(first, again) {
		t.Fatal("key overwritten")
	}
}

func TestSyncRefusesLooseSecretsFile(t *testing.T) {
	e := newEnv(t, `{}`, `{"civitai":{"api_key":"k"}}`)
	if err := os.Chmod(filepath.Join(e.dir, "secrets.json"), 0640); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if code := runWith(e.args, &out, deps{}); code != 1 {
		t.Fatalf("code=%d", code)
	}
}

func TestReportContainsNoMediaURLs(t *testing.T) {
	const (
		title   = "sensitive-title"
		creator = "sensitive-creator"
		model   = "sensitive-model"
	)
	var srv *httptest.Server
	srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/media/") {
			_, _ = w.Write([]byte("media-bytes"))
			return
		}
		if r.URL.Path != "/api/v1/images" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"items":[{"id":77,"url":%q,"nsfwLevel":"None","username":%q,"baseModel":%q,"title":%q}],"metadata":{}}`,
			srv.URL+"/media/77.jpeg", creator, model, title)
	}))
	defer srv.Close()
	e := newEnv(t, `{"civitai":{"tags":["4"],"max_content_level":"X","allowed_licenses":["provider-terms:civitai"]}}`, `{"civitai":{"api_key":"k"}}`)
	var out bytes.Buffer
	code := runWith(e.args, &out, deps{civitaiBase: srv.URL, hc: srv.Client()})
	if code != 0 {
		t.Fatalf("code=%d out=%s", code, out.String())
	}
	var rep struct {
		Providers map[string]struct {
			State  string `json:"state"`
			Stored int    `json:"stored"`
		} `json:"providers"`
	}
	if err := json.Unmarshal(out.Bytes(), &rep); err != nil {
		t.Fatal(err)
	}
	if got := rep.Providers["civitai"]; got.State != "ok" || got.Stored != 1 {
		t.Fatalf("report=%s", out.String())
	}
	host := strings.TrimPrefix(srv.URL, "https://")
	for _, secret := range []string{"http", "media", "77.jpeg", host, creator, model, title, "127.0.0.1"} {
		if strings.Contains(out.String(), secret) {
			t.Fatalf("report leaks %q: %s", secret, out.String())
		}
	}
}
