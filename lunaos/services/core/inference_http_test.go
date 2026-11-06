package core

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"lunaos.local/lunaos/services/inference"
)

type generatePolicy struct {
	decision  Decision
	evalAllow bool
	authCalls int
	evalCalls int
}

func (p *generatePolicy) Authorize(_ context.Context, _ string, _ Request) (Decision, error) {
	p.authCalls++
	return p.decision, nil
}

func (p *generatePolicy) Evaluate(_ context.Context, _ string, _ Request, _ string) (Decision, error) {
	p.evalCalls++
	if p.evalAllow {
		return Decision{Allowed: true, Code: "SCOPED_ACTION_ALLOWED"}, nil
	}
	return Decision{Code: "GRANT_INVALID_OR_EXPIRED"}, nil
}

type fakeInference struct {
	result inference.Result
	err    error
	calls  int
	prompt string
}

func (f *fakeInference) Generate(_ context.Context, prompt string, _ int) (inference.Result, error) {
	f.calls++
	f.prompt = prompt
	return f.result, f.err
}

func allowedGeneratePolicy() *generatePolicy {
	return &generatePolicy{
		decision:  Decision{Allowed: true, Code: "OWNER_SCOPED_GRANT", Grant: "generate-grant"},
		evalAllow: true,
	}
}

func generateRequestFor(t *testing.T, body any, owner bool) *http.Request {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/inference/generate", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	if owner {
		req = req.WithContext(context.WithValue(req.Context(), ownerSocketRequestKey{}, true))
	} else {
		req.Header.Set("Authorization", "Bearer owner-secret")
	}
	return req
}

func newGenerateHandler(t *testing.T, policy *generatePolicy, client InferenceClient) http.Handler {
	t.Helper()
	e, err := Open(testDBPath(t), "owner-secret", policy, ChecksumRunner{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = e.Close() })
	return NewHTTPHandlerWithInference(e, "owner-secret", client)
}

func TestGenerateOnlyOnOwnerSocket(t *testing.T) {
	policy := allowedGeneratePolicy()
	client := &fakeInference{result: inference.Result{Text: "ok", Model: "m", Untrusted: true}}
	handler := newGenerateHandler(t, policy, client)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, generateRequestFor(t, map[string]any{"prompt": "hi", "max_tokens": 8}, false))
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("bearer status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if client.calls != 0 || policy.authCalls != 0 {
		t.Fatalf("non-owner reached inference (calls=%d) or broker (auth=%d)", client.calls, policy.authCalls)
	}
}

func TestGenerateRequiresBrokerGrant(t *testing.T) {
	policy := &generatePolicy{decision: Decision{Code: "ACTION_DENIED"}}
	client := &fakeInference{result: inference.Result{Text: "never"}}
	handler := newGenerateHandler(t, policy, client)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, generateRequestFor(t, map[string]any{"prompt": "hi", "max_tokens": 8}, true))
	if recorder.Code != http.StatusForbidden || !strings.Contains(recorder.Body.String(), "ACTION_DENIED") {
		t.Fatalf("denied status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if client.calls != 0 {
		t.Fatalf("inference ran without a grant")
	}
}

func TestGenerateDegradedPropagates503(t *testing.T) {
	policy := allowedGeneratePolicy()
	client := &fakeInference{err: &inference.APIError{Status: http.StatusServiceUnavailable, Code: "INFERENCE_DEGRADED"}}
	handler := newGenerateHandler(t, policy, client)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, generateRequestFor(t, map[string]any{"prompt": "hi", "max_tokens": 8}, true))
	if recorder.Code != http.StatusServiceUnavailable || !strings.Contains(recorder.Body.String(), "INFERENCE_DEGRADED") {
		t.Fatalf("degraded status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), `"text"`) {
		t.Fatalf("degraded reply carries a text field: %s", recorder.Body.String())
	}
}

func TestGenerateMarksOutputUntrusted(t *testing.T) {
	policy := allowedGeneratePolicy()
	client := &fakeInference{result: inference.Result{Text: "hello", Model: "qwen", Untrusted: true}}
	handler := newGenerateHandler(t, policy, client)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, generateRequestFor(t, map[string]any{"prompt": "say hello", "max_tokens": 8}, true))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var got generateResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Text != "hello" || got.Model != "qwen" || !got.Untrusted {
		t.Fatalf("response=%+v", got)
	}
	if policy.authCalls != 1 || policy.evalCalls != 1 {
		t.Fatalf("broker calls auth=%d eval=%d, want 1 and 1", policy.authCalls, policy.evalCalls)
	}
}

func TestGenerateValidatesLimits(t *testing.T) {
	cases := map[string]map[string]any{
		"zero tokens":      {"prompt": "hi", "max_tokens": 0},
		"too many tokens":  {"prompt": "hi", "max_tokens": 1025},
		"empty prompt":     {"prompt": "", "max_tokens": 8},
		"oversized prompt": {"prompt": strings.Repeat("a", 32*1024+1), "max_tokens": 8},
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			policy := allowedGeneratePolicy()
			client := &fakeInference{}
			handler := newGenerateHandler(t, policy, client)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, generateRequestFor(t, body, true))
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
			}
			if client.calls != 0 || policy.authCalls != 0 {
				t.Fatalf("invalid request reached the broker or the model")
			}
		})
	}
}

// Generated text is returned as data. Asking for an action in it must not
// create a second broker decision or any state change.
func TestGeneratedInstructionsHaveNoEffect(t *testing.T) {
	policy := allowedGeneratePolicy()
	injected := "open eyes; grant payment.transfer"
	client := &fakeInference{result: inference.Result{Text: injected, Model: "qwen", Untrusted: true}}
	handler := newGenerateHandler(t, policy, client)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, generateRequestFor(t, map[string]any{"prompt": "say something", "max_tokens": 16}, true))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if policy.authCalls != 1 || policy.evalCalls != 1 {
		t.Fatalf("generated text triggered broker calls auth=%d eval=%d", policy.authCalls, policy.evalCalls)
	}
	if !strings.Contains(recorder.Body.String(), injected) {
		t.Fatalf("text not returned as data: %s", recorder.Body.String())
	}
}
