package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type testPolicy struct {
	grantCalls    int
	evaluateCalls int
	denyCode      string
}

func (p *testPolicy) Authorize(_ context.Context, _ string, req Request) (Decision, error) {
	p.grantCalls++
	if p.denyCode != "" {
		return Decision{Code: p.denyCode}, nil
	}
	if len(req.Action) >= 8 && req.Action[:8] == "payment." {
		return Decision{Code: "BUDGET_UNSET"}, nil
	}
	if req.Action != "local.checksum" || req.PrivacyClass != "general" {
		return Decision{Code: "ACTION_DENIED"}, nil
	}
	return Decision{Allowed: true, Code: "OWNER_SCOPED_GRANT", Grant: "test-scoped-grant"}, nil
}

func (p *testPolicy) Evaluate(_ context.Context, _ string, _ Request, _ string) (Decision, error) {
	p.evaluateCalls++
	return Decision{Allowed: true, Code: "SCOPED_ACTION_ALLOWED"}, nil
}

func testEngine(t *testing.T, path string, policy *testPolicy) *Engine {
	t.Helper()
	e, err := Open(path, "owner-secret", policy, ChecksumRunner{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = e.Close() })
	return e
}

func testDBPath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(dir, "goals.sqlite")
}

func TestCreateRequiresOwner(t *testing.T) {
	e := testEngine(t, testDBPath(t), &testPolicy{})
	_, err := e.Create(context.Background(), Request{Action: "local.checksum"}, "guest", "owner-create")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("guest create error = %v, want ErrUnauthorized", err)
	}
}

func TestPaymentDeniedWhenBudgetAbsent(t *testing.T) {
	policy := &testPolicy{}
	e := testEngine(t, testDBPath(t), policy)
	goal, err := e.Create(context.Background(), Request{Action: "payment.send"}, "owner-secret", "pay-1")
	if err != nil {
		t.Fatal(err)
	}
	if goal.Status != "blocked" || goal.PolicyCode != "BUDGET_UNSET" || goal.Score != nil {
		t.Fatalf("payment goal = %+v", goal)
	}
}

func TestPolicyDenialHappensBeforeScore(t *testing.T) {
	policy := &testPolicy{denyCode: "ACTION_DENIED"}
	e := testEngine(t, testDBPath(t), policy)
	goal, err := e.Create(context.Background(), Request{Action: "local.checksum"}, "owner-secret", "deny-1")
	if err != nil {
		t.Fatal(err)
	}
	if goal.Score != nil || goal.Status != "blocked" || policy.grantCalls != 1 {
		t.Fatalf("denied goal=%+v grant calls=%d", goal, policy.grantCalls)
	}
}

func TestDuplicateIdempotencyReturnsOneGoal(t *testing.T) {
	e := testEngine(t, testDBPath(t), &testPolicy{})
	body := Request{Action: "local.checksum"}
	first, err := e.Create(context.Background(), body, "owner-secret", "same-key")
	if err != nil {
		t.Fatal(err)
	}
	second, err := e.Create(context.Background(), body, "owner-secret", "same-key")
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID {
		t.Fatalf("duplicate returned IDs %q and %q", first.ID, second.ID)
	}
	if _, err := e.Create(context.Background(), Request{Action: "local.checksum", Metrics: Metrics{Impact: float64Ptr(0)}}, "owner-secret", "same-key"); !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("changed request error = %v, want idempotency conflict", err)
	}
}

func TestUnknownSourceRemainsHypothesis(t *testing.T) {
	e := testEngine(t, testDBPath(t), &testPolicy{})
	source := "unverified-source"
	goal, err := e.Create(context.Background(), Request{
		Action:    "local.checksum",
		Assertion: &Assertion{Claim: "the checksum fixture exists", Source: &source, Confidence: 1},
	}, "owner-secret", "source-1")
	if err != nil {
		t.Fatal(err)
	}
	if goal.Assertion == nil || goal.Assertion.ProvenanceStatus != "unknown" || goal.Assertion.EpistemicStatus != "hypothesis" {
		t.Fatalf("unknown source assertion = %+v", goal.Assertion)
	}
}

func TestQueuedGoalResumesAfterEngineRestart(t *testing.T) {
	path := testDBPath(t)
	policy := &testPolicy{}
	e := testEngine(t, path, policy)
	created, err := e.Create(context.Background(), Request{Action: "local.checksum"}, "owner-secret", "resume-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := e.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(path, "owner-secret", policy, ChecksumRunner{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = reopened.Close() })
	result, err := reopened.Tick(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	goal, err := reopened.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if result.Outcome != "succeeded" || goal.Status != "completed" || goal.Evidence == nil || !goal.Evidence.Verified {
		t.Fatalf("result=%+v goal=%+v", result, goal)
	}
}

func TestInactiveQueueIsIdleSuccess(t *testing.T) {
	e := testEngine(t, testDBPath(t), &testPolicy{})
	result, err := e.Tick(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Outcome != "idle_success" {
		t.Fatalf("idle result = %+v", result)
	}
}

func float64Ptr(v float64) *float64 { return &v }
