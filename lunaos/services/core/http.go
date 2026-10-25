package core

import (
	"context"
	"crypto/hmac"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"lunaos.local/bootstrap/services/intimacy"
)

const maxRequestBytes = 8192

type ownerSocketRequestKey struct{}

// OwnerSocketConnContext marks requests only when their accepted connection is a
// Unix socket. Mount it on the private, owner-group-restricted listener only.
func OwnerSocketConnContext(ctx context.Context, conn net.Conn) context.Context {
	if _, ok := conn.(*net.UnixConn); !ok {
		return ctx
	}
	return context.WithValue(ctx, ownerSocketRequestKey{}, true)
}

type HTTPHandler struct {
	engine *Engine
	token  string
	vault  *intimacy.Vault
}

func NewHTTPHandler(engine *Engine, ownerToken string, vaults ...*intimacy.Vault) http.Handler {
	handler := &HTTPHandler{engine: engine, token: ownerToken}
	if len(vaults) > 0 {
		handler.vault = vaults[0]
	}
	return handler
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if r.Header.Get("Origin") != "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"code": "BROWSER_ORIGIN_NOT_SUPPORTED"})
		return
	}
	authorization := r.Header.Get("Authorization")
	trustedUnixOwner, _ := r.Context().Value(ownerSocketRequestKey{}).(bool)
	if !trustedUnixOwner && (!strings.HasPrefix(authorization, "Bearer ") || !hmac.Equal([]byte(strings.TrimPrefix(authorization, "Bearer ")), []byte(h.token))) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"code": ErrUnauthorized.Error()})
		return
	}
	ownerPrincipal := strings.TrimPrefix(authorization, "Bearer ")
	if trustedUnixOwner {
		ownerPrincipal = h.token
	}
	if r.Method == http.MethodGet && r.URL.Path == "/v1/intimacy/relationship/state" {
		state, err := h.engine.IntimacyState(r.Context())
		if err != nil {
			writeEngineError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, state)
		return
	}
	if r.Method == http.MethodPost && r.URL.Path == "/v1/intimacy/stop" {
		var empty map[string]any
		if err := decodeBody(w, r, &empty); err != nil {
			writeEngineError(w, err)
			return
		}
		if len(empty) != 0 {
			writeEngineError(w, fmt.Errorf("%w: EMPTY_OBJECT_REQUIRED", ErrInvalid))
			return
		}
		if err := h.engine.StopIntimacy(r.Context()); err != nil {
			writeEngineError(w, err)
			return
		}
		state, err := h.engine.IntimacyState(r.Context())
		if err != nil {
			writeEngineError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, state)
		return
	}
	if r.Method == http.MethodPost && r.URL.Path == "/v1/intimacy/conversation" {
		state, err := h.engine.IntimacyState(r.Context())
		if err != nil {
			writeEngineError(w, err)
			return
		}
		if state.Stopped || state.PrivacyMode != "private" {
			writeJSON(w, http.StatusLocked, map[string]string{"code": "BLOCKED_PRIVACY"})
			return
		}
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"code": "LOCAL_MODEL_UNAVAILABLE"})
		return
	}
	if strings.HasPrefix(r.URL.Path, "/v1/intimacy/") && h.vault == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"code": "INTIMACY_VAULT_UNAVAILABLE"})
		return
	}
	if r.Method == http.MethodPost && r.URL.Path == "/v1/intimacy/consents" {
		var body struct {
			Scopes        []intimacy.Scope `json:"scopes"`
			ExpiresAt     string           `json:"expires_at"`
			ConfirmOwner  bool             `json:"confirm_owner"`
			AdultAttested bool             `json:"adult_attested"`
		}
		if err := decodeBody(w, r, &body); err != nil {
			writeEngineError(w, err)
			return
		}
		expires, err := time.Parse(time.RFC3339, body.ExpiresAt)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"code": "INVALID_CONSENT"})
			return
		}
		consent, err := h.vault.RecordOwnerConsent(r.Context(), "owner", body.Scopes, expires, body.ConfirmOwner, body.AdultAttested)
		if err != nil {
			writeIntimacyError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, consent)
		return
	}
	if r.Method == http.MethodPost && r.URL.Path == "/v1/intimacy/memories" {
		var body struct {
			ConsentID string          `json:"consent_id"`
			Origin    intimacy.Origin `json:"origin_type"`
			Content   string          `json:"content"`
		}
		if err := decodeBody(w, r, &body); err != nil {
			writeEngineError(w, err)
			return
		}
		memory, err := h.vault.StoreMemory(r.Context(), body.ConsentID, "owner", body.Origin, body.Content)
		if err != nil {
			writeIntimacyError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"memory_id": memory.ID, "subject_id": memory.SubjectID, "origin_type": memory.Origin, "created_at": memory.CreatedAt})
		return
	}
	if r.Method == http.MethodPost && r.URL.Path == "/v1/intimacy/preferences" {
		var body struct {
			ConsentID string  `json:"consent_id"`
			Dimension string  `json:"dimension"`
			Value     string  `json:"value"`
			Strength  float64 `json:"strength"`
			Timescale string  `json:"timescale"`
			Confirm   bool    `json:"confirm_preference"`
		}
		if err := decodeBody(w, r, &body); err != nil {
			writeEngineError(w, err)
			return
		}
		preference, err := h.vault.AddPreference(r.Context(), body.ConsentID, intimacy.PreferenceInput{Dimension: body.Dimension, Value: body.Value, Strength: body.Strength, Timescale: body.Timescale, Confirmed: body.Confirm})
		if err != nil {
			writeIntimacyError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, safePreferenceMetadata(preference))
		return
	}
	if r.Method == http.MethodGet && r.URL.Path == "/v1/intimacy/preferences" {
		preferences, err := h.vault.ListPreferences(r.Context(), r.URL.Query().Get("consent_id"))
		if err != nil {
			writeIntimacyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"preferences": preferences})
		return
	}
	if r.URL.Path == "/v1/intimacy/discovery/policy" && r.Method == http.MethodPut {
		var body struct {
			ConsentID string                      `json:"consent_id"`
			Policy    intimacy.DiscoveryAllowlist `json:"policy"`
			Confirm   bool                        `json:"confirm_owner"`
		}
		if err := decodeBody(w, r, &body); err != nil {
			writeEngineError(w, err)
			return
		}
		if err := h.vault.SetDiscoveryPolicy(r.Context(), body.ConsentID, body.Policy, body.Confirm); err != nil {
			writeIntimacyError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.URL.Path == "/v1/intimacy/discovery/policy" && r.Method == http.MethodGet {
		policy, err := h.vault.ReadDiscoveryPolicy(r.Context(), r.URL.Query().Get("consent_id"))
		if err != nil {
			writeIntimacyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, policy)
		return
	}
	if r.Method == http.MethodPost && r.URL.Path == "/v1/intimacy/discovery/evaluate" {
		var body struct {
			ConsentID string                      `json:"consent_id"`
			Candidate intimacy.DiscoveryCandidate `json:"candidate"`
		}
		if err := decodeBody(w, r, &body); err != nil {
			writeEngineError(w, err)
			return
		}
		decision, err := h.vault.EvaluateDiscovery(r.Context(), body.ConsentID, body.Candidate)
		if err != nil {
			writeIntimacyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, decision)
		return
	}
	if r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/v1/intimacy/consents/") && strings.HasSuffix(r.URL.Path, "/revoke") {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) != 5 || parts[0] != "v1" || parts[1] != "intimacy" || parts[2] != "consents" || parts[4] != "revoke" {
			writeJSON(w, http.StatusNotFound, map[string]string{"code": "NOT_FOUND"})
			return
		}
		var empty map[string]any
		if err := decodeBody(w, r, &empty); err != nil || len(empty) != 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"code": "INVALID_REQUEST"})
			return
		}
		if err := h.vault.RevokeConsent(r.Context(), parts[3], "owner"); err != nil {
			writeIntimacyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"code": "CONSENT_REVOKED"})
		return
	}
	if strings.HasPrefix(r.URL.Path, "/v1/intimacy/memories/") {
		id := strings.TrimPrefix(r.URL.Path, "/v1/intimacy/memories/")
		if id == "" || strings.Contains(id, "/") {
			writeJSON(w, http.StatusNotFound, map[string]string{"code": "NOT_FOUND"})
			return
		}
		if r.Method == http.MethodGet {
			memory, err := h.vault.ReadMemory(r.Context(), id, r.URL.Query().Get("consent_id"), "owner")
			if err != nil {
				writeIntimacyError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, memory)
			return
		}
		if r.Method == http.MethodDelete {
			if err := h.vault.DeleteMemory(r.Context(), id, "owner"); err != nil {
				writeIntimacyError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	if strings.HasPrefix(r.URL.Path, "/v1/intimacy/preferences/") {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/intimacy/preferences/"), "/")
		if len(parts) == 2 && parts[0] != "" && parts[1] == "correction" && r.Method == http.MethodPost {
			var body struct {
				ConsentID string  `json:"consent_id"`
				Value     string  `json:"value"`
				Strength  float64 `json:"strength"`
				Disagreed bool    `json:"disagreed_with_previous"`
				Confirm   bool    `json:"confirm_correction"`
			}
			if err := decodeBody(w, r, &body); err != nil {
				writeEngineError(w, err)
				return
			}
			preference, err := h.vault.CorrectPreference(r.Context(), parts[0], body.ConsentID, body.Value, body.Strength, body.Disagreed, body.Confirm)
			if err != nil {
				writeIntimacyError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, safePreferenceMetadata(preference))
			return
		}
		if len(parts) == 1 && parts[0] != "" && r.Method == http.MethodDelete {
			if err := h.vault.DeletePreference(r.Context(), parts[0], "owner"); err != nil {
				writeIntimacyError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	if r.Method == http.MethodGet && r.URL.Path == "/v1/goals" {
		goals, err := h.engine.List(r.Context())
		if err != nil {
			writeEngineError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"goals": goals})
		return
	}
	if r.Method == http.MethodPost && r.URL.Path == "/v1/goals" {
		var body Request
		if err := decodeBody(w, r, &body); err != nil {
			writeEngineError(w, err)
			return
		}
		goal, err := h.engine.Create(r.Context(), body, ownerPrincipal, r.Header.Get("Idempotency-Key"))
		if err != nil {
			writeEngineError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, goal)
		return
	}
	if r.Method == http.MethodPost && r.URL.Path == "/v1/scheduler/tick" {
		var empty map[string]any
		if err := decodeBody(w, r, &empty); err != nil {
			writeEngineError(w, err)
			return
		}
		if len(empty) != 0 {
			writeEngineError(w, fmt.Errorf("%w: EMPTY_OBJECT_REQUIRED", ErrInvalid))
			return
		}
		result, err := h.engine.Tick(r.Context())
		if err != nil {
			writeEngineError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
		return
	}
	if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v1/goals/") {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/goals/"), "/")
		if len(parts) == 1 && parts[0] != "" {
			goal, err := h.engine.Get(r.Context(), parts[0])
			if err != nil {
				writeEngineError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, goal)
			return
		}
		if len(parts) == 2 && parts[0] != "" && parts[1] == "evidence" {
			goal, err := h.engine.Get(r.Context(), parts[0])
			if err != nil {
				writeEngineError(w, err)
				return
			}
			if goal.Evidence == nil {
				writeJSON(w, http.StatusNotFound, map[string]string{"code": ErrNotFound.Error()})
				return
			}
			writeJSON(w, http.StatusOK, goal.Evidence)
			return
		}
	}
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"code": "METHOD_NOT_ALLOWED"})
		return
	}
	writeJSON(w, http.StatusNotFound, map[string]string{"code": "NOT_FOUND"})
}

