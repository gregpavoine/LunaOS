//go:build linux && integration

package core

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGoCoreAndPolicyBrokerProcessesRecoverQueuedGoal(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		t.Fatal(err)
	}
	token := hex.EncodeToString(tokenBytes)
	tokenPath := filepath.Join(dir, "owner-token")
	if err := os.WriteFile(tokenPath, []byte(token), 0600); err != nil {
		t.Fatal(err)
	}
	policyToken := strings.Repeat("b", 64)
	policyTokenPath := filepath.Join(dir, "policy-token")
	if err := os.WriteFile(policyTokenPath, []byte(policyToken), 0600); err != nil {
		t.Fatal(err)
	}

	brokerPath := filepath.Join(dir, "luna-policy-broker")
	corePath := filepath.Join(dir, "luna-core")
	build := func(output, target string) {
		t.Helper()
		command := exec.Command("go", "build", "-buildvcs=false", "-o", output, target)
		if result, err := command.CombinedOutput(); err != nil {
			t.Fatalf("build %s: %v\n%s", target, err, result)
		}
	}
	build(brokerPath, "../policy")
	build(corePath, "./cmd/luna-core")

	socket := filepath.Join(dir, "policy.sock")
	startTestProcess(t, brokerPath, "--socket="+socket, "--key-file="+policyTokenPath, "--client-user="+currentUsername(t))
	waitForCondition(t, "policy socket", func() bool {
		info, err := os.Lstat(socket)
		return err == nil && info.Mode()&os.ModeSocket != 0
	})

	database := filepath.Join(dir, "goals.sqlite")
	policy, err := NewUnixPolicyClient(socket, policyToken)
	if err != nil {
		t.Fatal(err)
	}
	engine, err := Open(database, token, policy, ChecksumRunner{})
	if err != nil {
		t.Fatal(err)
	}
	queued, err := engine.Create(context.Background(), Request{Action: "local.checksum"}, token, "preexisting-queued-goal")
	if err != nil {
		_ = engine.Close()
		t.Fatal(err)
	}
	if queued.Status != "approved" {
		_ = engine.Close()
		t.Fatalf("expected queued goal to be approved before daemon start, got %q", queued.Status)
	}
	if err := engine.Close(); err != nil {
		t.Fatal(err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	coreEnv := append(os.Environ(),
		"LUNA_OWNER_TOKEN_FILE="+tokenPath,
		"LUNA_DATABASE="+database,
		"LUNA_OWNER_SOCKET="+filepath.Join(dir, "owner-api.sock"),
		"LUNA_POLICY_SOCKET="+socket,
		"LUNA_POLICY_TOKEN_FILE="+policyTokenPath,
		"LUNA_LISTEN="+address,
	)
	startTestProcessWithEnv(t, corePath, coreEnv)
	waitForCondition(t, "owner API socket", func() bool {
		info, err := os.Lstat(filepath.Join(dir, "owner-api.sock"))
		return err == nil && info.Mode()&os.ModeSocket != 0 && info.Mode().Perm() == 0660
	})
	ownerClient := newOwnerSocketTestClient(filepath.Join(dir, "owner-api.sock"))
	ownerResponse, err := ownerClient.Get("http://luna/v1/goals")
	if err != nil {
		t.Fatal(err)
	}
	ownerResponse.Body.Close()
	if ownerResponse.StatusCode != http.StatusOK {
		t.Fatalf("owner Unix API status=%d", ownerResponse.StatusCode)
	}
	ownerGoalRequest, err := http.NewRequest(http.MethodPost, "http://luna/v1/goals", strings.NewReader(`{"action":"local.checksum","privacy_class":"general"}`))
	if err != nil {
		t.Fatal(err)
	}
	ownerGoalRequest.Header.Set("Content-Type", "application/json")
	ownerGoalRequest.Header.Set("Idempotency-Key", "owner-socket-goal")
	ownerGoalResponse, err := ownerClient.Do(ownerGoalRequest)
	if err != nil {
		t.Fatal(err)
	}
	var ownerGoal Goal
	if err := json.NewDecoder(ownerGoalResponse.Body).Decode(&ownerGoal); err != nil {
		t.Fatal(err)
	}
	ownerGoalResponse.Body.Close()
	if ownerGoalResponse.StatusCode != http.StatusCreated || ownerGoal.ID == "" {
		t.Fatalf("owner Unix create status=%d goal=%+v", ownerGoalResponse.StatusCode, ownerGoal)
	}

	baseURL := "http://" + address
	client := &http.Client{Timeout: time.Second}
	waitForCondition(t, "Core HTTP API", func() bool {
		response, err := client.Get(baseURL + "/v1/goals")
		if err != nil {
			return false
		}
		defer response.Body.Close()
		return response.StatusCode == http.StatusUnauthorized
	})
	assertGoalCompletesWithEvidence(t, client, baseURL, token, queued.ID)

	apiBody := `{"action":"local.checksum","privacy_class":"general"}`
	first := postGoal(t, client, baseURL, token, "process-api-idempotency", apiBody)
	second := postGoal(t, client, baseURL, token, "process-api-idempotency", apiBody)
	if first.ID == "" || first.ID != second.ID {
		t.Fatalf("process API did not preserve duplicate idempotency: first=%q second=%q", first.ID, second.ID)
	}
	assertGoalCompletesWithEvidence(t, client, baseURL, token, first.ID)

}

type testProcess struct {
	cmd    *exec.Cmd
	output bytes.Buffer
}

func currentUsername(t *testing.T) string {
	t.Helper()
	current, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	return current.Username
}

func startTestProcess(t *testing.T, executable string, args ...string) *testProcess {
	return startTestProcessWithEnv(t, executable, os.Environ(), args...)
}

func startTestProcessWithEnv(t *testing.T, executable string, env []string, args ...string) *testProcess {
	t.Helper()
	process := &testProcess{cmd: exec.Command(executable, args...)}
	process.cmd.Env = env
	process.cmd.Stdout = &process.output
	process.cmd.Stderr = &process.output
	if err := process.cmd.Start(); err != nil {
		t.Fatalf("start %s: %v", executable, err)
	}
	t.Cleanup(func() { stopTestProcess(t, process) })
	return process
}

func stopTestProcess(t *testing.T, process *testProcess) {
	t.Helper()
	if process == nil || process.cmd == nil || process.cmd.Process == nil {
		return
	}
	if process.cmd.ProcessState != nil {
		return
	}
	_ = process.cmd.Process.Signal(os.Interrupt)
	done := make(chan error, 1)
	go func() { done <- process.cmd.Wait() }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		_ = process.cmd.Process.Kill()
		<-done
	}
}

