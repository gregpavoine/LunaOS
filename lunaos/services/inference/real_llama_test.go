package inference

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestRealLlamaServerGenerates runs the real llama-server on a manifest model.
// It is skipped unless both are given:
//
//	LUNA_TEST_LLAMA_SERVER  absolute path of llama-server
//	LUNA_TEST_MODEL         path of a downloaded <id>.gguf from models/manifest.json
func TestRealLlamaServerGenerates(t *testing.T) {
	binary, model := os.Getenv("LUNA_TEST_LLAMA_SERVER"), os.Getenv("LUNA_TEST_MODEL")
	if binary == "" || model == "" {
		t.Skip("set LUNA_TEST_LLAMA_SERVER and LUNA_TEST_MODEL to run against a real llama-server")
	}
	manifest := filepath.Join("..", "..", "models", "manifest.json")
	models, err := LoadManifest(manifest)
	if err != nil {
		t.Fatal(err)
	}
	id := strings.TrimSuffix(filepath.Base(model), ".gguf")
	var want string
	for _, m := range models {
		if m.ID == id {
			want = m.SHA256
		}
	}
	if want == "" {
		t.Fatalf("%s is not named <id>.gguf for a manifest id", model)
	}
	// Check first: NewServer deletes a file with a wrong hash.
	if got, err := fileHash(model); err != nil || got != want {
		t.Fatalf("%s does not match the manifest sha256 (%v); not touching it", model, err)
	}

	srv, err := NewServer(Config{
		Manifest:    manifest,
		ModelDir:    filepath.Dir(model),
		ModelID:     id,
		LlamaServer: binary,
		AllowedUID:  coreUID,
		StartWait:   5 * time.Minute,
		Log:         io.Discard,
	})
	if err != nil {
		t.Fatal(err)
	}
	if srv.sup == nil {
		t.Fatal("model was not accepted")
	}
	srv.sup.StartTimeout = 5 * time.Minute
	t.Cleanup(func() { srv.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	res, err := NewClient(listenUnix(t, srv, coreUID)).Generate(ctx, "Say OK", 8)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(res.Text) == "" || !res.Untrusted || res.Model != id {
		t.Fatalf("result = %+v", res)
	}
	t.Logf("real llama-server answered %q", res.Text)
}
