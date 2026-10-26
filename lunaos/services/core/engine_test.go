package core

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
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

type blockingRunner struct {
	started chan struct{}
	resume  chan struct{}
}

func (r blockingRunner) Execute(ctx context.Context, action string) (ArtifactEvidence, error) {
	close(r.started)
	select {
	case <-r.resume:
		return (ChecksumRunner{}).Execute(ctx, action)
	case <-ctx.Done():
		return ArtifactEvidence{}, ctx.Err()
	}
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

func TestReprioritizeRequiresOwnerAndExplicitBoundedMetrics(t *testing.T) {
	e := testEngine(t, testDBPath(t), &testPolicy{})
	goal, err := e.Create(context.Background(), Request{Action: "local.checksum"}, "owner-secret", "reprio-validation")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.Reprioritize(context.Background(), goal.ID, Metrics{Impact: float64Ptr(.1), Urgency: float64Ptr(.1), UserValue: float64Ptr(.1), Confidence: float64Ptr(.1)}, "owner-secret"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("incomplete metrics error=%v, want ErrInvalid", err)
	}
	if _, err := e.Reprioritize(context.Background(), goal.ID, validMetrics(2), "owner-secret"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("out-of-range metrics error=%v, want ErrInvalid", err)
	}
	if _, err := e.Reprioritize(context.Background(), goal.ID, validMetrics(.4), "guest"); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("guest reprioritize error=%v, want ErrUnauthorized", err)
	}
	if _, err := e.Pause(context.Background(), goal.ID, "owner-secret"); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Reprioritize(context.Background(), goal.ID, validMetrics(.4), "owner-secret"); !errors.Is(err, ErrTransitionConflict) {
		t.Fatalf("paused reprioritize error=%v, want ErrTransitionConflict", err)
	}
	current, err := e.Get(context.Background(), goal.ID)
	if err != nil || current.Score == nil || *current.Score != 100 {
		t.Fatalf("rejected requests changed goal score: goal=%+v err=%v", current, err)
	}
}

func TestReprioritizeDenialBlocksWithoutScoringAndAllowedUpdatePersists(t *testing.T) {
	policy := &testPolicy{}
	e := testEngine(t, testDBPath(t), policy)
	goal, err := e.Create(context.Background(), Request{Action: "local.checksum"}, "owner-secret", "reprio-policy")
	if err != nil {
		t.Fatal(err)
	}
	policy.denyCode = "ACTION_DENIED"
	denied, err := e.Reprioritize(context.Background(), goal.ID, validMetrics(.2), "owner-secret")
	if err != nil || denied.Status != "blocked" || denied.Score != nil || denied.PolicyCode != "ACTION_DENIED" {
		t.Fatalf("denied reprioritize=%+v err=%v", denied, err)
	}
	var jobState string
	if err := e.db.QueryRow(`SELECT state FROM job WHERE goal_id=?`, goal.ID).Scan(&jobState); err != nil || jobState != "blocked" {
		t.Fatalf("denied reprioritize job state=%q err=%v", jobState, err)
	}
	policy.denyCode = ""
	updated, err := e.Reprioritize(context.Background(), goal.ID, validMetrics(.2), "owner-secret")
	if err != nil || updated.Status != "approved" || updated.Score == nil || *updated.Score != Score(validMetrics(.2)) {
		t.Fatalf("allowed reprioritize=%+v err=%v", updated, err)
	}
	if eventCount(t, e, goal.ID, "goal.reprioritized") != 1 {
		t.Fatal("successful reprioritize did not create one audit event")
	}
}

