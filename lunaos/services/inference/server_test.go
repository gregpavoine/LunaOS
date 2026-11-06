package inference

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"lunaos.local/lunaos/services/internal/peer"
)

const (
	coreUID  = 4242
	otherUID = 4243
)

var modelBytes = []byte("not really gguf weights")

type testEnv struct {
	srv     *Server
	cfg     Config
	url     string
	client  *http.Client
	pidfile string
}

// stampUID returns a ConnContext that records uid as the socket peer.
func stampUID(uid int) func(context.Context, net.Conn) context.Context {
	return func(ctx context.Context, _ net.Conn) context.Context { return peer.WithUID(ctx, uid) }
}

// testConfig writes a model file m1.gguf and a manifest that lists it.
func testConfig(t *testing.T) Config {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "m1.gguf"), modelBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	manifest := writeManifest(t, "["+entry("m1", "https://example.test/m1.gguf", sum(modelBytes), "MIT", len(modelBytes))+"]")
	return Config{
		Manifest:    manifest,
		ModelDir:    dir,
		ModelID:     "m1",
		LlamaServer: buildFakeLlama(t),
		AllowedUID:  coreUID,
		StartWait:   5 * time.Second,
		Log:         io.Discard,
	}
}

// newTestServer builds a Server from testConfig (adjusted by mutate) behind an
// httptest server that stamps the Core UID. setup runs before serving starts.
func newTestServer(t *testing.T, mutate func(*Config), setup func(*Server)) *testEnv {
	t.Helper()
	pidfile := filepath.Join(t.TempDir(), "pids")
	t.Setenv("FAKE_PIDFILE", pidfile)
	cfg := testConfig(t)
	if mutate != nil {
		mutate(&cfg)
	}
	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if srv.sup != nil {
		srv.sup.StartTimeout = 5 * time.Second
		srv.sup.StopGrace = time.Second
	}
	if setup != nil {
		setup(srv)
	}
	t.Cleanup(func() { srv.Close() })
	hs := httptest.NewUnstartedServer(srv)
	hs.Config.ConnContext = stampUID(coreUID)
	hs.Start()
	t.Cleanup(hs.Close)
	return &testEnv{srv: srv, cfg: cfg, url: hs.URL, client: hs.Client(), pidfile: pidfile}
}

// do sends a request and decodes a JSON reply into a map (nil when empty).
func (e *testEnv) do(t *testing.T, method, path, body string) (int, map[string]any) {
	t.Helper()
	return doReq(t, e.client, e.url, method, path, body)
}

func doReq(t *testing.T, c *http.Client, base, method, path, body string) (int, map[string]any) {
	t.Helper()
	status, m, err := request(c, base, method, path, body)
	if err != nil {
		t.Fatal(err)
	}
	return status, m
}

