package inference

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"
)

// State is the lifecycle state of the supervised llama-server.
type State int

const (
	StateStopped State = iota
	StateStarting
	StateReady
	StateDegraded
)

func (s State) String() string {
	switch s {
	case StateStarting:
		return "starting"
	case StateReady:
		return "ready"
	case StateDegraded:
		return "degraded"
	default:
		return "stopped"
	}
}

const (
	// maxFailures failures within failureWindow make the supervisor degraded.
	maxFailures   = 3
	failureWindow = 60 * time.Second

	defaultStopGrace    = 10 * time.Second
	defaultStartTimeout = 60 * time.Second
	healthInterval      = 100 * time.Millisecond
)

// Supervisor runs one llama-server child bound to 127.0.0.1. It does not
// restart in the background: callers Start it lazily, and repeated failures
// make it degraded until Reset.
type Supervisor struct {
	Binary       string // absolute path; no PATH lookup
	ModelPath    string
	GPULayers    int
	ContextSize  int           // llama-server -c; 0 -> llama-server's default
	StartTimeout time.Duration // 0 -> 60 s
	StopGrace    time.Duration // SIGTERM to SIGKILL delay; 0 -> 10 s
	Log          io.Writer     // child stdout/stderr; nil -> os.Stderr (journal)
	// KeyDir holds the per-spawn API key file. It must be a private 0700
	// directory owned by this UID (created 0700 if absent). "" -> a fresh
	// os.MkdirTemp directory per spawn, removed when the child ends.
	KeyDir string

	mu       sync.Mutex // never held while waiting on the child or the network
	state    State
	run      *run
	baseURL  string
	key      string // API key of the ready child, published with baseURL
	failures []time.Time
	cur      *start // start in progress, if any

	// Test hooks; nil in production.
	beforeSpawn func() // launch, before the binary check
	afterReady  func() // launch, just after state became Ready, before it returns
	stopHook    func() // Stop, after it inspected state, before it terminates
}

// run is one child process. stopped is guarded by Supervisor.mu.
type run struct {
	cmd     *exec.Cmd
	done    chan struct{} // closed after the child is reaped
	stopped bool          // Stop asked it to end
	waitErr error         // cmd.Wait result, set before done closes
	cleanup func()        // removes the key file once the child is reaped
}

// start is one shared start attempt.
type start struct {
	done   chan struct{}
	cancel context.CancelFunc
	url    string
	key    string
	err    error
}

// probeClient checks /health and the key; loopback only, so no proxy.
var probeClient = &http.Client{
	Timeout:       time.Second,
	Transport:     &http.Transport{Proxy: nil},
	CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
}

// State returns the current state; it never waits for a start in progress.
func (s *Supervisor) State() State {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}

// Reset clears the failure history; a degraded supervisor becomes stopped.
func (s *Supervisor) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failures = nil
	if s.state == StateDegraded {
		s.state = StateStopped
	}
}

// recordFailure counts one failure and may degrade. Caller holds s.mu.
func (s *Supervisor) recordFailure() {
	now := time.Now()
	kept := s.failures[:0]
	for _, t := range s.failures {
		if now.Sub(t) < failureWindow {
			kept = append(kept, t)
		}
	}
	s.failures = append(kept, now)
	if len(s.failures) >= maxFailures {
		s.state = StateDegraded
	}
}

// Start is Connect without the key.
func (s *Supervisor) Start(ctx context.Context) (string, error) {
	u, _, err := s.Connect(ctx)
	return u, err
}

// Connect launches llama-server if needed and returns its base URL and the
// per-spawn API key once /health answers 200 and the child proved it holds
// the key. Requests to the child must send "Authorization: Bearer <key>".
// Concurrent callers share one start. The start runs detached from the
// callers: a caller's ctx only bounds its own wait, and only StartTimeout or
// Stop abort the start itself.
func (s *Supervisor) Connect(ctx context.Context) (string, string, error) {
	s.mu.Lock()
	switch s.state {
	case StateReady:
		u, k := s.baseURL, s.key
		s.mu.Unlock()
		return u, k, nil
	case StateDegraded:
		s.mu.Unlock()
		return "", "", errors.New("llama-server is degraded")
	case StateStopped:
		st := &start{done: make(chan struct{})}
		startCtx, cancel := context.WithCancel(context.Background())
		st.cancel = cancel
		s.state = StateStarting
		s.cur = st
		go func() {
			st.url, st.key, st.err = s.launch(startCtx)
			cancel()
			s.mu.Lock()
			if s.cur == st {
				s.cur = nil
			}
			s.mu.Unlock()
			close(st.done)
		}()
	}
	st := s.cur
	s.mu.Unlock()

	select {
	case <-st.done:
		return st.url, st.key, st.err
	case <-ctx.Done():
		return "", "", ctx.Err()
	}
}

