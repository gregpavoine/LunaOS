package inference

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"lunaos.local/lunaos/services/internal/peer"
)

const (
	maxPromptBytes = 32 * 1024 // prompt limit, in bytes of UTF-8
	maxTokensLimit = 1024      // max_tokens is in [1, maxTokensLimit]
	// An encoder may write each prompt byte as a six-byte \uXXXX escape
	// (Go's writes '<' as <), so the body limit allows for that.
	maxBodyBytes = 6*maxPromptBytes + 4096
	// llama-server's reply carries the text plus timings and settings.
	maxUpstreamBytes = 4 << 20

	// DefaultStartWait bounds how long a generate request waits for the
	// model to start before INFERENCE_TIMEOUT.
	DefaultStartWait = 60 * time.Second
	// DefaultStartTimeout bounds the model load itself. It is longer than the
	// request wait: a slow load goes on after the request times out, and a
	// later request finds the model ready.
	DefaultStartTimeout = 5 * time.Minute
	// DefaultGenerateTimeout bounds one llama-server completion.
	DefaultGenerateTimeout = 5 * time.Minute
	// DefaultContextSize holds a prompt at maxPromptBytes plus maxTokensLimit
	// tokens; llama-server's own default (often 4096) does not.
	DefaultContextSize = 16384
)

// Config is the startup configuration of the inference service.
type Config struct {
	Manifest        string // manifest path; required
	ModelDir        string // holds <id>.gguf; required
	ModelID         string // "" -> serve with no model
	LlamaServer     string // absolute path of llama-server; required
	GPULayers       int
	ContextSize     int           // llama-server context in tokens; 0 -> DefaultContextSize
	AllowedUID      int           // the only peer UID served (the Core)
	StartWait       time.Duration // 0 -> DefaultStartWait
	StartTimeout    time.Duration // 0 -> DefaultStartTimeout; must exceed StartWait
	GenerateTimeout time.Duration // 0 -> DefaultGenerateTimeout
	Log             io.Writer     // service and llama-server log; nil -> os.Stderr
	KeyDir          string        // private dir for llama-server's API key; "" -> temp dir
}

// Server is the HTTP API of luna-inference. Every route first checks that the
// socket peer is the allowed UID.
type Server struct {
	modelID    string      // "" when no verified model is available
	sup        *Supervisor // nil when no verified model is available
	allowedUID int
	startWait  time.Duration
	upstream   *http.Client
	log        io.Writer

	// Generations hold gen for reading; unload holds it for writing, so it
	// waits for generations in flight and new ones wait for it.
	gen sync.RWMutex

	// Test hooks; nil in production.
	relayHook  func() // generate, holding gen, just before the upstream call
	unloadHook func() // unload, just before it takes gen
}

// NewServer validates cfg and verifies the selected model file against the
// manifest once. A model that is not configured, not in the manifest or not on
// disk leaves the server running with no model; a file with a wrong sha256 is
// deleted first. Configuration errors are returned.
func NewServer(cfg Config) (*Server, error) {
	switch {
	case cfg.Manifest == "" || cfg.ModelDir == "" || cfg.LlamaServer == "":
		return nil, errors.New("manifest, model directory and llama-server are required")
	case !filepath.IsAbs(cfg.LlamaServer):
		return nil, fmt.Errorf("llama-server %q must be an absolute path", cfg.LlamaServer)
	case cfg.AllowedUID < 0:
		return nil, errors.New("allowed UID must not be negative")
	case cfg.GPULayers < 0:
		return nil, errors.New("GPU layers must not be negative")
	case cfg.ContextSize < 0:
		return nil, errors.New("context size must not be negative")
	case cfg.StartWait < 0 || cfg.StartTimeout < 0 || cfg.GenerateTimeout < 0:
		return nil, errors.New("timeouts must not be negative")
	}
	startWait, startTimeout := cfg.StartWait, cfg.StartTimeout
	if startWait == 0 {
		startWait = DefaultStartWait
	}
	if startTimeout == 0 {
		startTimeout = DefaultStartTimeout
	}
	if startTimeout <= startWait {
		// Otherwise a load is killed when the request gives up and never comes up.
		return nil, fmt.Errorf("start timeout %v must exceed the start wait %v", startTimeout, startWait)
	}
	models, err := LoadManifest(cfg.Manifest)
	if err != nil {
		return nil, err
	}
	s := &Server{
		allowedUID: cfg.AllowedUID,
		startWait:  startWait,
		log:        cfg.Log,
	}
	if s.log == nil {
		s.log = os.Stderr
	}
	genTimeout := cfg.GenerateTimeout
	if genTimeout == 0 {
		genTimeout = DefaultGenerateTimeout
	}
	s.upstream = loopbackClient(genTimeout)
	ctxSize := cfg.ContextSize
	if ctxSize == 0 {
		ctxSize = DefaultContextSize
	}

	if path, ok := s.verifiedModel(models, cfg.ModelDir, cfg.ModelID); ok {
		s.modelID = cfg.ModelID
		s.sup = &Supervisor{Binary: cfg.LlamaServer, ModelPath: path, GPULayers: cfg.GPULayers,
			ContextSize: ctxSize, StartTimeout: startTimeout, Log: s.log, KeyDir: cfg.KeyDir}
	}
	return s, nil
}

