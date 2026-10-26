package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	apiBase        = "http://127.0.0.1:7766"
	expectedSHA256 = "07be6a1cc7b1b323fd3baca70abf34bfea13e09f9d28cf417d11aca77d4220ce"
	ownerTokenPath = "/var/lib/lunaos-identity/owner-token"
)

type capability struct {
	Status string `json:"status"`
}

type capabilities struct {
	Core         capability `json:"core"`
	Storage      capability `json:"storage"`
	PolicyBroker capability `json:"policy_broker"`
}

type evidence struct {
	SHA256     string `json:"sha256"`
	ByteCount  int64  `json:"byte_count"`
	Provenance string `json:"provenance"`
	Verified   bool   `json:"verified"`
}

type goal struct {
	ID       string    `json:"id"`
	Status   string    `json:"status"`
	Evidence *evidence `json:"evidence"`
}

func main() {
	if err := run(); err != nil {
		fmt.Printf("LUNAOS_INSTALL_SMOKE=FAIL reason=%q\n", err.Error())
		_ = exec.Command("/usr/bin/systemctl", "poweroff", "--no-block").Run()
		os.Exit(1)
	}
	fmt.Println("LUNAOS_INSTALL_SMOKE=PASS installed_boot=1 core=active broker=active goal=completed evidence=verified")
	if err := exec.Command("/usr/bin/systemctl", "poweroff", "--no-block").Run(); err != nil {
		fmt.Printf("LUNAOS_INSTALL_SMOKE=FAIL reason=%q\n", "poweroff request failed")
		os.Exit(1)
	}
}

func run() error {
	for _, unit := range []string{"luna-core.service", "luna-policy-broker.service"} {
		output, err := exec.Command("/usr/bin/systemctl", "is-active", "--quiet", unit).CombinedOutput()
		if err != nil {
			return fmt.Errorf("systemd unit %s is not active: %s", unit, strings.TrimSpace(string(output)))
		}
	}
	if err := removeTemporaryRepository("/etc/pacman.conf", "lunaos-ci"); err != nil {
		return fmt.Errorf("remove CI-only package repository: %w", err)
	}
	tokenBytes, err := os.ReadFile(ownerTokenPath)
	if err != nil {
		return fmt.Errorf("read owner token: %w", err)
	}
	token := strings.TrimSpace(string(tokenBytes))
	if token == "" {
		return fmt.Errorf("owner token is empty")
	}
	client := &http.Client{Timeout: 5 * time.Second}
	deadline := time.Now().Add(90 * time.Second)
	for time.Now().Before(deadline) {
		var state capabilities
		if err := requestJSON(client, http.MethodGet, apiBase+"/v1/system/capabilities", token, "", nil, &state, http.StatusOK); err == nil {
			if state.Core.Status == "ready" && state.Storage.Status == "ready" && state.PolicyBroker.Status == "ready" {
				return createAndVerifyGoal(client, token)
			}
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("Core and Policy Broker did not become ready before timeout")
}

func removeTemporaryRepository(path, repository string) error {
	contents, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var output strings.Builder
	skipping := false
	for _, line := range strings.Split(string(contents), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			skipping = trimmed == "["+repository+"]"
		}
		if !skipping {
			output.WriteString(line)
			output.WriteByte('\n')
		}
	}
	temporary := path + ".lunaos-ci"
	if err := os.WriteFile(temporary, []byte(output.String()), 0644); err != nil {
		return err
	}
	return os.Rename(temporary, path)
}

func createAndVerifyGoal(client *http.Client, token string) error {
	var created goal
	body := bytes.NewBufferString(`{"action":"local.checksum","privacy_class":"general"}`)
	key := fmt.Sprintf("install-smoke-%d", time.Now().UnixNano())
	if err := requestJSON(client, http.MethodPost, apiBase+"/v1/goals", token, key, body, &created, http.StatusCreated); err != nil {
		return fmt.Errorf("create owner-authorized checksum goal: %w", err)
	}
	if created.ID == "" {
		return fmt.Errorf("Core returned an empty goal id")
	}
	deadline := time.Now().Add(90 * time.Second)
	for time.Now().Before(deadline) {
		var current goal
		if err := requestJSON(client, http.MethodGet, apiBase+"/v1/goals/"+created.ID, token, "", nil, &current, http.StatusOK); err == nil {
			if current.Status == "failed" || current.Status == "blocked" {
				return fmt.Errorf("checksum goal ended in %s", current.Status)
			}
			if current.Status == "completed" {
				if current.Evidence == nil || !current.Evidence.Verified || current.Evidence.SHA256 != expectedSHA256 || current.Evidence.ByteCount != 29 || current.Evidence.Provenance != "local_public_fixture_v1" {
					return fmt.Errorf("completed goal evidence did not match the public fixture contract")
				}
				return nil
			}
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("checksum goal did not complete before timeout")
}

func requestJSON(client *http.Client, method, url, token, idempotencyKey string, body io.Reader, output any, expectedStatus int) error {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if idempotencyKey != "" {
		request.Header.Set("Idempotency-Key", idempotencyKey)
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != expectedStatus {
		payload, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("HTTP %d: %s", response.StatusCode, strings.TrimSpace(string(payload)))
	}
	if output != nil {
		return json.NewDecoder(response.Body).Decode(output)
	}
	return nil
}
