package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestReadVaultKeyAcceptsPrivateRawAndHex(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	for name, value := range map[string][]byte{
		"raw": bytes.Repeat([]byte{0x41}, 32),
		"hex": []byte("4141414141414141414141414141414141414141414141414141414141414141\n"),
	} {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, value, 0600); err != nil {
			t.Fatal(err)
		}
		key, err := readVaultKey(path)
		if err != nil || !bytes.Equal(key, bytes.Repeat([]byte{0x41}, 32)) {
			t.Fatalf("%s key read len=%d err=%v", name, len(key), err)
		}
	}
}

func TestReadVaultKeyRejectsBroadPermissionsAndSymlink(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	secret := filepath.Join(dir, "secret")
	if err := os.WriteFile(secret, bytes.Repeat([]byte{0x41}, 32), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := readVaultKey(secret); err == nil {
		t.Fatal("accepted group-readable vault key")
	}
	if err := os.Chmod(secret, 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(secret, link); err != nil {
		t.Fatal(err)
	}
	if _, err := readVaultKey(link); err == nil {
		t.Fatal("accepted symlinked vault key")
	}
}
