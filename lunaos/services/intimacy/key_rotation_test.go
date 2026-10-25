package intimacy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestRotateKeyReencryptsEveryPrivatePayloadAtomically(t *testing.T) {
	path := testVaultPath(t)
	oldKey := bytes.Repeat([]byte{0x81}, 32)
	newKey := bytes.Repeat([]byte{0x82}, 32)
	vault, err := Open(path, oldKey)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	consent, err := vault.RecordOwnerConsent(ctx, "owner", []Scope{
		ScopeMemoryRead, ScopeMemoryWrite, ScopePreferenceRead, ScopePreferenceWrite,
		ScopeDiscoveryPolicyRead, ScopeDiscoveryPolicyWrite, ScopeDiscoveryRun,
	}, time.Now().Add(time.Hour), true, true)
	if err != nil {
		t.Fatal(err)
	}
	memory, err := vault.StoreMemory(ctx, consent.ID, "owner", OriginUserAsserted, "rotation-memory-canary")
	if err != nil {
		t.Fatal(err)
	}
	preference, err := vault.AddPreference(ctx, consent.ID, PreferenceInput{Dimension: "visual_style", Value: "rotation-preference-canary", Strength: .8, Timescale: "recurring", Confirmed: true})
	if err != nil {
		t.Fatal(err)
	}
	policy := DiscoveryAllowlist{Enabled: true, Sources: []string{"rotation-source-canary"}, Creators: []string{"creator"}, Categories: []string{"category"}, ModelFamilies: []string{"family"}, Tags: []string{"tag"}}
	if err = vault.SetDiscoveryPolicy(ctx, consent.ID, policy, true); err != nil {
		t.Fatal(err)
	}
	if err = vault.RotateKey(newKey, true); err != nil {
		t.Fatal(err)
	}
	var rotations int
	if err = vault.db.QueryRow(`SELECT count(*) FROM vault_audit WHERE event_code='KEY_ROTATED'`).Scan(&rotations); err != nil || rotations != 1 {
		t.Fatalf("rotation audit count=%d err=%v", rotations, err)
	}
	readMemory, err := vault.ReadMemory(ctx, memory.ID, consent.ID, "owner")
	if err != nil || readMemory.Content != "rotation-memory-canary" {
		t.Fatalf("post-rotation memory=%+v err=%v", readMemory, err)
	}
	preferences, err := vault.ListPreferences(ctx, consent.ID)
	if err != nil || len(preferences) != 1 || preferences[0].Value != preference.Value {
		t.Fatalf("post-rotation preferences=%+v err=%v", preferences, err)
	}
	loadedPolicy, err := vault.ReadDiscoveryPolicy(ctx, consent.ID)
	if err != nil || !loadedPolicy.Enabled || loadedPolicy.Sources[0] != policy.Sources[0] {
		t.Fatalf("post-rotation policy=%+v err=%v", loadedPolicy, err)
	}
	if err = vault.Close(); err != nil {
		t.Fatal(err)
	}
	vault, err = Open(path, newKey)
	if err != nil {
		t.Fatal(err)
	}
	defer vault.Close()
	readMemory, err = vault.ReadMemory(ctx, memory.ID, consent.ID, "owner")
	if err != nil || readMemory.Content != "rotation-memory-canary" {
		t.Fatalf("reopened memory=%+v err=%v", readMemory, err)
	}
	preferences, err = vault.ListPreferences(ctx, consent.ID)
	if err != nil || len(preferences) != 1 || preferences[0].Value != preference.Value {
		t.Fatalf("reopened preferences=%+v err=%v", preferences, err)
	}
	loadedPolicy, err = vault.ReadDiscoveryPolicy(ctx, consent.ID)
	if err != nil || !loadedPolicy.Enabled || loadedPolicy.Sources[0] != policy.Sources[0] {
		t.Fatalf("reopened policy=%+v err=%v", loadedPolicy, err)
	}
	if err = vault.Close(); err != nil {
		t.Fatal(err)
	}
	oldVault, err := Open(path, oldKey)
	if err != nil {
		t.Fatal(err)
	}
	defer oldVault.Close()
	if _, err = oldVault.ReadMemory(ctx, memory.ID, consent.ID, "owner"); !errors.Is(err, ErrMemoryNotFound) {
		t.Fatalf("old key read memory err=%v", err)
	}
	if _, err = oldVault.ListPreferences(ctx, consent.ID); !errors.Is(err, ErrPreferenceNotFound) {
		t.Fatalf("old key read preferences err=%v", err)
	}
	if _, err = oldVault.ReadDiscoveryPolicy(ctx, consent.ID); !errors.Is(err, ErrInvalidDiscoveryAllowlist) {
		t.Fatalf("old key read policy err=%v", err)
	}
}

