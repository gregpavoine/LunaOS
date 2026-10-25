// Package intimacy is the encrypted, consent-scoped local data layer. It does
// not implement inference, media discovery, or generation.
package intimacy

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Scope string
type Origin string

const (
	ScopeMemoryRead           Scope = "memory.read"
	ScopeMemoryWrite          Scope = "memory.write"
	ScopePreferenceRead       Scope = "preferences.read"
	ScopePreferenceWrite      Scope = "preferences.write"
	ScopeDiscoveryPolicyRead  Scope = "discovery.policy.read"
	ScopeDiscoveryPolicyWrite Scope = "discovery.policy.write"
	ScopeDiscoveryRun         Scope = "discovery.run"

	OriginUserAsserted      Origin = "user_asserted"
	OriginObserved          Origin = "observed"
	OriginInferred          Origin = "inferred"
	OriginFictionalRoleplay Origin = "fictional_roleplay"
	OriginGeneratedArtifact Origin = "generated_artifact"
	OriginExternalMedia     Origin = "external_media"
)

var (
	ErrInvalidKey               = errors.New("INVALID_VAULT_KEY")
	ErrInvalidConsent           = errors.New("INVALID_CONSENT")
	ErrConsentDenied            = errors.New("CONSENT_DENIED")
	ErrInvalidOrigin            = errors.New("INVALID_MEMORY_ORIGIN")
	ErrMemoryNotFound           = errors.New("MEMORY_NOT_FOUND")
	ErrPreferenceNotFound       = errors.New("PREFERENCE_NOT_FOUND")
	ErrInvalidPreference        = errors.New("INVALID_PREFERENCE")
	ErrVaultRotationUnconfirmed = errors.New("VAULT_ROTATION_UNCONFIRMED")
)

type Vault struct {
	db             *sql.DB
	aead           cipher.AEAD
	keyFingerprint [32]byte
}