// launch spawns the child and waits for readiness. It leaves s.state at Ready,
// Stopped or Degraded. startCtx is cancelled only by Stop.
func (s *Supervisor) launch(startCtx context.Context) (string, string, error) {
	fail := func(err error, count bool) (string, string, error) {
		s.mu.Lock()
		if count {
			s.recordFailure()
		}
		if s.state != StateDegraded {
			s.state = StateStopped
		}
		s.mu.Unlock()
		return "", "", err
	}
	degrade := func(err error) (string, string, error) {
		s.mu.Lock()
		s.state = StateDegraded
		s.mu.Unlock()
		return "", "", err
	}
	errStopped := errors.New("llama-server stopped during start")

	if s.beforeSpawn != nil {
		s.beforeSpawn()
	}
	if !filepath.IsAbs(s.Binary) {
		return degrade(fmt.Errorf("llama-server binary %q must be an absolute path", s.Binary))
	}
	if fi, err := os.Stat(s.Binary); err != nil || fi.IsDir() || fi.Mode().Perm()&0o111 == 0 {
		return degrade(fmt.Errorf("llama-server binary %q missing or not executable", s.Binary))
	}
	if startCtx.Err() != nil {
		return fail(errStopped, false)
	}
	port, err := freePort()
	if err != nil {
		return fail(err, false)
	}
	key, keyFile, cleanup, err := s.writeKey()
	if err != nil {
		return degrade(fmt.Errorf("llama-server API key: %w", err))
	}
	// The key goes through a file: argv is world-readable in /proc.
	args := []string{"-m", s.ModelPath, "--host", "127.0.0.1",
		"--port", strconv.Itoa(port), "-ngl", strconv.Itoa(s.GPULayers),
		"--api-key-file", keyFile}
	if s.ContextSize > 0 {
		args = append(args, "-c", strconv.Itoa(s.ContextSize))
	}
	cmd := exec.Command(s.Binary, args...)
	var out io.Writer = os.Stderr
	if s.Log != nil {
		out = s.Log
	}
	cmd.Stdout, cmd.Stderr = out, out
	// A grandchild holding the output pipe must not block Wait forever.
	cmd.WaitDelay = s.grace()
	if err := cmd.Start(); err != nil {
		cleanup()
		return degrade(fmt.Errorf("start llama-server: %w", err))
	}
	r := &run{cmd: cmd, done: make(chan struct{}), cleanup: cleanup}
	s.mu.Lock()
	s.run = r
	s.mu.Unlock()
	go s.reap(r)
	if startCtx.Err() != nil { // Stop ran before the child existed
		s.terminate(r, true)
		return fail(errStopped, false)
	}

	timeout := s.StartTimeout
	if timeout <= 0 {
		timeout = defaultStartTimeout
	}
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	tick := time.NewTicker(healthInterval)
	defer tick.Stop()
	base := "http://127.0.0.1:" + strconv.Itoa(port)

	for {
		select {
		case <-r.done:
			s.mu.Lock()
			stopped, werr := r.stopped, r.waitErr
			s.mu.Unlock()
			if stopped || startCtx.Err() != nil {
				return fail(errStopped, false)
			}
			return fail(fmt.Errorf("llama-server exited before becoming healthy: %v", werr), true)
		case <-startCtx.Done():
			s.terminate(r, true)
			return fail(errStopped, false)
		case <-deadline.C:
			s.terminate(r, false)
			return fail(fmt.Errorf("llama-server not healthy after %v", timeout), startCtx.Err() == nil)
		case <-tick.C:
			if !healthy(base) {
				continue
			}
			// /health is public: whoever holds the port answers it. Only a
			// server that knows the key both refuses a wrong one and accepts it.
			if err := provesKey(base, key); err != nil {
				s.terminate(r, false)
				return fail(fmt.Errorf("llama-server on %s failed the key check: %w", base, err), startCtx.Err() == nil)
			}
			s.mu.Lock()
			select {
			case <-r.done: // died right after answering; next loop turn handles it
				s.mu.Unlock()
				continue
			default:
			}
			if startCtx.Err() != nil { // Stop cancels under mu, so this is final
				s.mu.Unlock()
				continue
			}
			s.state = StateReady
			s.baseURL = base // published with the state, never after it
			s.key = key
			s.mu.Unlock()
			if s.afterReady != nil {
				s.afterReady()
			}
			return base, key, nil
		}
	}
}

