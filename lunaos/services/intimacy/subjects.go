package intimacy

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Per-person consent and the private identity vault (Mont Blanc U1/U4).
//
// Every person is a subject, and starts with no permission at all. Each scope
// is granted separately, by the subject, after reading the current disclosure.
// The owner cannot consent for anyone else: a grant for another person must
// carry the subject's own proof, either a click in the subject-operated consent
// screen or a single-use spoken challenge issued to that subject and answered
// within two minutes. Intimate scopes also require the subject to acknowledge
// the disclosure that names nudity and genitals. A subject whose adulthood is
// not attested, or who is known or suspected to be a minor, can hold no
// intimate scope and no intimate asset.
//
// Assets are sealed with a data key of their own subject, itself sealed by the
// vault key. Revoking a scope stops every new use at once and erases the assets
// that depended on it in the same transaction; deleting a subject also destroys
// that subject's data key, so any copy of the ciphertext left in a WAL frame or
// an older backup can no longer be opened by this vault. Audit rows hold an
// event code and an opaque reference only.

type SubjectScope string

const (
	SubjectFaceReference            SubjectScope = "FACE_REFERENCE"
	SubjectFaceRecognition          SubjectScope = "FACE_RECOGNITION"
	SubjectBodyProfile              SubjectScope = "BODY_PROFILE"
	SubjectIntimateView             SubjectScope = "INTIMATE_VIEW"
	SubjectIntimateStillCapture     SubjectScope = "INTIMATE_STILL_CAPTURE"
	SubjectIntimateVideoCapture     SubjectScope = "INTIMATE_VIDEO_CAPTURE"
	SubjectIntimateDataRetention    SubjectScope = "INTIMATE_DATA_RETENTION"
	SubjectTrainingFace             SubjectScope = "TRAINING_FACE"
	SubjectTrainingBody             SubjectScope = "TRAINING_BODY"
	SubjectTrainingBlind            SubjectScope = "TRAINING_BLIND"
	SubjectOngoingAutonomousCapture SubjectScope = "ONGOING_AUTONOMOUS_CAPTURE"
	SubjectNASEncryptedBackup       SubjectScope = "NAS_ENCRYPTED_BACKUP"
	SubjectGenerationReference      SubjectScope = "GENERATION_REFERENCE"

	// CurrentDisclosureRevision is the consent text the subject must have read.
	// Changing that text means bumping this, which invalidates grants made
	// against the old one for new requests.
	CurrentDisclosureRevision = "2026-11-01.fr.v1"

	ProofSubjectClick    = "subject_click"
	ProofVoiceChallenge  = "voice_challenge"
	voiceChallengeWindow = 2 * time.Minute
	maxConsentDuration   = 366 * 24 * time.Hour
)

var (
	ErrSubjectNotFound  = errors.New("SUBJECT_NOT_FOUND")
	ErrSubjectConsent   = errors.New("SUBJECT_CONSENT_REQUIRED")
	ErrSubjectNotAdult  = errors.New("SUBJECT_ADULTHOOD_NOT_ATTESTED")
	ErrInvalidSubject   = errors.New("INVALID_SUBJECT_REQUEST")
	ErrChallengeInvalid = errors.New("CONSENT_CHALLENGE_INVALID")

	intimateScopes = map[SubjectScope]bool{
		SubjectBodyProfile: true, SubjectIntimateView: true, SubjectIntimateStillCapture: true,
		SubjectIntimateVideoCapture: true, SubjectIntimateDataRetention: true, SubjectTrainingBody: true,
		SubjectTrainingBlind: true, SubjectOngoingAutonomousCapture: true, SubjectGenerationReference: true,
	}
	knownScopes = map[SubjectScope]bool{
		SubjectFaceReference: true, SubjectFaceRecognition: true, SubjectTrainingFace: true, SubjectNASEncryptedBackup: true,
	}
)

func init() {
	for scope := range intimateScopes {
		knownScopes[scope] = true
	}
}

