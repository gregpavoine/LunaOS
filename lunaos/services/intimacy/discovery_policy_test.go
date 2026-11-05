package intimacy

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDiscoveryAllowlistRequiresEveryConfiguredDimension(t *testing.T) {
	policy := DiscoveryAllowlist{
		Enabled: true,
		Sources: []string{"source-a"}, Creators: []string{"creator-a"},
		Categories: []string{"category-a"}, ModelFamilies: []string{"family-a"},
		Tags: []string{"tag-a", "tag-b"},
	}
	candidate := DiscoveryCandidate{Source: "source-a", Creator: "creator-a", Category: "category-a", ModelFamily: "family-a", Tags: []string{"tag-a"}}
	if decision := EvaluateDiscoveryCandidate(policy, candidate); !decision.Allowed || decision.Code != "ALLOWLIST_MATCH" {
		t.Fatalf("configured candidate decision=%+v", decision)
	}
	tests := []struct {
		name string
		edit func(*DiscoveryCandidate)
	}{
		{"unknown source", func(c *DiscoveryCandidate) { c.Source = "source-unknown" }},
		{"unknown creator", func(c *DiscoveryCandidate) { c.Creator = "creator-unknown" }},
		{"unknown category", func(c *DiscoveryCandidate) { c.Category = "category-unknown" }},
		{"unknown model family", func(c *DiscoveryCandidate) { c.ModelFamily = "family-unknown" }},
		{"unknown tag", func(c *DiscoveryCandidate) { c.Tags = append(c.Tags, "tag-unknown") }},
		{"empty metadata", func(c *DiscoveryCandidate) { c.Creator = " " }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidate := candidate
			candidate.Tags = append([]string(nil), candidate.Tags...)
			test.edit(&candidate)
			decision := EvaluateDiscoveryCandidate(policy, candidate)
			if decision.Allowed || decision.Code != "ALLOWLIST_MISS" {
				t.Fatalf("candidate bypassed deny-by-default policy: %+v", decision)
			}
		})
	}
	policy.Enabled = false
	if decision := EvaluateDiscoveryCandidate(policy, candidate); decision.Allowed || decision.Code != "POLICY_DISABLED" {
		t.Fatalf("disabled policy decision=%+v", decision)
	}
}

func TestDiscoveryAllowlistExpansionRequiresExplicitOwnerConfirmation(t *testing.T) {
	policy := DiscoveryAllowlist{Enabled: true, Sources: []string{"source-a"}, Creators: []string{"creator-a"}, Categories: []string{"category-a"}, ModelFamilies: []string{"family-a"}, Tags: []string{"tag-a"}}
	request := DiscoveryAllowlistExpansion{Sources: []string{"source-b"}, Creators: []string{"creator-b"}, Categories: []string{"category-b"}, ModelFamilies: []string{"family-b"}, Tags: []string{"tag-b"}}
	if _, err := ExpandDiscoveryAllowlist(policy, request, false); err != ErrDiscoveryExpansionUnconfirmed {
		t.Fatalf("unconfirmed expansion err=%v", err)
	}
	updated, err := ExpandDiscoveryAllowlist(policy, request, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Sources) != 2 || len(updated.Creators) != 2 || len(updated.Categories) != 2 || len(updated.ModelFamilies) != 2 || len(updated.Tags) != 2 {
		t.Fatalf("confirmed expansion=%+v", updated)
	}
	if len(policy.Sources) != 1 || policy.Sources[0] != "source-a" {
		t.Fatalf("expansion mutated previous policy: %+v", policy)
	}
}

func TestDiscoveryPolicyRequiresScopedConsentAndPersistsEncrypted(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "vault.sqlite")
	vault, err := Open(path, bytes.Repeat([]byte{0x71}, 32))
	if err != nil {
		t.Fatal(err)
	}
	defer vault.Close()
	ctx := context.Background()
	readOnly, err := vault.RecordOwnerConsent(ctx, "owner", []Scope{ScopeDiscoveryPolicyRead}, time.Now().Add(time.Hour), true, true)
	if err != nil {
		t.Fatal(err)
	}
	policy := DiscoveryAllowlist{Enabled: true, Sources: []string{"source-private"}, Creators: []string{"creator-private"}, Categories: []string{"category-private"}, ModelFamilies: []string{"family-private"}, Tags: []string{"tag-private"}}
	if err := vault.SetDiscoveryPolicy(ctx, readOnly.ID, policy, true); !errors.Is(err, ErrConsentDenied) {
		t.Fatalf("policy write without write scope err=%v", err)
	}
	if _, err := vault.EvaluateDiscovery(ctx, readOnly.ID, DiscoveryCandidate{Source: "source-private"}); !errors.Is(err, ErrConsentDenied) {
		t.Fatalf("evaluation without discovery.run scope err=%v", err)
	}
	consent, err := vault.RecordOwnerConsent(ctx, "owner", []Scope{ScopeDiscoveryPolicyRead, ScopeDiscoveryPolicyWrite, ScopeDiscoveryRun}, time.Now().Add(time.Hour), true, true)
	if err != nil {
		t.Fatal(err)
	}
	if err = vault.SetDiscoveryPolicy(ctx, consent.ID, policy, false); !errors.Is(err, ErrDiscoveryPolicyUnconfirmed) {
		t.Fatalf("unconfirmed save err=%v", err)
	}
	if err = vault.SetDiscoveryPolicy(ctx, consent.ID, policy, true); err != nil {
		t.Fatal(err)
	}
	loaded, err := vault.ReadDiscoveryPolicy(ctx, consent.ID)
	if err != nil || !loaded.Enabled || len(loaded.Sources) != 1 {
		t.Fatalf("loaded=%+v err=%v", loaded, err)
	}
	candidate := DiscoveryCandidate{Source: "source-private", Creator: "creator-private", Category: "category-private", ModelFamily: "family-private", Tags: []string{"tag-private"}}
	decision, err := vault.EvaluateDiscovery(ctx, consent.ID, candidate)
	if err != nil || !decision.Allowed {
		t.Fatalf("decision=%+v err=%v", decision, err)
	}
	miss, err := vault.EvaluateDiscovery(ctx, consent.ID, DiscoveryCandidate{Source: "unlisted"})
	if err != nil || miss.Allowed || miss.Code != "ALLOWLIST_MISS" {
		t.Fatalf("miss=%+v err=%v", miss, err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		data, readErr := os.ReadFile(filepath.Join(dir, entry.Name()))
		if readErr != nil {
			t.Fatal(readErr)
		}
		for _, secret := range []string{"source-private", "creator-private", "category-private", "tag-private"} {
			if bytes.Contains(data, []byte(secret)) {
				t.Fatalf("allowlist plaintext leaked through %s: %s", entry.Name(), secret)
			}
		}
	}
	if err = vault.RevokeConsent(ctx, consent.ID, "owner"); err != nil {
		t.Fatal(err)
	}
	if _, err = vault.EvaluateDiscovery(ctx, consent.ID, candidate); !errors.Is(err, ErrConsentDenied) {
		t.Fatalf("evaluation after revoke err=%v", err)
	}
}