func TestRotateKeyRejectsInvalidKeyWithoutChangingCurrentVault(t *testing.T) {
	vault, err := Open(testVaultPath(t), bytes.Repeat([]byte{0x83}, 32))
	if err != nil {
		t.Fatal(err)
	}
	defer vault.Close()
	if err = vault.RotateKey(bytes.Repeat([]byte{0x86}, 32), false); !errors.Is(err, ErrVaultRotationUnconfirmed) {
		t.Fatalf("unconfirmed rotation err=%v", err)
	}
	if err = vault.RotateKey([]byte("too-short"), true); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("invalid rotation key err=%v", err)
	}
	ctx := context.Background()
	consent, err := vault.RecordOwnerConsent(ctx, "owner", []Scope{ScopePreferenceRead, ScopePreferenceWrite}, time.Now().Add(time.Hour), true, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = vault.AddPreference(ctx, consent.ID, PreferenceInput{Dimension: "mood", Value: "still-readable", Strength: .5, Timescale: "persistent", Confirmed: true}); err != nil {
		t.Fatalf("vault changed after rejected rotation: %v", err)
	}
}

func TestRotateKeyRollsBackAllRowsWhenAnyCiphertextIsInvalid(t *testing.T) {
	vault, err := Open(testVaultPath(t), bytes.Repeat([]byte{0x84}, 32))
	if err != nil {
		t.Fatal(err)
	}
	defer vault.Close()
	ctx := context.Background()
	consent, err := vault.RecordOwnerConsent(ctx, "owner", []Scope{ScopeMemoryRead}, time.Now().Add(time.Hour), true, true)
	if err != nil {
		t.Fatal(err)
	}
	firstID := "00000000-0000-4000-8000-000000000001"
	secondID := "00000000-0000-4000-8000-000000000002"
	origin := OriginUserAsserted
	plain, err := json.Marshal(struct {
		Content string `json:"content"`
	}{Content: "rollback-canary"})
	if err != nil {
		t.Fatal(err)
	}
	ciphertext, err := sealWithAEAD(vault.aead, plain, associatedData(firstID, "owner", origin))
	if err != nil {
		t.Fatal(err)
	}
	created := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err = vault.db.Exec(`INSERT INTO private_memory(id,subject_id,origin_type,ciphertext,created_at_utc) VALUES(?,?,?,?,?)`, firstID, "owner", origin, ciphertext, created); err != nil {
		t.Fatal(err)
	}
	if _, err = vault.db.Exec(`INSERT INTO private_memory(id,subject_id,origin_type,ciphertext,created_at_utc) VALUES(?,?,?,?,?)`, secondID, "owner", origin, []byte{0x01}, created); err != nil {
		t.Fatal(err)
	}
	if err = vault.RotateKey(bytes.Repeat([]byte{0x85}, 32), true); !errors.Is(err, ErrMemoryNotFound) {
		t.Fatalf("rotation with corrupt row err=%v", err)
	}
	read, err := vault.ReadMemory(ctx, firstID, consent.ID, "owner")
	if err != nil || read.Content != "rollback-canary" {
		t.Fatalf("valid row changed despite failed rotation: %+v err=%v", read, err)
	}
}
