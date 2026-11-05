package core

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPrivacyStartsLockedAndRefusesCapture(t *testing.T) {
	e := testEngine(t, testDBPath(t), &testPolicy{})
	state, err := e.PrivacyStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if state.Mode != PrivacyStartupLocked || state.CaptureAllowed || !state.RequiresOpenEye {
		t.Fatalf("first boot must start locked without capture: %+v", state)
	}
	if _, err := e.RequestCaptureLease(context.Background(), "owner-secret", "blind_detection"); err != ErrPrivacyBlocked {
		t.Fatalf("capture lease while locked: err=%v", err)
	}
}

func TestOpenEyesNeedsOwnerOnTrustedUI(t *testing.T) {
	e := testEngine(t, testDBPath(t), &testPolicy{})
	ctx := context.Background()
	if _, err := e.OpenEyes(ctx, "wrong", true); err != ErrUnauthorized {
		t.Fatalf("non-owner reopened eyes: err=%v", err)
	}
	if _, err := e.OpenEyes(ctx, "owner-secret", false); err != ErrTrustedUIRequired {
		t.Fatalf("owner token without the trusted UI reopened eyes: err=%v", err)
	}
	state, err := e.OpenEyes(ctx, "owner-secret", true)
	if err != nil || state.Mode != PrivacyNormal || !state.CaptureAllowed {
		t.Fatalf("trusted owner click must open eyes: state=%+v err=%v", state, err)
	}
}

func TestBlindRevokesLeasesAndSurvivesRestart(t *testing.T) {
	path := testDBPath(t)
	e := testEngine(t, path, &testPolicy{})
	ctx := context.Background()
	if _, err := e.OpenEyes(ctx, "owner-secret", true); err != nil {
		t.Fatal(err)
	}
	lease, err := e.RequestCaptureLease(ctx, "owner-secret", "blind_detection")
	if err != nil {
		t.Fatal(err)
	}
	if valid, _ := e.LeaseValid(ctx, lease); !valid {
		t.Fatal("fresh detection lease must be valid in normal mode")
	}
	if _, err := e.RequestCaptureLease(ctx, "owner-secret", "consented_capture"); err != ErrPrivacyBlocked {
		t.Fatalf("consented capture outside an authorized session: err=%v", err)
	}
	for i := 0; i < 2; i++ {
		if _, err := e.EnterBlind(ctx, "intimate_scene"); err != nil {
			t.Fatal(err)
		}
	}
	if valid, _ := e.LeaseValid(ctx, lease); valid {
		t.Fatal("a lease issued before BLIND must stop validating")
	}
	var events int
	if err := e.db.QueryRow(`SELECT count(*) FROM privacy_event WHERE kind='privacy.blind_entered'`).Scan(&events); err != nil || events != 1 {
		t.Fatalf("repeated BLIND must record one transition: events=%d err=%v", events, err)
	}
	if _, err := e.db.Exec(`DELETE FROM privacy_event`); err == nil {
		t.Fatal("privacy audit must be append-only")
	}
	if err := e.Close(); err != nil {
		t.Fatal(err)
	}

	reopened := testEngine(t, path, &testPolicy{})
	state, err := reopened.PrivacyStatus(ctx)
	if err != nil || state.Mode != PrivacyBlind || state.CaptureAllowed {
		t.Fatalf("BLIND must persist across restart: state=%+v err=%v", state, err)
	}
	if _, err := reopened.OpenEyes(ctx, "owner-secret", true); err != nil {
		t.Fatal(err)
	}
	if valid, _ := reopened.LeaseValid(ctx, lease); valid {
		t.Fatal("reopening eyes must not revive a lease from before BLIND")
	}
	if _, err := reopened.EnterBlind(ctx, "because"); err == nil {
		t.Fatal("unknown BLIND reasons must be rejected")
	}
	forged := lease
	forged.Epoch++
	if valid, _ := reopened.LeaseValid(ctx, forged); valid {
		t.Fatal("a lease with an altered epoch must not validate")
	}
}

func TestHTTPOpenEyesOnlyFromOwnerSocket(t *testing.T) {
	e := testEngine(t, testDBPath(t), &testPolicy{})
	handler := NewHTTPHandler(e, "owner-secret")
	post := func(path, body string, trusted bool) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
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

	if response := post("/v1/privacy/open-eyes", `{"intent":"explicit_user_click"}`, false); response.Code != http.StatusForbidden {
		t.Fatalf("loopback bearer reopened eyes: %d %s", response.Code, response.Body.String())
	}
	if response := post("/v1/privacy/open-eyes", `{"intent":"agent"}`, true); response.Code != http.StatusBadRequest {
		t.Fatalf("non-click intent accepted: %d", response.Code)
	}
	if response := post("/v1/privacy/open-eyes", `{"intent":"explicit_user_click"}`, true); response.Code != http.StatusOK {
		t.Fatalf("trusted UI click refused: %d %s", response.Code, response.Body.String())
	}
	response := post("/v1/privacy/blind", `{"reason":"manual"}`, false)
	var state PrivacyState
	if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &state) != nil || state.Mode != PrivacyBlind {
		t.Fatalf("owner BLIND over loopback must work: %d %s", response.Code, response.Body.String())
	}
	if response := post("/v1/privacy/leases", `{"purpose":"blind_detection"}`, false); response.Code != http.StatusLocked {
		t.Fatalf("capture lease during BLIND: %d", response.Code)
	}
	unauthenticated := httptest.NewRequest(http.MethodPost, "/v1/privacy/blind", bytes.NewBufferString(`{"reason":"manual"}`))
	unauthenticated.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, unauthenticated)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated BLIND request: %d", recorder.Code)
	}
}