func writeIntimacyError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, intimacy.ErrConsentDenied):
		writeJSON(w, http.StatusForbidden, map[string]string{"code": "CONSENT_DENIED"})
	case errors.Is(err, intimacy.ErrInvalidConsent), errors.Is(err, intimacy.ErrInvalidOrigin):
		writeJSON(w, http.StatusBadRequest, map[string]string{"code": err.Error()})
	case errors.Is(err, intimacy.ErrInvalidPreference):
		writeJSON(w, http.StatusBadRequest, map[string]string{"code": "INVALID_PREFERENCE"})
	case errors.Is(err, intimacy.ErrInvalidDiscoveryAllowlist), errors.Is(err, intimacy.ErrDiscoveryPolicyUnconfirmed):
		writeJSON(w, http.StatusBadRequest, map[string]string{"code": err.Error()})
	case errors.Is(err, intimacy.ErrMemoryNotFound), errors.Is(err, intimacy.ErrPreferenceNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"code": err.Error()})
	default:
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"code": "INTIMACY_VAULT_UNAVAILABLE"})
	}
}

func safePreferenceMetadata(preference intimacy.Preference) map[string]any {
	return map[string]any{
		"preference_id": preference.ID, "dimension": preference.Dimension,
		"source_type": preference.SourceType, "confidence": preference.Confidence,
		"strength": preference.Strength, "timescale": preference.Timescale,
		"first_seen_at": preference.FirstSeenAt, "last_seen_at": preference.LastSeenAt,
		"evidence_count":          preference.EvidenceCount,
		"negative_feedback_count": preference.NegativeFeedbackCount,
		"consent_scope":           preference.ConsentScope,
	}
}

func decodeBody(w http.ResponseWriter, r *http.Request, target any) error {
	if r.Header.Get("Content-Type") != "application/json" || len(r.TransferEncoding) != 0 {
		return fmt.Errorf("%w: JSON_CONTENT_TYPE_REQUIRED", ErrInvalid)
	}
	if r.ContentLength <= 0 || r.ContentLength > maxRequestBytes {
		return fmt.Errorf("%w: BODY_SIZE_INVALID", ErrInvalid)
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("%w: INVALID_JSON", ErrInvalid)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: TRAILING_JSON", ErrInvalid)
	}
	return nil
}

func writeEngineError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrUnauthorized):
		writeJSON(w, http.StatusUnauthorized, map[string]string{"code": ErrUnauthorized.Error()})
	case errors.Is(err, ErrInvalid), errors.Is(err, ErrIdempotencyConflict):
		writeJSON(w, http.StatusBadRequest, map[string]string{"code": err.Error()})
	case errors.Is(err, ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"code": ErrNotFound.Error()})
	default:
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"code": "LOCAL_SERVICE_UNAVAILABLE"})
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