// verifiedModel returns the path of <dir>/<id>.gguf when its sha256 matches
// the manifest. A mismatching file is deleted.
func (s *Server) verifiedModel(models []Model, dir, id string) (string, bool) {
	if id == "" {
		s.logf("no model selected; generation is unavailable")
		return "", false
	}
	var m *Model
	for i := range models {
		if models[i].ID == id {
			m = &models[i]
		}
	}
	if m == nil {
		s.logf("model %q is not in the manifest; generation is unavailable", id)
		return "", false
	}
	path := filepath.Join(dir, m.ID+".gguf")
	got, err := fileHash(path)
	if err != nil {
		s.logf("model %s unavailable: %v", m.ID, err)
		return "", false
	}
	if got != m.SHA256 {
		if err := os.Remove(path); err != nil {
			s.logf("model %s has a wrong sha256 and could not be deleted: %v", m.ID, err)
		} else {
			s.logf("model %s has a wrong sha256; file deleted", m.ID)
		}
		return "", false
	}
	return path, true
}

// loopbackClient talks to llama-server only: no proxy, no redirect, and no
// connection to a non-loopback address.
func loopbackClient(timeout time.Duration) *http.Client {
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: nil,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				host, _, err := net.SplitHostPort(addr)
				if err != nil {
					return nil, err
				}
				if ip := net.ParseIP(host); ip == nil || !ip.IsLoopback() {
					return nil, fmt.Errorf("llama-server address %q is not loopback", addr)
				}
				return dialer.DialContext(ctx, network, addr)
			},
			MaxIdleConns:    4,
			IdleConnTimeout: 30 * time.Second,
		},
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}

// Close stops llama-server; no child outlives it.
func (s *Server) Close() error {
	if s.sup == nil {
		return nil
	}
	return s.sup.Stop()
}

func (s *Server) logf(format string, args ...any) {
	fmt.Fprintf(s.log, "luna-inference: "+format+"\n", args...)
}

type errorReply struct {
	Code string `json:"code"`
}

func reply(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, status int, code string) {
	reply(w, status, errorReply{Code: code})
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !peer.Allowed(r.Context(), s.allowedUID) {
		fail(w, http.StatusForbidden, "SOCKET_PEER_DENIED")
		return
	}
	var method string
	var handle func(http.ResponseWriter, *http.Request)
	switch r.URL.Path {
	case "/v1/health":
		method, handle = http.MethodGet, s.health
	case "/v1/generate":
		method, handle = http.MethodPost, s.generate
	case "/v1/unload":
		method, handle = http.MethodPost, s.unload
	default:
		fail(w, http.StatusNotFound, "NOT_FOUND")
		return
	}
	if r.Method != method {
		w.Header().Set("Allow", method)
		fail(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED")
		return
	}
	handle(w, r)
}

type healthReply struct {
	State string `json:"state"`
	Model string `json:"model"`
}

// health reports the supervisor state; it never starts or waits for the model.
func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	state := StateStopped
	if s.sup != nil {
		state = s.sup.State()
	}
	reply(w, http.StatusOK, healthReply{State: state.String(), Model: s.modelID})
}

