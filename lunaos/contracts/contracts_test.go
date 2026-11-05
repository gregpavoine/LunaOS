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
		if err := validActionName(action.Name); err != nil {
			t.Fatalf("registered action is unsafe: %v", err)
		}
		seen[action.Name] = true
		if action.PrivacyClass != "general" {
			t.Fatalf("action %q has unknown privacy class %q", action.Name, action.PrivacyClass)
		}
	}
}

func TestValidActionName(t *testing.T) {
	for _, name := range []string{"local.checksum", "inference.generate", "memory.read_all", "a.b.c"} {
		if err := validActionName(name); err != nil {
			t.Errorf("%q rejected: %v", name, err)
		}
	}
	for _, name := range []string{
		"", "payment.send", "payment.", "payment.refund_all", "checksum", "Local.checksum",
		"local.Checksum", "local.check-sum", "local..x", ".local", "local.", "local.x1", "1local.x", "shell exec.run", "local.x\n",
	} {
		if err := validActionName(name); err == nil {
			t.Errorf("%q accepted", name)
		}
	}
}