// KnownScopes lists every consent scope the vault accepts, sorted.
func KnownScopes() []SubjectScope {
	scopes := make([]SubjectScope, 0, len(knownScopes))
	for scope := range knownScopes {
		scopes = append(scopes, scope)
	}
	sort.Slice(scopes, func(i, j int) bool { return scopes[i] < scopes[j] })
	return scopes
}

type Subject struct {
	ID          string `json:"subject_id"`
	Kind        string `json:"kind"`
	AdultStatus string `json:"adult_status"`
}

type SubjectConsentRequest struct {
	Scopes                 []SubjectScope `json:"scopes"`
	DisclosureRevision     string         `json:"disclosure_revision"`
	IntimateDisclosureRead bool           `json:"intimate_disclosure_acknowledged"`
	Proof                  string         `json:"proof"`
	ChallengeResponse      string         `json:"challenge_response,omitempty"`
	ExpiresAt              time.Time      `json:"expires_at"`
}

type SubjectGrant struct {
	Scope     SubjectScope `json:"scope"`
	ExpiresAt time.Time    `json:"expires_at"`
}

type ErasureReceipt struct {
	SubjectID      string `json:"subject_id"`
	Reason         string `json:"reason"`
	AssetsErased   int64  `json:"assets_erased"`
	KeyDestroyed   bool   `json:"key_destroyed"`
	CompletedAtUTC string `json:"completed_at_utc"`
}

type SubjectAsset struct {
	ID        string       `json:"asset_id"`
	SubjectID string       `json:"subject_id"`
	Purpose   SubjectScope `json:"purpose"`
	Data      []byte       `json:"-"`
}

