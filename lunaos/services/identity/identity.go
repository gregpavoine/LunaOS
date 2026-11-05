package identity

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

var ErrAlreadyProvisioned = errors.New("owner credential already exists")

// ProvisionOwnerToken creates a random 32-byte hex credential atomically and
// never replaces an existing path. The caller must be the credential owner.
func ProvisionOwnerToken(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("credential path is required")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve credential path: %w", err)
	}
	parent := filepath.Dir(absolute)
	if err := ensurePrivateDirectory(parent); err != nil {
		return "", err
	}

	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", fmt.Errorf("generate owner credential: %w", err)
	}
	token := hex.EncodeToString(secret)
	file, err := os.CreateTemp(parent, ".luna-owner-*.tmp")
	if err != nil {
		return "", fmt.Errorf("create private credential temporary file: %w", err)
	}
	temporaryPath := file.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err := file.Chmod(0600); err != nil {
		_ = file.Close()
		return "", fmt.Errorf("secure credential file: %w", err)
	}
	if _, err := io.WriteString(file, token+"\n"); err != nil {
		_ = file.Close()
		return "", fmt.Errorf("write owner credential: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return "", fmt.Errorf("sync owner credential: %w", err)
	}
	if err := file.Close(); err != nil {
		return "", fmt.Errorf("close owner credential: %w", err)
	}
	if err := os.Link(temporaryPath, absolute); err != nil {
		if errors.Is(err, os.ErrExist) {
			return "", ErrAlreadyProvisioned
		}
		return "", fmt.Errorf("publish owner credential: %w", err)
	}
	if err := syncDirectory(parent); err != nil {
		return "", fmt.Errorf("sync credential directory: %w", err)
	}
	return token, nil
}

func ensurePrivateDirectory(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(path, 0700); err != nil {
			return fmt.Errorf("create credential directory: %w", err)
		}
		if err := os.Chmod(path, 0700); err != nil {
			return fmt.Errorf("secure credential directory: %w", err)
		}
		info, err = os.Lstat(path)
	}
	if err != nil {
		return fmt.Errorf("inspect credential directory: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() || info.Mode().Perm()&0022 != 0 {
		return fmt.Errorf("credential directory must be a real directory not writable by group or other")
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); ok && int(stat.Uid) != os.Getuid() {
		return fmt.Errorf("credential directory owner mismatch")
	}
	return nil
}

func syncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}

// ProvisionServiceCredentials creates independent owner/API and Core-to-Broker
// credentials. It refuses to replace either path and rolls back a newly created
// owner credential if publishing the broker credential fails.
func ProvisionServiceCredentials(ownerPath, policyPath string) (string, string, error) {
	if ownerPath == "" || policyPath == "" {
		return "", "", fmt.Errorf("owner and policy credential paths are required")
	}
	ownerAbsolute, err := filepath.Abs(ownerPath)
	if err != nil {
		return "", "", fmt.Errorf("resolve owner credential path: %w", err)
	}
	policyAbsolute, err := filepath.Abs(policyPath)
	if err != nil {
		return "", "", fmt.Errorf("resolve policy credential path: %w", err)
	}
	if ownerAbsolute == policyAbsolute {
		return "", "", fmt.Errorf("owner and policy credential paths must differ")
	}
	for _, path := range []string{ownerAbsolute, policyAbsolute} {
		if err := ensurePrivateDirectory(filepath.Dir(path)); err != nil {
			return "", "", err
		}
		if _, err := os.Lstat(path); err == nil {
			return "", "", ErrAlreadyProvisioned
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", "", fmt.Errorf("inspect credential path: %w", err)
		}
	}
	owner, err := ProvisionOwnerToken(ownerAbsolute)
	if err != nil {
		return "", "", err
	}
	policy, err := ProvisionOwnerToken(policyAbsolute)
	if err != nil {
		if removeErr := os.Remove(ownerAbsolute); removeErr != nil {
			return "", "", fmt.Errorf("publish policy credential: %w (owner credential rollback failed: %v)", err, removeErr)
		}
		return "", "", err
	}
	return owner, policy, nil
}
