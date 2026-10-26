package core

import (
	"context"
	"errors"
)

var (
	ErrUnauthorized        = errors.New("OWNER_AUTH_REQUIRED")
	ErrIdempotencyConflict = errors.New("IDEMPOTENCY_CONFLICT")
	ErrNotFound            = errors.New("GOAL_NOT_FOUND")
	ErrInvalid             = errors.New("INVALID_REQUEST")
	ErrTransitionConflict  = errors.New("GOAL_TRANSITION_CONFLICT")
)

type Metrics struct {
	Impact     *float64 `json:"impact"`
	Urgency    *float64 `json:"urgency"`
	UserValue  *float64 `json:"user_value"`
	Confidence *float64 `json:"confidence"`
	Efficiency *float64 `json:"efficiency"`
}

type Assertion struct {
	Claim            string  `json:"claim"`
	Source           *string `json:"source"`
	Confidence       float64 `json:"confidence"`
	ProvenanceStatus string  `json:"provenance_status"`
	EpistemicStatus  string  `json:"epistemic_status"`
}

type Request struct {
	Action            string     `json:"action"`
	PrivacyClass      string     `json:"privacy_class,omitempty"`
	Metrics           Metrics    `json:"metrics,omitempty"`
	Assertion         *Assertion `json:"assertion,omitempty"`
	ExpectedCostMinor int64      `json:"expected_cost_minor,omitempty"`
}

type Decision struct {
	Allowed bool   `json:"allowed"`
	Code    string `json:"code"`
	Grant   string `json:"grant,omitempty"`
}

type Policy interface {
	Authorize(context.Context, string, Request) (Decision, error)
	Evaluate(context.Context, string, Request, string) (Decision, error)
}

type ArtifactEvidence struct {
	SHA256     string `json:"sha256"`
	ByteCount  int64  `json:"byte_count"`
	Provenance string `json:"provenance"`
	Verified   bool   `json:"verified"`
}

type Runner interface {
	Execute(context.Context, string) (ArtifactEvidence, error)
}

type Goal struct {
	ID               string            `json:"id"`
	Action           string            `json:"action"`
	PrivacyClass     string            `json:"privacy_class"`
	Status           string            `json:"status"`
	Metrics          Metrics           `json:"metrics"`
	Score            *float64          `json:"score"`
	PolicyCode       string            `json:"policy_code"`
	ProvenanceStatus string            `json:"provenance_status"`
	Assertion        *Assertion        `json:"assertion,omitempty"`
	Evidence         *ArtifactEvidence `json:"evidence,omitempty"`
	CreatedAtUTC     string            `json:"created_at_utc"`
	UpdatedAtUTC     string            `json:"updated_at_utc"`
}

type TickResult struct {
	Outcome string `json:"outcome"`
	GoalID  string `json:"goal_id,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

// IntimacyState is privacy-safe control metadata; it never contains intimate payloads.
type IntimacyState struct {
	Stopped          bool   `json:"stopped"`
	PrivacyMode      string `json:"privacy_mode"`
	BackendAvailable bool   `json:"backend_available"`
	UpdatedAtUTC     string `json:"updated_at_utc"`
}