// request is doReq without t, for use from other goroutines.
func request(c *http.Client, base, method, path, body string) (int, map[string]any, error) {
	req, err := http.NewRequest(method, base+path, strings.NewReader(body))
	if err != nil {
		return 0, nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	if len(raw) == 0 {
		return resp.StatusCode, nil, nil
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return 0, nil, fmt.Errorf("%s %s: reply is not a JSON object: %q", method, path, raw)
	}
	return resp.StatusCode, m, nil
}

func genBody(prompt string, maxTokens int) string {
	b, _ := json.Marshal(map[string]any{"prompt": prompt, "max_tokens": maxTokens})
	return string(b)
}

func wantCode(t *testing.T, what string, status int, body map[string]any, wantStatus int, code string) {
	t.Helper()
	if status != wantStatus || body["code"] != code {
		t.Fatalf("%s: got %d %v, want %d code %s", what, status, body, wantStatus, code)
	}
}

func TestPeerOtherThanCoreDenied(t *testing.T) {
	env := newTestServer(t, nil, nil)
	other := httptest.NewUnstartedServer(env.srv)
	other.Config.ConnContext = stampUID(otherUID)
	other.Start()
	defer other.Close()
	unstamped := httptest.NewServer(env.srv) // no peer UID at all
	defer unstamped.Close()

	requests := []struct{ method, path, body string }{
		{"GET", "/v1/health", ""},
		{"POST", "/v1/generate", genBody("hello", 8)},
		{"POST", "/v1/generate", "not json"},
		{"POST", "/v1/unload", ""},
		{"GET", "/v1/generate", ""},
		{"GET", "/nope", ""},
	}
	for _, s := range []*httptest.Server{other, unstamped} {
		for _, r := range requests {
			status, body := doReq(t, s.Client(), s.URL, r.method, r.path, r.body)
			wantCode(t, r.method+" "+r.path, status, body, http.StatusForbidden, "SOCKET_PEER_DENIED")
		}
	}
	if pids := readPIDs(t, env.pidfile); len(pids) != 0 {
		t.Fatalf("a denied peer started llama-server: %v", pids)
	}
	if status, _ := env.do(t, "GET", "/v1/health", ""); status != http.StatusOK {
		t.Fatalf("Core health = %d, want 200", status)
	}
}

func TestGenerateValidatesLimits(t *testing.T) {
	env := newTestServer(t, nil, nil)
	big := strings.Repeat("a", maxPromptBytes+1)
	bad := map[string]string{
		"zero tokens":      genBody("hi", 0),
		"1025 tokens":      genBody("hi", 1025),
		"negative tokens":  genBody("hi", -1),
		"missing tokens":   `{"prompt":"hi"}`,
		"prompt 32 KiB+1":  genBody(big, 8),
		"empty prompt":     genBody("", 8),
		"missing prompt":   `{"max_tokens":8}`,
		"unknown field":    `{"prompt":"hi","max_tokens":8,"temperature":2}`,
		"trailing data":    genBody("hi", 8) + `{}`,
		"not json":         "prompt=hi",
		"wrong type":       `{"prompt":7,"max_tokens":8}`,
		"body over limit":  `{"prompt":"` + strings.Repeat("a", maxBodyBytes) + `","max_tokens":8}`,
		"multibyte 32KiB+": genBody(strings.Repeat("é", maxPromptBytes/2)+"a", 8),
	}
	for name, body := range bad {
		status, reply := env.do(t, "POST", "/v1/generate", body)
		wantCode(t, name, status, reply, http.StatusBadRequest, "INVALID_REQUEST")
	}
	if pids := readPIDs(t, env.pidfile); len(pids) != 0 {
		t.Fatalf("an invalid request started llama-server: %v", pids)
	}

	// Boundaries are accepted.
	exact := strings.Repeat("a", maxPromptBytes)
	if status, reply := env.do(t, "POST", "/v1/generate", genBody(exact, maxTokensLimit)); status != http.StatusOK {
		t.Fatalf("32 KiB prompt, 1024 tokens: %d %v", status, reply)
	}
	// Go's encoder writes '<' as \u003c: a 32 KiB prompt can take six times
	// its size on the wire and must still fit the body limit.
	escaped := genBody(strings.Repeat("<", maxPromptBytes), 1)
	if !strings.Contains(escaped, `\u003c`) || len(escaped) < 6*maxPromptBytes {
		t.Fatalf("test body is not escaped as expected")
	}
	if status, reply := env.do(t, "POST", "/v1/generate", escaped); status != http.StatusOK {
		t.Fatalf("escaped 32 KiB prompt: %d %v", status, reply)
	}

	methods := []struct{ method, path string }{
		{"GET", "/v1/generate"},
		{"PUT", "/v1/generate"},
		{"GET", "/v1/unload"},
		{"POST", "/v1/health"},
	}
	for _, m := range methods {
		status, reply := env.do(t, m.method, m.path, "")
		wantCode(t, m.method+" "+m.path, status, reply, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED")
	}
	status, reply := env.do(t, "GET", "/v1/nope", "")
	wantCode(t, "unknown path", status, reply, http.StatusNotFound, "NOT_FOUND")
}

func TestGenerateMarksOutputUntrusted(t *testing.T) {
	env := newTestServer(t, nil, nil)
	status, reply := env.do(t, "POST", "/v1/generate", genBody("hello", 8))
	if status != http.StatusOK {
		t.Fatalf("generate = %d %v", status, reply)
	}
	if u, ok := reply["untrusted"].(bool); !ok || !u {
		t.Fatalf("untrusted = %#v, want true", reply["untrusted"])
	}
	if reply["text"] != "echo(8):hello" || reply["model"] != "m1" || len(reply) != 3 {
		t.Fatalf("reply = %v", reply)
	}
}

func TestGenerateStartsModelLazily(t *testing.T) {
	env := newTestServer(t, nil, nil)
	status, health := env.do(t, "GET", "/v1/health", "")
	if status != http.StatusOK || health["state"] != "stopped" || health["model"] != "m1" {
		t.Fatalf("health before use = %d %v", status, health)
	}
	time.Sleep(200 * time.Millisecond)
	if pids := readPIDs(t, env.pidfile); len(pids) != 0 {
		t.Fatalf("llama-server started before any generate: %v", pids)
	}
	for i := 0; i < 2; i++ {
		if status, reply := env.do(t, "POST", "/v1/generate", genBody("hi", 4)); status != http.StatusOK {
			t.Fatalf("generate %d = %d %v", i, status, reply)
		}
	}
	if _, health := env.do(t, "GET", "/v1/health", ""); health["state"] != "ready" {
		t.Fatalf("health after generate = %v", health)
	}
	if pids := readPIDs(t, env.pidfile); len(pids) != 1 {
		t.Fatalf("pids = %v, want exactly one llama-server", pids)
	}
}

func TestGenerateWhileDegradedReturns503(t *testing.T) {
	t.Run("missing binary", func(t *testing.T) {
		env := newTestServer(t, func(c *Config) {
			c.LlamaServer = "/nonexistent/llama-server"
			c.StartWait = 30 * time.Second
		}, nil)
		status, reply := env.do(t, "POST", "/v1/generate", genBody("hi", 4))
		wantCode(t, "first generate", status, reply, http.StatusServiceUnavailable, "INFERENCE_DEGRADED")
		if _, health := env.do(t, "GET", "/v1/health", ""); health["state"] != "degraded" {
			t.Fatalf("health = %v, want degraded", health)
		}
		begin := time.Now()
		status, reply = env.do(t, "POST", "/v1/generate", genBody("hi", 4))
		wantCode(t, "generate while degraded", status, reply, http.StatusServiceUnavailable, "INFERENCE_DEGRADED")
		if d := time.Since(begin); d > time.Second {
			t.Fatalf("degraded generate took %v; it must not block", d)
		}
	})
	t.Run("crash loop", func(t *testing.T) {
		t.Setenv("FAKE_EXIT_CODE", "1")
		env := newTestServer(t, nil, nil)
		for i := 0; i < maxFailures; i++ {
			status, reply := env.do(t, "POST", "/v1/generate", genBody("hi", 4))
			if status != http.StatusServiceUnavailable || reply["code"] != "INFERENCE_DEGRADED" {
				t.Fatalf("generate %d = %d %v", i, status, reply)
			}
		}
		if _, health := env.do(t, "GET", "/v1/health", ""); health["state"] != "degraded" {
			t.Fatalf("health = %v, want degraded", health)
		}
		status, reply := env.do(t, "POST", "/v1/generate", genBody("hi", 4))
		wantCode(t, "generate while degraded", status, reply, http.StatusServiceUnavailable, "INFERENCE_DEGRADED")
		if pids := readPIDs(t, env.pidfile); len(pids) != maxFailures {
			t.Fatalf("pids = %v, want %d spawns and none while degraded", pids, maxFailures)
		}
	})
}

func TestGenerateDuringSlowStartTimesOut(t *testing.T) {
	t.Setenv("FAKE_HEALTH_DELAY_MS", "1500")
	env := newTestServer(t, func(c *Config) { c.StartWait = 200 * time.Millisecond }, nil)
	begin := time.Now()
	status, reply := env.do(t, "POST", "/v1/generate", genBody("hi", 4))
	wantCode(t, "generate during start", status, reply, http.StatusServiceUnavailable, "INFERENCE_TIMEOUT")
	if d := time.Since(begin); d > 1200*time.Millisecond {
		t.Fatalf("generate waited %v, want about the 200 ms start wait", d)
	}
	if _, health := env.do(t, "GET", "/v1/health", ""); health["state"] != "starting" {
		t.Fatalf("health = %v, want starting", health)
	}
	// The start goes on without the caller; a later request uses it.
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, health := env.do(t, "GET", "/v1/health", ""); health["state"] == "ready" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("model never became ready")
		}
		time.Sleep(50 * time.Millisecond)
	}
	if status, reply := env.do(t, "POST", "/v1/generate", genBody("hi", 4)); status != http.StatusOK {
		t.Fatalf("generate after start = %d %v", status, reply)
	}
	if pids := readPIDs(t, env.pidfile); len(pids) != 1 {
		t.Fatalf("pids = %v, want one", pids)
	}
}

