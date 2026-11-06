package inference

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

// fakeLlamaSource is a stand-in for llama-server. It parses -m, --host, --port,
// -ngl, -c and --api-key-file, then serves GET /health on <host>:<port>. Like
// llama-server, /health is public and GET /props and POST /completion need
// "Authorization: Bearer <key>" (401 otherwise). Stricter than llama-server, a
// fake started without --api-key-file accepts no key at all, so a caller that
// forgets the key fails. Behaviour switches come from environment variables:
//
//	FAKE_EXIT_CODE=n         exit immediately with code n
//	FAKE_NO_HEALTH=1         listen but never answer /health with 200 (503)
//	FAKE_IGNORE_TERM=1       ignore SIGTERM
//	FAKE_HEALTH_DELAY_MS=n   start listening only after n ms
//	FAKE_CRASH_AFTER_MS=n    exit(3) n ms after listening
//	FAKE_PIDFILE=path        append "<pid>\n" to path at startup
//	FAKE_ARGV_FILE=path      append its argv, one line per spawn, to path
//	FAKE_ACCEPT_ANY_KEY=1    impostor: accept every key, or none
//
// POST /completion decodes {"prompt","n_predict"} strictly and answers
// {"content":"echo(<n_predict>):<prompt>"}. Its switches:
//
//	FAKE_COMPLETION_GATE=path    answer only once path exists (polled, 30 s cap)
//	FAKE_COMPLETION_STATUS=n     answer with HTTP status n and an empty body
//	FAKE_COMPLETION_BODY=raw     answer 200 with this raw body
const fakeLlamaSource = `package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
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
	flag.Int("c", 0, "context size")
	keyFile := flag.String("api-key-file", "", "API key file")
	flag.Parse()

	if p := os.Getenv("FAKE_ARGV_FILE"); p != "" {
		f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err == nil {
			fmt.Fprintln(f, strings.Join(os.Args, " "))
			f.Close()
		}
	}
	var key string
	if *keyFile != "" {
		b, err := os.ReadFile(*keyFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(4)
		}
		key = strings.TrimSpace(string(b))
	}
	authorized := func(w http.ResponseWriter, r *http.Request) bool {
		if os.Getenv("FAKE_ACCEPT_ANY_KEY") != "" {
			return true
		}
		if key == "" || r.Header.Get("Authorization") != "Bearer "+key {
			http.Error(w, "{\"error\":\"Invalid API Key\"}", http.StatusUnauthorized)
			return false
		}
		return true
	}

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
	mux.HandleFunc("GET /props", func(w http.ResponseWriter, r *http.Request) {
		if authorized(w, r) {
			w.Write([]byte("{}"))
		}
	})
	mux.HandleFunc("POST /completion", func(w http.ResponseWriter, r *http.Request) {
		if !authorized(w, r) {
			return
		}
		var req struct {
			Prompt   string ` + "`json:\"prompt\"`" + `
			NPredict int    ` + "`json:\"n_predict\"`" + `
		}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil || req.NPredict <= 0 {
			http.Error(w, "bad completion request", http.StatusBadRequest)
			return
		}
		if g := os.Getenv("FAKE_COMPLETION_GATE"); g != "" {
			for i := 0; i < 3000; i++ {
				if _, err := os.Stat(g); err == nil {
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
		if n, ok := envInt("FAKE_COMPLETION_STATUS"); ok {
			w.WriteHeader(n)
			return
		}
		if b, ok := os.LookupEnv("FAKE_COMPLETION_BODY"); ok {
			w.Write([]byte(b))
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"content": fmt.Sprintf("echo(%d):%s", req.NPredict, req.Prompt),
			"stop":    true,
		})
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