type generateRequest struct {
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens"`
}

type generateReply struct {
	Text      string `json:"text"`
	Model     string `json:"model"`
	Untrusted bool   `json:"untrusted"`
}

func (s *Server) generate(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	var req generateRequest
	if dec.Decode(&req) != nil || dec.Decode(new(any)) != io.EOF ||
		req.Prompt == "" || len(req.Prompt) > maxPromptBytes ||
		req.MaxTokens < 1 || req.MaxTokens > maxTokensLimit {
		fail(w, http.StatusBadRequest, "INVALID_REQUEST")
		return
	}
	if s.sup == nil {
		fail(w, http.StatusServiceUnavailable, "INFERENCE_NO_MODEL")
		return
	}
	if s.sup.State() == StateDegraded { // answer at once, even during an unload
		fail(w, http.StatusServiceUnavailable, "INFERENCE_DEGRADED")
		return
	}

	s.gen.RLock()
	defer s.gen.RUnlock()
	startCtx, cancel := context.WithTimeout(r.Context(), s.startWait)
	base, key, err := s.sup.Connect(startCtx)
	waited := startCtx.Err() != nil
	cancel()
	if err != nil {
		switch {
		case waited:
			fail(w, http.StatusServiceUnavailable, "INFERENCE_TIMEOUT")
		case s.sup.State() == StateDegraded: // set before Connect returned
			s.logf("llama-server start failed: %v", err)
			fail(w, http.StatusServiceUnavailable, "INFERENCE_DEGRADED")
		default:
			// Within the failure budget: the next request starts it again.
			s.logf("llama-server start failed: %v", err)
			fail(w, http.StatusServiceUnavailable, "INFERENCE_UNAVAILABLE")
		}
		return
	}
	if s.relayHook != nil {
		s.relayHook()
	}
	text, err := s.complete(r.Context(), base, key, req)
	if err != nil {
		s.logf("completion failed: %v", err)
		fail(w, http.StatusBadGateway, "INFERENCE_UPSTREAM")
		return
	}
	// Model output is data, never an instruction: always flagged untrusted.
	reply(w, http.StatusOK, generateReply{Text: text, Model: s.modelID, Untrusted: true})
}

// complete relays one request to llama-server's POST /completion with the
// per-spawn API key. Anything but
// a 200 with a non-empty string "content" is an error, so an empty or failed
// completion is never passed on as a success.
func (s *Server) complete(ctx context.Context, base, key string, req generateRequest) (string, error) {
	body, err := json.Marshal(struct {
		Prompt   string `json:"prompt"`
		NPredict int    `json:"n_predict"`
	}{req.Prompt, req.MaxTokens})
	if err != nil {
		return "", err
	}
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/completion", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	hreq.Header.Set("Content-Type", "application/json")
	hreq.Header.Set("Authorization", "Bearer "+key)
	resp, err := s.upstream.Do(hreq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("llama-server answered HTTP %d", resp.StatusCode)
	}
	var out struct {
		Content *string `json:"content"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxUpstreamBytes)).Decode(&out); err != nil {
		return "", fmt.Errorf("llama-server reply: %w", err)
	}
	if out.Content == nil || *out.Content == "" {
		return "", errors.New("llama-server reply has no content")
	}
	return *out.Content, nil
}

// unload waits for generations in flight, then stops llama-server. Requests
// arriving meanwhile wait for it and then start the model again lazily.
func (s *Server) unload(w http.ResponseWriter, _ *http.Request) {
	if s.sup != nil {
		if s.unloadHook != nil {
			s.unloadHook()
		}
		s.gen.Lock()
		s.sup.Stop()
		s.gen.Unlock()
	}
	w.WriteHeader(http.StatusNoContent)
}
