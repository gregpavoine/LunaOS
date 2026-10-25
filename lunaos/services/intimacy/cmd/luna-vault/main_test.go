package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"lunaos.local/bootstrap/services/intimacy"
)

func TestVaultCommandBacksUpAndRestoresWithoutOverwriting(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	vaultKey := bytes.Repeat([]byte{0xa1}, 32)
	backupKey := bytes.Repeat([]byte{0xa2}, 32)
	vaultKeyFile := filepath.Join(dir, "vault.key")
	backupKeyFile := filepath.Join(dir, "backup.key")
	if err := os.WriteFile(vaultKeyFile, vaultKey, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(backupKeyFile, backupKey, 0600); err != nil {
		t.Fatal(err)
	}
	database := filepath.Join(dir, "vault.sqlite")
	vault, err := intimacy.Open(database, vaultKey)
	if err != nil {
		t.Fatal(err)
	}
	consent, err := vault.RecordOwnerConsent(context.Background(), "owner", []intimacy.Scope{intimacy.ScopeMemoryRead, intimacy.ScopeMemoryWrite}, time.Now().Add(time.Hour), true, true)
	if err != nil {
		t.Fatal(err)
	}
	memory, err := vault.StoreMemory(context.Background(), consent.ID, "owner", intimacy.OriginUserAsserted, "cli-backup-canary")
	if err != nil {
		t.Fatal(err)
	}
	if err = vault.Close(); err != nil {
		t.Fatal(err)
	}
	backup := filepath.Join(dir, "vault.lunabak")
	var output bytes.Buffer
	if err = run([]string{"backup", "--database", database, "--vault-key-file", vaultKeyFile, "--backup-key-file", backupKeyFile, "--destination", backup}, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "backup created") {
		t.Fatalf("backup output=%q", output.String())
	}
	restoredPath := filepath.Join(dir, "restored.sqlite")
	output.Reset()
	if err = run([]string{"restore", "--backup-file", backup, "--backup-key-file", backupKeyFile, "--vault-key-file", vaultKeyFile, "--destination", restoredPath}, &output); err != nil {
		t.Fatal(err)
	}
	restored, err := intimacy.Open(restoredPath, vaultKey)
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	got, err := restored.ReadMemory(context.Background(), memory.ID, consent.ID, "owner")
	if err != nil || got.Content != "cli-backup-canary" {
		t.Fatalf("restored=%+v err=%v", got, err)
	}
	if err = run([]string{"restore", "--backup-file", backup, "--backup-key-file", backupKeyFile, "--vault-key-file", vaultKeyFile, "--destination", restoredPath}, &output); !errors.Is(err, intimacy.ErrBackupDestinationExists) {
		t.Fatalf("restore overwrite err=%v", err)
	}
}

func TestVaultCommandRejectsMissingOptionsAndUnknownCommand(t *testing.T) {
	if err := run([]string{"backup"}, &bytes.Buffer{}); err == nil {
		t.Fatal("accepted missing backup options")
	}
	if err := run([]string{"restore"}, &bytes.Buffer{}); err == nil {
		t.Fatal("accepted missing restore options")
	}
	if err := run([]string{"unknown"}, &bytes.Buffer{}); err == nil {
		t.Fatal("accepted unknown command")
	}
}

func TestVaultCommandKeygenDoesNotPrintKeyMaterial(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "backup.key")
	var output bytes.Buffer
	if err := run([]string{"keygen", "--destination", path}, &output); err != nil {
		t.Fatal(err)
	}
	key, err := intimacy.ReadOwnerKeyFile(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		for index := range key {
			key[index] = 0
		}
	}()
	if bytes.Contains(output.Bytes(), key) {
		t.Fatal("keygen printed key bytes")
	}
	if err = run([]string{"keygen", "--destination", path}, &output); !errors.Is(err, intimacy.ErrBackupDestinationExists) {
		t.Fatalf("keygen overwrite err=%v", err)
	}
}