// CreateSubject registers a person. Adulthood starts unverified unless the
// trusted consent screen records the subject's own attestation; "minor" and
// "suspected_minor" are terminal for every intimate path.
func (v *Vault) CreateSubject(ctx context.Context, kind, adultStatus string) (Subject, error) {
	if (kind != "SELF" && kind != "OTHER") || !validAdultStatus(adultStatus) {
		return Subject{}, ErrInvalidSubject
	}
	id, err := randomID()
	if err != nil {
		return Subject{}, err
	}
	dataKey := make([]byte, 32)
	if _, err = rand.Read(dataKey); err != nil {
		return Subject{}, err
	}
	wrapped, err := sealWithAEAD(v.aead, dataKey, subjectKeyAAD(id))
	if err != nil {
		return Subject{}, err
	}
	now := utcNow()
	tx, err := v.db.BeginTx(ctx, nil)
	if err != nil {
		return Subject{}, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `INSERT INTO private_subject(id,kind,adult_status,created_at_utc) VALUES(?,?,?,?)`, id, kind, adultStatus, now); err != nil {
		return Subject{}, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO subject_key(subject_id,wrapped_key) VALUES(?,?)`, id, wrapped); err != nil {
		return Subject{}, err
	}
	if err = subjectAudit(ctx, tx, "SUBJECT_CREATED", id, now); err != nil {
		return Subject{}, err
	}
	if err = tx.Commit(); err != nil {
		return Subject{}, err
	}
	return Subject{ID: id, Kind: kind, AdultStatus: adultStatus}, nil
}

// IssueVoiceChallenge returns a random phrase the subject must say back. It
// is single use and expires after two minutes, so a recording of an earlier
// "OK, je valide" cannot grant anything.
func (v *Vault) IssueVoiceChallenge(ctx context.Context, subjectID string) (string, error) {
	if _, err := v.subject(ctx, v.db, subjectID); err != nil {
		return "", err
	}
	raw := make([]byte, 4)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	phrase := "OK, je valide " + hex.EncodeToString(raw)
	now := time.Now().UTC()
	_, err := v.db.ExecContext(ctx, `INSERT OR REPLACE INTO subject_challenge(subject_id,phrase,expires_at_utc) VALUES(?,?,?)`,
		subjectID, phrase, now.Add(voiceChallengeWindow).Format(time.RFC3339Nano))
	return phrase, err
}

// GrantSubjectConsent records the scopes a subject granted. The caller is the
// trusted consent screen; the HTTP layer accepts it only from the owner socket.
func (v *Vault) GrantSubjectConsent(ctx context.Context, subjectID string, request SubjectConsentRequest) ([]SubjectGrant, error) {
	if request.DisclosureRevision != CurrentDisclosureRevision || len(request.Scopes) == 0 || len(request.Scopes) > len(knownScopes) {
		return nil, ErrInvalidSubject
	}
	now := time.Now().UTC()
	if !request.ExpiresAt.After(now) || request.ExpiresAt.Sub(now) > maxConsentDuration {
		return nil, ErrInvalidSubject
	}
	tx, err := v.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	subject, err := v.subject(ctx, tx, subjectID)
	if err != nil {
		return nil, err
	}
	intimate := false
	seen := map[SubjectScope]bool{}
	for _, scope := range request.Scopes {
		if !knownScopes[scope] || seen[scope] {
			return nil, ErrInvalidSubject
		}
		seen[scope] = true
		intimate = intimate || intimateScopes[scope]
	}
	if intimate && (subject.AdultStatus != "attested_adult" || !request.IntimateDisclosureRead) {
		if subject.AdultStatus != "attested_adult" {
			return nil, ErrSubjectNotAdult
		}
		return nil, ErrSubjectConsent
	}
	switch request.Proof {
	case ProofSubjectClick:
	case ProofVoiceChallenge:
		var phrase, expires string
		err = tx.QueryRowContext(ctx, `SELECT phrase,expires_at_utc FROM subject_challenge WHERE subject_id=?`, subjectID).Scan(&phrase, &expires)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrChallengeInvalid
		}
		if err != nil {
			return nil, err
		}
		deadline, parseErr := time.Parse(time.RFC3339Nano, expires)
		// The challenge is consumed whatever the answer, so a wrong guess
		// cannot be retried against the same phrase.
		if _, err = tx.ExecContext(ctx, `DELETE FROM subject_challenge WHERE subject_id=?`, subjectID); err != nil {
			return nil, err
		}
		if parseErr != nil || !deadline.After(now) || subtle.ConstantTimeCompare([]byte(normalizePhrase(phrase)), []byte(normalizePhrase(request.ChallengeResponse))) != 1 {
			if commitErr := tx.Commit(); commitErr != nil {
				return nil, commitErr
			}
			return nil, ErrChallengeInvalid
		}
	default:
		// "owner" and anything else: nobody consents for another person.
		return nil, ErrSubjectConsent
	}
	stamp := now.Format(time.RFC3339Nano)
	grants := make([]SubjectGrant, 0, len(request.Scopes))
	for _, scope := range request.Scopes {
		if _, err = tx.ExecContext(ctx, `UPDATE subject_consent SET revoked_at_utc=? WHERE subject_id=? AND scope=? AND revoked_at_utc IS NULL`, stamp, subjectID, string(scope)); err != nil {
			return nil, err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO subject_consent(subject_id,scope,disclosure_revision,proof_method,granted_at_utc,expires_at_utc) VALUES(?,?,?,?,?,?)`,
			subjectID, string(scope), request.DisclosureRevision, request.Proof, stamp, request.ExpiresAt.UTC().Format(time.RFC3339Nano)); err != nil {
			return nil, err
		}
		grants = append(grants, SubjectGrant{Scope: scope, ExpiresAt: request.ExpiresAt.UTC()})
	}
	if err = subjectAudit(ctx, tx, "CONSENT_GRANTED", subjectID, stamp); err != nil {
		return nil, err
	}
	return grants, tx.Commit()
}

