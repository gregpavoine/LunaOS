// Command luna-discovery syncs the owner's allowlisted provider scopes into
// the encrypted discovery store and generates the store key. Its output is
// codes and counters only; provider content never reaches stdout or stderr.
package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"lunaos.local/lunaos/services/discovery"
)

const (
	civitaiBase     = "https://civitai.com"
	defaultTmpDir   = "/var/lib/luna-discovery/tmp"
	maxSecretsBytes = 1 << 20
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout))
}

// deps carries the endpoints and HTTP client a test replaces; production
// leaves every field zero and gets the official hosts.
type deps struct {
	civitaiBase string
	reddit      discovery.RedditConfig // TokenURL and APIBase only
	hc          *http.Client
}

func run(args []string, stdout io.Writer) int {
	return runWith(args, stdout, deps{})
}

func fail(code string) int {
	fmt.Fprintln(os.Stderr, code)
	return 1
}

func runWith(args []string, stdout io.Writer, d deps) int {
	if len(args) == 0 {
		return fail("DISCOVERY_USAGE")
	}
	switch args[0] {
	case "keygen":
		return keygen(args[1:], stdout)
	case "sync":
		return syncCmd(args[1:], stdout, d)
	}
	return fail("DISCOVERY_USAGE")
}

func keygen(args []string, stdout io.Writer) int {
	fs := flag.NewFlagSet("keygen", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	path := fs.String("key", "", "new 32-byte key file")
	if fs.Parse(args) != nil || *path == "" || fs.NArg() != 0 {
		return fail("DISCOVERY_USAGE")
	}
	key := make([]byte, 32)
	defer clear(key)
	if _, err := rand.Read(key); err != nil {
		return fail("DISCOVERY_KEYGEN")
	}
	f, err := os.OpenFile(*path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fail("DISCOVERY_KEYGEN")
	}
	_, werr := f.Write(key)
	if cerr := f.Close(); werr != nil || cerr != nil {
		_ = os.Remove(*path)
		return fail("DISCOVERY_KEYGEN")
	}
	fmt.Fprintln(stdout, "key file created")
	return 0
}

// secrets is the owner's provider credentials file.
type secrets struct {
	Civitai *struct {
		APIKey string `json:"api_key"`
	} `json:"civitai"`
	Reddit *struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		RefreshToken string `json:"refresh_token"`
		UserAgent    string `json:"user_agent"`
	} `json:"reddit"`
}

var errSecrets = errors.New("DISCOVERY_SECRETS")

// loadSecrets reads the credentials file: a regular file with no permission
// bit outside 0600 and no unknown fields. A missing file means no provider.
func loadSecrets(path string) (secrets, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return secrets{}, nil
	}
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&^0600 != 0 {
		return secrets{}, errSecrets
	}
	f, err := os.Open(path)
	if err != nil {
		return secrets{}, errSecrets
	}
	defer f.Close()
	if opened, err := f.Stat(); err != nil || !os.SameFile(info, opened) {
		return secrets{}, errSecrets
	}
	data, err := io.ReadAll(io.LimitReader(f, maxSecretsBytes+1))
	defer clear(data)
	if err != nil || len(data) > maxSecretsBytes {
		return secrets{}, errSecrets
	}
	var s secrets
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if dec.Decode(&s) != nil {
		return secrets{}, errSecrets
	}
	if _, err := dec.Token(); err != io.EOF {
		return secrets{}, errSecrets
	}
	return s, nil
}

