package inference

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

func newSup(t *testing.T) *Supervisor {
	t.Helper()
	s := &Supervisor{
		Binary:       buildFakeLlama(t),
		ModelPath:    "model.gguf",
		GPULayers:    1,
		StartTimeout: 5 * time.Second,
		StopGrace:    time.Second,
	}
	t.Cleanup(func() { s.Stop() })
	return s
}

func readPIDs(t *testing.T, path string) []int {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var pids []int
	for _, f := range strings.Fields(string(b)) {
		n, _ := strconv.Atoi(f)
		pids = append(pids, n)
	}
	return pids
}

func alive(pid int) bool { return syscall.Kill(pid, 0) == nil }

func TestStartWaitsForHealth(t *testing.T) {
	t.Setenv("FAKE_HEALTH_DELAY_MS", "400")
	s := newSup(t)
	type res struct {
		url string
		err error
	}
	ch := make(chan res, 1)
	go func() {
		u, err := s.Start(context.Background())
		ch <- res{u, err}
	}()
	time.Sleep(100 * time.Millisecond)
	st := make(chan State, 1)
	go func() { st <- s.State() }()
	select {
	case got := <-st:
		if got != StateStarting {
			t.Fatalf("state during start = %v, want starting", got)
		}
	case <-time.After(time.Second):
		t.Fatal("State() blocked while a start is in progress")
	}
	r := <-ch
	if r.err != nil {
		t.Fatal(r.err)
	}
	if !strings.HasPrefix(r.url, "http://127.0.0.1:") {
		t.Fatalf("baseURL = %q", r.url)
	}
	if s.State() != StateReady {
		t.Fatalf("state = %v, want ready", s.State())
	}
	resp, err := http.Get(r.url + "/health")
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("health: %v %v", resp, err)
	}
	resp.Body.Close()
	if err := s.Stop(); err != nil {
		t.Fatal(err)
	}
	if s.State() != StateStopped {
		t.Fatalf("state after stop = %v", s.State())
	}
	if _, err := http.Get(r.url + "/health"); err == nil {
		t.Fatal("process still serving after Stop")
	}
}

func TestStartTimesOut(t *testing.T) {
	t.Setenv("FAKE_NO_HEALTH", "1")
	pidfile := t.TempDir() + "/pids"
	t.Setenv("FAKE_PIDFILE", pidfile)
	s := newSup(t)
	s.StartTimeout = 500 * time.Millisecond
	if _, err := s.Start(context.Background()); err == nil {
		t.Fatal("expected timeout error")
	}
	if s.State() != StateStopped {
		t.Fatalf("state = %v, want stopped", s.State())
	}
	for _, pid := range readPIDs(t, pidfile) {
		if alive(pid) {
			t.Fatalf("child %d outlived failed start", pid)
		}
	}
}

func TestCrashLoopBecomesDegraded(t *testing.T) {
	t.Setenv("FAKE_EXIT_CODE", "1")
	pidfile := t.TempDir() + "/pids"
	t.Setenv("FAKE_PIDFILE", pidfile)
	s := newSup(t)
	for i := 0; i < maxFailures; i++ {
		if _, err := s.Start(context.Background()); err == nil {
			t.Fatal("expected start failure")
		}
	}
	if s.State() != StateDegraded {
		t.Fatalf("state = %v, want degraded", s.State())
	}
	spawned := len(readPIDs(t, pidfile))
	begin := time.Now()
	if _, err := s.Start(context.Background()); err == nil {
		t.Fatal("degraded Start must fail")
	}
	if time.Since(begin) > 200*time.Millisecond || len(readPIDs(t, pidfile)) != spawned {
		t.Fatal("degraded Start must return immediately without spawning")
	}
	s.Reset()
	if s.State() != StateStopped {
		t.Fatalf("state after reset = %v", s.State())
	}
	t.Setenv("FAKE_EXIT_CODE", "")
	if _, err := s.Start(context.Background()); err != nil {
		t.Fatalf("start after reset: %v", err)
	}
}

func TestStopKillsAfterGrace(t *testing.T) {
	t.Setenv("FAKE_IGNORE_TERM", "1")
	pidfile := t.TempDir() + "/pids"
	t.Setenv("FAKE_PIDFILE", pidfile)
	s := newSup(t)
	s.StopGrace = 300 * time.Millisecond
	if _, err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	begin := time.Now()
	if err := s.Stop(); err != nil {
		t.Fatal(err)
	}
	if d := time.Since(begin); d < 300*time.Millisecond || d > 5*time.Second {
		t.Fatalf("Stop took %v", d)
	}
	for _, pid := range readPIDs(t, pidfile) {
		if alive(pid) {
			t.Fatalf("child %d still alive", pid)
		}
	}
}

func TestMissingBinaryIsDegraded(t *testing.T) {
	for _, bin := range []string{t.TempDir() + "/nope", t.TempDir()} {
		s := &Supervisor{Binary: bin, ModelPath: "m.gguf", StartTimeout: time.Second}
		if _, err := s.Start(context.Background()); err == nil {
			t.Fatalf("%s: expected error", bin)
		}
		if s.State() != StateDegraded {
			t.Fatalf("%s: state = %v, want degraded", bin, s.State())
		}
	}
	plain := t.TempDir() + "/plain"
	os.WriteFile(plain, []byte("x"), 0o644)
	s := &Supervisor{Binary: plain, ModelPath: "m.gguf"}
	if _, err := s.Start(context.Background()); err == nil || s.State() != StateDegraded {
		t.Fatalf("non-executable binary: err=%v state=%v", err, s.State())
	}
}

func TestConcurrentStartsShareOneProcess(t *testing.T) {
	t.Setenv("FAKE_HEALTH_DELAY_MS", "300")
	pidfile := t.TempDir() + "/pids"
	t.Setenv("FAKE_PIDFILE", pidfile)
	s := newSup(t)
	var wg sync.WaitGroup
	urls := make([]string, 6)
	for i := range urls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			u, err := s.Start(context.Background())
			if err != nil {
				t.Error(err)
			}
			urls[i] = u
		}()
	}
	wg.Wait()
	for _, u := range urls {
		if u == "" || u != urls[0] {
			t.Fatalf("urls differ: %v", urls)
		}
	}
	if n := len(readPIDs(t, pidfile)); n != 1 {
		t.Fatalf("spawned %d processes, want 1", n)
	}
}

func TestStopDuringStart(t *testing.T) {
	t.Setenv("FAKE_NO_HEALTH", "1")
	pidfile := t.TempDir() + "/pids"
	t.Setenv("FAKE_PIDFILE", pidfile)
	s := newSup(t)
	errc := make(chan error, 1)
	go func() { _, err := s.Start(context.Background()); errc <- err }()
	time.Sleep(300 * time.Millisecond)
	if err := s.Stop(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-errc:
		if err == nil {
			t.Fatal("Start should fail after Stop")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Start did not return after Stop")
	}
	if s.State() != StateStopped {
		t.Fatalf("state = %v", s.State())
	}
	for _, pid := range readPIDs(t, pidfile) {
		if alive(pid) {
			t.Fatalf("child %d still alive", pid)
		}
	}
}

func TestUnexpectedExitMovesToStopped(t *testing.T) {
	t.Setenv("FAKE_CRASH_AFTER_MS", "300")
	s := newSup(t)
	if _, err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for s.State() == StateReady && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if s.State() != StateStopped {
		t.Fatalf("state = %v, want stopped", s.State())
	}
	if err := s.Stop(); err != nil {
		t.Fatalf("Stop after death: %v", err)
	}
}
