package intimacy

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// RotateKey atomically re-encrypts every private payload with a caller-supplied
// 32-byte key. The caller must securely persist the new key before invoking this
// method and securely retire the old key only after a verified backup exists.
func (v *Vault) RotateKey(newKey []byte, ownerConfirmed bool) error {
	if !ownerConfirmed {
		return ErrVaultRotationUnconfirmed
	}
	if len(newKey) != 32 {
		return ErrInvalidKey
	}
	block, err := aes.NewCipher(append([]byte(nil), newKey...))
	if err != nil {
		return ErrInvalidKey
	}
	newAEAD, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	ctx := context.Background()
	tx, err := v.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = rotateMemoryRows(ctx, tx, v.aead, newAEAD); err != nil {
		return err
	}
	if err = rotatePreferenceRows(ctx, tx, v.aead, newAEAD); err != nil {
		return err
	}
	if err = rotateDiscoveryPolicy(ctx, tx, v.aead, newAEAD); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err = tx.ExecContext(ctx, `INSERT INTO vault_audit(event_code,subject_id,created_at_utc) VALUES('KEY_ROTATED','owner',?)`, now); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	v.aead = newAEAD
	v.keyFingerprint = sha256.Sum256(newKey)
	return nil
}

type encryptedRow struct {
	id         string
	origin     Origin
	ciphertext []byte
}

func rotateMemoryRows(ctx context.Context, tx *sql.Tx, oldAEAD, newAEAD cipher.AEAD) error {
	lastID := ""
	for {
		rows, err := tx.QueryContext(ctx, `SELECT id,origin_type,ciphertext FROM private_memory WHERE id>? ORDER BY id LIMIT 64`, lastID)
		if err != nil {
			return err
		}
		batch := make([]encryptedRow, 0, 64)
		for rows.Next() {
			var row encryptedRow
			if err = rows.Scan(&row.id, &row.origin, &row.ciphertext); err != nil {
				_ = rows.Close()
				return err
			}
			batch = append(batch, row)
		}
		if err = rows.Err(); err != nil {
			_ = rows.Close()
			return err
		}
		if err = rows.Close(); err != nil {
			return err
		}
		if len(batch) == 0 {
			return nil
		}
		for _, row := range batch {
			lastID = row.id
			if !validOrigin(row.origin) {
				return ErrMemoryNotFound
			}
			plain, openErr := openWithAEAD(oldAEAD, row.ciphertext, associatedData(row.id, "owner", row.origin))
			if openErr != nil {
				return ErrMemoryNotFound
			}
			ciphertext, sealErr := sealWithAEAD(newAEAD, plain, associatedData(row.id, "owner", row.origin))
			if sealErr != nil {
				return sealErr
			}
			if _, err = tx.ExecContext(ctx, `UPDATE private_memory SET ciphertext=? WHERE id=? AND subject_id='owner'`, ciphertext, row.id); err != nil {
				return err
			}
		}
	}
}

func rotatePreferenceRows(ctx context.Context, tx *sql.Tx, oldAEAD, newAEAD cipher.AEAD) error {
	lastID := ""
	for {
		rows, err := tx.QueryContext(ctx, `SELECT id,ciphertext FROM private_preference WHERE id>? ORDER BY id LIMIT 64`, lastID)
		if err != nil {
			return err
		}
		batch := make([]encryptedRow, 0, 64)
		for rows.Next() {
			var row encryptedRow
			if err = rows.Scan(&row.id, &row.ciphertext); err != nil {
				_ = rows.Close()
				return err
			}
			batch = append(batch, row)
		}
		if err = rows.Err(); err != nil {
			_ = rows.Close()
			return err
		}
		if err = rows.Close(); err != nil {
			return err
		}
		if len(batch) == 0 {
			return nil
		}
		for _, row := range batch {
			lastID = row.id
			aad := associatedData(row.id, "owner", Origin("preferences.v1"))
			plain, openErr := openWithAEAD(oldAEAD, row.ciphertext, aad)
			if openErr != nil {
				return ErrPreferenceNotFound
			}
			var preference Preference
			if json.Unmarshal(plain, &preference) != nil || preference.ID != row.id || preference.SourceType != "declared" || preference.ConsentScope != "local_only" {
				return ErrPreferenceNotFound
			}
			ciphertext, sealErr := sealWithAEAD(newAEAD, plain, aad)
			if sealErr != nil {
				return sealErr
			}
			if _, err = tx.ExecContext(ctx, `UPDATE private_preference SET ciphertext=? WHERE id=? AND subject_id='owner'`, ciphertext, row.id); err != nil {
				return err
			}
		}
	}
}

func rotateDiscoveryPolicy(ctx context.Context, tx *sql.Tx, oldAEAD, newAEAD cipher.AEAD) error {
	var ciphertext []byte
	err := tx.QueryRowContext(ctx, `SELECT ciphertext FROM private_discovery_policy WHERE subject_id='owner'`).Scan(&ciphertext)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	plain, err := openWithAEAD(oldAEAD, ciphertext, discoveryPolicyAAD())
	if err != nil {
		return ErrInvalidDiscoveryAllowlist
	}
	var policy DiscoveryAllowlist
	if json.Unmarshal(plain, &policy) != nil || !validPersistedPolicy(policy) {
		return ErrInvalidDiscoveryAllowlist
	}
	rotated, err := sealWithAEAD(newAEAD, plain, discoveryPolicyAAD())
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE private_discovery_policy SET ciphertext=?,updated_at_utc=? WHERE subject_id='owner'`, rotated, time.Now().UTC().Format(time.RFC3339Nano))
	return err
}

func sealWithAEAD(aead cipher.AEAD, plain, aad []byte) ([]byte, error) {
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return aead.Seal(nonce, nonce, plain, aad), nil
}

func openWithAEAD(aead cipher.AEAD, ciphertext, aad []byte) ([]byte, error) {
	if len(ciphertext) < aead.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, body := ciphertext[:aead.NonceSize()], ciphertext[aead.NonceSize():]
	return aead.Open(nil, nonce, body, aad)
}
