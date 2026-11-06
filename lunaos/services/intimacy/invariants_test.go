package intimacy

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
)

// Mont Blanc U1 invariants that the imported tests only pin in part. The
// others are covered by the imported tests:
//   - TestNobodyConsentsForAnotherPerson (owner cannot consent for a subject)
//   - TestMinorsAndUnverifiedSubjectsGetNoIntimateScope (unverified,
//     suspected_minor, minor get no intimate scope)
//   - TestVoiceConsentNeedsTheFreshSingleUseChallenge (a voice phrase without
//     an active challenge is refused)

func adultWithIntimateAssets(t *testing.T, v *Vault) (Subject, SubjectAsset, SubjectAsset) {
	t.Helper()
	ctx := context.Background()
	guest, err := v.CreateSubject(ctx, "OTHER", "attested_adult")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = v.GrantSubjectConsent(ctx, guest.ID, intimateRequest(ProofSubjectClick,
		SubjectFaceReference, SubjectIntimateStillCapture, SubjectIntimateDataRetention)); err != nil {
		t.Fatal(err)
	}
	face, err := v.StoreSubjectAsset(ctx, guest.ID, SubjectFaceReference, []byte("face-canary"))
	if err != nil {
		t.Fatal(err)
	}
	intimate, err := v.StoreSubjectAsset(ctx, guest.ID, SubjectIntimateStillCapture, []byte("intimate-canary"))
	if err != nil {
		t.Fatal(err)
	}
	return guest, face, intimate
}

