package core

import (
	"net/http"
	"strings"

	"lunaos.local/lunaos/services/intimacy"
)

// serveLibrary exposes the private media library. Anything that returns file
// names or tags answers only in a private context with intimacy not stopped;
// otherwise it reports BLOCKED_PRIVACY and reveals nothing.
func (h *HTTPHandler) serveLibrary(w http.ResponseWriter, r *http.Request) {
	if h.vault == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"code": "INTIMACY_VAULT_UNAVAILABLE"})
		return
	}
	ctx := r.Context()
	allowed, err := h.engine.ExplicitContentAllowed(ctx)
	if err != nil {
		writeEngineError(w, err)
		return
	}
	if !allowed {
		writeJSON(w, http.StatusLocked, map[string]string{"code": "BLOCKED_PRIVACY"})
		return
	}
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/v1/intimacy/library/"), "/"), "/")
	switch {
	case len(parts) == 1 && parts[0] == "sources" && r.Method == http.MethodGet:
		sources, err := h.vault.ListLibrarySources(ctx, r.URL.Query().Get("consent_id"))
		if err != nil {
			writeIntimacyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"sources": sources})
	case len(parts) == 1 && parts[0] == "sources" && r.Method == http.MethodPost:
		var body struct {
			ConsentID string `json:"consent_id"`
			Path      string `json:"path"`
			Label     string `json:"label"`
		}
		if err := decodeBody(w, r, &body); err != nil {
			writeEngineError(w, err)
			return
		}
		source, err := h.vault.AddLibrarySource(ctx, body.ConsentID, body.Path, body.Label)
		if err != nil {
			writeIntimacyError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, source)
	case len(parts) == 3 && parts[0] == "sources" && parts[2] == "scan" && r.Method == http.MethodPost:
		var body struct {
			ConsentID string `json:"consent_id"`
		}
		if err := decodeBody(w, r, &body); err != nil {
			writeEngineError(w, err)
			return
		}
		report, err := h.vault.ScanLibrary(ctx, body.ConsentID, parts[1])
		if err != nil {
			writeIntimacyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, report)
	case len(parts) == 1 && parts[0] == "search" && r.Method == http.MethodPost:
		var body struct {
			ConsentID string `json:"consent_id"`
			intimacy.LibraryQuery
		}
		if err := decodeBody(w, r, &body); err != nil {
			writeEngineError(w, err)
			return
		}
		items, err := h.vault.SearchLibrary(ctx, body.ConsentID, body.LibraryQuery)
		if err != nil {
			writeIntimacyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	case len(parts) == 3 && parts[0] == "items" && parts[2] == "tags" && r.Method == http.MethodPut:
		var body struct {
			ConsentID string   `json:"consent_id"`
			Tags      []string `json:"tags"`
		}
		if err := decodeBody(w, r, &body); err != nil {
			writeEngineError(w, err)
			return
		}
		if err := h.vault.TagLibraryItem(ctx, body.ConsentID, parts[1], body.Tags); err != nil {
			writeIntimacyError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	case len(parts) == 2 && parts[0] == "items" && r.Method == http.MethodDelete:
		if err := h.vault.RemoveLibraryItem(ctx, r.URL.Query().Get("consent_id"), parts[1]); err != nil {
			writeIntimacyError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		writeJSON(w, http.StatusNotFound, map[string]string{"code": "NOT_FOUND"})
	}
}
