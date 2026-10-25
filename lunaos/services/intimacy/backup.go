package intimacy

import (
	"bufio"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

var (
	ErrBackupAuthentication    = errors.New("BACKUP_AUTHENTICATION_FAILED")
	ErrInvalidBackup           = errors.New("INVALID_ENCRYPTED_BACKUP")
	ErrBackupDestinationExists = errors.New("BACKUP_DESTINATION_EXISTS")
	ErrBackupKeysMustDiffer    = errors.New("BACKUP_AND_VAULT_KEYS_MUST_DIFFER")
)

const (
	backupChunkSize = 1 << 20
	backupMagic     = "LUNABAK1"
)

// CreateEncryptedBackup uses SQLite VACUUM INTO for a consistent snapshot,
// encrypts it in authenticated chunks, and publishes the file without replacing
// an existing destination. The backup key must be an independent 32-byte key.
func (v *Vault) CreateEncryptedBackup(ctx context.Context, destination string, backupKey []byte) error {
	if len(backupKey) != 32 {
		return ErrInvalidKey
	}
	backupFingerprint := sha256.Sum256(backupKey)
	if subtle.ConstantTimeCompare(v.keyFingerprint[:], backupFingerprint[:]) == 1 {
		return ErrBackupKeysMustDiffer
	}
	absolute, err := privateDestination(destination)
	if err != nil {
		return err
	}
	if err = ensureDoesNotExist(absolute); err != nil {
		return err
	}
	snapshot, err := makePrivateTempPath(filepath.Dir(absolute), ".luna-vault-snapshot-*.sqlite")
	if err != nil {
		return err
	}
	defer removeDatabaseFiles(snapshot)
	if _, err = v.db.ExecContext(ctx, `VACUUM INTO ?`, snapshot); err != nil {
		return err
	}
	if err = os.Chmod(snapshot, 0600); err != nil {
		return err
	}
	backupTemp, err := os.CreateTemp(filepath.Dir(absolute), ".luna-vault-backup-*.tmp")
	if err != nil {
		return err
	}
	backupTempPath := backupTemp.Name()
	defer os.Remove(backupTempPath)
	if err = encryptBackupFile(snapshot, backupTemp, backupKey); err != nil {
		_ = backupTemp.Close()
		return err
	}
	if err = backupTemp.Sync(); err != nil {
		_ = backupTemp.Close()
		return err
	}
	if err = backupTemp.Close(); err != nil {
		return err
	}
	if err = os.Link(backupTempPath, absolute); err != nil {
		if errors.Is(err, os.ErrExist) {
			return ErrBackupDestinationExists
		}
		return err
	}
	if err = syncDirectory(filepath.Dir(absolute)); err != nil {
		_ = os.Remove(absolute)
		return err
	}
	return os.Remove(backupTempPath)
}

// RestoreEncryptedBackup decrypts to a private temporary database, validates
// SQLite integrity and every encrypted payload, and atomically publishes only a
// verified database. The destination must not already exist.
func RestoreEncryptedBackup(ctx context.Context, backupPath, destination string, backupKey, vaultKey []byte) error {
	if len(backupKey) != 32 || len(vaultKey) != 32 {
		return ErrInvalidKey
	}
	if subtle.ConstantTimeCompare(backupKey, vaultKey) == 1 {
		return ErrBackupKeysMustDiffer
	}
	source, err := secureOwnerFile(backupPath)
	if err != nil {
		return err
	}
	destinationPath, err := privateDestination(destination)
	if err != nil {
		return err
	}
	if err = ensureDoesNotExist(destinationPath); err != nil {
		return err
	}
	tempDB, err := makePrivateTempPath(filepath.Dir(destinationPath), ".luna-vault-restore-*.sqlite")
	if err != nil {
		return err
	}
	defer removeDatabaseFiles(tempDB)
	output, err := os.OpenFile(tempDB, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	input, err := os.Open(source)
	if err != nil {
		_ = output.Close()
		return err
	}
	decryptErr := decryptBackupFile(input, output, backupKey)
	closeInputErr := input.Close()
	if decryptErr != nil {
		_ = output.Close()
		return decryptErr
	}
	if closeInputErr != nil {
		_ = output.Close()
		return closeInputErr
	}
	if err = output.Sync(); err != nil {
		_ = output.Close()
		return err
	}
	if err = output.Close(); err != nil {
		return err
	}
	if err = verifyRestoredVault(ctx, tempDB, vaultKey); err != nil {
		return err
	}
	if err = syncFile(tempDB); err != nil {
		return err
	}
	if err = os.Link(tempDB, destinationPath); err != nil {
		if errors.Is(err, os.ErrExist) {
			return ErrBackupDestinationExists
		}
		return err
	}
	if err = syncDirectory(filepath.Dir(destinationPath)); err != nil {
		_ = os.Remove(destinationPath)
		return err
	}
	return nil
}

func encryptBackupFile(snapshotPath string, destination io.Writer, key []byte) error {
	aead, err := backupAEAD(key)
	if err != nil {
		return err
	}
	input, err := os.Open(snapshotPath)
	if err != nil {
		return err
	}
	defer input.Close()
	if err = writeAll(destination, []byte(backupMagic)); err != nil {
		return err
	}
	buffer := make([]byte, backupChunkSize)
	var counter uint64
	for {
		count, readErr := input.Read(buffer)
		if count > 0 {
			if err = writeEncryptedChunk(destination, aead, counter, buffer[:count]); err != nil {
				return err
			}
			counter++
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	var end [4]byte
	return writeAll(destination, end[:])
}

func decryptBackupFile(source io.Reader, destination io.Writer, key []byte) error {
	aead, err := backupAEAD(key)
	if err != nil {
		return err
	}
	reader := bufio.NewReader(source)
	header := make([]byte, len(backupMagic))
	if _, err = io.ReadFull(reader, header); err != nil || string(header) != backupMagic {
		return ErrBackupAuthentication
	}
	var counter uint64
	for {
		var size uint32
		if err = binary.Read(reader, binary.BigEndian, &size); err != nil {
			return ErrBackupAuthentication
		}
		if size == 0 {
			if _, err = reader.ReadByte(); !errors.Is(err, io.EOF) {
				return ErrBackupAuthentication
			}
			return nil
		}
		if size > backupChunkSize {
			return ErrBackupAuthentication
		}
		nonce := make([]byte, aead.NonceSize())
		if _, err = io.ReadFull(reader, nonce); err != nil {
			return ErrBackupAuthentication
		}
		ciphertext := make([]byte, int(size)+aead.Overhead())
		if _, err = io.ReadFull(reader, ciphertext); err != nil {
			return ErrBackupAuthentication
		}
		plain, openErr := aead.Open(nil, nonce, ciphertext, backupAAD(counter))
		if openErr != nil || len(plain) != int(size) {
			return ErrBackupAuthentication
		}
		if _, err = destination.Write(plain); err != nil {
			return err
		}
		counter++
	}
}

func writeEncryptedChunk(writer io.Writer, aead cipher.AEAD, counter uint64, plain []byte) error {
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return err
	}
	ciphertext := aead.Seal(nil, nonce, plain, backupAAD(counter))
	var size [4]byte
	binary.BigEndian.PutUint32(size[:], uint32(len(plain)))
	if err := writeAll(writer, size[:]); err != nil {
		return err
	}
	if err := writeAll(writer, nonce); err != nil {
		return err
	}
	return writeAll(writer, ciphertext)
}

func writeAll(writer io.Writer, data []byte) error {
	for len(data) > 0 {
		written, err := writer.Write(data)
		if err != nil {
			return err
		}
		if written == 0 {
			return io.ErrShortWrite
		}
		data = data[written:]
	}
	return nil
}

func backupAEAD(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(append([]byte(nil), key...))
	if err != nil {
		return nil, ErrInvalidKey
	}
	return cipher.NewGCM(block)
}

func backupAAD(counter uint64) []byte {
	buffer := make([]byte, len(backupMagic)+8)
	copy(buffer, backupMagic)
	binary.BigEndian.PutUint64(buffer[len(backupMagic):], counter)
	return buffer
}

func verifyRestoredVault(ctx context.Context, path string, key []byte) error {
	vault, err := Open(path, key)
	if err != nil {
		return ErrInvalidBackup
	}
	defer vault.Close()
	var integrity string
	if err = vault.db.QueryRowContext(ctx, `PRAGMA integrity_check`).Scan(&integrity); err != nil || integrity != "ok" {
		return ErrInvalidBackup
	}
	rows, err := vault.db.QueryContext(ctx, `SELECT id,subject_id,origin_type,ciphertext FROM private_memory ORDER BY id`)
	if err != nil {
		return ErrInvalidBackup
	}
	memoryRows := make([]encryptedRow, 0)
	for rows.Next() {
		var row encryptedRow
		var subject string
		if err = rows.Scan(&row.id, &subject, &row.origin, &row.ciphertext); err != nil || subject != "owner" {
			_ = rows.Close()
			return ErrInvalidBackup
		}
		memoryRows = append(memoryRows, row)
	}
	if err = rows.Err(); err != nil {
		_ = rows.Close()
		return ErrInvalidBackup
	}
	if err = rows.Close(); err != nil {
		return ErrInvalidBackup
	}
	for _, row := range memoryRows {
		plain, openErr := openWithAEAD(vault.aead, row.ciphertext, associatedData(row.id, "owner", row.origin))
		var payload struct {
			Content string `json:"content"`
		}
		if openErr != nil || !validOrigin(row.origin) || json.Unmarshal(plain, &payload) != nil || payload.Content == "" {
			return ErrInvalidBackup
		}
	}
	rows, err = vault.db.QueryContext(ctx, `SELECT id FROM private_preference ORDER BY id`)
	if err != nil {
		return ErrInvalidBackup
	}
	preferenceIDs := make([]string, 0)
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			_ = rows.Close()
			return ErrInvalidBackup
		}
		preferenceIDs = append(preferenceIDs, id)
	}
	if err = rows.Err(); err != nil {
		_ = rows.Close()
		return ErrInvalidBackup
	}
	if err = rows.Close(); err != nil {
		return ErrInvalidBackup
	}
	for _, id := range preferenceIDs {
		if _, err = vault.readPreferenceQuery(ctx, vault.db, id); err != nil {
			return ErrInvalidBackup
		}
	}
	if _, err = vault.readDiscoveryPolicyQuery(ctx, vault.db); err != nil {
		return ErrInvalidBackup
	}
	return nil
}

func privateDestination(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("destination path is required")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if err = privateDirectory(filepath.Dir(absolute)); err != nil {
		return "", err
	}
	return absolute, nil
}

func ensureDoesNotExist(path string) error {
	_, err := os.Lstat(path)
	if err == nil {
		return ErrBackupDestinationExists
	}
	if !os.IsNotExist(err) {
		return err
	}
	return nil
}

func secureOwnerFile(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	info, err := os.Lstat(absolute)
	if err != nil {
		return "", err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 || !ok || int(stat.Uid) != os.Getuid() {
		return "", fmt.Errorf("backup file must be owner-only regular file")
	}
	return absolute, nil
}

func makePrivateTempPath(directory, pattern string) (string, error) {
	file, err := os.CreateTemp(directory, pattern)
	if err != nil {
		return "", err
	}
	path := file.Name()
	if err = file.Close(); err != nil {
		_ = os.Remove(path)
		return "", err
	}
	if err = os.Remove(path); err != nil {
		return "", err
	}
	return path, nil
}

func removeDatabaseFiles(path string) {
	_ = os.Remove(path)
	_ = os.Remove(path + "-wal")
	_ = os.Remove(path + "-shm")
	_ = os.Remove(path + "-journal")
}

func syncFile(path string) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	return file.Sync()
}

func syncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}

