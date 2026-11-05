// Package contracts holds the registry shared by every LunaOS track: the named
// actions and the per-person consent scopes. The JSON files are the source of
// truth; tests keep them aligned with the vault.
package contracts

import (
	_ "embed"
	"encoding/json"
)

//go:embed actions.json
var actionsJSON []byte

//go:embed consent-scopes.json
var consentScopesJSON []byte

type Action struct {
	Name         string `json:"name"`
	PrivacyClass string `json:"privacy_class"`
}

var (
	actions       []Action
	consentScopes []string
)

func init() {
	var a struct {
		Actions []Action `json:"actions"`
	}
	var s struct {
		Scopes []string `json:"scopes"`
	}
	if err := json.Unmarshal(actionsJSON, &a); err != nil {
		panic("contracts: actions.json: " + err.Error())
	}
	if err := json.Unmarshal(consentScopesJSON, &s); err != nil {
		panic("contracts: consent-scopes.json: " + err.Error())
	}
	actions, consentScopes = a.Actions, s.Scopes
}

// Actions returns a copy of the registered actions.
func Actions() []Action { return append([]Action(nil), actions...) }

// ConsentScopes returns a copy of the consent scope names.
func ConsentScopes() []string { return append([]string(nil), consentScopes...) }
