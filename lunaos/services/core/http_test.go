package core

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHTTPGoalLifecycleRequiresOwnerAndReturnsEvidence(t *testing.T) {
	e := testEngine(t, testDBPath(t), &testPolicy{})
	handler := NewHTTPHandler(e, "owner-secret")

	unauthorized := httptest.NewRecorder()
	handler.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/v1/goals", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized list status=%d body=%s", unauthorized.Code, unauthorized.Body.String())
	}

	body, _ := json.Marshal(Request{Action: "local.checksum"})
	create := httptest.NewRequest(http.MethodPost, "/v1/goals", bytes.NewReader(body))
	create.Header.Set("Authorization", "Bearer owner-secret")
	create.Header.Set("Content-Type", "application/json")
	create.Header.Set("Idempotency-Key", "http-create-1")
	created := httptest.NewRecorder()
	handler.ServeHTTP(created, create)
	if created.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", created.Code, created.Body.String())
	}
	var goal Goal
	if err := json.Unmarshal(created.Body.Bytes(), &goal); err != nil {
		t.Fatal(err)
	}

	tick := httptest.NewRequest(http.MethodPost, "/v1/scheduler/tick", bytes.NewBufferString("{}"))
	tick.Header.Set("Authorization", "Bearer owner-secret")
	tick.Header.Set("Content-Type", "application/json")
	tickResponse := httptest.NewRecorder()
	handler.ServeHTTP(tickResponse, tick)
	if tickResponse.Code != http.StatusOK {
		t.Fatalf("tick status=%d body=%s", tickResponse.Code, tickResponse.Body.String())
	}

	get := httptest.NewRequest(http.MethodGet, "/v1/goals/"+goal.ID, nil)
	get.Header.Set("Authorization", "Bearer owner-secret")
	got := httptest.NewRecorder()
	handler.ServeHTTP(got, get)
	var completed Goal
	if got.Code != http.StatusOK || json.Unmarshal(got.Body.Bytes(), &completed) != nil {
		t.Fatalf("get status=%d body=%s", got.Code, got.Body.String())
	}
	if completed.Status != "completed" || completed.Evidence == nil || !completed.Evidence.Verified {
		t.Fatalf("completed goal lacks verified evidence: %+v", completed)
	}
}

func TestHTTPRejectsUnknownFieldsAndBrowserOrigins(t *testing.T) {
	e := testEngine(t, testDBPath(t), &testPolicy{})
	handler := NewHTTPHandler(e, "owner-secret")
	unknown := httptest.NewRequest(http.MethodPost, "/v1/goals", bytes.NewBufferString(`{"action":"local.checksum","owner":true}`))
	unknown.Header.Set("Authorization", "Bearer owner-secret")
	unknown.Header.Set("Content-Type", "application/json")
	unknown.Header.Set("Idempotency-Key", "http-invalid-1")
	unknownResponse := httptest.NewRecorder()
	handler.ServeHTTP(unknownResponse, unknown)
	if unknownResponse.Code != http.StatusBadRequest {
		t.Fatalf("unknown field status=%d body=%s", unknownResponse.Code, unknownResponse.Body.String())
	}

	origin := httptest.NewRequest(http.MethodPost, "/v1/goals", bytes.NewBufferString(`{"action":"local.checksum"}`))
	origin.Header.Set("Authorization", "Bearer owner-secret")
	origin.Header.Set("Content-Type", "application/json")
	origin.Header.Set("Idempotency-Key", "http-invalid-2")
	origin.Header.Set("Origin", "https://example.invalid")
	originResponse := httptest.NewRecorder()
	handler.ServeHTTP(originResponse, origin)
	if originResponse.Code != http.StatusUnauthorized {
		t.Fatalf("browser origin status=%d body=%s", originResponse.Code, originResponse.Body.String())
	}
}