func TestGenerateUpstreamFailureIsNever200(t *testing.T) {
	cases := map[string][2]string{
		"status 500":      {"FAKE_COMPLETION_STATUS", "500"},
		"invalid json":    {"FAKE_COMPLETION_BODY", "not json"},
		"missing content": {"FAKE_COMPLETION_BODY", `{"tokens":[]}`},
		"empty content":   {"FAKE_COMPLETION_BODY", `{"content":""}`},
		"content number":  {"FAKE_COMPLETION_BODY", `{"content":5}`},
	}
	for name, env := range cases {
		t.Run(name, func(t *testing.T) {
			t.Setenv(env[0], env[1])
			e := newTestServer(t, nil, nil)
			status, reply := e.do(t, "POST", "/v1/generate", genBody("hi", 4))
			wantCode(t, name, status, reply, http.StatusBadGateway, "INFERENCE_UPSTREAM")
		})
	}
}

func TestUnloadDuringGenerateCompletesThenStops(t *testing.T) {
	gate := filepath.Join(t.TempDir(), "gate")
	t.Setenv("FAKE_COMPLETION_GATE", gate)
	inRelay := make(chan struct{})
	unloadEntered := make(chan struct{})
	var relayOnce sync.Once
	env := newTestServer(t, nil, func(s *Server) {
		s.relayHook = func() { relayOnce.Do(func() { close(inRelay) }) }
		s.unloadHook = func() { close(unloadEntered) }
	})

	type result struct {
		status int
		body   map[string]any
		err    error
	}
	gen := make(chan result, 1)
	go func() {
		status, body, err := request(env.client, env.url, "POST", "/v1/generate", genBody("hi", 4))
		gen <- result{status, body, err}
	}()
	<-inRelay
	pids := readPIDs(t, env.pidfile)
	if len(pids) != 1 {
		t.Fatalf("pids = %v", pids)
	}
	unload := make(chan int, 1)
	go func() {
		status, _, _ := request(env.client, env.url, "POST", "/v1/unload", "")
		unload <- status // 0 on a transport error
	}()
	<-unloadEntered
	time.Sleep(150 * time.Millisecond)
	select {
	case s := <-unload:
		t.Fatalf("unload returned %d while a generation was in flight", s)
	default:
	}
	if st := env.srv.sup.State(); st != StateReady || !alive(pids[0]) {
		t.Fatalf("model released during a generation: state %v, alive %v", st, alive(pids[0]))
	}

	if err := os.WriteFile(gate, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	r := <-gen
	if r.err != nil || r.status != http.StatusOK || r.body["text"] != "echo(4):hi" || r.body["untrusted"] != true {
		t.Fatalf("in-flight generate = %d %v", r.status, r.body)
	}
	if s := <-unload; s != http.StatusNoContent {
		t.Fatalf("unload = %d, want 204", s)
	}
	if st := env.srv.sup.State(); st != StateStopped {
		t.Fatalf("state after unload = %v, want stopped", st)
	}
	if alive(pids[0]) {
		t.Fatal("llama-server still running after unload")
	}
	// The next request starts the model again.
	if status, reply := env.do(t, "POST", "/v1/generate", genBody("again", 4)); status != http.StatusOK {
		t.Fatalf("generate after unload = %d %v", status, reply)
	}
	if pids := readPIDs(t, env.pidfile); len(pids) != 2 {
		t.Fatalf("pids = %v, want a second spawn", pids)
	}
}

func TestNoModelConfiguredReturns503(t *testing.T) {
	env := newTestServer(t, func(c *Config) { c.ModelID = "" }, nil)
	status, health := env.do(t, "GET", "/v1/health", "")
	if status != http.StatusOK || health["state"] != "stopped" || health["model"] != "" {
		t.Fatalf("health = %d %v", status, health)
	}
	status, reply := env.do(t, "POST", "/v1/generate", genBody("hi", 4))
	wantCode(t, "generate", status, reply, http.StatusServiceUnavailable, "INFERENCE_NO_MODEL")
	if status, _ := env.do(t, "POST", "/v1/unload", ""); status != http.StatusNoContent {
		t.Fatalf("unload = %d, want 204", status)
	}
	if pids := readPIDs(t, env.pidfile); len(pids) != 0 {
		t.Fatalf("pids = %v, want none", pids)
	}
}

func TestStartupVerifiesModelOnce(t *testing.T) {
	noModel := func(t *testing.T, env *testEnv) {
		t.Helper()
		if _, health := env.do(t, "GET", "/v1/health", ""); health["state"] != "stopped" || health["model"] != "" {
			t.Fatalf("health = %v, want stopped with no model", health)
		}
		status, reply := env.do(t, "POST", "/v1/generate", genBody("hi", 4))
		wantCode(t, "generate", status, reply, http.StatusServiceUnavailable, "INFERENCE_NO_MODEL")
	}
	t.Run("wrong hash is deleted", func(t *testing.T) {
		var path string
		env := newTestServer(t, func(c *Config) {
			path = filepath.Join(c.ModelDir, "m1.gguf")
			if err := os.WriteFile(path, []byte("tampered weights"), 0o644); err != nil {
				t.Fatal(err)
			}
		}, nil)
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("model with a wrong hash was kept: %v", err)
		}
		noModel(t, env)
	})
	t.Run("missing file", func(t *testing.T) {
		env := newTestServer(t, func(c *Config) { os.Remove(filepath.Join(c.ModelDir, "m1.gguf")) }, nil)
		noModel(t, env)
	})
	t.Run("unknown id", func(t *testing.T) {
		var path string
		env := newTestServer(t, func(c *Config) {
			path = filepath.Join(c.ModelDir, "m1.gguf")
			c.ModelID = "m2"
		}, nil)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("an unrelated model file was touched: %v", err)
		}
		noModel(t, env)
	})
	t.Run("verified model", func(t *testing.T) {
		env := newTestServer(t, nil, nil)
		if env.srv.sup == nil || env.srv.sup.ModelPath != filepath.Join(env.cfg.ModelDir, "m1.gguf") {
			t.Fatalf("supervisor not configured for the verified model: %+v", env.srv.sup)
		}
	})
	t.Run("invalid config", func(t *testing.T) {
		for name, mutate := range map[string]func(*Config){
			"bad manifest":      func(c *Config) { c.Manifest = writeManifest(t, "{") },
			"relative binary":   func(c *Config) { c.LlamaServer = "llama-server" },
			"negative uid":      func(c *Config) { c.AllowedUID = -1 },
			"no model dir":      func(c *Config) { c.ModelDir = "" },
			"negative gpu":      func(c *Config) { c.GPULayers = -1 },
			"missing manifest":  func(c *Config) { c.Manifest = filepath.Join(t.TempDir(), "absent.json") },
			"no llama-server":   func(c *Config) { c.LlamaServer = "" },
			"negative waiting":  func(c *Config) { c.StartWait = -time.Second },
			"negative generate": func(c *Config) { c.GenerateTimeout = -time.Second },
		} {
			cfg := testConfig(t)
			mutate(&cfg)
			if srv, err := NewServer(cfg); err == nil {
				srv.Close()
				t.Errorf("%s: NewServer accepted the config", name)
			}
		}
	})
}

