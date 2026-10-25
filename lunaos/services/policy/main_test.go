package main

import (
	"bytes"
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

func TestUnixSocketPeerMustMatchBrokerUID(t *testing.T) {
	uid := os.Getuid()
	if !peerUIDAllowed(uid, uid) {
		t.Fatal("same-UID Core service must pass the broker peer check")
	}
	if peerUIDAllowed(uid+1, uid) {
		t.Fatal("different-UID process passed the broker peer check")
	}
}
