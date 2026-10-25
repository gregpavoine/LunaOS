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

func TestDiscoveryPolicyAPIRequiresOwnerConfirmationAndFailsClosed(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	vault, err := intimacy.Open(filepath.Join(dir, "vault.sqlite"), bytes.Repeat([]byte{0x72}, 32))
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
	unauthorized := call("wrong", http.MethodPost, "/v1/intimacy/discovery/evaluate", `{}`)
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status=%d", unauthorized.Code)
	}
	consent := call("owner-secret", http.MethodPost, "/v1/intimacy/consents", `{"scopes":["discovery.policy.read","discovery.policy.write","discovery.run"],"expires_at":"`+time.Now().Add(time.Hour).UTC().Format(time.RFC3339)+`","confirm_owner":true,"adult_attested":true}`)
	if consent.Code != http.StatusCreated {
		t.Fatalf("consent status=%d body=%s", consent.Code, consent.Body.String())
	}
	var grant struct {
		ID string `json:"consent_id"`
	}
	if err := json.Unmarshal(consent.Body.Bytes(), &grant); err != nil {
		t.Fatal(err)
	}
	readGrantResponse := call("owner-secret", http.MethodPost, "/v1/intimacy/consents", `{"scopes":["discovery.policy.read"],"expires_at":"`+time.Now().Add(time.Hour).UTC().Format(time.RFC3339)+`","confirm_owner":true,"adult_attested":true}`)
	var readGrant struct {
		ID string `json:"consent_id"`
	}
	if readGrantResponse.Code != http.StatusCreated || json.Unmarshal(readGrantResponse.Body.Bytes(), &readGrant) != nil {
		t.Fatalf("read-only consent status=%d body=%s", readGrantResponse.Code, readGrantResponse.Body.String())
	}
	readPolicy := call("owner-secret", http.MethodGet, "/v1/intimacy/discovery/policy?consent_id="+readGrant.ID, "")
	if readPolicy.Code != http.StatusOK {
		t.Fatalf("policy read with read scope status=%d body=%s", readPolicy.Code, readPolicy.Body.String())
	}
	readOnlyEvaluation := call("owner-secret", http.MethodPost, "/v1/intimacy/discovery/evaluate", `{"consent_id":"`+readGrant.ID+`","candidate":{"source":"source-a","creator":"creator-a","category":"category-a","model_family":"family-a","tags":["tag-a"]}}`)
	if readOnlyEvaluation.Code != http.StatusForbidden {
		t.Fatalf("evaluation without discovery.run scope status=%d body=%s", readOnlyEvaluation.Code, readOnlyEvaluation.Body.String())
	}
	initial := call("owner-secret", http.MethodPost, "/v1/intimacy/discovery/evaluate", `{"consent_id":"`+grant.ID+`","candidate":{"source":"source-a","creator":"creator-a","category":"category-a","model_family":"family-a","tags":["tag-a"]}}`)
	if initial.Code != http.StatusOK || !strings.Contains(initial.Body.String(), `"code":"POLICY_DISABLED"`) {
		t.Fatalf("default decision status=%d body=%s", initial.Code, initial.Body.String())
	}
	policy := `"policy":{"enabled":true,"sources":["source-a"],"creators":["creator-a"],"categories":["category-a"],"model_families":["family-a"],"tags":["tag-a"]}`
	unconfirmed := call("owner-secret", http.MethodPut, "/v1/intimacy/discovery/policy", `{"consent_id":"`+grant.ID+`",`+policy+`,"confirm_owner":false}`)
	if unconfirmed.Code != http.StatusBadRequest {
		t.Fatalf("unconfirmed update status=%d body=%s", unconfirmed.Code, unconfirmed.Body.String())
	}
	updated := call("owner-secret", http.MethodPut, "/v1/intimacy/discovery/policy", `{"consent_id":"`+grant.ID+`",`+policy+`,"confirm_owner":true}`)
	if updated.Code != http.StatusNoContent {
		t.Fatalf("policy update status=%d body=%s", updated.Code, updated.Body.String())
	}
	allowed := call("owner-secret", http.MethodPost, "/v1/intimacy/discovery/evaluate", `{"consent_id":"`+grant.ID+`","candidate":{"source":"source-a","creator":"creator-a","category":"category-a","model_family":"family-a","tags":["tag-a"]}}`)
	if allowed.Code != http.StatusOK || !strings.Contains(allowed.Body.String(), `"allowed":true`) {
		t.Fatalf("allowlisted decision status=%d body=%s", allowed.Code, allowed.Body.String())
	}
	denied := call("owner-secret", http.MethodPost, "/v1/intimacy/discovery/evaluate", `{"consent_id":"`+grant.ID+`","candidate":{"source":"unlisted","creator":"creator-a","category":"category-a","model_family":"family-a","tags":["tag-a"]}}`)
	if denied.Code != http.StatusOK || !strings.Contains(denied.Body.String(), `"code":"ALLOWLIST_MISS"`) {
		t.Fatalf("unlisted decision status=%d body=%s", denied.Code, denied.Body.String())
	}
	revoked := call("owner-secret", http.MethodPost, "/v1/intimacy/consents/"+grant.ID+"/revoke", `{}`)
	if revoked.Code != http.StatusOK {
		t.Fatalf("revoke status=%d body=%s", revoked.Code, revoked.Body.String())
	}
	afterRevoke := call("owner-secret", http.MethodPost, "/v1/intimacy/discovery/evaluate", `{"consent_id":"`+grant.ID+`","candidate":{"source":"source-a","creator":"creator-a","category":"category-a","model_family":"family-a","tags":["tag-a"]}}`)
	if afterRevoke.Code != http.StatusForbidden {
		t.Fatalf("evaluation after revoke status=%d body=%s", afterRevoke.Code, afterRevoke.Body.String())
	}
}
