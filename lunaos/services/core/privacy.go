package core

import (
	"context"
	"crypto/hmac"
	"database/sql"
	"errors"
	"fmt"
	"hash/crc32"
)

// The privacy controller is ordinary, deterministic code: no model output can
// reach it. BLIND is a latch. Anything may close LUNA's eyes (the owner, a
// detector, a failing sensor), and closing them again is harmless; only an
// explicit action from the trusted local UI opens them, and nothing reopens
// them because the scene changed or the machine rebooted.
//
// Capture is leased. A sensor broker asks for a lease before reading a frame
// and checks it before keeping anything; every entry into BLIND bumps the
// capture epoch, so leases issued before it stop validating at once.

const (
	PrivacyStartupLocked     = "STARTUP_LOCKED"
	PrivacyNormal            = "NORMAL_EPHEMERAL_DETECTION"
	PrivacyAuthorizedCapture = "AUTHORIZED_CAPTURE_SESSION"
	PrivacyBlind             = "BLIND_LATCHED"
	PrivacySensorUnavailable = "SENSOR_UNAVAILABLE"
)

var (
	ErrPrivacyBlocked     = errors.New("PRIVACY_BLOCKED")
	ErrTrustedUIRequired  = errors.New("TRUSTED_UI_REQUIRED")
	privacyBlindReasons   = map[string]string{"manual": "MANUAL", "intimate_scene": "INTIMATE_SCENE", "sensor_error": "SENSOR_ERROR", "consent_revoked": "CONSENT_REVOKED"}
	privacyLeasePurposes  = map[string]bool{"blind_detection": true}
	privacyCapturePurpose = "consented_capture"
)

// PrivacyState is what any caller may see: the mode and whether capture is
// possible. It never says why BLIND was entered.
type PrivacyState struct {
	Mode            string `json:"mode"`
	CaptureAllowed  bool   `json:"capture_allowed"`
	RequiresOpenEye bool   `json:"requires_open_eyes"`
	UpdatedAtUTC    string `json:"updated_at_utc"`
}

// CaptureLease lets a sensor broker read frames for one purpose until the
// next entry into BLIND.
type CaptureLease struct {
	Purpose string `json:"purpose"`
	Epoch   int64  `json:"epoch"`
	Check   uint32 `json:"check"`
}

func privacyState(mode, updated string) PrivacyState {
	return PrivacyState{
		Mode:            mode,
		CaptureAllowed:  mode == PrivacyNormal || mode == PrivacyAuthorizedCapture,
		RequiresOpenEye: mode == PrivacyBlind || mode == PrivacyStartupLocked || mode == PrivacySensorUnavailable,
		UpdatedAtUTC:    updated,
	}
}

func (e *Engine) PrivacyStatus(ctx context.Context) (PrivacyState, error) {
	var mode, updated string
	if err := e.db.QueryRowContext(ctx, `SELECT mode,updated_at_utc FROM privacy_control WHERE singleton=1`).Scan(&mode, &updated); err != nil {
		return PrivacyState{}, err
	}
	return privacyState(mode, updated), nil
}

// EnterBlind latches BLIND and invalidates every capture lease. It is
// idempotent and records only a coarse reason code.
func (e *Engine) EnterBlind(ctx context.Context, reason string) (PrivacyState, error) {
	code, ok := privacyBlindReasons[reason]
	if !ok {
		return PrivacyState{}, fmt.Errorf("%w: INVALID_BLIND_REASON", ErrInvalid)
	}
	return e.setPrivacyMode(ctx, PrivacyBlind, "privacy.blind_entered", code, true)
}

// OpenEyes leaves BLIND (or the startup lock) for ephemeral detection. Only the
// trusted local UI may call it: the HTTP layer passes trustedUI only for a
// request that arrived on the owner's Unix socket.
func (e *Engine) OpenEyes(ctx context.Context, ownerToken string, trustedUI bool) (PrivacyState, error) {
	if !e.owner(ownerToken) {
		return PrivacyState{}, ErrUnauthorized
	}
	if !trustedUI {
		return PrivacyState{}, ErrTrustedUIRequired
	}
	return e.setPrivacyMode(ctx, PrivacyNormal, "privacy.blind_exited", "OWNER_OPENED_EYES", false)
}

