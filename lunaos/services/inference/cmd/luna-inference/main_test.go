package main

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"lunaos.local/lunaos/services/inference"
)

func serveArgs(extra ...string) []string {
	return append([]string{"--socket", "/run/x.sock", "--model-dir", "/m", "--manifest", "/m.json", "--llama-server", "/usr/bin/llama-server"}, extra...)
}

func TestParseServeRequiresExactlyOneAllowedPeer(t *testing.T) {
	for name, args := range map[string][]string{
		"neither":      serveArgs(),
		"both":         serveArgs("--allowed-uid", "1000", "--allowed-user", "root"),
		"negative uid": serveArgs("--allowed-uid", "-1"),
		"non-numeric":  serveArgs("--allowed-uid", "core"),
		"unknown user": serveArgs("--allowed-user", "no-such-luna-user"),
		"no socket":    {"--model-dir", "/m", "--manifest", "/m.json", "--llama-server", "/l", "--allowed-uid", "1"},
		"extra arg":    serveArgs("--allowed-uid", "1", "stray"),
	} {
		if _, _, err := parseServe(args); err == nil {
			t.Errorf("%s: parseServe accepted %q", name, args)
		}
	}

	socket, cfg, err := parseServe(serveArgs("--allowed-uid", "1000", "--model", "m1", "--gpu-layers", "99"))
	if err != nil {
		t.Fatal(err)
	}
	if socket != "/run/x.sock" || cfg.AllowedUID != 1000 || cfg.ModelID != "m1" || cfg.GPULayers != 99 ||
		cfg.ModelDir != "/m" || cfg.Manifest != "/m.json" || cfg.LlamaServer != "/usr/bin/llama-server" || cfg.KeyDir != "/run/llama-key" {
		t.Fatalf("parsed %q %+v", socket, cfg)
	}

	me, err := user.Current()
	if err != nil {
		t.Skip(err)
	}
	_, cfg, err = parseServe(serveArgs("--allowed-user", me.Username))
	if err != nil {
		t.Fatal(err)
	}
	if strconv.Itoa(cfg.AllowedUID) != me.Uid {
		t.Fatalf("--allowed-user %s resolved to %d, want %s", me.Username, cfg.AllowedUID, me.Uid)
	}
}

func TestParseFetchNeedsOneID(t *testing.T) {
	for _, args := range [][]string{
		{"--manifest", "/m.json", "--model-dir", "/m"},
		{"--manifest", "/m.json", "--model-dir", "/m", "a", "b"},
		{"--model-dir", "/m", "a"},
	} {
		if _, _, _, err := parseFetch(args); err == nil {
			t.Errorf("parseFetch accepted %q", args)
		}
	}
	if m, d, id, err := parseFetch([]string{"--manifest", "/m.json", "--model-dir", "/m", "m1"}); err != nil || m != "/m.json" || d != "/m" || id != "m1" {
		t.Fatalf("parseFetch = %q %q %q %v", m, d, id, err)
	}
}

func TestListenSocketMode0660AndNoReplacement(t *testing.T) {
	dir, err := os.MkdirTemp("/tmp", "infsock") // short path for macOS
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "s.sock")
	l, err := listenSocket(path)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	fi, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode()&os.ModeSocket == 0 || fi.Mode().Perm() != 0o660 {
		t.Fatalf("socket mode = %v, want socket 0660", fi.Mode())
	}
	if _, err := listenSocket(path); err == nil || !strings.Contains(err.Error(), "refusing replacement") {
		t.Fatalf("second listen = %v, want refusal", err)
	}
	stale := filepath.Join(dir, "stale")
	if err := os.WriteFile(stale, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := listenSocket(stale); err == nil {
		t.Fatal("listenSocket replaced an existing file")
	}
}

func TestRunRejectsBadCommandsAndNonLinux(t *testing.T) {
	for _, args := range [][]string{nil, {"serve"}, {"bogus"}} {
		if err := run(args); err == nil {
			t.Errorf("run(%q) succeeded", args)
		}
	}
	if runtime.GOOS != "linux" {
		err := serve(filepath.Join(t.TempDir(), "s.sock"), inference.Config{})
		if err == nil || !strings.Contains(err.Error(), "requires Linux") {
			t.Fatalf("serve off Linux = %v, want a refusal", err)
		}
	}
}

// TestUnitExecStartParses catches drift between the systemd unit and the CLI:
// a flag renamed in parseServe while the unit still passes the old one.
func TestUnitExecStartParses(t *testing.T) {
	unit, err := os.ReadFile("../../systemd/luna-inference.service")
	if err != nil {
		t.Fatal(err)
	}
	var args []string
	for _, line := range strings.Split(string(unit), "\n") {
		if rest, ok := strings.CutPrefix(line, "ExecStart="); ok {
			fields := strings.Fields(strings.ReplaceAll(rest, "${LUNA_INFERENCE_MODEL}", ""))
			for i, f := range fields {
				if f == "serve" {
					args = fields[i+1:]
				}
			}
		}
	}
	if len(args) == 0 {
		t.Fatal("no ExecStart serve arguments in the unit")
	}
	if _, err := user.Lookup("lunaos"); err != nil {
		// --allowed-user=lunaos cannot resolve here; swap in the current user
		// so the rest of the arguments are still checked.
		current, uerr := user.Current()
		if uerr != nil {
			t.Fatal(uerr)
		}
		for i, a := range args {
			if a == "--allowed-user=lunaos" {
				args[i] = "--allowed-user=" + current.Username
			}
		}
	}
	socket, cfg, err := parseServe(args)
	if err != nil {
		t.Fatalf("parseServe rejected the unit's arguments %q: %v", args, err)
	}
	if socket != "/run/luna-inference/inference.sock" {
		t.Errorf("socket = %q", socket)
	}
	if cfg.KeyDir != "/run/luna-inference/llama-key" {
		t.Errorf("KeyDir = %q", cfg.KeyDir)
	}
	if cfg.ModelID != "" {
		t.Errorf("empty model expected, got %q", cfg.ModelID)
	}
}
