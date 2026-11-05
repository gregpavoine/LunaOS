package discovery

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeKey(t *testing.T, path string, mode os.FileMode, n int) {
	t.Helper()
	key := make([]byte, n)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, key, mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
}

func newTestStore(t *testing.T) (*Store, string) {
	t.Helper()
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "key")
	writeKey(t, keyPath, 0600, 32)
	s, err := OpenStore(filepath.Join(dir, "d.db"), filepath.Join(dir, "blobs"), keyPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s, dir
}

func shaHex(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

func blobFiles(t *testing.T, s *Store) []string {
	t.Helper()
	entries, err := os.ReadDir(s.blobDir)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, e := range entries {
		out = append(out, filepath.Join(s.blobDir, e.Name()))
	}
	return out
}

func TestPutEncryptsAtRest(t *testing.T) {
	s, dir := newTestStore(t)
	plain := append(bytes.Repeat([]byte("PLAINTEXT-MARKER-"), 10000), []byte("tail")...) // spans several chunks
	sha := shaHex(plain)
	it := Item{Provider: "civitai", ExternalID: "1", SourceURL: "https://example.test/secret-url", Creator: "secret-creator", Tags: []string{"secret-tag"}, ContentLevel: "adult", License: "l"}
	stored, err := s.Put(context.Background(), it, bytes.NewReader(plain))
	if err != nil || !stored {
		t.Fatalf("stored=%v err=%v", stored, err)
	}
	rc, err := s.Open(context.Background(), sha)
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(rc)
	rc.Close()
	if err != nil || !bytes.Equal(got, plain) {
		t.Fatalf("round trip failed: %v", err)
	}
	needles := [][]byte{[]byte("PLAINTEXT-MARKER-"), []byte("secret-url"), []byte("secret-creator"), []byte("secret-tag"), []byte(sha)}
	filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Base(p) == "key" {
			return nil
		}
		data, _ := os.ReadFile(p)
		for _, n := range needles {
			if bytes.Contains(data, n) {
				t.Errorf("%s contains a plaintext needle", filepath.Base(p))
			}
		}
		return nil
	})
	if strings.Contains(filepath.Base(blobFiles(t, s)[0]), sha) {
		t.Error("blob name exposes the media hash")
	}
}

func TestPutDeduplicatesBySHA256(t *testing.T) {
	s, _ := newTestStore(t)
	ctx := context.Background()
	media := []byte("same media")
	stored, err := s.Put(ctx, Item{Provider: "p", ExternalID: "a"}, bytes.NewReader(media))
	if err != nil || !stored {
		t.Fatal(stored, err)
	}
	stored, err = s.Put(ctx, Item{Provider: "p", ExternalID: "b"}, bytes.NewReader(media))
	if err != nil || stored {
		t.Fatalf("same sha must not store again: %v %v", stored, err)
	}
	if len(blobFiles(t, s)) != 1 {
		t.Fatalf("want one blob, got %d", len(blobFiles(t, s)))
	}
	var rows int
	s.db.QueryRow(`SELECT COUNT(*) FROM items`).Scan(&rows)
	if rows != 2 {
		t.Fatalf("second external id must get its own row, got %d", rows)
	}
	if stored, _ = s.Put(ctx, Item{Provider: "p", ExternalID: "a"}, bytes.NewReader([]byte("other"))); stored {
		t.Fatal("re-put of same key must not store")
	}
	if _, err = s.Put(ctx, Item{Provider: "p", ExternalID: "c", SHA256: shaHex([]byte("nope"))}, bytes.NewReader(media)); err == nil {
		t.Fatal("declared hash mismatch must fail")
	}
	if len(blobFiles(t, s)) != 1 {
		t.Fatal("mismatch left files behind")
	}
}

func TestPutMetadataOnlyCreatesNoBlob(t *testing.T) {
	s, _ := newTestStore(t)
	ctx := context.Background()
	stored, err := s.Put(ctx, Item{Provider: "reddit", ExternalID: "x", SourceURL: "u"}, nil)
	if err != nil || !stored {
		t.Fatal(stored, err)
	}
	stored, err = s.Put(ctx, Item{Provider: "reddit", ExternalID: "x"}, nil)
	if err != nil || stored {
		t.Fatal(stored, err)
	}
	if len(blobFiles(t, s)) != 0 {
		t.Fatal("metadata-only must not create a blob")
	}
}