// listenUnix serves h on a short Unix socket path (macOS caps them near 104
// bytes) and stamps uid as the peer.
func listenUnix(t *testing.T, h http.Handler, uid int) string {
	t.Helper()
	dir, err := os.MkdirTemp("/tmp", "inf")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	sock := filepath.Join(dir, "s.sock")
	l, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	hs := &http.Server{Handler: h, ConnContext: stampUID(uid)}
	go hs.Serve(l)
	t.Cleanup(func() { hs.Close() })
	return sock
}

func TestClientRoundTrip(t *testing.T) {
	env := newTestServer(t, nil, nil)
	c := NewClient(listenUnix(t, env.srv, coreUID))
	ctx := context.Background()
	res, err := c.Generate(ctx, "hello", 8)
	if err != nil {
		t.Fatal(err)
	}
	if res != (Result{Text: "echo(8):hello", Model: "m1", Untrusted: true}) {
		t.Fatalf("result = %+v", res)
	}

	_, err = c.Generate(ctx, "hello", 0)
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusBadRequest || apiErr.Code != "INVALID_REQUEST" {
		t.Fatalf("invalid generate error = %v", err)
	}

	denied := NewClient(listenUnix(t, env.srv, otherUID))
	_, err = denied.Generate(ctx, "hello", 8)
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusForbidden || apiErr.Code != "SOCKET_PEER_DENIED" {
		t.Fatalf("denied generate error = %v", err)
	}
}

func TestClientRejectsResultNotMarkedUntrusted(t *testing.T) {
	for name, body := range map[string]string{
		"untrusted missing": `{"text":"rm -rf","model":"m1"}`,
		"untrusted false":   `{"text":"rm -rf","model":"m1","untrusted":false}`,
		"not json":          `ok`,
		"trailing data":     `{"text":"x","model":"m1","untrusted":true}{}`,
	} {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, body)
		})
		c := NewClient(listenUnix(t, h, coreUID))
		if res, err := c.Generate(context.Background(), "hi", 4); err == nil {
			t.Errorf("%s: Generate accepted %+v", name, res)
		}
	}
}

func TestClientErrorWithoutCodeKeepsStatus(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})
	_, err := NewClient(listenUnix(t, h, coreUID)).Generate(context.Background(), "hi", 4)
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusInternalServerError {
		t.Fatalf("error = %v", err)
	}
	if !strings.Contains(err.Error(), strconv.Itoa(http.StatusInternalServerError)) {
		t.Fatalf("error text %q lacks the status", err)
	}
}