// ActiveScopes lists the scopes a subject currently grants.
func (v *Vault) ActiveScopes(ctx context.Context, subjectID string) ([]SubjectGrant, error) {
	if _, err := v.subject(ctx, v.db, subjectID); err != nil {
		return nil, err
	}
	rows, err := v.db.QueryContext(ctx, `SELECT scope,expires_at_utc FROM subject_consent WHERE subject_id=? AND revoked_at_utc IS NULL AND expires_at_utc>? ORDER BY scope`,
		subjectID, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	grants := []SubjectGrant{}
	for rows.Next() {
		var scope, expires string
		if err = rows.Scan(&scope, &expires); err != nil {
			return nil, err
		}
		deadline, err := time.Parse(time.RFC3339Nano, expires)
		if err != nil {
			return nil, err
		}
		grants = append(grants, SubjectGrant{Scope: SubjectScope(scope), ExpiresAt: deadline})
	}
	return grants, rows.Err()
}

// StoreSubjectAsset keeps data about a subject for one consented purpose.
// Intimate purposes also need INTIMATE_DATA_RETENTION.
func (v *Vault) StoreSubjectAsset(ctx context.Context, subjectID string, purpose SubjectScope, data []byte) (SubjectAsset, error) {
	if !knownScopes[purpose] || len(data) == 0 || len(data) > 64<<20 {
		return SubjectAsset{}, ErrInvalidSubject
	}
	tx, err := v.db.BeginTx(ctx, nil)
	if err != nil {
		return SubjectAsset{}, err
	}
	defer tx.Rollback()
	subject, err := v.subject(ctx, tx, subjectID)
	if err != nil {
		return SubjectAsset{}, err
	}
	if intimateScopes[purpose] && subject.AdultStatus != "attested_adult" {
		return SubjectAsset{}, ErrSubjectNotAdult
	}
	required := []SubjectScope{purpose}
	if intimateScopes[purpose] && purpose != SubjectIntimateDataRetention {
		required = append(required, SubjectIntimateDataRetention)
	}
	for _, scope := range required {
		if err = activeScope(ctx, tx, subjectID, scope); err != nil {
			return SubjectAsset{}, err
		}
	}
	aead, err := v.subjectAEAD(ctx, tx, subjectID)
	if err != nil {
		return SubjectAsset{}, err
	}
	id, err := randomID()
	if err != nil {
		return SubjectAsset{}, err
	}
	sealed, err := sealWithAEAD(aead, data, subjectAssetAAD(id, subjectID, purpose))
	if err != nil {
		return SubjectAsset{}, err
	}
	now := utcNow()
	if _, err = tx.ExecContext(ctx, `INSERT INTO subject_asset(id,subject_id,purpose,ciphertext,created_at_utc) VALUES(?,?,?,?,?)`, id, subjectID, string(purpose), sealed, now); err != nil {
		return SubjectAsset{}, err
	}
	if err = subjectAudit(ctx, tx, "ASSET_STORED", subjectID, now); err != nil {
		return SubjectAsset{}, err
	}
	return SubjectAsset{ID: id, SubjectID: subjectID, Purpose: purpose}, tx.Commit()
}

// ReadSubjectAsset opens an asset only while its purpose is still consented.
func (v *Vault) ReadSubjectAsset(ctx context.Context, subjectID, assetID string) (SubjectAsset, error) {
	var purpose string
	var sealed []byte
	err := v.db.QueryRowContext(ctx, `SELECT purpose,ciphertext FROM subject_asset WHERE id=? AND subject_id=?`, assetID, subjectID).Scan(&purpose, &sealed)
	if errors.Is(err, sql.ErrNoRows) {
		return SubjectAsset{}, ErrSubjectNotFound
	}
	if err != nil {
		return SubjectAsset{}, err
	}
	if err = activeScope(ctx, v.db, subjectID, SubjectScope(purpose)); err != nil {
		return SubjectAsset{}, err
	}
	aead, err := v.subjectAEAD(ctx, v.db, subjectID)
	if err != nil {
		return SubjectAsset{}, err
	}
	data, err := openWithAEAD(aead, sealed, subjectAssetAAD(assetID, subjectID, SubjectScope(purpose)))
	if err != nil {
		return SubjectAsset{}, ErrSubjectNotFound
	}
	return SubjectAsset{ID: assetID, SubjectID: subjectID, Purpose: SubjectScope(purpose), Data: data}, nil
}

// RevokeSubjectScopes ends the listed scopes now and erases the assets that
// depended on them. Withdrawing INTIMATE_DATA_RETENTION erases every intimate
// asset of the subject.
func (v *Vault) RevokeSubjectScopes(ctx context.Context, subjectID string, scopes []SubjectScope) (ErasureReceipt, error) {
	if len(scopes) == 0 {
		return ErasureReceipt{}, ErrInvalidSubject
	}
	tx, err := v.db.BeginTx(ctx, nil)
	if err != nil {
		return ErasureReceipt{}, err
	}
	defer tx.Rollback()
	if _, err = v.subject(ctx, tx, subjectID); err != nil {
		return ErasureReceipt{}, err
	}
	now := utcNow()
	receipt := ErasureReceipt{SubjectID: subjectID, Reason: "CONSENT_REVOKED", CompletedAtUTC: now}
	for _, scope := range scopes {
		if !knownScopes[scope] {
			return ErasureReceipt{}, ErrInvalidSubject
		}
		if _, err = tx.ExecContext(ctx, `UPDATE subject_consent SET revoked_at_utc=? WHERE subject_id=? AND scope=? AND revoked_at_utc IS NULL`, now, subjectID, string(scope)); err != nil {
			return ErasureReceipt{}, err
		}
		query, args := `DELETE FROM subject_asset WHERE subject_id=? AND purpose=?`, []any{subjectID, string(scope)}
		if scope == SubjectIntimateDataRetention {
			placeholders := make([]string, 0, len(intimateScopes))
			args = []any{subjectID}
			for intimate := range intimateScopes {
				placeholders = append(placeholders, "?")
				args = append(args, string(intimate))
			}
			query = `DELETE FROM subject_asset WHERE subject_id=? AND purpose IN (` + strings.Join(placeholders, ",") + `)`
		}
		result, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return ErasureReceipt{}, err
		}
		erased, _ := result.RowsAffected()
		receipt.AssetsErased += erased
	}
	if err = subjectAudit(ctx, tx, "CONSENT_REVOKED", subjectID, now); err != nil {
		return ErasureReceipt{}, err
	}
	return receipt, tx.Commit()
}

