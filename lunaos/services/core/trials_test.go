package core

import (
	"context"
	"errors"
	"testing"
)

// failingRunner fails the first `failures` executions, then behaves like the
// real checksum runner, so the second strategy can genuinely succeed.
type failingRunner struct{ failures int }

func (r *failingRunner) Execute(ctx context.Context, action string) (ArtifactEvidence, error) {
	if r.failures > 0 {
		r.failures--
		return ArtifactEvidence{}, errors.New("simulated tool failure")
	}
	return ChecksumRunner{}.Execute(ctx, action)
}

func trialEngine(t *testing.T, policy *testPolicy, failures int) *Engine {
	t.Helper()
	e, err := Open(testDBPath(t), "owner-secret", policy, &failingRunner{failures: failures})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = e.Close() })
	return e
}

func failedGoal(t *testing.T, e *Engine, key string) Goal {
	t.Helper()
	ctx := context.Background()
	goal, err := e.Create(ctx, Request{Action: "local.checksum"}, "owner-secret", key)
	if err != nil {
		t.Fatal(err)
	}
	if result, err := e.Tick(ctx); err != nil || result.Outcome != "failed" {
		t.Fatalf("first attempt should fail: %+v err=%v", result, err)
	}
	return goal
}

func TestRetryIsAnExplicitPolicyCheckedTrialWithMeasuredOutcome(t *testing.T) {
	policy := &testPolicy{}
	e := trialEngine(t, policy, 1)
	ctx := context.Background()
	goal := failedGoal(t, e, "trial-success")

	if _, _, err := e.Retry(ctx, goal.ID, "intruder", "retry.same", "try again"); err != ErrUnauthorized {
		t.Fatalf("non-owner retry: err=%v", err)
	}
	if _, _, err := e.Retry(ctx, goal.ID, "owner-secret", "Bad Strategy!", "x"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("malformed strategy accepted: err=%v", err)
	}
	if _, _, err := e.Retry(ctx, goal.ID, "owner-secret", "retry.same", "   "); !errors.Is(err, ErrInvalid) {
		t.Fatalf("empty hypothesis accepted: err=%v", err)
	}

	authorizations := policy.grantCalls
	retried, trial, err := e.Retry(ctx, goal.ID, "owner-secret", "retry.backoff", "the tool failure was transient")
	if err != nil {
		t.Fatal(err)
	}
	if policy.grantCalls != authorizations+1 {
		t.Fatal("a retry must go back through the Policy Broker before it is scored")
	}
	if retried.Status != "approved" || retried.Score == nil || trial.Attempt != 1 || trial.Outcome != "pending" || trial.BaselineCode != "ACTION_FAILED" {
		t.Fatalf("retry did not start trial 1 from the failure: goal=%+v trial=%+v", retried, trial)
	}
	if result, err := e.Tick(ctx); err != nil || result.Outcome != "succeeded" {
		t.Fatalf("second strategy should succeed: %+v err=%v", result, err)
	}
	trials, err := e.Trials(ctx, goal.ID)
	if err != nil || len(trials) != 1 || trials[0].Outcome != "succeeded" || trials[0].OutcomeCode != "ARTIFACT_VERIFIED" || trials[0].FinishedAtUTC == "" {
		t.Fatalf("trial outcome must come from the verified result: %+v err=%v", trials, err)
	}
	completed, _ := e.Get(ctx, goal.ID)
	if completed.Evidence == nil || !completed.Evidence.Verified {
		t.Fatalf("completed trial without verified evidence: %+v", completed)
	}
	if _, _, err := e.Retry(ctx, goal.ID, "owner-secret", "retry.again", "already done"); err != ErrTransitionConflict {
		t.Fatalf("a completed goal must not be retried: err=%v", err)
	}
}

func TestRetriesStopAtTheTrialLimit(t *testing.T) {
	e := trialEngine(t, &testPolicy{}, 100)
	ctx := context.Background()
	goal := failedGoal(t, e, "trial-limit")
	for attempt := 1; attempt <= maxStrategyTrials; attempt++ {
		if _, _, err := e.Retry(ctx, goal.ID, "owner-secret", "retry.again", "maybe now"); err != nil {
			t.Fatalf("trial %d: %v", attempt, err)
		}
		if result, err := e.Tick(ctx); err != nil || result.Outcome != "failed" {
			t.Fatalf("trial %d should fail: %+v err=%v", attempt, result, err)
		}
	}
	if _, _, err := e.Retry(ctx, goal.ID, "owner-secret", "retry.again", "one more"); err != ErrTrialLimit {
		t.Fatalf("retry beyond the limit: err=%v", err)
	}
	trials, _ := e.Trials(ctx, goal.ID)
	for _, trial := range trials {
		if trial.Outcome != "failed" || trial.OutcomeCode != "ACTION_FAILED" {
			t.Fatalf("every exhausted trial must record its failure: %+v", trials)
		}
	}
}

func TestDeniedRetryBlocksWithoutScore(t *testing.T) {
	policy := &testPolicy{}
	e := trialEngine(t, policy, 1)
	ctx := context.Background()
	goal := failedGoal(t, e, "trial-denied")
	policy.denyCode = "OWNER_GRANT_REVOKED"
	blocked, trial, err := e.Retry(ctx, goal.ID, "owner-secret", "retry.same", "try again")
	if err != nil {
		t.Fatal(err)
	}
	if blocked.Status != "blocked" || blocked.Score != nil || trial.Outcome != "blocked" || trial.OutcomeCode != "OWNER_GRANT_REVOKED" {
		t.Fatalf("denied retry must block before scoring: goal=%+v trial=%+v", blocked, trial)
	}
	if result, err := e.Tick(ctx); err != nil || result.Outcome == "succeeded" || result.Outcome == "failed" {
		t.Fatalf("a denied retry must not run: %+v err=%v", result, err)
	}
}