func TestDeleteRemovesBlobAndRow(t *testing.T) {
	s, _ := newTestStore(t)
	ctx := context.Background()
	media := []byte("media")
	s.Put(ctx, Item{Provider: "p", ExternalID: "a"}, bytes.NewReader(media))
	s.Put(ctx, Item{Provider: "p", ExternalID: "b"}, bytes.NewReader(media))
	if err := s.Delete(ctx, "p", "a"); err != nil {
		t.Fatal(err)
	}
	if len(blobFiles(t, s)) != 1 {
		t.Fatal("blob still referenced must stay")
	}
	if err := s.Delete(ctx, "p", "b"); err != nil {
		t.Fatal(err)
	}
	if len(blobFiles(t, s)) != 0 {
		t.Fatal("unreferenced blob must go")
	}
	var rows int
	s.db.QueryRow(`SELECT COUNT(*) FROM items`).Scan(&rows)
	if rows != 0 {
		t.Fatal("rows remain")
	}
	if err := s.Delete(ctx, "p", "missing"); err != nil {
		t.Fatalf("missing delete must not error: %v", err)
	}
}

func TestOpenWithWrongKeyFails(t *testing.T) {
	s, dir := newTestStore(t)
	ctx := context.Background()
	media := bytes.Repeat([]byte("m"), 70000)
	sha := shaHex(media)
	s.Put(ctx, Item{Provider: "p", ExternalID: "a"}, bytes.NewReader(media))
	keyB := filepath.Join(dir, "keyB")
	writeKey(t, keyB, 0600, 32)
	other, err := OpenStore(filepath.Join(dir, "o.db"), filepath.Join(dir, "oblobs"), keyB)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	// Plant the blob under the name the wrong key would look up.
	data, _ := os.ReadFile(blobFiles(t, s)[0])
	os.WriteFile(filepath.Join(other.blobDir, other.blobName(sha)), data, 0600)
	if rc, err := other.Open(ctx, sha); err == nil {
		rc.Close()
		t.Fatal("wrong key must fail")
	}
	// Tampering: flip a late byte, reading must error.
	path := blobFiles(t, s)[0]
	data[len(data)-5] ^= 0xff
	os.WriteFile(path, data, 0600)
	rc, err := s.Open(ctx, sha)
	if err == nil {
		_, err = io.ReadAll(rc)
		rc.Close()
	}
	if err == nil {
		t.Fatal("tampered blob must fail")
	}
	// Truncation of whole last chunk must also fail.
	data2 := data[:8+chunkSize+16]
	os.WriteFile(path, data2, 0600)
	rc, err = s.Open(ctx, sha)
	if err == nil {
		_, err = io.ReadAll(rc)
		rc.Close()
	}
	if err == nil {
		t.Fatal("truncated blob must fail")
	}
}

func TestRejectsWeakKeyFile(t *testing.T) {
	dir := t.TempDir()
	for name, c := range map[string]struct {
		mode os.FileMode
		n    int
	}{"wide": {0644, 32}, "group": {0640, 32}, "short": {0600, 16}, "long": {0600, 64}} {
		p := filepath.Join(dir, name)
		writeKey(t, p, c.mode, c.n)
		if _, err := OpenStore(filepath.Join(dir, name+".db"), filepath.Join(dir, name+"-b"), p); err == nil {
			t.Errorf("%s key must be refused", name)
		}
	}
}

func TestCursorRoundTrip(t *testing.T) {
	s, _ := newTestStore(t)
	ctx := context.Background()
	if c, err := s.Cursor(ctx, "p", "s"); err != nil || c != "" {
		t.Fatal(c, err)
	}
	s.SaveCursor(ctx, "p", "s", "one")
	s.SaveCursor(ctx, "p", "s", "two")
	if c, _ := s.Cursor(ctx, "p", "s"); c != "two" {
		t.Fatal(c)
	}
	if c, _ := s.Cursor(ctx, "p", "other"); c != "" {
		t.Fatal(c)
	}
}

func TestOpenSwappedBlobsFail(t *testing.T) {
	s, _ := newTestStore(t)
	ctx := context.Background()
	a, b := []byte("media A"), []byte("media B")
	s.Put(ctx, Item{Provider: "p", ExternalID: "a"}, bytes.NewReader(a))
	s.Put(ctx, Item{Provider: "p", ExternalID: "b"}, bytes.NewReader(b))
	pa := filepath.Join(s.blobDir, s.blobName(shaHex(a)))
	pb := filepath.Join(s.blobDir, s.blobName(shaHex(b)))
	da, _ := os.ReadFile(pa)
	db, _ := os.ReadFile(pb)
	os.WriteFile(pa, db, 0600)
	os.WriteFile(pb, da, 0600)
	for _, m := range [][]byte{a, b} {
		rc, err := s.Open(ctx, shaHex(m))
		if err == nil {
			_, err = io.ReadAll(rc)
			rc.Close()
		}
		if err == nil {
			t.Fatal("swapped blob must fail to open")
		}
	}
}

func TestRejectsWritableDatabaseDirectory(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "key")
	writeKey(t, keyPath, 0600, 32)
	open := filepath.Join(dir, "open")
	os.Mkdir(open, 0777)
	os.Chmod(open, 0777)
	if _, err := OpenStore(filepath.Join(open, "d.db"), filepath.Join(dir, "blobs"), keyPath); err == nil {
		t.Fatal("group/other-writable db directory must be refused")
	}
}