func validMetrics(value float64) Metrics {
	return Metrics{Impact: float64Ptr(value), Urgency: float64Ptr(value), UserValue: float64Ptr(value), Confidence: float64Ptr(value), Efficiency: float64Ptr(value)}
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

func TestGoalPauseRequiresOwner(t *testing.T) {
	e := testEngine(t, testDBPath(t), &testPolicy{})
	goal, err := e.Create(context.Background(), Request{Action: "local.checksum"}, "owner-secret", "pause-owner-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = e.Pause(context.Background(), goal.ID, "guest"); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("guest pause error = %v, want ErrUnauthorized", err)
	}
	current, err := e.Get(context.Background(), goal.ID)
	if err != nil || current.Status != "approved" {
		t.Fatalf("unauthorized pause changed goal: goal=%+v err=%v", current, err)
	}
}

func TestGoalPauseBlocksTickAndResumeAfterRestart(t *testing.T) {
	path := testDBPath(t)
	policy := &testPolicy{}
	e := testEngine(t, path, policy)
	goal, err := e.Create(context.Background(), Request{Action: "local.checksum"}, "owner-secret", "pause-resume-1")
	if err != nil {
		t.Fatal(err)
	}
	paused, err := e.Pause(context.Background(), goal.ID, "owner-secret")
	if err != nil || paused.Status != "paused" {
		t.Fatalf("pause goal=%+v err=%v", paused, err)
	}
	if result, err := e.Tick(context.Background()); err != nil || result.Outcome != "idle_success" {
		t.Fatalf("tick while paused result=%+v err=%v", result, err)
	}
	if count := eventCount(t, e, goal.ID, "goal.paused"); count != 1 {
		t.Fatalf("goal.paused events=%d, want 1", count)
	}
	if err := e.Close(); err != nil {
		t.Fatal(err)
	}
	reopened := testEngine(t, path, policy)
	current, err := reopened.Get(context.Background(), goal.ID)
	if err != nil || current.Status != "paused" {
		t.Fatalf("goal after restart=%+v err=%v", current, err)
	}
	resumed, err := reopened.Resume(context.Background(), goal.ID, "owner-secret")
	if err != nil || resumed.Status != "approved" {
		t.Fatalf("resume goal=%+v err=%v", resumed, err)
	}
	if result, err := reopened.Tick(context.Background()); err != nil || result.Outcome != "succeeded" {
		t.Fatalf("tick after resume result=%+v err=%v", result, err)
	}
}

func TestGoalCancelIsIdempotentAndNeverRuns(t *testing.T) {
	policy := &testPolicy{}
	e := testEngine(t, testDBPath(t), policy)
	goal, err := e.Create(context.Background(), Request{Action: "local.checksum"}, "owner-secret", "cancel-1")
	if err != nil {
		t.Fatal(err)
	}
	first, err := e.Cancel(context.Background(), goal.ID, "owner-secret")
	if err != nil || first.Status != "cancelled" {
		t.Fatalf("cancel goal=%+v err=%v", first, err)
	}
	second, err := e.Cancel(context.Background(), goal.ID, "owner-secret")
	if err != nil || second.Status != "cancelled" {
		t.Fatalf("repeat cancel goal=%+v err=%v", second, err)
	}
	if count := eventCount(t, e, goal.ID, "goal.cancelled"); count != 1 {
		t.Fatalf("goal.cancelled events=%d, want 1", count)
	}
	if result, err := e.Tick(context.Background()); err != nil || result.Outcome != "idle_success" {
		t.Fatalf("tick after cancel result=%+v err=%v", result, err)
	}
	if policy.evaluateCalls != 0 {
		t.Fatalf("cancelled goal reached policy evaluation %d times", policy.evaluateCalls)
	}
}

func TestGoalPauseRejectsRunningJob(t *testing.T) {
	path := testDBPath(t)
	policy := &testPolicy{}
	runner := blockingRunner{started: make(chan struct{}), resume: make(chan struct{})}
	e, err := Open(path, "owner-secret", policy, runner)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = e.Close() })
	goal, err := e.Create(context.Background(), Request{Action: "local.checksum"}, "owner-secret", "pause-running-1")
	if err != nil {
		t.Fatal(err)
	}
	tickResult := make(chan error, 1)
	go func() {
		_, tickErr := e.Tick(context.Background())
		tickResult <- tickErr
	}()
	select {
	case <-runner.started:
	case <-time.After(2 * time.Second):
		t.Fatal("runner did not start")
	}
	if _, err := e.Pause(context.Background(), goal.ID, "owner-secret"); !errors.Is(err, ErrTransitionConflict) {
		t.Fatalf("pause during execution error = %v, want ErrTransitionConflict", err)
	}
	close(runner.resume)
	select {
	case err := <-tickResult:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("running goal did not finish after release")
	}
}

