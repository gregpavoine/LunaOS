package core

import (
	"context"
	"fmt"
)

// The intimacy context says whether explicit material may appear: "private"
// only when the owner says so from the trusted local UI, "public" whenever
// anyone (an agent, a screen-share detector) says so, and "unknown" counts as
// not private. Lifting STOP_INTIMACY is likewise reserved to the trusted UI,
// while anyone may stop.

// SetIntimacyContext changes the privacy context. Raising it to private needs
// the trusted UI; lowering it to public never does.
func (e *Engine) SetIntimacyContext(ctx context.Context, mode string, trustedUI bool) (IntimacyState, error) {
	switch mode {
	case "public":
	case "private":
		if !trustedUI {
			return IntimacyState{}, ErrTrustedUIRequired
		}
	default:
		return IntimacyState{}, fmt.Errorf("%w: INVALID_PRIVACY_MODE", ErrInvalid)
	}
	if _, err := e.db.ExecContext(ctx, `UPDATE intimacy_control SET privacy_mode=?,updated_at_utc=? WHERE singleton=1`, mode, timestamp()); err != nil {
		return IntimacyState{}, err
	}
	return e.IntimacyState(ctx)
}

// ResumeIntimacy lifts STOP_INTIMACY from the trusted UI only.
func (e *Engine) ResumeIntimacy(ctx context.Context, trustedUI bool) (IntimacyState, error) {
	if !trustedUI {
		return IntimacyState{}, ErrTrustedUIRequired
	}
	if _, err := e.db.ExecContext(ctx, `UPDATE intimacy_control SET stopped=0,updated_at_utc=? WHERE singleton=1`, timestamp()); err != nil {
		return IntimacyState{}, err
	}
	return e.IntimacyState(ctx)
}

// ExplicitContentAllowed is true only in a private context with intimacy not stopped.
func (e *Engine) ExplicitContentAllowed(ctx context.Context) (bool, error) {
	state, err := e.IntimacyState(ctx)
	if err != nil {
		return false, err
	}
	return !state.Stopped && state.PrivacyMode == "private", nil
}
