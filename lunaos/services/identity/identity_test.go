package identity

import (
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestProvisionOwnerTokenCreatesPrivateHexCredential(t *testing.T) {
	path := filepath.Join(t.TempDir(), "new-private-dir", "owner-token")
	token, err := ProvisionOwnerToken(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(token) != 64 {
		t.Fatalf("credential length = %d, want 64 hex characters", len(token))
	}
	decoded, err := hex.DecodeString(token)
	if err != nil || len(decoded) != 32 {
		t.Fatalf("credential is not 32 random bytes encoded as hex: %v", err)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != token+"\n" {
		t.Fatalf("credential file does not match generated token")
	}
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm() != 0600 {
		t.Fatalf("credential file mode = %v, want regular 0600 file", info.Mode())
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); ok && int(stat.Uid) != os.Getuid() {
		t.Fatalf("credential owner UID = %d, want %d", stat.Uid, os.Getuid())
	}
	parent, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if parent.Mode().Perm() != 0700 {
		t.Fatalf("new credential directory mode = %v, want 0700", parent.Mode().Perm())
	}
}

func TestProvisionOwnerTokenRefusesReplacement(t *testing.T) {
	path := filepath.Join(t.TempDir(), "owner-token")
	if err := os.WriteFile(path, []byte("keep-existing"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ProvisionOwnerToken(path); !errors.Is(err, ErrAlreadyProvisioned) {
		t.Fatalf("replacement error = %v, want ErrAlreadyProvisioned", err)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "keep-existing" {
		t.Fatalf("existing credential was changed: %q", contents)
	}
}

func TestProvisionOwnerTokenRejectsSymlinkedCredential(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	path := filepath.Join(dir, "owner-token")
	if err := os.WriteFile(target, []byte("keep-existing"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
	if _, err := ProvisionOwnerToken(path); !errors.Is(err, ErrAlreadyProvisioned) {
		t.Fatalf("symlink destination error = %v, want ErrAlreadyProvisioned", err)
	}
	contents, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "keep-existing" {
		t.Fatalf("symlink target was changed: %q", contents)
	}
}

func TestProvisionServiceCredentialsCreatesDistinctPrivateTokens(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	ownerPath := filepath.Join(dir, "owner-token")
	policyPath := filepath.Join(dir, "policy-token")
	owner, policy, err := ProvisionServiceCredentials(ownerPath, policyPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(owner) != 64 || len(policy) != 64 || owner == policy {
		t.Fatalf("credentials are not distinct 32-byte hex tokens")
	}
	for _, path := range []string{ownerPath, policyPath} {
		info, err := os.Stat(path)
		if err != nil || info.Mode().Perm() != 0600 {
			t.Fatalf("credential file %s mode=%v err=%v", path, info.Mode(), err)
		}
	}
	if _, _, err := ProvisionServiceCredentials(ownerPath, policyPath); !errors.Is(err, ErrAlreadyProvisioned) {
		t.Fatalf("second provisioning error=%v", err)
	}
}

func TestProvisionServiceCredentialsDoesNotLeavePartialPair(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	ownerPath := filepath.Join(dir, "owner-token")
	policyPath := filepath.Join(dir, "policy-token")
	if err := os.WriteFile(policyPath, []byte("existing-policy"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ProvisionServiceCredentials(ownerPath, policyPath); !errors.Is(err, ErrAlreadyProvisioned) {
		t.Fatalf("expected existing policy credential to fail atomically, got %v", err)
	}
	if _, err := os.Lstat(ownerPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("partial owner credential left behind: %v", err)
	}
}
