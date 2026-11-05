package core

import (
	"net/http"
	"strings"

	"lunaos.local/lunaos/services/intimacy"
)

// serveSubjects is the API behind the subject-operated consent screen. Reading
// and withdrawing consent work from any owner channel; issuing a spoken
// challenge and recording a grant are accepted only from the owner's Unix
// socket, where the trusted local consent screen runs. Any withdrawal also
// closes LUNA's eyes, so a capture lease cannot outlive the consent it relied on.
func (h *HTTPHandler) serveSubjects(w http.ResponseWriter, r *http.Request, trustedUI bool) {
	if h.vault == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"code": "INTIMACY_VAULT_UNAVAILABLE"})
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	ctx := r.Context()
	switch {
	case len(parts) == 2 && r.Method == http.MethodPost:
		var body struct {
			Kind        string `json:"kind"`
			AdultStatus string `json:"adult_status"`
		}
		if err := decodeBody(w, r, &body); err != nil {
			writeEngineError(w, err)
			return
		}
		if body.AdultStatus == "attested_adult" && !trustedUI {
			writeEngineError(w, ErrTrustedUIRequired)
			return
		}
		subject, err := h.vault.CreateSubject(ctx, body.Kind, body.AdultStatus)
		if err != nil {
			writeIntimacyError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, subject)
	case len(parts) == 4 && parts[3] == "challenge" && r.Method == http.MethodPost:
		if !trustedUI {
			writeEngineError(w, ErrTrustedUIRequired)
			return
		}
		phrase, err := h.vault.IssueVoiceChallenge(ctx, parts[2])
		if err != nil {
			writeIntimacyError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"phrase": phrase})
	case len(parts) == 4 && parts[3] == "consents" && r.Method == http.MethodPost:
		if !trustedUI {
			writeEngineError(w, ErrTrustedUIRequired)
			return
		}
		var request intimacy.SubjectConsentRequest
		if err := decodeBody(w, r, &request); err != nil {
			writeEngineError(w, err)
			return
		}
		grants, err := h.vault.GrantSubjectConsent(ctx, parts[2], request)
		if err != nil {
			writeIntimacyError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"grants": grants})
	case len(parts) == 4 && parts[3] == "scopes" && r.Method == http.MethodGet:
		grants, err := h.vault.ActiveScopes(ctx, parts[2])
		if err != nil {
			writeIntimacyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"grants": grants})
	case len(parts) == 4 && parts[3] == "revoke" && r.Method == http.MethodPost:
		var body struct {
			Scopes []intimacy.SubjectScope `json:"scopes"`
		}
		if err := decodeBody(w, r, &body); err != nil {
			writeEngineError(w, err)
			return
		}
		if _, err := h.engine.EnterBlind(ctx, "consent_revoked"); err != nil {
			writeEngineError(w, err)
			return
		}
		receipt, err := h.vault.RevokeSubjectScopes(ctx, parts[2], body.Scopes)
		if err != nil {
			writeIntimacyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, receipt)
	case len(parts) == 3 && r.Method == http.MethodDelete:
		if _, err := h.engine.EnterBlind(ctx, "consent_revoked"); err != nil {
			writeEngineError(w, err)
			return
		}
		receipt, err := h.vault.DeleteSubject(ctx, parts[2])
		if err != nil {
			writeIntimacyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, receipt)
	default:
		writeJSON(w, http.StatusNotFound, map[string]string{"code": "NOT_FOUND"})
	}
}
