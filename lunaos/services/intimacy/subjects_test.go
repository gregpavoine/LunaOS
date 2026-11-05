package intimacy

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"
)

func subjectVault(t *testing.T) (*Vault, string, []byte) {
	t.Helper()
	path := testVaultPath(t)
	key := bytes.Repeat([]byte{7}, 32)
	v, err := Open(path, key)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = v.Close() })
	return v, path, key
}

func intimateRequest(proof string, scopes ...SubjectScope) SubjectConsentRequest {
	return SubjectConsentRequest{
		Scopes:                 scopes,
		DisclosureRevision:     CurrentDisclosureRevision,
		IntimateDisclosureRead: true,
		Proof:                  proof,
		ExpiresAt:              time.Now().Add(24 * time.Hour),
	}
}

func TestNobodyConsentsForAnotherPerson(t *testing.T) {
	v, _, _ := subjectVault(t)
	ctx := context.Background()
	guest, err := v.CreateSubject(ctx, "OTHER", "attested_adult")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.GrantSubjectConsent(ctx, guest.ID, intimateRequest("owner", SubjectIntimateStillCapture)); err != ErrSubjectConsent {
		t.Fatalf("owner-declared consent for a guest: err=%v", err)
	}
	stale := intimateRequest(ProofSubjectClick, SubjectFaceRecognition)
	stale.DisclosureRevision = "2025-01-01.v0"
	if _, err := v.GrantSubjectConsent(ctx, guest.ID, stale); err != ErrInvalidSubject {
		t.Fatalf("consent against a stale disclosure: err=%v", err)
	}
	unread := intimateRequest(ProofSubjectClick, SubjectIntimateView)
	unread.IntimateDisclosureRead = false
	if _, err := v.GrantSubjectConsent(ctx, guest.ID, unread); err != ErrSubjectConsent {
		t.Fatalf("intimate scope without the nudity disclosure: err=%v", err)
	}
	if _, err := v.GrantSubjectConsent(ctx, guest.ID, intimateRequest(ProofSubjectClick, SubjectFaceRecognition)); err != nil {
		t.Fatal(err)
	}
	if _, err := v.StoreSubjectAsset(ctx, guest.ID, SubjectIntimateStillCapture, []byte("frame")); err != ErrSubjectConsent {
		t.Fatalf("face recognition consent must not allow intimate capture: err=%v", err)
	}
}

func TestMinorsAndUnverifiedSubjectsGetNoIntimateScope(t *testing.T) {
	v, _, _ := subjectVault(t)
	ctx := context.Background()
	for _, status := range []string{"unverified", "suspected_minor", "minor"} {
		subject, err := v.CreateSubject(ctx, "OTHER", status)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := v.GrantSubjectConsent(ctx, subject.ID, intimateRequest(ProofSubjectClick, SubjectIntimateView)); err != ErrSubjectNotAdult {
			t.Fatalf("%s subject granted an intimate scope: err=%v", status, err)
		}
	}
	adult, _ := v.CreateSubject(ctx, "OTHER", "attested_adult")
	if _, err := v.GrantSubjectConsent(ctx, adult.ID, intimateRequest(ProofSubjectClick, SubjectIntimateStillCapture, SubjectIntimateDataRetention)); err != nil {
		t.Fatal(err)
	}
	if _, err := v.StoreSubjectAsset(ctx, adult.ID, SubjectIntimateStillCapture, []byte("frame")); err != nil {
		t.Fatal(err)
	}
	receipt, err := v.FlagSuspectedMinor(ctx, adult.ID)
	if err != nil || receipt.AssetsErased != 1 {
		t.Fatalf("a suspected minor flag must erase intimate assets: %+v err=%v", receipt, err)
	}
	if _, err := v.StoreSubjectAsset(ctx, adult.ID, SubjectIntimateStillCapture, []byte("frame")); err != ErrSubjectNotAdult {
		t.Fatalf("intimate storage after a minor flag: err=%v", err)
	}
}

func TestVoiceConsentNeedsTheFreshSingleUseChallenge(t *testing.T) {
	v, _, _ := subjectVault(t)
	ctx := context.Background()
	guest, _ := v.CreateSubject(ctx, "OTHER", "attested_adult")
	if _, err := v.GrantSubjectConsent(ctx, guest.ID, intimateRequest(ProofVoiceChallenge, SubjectFaceReference)); err != ErrChallengeInvalid {
		t.Fatalf("voice consent without a challenge: err=%v", err)
	}
	phrase, err := v.IssueVoiceChallenge(ctx, guest.ID)
	if err != nil {
		t.Fatal(err)
	}
	replayed := intimateRequest(ProofVoiceChallenge, SubjectFaceReference)
	replayed.ChallengeResponse = "OK, je valide"
	if _, err := v.GrantSubjectConsent(ctx, guest.ID, replayed); err != ErrChallengeInvalid {
		t.Fatalf("a recorded bare phrase granted consent: err=%v", err)
	}
	answered := intimateRequest(ProofVoiceChallenge, SubjectFaceReference)
	answered.ChallengeResponse = phrase
	if _, err := v.GrantSubjectConsent(ctx, guest.ID, answered); err != ErrChallengeInvalid {
		t.Fatalf("a challenge must be consumed by the first answer, right or wrong: err=%v", err)
	}
	phrase, _ = v.IssueVoiceChallenge(ctx, guest.ID)
	answered.ChallengeResponse = "  ok je VALIDE " + phrase[len("OK, je valide "):]
	if _, err := v.GrantSubjectConsent(ctx, guest.ID, answered); err != nil {
		t.Fatalf("the spoken challenge, as heard, must grant consent: err=%v", err)
	}
	if _, err := v.GrantSubjectConsent(ctx, guest.ID, answered); err != ErrChallengeInvalid {
		t.Fatalf("a used challenge was replayed: err=%v", err)
	}
}

