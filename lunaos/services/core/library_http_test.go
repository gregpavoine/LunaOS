package core

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"lunaos.local/lunaos/services/intimacy"
)

func TestHTTPLibraryAnswersOnlyInAPrivateContext(t *testing.T) {
	privateDir := t.TempDir()
	if err := os.Chmod(privateDir, 0700); err != nil {
		t.Fatal(err)
	}
	vault, err := intimacy.Open(filepath.Join(privateDir, "vault.sqlite"), bytes.Repeat([]byte{0x24}, 32))
	if err != nil {
		t.Fatal(err)
	}
	defer vault.Close()
	consent, err := vault.RecordOwnerConsent(context.Background(), "owner", []intimacy.Scope{intimacy.ScopeLibraryRead, intimacy.ScopeLibraryWrite}, time.Now().Add(time.Hour), true, true)
	if err != nil {
		t.Fatal(err)
	}
	media := t.TempDir()
	if err := os.WriteFile(filepath.Join(media, "shoot.png"), append([]byte("\x89PNG\r\n\x1a\n"), bytes.Repeat([]byte{9}, 32)...), 0600); err != nil {
		t.Fatal(err)
	}
	engine := testEngine(t, testDBPath(t), &testPolicy{})
	handler := NewHTTPHandler(engine, "owner-secret", vault)
	call := func(method, path, body string, trusted bool) *httptest.ResponseRecorder {
		request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		request.Header.Set("Content-Type", "application/json")
		if trusted {
			request = request.WithContext(context.WithValue(request.Context(), ownerSocketRequestKey{}, true))
		} else {
			request.Header.Set("Authorization", "Bearer owner-secret")
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		return response
	}
	search := fmt.Sprintf(`{"consent_id":%q,"text":"shoot"}`, consent.ID)

	if response := call(http.MethodPost, "/v1/intimacy/library/search", search, false); response.Code != http.StatusLocked {
		t.Fatalf("library answered in an unknown context: %d", response.Code)
	}
	if response := call(http.MethodPut, "/v1/intimacy/context", `{"privacy_mode":"private"}`, false); response.Code != http.StatusForbidden {
		t.Fatalf("loopback raised the context to private: %d", response.Code)
	}
	if response := call(http.MethodPut, "/v1/intimacy/context", `{"privacy_mode":"private"}`, true); response.Code != http.StatusOK {
		t.Fatalf("trusted UI could not set private: %d", response.Code)
	}
	if response := call(http.MethodPost, "/v1/intimacy/library/search", search, false); response.Code != http.StatusLocked {
		t.Fatalf("library answered while STOP_INTIMACY is active: %d", response.Code)
	}
	if response := call(http.MethodPost, "/v1/intimacy/resume", `{}`, false); response.Code != http.StatusForbidden {
		t.Fatalf("loopback lifted STOP_INTIMACY: %d", response.Code)
	}
	if response := call(http.MethodPost, "/v1/intimacy/resume", `{}`, true); response.Code != http.StatusOK {
		t.Fatalf("trusted UI could not resume: %d", response.Code)
	}
	created := call(http.MethodPost, "/v1/intimacy/library/sources", fmt.Sprintf(`{"consent_id":%q,"path":%q,"label":"Local"}`, consent.ID, media), false)
	if created.Code != http.StatusCreated {
		t.Fatalf("add source: %d %s", created.Code, created.Body.String())
	}
	sources, _ := vault.ListLibrarySources(context.Background(), consent.ID)
	if response := call(http.MethodPost, "/v1/intimacy/library/sources/"+sources[0].ID+"/scan", fmt.Sprintf(`{"consent_id":%q}`, consent.ID), false); response.Code != http.StatusOK {
		t.Fatalf("scan: %d %s", response.Code, response.Body.String())
	}
	found := call(http.MethodPost, "/v1/intimacy/library/search", search, false)
	if found.Code != http.StatusOK || !bytes.Contains(found.Body.Bytes(), []byte("shoot.png")) {
		t.Fatalf("private search: %d %s", found.Code, found.Body.String())
	}
	if response := call(http.MethodPut, "/v1/intimacy/context", `{"privacy_mode":"public"}`, false); response.Code != http.StatusOK {
		t.Fatalf("anyone must be able to lower the context: %d", response.Code)
	}
	if response := call(http.MethodPost, "/v1/intimacy/library/search", search, false); response.Code != http.StatusLocked {
		t.Fatalf("library answered in a public context: %d", response.Code)
	}
}