func waitForCondition(t *testing.T, description string, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(40 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", description)
}

func postGoal(t *testing.T, client *http.Client, baseURL, token, key, body string) Goal {
	t.Helper()
	request, err := http.NewRequest(http.MethodPost, baseURL+"/v1/goals", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", key)
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var goal Goal
	if err := json.NewDecoder(response.Body).Decode(&goal); err != nil {
		t.Fatalf("decode goal response: %v", err)
	}
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("create goal status=%d response=%+v", response.StatusCode, goal)
	}
	return goal
}

func assertGoalCompletesWithEvidence(t *testing.T, client *http.Client, baseURL, token, goalID string) {
	t.Helper()
	var last Goal
	waitForCondition(t, "completed goal and verified evidence", func() bool {
		request, err := http.NewRequest(http.MethodGet, baseURL+"/v1/goals/"+goalID, nil)
		if err != nil {
			return false
		}
		request.Header.Set("Authorization", "Bearer "+token)
		response, err := client.Do(request)
		if err != nil {
			return false
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK || json.NewDecoder(response.Body).Decode(&last) != nil {
			return false
		}
		return last.Status == "completed" && last.Evidence != nil && last.Evidence.Verified
	})
	if last.ID != goalID || last.Evidence.SHA256 == "" {
		t.Fatalf("goal completed without verifiable evidence: %+v", last)
	}
	if last.Evidence.ByteCount <= 0 || last.Evidence.Provenance != "local_public_fixture_v1" {
		t.Fatalf("unexpected evidence provenance: %+v", last.Evidence)
	}
}

func newOwnerSocketTestClient(socket string) *http.Client {
	transport := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{Timeout: time.Second}).DialContext(ctx, "unix", socket)
	}}
	return &http.Client{Timeout: time.Second, Transport: transport}
}
