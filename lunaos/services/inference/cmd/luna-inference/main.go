// luna-inference serves local generation to the Core over a peer-checked Unix
// socket. Only the fetch command reaches the network.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"lunaos.local/lunaos/services/inference"
	"lunaos.local/lunaos/services/internal/peer"
)

const usage = `usage:
  luna-inference serve --socket <path> --model-dir <dir> [--model <id>] --manifest <path> --llama-server <path> (--allowed-uid <uid> | --allowed-user <name>) [--gpu-layers N]
  luna-inference fetch --manifest <path> --model-dir <dir> <id>`

// shutdownGrace bounds the wait for requests in flight on SIGTERM/SIGINT.
const shutdownGrace = 30 * time.Second

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "luna-inference:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New(usage)
	}
	switch args[0] {
	case "serve":
		socket, cfg, err := parseServe(args[1:])
		if err != nil {
			return err
		}
		return serve(socket, cfg)
	case "fetch":
		manifest, dir, id, err := parseFetch(args[1:])
		if err != nil {
			return err
		}
		return fetch(manifest, dir, id)
	default:
		return fmt.Errorf("unknown command %q\n%s", args[0], usage)
	}
}

func newFlags(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard) // errors are returned with the usage text
	return fs
}

func parseServe(args []string) (string, inference.Config, error) {
	fs := newFlags("serve")
	socket := fs.String("socket", "", "Unix socket path")
	modelDir := fs.String("model-dir", "", "directory of <id>.gguf files")
	model := fs.String("model", "", "manifest id of the model to serve")
	manifest := fs.String("manifest", "", "model manifest (JSON)")
	llama := fs.String("llama-server", "", "absolute path of llama-server")
	uidFlag := fs.String("allowed-uid", "", "the only peer UID served")
	userFlag := fs.String("allowed-user", "", "account whose UID is the only peer served")
	gpuLayers := fs.Int("gpu-layers", 0, "layers offloaded to the GPU")
	if err := fs.Parse(args); err != nil {
		return "", inference.Config{}, fmt.Errorf("%v\n%s", err, usage)
	}
	if fs.NArg() != 0 {
		return "", inference.Config{}, fmt.Errorf("unexpected arguments %q\n%s", fs.Args(), usage)
	}
	if *socket == "" || *modelDir == "" || *manifest == "" || *llama == "" {
		return "", inference.Config{}, fmt.Errorf("--socket, --model-dir, --manifest and --llama-server are required\n%s", usage)
	}
	uid, err := allowedUID(*uidFlag, *userFlag)
	if err != nil {
		return "", inference.Config{}, err
	}
	return *socket, inference.Config{
		Manifest:    *manifest,
		ModelDir:    *modelDir,
		ModelID:     *model,
		LlamaServer: *llama,
		GPULayers:   *gpuLayers,
		AllowedUID:  uid,
		// llama-server's per-spawn API key lives next to the socket, in the
		// service's private runtime directory.
		KeyDir: filepath.Join(filepath.Dir(*socket), "llama-key"),
	}, nil
}

// allowedUID resolves exactly one of --allowed-uid and --allowed-user.
func allowedUID(uidFlag, userFlag string) (int, error) {
	if (uidFlag == "") == (userFlag == "") {
		return -1, errors.New("exactly one of --allowed-uid and --allowed-user is required")
	}
	raw := uidFlag
	if userFlag != "" {
		account, err := user.Lookup(userFlag)
		if err != nil {
			return -1, fmt.Errorf("lookup %s: %w", userFlag, err)
		}
		raw = account.Uid
	}
	uid, err := strconv.Atoi(raw)
	if err != nil || uid < 0 {
		return -1, fmt.Errorf("invalid allowed UID %q", raw)
	}
	return uid, nil
}

func parseFetch(args []string) (manifest, dir, id string, err error) {
	fs := newFlags("fetch")
	m := fs.String("manifest", "", "model manifest (JSON)")
	d := fs.String("model-dir", "", "directory of <id>.gguf files")
	if err := fs.Parse(args); err != nil {
		return "", "", "", fmt.Errorf("%v\n%s", err, usage)
	}
	if *m == "" || *d == "" || fs.NArg() != 1 {
		return "", "", "", fmt.Errorf("fetch needs --manifest, --model-dir and one model id\n%s", usage)
	}
	return *m, *d, fs.Arg(0), nil
}

// listenSocket creates the socket with mode 0660. It refuses to replace an
// existing path, and the umask closes any window with a wider mode.
func listenSocket(path string) (net.Listener, error) {
	if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("socket path %s exists or cannot be inspected; refusing replacement", path)
	}
	old := syscall.Umask(0o077)
	l, err := net.Listen("unix", path)
	syscall.Umask(old)
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(path, 0o660); err != nil {
		l.Close()
		return nil, err
	}
	return l, nil
}

func serve(socket string, cfg inference.Config) error {
	if runtime.GOOS != "linux" {
		return errors.New("luna-inference requires Linux Unix-socket peer credentials")
	}
	srv, err := inference.NewServer(cfg)
	if err != nil {
		return err
	}
	defer srv.Close() // no llama-server outlives the service
	l, err := listenSocket(socket)
	if err != nil {
		return err
	}
	hs := &http.Server{
		Handler:           srv,
		ConnContext:       peer.ConnContext,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		// A generate may wait for the model start, then for the completion.
		WriteTimeout:   inference.DefaultStartWait + inference.DefaultGenerateTimeout + 30*time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 8192,
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	served := make(chan error, 1)
	go func() { served <- hs.Serve(l) }()
	select {
	case err := <-served:
		return err
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownGrace)
	defer cancel()
	if err := hs.Shutdown(shutdownCtx); err != nil {
		hs.Close()
	}
	return nil
}

func fetch(manifest, dir, id string) error {
	models, err := inference.LoadManifest(manifest)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	for _, m := range models {
		if m.ID == id {
			path, err := inference.Fetch(ctx, m, dir, nil)
			if err != nil {
				return err
			}
			fmt.Println(path)
			return nil
		}
	}
	return fmt.Errorf("model %q is not in the manifest", id)
}
