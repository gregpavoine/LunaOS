package inference

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

// fakeLlamaSource is a stand-in for llama-server. It parses -m, --host, --port
// and -ngl, then serves GET /health on <host>:<port>. Behaviour switches come
// from environment variables:
//
//	FAKE_EXIT_CODE=n         exit immediately with code n
//	FAKE_NO_HEALTH=1         listen but never answer /health with 200 (503)
//	FAKE_IGNORE_TERM=1       ignore SIGTERM
//	FAKE_HEALTH_DELAY_MS=n   start listening only after n ms
//	FAKE_CRASH_AFTER_MS=n    exit(3) n ms after listening
//	FAKE_PIDFILE=path        append "<pid>\n" to path at startup
//
// Extend the mux in main (e.g. POST /completion) for further tests.
const fakeLlamaSource = `package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func envInt(k string) (int, bool) {
	v := os.Getenv(k)
	if v == "" {
		return 0, false
	}
	n, _ := strconv.Atoi(v)
	return n, true
}

func main() {
	flag.String("m", "", "model")
	host := flag.String("host", "127.0.0.1", "host")
	port := flag.Int("port", 0, "port")
	flag.Int("ngl", 0, "gpu layers")
	flag.Parse()

	if p := os.Getenv("FAKE_PIDFILE"); p != "" {
		f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err == nil {
			fmt.Fprintf(f, "%d\n", os.Getpid())
			f.Close()
		}
	}
	if os.Getenv("FAKE_IGNORE_TERM") != "" {
		signal.Ignore(syscall.SIGTERM)
	}
	if n, ok := envInt("FAKE_EXIT_CODE"); ok {
		os.Exit(n)
	}
	if n, ok := envInt("FAKE_HEALTH_DELAY_MS"); ok {
		time.Sleep(time.Duration(n) * time.Millisecond)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("FAKE_NO_HEALTH") != "" {
			http.Error(w, "loading model", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("{\"status\":\"ok\"}"))
	})
	ln, err := net.Listen("tcp", net.JoinHostPort(*host, strconv.Itoa(*port)))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if n, ok := envInt("FAKE_CRASH_AFTER_MS"); ok {
		go func() {
			time.Sleep(time.Duration(n) * time.Millisecond)
			os.Exit(3)
		}()
	}
	http.Serve(ln, mux)
}
`

var (
	fakeLlamaOnce sync.Once
	fakeLlamaDir  string
	fakeLlamaBin  string
	fakeLlamaErr  error
)

// buildFakeLlama compiles the fake once per test binary and returns its path.
func buildFakeLlama(t *testing.T) string {
	t.Helper()
	fakeLlamaOnce.Do(func() {
		fakeLlamaDir, fakeLlamaErr = os.MkdirTemp("", "fakellama")
		if fakeLlamaErr != nil {
			return
		}
		src := filepath.Join(fakeLlamaDir, "main.go")
		if fakeLlamaErr = os.WriteFile(src, []byte(fakeLlamaSource), 0o644); fakeLlamaErr != nil {
			return
		}
		if fakeLlamaErr = os.WriteFile(filepath.Join(fakeLlamaDir, "go.mod"), []byte("module fakellama\n\ngo 1.24\n"), 0o644); fakeLlamaErr != nil {
			return
		}
		fakeLlamaBin = filepath.Join(fakeLlamaDir, "llama-server")
		cmd := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-o", fakeLlamaBin, ".")
		cmd.Dir = fakeLlamaDir
		cmd.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOFLAGS=-buildvcs=false")
		if out, err := cmd.CombinedOutput(); err != nil {
			fakeLlamaErr = &buildError{err, string(out)}
		}
	})
	if fakeLlamaErr != nil {
		t.Fatalf("build fake llama-server: %v", fakeLlamaErr)
	}
	return fakeLlamaBin
}

type buildError struct {
	err error
	out string
}

func (e *buildError) Error() string { return e.err.Error() + "\n" + e.out }

// TestMain removes the cached fake binary after the run.
func TestMain(m *testing.M) {
	code := m.Run()
	if fakeLlamaDir != "" {
		os.RemoveAll(fakeLlamaDir)
	}
	os.Exit(code)
}
