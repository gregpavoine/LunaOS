package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func call(t *testing.T, b *broker, path, token string, req request) response {
	t.Helper()
	raw, _ := json.Marshal(req)
	r := httptest.NewRequest("POST", path, bytes.NewReader(raw))
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	b.ServeHTTP(w, r)
	var out response
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}
func TestOwnerAndBudget(t *testing.T) {
	b := &broker{key: []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")}
	for _, token := range []string{"", "guest", "ok je valide"} {
		if call(t, b, "/v1/grants", token, request{GoalID: "g", Action: "local.checksum", Privacy: "general"}).Allowed {
			t.Fatal("non-owner allowed")
		}
	}
	if r := call(t, b, "/v1/grants", string(b.key), request{GoalID: "g", Action: "payment.send", Privacy: "general"}); r.Code != "BUDGET_UNSET" || r.Allowed {
		t.Fatal(r)
	}
}
func TestScopedSignedGrant(t *testing.T) {
	b := &broker{key: []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")}
	req := request{GoalID: "g", Action: "local.checksum", Privacy: "general"}
	r := call(t, b, "/v1/grants", string(b.key), req)
	if !r.Allowed || r.Grant == "" {
		t.Fatal(r)
	}
	req.Grant = r.Grant
	if !call(t, b, "/v1/policy/evaluate", string(b.key), req).Allowed {
		t.Fatal("valid grant denied")
	}
	req.GoalID = "other"
	if call(t, b, "/v1/policy/evaluate", string(b.key), req).Allowed {
		t.Fatal("scope bypass")
	}
	req.GoalID = "g"
	req.Grant = "forged"
	if call(t, b, "/v1/policy/evaluate", string(b.key), req).Allowed {
		t.Fatal("forgery")
	}
	req.Grant = b.sign(grant{GoalID: "g", Action: "local.checksum", Principal: "owner", Expires: time.Now().Unix() - 1})
	if call(t, b, "/v1/policy/evaluate", string(b.key), req).Allowed {
		t.Fatal("expired accepted")
	}
	req.Grant = ""
	if call(t, b, "/v1/policy/evaluate", string(b.key), req).Allowed {
		t.Fatal("absent grant accepted")
	}
	req.Privacy = "intimate"
	if call(t, b, "/v1/grants", string(b.key), req).Allowed {
		t.Fatal("sensitive scope accepted")
	}
}
func TestUnknownAndMalformed(t *testing.T) {
	b := &broker{key: []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")}
	if call(t, b, "/v1/grants", string(b.key), request{GoalID: "g", Action: "shell", Privacy: "general"}).Allowed {
		t.Fatal("shell allowed")
	}
	r := httptest.NewRequest("GET", "/v1/grants", nil)
	r.Header.Set("Authorization", "Bearer "+string(b.key))
	w := httptest.NewRecorder()
	b.ServeHTTP(w, r)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatal(w.Code)
	}
}

func TestBrokerHealthRequiresCoreCredential(t *testing.T) {
	b := &broker{key: []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")}
	request := func(token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/v1/system/health", nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		response := httptest.NewRecorder()
		b.ServeHTTP(response, req)
		return response
	}
	if response := request("guest"); response.Code != http.StatusUnauthorized {
		t.Fatalf("guest health status=%d body=%s", response.Code, response.Body.String())
	}
	response := request(string(b.key))
	if response.Code != http.StatusOK || !bytes.Contains(response.Body.Bytes(), []byte(`"status":"ready"`)) || !bytes.Contains(response.Body.Bytes(), []byte(`"code":"BROKER_READY"`)) {
		t.Fatalf("Core health status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestUnixSocketPeerMustMatchConfiguredCoreUID(t *testing.T) {
	brokerUID := os.Getuid()
	coreUID := brokerUID + 1
	if !peerUIDAllowed(coreUID, coreUID) {
		t.Fatal("configured Core service UID must pass the broker peer check")
	}
	if peerUIDAllowed(brokerUID, coreUID) {
		t.Fatal("broker's own UID must not pass as the Core peer")
	}
	if peerUIDAllowed(coreUID+1, coreUID) {
		t.Fatal("unconfigured process passed the broker peer check")
	}
}

func TestOwnerTokenIsNotBrokerCredential(t *testing.T) {
	brokerToken := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	ownerToken := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	b := &broker{key: []byte(brokerToken)}
	if call(t, b, "/v1/grants", ownerToken, request{GoalID: "g", Action: "local.checksum", Privacy: "general"}).Allowed {
		t.Fatal("owner API credential also authenticated directly to policy broker")
	}
	if got := call(t, b, "/v1/grants", brokerToken, request{GoalID: "g", Action: "local.checksum", Privacy: "general"}); !got.Allowed {
		t.Fatalf("dedicated Core-to-Broker credential was rejected: %+v", got)
	}
}

func TestBrokerRequiresConfiguredCorePeerBeforeCredential(t *testing.T) {
	brokerUID := os.Getuid()
	coreUID := brokerUID + 1
	b := &broker{key: []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"), requirePeerUID: true, allowedPeerUID: coreUID}
	invoke := func(uid int) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, "/v1/grants", bytes.NewBufferString(`{"goal_id":"g","action":"local.checksum","privacy_class":"general"}`))
		r.Header.Set("Authorization", "Bearer "+string(b.key))
		r = r.WithContext(context.WithValue(r.Context(), peerUIDContextKey{}, uid))
		w := httptest.NewRecorder()
		b.ServeHTTP(w, r)
		return w
	}
	if response := invoke(coreUID); response.Code != http.StatusOK {
		t.Fatalf("configured Core UID status=%d body=%s", response.Code, response.Body.String())
	}
	if response := invoke(brokerUID); response.Code != http.StatusForbidden {
		t.Fatalf("Broker's own UID status=%d, want forbidden", response.Code)
	}
	health := func(uid int) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/v1/system/health", nil)
		req.Header.Set("Authorization", "Bearer "+string(b.key))
		req = req.WithContext(context.WithValue(req.Context(), peerUIDContextKey{}, uid))
		response := httptest.NewRecorder()
		b.ServeHTTP(response, req)
		return response
	}
	if response := health(coreUID); response.Code != http.StatusOK {
		t.Fatalf("configured Core UID health status=%d body=%s", response.Code, response.Body.String())
	}
	if response := health(brokerUID); response.Code != http.StatusForbidden {
		t.Fatalf("Broker's own UID health status=%d, want forbidden", response.Code)
	}
}

func TestBrokerRuntimeDirectoryAllowsRestrictedCoreGroupTraversal(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0750); err != nil {
		t.Fatal(err)
	}
	if err := private(dir, true); err != nil {
		t.Fatalf("restricted 0750 runtime directory rejected: %v", err)
	}
	if err := os.Chmod(dir, 0751); err != nil {
		t.Fatal(err)
	}
	if err := private(dir, true); err == nil {
		t.Fatal("runtime directory with world traversal was accepted")
	}
	if err := os.Chmod(dir, 0770); err != nil {
		t.Fatal(err)
	}
	if err := private(dir, true); err == nil {
		t.Fatal("runtime directory writable by the socket group was accepted")
	}
}
