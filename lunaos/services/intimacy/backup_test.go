package intimacy

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEncryptedBackupRestoresAllPrivateData(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	vaultKey := bytes.Repeat([]byte{0x91}, 32)
	backupKey := bytes.Repeat([]byte{0x92}, 32)
	vaultPath := filepath.Join(dir, "vault.sqlite")
	vault, err := Open(vaultPath, vaultKey)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	consent, err := vault.RecordOwnerConsent(ctx, "owner", []Scope{ScopeMemoryRead, ScopeMemoryWrite, ScopePreferenceRead, ScopePreferenceWrite, ScopeDiscoveryPolicyRead, ScopeDiscoveryPolicyWrite, ScopeDiscoveryRun}, time.Now().Add(time.Hour), true, true)
	if err != nil {
		t.Fatal(err)
	}
	memory, err := vault.StoreMemory(ctx, consent.ID, "owner", OriginUserAsserted, "backup-memory-canary")
	if err != nil {
		t.Fatal(err)
	}
	preference, err := vault.AddPreference(ctx, consent.ID, PreferenceInput{Dimension: "visual_style", Value: "backup-preference-canary", Strength: .7, Timescale: "recurring", Confirmed: true})
	if err != nil {
		t.Fatal(err)
	}
	policy := DiscoveryAllowlist{Enabled: true, Sources: []string{"backup-source-canary"}, Creators: []string{"creator"}, Categories: []string{"category"}, ModelFamilies: []string{"family"}, Tags: []string{"tag"}}
	if err = vault.SetDiscoveryPolicy(ctx, consent.ID, policy, true); err != nil {
		t.Fatal(err)
	}
	backupPath := filepath.Join(dir, "vault.lunabak")
	if err = vault.CreateEncryptedBackup(ctx, backupPath, backupKey); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(backupPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("backup file mode=%o", info.Mode().Perm())
	}
	backupBytes, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"backup-memory-canary", "backup-preference-canary", "backup-source-canary"} {
		if bytes.Contains(backupBytes, []byte(secret)) {
			t.Fatalf("backup contains plaintext: %s", secret)
		}
	}
	restoredPath := filepath.Join(dir, "restored.sqlite")
	if err = RestoreEncryptedBackup(ctx, backupPath, restoredPath, backupKey, vaultKey); err != nil {
		t.Fatal(err)
	}
	if err = RestoreEncryptedBackup(ctx, backupPath, restoredPath, backupKey, vaultKey); !errors.Is(err, ErrBackupDestinationExists) {
		t.Fatalf("existing restore destination err=%v", err)
	}
	restored, err := Open(restoredPath, vaultKey)
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	gotMemory, err := restored.ReadMemory(ctx, memory.ID, consent.ID, "owner")
	if err != nil || gotMemory.Content != "backup-memory-canary" {
		t.Fatalf("restored memory=%+v err=%v", gotMemory, err)
	}
	gotPreferences, err := restored.ListPreferences(ctx, consent.ID)
	if err != nil || len(gotPreferences) != 1 || gotPreferences[0].Value != preference.Value {
		t.Fatalf("restored preferences=%+v err=%v", gotPreferences, err)
	}
	gotPolicy, err := restored.ReadDiscoveryPolicy(ctx, consent.ID)
	if err != nil || !gotPolicy.Enabled || gotPolicy.Sources[0] != policy.Sources[0] {
		t.Fatalf("restored policy=%+v err=%v", gotPolicy, err)
	}
}

