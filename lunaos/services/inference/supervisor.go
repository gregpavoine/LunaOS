package inference

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
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
	Binary, ModelPath string
	GPULayers         int
	StartTimeout      time.Duration // 0 -> 60 s
	StopGrace         time.Duration // SIGTERM to SIGKILL delay; 0 -> 10 s

	mu        sync.Mutex // never held while waiting on the child or the network
	state     State
	run       *run
	baseURL   string
	failures  []time.Time
	startDone chan struct{} // non-nil while a start is in progress
	startErr  error
}

// run is one child process. stopped is guarded by Supervisor.mu.
type run struct {
	cmd     *exec.Cmd
	done    chan struct{} // closed after the child is reaped
	stopped bool          // Stop or a cancelled Start asked it to end
}

var healthClient = &http.Client{Timeout: time.Second}

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

// Start launches llama-server if needed and returns its base URL once /health
// answers 200. Concurrent callers share one start.
func (s *Supervisor) Start(ctx context.Context) (string, error) {
	s.mu.Lock()
	switch s.state {
	case StateReady:
		u := s.baseURL
		s.mu.Unlock()
		return u, nil
	case StateDegraded:
		s.mu.Unlock()
		return "", errors.New("llama-server is degraded")
	case StateStarting:
		done := s.startDone
		s.mu.Unlock()
		select {
		case <-done:
		case <-ctx.Done():
			return "", ctx.Err()
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.startErr != nil {
			return "", s.startErr
		}
		return s.baseURL, nil
	}
	s.state = StateStarting
	s.startDone = make(chan struct{})
	done := s.startDone
	s.mu.Unlock()

	url, err := s.launch(ctx)

	s.mu.Lock()
	s.startErr = err
	if err == nil {
		s.baseURL = url
	}
	s.startDone = nil
	close(done)
	s.mu.Unlock()
	return url, err
}

// launch spawns the child and waits for readiness. It leaves s.state at Ready,
// Stopped or Degraded.
func (s *Supervisor) launch(ctx context.Context) (string, error) {
	fail := func(err error, count bool) (string, error) {
		s.mu.Lock()
		if count {
			s.recordFailure()
		}
		if s.state != StateDegraded {
			s.state = StateStopped
		}
		s.mu.Unlock()
		return "", err
	}

	if fi, err := os.Stat(s.Binary); err != nil || fi.IsDir() || fi.Mode().Perm()&0o111 == 0 {
		s.mu.Lock()
		s.state = StateDegraded
		s.mu.Unlock()
		return "", fmt.Errorf("llama-server binary %q missing or not executable", s.Binary)
	}
	port, err := freePort()
	if err != nil {
		return fail(err, false)
	}
	cmd := exec.Command(s.Binary, "-m", s.ModelPath, "--host", "127.0.0.1",
		"--port", strconv.Itoa(port), "-ngl", strconv.Itoa(s.GPULayers))
	if err := cmd.Start(); err != nil {
		s.mu.Lock()
		s.state = StateDegraded
		s.mu.Unlock()
		return "", fmt.Errorf("start llama-server: %w", err)
	}
	r := &run{cmd: cmd, done: make(chan struct{})}
	s.mu.Lock()
	s.run = r
	s.mu.Unlock()
	go s.reap(r)

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
			stopped := r.stopped
			s.mu.Unlock()
			if stopped {
				return fail(errors.New("llama-server stopped during start"), false)
			}
			return fail(errors.New("llama-server exited before becoming healthy"), true)
		case <-ctx.Done():
			s.terminate(r, true)
			return fail(ctx.Err(), false)
		case <-deadline.C:
			s.terminate(r, false)
			return fail(fmt.Errorf("llama-server not healthy after %v", timeout), true)
		case <-tick.C:
			if !healthy(base) {
				continue
			}
			s.mu.Lock()
			select {
			case <-r.done: // died right after answering; next loop turn handles it
				s.mu.Unlock()
				continue
			default:
			}
			if r.stopped {
				s.mu.Unlock()
				continue
			}
			s.state = StateReady
			s.mu.Unlock()
			return base, nil
		}
	}
}

// reap waits for the child and applies an unexpected exit to the state.
func (s *Supervisor) reap(r *run) {
	r.cmd.Wait()
	s.mu.Lock()
	if s.run == r {
		s.run = nil
		// During Starting the starter observes done and accounts for it.
		if s.state == StateReady && !r.stopped {
			s.state = StateStopped
			s.recordFailure()
		}
	}
	close(r.done)
	s.mu.Unlock()
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
	grace := s.StopGrace
	if grace <= 0 {
		grace = defaultStopGrace
	}
	t := time.NewTimer(grace)
	defer t.Stop()
	select {
	case <-r.done:
	case <-t.C:
		r.cmd.Process.Kill()
		<-r.done
	}
}

// Stop terminates the child (SIGTERM, then SIGKILL after StopGrace). It is a
// no-op when nothing is running and also ends a start in progress.
func (s *Supervisor) Stop() error {
	s.mu.Lock()
	r := s.run
	if r != nil && s.state == StateReady {
		s.state = StateStopped
	}
	s.mu.Unlock()
	if r == nil {
		return nil
	}
	s.terminate(r, true)
	return nil
}

func healthy(base string) bool {
	resp, err := healthClient.Get(base + "/health")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
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
