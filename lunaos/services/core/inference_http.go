package core

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"lunaos.local/lunaos/services/inference"
)

// Text generation is the only route that carries a prompt, so it gets its own
// limit: the prompt cap plus room for the JSON envelope.
const (
	maxGenerateBodyBytes  = 40 * 1024
	maxGeneratePromptSize = 32 * 1024
	maxGenerateTokens     = 1024
	generateTimeout       = 90 * time.Second
	generateAction        = "inference.generate"
)

// InferenceClient is the part of the local inference service the Core uses.
// The Core never holds a model or a token: it only sends a prompt over the
// inference socket.
type InferenceClient interface {
	Generate(ctx context.Context, prompt string, maxTokens int) (inference.Result, error)
}

type generateRequest struct {
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens"`
}

type generateResponse struct {
	Text      string `json:"text"`
	Model     string `json:"model"`
	Untrusted bool   `json:"untrusted"`
}

// handleGenerate runs a text generation only for the owner socket and only
// when the Policy Broker grants the registered inference.generate action. The
// generated text is returned to the caller as untrusted data: it is never
// stored, logged, or interpreted as an action here.
func (h *HTTPHandler) handleGenerate(w http.ResponseWriter, r *http.Request, trustedUnixOwner bool) {
	if !trustedUnixOwner {
		writeJSON(w, http.StatusForbidden, map[string]string{"code": "OWNER_SOCKET_REQUIRED"})
		return
	}
	if h.inference == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"code": "INFERENCE_UNAVAILABLE"})
		return
	}
	var body generateRequest
	decoder := json.NewDecoder(io.LimitReader(r.Body, maxGenerateBodyBytes))
	if err := decoder.Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"code": "INVALID_REQUEST"})
		return
	}
	if body.Prompt == "" || len(body.Prompt) > maxGeneratePromptSize ||
		body.MaxTokens < 1 || body.MaxTokens > maxGenerateTokens {
		writeJSON(w, http.StatusBadRequest, map[string]string{"code": "INVALID_REQUEST"})
		return
	}

	goalID, err := newID()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"code": "INTERNAL"})
		return
	}
	req := Request{Action: generateAction, PrivacyClass: "general", Title: "Generate text"}
	decision, err := h.engine.policy.Authorize(r.Context(), goalID, req)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"code": "BROKER_UNAVAILABLE"})
		return
	}
	if !decision.Allowed || decision.Grant == "" {
		code := decision.Code
		if code == "" {
			code = "ACTION_DENIED"
		}
		writeJSON(w, http.StatusForbidden, map[string]string{"code": code})
		return
	}
	checked, err := h.engine.policy.Evaluate(r.Context(), goalID, req, decision.Grant)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"code": "BROKER_UNAVAILABLE"})
		return
	}
	if !checked.Allowed {
		code := checked.Code
		if code == "" {
			code = "GRANT_INVALID_OR_EXPIRED"
		}
		writeJSON(w, http.StatusForbidden, map[string]string{"code": code})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), generateTimeout)
	defer cancel()
	result, err := h.inference.Generate(ctx, body.Prompt, body.MaxTokens)
	if err != nil {
		writeInferenceError(w, err)
		return
	}
	// Explicit fields only: nothing from the inference reply is forwarded
	// beyond the text, the model name and the untrusted marker.
	writeJSON(w, http.StatusOK, generateResponse{Text: result.Text, Model: result.Model, Untrusted: true})
}

// writeInferenceError keeps a degraded or timed-out model visible to the owner
// as a 503 with its code, never as an empty text presented as a success.
func writeInferenceError(w http.ResponseWriter, err error) {
	var apiErr *inference.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.Code {
		case "INFERENCE_DEGRADED", "INFERENCE_TIMEOUT", "INFERENCE_NO_MODEL":
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"code": apiErr.Code})
			return
		}
		writeJSON(w, http.StatusBadGateway, map[string]string{"code": "INFERENCE_ERROR"})
		return
	}
	writeJSON(w, http.StatusServiceUnavailable, map[string]string{"code": "INFERENCE_UNAVAILABLE"})
}