func eventCount(t *testing.T, e *Engine, goalID, kind string) int {
	t.Helper()
	var count int
	if err := e.db.QueryRow(`SELECT count(*) FROM goal_event WHERE goal_id=? AND kind=?`, goalID, kind).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
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

func TestGoalPlanRequiresOwnerAndRejectsCycles(t *testing.T) {
	e := testEngine(t, testDBPath(t), &testPolicy{})
	a, err := e.Create(context.Background(), Request{Action: "local.checksum"}, "owner-secret", "plan-a")
	if err != nil {
		t.Fatal(err)
	}
	b, err := e.Create(context.Background(), Request{Action: "local.checksum"}, "owner-secret", "plan-b")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.SetPlan(context.Background(), a.ID, []string{b.ID}, nil, "guest"); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("guest plan error = %v, want ErrUnauthorized", err)
	}
	if _, err := e.SetPlan(context.Background(), a.ID, []string{b.ID, b.ID}, nil, "owner-secret"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("duplicate dependency error = %v, want ErrInvalid", err)
	}
	if _, err := e.SetPlan(context.Background(), a.ID, []string{a.ID}, nil, "owner-secret"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("self dependency error = %v, want ErrInvalid", err)
	}
	if _, err := e.SetPlan(context.Background(), a.ID, []string{b.ID}, nil, "owner-secret"); err != nil {
		t.Fatal(err)
	}
	if _, err := e.SetPlan(context.Background(), b.ID, []string{a.ID}, nil, "owner-secret"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("cycle error = %v, want ErrInvalid", err)
	}
	if _, err := e.SetPlan(context.Background(), a.ID, []string{}, &b.ID, "owner-secret"); err != nil {
		t.Fatal(err)
	}
	if _, err := e.SetPlan(context.Background(), b.ID, []string{a.ID}, nil, "owner-secret"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("combined hierarchy/dependency cycle error = %v, want ErrInvalid", err)
	}
	if _, err := e.SetPlan(context.Background(), a.ID, []string{"missing-goal"}, nil, "owner-secret"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing dependency error = %v, want ErrNotFound", err)
	}
}

func TestSchedulerWaitsForDependenciesAndPlanSurvivesRestart(t *testing.T) {
	path := testDBPath(t)
	policy := &testPolicy{}
	e := testEngine(t, path, policy)
	dependent, err := e.Create(context.Background(), Request{Action: "local.checksum"}, "owner-secret", "dependent-goal")
	if err != nil {
		t.Fatal(err)
	}
	prerequisite, err := e.Create(context.Background(), Request{Action: "local.checksum"}, "owner-secret", "prerequisite-goal")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.SetPlan(context.Background(), dependent.ID, []string{prerequisite.ID}, nil, "owner-secret"); err != nil {
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
	current, err := reopened.Get(context.Background(), dependent.ID)
	if err != nil || len(current.DependencyIDs) != 1 || current.DependencyIDs[0] != prerequisite.ID {
		t.Fatalf("plan after restart = %+v, err=%v", current.DependencyIDs, err)
	}
	first, err := reopened.Tick(context.Background())
	if err != nil || first.GoalID != prerequisite.ID || first.Outcome != "succeeded" {
		t.Fatalf("first scheduler tick = %+v, err=%v; want prerequisite", first, err)
	}
	second, err := reopened.Tick(context.Background())
	if err != nil || second.GoalID != dependent.ID || second.Outcome != "succeeded" {
		t.Fatalf("second scheduler tick = %+v, err=%v; want dependent", second, err)
	}
}

func TestGoalDependencyMigrationPreservesExistingGoal(t *testing.T) {
	path := testDBPath(t)
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(schemaSQL); err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO goal(id,action,privacy_class,status,metrics_json,score,policy_code,provenance_status,idempotency_key,request_hash,created_at_utc,updated_at_utc)
		VALUES('legacy-goal','local.checksum','general','approved','{}',50,'OWNER_SCOPED_GRANT','unknown','legacy-key','legacy-hash','2026-01-01T00:00:00Z','2026-01-01T00:00:00Z')`)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0600); err != nil {
		t.Fatal(err)
	}
	e, err := Open(path, "owner-secret", &testPolicy{}, ChecksumRunner{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = e.Close() })
	goal, err := e.Get(context.Background(), "legacy-goal")
	if err != nil || goal.ID != "legacy-goal" || goal.Status != "approved" || len(goal.DependencyIDs) != 0 {
		t.Fatalf("legacy goal after additive migration = %+v, err=%v", goal, err)
	}
	var version int
	if err := e.db.QueryRow(`SELECT max(version) FROM schema_migration`).Scan(&version); err != nil || version != 3 {
		t.Fatalf("migration version=%d err=%v; want 3", version, err)
	}
}

func float64Ptr(v float64) *float64 { return &v }