func (e *Engine) setPrivacyMode(ctx context.Context, mode, kind, code string, bumpEpoch bool) (PrivacyState, error) {
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return PrivacyState{}, err
	}
	defer tx.Rollback()
	var current string
	if err = tx.QueryRowContext(ctx, `SELECT mode FROM privacy_control WHERE singleton=1`).Scan(&current); err != nil {
		return PrivacyState{}, err
	}
	now := timestamp()
	epochStep := 0
	if bumpEpoch {
		epochStep = 1
	}
	if current != mode || bumpEpoch {
		if _, err = tx.ExecContext(ctx, `UPDATE privacy_control SET mode=?,capture_epoch=capture_epoch+?,updated_at_utc=? WHERE singleton=1`, mode, epochStep, now); err != nil {
			return PrivacyState{}, err
		}
	}
	if current != mode {
		if _, err = tx.ExecContext(ctx, `INSERT INTO privacy_event(kind,code,created_at_utc) VALUES(?,?,?)`, kind, code, now); err != nil {
			return PrivacyState{}, err
		}
	}
	var updated string
	if err = tx.QueryRowContext(ctx, `SELECT updated_at_utc FROM privacy_control WHERE singleton=1`).Scan(&updated); err != nil {
		return PrivacyState{}, err
	}
	if err = tx.Commit(); err != nil {
		return PrivacyState{}, err
	}
	return privacyState(mode, updated), nil
}

// RequestCaptureLease issues a lease only while capture is allowed for the
// purpose: ephemeral detection in normal mode, consented capture only inside
// an authorized capture session.
func (e *Engine) RequestCaptureLease(ctx context.Context, ownerToken, purpose string) (CaptureLease, error) {
	if !e.owner(ownerToken) {
		return CaptureLease{}, ErrUnauthorized
	}
	if !privacyLeasePurposes[purpose] && purpose != privacyCapturePurpose {
		return CaptureLease{}, fmt.Errorf("%w: INVALID_CAPTURE_PURPOSE", ErrInvalid)
	}
	var mode string
	var epoch int64
	if err := e.db.QueryRowContext(ctx, `SELECT mode,capture_epoch FROM privacy_control WHERE singleton=1`).Scan(&mode, &epoch); err != nil {
		return CaptureLease{}, err
	}
	switch {
	case purpose == privacyCapturePurpose && mode == PrivacyAuthorizedCapture:
	case purpose == "blind_detection" && (mode == PrivacyNormal || mode == PrivacyAuthorizedCapture):
	default:
		return CaptureLease{}, ErrPrivacyBlocked
	}
	return CaptureLease{Purpose: purpose, Epoch: epoch, Check: leaseCheck(purpose, epoch)}, nil
}

// LeaseValid reports whether a lease still permits capture right now.
func (e *Engine) LeaseValid(ctx context.Context, lease CaptureLease) (bool, error) {
	if lease.Check != leaseCheck(lease.Purpose, lease.Epoch) {
		return false, nil
	}
	var mode string
	var epoch int64
	err := e.db.QueryRowContext(ctx, `SELECT mode,capture_epoch FROM privacy_control WHERE singleton=1`).Scan(&mode, &epoch)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if epoch != lease.Epoch {
		return false, nil
	}
	if lease.Purpose == privacyCapturePurpose {
		return mode == PrivacyAuthorizedCapture, nil
	}
	return mode == PrivacyNormal || mode == PrivacyAuthorizedCapture, nil
}

// leaseCheck catches a malformed lease; authority comes from the epoch and the
// owner credential, not from this checksum.
func leaseCheck(purpose string, epoch int64) uint32 {
	return crc32.ChecksumIEEE([]byte(fmt.Sprintf("%s:%d", purpose, epoch)))
}

func (e *Engine) owner(token string) bool {
	return hmac.Equal([]byte(token), []byte(e.ownerToken))
}
