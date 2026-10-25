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

func TestIntimacyMemoryAPIRequiresConsentAndSupportsRevocationAndDeletion(t *testing.T) {
	privateDir := t.TempDir()
	if err := os.Chmod(privateDir, 0700); err != nil {
		t.Fatal(err)
	}
	vault, err := intimacy.Open(filepath.Join(privateDir, "vault.sqlite"), bytes.Repeat([]byte{0x31}, 32))
	if err != nil {
		t.Fatal(err)
	}
	defer vault.Close()
	engine := testEngine(t, testDBPath(t), &testPolicy{})
	handler := NewHTTPHandler(engine, "owner-secret", vault)
	call := func(method, path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer owner-secret")
		req.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		return response
	}
	blocked := call(http.MethodPost, "/v1/intimacy/memories", `{"consent_id":"missing","origin_type":"user_asserted","content":"synthetic-canary"}`)
	if blocked.Code != http.StatusForbidden {
		t.Fatalf("write without consent status=%d body=%s", blocked.Code, blocked.Body.String())
	}
	consentResponse := call(http.MethodPost, "/v1/intimacy/consents", `{"scopes":["memory.read","memory.write"],"expires_at":"`+time.Now().Add(time.Hour).UTC().Format(time.RFC3339)+`","confirm_owner":true,"adult_attested":true}`)
	if consentResponse.Code != http.StatusCreated {
		t.Fatalf("consent status=%d body=%s", consentResponse.Code, consentResponse.Body.String())
	}
	var consent intimacy.Consent
	if err := json.Unmarshal(consentResponse.Body.Bytes(), &consent); err != nil {
		t.Fatal(err)
	}
	canary := "synthetic-private-canary-route"
	write := call(http.MethodPost, "/v1/intimacy/memories", `{"consent_id":"`+consent.ID+`","origin_type":"user_asserted","content":"`+canary+`"}`)
	if write.Code != http.StatusCreated {
		t.Fatalf("write status=%d body=%s", write.Code, write.Body.String())
	}
	var memory struct {
		ID string `json:"memory_id"`
	}
	if err := json.Unmarshal(write.Body.Bytes(), &memory); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(write.Body.String(), canary) {
		t.Fatal("write response echoed intimate plaintext")
	}
	read := call(http.MethodGet, "/v1/intimacy/memories/"+memory.ID+"?consent_id="+consent.ID, "")
	if read.Code != http.StatusOK || !strings.Contains(read.Body.String(), canary) {
		t.Fatalf("read status=%d body=%s", read.Code, read.Body.String())
	}
	revoke := call(http.MethodPost, "/v1/intimacy/consents/"+consent.ID+"/revoke", `{}`)
	if revoke.Code != http.StatusOK {
		t.Fatalf("revoke status=%d body=%s", revoke.Code, revoke.Body.String())
	}
	denied := call(http.MethodGet, "/v1/intimacy/memories/"+memory.ID+"?consent_id="+consent.ID, "")
	if denied.Code != http.StatusForbidden {
		t.Fatalf("read after revoke status=%d", denied.Code)
	}
	deleted := call(http.MethodDelete, "/v1/intimacy/memories/"+memory.ID, "")
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("delete after revoke status=%d body=%s", deleted.Code, deleted.Body.String())
	}
}

func TestIntimacyMemoryAPIRequiresOwnerAndAdultAttestation(t *testing.T) {
	privateDir := t.TempDir()
	if err := os.Chmod(privateDir, 0700); err != nil {
		t.Fatal(err)
	}
	vault, err := intimacy.Open(filepath.Join(privateDir, "vault.sqlite"), bytes.Repeat([]byte{0x32}, 32))
	if err != nil {
		t.Fatal(err)
	}
	defer vault.Close()
	engine := testEngine(t, testDBPath(t), &testPolicy{})
	handler := NewHTTPHandler(engine, "owner-secret", vault)
	request := httptest.NewRequest(http.MethodPost, "/v1/intimacy/consents", strings.NewReader(`{"scopes":["memory.write"],"expires_at":"`+time.Now().Add(time.Hour).UTC().Format(time.RFC3339)+`","confirm_owner":true,"adult_attested":false}`))
	request.Header.Set("Authorization", "Bearer owner-secret")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("no adult attestation status=%d body=%s", response.Code, response.Body.String())
	}
}
