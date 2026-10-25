package core

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"lunaos.local/bootstrap/services/intimacy"
)

func TestPreferenceAPIIsOwnerAuthenticatedConsentScopedAndCorrectable(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	vault, err := intimacy.Open(filepath.Join(dir, "vault.sqlite"), bytes.Repeat([]byte{0x51}, 32))
	if err != nil {
		t.Fatal(err)
	}
	defer vault.Close()
	handler := NewHTTPHandler(testEngine(t, testDBPath(t), &testPolicy{}), "owner-secret", vault)
	call := func(token, method, path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		return response
	}
	unauth := call("wrong", http.MethodPost, "/v1/intimacy/preferences", `{}`)
	if unauth.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status=%d", unauth.Code)
	}
	denied := call("owner-secret", http.MethodPost, "/v1/intimacy/preferences", `{"consent_id":"missing","dimension":"visual_style","value":"private-canary","strength":0.8,"timescale":"recurring","confirm_preference":true}`)
	if denied.Code != http.StatusForbidden {
		t.Fatalf("no consent status=%d body=%s", denied.Code, denied.Body.String())
	}
	consentResponse := call("owner-secret", http.MethodPost, "/v1/intimacy/consents", `{"scopes":["preferences.read","preferences.write"],"expires_at":"`+time.Now().Add(time.Hour).UTC().Format(time.RFC3339)+`","confirm_owner":true,"adult_attested":true}`)
	if consentResponse.Code != http.StatusCreated {
		t.Fatalf("consent status=%d body=%s", consentResponse.Code, consentResponse.Body.String())
	}
	var consent intimacy.Consent
	if err := json.Unmarshal(consentResponse.Body.Bytes(), &consent); err != nil {
		t.Fatal(err)
	}
	canary := "private-preference-canary"
	created := call("owner-secret", http.MethodPost, "/v1/intimacy/preferences", `{"consent_id":"`+consent.ID+`","dimension":"visual_style","value":"`+canary+`","strength":0.6,"timescale":"recurring","confirm_preference":true}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", created.Code, created.Body.String())
	}
	if strings.Contains(created.Body.String(), canary) {
		t.Fatal("create response exposed preference value")
	}
	var item intimacy.Preference
	if err := json.Unmarshal(created.Body.Bytes(), &item); err != nil {
		t.Fatal(err)
	}
	if item.SourceType != "declared" || item.Confidence != 1 || item.ConsentScope != "local_only" {
		t.Fatalf("provenance=%+v", item)
	}
	listed := call("owner-secret", http.MethodGet, "/v1/intimacy/preferences?consent_id="+consent.ID, "")
	if listed.Code != http.StatusOK || !strings.Contains(listed.Body.String(), canary) {
		t.Fatalf("list status=%d body=%s", listed.Code, listed.Body.String())
	}
	corrected := call("owner-secret", http.MethodPost, "/v1/intimacy/preferences/"+item.ID+"/correction", `{"consent_id":"`+consent.ID+`","value":"corrected-canary","strength":0.9,"disagreed_with_previous":true,"confirm_correction":true}`)
	if corrected.Code != http.StatusOK || !strings.Contains(corrected.Body.String(), `"source_type":"declared"`) || strings.Contains(corrected.Body.String(), "corrected-canary") {
		t.Fatalf("correct status=%d body=%s", corrected.Code, corrected.Body.String())
	}
	listed = call("owner-secret", http.MethodGet, "/v1/intimacy/preferences?consent_id="+consent.ID, "")
	if listed.Code != http.StatusOK || !strings.Contains(listed.Body.String(), "corrected-canary") || !strings.Contains(listed.Body.String(), `"negative_feedback_count":1`) {
		t.Fatalf("corrected list status=%d body=%s", listed.Code, listed.Body.String())
	}
	revoked := call("owner-secret", http.MethodPost, "/v1/intimacy/consents/"+consent.ID+"/revoke", `{}`)
	if revoked.Code != http.StatusOK {
		t.Fatalf("revoke status=%d", revoked.Code)
	}
	blocked := call("owner-secret", http.MethodGet, "/v1/intimacy/preferences?consent_id="+consent.ID, "")
	if blocked.Code != http.StatusForbidden {
		t.Fatalf("list after revocation status=%d", blocked.Code)
	}
	deleted := call("owner-secret", http.MethodDelete, "/v1/intimacy/preferences/"+item.ID, "")
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("delete after revocation status=%d body=%s", deleted.Code, deleted.Body.String())
	}
}