type Consent struct {
	ID        string     `json:"consent_id"`
	SubjectID string     `json:"subject_id"`
	Scopes    []Scope    `json:"scopes"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

type Memory struct {
	ID        string    `json:"memory_id"`
	SubjectID string    `json:"subject_id"`
	Origin    Origin    `json:"origin_type"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func Open(path string, key []byte) (*Vault, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKey
	}
	if path == "" {
		return nil, fmt.Errorf("invalid vault path")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if err = privateDirectory(filepath.Dir(absolute)); err != nil {
		return nil, err
	}
	if err = prepareDatabase(absolute); err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(append([]byte(nil), key...))
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	uri := (&url.URL{Scheme: "file", Path: filepath.ToSlash(absolute)}).String()
	db, err := sql.Open("sqlite3", uri+"?_busy_timeout=5000&_journal_mode=WAL&_synchronous=FULL&_foreign_keys=on")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err = db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err = os.Chmod(absolute, 0600); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err = db.Exec(vaultSchema); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Vault{db: db, aead: aead, keyFingerprint: sha256.Sum256(key)}, nil
}

func (v *Vault) Close() error { return v.db.Close() }

// RecordOwnerConsent only records the local owner's explicit declaration.
// Third-party, biometric and intimate-image consent need a separate verified flow.
func (v *Vault) RecordOwnerConsent(ctx context.Context, subjectID string, scopes []Scope, expires time.Time, ownerConfirmed, adultAttested bool) (Consent, error) {
	if subjectID != "owner" || !ownerConfirmed || !adultAttested || !expires.After(time.Now()) || len(scopes) == 0 {
		return Consent{}, ErrInvalidConsent
	}
	allowed := map[Scope]bool{ScopeMemoryRead: true, ScopeMemoryWrite: true, ScopePreferenceRead: true, ScopePreferenceWrite: true, ScopeDiscoveryPolicyRead: true, ScopeDiscoveryPolicyWrite: true, ScopeDiscoveryRun: true}
	seen := make(map[Scope]bool, len(scopes))
	for _, scope := range scopes {
		if !allowed[scope] || seen[scope] {
			return Consent{}, ErrInvalidConsent
		}
		seen[scope] = true
	}
	id, err := randomID()
	if err != nil {
		return Consent{}, err
	}
	scopeJSON, err := json.Marshal(scopes)
	if err != nil {
		return Consent{}, err
	}
	now := time.Now().UTC()
	tx, err := v.db.BeginTx(ctx, nil)
	if err != nil {
		return Consent{}, err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `INSERT INTO private_consent(id,subject_id,scopes_json,expires_at_utc,adult_attested_at_utc,created_at_utc) VALUES(?,?,?,?,?,?)`, id, subjectID, string(scopeJSON), expires.UTC().Format(time.RFC3339Nano), now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano))
	if err != nil {
		return Consent{}, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO private_audit(event_code,subject_id,created_at_utc) VALUES('CONSENT_GRANTED','owner',?)`, now.Format(time.RFC3339Nano)); err != nil {
		return Consent{}, err
	}
	if err = tx.Commit(); err != nil {
		return Consent{}, err
	}
	return Consent{ID: id, SubjectID: subjectID, Scopes: scopes, ExpiresAt: expires.UTC()}, nil
}

func (v *Vault) StoreMemory(ctx context.Context, consentID, subjectID string, origin Origin, content string) (Memory, error) {
	if !validOrigin(origin) {
		return Memory{}, ErrInvalidOrigin
	}
	if len(content) == 0 || len(content) > 1<<20 {
		return Memory{}, ErrInvalidConsent
	}
	if err := v.authorize(ctx, consentID, subjectID, ScopeMemoryWrite); err != nil {
		return Memory{}, err
	}
	id, err := randomID()
	if err != nil {
		return Memory{}, err
	}
	nonce := make([]byte, v.aead.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return Memory{}, err
	}
	created := time.Now().UTC()
	plain, err := json.Marshal(struct {
		Content string `json:"content"`
	}{Content: content})
	if err != nil {
		return Memory{}, err
	}
	ciphertext := v.aead.Seal(nonce, nonce, plain, associatedData(id, subjectID, origin))
	_, err = v.db.ExecContext(ctx, `INSERT INTO private_memory(id,subject_id,origin_type,ciphertext,created_at_utc) VALUES(?,?,?,?,?)`, id, subjectID, string(origin), ciphertext, created.Format(time.RFC3339Nano))
	if err != nil {
		return Memory{}, err
	}
	return Memory{ID: id, SubjectID: subjectID, Origin: origin, Content: content, CreatedAt: created}, nil
}

func (v *Vault) ReadMemory(ctx context.Context, id, consentID, subjectID string) (Memory, error) {
	if err := v.authorize(ctx, consentID, subjectID, ScopeMemoryRead); err != nil {
		return Memory{}, err
	}
	var memory Memory
	var ciphertext []byte
	var created string
	err := v.db.QueryRowContext(ctx, `SELECT id,subject_id,origin_type,ciphertext,created_at_utc FROM private_memory WHERE id=? AND subject_id=?`, id, subjectID).Scan(&memory.ID, &memory.SubjectID, &memory.Origin, &ciphertext, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return Memory{}, ErrMemoryNotFound
	}
	if err != nil {
		return Memory{}, err
	}
	if !validOrigin(memory.Origin) || len(ciphertext) < v.aead.NonceSize() {
		return Memory{}, ErrMemoryNotFound
	}
	nonce, encrypted := ciphertext[:v.aead.NonceSize()], ciphertext[v.aead.NonceSize():]
	plain, err := v.aead.Open(nil, nonce, encrypted, associatedData(memory.ID, memory.SubjectID, memory.Origin))
	if err != nil {
		return Memory{}, ErrMemoryNotFound
	}
	var payload struct {
		Content string `json:"content"`
	}
	if err = json.Unmarshal(plain, &payload); err != nil {
		return Memory{}, ErrMemoryNotFound
	}
	memory.Content = payload.Content
	memory.CreatedAt, err = time.Parse(time.RFC3339Nano, created)
	return memory, err
}

func (v *Vault) RevokeConsent(ctx context.Context, consentID, subjectID string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	tx, err := v.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `UPDATE private_consent SET revoked_at_utc=? WHERE id=? AND subject_id=? AND revoked_at_utc IS NULL`, now, consentID, subjectID)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed != 1 {
		return ErrConsentDenied
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO private_audit(event_code,subject_id,created_at_utc) VALUES('CONSENT_REVOKED','owner',?)`, now); err != nil {
		return err
	}
	return tx.Commit()
}