// buildConnectors returns a connector only for a provider whose secrets are
// complete and whose policy allowlist is non-empty.
func buildConnectors(s secrets, p discovery.Policy, hc *http.Client, d deps) []discovery.Connector {
	var out []discovery.Connector
	if c := s.Civitai; c != nil && strings.TrimSpace(c.APIKey) != "" && len(p.Civitai.Tags)+len(p.Civitai.Creators) > 0 {
		base := d.civitaiBase
		if base == "" {
			base = civitaiBase
		}
		out = append(out, discovery.NewCivitai(base, c.APIKey, p.Civitai, hc))
	}
	if r := s.Reddit; r != nil && r.ClientID != "" && r.RefreshToken != "" && strings.TrimSpace(r.UserAgent) != "" && len(p.Reddit.Subreddits) > 0 {
		cfg := d.reddit
		cfg.ClientID, cfg.ClientSecret, cfg.RefreshToken, cfg.UserAgent = r.ClientID, r.ClientSecret, r.RefreshToken, r.UserAgent
		out = append(out, discovery.NewReddit(cfg, p.Reddit, hc))
	}
	return out
}

func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Minute,
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           (&net.Dialer{Timeout: 15 * time.Second}).DialContext,
			TLSHandshakeTimeout:   15 * time.Second,
			ResponseHeaderTimeout: 60 * time.Second,
		},
	}
}

type providerOut struct {
	State      string `json:"state"`
	Stored     int    `json:"stored"`
	Duplicates int    `json:"duplicates"`
	Rejected   int    `json:"rejected"`
	Removed    int    `json:"removed"`
}

func syncCmd(args []string, stdout io.Writer, d deps) int {
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	policyPath := fs.String("policy", "/etc/lunaos/discovery/policy.json", "owner policy file")
	secretsPath := fs.String("secrets", "/etc/lunaos/discovery/secrets.json", "provider credentials file")
	db := fs.String("db", "/var/lib/luna-discovery/index.db", "index database")
	blobs := fs.String("blobs", "/var/lib/luna-discovery/blobs", "encrypted blob directory")
	key := fs.String("key", "/var/lib/luna-discovery/key", "32-byte key file")
	tmp := fs.String("tmp", defaultTmpDir, "private download directory")
	maxBytes := fs.Int64("max-bytes", discovery.DefaultMaxBytes, "per-download size cap")
	if fs.Parse(args) != nil || fs.NArg() != 0 || *maxBytes <= 0 {
		return fail("DISCOVERY_USAGE")
	}
	policy, err := discovery.LoadPolicy(*policyPath)
	if err != nil {
		return fail(codeOf(err, "DISCOVERY_POLICY"))
	}
	sec, err := loadSecrets(*secretsPath)
	if err != nil {
		return fail("DISCOVERY_SECRETS")
	}
	if err = os.MkdirAll(*tmp, 0700); err != nil {
		return fail("DISCOVERY_TMP")
	}
	if err = os.Chmod(*tmp, 0700); err != nil {
		return fail("DISCOVERY_TMP")
	}
	store, err := discovery.OpenStore(*db, *blobs, *key)
	if err != nil {
		return fail(codeOf(err, "DISCOVERY_STORE"))
	}
	defer store.Close()

	hc := d.hc
	if hc == nil {
		hc = newHTTPClient()
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	engine := discovery.Engine{
		Store:      store,
		Connectors: buildConnectors(sec, policy, hc, d),
		Policy:     policy,
		Now:        time.Now,
		Sleep:      sleep,
		Download:   discovery.NewDownloader(hc, *maxBytes, *tmp),
	}
	report := engine.Sync(ctx)

	out := struct {
		Providers map[string]providerOut `json:"providers"`
	}{Providers: map[string]providerOut{}}
	anyOK := len(report.PerProvider) == 0
	for name, r := range report.PerProvider {
		out.Providers[name] = providerOut{r.State, r.Stored, r.Duplicates, r.Rejected, r.Removed}
		anyOK = anyOK || r.State == "ok"
	}
	if err = json.NewEncoder(stdout).Encode(out); err != nil {
		return fail("DISCOVERY_OUTPUT")
	}
	if !anyOK {
		return 1
	}
	return 0
}

// codeOf returns the error text when it is a bare DISCOVERY_ code, else the
// fallback, so no path or content ever reaches stderr.
func codeOf(err error, fallback string) string {
	if s := err.Error(); strings.HasPrefix(s, "DISCOVERY_") && !strings.ContainsAny(s, " /:") {
		return s
	}
	return fallback
}

func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