// ReadOwnerKeyFile loads a 32-byte raw or 64-character hexadecimal key from a
// regular owner-only file. The caller should clear the returned buffer when done.
func ReadOwnerKeyFile(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 {
		return nil, fmt.Errorf("key must be an owner-only regular file")
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || int(stat.Uid) != os.Getuid() {
		return nil, fmt.Errorf("key file owner mismatch")
	}
	key, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("key file unavailable")
	}
	if len(key) == 65 && key[64] == '\n' {
		key = key[:64]
	}
	if len(key) == 64 {
		decoded := make([]byte, 32)
		if _, err = hex.Decode(decoded, key); err == nil {
			for i := range key {
				key[i] = 0
			}
			key = decoded
		}
	}
	if len(key) != 32 {
		for i := range key {
			key[i] = 0
		}
		return nil, ErrInvalidKey
	}
	return key, nil
}

// GenerateOwnerKeyFile creates a random 32-byte key in a private directory,
// publishes it without replacing an existing path, and never returns key bytes.
func GenerateOwnerKeyFile(destination string) error {
	absolute, err := privateDestination(destination)
	if err != nil {
		return err
	}
	if err = ensureDoesNotExist(absolute); err != nil {
		return err
	}
	key := make([]byte, 32)
	defer func() {
		for index := range key {
			key[index] = 0
		}
	}()
	if _, err = rand.Read(key); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(absolute), ".luna-key-*.tmp")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if err = writeAll(temp, key); err != nil {
		_ = temp.Close()
		return err
	}
	if err = temp.Sync(); err != nil {
		_ = temp.Close()
		return err
	}
	if err = temp.Close(); err != nil {
		return err
	}
	if err = os.Link(tempPath, absolute); err != nil {
		if errors.Is(err, os.ErrExist) {
			return ErrBackupDestinationExists
		}
		return err
	}
	if err = syncDirectory(filepath.Dir(absolute)); err != nil {
		_ = os.Remove(absolute)
		return err
	}
	return os.Remove(tempPath)
}