// Deletion remains available after revocation so withdrawal does not trap the
// owner into retaining data whose processing permission has ended.
func (v *Vault) DeleteMemory(ctx context.Context, id, subjectID string) error {
	tx, err := v.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `DELETE FROM private_memory WHERE id=? AND subject_id=?`, id, subjectID)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed != 1 {
		return ErrMemoryNotFound
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO private_audit(event_code,subject_id,created_at_utc) VALUES('MEMORY_DELETED','owner',?)`, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		return err
	}
	return tx.Commit()
}

func (v *Vault) authorize(ctx context.Context, consentID, subjectID string, scope Scope) error {
	return authorizeQuery(ctx, v.db, consentID, subjectID, scope)
}

func privateDirectory(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("vault directory must already exist with mode 0700")
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() || info.Mode().Perm()&0077 != 0 {
		return fmt.Errorf("vault requires a private non-symlink directory")
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || int(stat.Uid) != os.Getuid() {
		return fmt.Errorf("vault directory owner mismatch")
	}
	return nil
}

func prepareDatabase(path string) error {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		file, createErr := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
		if createErr != nil {
			return createErr
		}
		return file.Close()
	}
	if err != nil {
		return err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 || !ok || int(stat.Uid) != os.Getuid() {
		return fmt.Errorf("vault database must be an owner-only regular file")
	}
	return nil
}

func validOrigin(origin Origin) bool {
	switch origin {
	case OriginUserAsserted, OriginObserved, OriginInferred, OriginFictionalRoleplay, OriginGeneratedArtifact, OriginExternalMedia:
		return true
	default:
		return false
	}
}

func associatedData(id, subjectID string, origin Origin) []byte {
	return []byte(id + "\x00" + subjectID + "\x00" + string(origin))
}

func randomID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	bytes[6] = bytes[6]&0x0f | 0x40
	bytes[8] = bytes[8]&0x3f | 0x80
	encoded := hex.EncodeToString(bytes)
	return encoded[:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:], nil
}

const vaultSchema = `
PRAGMA foreign_keys=ON;
CREATE TABLE IF NOT EXISTS private_schema_version (
 version INTEGER PRIMARY KEY,
 applied_at_utc TEXT NOT NULL
);
INSERT OR IGNORE INTO private_schema_version(version,applied_at_utc)
VALUES (1,strftime('%Y-%m-%dT%H:%M:%fZ','now'));
INSERT OR IGNORE INTO private_schema_version(version,applied_at_utc)
VALUES (2,strftime('%Y-%m-%dT%H:%M:%fZ','now'));
INSERT OR IGNORE INTO private_schema_version(version,applied_at_utc)
VALUES (3,strftime('%Y-%m-%dT%H:%M:%fZ','now'));
INSERT OR IGNORE INTO private_schema_version(version,applied_at_utc)
VALUES (4,strftime('%Y-%m-%dT%H:%M:%fZ','now'));
CREATE TABLE IF NOT EXISTS private_consent (
 id TEXT PRIMARY KEY,
 subject_id TEXT NOT NULL CHECK(subject_id='owner'),
 scopes_json TEXT NOT NULL,
 expires_at_utc TEXT NOT NULL,
 adult_attested_at_utc TEXT NOT NULL,
 revoked_at_utc TEXT,
 created_at_utc TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS private_memory (
 id TEXT PRIMARY KEY,
 subject_id TEXT NOT NULL CHECK(subject_id='owner'),
 origin_type TEXT NOT NULL,
 ciphertext BLOB NOT NULL,
 created_at_utc TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS private_audit (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 event_code TEXT NOT NULL CHECK(event_code IN ('CONSENT_GRANTED','CONSENT_REVOKED','MEMORY_DELETED')),
 subject_id TEXT NOT NULL CHECK(subject_id='owner'),
 created_at_utc TEXT NOT NULL
);
CREATE TRIGGER IF NOT EXISTS private_audit_no_update BEFORE UPDATE ON private_audit
BEGIN SELECT RAISE(ABORT,'private audit is append-only'); END;
CREATE TRIGGER IF NOT EXISTS private_audit_no_delete BEFORE DELETE ON private_audit
BEGIN SELECT RAISE(ABORT,'private audit is append-only'); END;
CREATE TABLE IF NOT EXISTS private_preference (
 id TEXT PRIMARY KEY,
 subject_id TEXT NOT NULL CHECK(subject_id='owner'),
 ciphertext BLOB NOT NULL,
 created_at_utc TEXT NOT NULL,
 updated_at_utc TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS preference_audit (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 event_code TEXT NOT NULL CHECK(event_code IN ('PREFERENCE_DECLARED','PREFERENCE_CORRECTED','PREFERENCE_DELETED')),
 subject_id TEXT NOT NULL CHECK(subject_id='owner'),
 created_at_utc TEXT NOT NULL
);
CREATE TRIGGER IF NOT EXISTS preference_audit_no_update BEFORE UPDATE ON preference_audit
BEGIN SELECT RAISE(ABORT,'preference audit is append-only'); END;
CREATE TRIGGER IF NOT EXISTS preference_audit_no_delete BEFORE DELETE ON preference_audit
BEGIN SELECT RAISE(ABORT,'preference audit is append-only'); END;
CREATE TABLE IF NOT EXISTS private_discovery_policy (
 subject_id TEXT PRIMARY KEY CHECK(subject_id='owner'),
 ciphertext BLOB NOT NULL,
 updated_at_utc TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS discovery_audit (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 event_code TEXT NOT NULL CHECK(event_code IN ('DISCOVERY_POLICY_UPDATED')),
 subject_id TEXT NOT NULL CHECK(subject_id='owner'),
 created_at_utc TEXT NOT NULL
);
CREATE TRIGGER IF NOT EXISTS discovery_audit_no_update BEFORE UPDATE ON discovery_audit
BEGIN SELECT RAISE(ABORT,'discovery audit is append-only'); END;
CREATE TRIGGER IF NOT EXISTS discovery_audit_no_delete BEFORE DELETE ON discovery_audit
BEGIN SELECT RAISE(ABORT,'discovery audit is append-only'); END;
CREATE TABLE IF NOT EXISTS vault_audit (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 event_code TEXT NOT NULL CHECK(event_code IN ('KEY_ROTATED')),
 subject_id TEXT NOT NULL CHECK(subject_id='owner'),
 created_at_utc TEXT NOT NULL
);
CREATE TRIGGER IF NOT EXISTS vault_audit_no_update BEFORE UPDATE ON vault_audit
BEGIN SELECT RAISE(ABORT,'vault audit is append-only'); END;
CREATE TRIGGER IF NOT EXISTS vault_audit_no_delete BEFORE DELETE ON vault_audit
BEGIN SELECT RAISE(ABORT,'vault audit is append-only'); END;
`