func TestEncryptedBackupRejectsWrongKeyTamperingAndOverwrite(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	vaultKey := bytes.Repeat([]byte{0x93}, 32)
	backupKey := bytes.Repeat([]byte{0x94}, 32)
	vault, err := Open(filepath.Join(dir, "vault.sqlite"), vaultKey)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	consent, err := vault.RecordOwnerConsent(ctx, "owner", []Scope{ScopeMemoryWrite}, time.Now().Add(time.Hour), true, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = vault.StoreMemory(ctx, consent.ID, "owner", OriginUserAsserted, "key-check-canary"); err != nil {
		t.Fatal(err)
	}
	backupPath := filepath.Join(dir, "vault.lunabak")
	if err = vault.CreateEncryptedBackup(context.Background(), filepath.Join(dir, "same-key.lunabak"), vaultKey); !errors.Is(err, ErrBackupKeysMustDiffer) {
		t.Fatalf("backup key reuse err=%v", err)
	}
	if err = vault.CreateEncryptedBackup(context.Background(), backupPath, backupKey); err != nil {
		t.Fatal(err)
	}
	if err = vault.CreateEncryptedBackup(context.Background(), backupPath, backupKey); !errors.Is(err, ErrBackupDestinationExists) {
		t.Fatalf("existing backup overwrite err=%v", err)
	}
	if err = vault.Close(); err != nil {
		t.Fatal(err)
	}
	wrongKeyPath := filepath.Join(dir, "wrong.sqlite")
	if err = RestoreEncryptedBackup(context.Background(), backupPath, wrongKeyPath, bytes.Repeat([]byte{0x95}, 32), vaultKey); !errors.Is(err, ErrBackupAuthentication) {
		t.Fatalf("wrong backup key err=%v", err)
	}
	if _, err = os.Stat(wrongKeyPath); !os.IsNotExist(err) {
		t.Fatalf("wrong-key restore left a destination: %v", err)
	}
	wrongVaultKeyPath := filepath.Join(dir, "wrong-vault-key.sqlite")
	if err = RestoreEncryptedBackup(context.Background(), backupPath, wrongVaultKeyPath, backupKey, bytes.Repeat([]byte{0x96}, 32)); !errors.Is(err, ErrInvalidBackup) {
		t.Fatalf("wrong vault key err=%v", err)
	}
	if _, err = os.Stat(wrongVaultKeyPath); !os.IsNotExist(err) {
		t.Fatalf("wrong-vault-key restore left a destination: %v", err)
	}
	contents, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatal(err)
	}
	contents[len(contents)-1] ^= 0xff
	if err = os.WriteFile(backupPath, contents, 0600); err != nil {
		t.Fatal(err)
	}
	tamperedPath := filepath.Join(dir, "tampered.sqlite")
	if err = RestoreEncryptedBackup(context.Background(), backupPath, tamperedPath, backupKey, vaultKey); !errors.Is(err, ErrBackupAuthentication) {
		t.Fatalf("tampered backup err=%v", err)
	}
	if _, err = os.Stat(tamperedPath); !os.IsNotExist(err) {
		t.Fatalf("tampered restore left a destination: %v", err)
	}
	if err = RestoreEncryptedBackup(context.Background(), backupPath, tamperedPath, backupKey, vaultKey); !errors.Is(err, ErrBackupAuthentication) {
		t.Fatalf("second tampered restore err=%v", err)
	}
}

func TestEncryptedBackupChunkBoundariesRoundTrip(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	sourcePath := filepath.Join(dir, "source.bin")
	plain := bytes.Repeat([]byte("chunk-boundary"), (backupChunkSize+31)/len("chunk-boundary"))
	if err := os.WriteFile(sourcePath, plain, 0600); err != nil {
		t.Fatal(err)
	}
	var encrypted bytes.Buffer
	key := bytes.Repeat([]byte{0x97}, 32)
	if err := encryptBackupFile(sourcePath, &encrypted, key); err != nil {
		t.Fatal(err)
	}
	var restored bytes.Buffer
	if err := decryptBackupFile(bytes.NewReader(encrypted.Bytes()), &restored, key); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(restored.Bytes(), plain) {
		t.Fatalf("chunk roundtrip length=%d expected=%d", restored.Len(), len(plain))
	}
}

func TestGenerateOwnerKeyFileCreatesPrivateRandomKeyAndNeverOverwrites(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "vault.key")
	if err := GenerateOwnerKeyFile(path); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("key file mode=%o", info.Mode().Perm())
	}
	key, err := ReadOwnerKeyFile(path)
	if err != nil || len(key) != 32 {
		t.Fatalf("key length=%d err=%v", len(key), err)
	}
	clearKey(key)
	if err = GenerateOwnerKeyFile(path); !errors.Is(err, ErrBackupDestinationExists) {
		t.Fatalf("overwrite err=%v", err)
	}
	keyAfter, err := ReadOwnerKeyFile(path)
	if err != nil || len(keyAfter) != 32 {
		t.Fatalf("existing key length=%d err=%v", len(keyAfter), err)
	}
	clearKey(keyAfter)
}

func TestGenerateOwnerKeyFileRequiresPrivateDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := GenerateOwnerKeyFile(filepath.Join(dir, "key")); err == nil {
		t.Fatal("generated key in a non-private directory")
	}
}

func clearKey(key []byte) {
	for index := range key {
		key[index] = 0
	}
}
