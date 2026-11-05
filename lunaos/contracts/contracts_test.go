package contracts

import (
	"sort"
	"testing"

	"lunaos.local/lunaos/services/intimacy"
)

func TestConsentScopesMatchVault(t *testing.T) {
	var vault []string
	for _, scope := range intimacy.KnownScopes() {
		vault = append(vault, string(scope))
	}
	got := append([]string(nil), ConsentScopes()...)
	sort.Strings(vault)
	sort.Strings(got)
	if len(got) != len(vault) {
		t.Fatalf("contract scopes %v, vault scopes %v", got, vault)
	}
	for i := range got {
		if got[i] != vault[i] {
			t.Fatalf("contract scopes %v, vault scopes %v", got, vault)
		}
	}
}

func TestActionsAreUniqueAndKnownPrivacy(t *testing.T) {
	actions := Actions()
	if len(actions) == 0 {
		t.Fatal("no actions")
	}
	seen := map[string]bool{}
	for _, action := range actions {
		if action.Name == "" || seen[action.Name] {
			t.Fatalf("empty or duplicate action %q", action.Name)
		}
		seen[action.Name] = true
		if action.PrivacyClass != "general" {
			t.Fatalf("action %q has unknown privacy class %q", action.Name, action.PrivacyClass)
		}
	}
}