// DeleteSubject erases everything about a person, including the data key, and
// leaves only a minimal audit receipt.
func (v *Vault) DeleteSubject(ctx context.Context, subjectID string) (ErasureReceipt, error) {
	tx, err := v.db.BeginTx(ctx, nil)
	if err != nil {
		return ErasureReceipt{}, err
	}
	defer tx.Rollback()
	if _, err = v.subject(ctx, tx, subjectID); err != nil {
		return ErasureReceipt{}, err
	}
	now := utcNow()
	result, err := tx.ExecContext(ctx, `DELETE FROM subject_asset WHERE subject_id=?`, subjectID)
	if err != nil {
		return ErasureReceipt{}, err
	}
	erased, _ := result.RowsAffected()
	for _, statement := range []string{
		`DELETE FROM subject_challenge WHERE subject_id=?`,
		`DELETE FROM subject_consent WHERE subject_id=?`,
		`DELETE FROM subject_key WHERE subject_id=?`,
		`DELETE FROM private_subject WHERE id=?`,
	} {
		if _, err = tx.ExecContext(ctx, statement, subjectID); err != nil {
			return ErasureReceipt{}, err
		}
	}
	if err = subjectAudit(ctx, tx, "SUBJECT_ERASED", subjectID, now); err != nil {
		return ErasureReceipt{}, err
	}
	if err = tx.Commit(); err != nil {
		return ErasureReceipt{}, err
	}
	return ErasureReceipt{SubjectID: subjectID, Reason: "SUBJECT_DELETED", AssetsErased: erased, KeyDestroyed: true, CompletedAtUTC: now}, nil
}