func TestRevocationErasesDependentAssetsAndDeletionDestroysTheKey(t *testing.T) {
	v, path, key := subjectVault(t)
	ctx := context.Background()
	guest, _ := v.CreateSubject(ctx, "OTHER", "attested_adult")
	if _, err := v.GrantSubjectConsent(ctx, guest.ID, intimateRequest(ProofSubjectClick,
		SubjectFaceReference, SubjectIntimateStillCapture, SubjectBodyProfile, SubjectIntimateDataRetention)); err != nil {
		t.Fatal(err)
	}
	face, err := v.StoreSubjectAsset(ctx, guest.ID, SubjectFaceReference, []byte("face-reference-canary"))
	if err != nil {
		t.Fatal(err)
	}
	for _, purpose := range []SubjectScope{SubjectIntimateStillCapture, SubjectBodyProfile} {
		if _, err := v.StoreSubjectAsset(ctx, guest.ID, purpose, []byte("intimate-canary")); err != nil {
			t.Fatal(err)
		}
	}
	if raw, _ := os.ReadFile(path); bytes.Contains(raw, []byte("intimate-canary")) || bytes.Contains(raw, []byte("face-reference-canary")) {
		t.Fatal("subject assets must not reach disk in plaintext")
	}
	receipt, err := v.RevokeSubjectScopes(ctx, guest.ID, []SubjectScope{SubjectIntimateDataRetention})
	if err != nil || receipt.AssetsErased != 2 {
		t.Fatalf("withdrawing retention must erase every intimate asset: %+v err=%v", receipt, err)
	}
	if _, err := v.StoreSubjectAsset(ctx, guest.ID, SubjectIntimateStillCapture, []byte("frame")); err != ErrSubjectConsent {
		t.Fatalf("intimate capture kept after retention was withdrawn: err=%v", err)
	}
	if asset, err := v.ReadSubjectAsset(ctx, guest.ID, face.ID); err != nil || string(asset.Data) != "face-reference-canary" {
		t.Fatalf("unrelated face reference must survive: %+v err=%v", asset, err)
	}

	if err := v.RotateKey(bytes.Repeat([]byte{9}, 32), true); err != nil {
		t.Fatal(err)
	}
	if asset, err := v.ReadSubjectAsset(ctx, guest.ID, face.ID); err != nil || string(asset.Data) != "face-reference-canary" {
		t.Fatalf("key rotation must rewrap subject keys: %+v err=%v", asset, err)
	}

	erased, err := v.DeleteSubject(ctx, guest.ID)
	if err != nil || !erased.KeyDestroyed || erased.AssetsErased != 1 {
		t.Fatalf("subject deletion receipt: %+v err=%v", erased, err)
	}
	var keys, consents int
	_ = v.db.QueryRow(`SELECT count(*) FROM subject_key`).Scan(&keys)
	_ = v.db.QueryRow(`SELECT count(*) FROM subject_consent`).Scan(&consents)
	if keys != 0 || consents != 0 {
		t.Fatalf("deletion left key=%d consent=%d rows", keys, consents)
	}
	if _, err := v.ActiveScopes(ctx, guest.ID); err != ErrSubjectNotFound {
		t.Fatalf("deleted subject still resolvable: err=%v", err)
	}
	if _, err := v.db.Exec(`DELETE FROM subject_audit`); err == nil {
		t.Fatal("subject audit must be append-only")
	}
	_ = key
}

func TestExpiredConsentDenies(t *testing.T) {
	v, _, _ := subjectVault(t)
	ctx := context.Background()
	guest, _ := v.CreateSubject(ctx, "OTHER", "attested_adult")
	if _, err := v.GrantSubjectConsent(ctx, guest.ID, intimateRequest(ProofSubjectClick, SubjectFaceReference)); err != nil {
		t.Fatal(err)
	}
	if _, err := v.db.Exec(`UPDATE subject_consent SET expires_at_utc=?`, time.Now().Add(-time.Minute).UTC().Format(time.RFC3339Nano)); err != nil {
		t.Fatal(err)
	}
	if _, err := v.StoreSubjectAsset(ctx, guest.ID, SubjectFaceReference, []byte("face")); err != ErrSubjectConsent {
		t.Fatalf("expired consent allowed storage: err=%v", err)
	}
	if scopes, _ := v.ActiveScopes(ctx, guest.ID); len(scopes) != 0 {
		t.Fatalf("expired consent still listed: %+v", scopes)
	}
}
