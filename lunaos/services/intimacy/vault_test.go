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

func TestVaultEncryptsMemoryAndRequiresScopedConsent(t *testing.T) {
	path := testVaultPath(t)
	key := bytes.Repeat([]byte{0x42}, 32)
	vault, err := Open(path, key)
	if err != nil {
		t.Fatal(err)
	}
	defer vault.Close()

	consent, err := vault.RecordOwnerConsent(context.Background(), "owner", []Scope{ScopeMemoryRead, ScopeMemoryWrite}, time.Now().Add(time.Hour), true, true)
	if err != nil {
		t.Fatal(err)
	}
	content := "synthetic-private-canary"
	memory, err := vault.StoreMemory(context.Background(), consent.ID, "owner", OriginUserAsserted, content)
	if err != nil {
		t.Fatal(err)
	}
	if memory.Content != content || memory.Origin != OriginUserAsserted {
		t.Fatalf("memory round trip = %+v", memory)
	}
	bytesOnDisk, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(bytesOnDisk, []byte(content)) {
		t.Fatal("private memory leaked in plaintext to the SQLite file")
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		data, readErr := os.ReadFile(filepath.Join(filepath.Dir(path), entry.Name()))
		if readErr != nil {
			t.Fatal(readErr)
		}
		if bytes.Contains(data, []byte(content)) {
			t.Fatalf("private memory leaked in plaintext to SQLite sidecar %q", entry.Name())
		}
	}
	readOnly, err := vault.RecordOwnerConsent(context.Background(), "owner", []Scope{ScopeMemoryRead}, time.Now().Add(time.Hour), true, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = vault.StoreMemory(context.Background(), readOnly.ID, "owner", OriginUserAsserted, "unauthorized"); !errors.Is(err, ErrConsentDenied) {
		t.Fatalf("write without write scope error = %v", err)
	}
	if _, err = vault.ReadMemory(context.Background(), memory.ID, consent.ID, "other-subject"); !errors.Is(err, ErrConsentDenied) {
		t.Fatalf("cross-subject read error = %v", err)
	}
}

func TestVaultRevocationBlocksReadAndWriteButDeletionRemainsAvailable(t *testing.T) {
	vault, err := Open(testVaultPath(t), bytes.Repeat([]byte{0x55}, 32))
	if err != nil {
		t.Fatal(err)
	}
	defer vault.Close()
	consent, err := vault.RecordOwnerConsent(context.Background(), "owner", []Scope{ScopeMemoryRead, ScopeMemoryWrite}, time.Now().Add(time.Hour), true, true)
	if err != nil {
		t.Fatal(err)
	}
	memory, err := vault.StoreMemory(context.Background(), consent.ID, "owner", OriginFictionalRoleplay, "fictional synthetic event")
	if err != nil {
		t.Fatal(err)
	}
	if err = vault.RevokeConsent(context.Background(), consent.ID, "owner"); err != nil {
		t.Fatal(err)
	}
	if _, err = vault.ReadMemory(context.Background(), memory.ID, consent.ID, "owner"); !errors.Is(err, ErrConsentDenied) {
		t.Fatalf("read after revocation error = %v", err)
	}
	if _, err = vault.StoreMemory(context.Background(), consent.ID, "owner", OriginUserAsserted, "new data"); !errors.Is(err, ErrConsentDenied) {
		t.Fatalf("write after revocation error = %v", err)
	}
	if err = vault.DeleteMemory(context.Background(), memory.ID, "owner"); err != nil {
		t.Fatal(err)
	}
	var remaining int
	if err = vault.db.QueryRow(`SELECT count(*) FROM private_memory WHERE id=?`, memory.ID).Scan(&remaining); err != nil || remaining != 0 {
		t.Fatalf("deleted memory rows=%d err=%v", remaining, err)
	}
}

func TestVaultFailsClosedOnUnknownOriginAndExpiredConsent(t *testing.T) {
	vault, err := Open(testVaultPath(t), bytes.Repeat([]byte{0x66}, 32))
	if err != nil {
		t.Fatal(err)
	}
	defer vault.Close()
	consent, err := vault.RecordOwnerConsent(context.Background(), "owner", []Scope{ScopeMemoryRead, ScopeMemoryWrite}, time.Now().Add(time.Hour), true, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = vault.db.Exec(`UPDATE private_consent SET expires_at_utc=? WHERE id=?`, time.Now().Add(-time.Second).UTC().Format(time.RFC3339Nano), consent.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = vault.StoreMemory(context.Background(), consent.ID, "owner", Origin("made_up"), "synthetic"); !errors.Is(err, ErrInvalidOrigin) {
		t.Fatalf("unknown origin error = %v", err)
	}
	if _, err = vault.StoreMemory(context.Background(), consent.ID, "owner", OriginUserAsserted, "synthetic"); !errors.Is(err, ErrConsentDenied) {
		t.Fatalf("expired consent error = %v", err)
	}
}

func TestVaultRequiresDedicatedPrivateDirectoryAndKey(t *testing.T) {
	dir := t.TempDir()
	if _, err := Open(filepath.Join(dir, "vault.sqlite"), []byte("short")); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("short key error = %v", err)
	}
}

func TestVaultConsentMustBeOwnerConfirmedAndCannotGrantUnknownScope(t *testing.T) {
	vault, err := Open(testVaultPath(t), bytes.Repeat([]byte{0x77}, 32))
	if err != nil {
		t.Fatal(err)
	}
	defer vault.Close()
	if _, err = vault.RecordOwnerConsent(context.Background(), "owner", []Scope{ScopeMemoryWrite}, time.Now().Add(time.Hour), false, true); !errors.Is(err, ErrInvalidConsent) {
		t.Fatalf("unconfirmed owner consent error = %v", err)
	}
	if _, err = vault.RecordOwnerConsent(context.Background(), "owner", []Scope{"intimacy.admin"}, time.Now().Add(time.Hour), true, true); !errors.Is(err, ErrInvalidConsent) {
		t.Fatalf("unknown scope error = %v", err)
	}
	if _, err = vault.RecordOwnerConsent(context.Background(), "third-party", []Scope{ScopeMemoryWrite}, time.Now().Add(time.Hour), true, true); !errors.Is(err, ErrInvalidConsent) {
		t.Fatalf("third-party consent error = %v", err)
	}
}

func TestVaultAssociatedDataPreventsMetadataTampering(t *testing.T) {
	vault, err := Open(testVaultPath(t), bytes.Repeat([]byte{0x88}, 32))
	if err != nil {
		t.Fatal(err)
	}
	defer vault.Close()
	consent, err := vault.RecordOwnerConsent(context.Background(), "owner", []Scope{ScopeMemoryRead, ScopeMemoryWrite}, time.Now().Add(time.Hour), true, true)
	if err != nil {
		t.Fatal(err)
	}
	memory, err := vault.StoreMemory(context.Background(), consent.ID, "owner", OriginUserAsserted, "synthetic")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = vault.db.Exec(`UPDATE private_memory SET origin_type=? WHERE id=?`, OriginInferred, memory.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = vault.ReadMemory(context.Background(), memory.ID, consent.ID, "owner"); !errors.Is(err, ErrMemoryNotFound) {
		t.Fatalf("tampered metadata read error = %v", err)
	}
}

func TestPreferenceStoreEncryptsDeclaredValuesAndRequiresPreferenceConsent(t *testing.T) {
	path := testVaultPath(t)
	vault, err := Open(path, bytes.Repeat([]byte{0x91}, 32))
	if err != nil {
		t.Fatal(err)
	}
	defer vault.Close()
	ctx := context.Background()
	memoryOnly, err := vault.RecordOwnerConsent(ctx, "owner", []Scope{ScopeMemoryRead, ScopeMemoryWrite}, time.Now().Add(time.Hour), true, true)
	if err != nil {
		t.Fatal(err)
	}
	preference, err := vault.AddPreference(ctx, memoryOnly.ID, PreferenceInput{Dimension: "visual_style", Value: "synthetic-private-value", Strength: .7, Timescale: "recurring", Confirmed: true})
	if !errors.Is(err, ErrConsentDenied) {
		t.Fatalf("preference write without preferences.write err=%v", err)
	}
	if preference.ID != "" {
		t.Fatalf("preference was written without permission: %+v", preference)
	}
	consent, err := vault.RecordOwnerConsent(ctx, "owner", []Scope{ScopePreferenceRead, ScopePreferenceWrite}, time.Now().Add(time.Hour), true, true)
	if err != nil {
		t.Fatal(err)
	}
	preference, err = vault.AddPreference(ctx, consent.ID, PreferenceInput{Dimension: "visual_style", Value: "synthetic-private-value", Strength: .7, Timescale: "recurring", Confirmed: true})
	if err != nil {
		t.Fatal(err)
	}
	if preference.SourceType != "declared" || preference.Confidence != 1 || preference.EvidenceCount != 1 || preference.ConsentScope != "local_only" {
		t.Fatalf("declared preference metadata=%+v", preference)
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		data, readErr := os.ReadFile(filepath.Join(filepath.Dir(path), entry.Name()))
		if readErr != nil {
			t.Fatal(readErr)
		}
		if bytes.Contains(data, []byte("synthetic-private-value")) {
			t.Fatalf("preference plaintext leaked through %s", entry.Name())
		}
	}
}

func TestPreferenceCorrectionOverridesValueAndPreservesDeclaredProvenance(t *testing.T) {
	vault, err := Open(testVaultPath(t), bytes.Repeat([]byte{0x92}, 32))
	if err != nil {
		t.Fatal(err)
	}
	defer vault.Close()
	ctx := context.Background()
	consent, err := vault.RecordOwnerConsent(ctx, "owner", []Scope{ScopePreferenceRead, ScopePreferenceWrite}, time.Now().Add(time.Hour), true, true)
	if err != nil {
		t.Fatal(err)
	}
	created, err := vault.AddPreference(ctx, consent.ID, PreferenceInput{Dimension: "visual_style", Value: "initial-synthetic", Strength: .4, Timescale: "occasional", Confirmed: true})
	if err != nil {
		t.Fatal(err)
	}
	corrected, err := vault.CorrectPreference(ctx, created.ID, consent.ID, "corrected-synthetic", .8, true, true)
	if err != nil {
		t.Fatal(err)
	}
	firstSeen, firstErr := time.Parse(time.RFC3339Nano, corrected.FirstSeenAt)
	lastSeen, lastErr := time.Parse(time.RFC3339Nano, corrected.LastSeenAt)
	if corrected.Value != "corrected-synthetic" || corrected.SourceType != "declared" || corrected.Confidence != 1 || corrected.Strength != .8 || corrected.EvidenceCount != 2 || corrected.NegativeFeedbackCount != 1 || firstErr != nil || lastErr != nil || !lastSeen.After(firstSeen) {
		t.Fatalf("corrected preference=%+v", corrected)
	}
}

func TestPreferenceConsentRevocationBlocksReadAndDeleteRemainsAvailable(t *testing.T) {
	vault, err := Open(testVaultPath(t), bytes.Repeat([]byte{0x93}, 32))
	if err != nil {
		t.Fatal(err)
	}
	defer vault.Close()
	ctx := context.Background()
	consent, err := vault.RecordOwnerConsent(ctx, "owner", []Scope{ScopePreferenceRead, ScopePreferenceWrite}, time.Now().Add(time.Hour), true, true)
	if err != nil {
		t.Fatal(err)
	}
	created, err := vault.AddPreference(ctx, consent.ID, PreferenceInput{Dimension: "format", Value: "synthetic-format", Strength: 1, Timescale: "persistent", Confirmed: true})
	if err != nil {
		t.Fatal(err)
	}
	if err = vault.RevokeConsent(ctx, consent.ID, "owner"); err != nil {
		t.Fatal(err)
	}
	if _, err = vault.ListPreferences(ctx, consent.ID); !errors.Is(err, ErrConsentDenied) {
		t.Fatalf("list after revoke err=%v", err)
	}
	if err = vault.DeletePreference(ctx, created.ID, "owner"); err != nil {
		t.Fatal(err)
	}
	var count int
	if err = vault.db.QueryRow(`SELECT count(*) FROM private_preference WHERE id=?`, created.ID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("deleted preference rows=%d err=%v", count, err)
	}
}

func TestPreferenceInputRejectsUnconfirmedAndInvalidValues(t *testing.T) {
	vault, err := Open(testVaultPath(t), bytes.Repeat([]byte{0x94}, 32))
	if err != nil {
		t.Fatal(err)
	}
	defer vault.Close()
	ctx := context.Background()
	consent, err := vault.RecordOwnerConsent(ctx, "owner", []Scope{ScopePreferenceWrite}, time.Now().Add(time.Hour), true, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range []PreferenceInput{
		{Dimension: "unknown", Value: "synthetic", Strength: .5, Timescale: "persistent", Confirmed: true},
		{Dimension: "visual_style", Value: "synthetic", Strength: 1.1, Timescale: "persistent", Confirmed: true},
		{Dimension: "visual_style", Value: "synthetic", Strength: .5, Timescale: "forever", Confirmed: true},
		{Dimension: "visual_style", Value: "synthetic", Strength: .5, Timescale: "persistent", Confirmed: false},
	} {
		if _, err := vault.AddPreference(ctx, consent.ID, input); !errors.Is(err, ErrInvalidPreference) {
			t.Fatalf("invalid preference %+v err=%v", input, err)
		}
	}
}

func testVaultPath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(dir, "vault.sqlite")
}