// FlagSuspectedMinor is for a detector or the owner: it can only lower a
// subject's status, and it ends every intimate scope and erases every intimate
// asset of that subject at once.
func (v *Vault) FlagSuspectedMinor(ctx context.Context, subjectID string) (ErasureReceipt, error) {
	scopes := make([]SubjectScope, 0, len(intimateScopes))
	for scope := range intimateScopes {
		scopes = append(scopes, scope)
	}
	receipt, err := v.RevokeSubjectScopes(ctx, subjectID, scopes)
	if err != nil {
		return ErasureReceipt{}, err
	}
	now := utcNow()
	tx, err := v.db.BeginTx(ctx, nil)
	if err != nil {
		return ErasureReceipt{}, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `UPDATE private_subject SET adult_status='suspected_minor' WHERE id=? AND adult_status!='minor'`, subjectID); err != nil {
		return ErasureReceipt{}, err
	}
	if err = subjectAudit(ctx, tx, "SUBJECT_MINOR_FLAGGED", subjectID, now); err != nil {
		return ErasureReceipt{}, err
	}
	receipt.Reason = "SUSPECTED_MINOR"
	return receipt, tx.Commit()
}

type rowQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (v *Vault) subject(ctx context.Context, q rowQuerier, id string) (Subject, error) {
	var subject Subject
	err := q.QueryRowContext(ctx, `SELECT id,kind,adult_status FROM private_subject WHERE id=?`, id).Scan(&subject.ID, &subject.Kind, &subject.AdultStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return Subject{}, ErrSubjectNotFound
	}
	return subject, err
}

func (v *Vault) subjectAEAD(ctx context.Context, q rowQuerier, subjectID string) (cipher.AEAD, error) {
	var wrapped []byte
	err := q.QueryRowContext(ctx, `SELECT wrapped_key FROM subject_key WHERE subject_id=?`, subjectID).Scan(&wrapped)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSubjectNotFound
	}
	if err != nil {
		return nil, err
	}
	dataKey, err := openWithAEAD(v.aead, wrapped, subjectKeyAAD(subjectID))
	if err != nil {
		return nil, fmt.Errorf("subject key unavailable: %w", err)
	}
	block, err := aes.NewCipher(dataKey)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func activeScope(ctx context.Context, q rowQuerier, subjectID string, scope SubjectScope) error {
	var granted int
	err := q.QueryRowContext(ctx, `SELECT count(*) FROM subject_consent WHERE subject_id=? AND scope=? AND revoked_at_utc IS NULL AND expires_at_utc>?`,
		subjectID, string(scope), time.Now().UTC().Format(time.RFC3339Nano)).Scan(&granted)
	if err != nil {
		return err
	}
	if granted == 0 {
		return ErrSubjectConsent
	}
	return nil
}

func subjectAudit(ctx context.Context, tx *sql.Tx, code, subjectID, now string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO subject_audit(event_code,subject_ref,created_at_utc) VALUES(?,?,?)`, code, subjectID, now)
	return err
}

func validAdultStatus(status string) bool {
	switch status {
	case "unverified", "attested_adult", "suspected_minor", "minor":
		return true
	}
	return false
}

func normalizePhrase(phrase string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.ReplaceAll(phrase, ",", " "))), " ")
}

func subjectKeyAAD(subjectID string) []byte { return []byte("lunaos-subject-key:" + subjectID) }

func subjectAssetAAD(id, subjectID string, purpose SubjectScope) []byte {
	return []byte("lunaos-subject-asset:" + id + ":" + subjectID + ":" + string(purpose))
}

func utcNow() string { return time.Now().UTC().Format(time.RFC3339Nano) }