func TestHTTPGoalControlsRequireOwnerAndEnforceTransitions(t *testing.T) {
	e := testEngine(t, testDBPath(t), &testPolicy{})
	handler := NewHTTPHandler(e, "owner-secret")
	goal, err := e.Create(context.Background(), Request{Action: "local.checksum"}, "owner-secret", "http-control-1")
	if err != nil {
		t.Fatal(err)
	}
	request := func(token, operation string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPatch, "/v1/goals/"+goal.ID, bytes.NewBufferString(`{"operation":"`+operation+`"}`))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		return response
	}
	if response := request("guest", "pause"); response.Code != http.StatusUnauthorized {
		t.Fatalf("guest pause status=%d", response.Code)
	}
	if response := request("owner-secret", "pause"); response.Code != http.StatusOK || !bytes.Contains(response.Body.Bytes(), []byte(`"status":"paused"`)) {
		t.Fatalf("pause status=%d body=%s", response.Code, response.Body.String())
	}
	if response := request("owner-secret", "resume"); response.Code != http.StatusOK || !bytes.Contains(response.Body.Bytes(), []byte(`"status":"approved"`)) {
		t.Fatalf("resume status=%d body=%s", response.Code, response.Body.String())
	}
	if response := request("owner-secret", "cancel"); response.Code != http.StatusOK || !bytes.Contains(response.Body.Bytes(), []byte(`"status":"cancelled"`)) {
		t.Fatalf("cancel status=%d body=%s", response.Code, response.Body.String())
	}
	if response := request("owner-secret", "resume"); response.Code != http.StatusConflict {
		t.Fatalf("invalid transition status=%d body=%s", response.Code, response.Body.String())
	}
	if response := request("owner-secret", "delete"); response.Code != http.StatusBadRequest {
		t.Fatalf("invalid operation status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestHTTPGoalPlanIsOwnerOnlyAndValidatesBody(t *testing.T) {
	e := testEngine(t, testDBPath(t), &testPolicy{})
	goal, err := e.Create(context.Background(), Request{Action: "local.checksum"}, "owner-secret", "http-plan-goal")
	if err != nil {
		t.Fatal(err)
	}
	dependency, err := e.Create(context.Background(), Request{Action: "local.checksum"}, "owner-secret", "http-plan-dependency")
	if err != nil {
		t.Fatal(err)
	}
	handler := NewHTTPHandler(e, "owner-secret")
	request := func(token, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/v1/goals/"+goal.ID+"/plan", bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		return response
	}
	if response := request("guest", `{"dependency_goal_ids":["`+dependency.ID+`"]}`); response.Code != http.StatusUnauthorized {
		t.Fatalf("guest plan status=%d body=%s", response.Code, response.Body.String())
	}
	if response := request("owner-secret", `{"dependency_goal_ids":["`+dependency.ID+`"],"unexpected":true}`); response.Code != http.StatusBadRequest {
		t.Fatalf("unknown plan field status=%d body=%s", response.Code, response.Body.String())
	}
	if response := request("owner-secret", `{"dependency_goal_ids":["`+dependency.ID+`"]}`); response.Code != http.StatusOK || !bytes.Contains(response.Body.Bytes(), []byte(dependency.ID)) {
		t.Fatalf("owner plan status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestHTTPReprioritizeRequiresOwnerAndAllMetrics(t *testing.T) {
	e := testEngine(t, testDBPath(t), &testPolicy{})
	goal, err := e.Create(context.Background(), Request{Action: "local.checksum"}, "owner-secret", "http-reprioritize")
	if err != nil {
		t.Fatal(err)
	}
	handler := NewHTTPHandler(e, "owner-secret")
	request := func(token, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPatch, "/v1/goals/"+goal.ID, bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		return response
	}
	body := `{"operation":"reprioritize","metrics":{"impact":0.2,"urgency":0.2,"user_value":0.2,"confidence":0.2,"efficiency":0.2}}`
	if response := request("guest", body); response.Code != http.StatusUnauthorized {
		t.Fatalf("guest reprioritize status=%d body=%s", response.Code, response.Body.String())
	}
	if response := request("owner-secret", `{"operation":"reprioritize"}`); response.Code != http.StatusBadRequest {
		t.Fatalf("missing metrics status=%d body=%s", response.Code, response.Body.String())
	}
	if response := request("owner-secret", body); response.Code != http.StatusOK || !bytes.Contains(response.Body.Bytes(), []byte(`"score":20`)) {
		t.Fatalf("reprioritize status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestHTTPGoalMetadataEditIsOwnerOnly(t *testing.T) {
	e := testEngine(t, testDBPath(t), &testPolicy{})
	goal, err := e.Create(context.Background(), Request{Action: "local.checksum"}, "owner-secret", "http-metadata-edit")
	if err != nil {
		t.Fatal(err)
	}
	handler := NewHTTPHandler(e, "owner-secret")
	request := func(token, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPatch, "/v1/goals/"+goal.ID, bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		return response
	}
	body := `{"operation":"edit","title":"My goal","description":"Owner metadata"}`
	if response := request("guest", body); response.Code != http.StatusUnauthorized {
		t.Fatalf("guest edit status=%d body=%s", response.Code, response.Body.String())
	}
	if response := request("owner-secret", `{"operation":"edit"}`); response.Code != http.StatusBadRequest {
		t.Fatalf("empty edit status=%d body=%s", response.Code, response.Body.String())
	}
	if response := request("owner-secret", body); response.Code != http.StatusOK || !bytes.Contains(response.Body.Bytes(), []byte(`"title":"My goal"`)) {
		t.Fatalf("owner edit status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestIntimacyStopIsOwnerOnlyAndPersistsAcrossRestart(t *testing.T) {
	path := testDBPath(t)
	policy := &testPolicy{}
	e := testEngine(t, path, policy)
	handler := NewHTTPHandler(e, "owner-secret")

	unauthorized := httptest.NewRecorder()
	handler.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodPost, "/v1/intimacy/stop", bytes.NewBufferString(`{}`)))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized stop status=%d", unauthorized.Code)
	}

	stop := httptest.NewRequest(http.MethodPost, "/v1/intimacy/stop", bytes.NewBufferString(`{}`))
	stop.Header.Set("Authorization", "Bearer owner-secret")
	stop.Header.Set("Content-Type", "application/json")
	stopped := httptest.NewRecorder()
	handler.ServeHTTP(stopped, stop)
	if stopped.Code != http.StatusOK || !bytes.Contains(stopped.Body.Bytes(), []byte(`"stopped":true`)) {
		t.Fatalf("stop response status=%d body=%s", stopped.Code, stopped.Body.String())
	}

	conversation := httptest.NewRequest(http.MethodPost, "/v1/intimacy/conversation", bytes.NewBufferString(`{"message":"hello"}`))
	conversation.Header.Set("Authorization", "Bearer owner-secret")
	conversation.Header.Set("Content-Type", "application/json")
	denied := httptest.NewRecorder()
	handler.ServeHTTP(denied, conversation)
	if denied.Code != http.StatusLocked {
		t.Fatalf("stopped conversation status=%d body=%s", denied.Code, denied.Body.String())
	}

	if err := e.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(path, "owner-secret", policy, ChecksumRunner{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = reopened.Close() })
	state, err := reopened.IntimacyState(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !state.Stopped || state.PrivacyMode != "unknown" || state.BackendAvailable {
		t.Fatalf("unsafe intimacy state after restart: %+v", state)
	}
}

func TestOwnerBearerMayBeOmittedOnlyOnTrustedUnixConnection(t *testing.T) {
	e := testEngine(t, testDBPath(t), &testPolicy{})
	handler := NewHTTPHandler(e, "owner-secret")
	temporaryFile, err := os.CreateTemp("/tmp", "luna-")
	if err != nil {
		t.Fatal(err)
	}
	path := temporaryFile.Name() + ".sock"
	if err := temporaryFile.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(temporaryFile.Name()); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: path, Net: "unix"})
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	client, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: path, Net: "unix"})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	server, err := listener.AcceptUnix()
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	request := httptest.NewRequest(http.MethodGet, "/v1/goals", nil)
	request = request.WithContext(OwnerSocketConnContext(request.Context(), server))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("owner Unix connection status=%d body=%s", response.Code, response.Body.String())
	}

	ordinary := httptest.NewRecorder()
	handler.ServeHTTP(ordinary, httptest.NewRequest(http.MethodGet, "/v1/goals", nil))
	if ordinary.Code != http.StatusUnauthorized {
		t.Fatalf("ordinary HTTP without bearer status=%d", ordinary.Code)
	}
}
