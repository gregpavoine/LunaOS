package intimacy

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"time"
)

type PreferenceInput struct {
	Dimension string
	Value     string
	Strength  float64
	Timescale string
	Confirmed bool
}

type Preference struct {
	ID                    string  `json:"preference_id"`
	Dimension             string  `json:"dimension"`
	Value                 string  `json:"value"`
	SourceType            string  `json:"source_type"`
	Confidence            float64 `json:"confidence"`
	Strength              float64 `json:"strength"`
	Timescale             string  `json:"timescale"`
	FirstSeenAt           string  `json:"first_seen_at"`
	LastSeenAt            string  `json:"last_seen_at"`
	EvidenceCount         int     `json:"evidence_count"`
	NegativeFeedbackCount int     `json:"negative_feedback_count"`
	ConsentScope          string  `json:"consent_scope"`
}

type rowQueryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

var preferenceDimensions = map[string]bool{
	"content_category": true, "consensual_activity": true, "fictional_scenario": true,
	"clothing_style": true, "visual_style": true, "body_descriptor": true,
	"format": true, "composition": true, "sound": true, "mood": true,
}

var preferenceTimescales = map[string]bool{
	"persistent": true, "recurring": true, "occasional": true,
}

func (v *Vault) AddPreference(ctx context.Context, consentID string, input PreferenceInput) (Preference, error) {
	if !input.Confirmed || !validPreferenceInput(input) {
		return Preference{}, ErrInvalidPreference
	}
	tx, err := v.db.BeginTx(ctx, nil)
	if err != nil {
		return Preference{}, err
	}
	defer tx.Rollback()
	if err = authorizeQuery(ctx, tx, consentID, "owner", ScopePreferenceWrite); err != nil {
		return Preference{}, err
	}
	id, err := randomID()
	if err != nil {
		return Preference{}, err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	preference := Preference{
		ID: id, Dimension: input.Dimension, Value: input.Value, SourceType: "declared",
		Confidence: 1, Strength: input.Strength, Timescale: input.Timescale,
		FirstSeenAt: now, LastSeenAt: now, EvidenceCount: 1,
		NegativeFeedbackCount: 0, ConsentScope: "local_only",
	}
	ciphertext, err := v.sealPreference(preference)
	if err != nil {
		return Preference{}, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO private_preference(id,subject_id,ciphertext,created_at_utc,updated_at_utc) VALUES(?,'owner',?,?,?)`, id, ciphertext, now, now); err != nil {
		return Preference{}, err
	}
	if err = auditPreference(ctx, tx, "PREFERENCE_DECLARED", now); err != nil {
		return Preference{}, err
	}
	if err = tx.Commit(); err != nil {
		return Preference{}, err
	}
	return preference, nil
}

func (v *Vault) ListPreferences(ctx context.Context, consentID string) ([]Preference, error) {
	tx, err := v.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err = authorizeQuery(ctx, tx, consentID, "owner", ScopePreferenceRead); err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT id FROM private_preference WHERE subject_id='owner' ORDER BY created_at_utc,id`)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			_ = rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	if err = rows.Close(); err != nil {
		return nil, err
	}
	preferences := make([]Preference, 0, len(ids))
	for _, id := range ids {
		preference, readErr := v.readPreferenceQuery(ctx, tx, id)
		if readErr != nil {
			return nil, readErr
		}
		preferences = append(preferences, preference)
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return preferences, nil
}

func (v *Vault) CorrectPreference(ctx context.Context, id, consentID, value string, strength float64, disagreed, confirmed bool) (Preference, error) {
	if !confirmed || len(value) == 0 || len(value) > 256 || math.IsNaN(strength) || math.IsInf(strength, 0) || strength < 0 || strength > 1 {
		return Preference{}, ErrInvalidPreference
	}
	tx, err := v.db.BeginTx(ctx, nil)
	if err != nil {
		return Preference{}, err
	}
	defer tx.Rollback()
	if err = authorizeQuery(ctx, tx, consentID, "owner", ScopePreferenceWrite); err != nil {
		return Preference{}, err
	}
	preference, err := v.readPreferenceQuery(ctx, tx, id)
	if err != nil {
		return Preference{}, err
	}
	preference.Value = value
	preference.SourceType = "declared"
	preference.Confidence = 1
	preference.Strength = strength
	preference.EvidenceCount++
	if disagreed {
		preference.NegativeFeedbackCount++
	}
	preference.LastSeenAt = time.Now().UTC().Format(time.RFC3339Nano)
	ciphertext, err := v.sealPreference(preference)
	if err != nil {
		return Preference{}, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE private_preference SET ciphertext=?,updated_at_utc=? WHERE id=? AND subject_id='owner'`, ciphertext, preference.LastSeenAt, id); err != nil {
		return Preference{}, err
	}
	if err = auditPreference(ctx, tx, "PREFERENCE_CORRECTED", preference.LastSeenAt); err != nil {
		return Preference{}, err
	}
	if err = tx.Commit(); err != nil {
		return Preference{}, err
	}
	return preference, nil
}

func (v *Vault) DeletePreference(ctx context.Context, id, subjectID string) error {
	if subjectID != "owner" {
		return ErrPreferenceNotFound
	}
	tx, err := v.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `DELETE FROM private_preference WHERE id=? AND subject_id='owner'`, id)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed != 1 {
		return ErrPreferenceNotFound
	}
	if err = auditPreference(ctx, tx, "PREFERENCE_DELETED", time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		return err
	}
	return tx.Commit()
}

func (v *Vault) readPreferenceQuery(ctx context.Context, queryer rowQueryer, id string) (Preference, error) {
	var ciphertext []byte
	err := queryer.QueryRowContext(ctx, `SELECT ciphertext FROM private_preference WHERE id=? AND subject_id='owner'`, id).Scan(&ciphertext)
	if errors.Is(err, sql.ErrNoRows) {
		return Preference{}, ErrPreferenceNotFound
	}
	if err != nil {
		return Preference{}, err
	}
	plain, err := v.openPreference(id, ciphertext)
	if err != nil {
		return Preference{}, ErrPreferenceNotFound
	}
	var preference Preference
	if json.Unmarshal(plain, &preference) != nil || preference.ID != id || preference.SourceType != "declared" || preference.ConsentScope != "local_only" {
		return Preference{}, ErrPreferenceNotFound
	}
	return preference, nil
}

func (v *Vault) sealPreference(preference Preference) ([]byte, error) {
	plain, err := json.Marshal(preference)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, v.aead.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}
	return v.aead.Seal(nonce, nonce, plain, associatedData(preference.ID, "owner", Origin("preferences.v1"))), nil
}

func (v *Vault) openPreference(id string, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < v.aead.NonceSize() {
		return nil, ErrPreferenceNotFound
	}
	nonce, body := ciphertext[:v.aead.NonceSize()], ciphertext[v.aead.NonceSize():]
	return v.aead.Open(nil, nonce, body, associatedData(id, "owner", Origin("preferences.v1")))
}

func authorizeQuery(ctx context.Context, queryer rowQueryer, consentID, subjectID string, scope Scope) error {
	if subjectID != "owner" {
		return ErrConsentDenied
	}
	var scopesJSON, expires string
	var revoked sql.NullString
	err := queryer.QueryRowContext(ctx, `SELECT scopes_json,expires_at_utc,revoked_at_utc FROM private_consent WHERE id=? AND subject_id=?`, consentID, subjectID).Scan(&scopesJSON, &expires, &revoked)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrConsentDenied
	}
	if err != nil {
		return err
	}
	deadline, err := time.Parse(time.RFC3339Nano, expires)
	if err != nil || revoked.Valid || !deadline.After(time.Now()) {
		return ErrConsentDenied
	}
	var scopes []Scope
	if json.Unmarshal([]byte(scopesJSON), &scopes) != nil {
		return ErrConsentDenied
	}
	for _, granted := range scopes {
		if granted == scope {
			return nil
		}
	}
	return ErrConsentDenied
}

func validPreferenceInput(input PreferenceInput) bool {
	return preferenceDimensions[input.Dimension] && len(input.Value) > 0 && len(input.Value) <= 256 && !math.IsNaN(input.Strength) && !math.IsInf(input.Strength, 0) && input.Strength >= 0 && input.Strength <= 1 && preferenceTimescales[input.Timescale]
}

func auditPreference(ctx context.Context, tx *sql.Tx, code, timestamp string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO preference_audit(event_code,subject_id,created_at_utc) VALUES(?,'owner',?)`, code, timestamp)
	return err
}