func TestFlagSuspectedMinorErasesIntimateAssets(t *testing.T) {
	v, _, _ := subjectVault(t)
	ctx := context.Background()
	guest, face, intimate := adultWithIntimateAssets(t, v)

	if _, err := v.FlagSuspectedMinor(ctx, guest.ID); err != nil {
		t.Fatal(err)
	}
	var status string
	if err := v.db.QueryRow(`SELECT adult_status FROM private_subject WHERE id=?`, guest.ID).Scan(&status); err != nil || status != "suspected_minor" {
		t.Fatalf("status after flag: %q err=%v", status, err)
	}
	var rows int
	if err := v.db.QueryRow(`SELECT count(*) FROM subject_asset WHERE id=?`, intimate.ID).Scan(&rows); err != nil || rows != 0 {
		t.Fatalf("intimate asset row survived the flag: rows=%d err=%v", rows, err)
	}
	if _, err := v.ReadSubjectAsset(ctx, guest.ID, intimate.ID); err == nil {
		t.Fatal("intimate asset readable after the flag")
	}
	scopes, err := v.ActiveScopes(ctx, guest.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, grant := range scopes {
		if intimateScopes[grant.Scope] {
			t.Fatalf("intimate scope %s still active after the flag", grant.Scope)
		}
	}
	if asset, err := v.ReadSubjectAsset(ctx, guest.ID, face.ID); err != nil || string(asset.Data) != "face-canary" {
		t.Fatalf("non-intimate asset must survive the flag: %+v err=%v", asset, err)
	}
	// A flag only lowers status: a known minor is never raised to suspected.
	minor, _ := v.CreateSubject(ctx, "OTHER", "minor")
	if _, err := v.FlagSuspectedMinor(ctx, minor.ID); err != nil {
		t.Fatal(err)
	}
	if err := v.db.QueryRow(`SELECT adult_status FROM private_subject WHERE id=?`, minor.ID).Scan(&status); err != nil || status != "minor" {
		t.Fatalf("flag changed a minor to %q err=%v", status, err)
	}
}

func TestRevocationErasesDependentAssetsInSameTransaction(t *testing.T) {
	v, _, _ := subjectVault(t)
	ctx := context.Background()
	guest, _, intimate := adultWithIntimateAssets(t, v)

	// Make the closing audit insert fail: if erasure were committed on its own
	// the asset would be gone while the revocation failed.
	if _, err := v.db.Exec(`CREATE TRIGGER fail_revoke_audit BEFORE INSERT ON subject_audit WHEN NEW.event_code='CONSENT_REVOKED'
BEGIN SELECT RAISE(ABORT,'forced failure'); END`); err != nil {
		t.Fatal(err)
	}
	if _, err := v.RevokeSubjectScopes(ctx, guest.ID, []SubjectScope{SubjectIntimateDataRetention}); err == nil {
		t.Fatal("revocation must fail when its audit row cannot be written")
	}
	if asset, err := v.ReadSubjectAsset(ctx, guest.ID, intimate.ID); err != nil || string(asset.Data) != "intimate-canary" {
		t.Fatalf("failed revocation must leave assets and consent untouched: %+v err=%v", asset, err)
	}
	if _, err := v.db.Exec(`DROP TRIGGER fail_revoke_audit`); err != nil {
		t.Fatal(err)
	}

	receipt, err := v.RevokeSubjectScopes(ctx, guest.ID, []SubjectScope{SubjectIntimateDataRetention})
	if err != nil || receipt.AssetsErased != 1 {
		t.Fatalf("revocation receipt: %+v err=%v", receipt, err)
	}
	var live int
	if err := v.db.QueryRow(`SELECT count(*) FROM subject_consent WHERE subject_id=? AND scope=? AND revoked_at_utc IS NULL`,
		guest.ID, string(SubjectIntimateDataRetention)).Scan(&live); err != nil || live != 0 {
		t.Fatalf("consent still live after revocation: %d err=%v", live, err)
	}
	if _, err := v.ReadSubjectAsset(ctx, guest.ID, intimate.ID); err == nil {
		t.Fatal("asset readable after revocation")
	}
}

func TestSubjectDeletionDestroysKey(t *testing.T) {
	v, _, _ := subjectVault(t)
	ctx := context.Background()
	guest, face, _ := adultWithIntimateAssets(t, v)

	var wrapped, sealed []byte
	if err := v.db.QueryRow(`SELECT wrapped_key FROM subject_key WHERE subject_id=?`, guest.ID).Scan(&wrapped); err != nil {
		t.Fatal(err)
	}
	if err := v.db.QueryRow(`SELECT ciphertext FROM subject_asset WHERE id=?`, face.ID).Scan(&sealed); err != nil {
		t.Fatal(err)
	}
	if _, err := v.DeleteSubject(ctx, guest.ID); err != nil {
		t.Fatal(err)
	}
	var keys int
	if err := v.db.QueryRow(`SELECT count(*) FROM subject_key WHERE subject_id=?`, guest.ID).Scan(&keys); err != nil || keys != 0 {
		t.Fatalf("data key survived deletion: %d err=%v", keys, err)
	}

	// A stray copy of the ciphertext (WAL frame, older backup) must stay
	// unreadable: re-creating the same subject gets a fresh key, not the old one.
	stamp := utcNow()
	if _, err := v.db.Exec(`INSERT INTO private_subject(id,kind,adult_status,created_at_utc) VALUES(?,?,?,?)`, guest.ID, "OTHER", "attested_adult", stamp); err != nil {
		t.Fatal(err)
	}
	fresh := bytes.Repeat([]byte{3}, 32)
	freshWrapped, err := sealWithAEAD(v.aead, fresh, subjectKeyAAD(guest.ID))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(freshWrapped, wrapped) {
		t.Fatal("test setup reused the old wrapped key")
	}
	if _, err := v.db.Exec(`INSERT INTO subject_key(subject_id,wrapped_key) VALUES(?,?)`, guest.ID, freshWrapped); err != nil {
		t.Fatal(err)
	}
	if _, err := v.db.Exec(`INSERT INTO subject_asset(id,subject_id,purpose,ciphertext,created_at_utc) VALUES(?,?,?,?,?)`,
		face.ID, guest.ID, string(SubjectFaceReference), sealed, stamp); err != nil {
		t.Fatal(err)
	}
	if _, err := v.GrantSubjectConsent(ctx, guest.ID, intimateRequest(ProofSubjectClick, SubjectFaceReference)); err != nil {
		t.Fatal(err)
	}
	if asset, err := v.ReadSubjectAsset(ctx, guest.ID, face.ID); err == nil {
		t.Fatalf("old ciphertext opened without the destroyed key: %q", asset.Data)
	}
}

func TestAuditHoldsNoIntimatePayload(t *testing.T) {
	v, _, _ := subjectVault(t)
	ctx := context.Background()
	guest, _, _ := adultWithIntimateAssets(t, v)
	if _, err := v.RevokeSubjectScopes(ctx, guest.ID, []SubjectScope{SubjectIntimateDataRetention}); err != nil {
		t.Fatal(err)
	}
	if _, err := v.FlagSuspectedMinor(ctx, guest.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := v.DeleteSubject(ctx, guest.ID); err != nil {
		t.Fatal(err)
	}

	names, err := v.db.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name LIKE '%\_audit' ESCAPE '\'`)
	if err != nil {
		t.Fatal(err)
	}
	var tables []string
	for names.Next() {
		var name string
		if err := names.Scan(&name); err != nil {
			t.Fatal(err)
		}
		tables = append(tables, name)
	}
	names.Close()
	if len(tables) == 0 {
		t.Fatal("no audit table found")
	}
	for _, table := range tables {
		rows, err := v.db.Query("SELECT * FROM " + table)
		if err != nil {
			t.Fatal(err)
		}
		cols, _ := rows.Columns()
		for rows.Next() {
			values := make([]any, len(cols))
			ptrs := make([]any, len(cols))
			for i := range values {
				ptrs[i] = &values[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				t.Fatal(err)
			}
			for _, value := range values {
				if strings.Contains(fmt.Sprint(value), "canary") {
					t.Fatalf("audit table %s holds payload bytes", table)
				}
				if raw, ok := value.([]byte); ok && len(raw) > 64 {
					t.Fatalf("audit table %s holds a %d byte blob", table, len(raw))
				}
			}
		}
		rows.Close()
	}
}
