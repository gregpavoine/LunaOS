package intimacy

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var (
	pngBytes = append([]byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR"), bytes.Repeat([]byte{1}, 64)...)
	jpgBytes = append([]byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00}, bytes.Repeat([]byte{2}, 64)...)
	mp4Bytes = append([]byte{0x00, 0x00, 0x00, 0x18, 'f', 't', 'y', 'p', 'i', 's', 'o', 'm'}, bytes.Repeat([]byte{3}, 64)...)
)

func libraryVault(t *testing.T) (*Vault, string, string) {
	t.Helper()
	path := testVaultPath(t)
	v, err := Open(path, bytes.Repeat([]byte{5}, 32))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = v.Close() })
	consent, err := v.RecordOwnerConsent(context.Background(), "owner", []Scope{ScopeLibraryRead, ScopeLibraryWrite}, time.Now().Add(time.Hour), true, true)
	if err != nil {
		t.Fatal(err)
	}
	return v, path, consent.ID
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestLibraryIndexesByContentDedupsAndKeepsNamesSealed(t *testing.T) {
	v, vaultPath, consent := libraryVault(t)
	ctx := context.Background()
	root := t.TempDir()
	outside := t.TempDir()
	writeFile(t, filepath.Join(root, "lingerie-shoot-canary.png"), pngBytes)
	writeFile(t, filepath.Join(root, "nested", "copy-of-shoot.png"), pngBytes)
	writeFile(t, filepath.Join(root, "clip.mp4"), mp4Bytes)
	writeFile(t, filepath.Join(root, "notes.txt"), []byte("not media"))
	writeFile(t, filepath.Join(outside, "outside.jpg"), jpgBytes)
	if err := os.Symlink(filepath.Join(outside, "outside.jpg"), filepath.Join(root, "link.jpg")); err != nil {
		t.Fatal(err)
	}

	if _, err := v.AddLibrarySource(ctx, consent, "relative/path", "x"); err != ErrLibrarySource {
		t.Fatalf("relative source accepted: err=%v", err)
	}
	source, err := v.AddLibrarySource(ctx, consent, root, "Private")
	if err != nil {
		t.Fatal(err)
	}
	report, err := v.ScanLibrary(ctx, consent, source.ID)
	if err != nil {
		t.Fatal(err)
	}
	if report.Indexed != 2 || report.Duplicates != 1 || report.Skipped != 1 {
		t.Fatalf("scan report: %+v", report)
	}
	items, err := v.SearchLibrary(ctx, consent, LibraryQuery{})
	if err != nil || len(items) != 2 {
		t.Fatalf("two distinct media expected, got %+v err=%v", items, err)
	}
	for _, item := range items {
		if item.Name == "outside.jpg" || item.Name == "link.jpg" {
			t.Fatal("a symlink must not pull files from outside the source into the index")
		}
	}
	videos, _ := v.SearchLibrary(ctx, consent, LibraryQuery{MediaType: "video"})
	if len(videos) != 1 || videos[0].MediaType != "video/mp4" {
		t.Fatalf("media type filter: %+v", videos)
	}

	again, err := v.ScanLibrary(ctx, consent, source.ID)
	if err != nil || again.Unchanged != 3 || again.Indexed != 0 {
		t.Fatalf("an unchanged folder must rescan without rehashing: %+v err=%v", again, err)
	}

	for _, file := range []string{vaultPath, vaultPath + "-wal"} {
		raw, _ := os.ReadFile(file)
		if bytes.Contains(raw, []byte("lingerie-shoot-canary")) || bytes.Contains(raw, []byte(root)) {
			t.Fatalf("file names or folder paths reached %s in plaintext", filepath.Base(file))
		}
	}
}

func TestLibraryTagsSearchRemovalAndOfflineSources(t *testing.T) {
	v, _, consent := libraryVault(t)
	ctx := context.Background()
	parent := t.TempDir()
	root := filepath.Join(parent, "nas-share")
	writeFile(t, filepath.Join(root, "a.png"), pngBytes)
	writeFile(t, filepath.Join(root, "b.jpg"), jpgBytes)
	source, err := v.AddLibrarySource(ctx, consent, root, "NAS")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.ScanLibrary(ctx, consent, source.ID); err != nil {
		t.Fatal(err)
	}
	items, _ := v.SearchLibrary(ctx, consent, LibraryQuery{Text: "a.png"})
	if len(items) != 1 {
		t.Fatalf("name search: %+v", items)
	}
	if err := v.TagLibraryItem(ctx, consent, items[0].ID, []string{" Lingerie ", "photoreal", "lingerie"}); err != nil {
		t.Fatal(err)
	}
	tagged, _ := v.SearchLibrary(ctx, consent, LibraryQuery{Tags: []string{"lingerie", "photoreal"}})
	if len(tagged) != 1 || len(tagged[0].Tags) != 2 {
		t.Fatalf("tag search: %+v", tagged)
	}

	if err := v.RotateKey(bytes.Repeat([]byte{6}, 32), true); err != nil {
		t.Fatal(err)
	}
	if rotated, err := v.SearchLibrary(ctx, consent, LibraryQuery{Tags: []string{"lingerie"}}); err != nil || len(rotated) != 1 {
		t.Fatalf("library must survive key rotation: %+v err=%v", rotated, err)
	}

	if err := os.Rename(root, root+"-unmounted"); err != nil {
		t.Fatal(err)
	}
	offline, err := v.ScanLibrary(ctx, consent, source.ID)
	if err != nil || offline.Online || offline.Unavailable != 2 {
		t.Fatalf("an offline share must keep its index and report it: %+v err=%v", offline, err)
	}
	kept, _ := v.SearchLibrary(ctx, consent, LibraryQuery{})
	if len(kept) != 2 || kept[0].Available {
		t.Fatalf("offline items must stay indexed and unavailable: %+v", kept)
	}
	if err := os.Rename(root+"-unmounted", root); err != nil {
		t.Fatal(err)
	}
	if back, err := v.ScanLibrary(ctx, consent, source.ID); err != nil || !back.Online || back.Unchanged != 2 {
		t.Fatalf("remounted share: %+v err=%v", back, err)
	}

	if err := v.RemoveLibraryItem(ctx, consent, items[0].ID); err != nil {
		t.Fatal(err)
	}
	if left, _ := v.SearchLibrary(ctx, consent, LibraryQuery{}); len(left) != 1 {
		t.Fatalf("removed item still indexed: %+v", left)
	}
	if _, err := os.Stat(filepath.Join(root, "a.png")); err != nil {
		t.Fatal("removing from the index must not touch the file")
	}

	other, _ := v.RecordOwnerConsent(ctx, "owner", []Scope{ScopeMemoryRead}, time.Now().Add(time.Hour), true, true)
	if _, err := v.SearchLibrary(ctx, other.ID, LibraryQuery{}); err != ErrConsentDenied {
		t.Fatalf("library read without its scope: err=%v", err)
	}
}
