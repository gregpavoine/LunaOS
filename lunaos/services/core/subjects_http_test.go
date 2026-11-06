package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"lunaos.local/lunaos/services/intimacy"
)

func TestHTTPConsentScreenIsTrustedUIOnlyAndRevocationClosesEyes(t *testing.T) {
	privateDir := t.TempDir()
	if err := os.Chmod(privateDir, 0700); err != nil {
		t.Fatal(err)
	}
	vault, err := intimacy.Open(filepath.Join(privateDir, "vault.sqlite"), bytes.Repeat([]byte{0x42}, 32))
	if err != nil {
		t.Fatal(err)
	}
	defer vault.Close()
	engine := testEngine(t, testDBPath(t), &testPolicy{})
	handler := NewHTTPHandler(engine, "owner-secret", vault)
	call := func(method, path, body string, trusted bool) *httptest.ResponseRecorder {
		request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		if body != "" {
			request.Header.Set("Content-Type", "application/json")
		}
		if trusted {
			request = request.WithContext(context.WithValue(request.Context(), ownerSocketRequestKey{}, true))
		} else {
			request.Header.Set("Authorization", "Bearer owner-secret")
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		return response
	}

	if response := call(http.MethodPost, "/v1/subjects", `{"kind":"OTHER","adult_status":"attested_adult"}`, false); response.Code != http.StatusForbidden {
		t.Fatalf("adult attestation from the loopback API: %d", response.Code)
	}
	created := call(http.MethodPost, "/v1/subjects", `{"kind":"OTHER","adult_status":"attested_adult"}`, true)
	var subject intimacy.Subject
	if created.Code != http.StatusCreated || json.Unmarshal(created.Body.Bytes(), &subject) != nil {
		t.Fatalf("trusted subject creation: %d %s", created.Code, created.Body.String())
	}
	grant := fmt.Sprintf(`{"scopes":["FACE_REFERENCE","INTIMATE_VIEW"],"disclosure_revision":%q,"intimate_disclosure_acknowledged":true,"proof":"subject_click","expires_at":%q}`,
		intimacy.CurrentDisclosureRevision, time.Now().Add(time.Hour).UTC().Format(time.RFC3339))
	path := "/v1/subjects/" + subject.ID
	if response := call(http.MethodPost, path+"/consents", grant, false); response.Code != http.StatusForbidden {
		t.Fatalf("consent recorded from outside the consent screen: %d", response.Code)
	}
	if response := call(http.MethodPost, path+"/consents", grant, true); response.Code != http.StatusCreated {
		t.Fatalf("consent from the trusted screen: %d %s", response.Code, response.Body.String())
	}
	if response := call(http.MethodGet, path+"/scopes", "", false); response.Code != http.StatusOK || !bytes.Contains(response.Body.Bytes(), []byte("INTIMATE_VIEW")) {
		t.Fatalf("scope listing: %d %s", response.Code, response.Body.String())
	}
	if _, err := engine.OpenEyes(context.Background(), "owner-secret", true); err != nil {
		t.Fatal(err)
	}
	if response := call(http.MethodPost, path+"/revoke", `{"scopes":["INTIMATE_VIEW"]}`, false); response.Code != http.StatusOK {
		t.Fatalf("revocation: %d %s", response.Code, response.Body.String())
	}
	if state, _ := engine.PrivacyStatus(context.Background()); state.Mode != PrivacyBlind {
		t.Fatalf("revocation must close LUNA's eyes, got %s", state.Mode)
	}
	if response := call(http.MethodDelete, path, "", false); response.Code != http.StatusOK || !bytes.Contains(response.Body.Bytes(), []byte(`"key_destroyed":true`)) {
		t.Fatalf("subject deletion: %d %s", response.Code, response.Body.String())
	}
}