// reap waits for the child and applies its exit to the state.
func (s *Supervisor) reap(r *run) {
	err := r.cmd.Wait()
	r.cleanup()
	s.mu.Lock()
	r.waitErr = err
	if s.run == r {
		s.run = nil
		// During Starting the starter observes done and accounts for it.
		if s.state == StateReady {
			s.state = StateStopped
			if !r.stopped {
				s.recordFailure()
			}
		}
	}
	close(r.done)
	s.mu.Unlock()
}

func (s *Supervisor) grace() time.Duration {
	if s.StopGrace > 0 {
		return s.StopGrace
	}
	return defaultStopGrace
}

// terminate sends SIGTERM, then SIGKILL after StopGrace, and waits for the
// child to be reaped. markStopped flags the exit as requested, not a failure.
func (s *Supervisor) terminate(r *run, markStopped bool) {
	if markStopped {
		s.mu.Lock()
		r.stopped = true
		s.mu.Unlock()
	}
	select {
	case <-r.done:
		return
	default:
	}
	r.cmd.Process.Signal(syscall.SIGTERM)
	t := time.NewTimer(s.grace())
	defer t.Stop()
	select {
	case <-r.done:
	case <-t.C:
		r.cmd.Process.Kill()
		<-r.done
	}
}

// Stop ends a start in progress and terminates the child (SIGTERM, then
// SIGKILL after StopGrace). When it returns no child is left. It is a no-op
// when nothing is running.
func (s *Supervisor) Stop() error {
	s.mu.Lock()
	st := s.cur
	if st != nil {
		st.cancel() // under mu: a start can no longer turn Ready
	}
	r := s.run
	if r != nil && s.state == StateReady {
		s.state = StateStopped
	}
	s.mu.Unlock()
	if s.stopHook != nil {
		s.stopHook()
	}
	if r != nil {
		s.terminate(r, true)
	}
	if st != nil {
		<-st.done // the start kills a child it spawned after we looked
	}
	return nil
}

func healthy(base string) bool {
	resp, err := probeClient.Get(base + "/health")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// provesKey checks that the server at base holds key: GET /props (behind the
// API key in llama-server) must answer 401 to a wrong key, then 200 to key.
// The wrong key goes first; the right one is only sent to a server that
// checks keys. If the check fails, the run is abandoned with its key.
func provesKey(base, key string) error {
	wrong, err := newKey()
	if err != nil {
		return err
	}
	for _, c := range []struct {
		key  string
		want int
	}{{wrong, http.StatusUnauthorized}, {key, http.StatusOK}} {
		req, err := http.NewRequest(http.MethodGet, base+"/props", nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		resp, err := probeClient.Do(req)
		if err != nil {
			return err
		}
		io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
		resp.Body.Close()
		if resp.StatusCode != c.want {
			return fmt.Errorf("GET /props answered %d, want %d", resp.StatusCode, c.want)
		}
	}
	return nil
}

func newKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// writeKey creates a fresh key in a 0600 file inside a private 0700
// directory. cleanup removes the file, and the directory when it is a
// temporary one.
func (s *Supervisor) writeKey() (key, path string, cleanup func(), err error) {
	key, err = newKey()
	if err != nil {
		return "", "", nil, err
	}
	dir, temp := s.KeyDir, s.KeyDir == ""
	if temp {
		if dir, err = os.MkdirTemp("", "luna-llama-key-"); err != nil { // mode 0700
			return "", "", nil, err
		}
	} else if err = privateDir(dir); err != nil {
		return "", "", nil, err
	}
	removeDir := func() {
		if temp {
			os.RemoveAll(dir)
		}
	}
	f, err := os.CreateTemp(dir, "key-") // mode 0600
	if err != nil {
		removeDir()
		return "", "", nil, err
	}
	path = f.Name()
	cleanup = func() {
		os.Remove(path)
		removeDir()
	}
	_, err = f.WriteString(key + "\n")
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		cleanup()
		return "", "", nil, err
	}
	return key, path, cleanup, nil
}

// privateDir creates dir 0700 if absent and checks that it is a real
// directory owned by this UID with no group or world access.
func privateDir(dir string) error {
	if err := os.Mkdir(dir, 0o700); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	fi, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	if !fi.IsDir() || fi.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("key directory %s must be a 0700 directory, not a symlink", dir)
	}
	if st, ok := fi.Sys().(*syscall.Stat_t); !ok || int(st.Uid) != os.Getuid() {
		return fmt.Errorf("key directory %s is not owned by this user", dir)
	}
	return nil
}

// freePort reserves a loopback port and releases it for the child.
func freePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
