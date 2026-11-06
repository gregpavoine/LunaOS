package main

import (
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunDoesNotPrintCredentialUnlessRequested(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "owner-token")
	policyPath := filepath.Join(dir, "policy-token")
	var output, errorsOutput bytes.Buffer
	if code := run([]string{"--path", path, "--policy-path", policyPath}, &output, &errorsOutput); code != 0 {
		t.Fatalf("run exit=%d stderr=%s", code, errorsOutput.String())
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	token := strings.TrimSpace(string(contents))
	if _, err := hex.DecodeString(token); err != nil || len(token) != 64 {
		t.Fatalf("generated token format invalid: %v", err)
	}
	if strings.Contains(output.String(), token) {
		t.Fatal("credential was printed without --reveal-token")
	}
}

func TestRunExplicitlyRevealsCredential(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "owner-token")
	policyPath := filepath.Join(dir, "policy-token")
	var output, errorsOutput bytes.Buffer
	if code := run([]string{"--path", path, "--policy-path", policyPath, "--reveal-token"}, &output, &errorsOutput); code != 0 {
		t.Fatalf("run exit=%d stderr=%s", code, errorsOutput.String())
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	token := strings.TrimSpace(string(contents))
	if !strings.Contains(output.String(), "\n"+token+"\n") {
		t.Fatalf("explicit reveal did not print generated credential")
	}
}
