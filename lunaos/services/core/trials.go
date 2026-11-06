package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Self-evaluation after failure. The evaluator already refuses to complete a
// goal without verified evidence; this is what happens next. The owner (or an
// agent acting through the owner's request) retries a failed goal as an
// explicit, numbered trial that names its strategy and hypothesis and starts
// from the failure it answers. The Policy Broker decides again before the goal
// is scored or queued, the scheduler revalidates before running it, and the
// trial's outcome is filled in from the same verified result as the goal's.
// Five trials per goal is the ceiling, so replanning cannot loop or bypass a
// limit.

const maxStrategyTrials = 5

var (
	ErrTrialLimit   = errors.New("STRATEGY_TRIAL_LIMIT")
	strategyPattern = regexp.MustCompile(`^[a-z][a-z0-9_.-]{0,63}$`)
)

type StrategyTrial struct {
	Attempt       int    `json:"attempt"`
	Strategy      string `json:"strategy"`
	Hypothesis    string `json:"hypothesis"`
	BaselineCode  string `json:"baseline_code"`
	Outcome       string `json:"outcome"`
	OutcomeCode   string `json:"outcome_code,omitempty"`
	CreatedAtUTC  string `json:"created_at_utc"`
	FinishedAtUTC string `json:"finished_at_utc,omitempty"`
}

// Retry starts trial N+1 for a failed goal.
func (e *Engine) Retry(ctx context.Context, goalID, ownerToken, strategy, hypothesis string) (Goal, StrategyTrial, error) {
	if !e.owner(ownerToken) {
		return Goal{}, StrategyTrial{}, ErrUnauthorized
	}
	hypothesis = strings.TrimSpace(hypothesis)
	if !strategyPattern.MatchString(strategy) || hypothesis == "" || len(hypothesis) > 1024 {
		return Goal{}, StrategyTrial{}, fmt.Errorf("%w: STRATEGY_AND_HYPOTHESIS_REQUIRED", ErrInvalid)
	}
	goal, err := e.Get(ctx, goalID)
	if err != nil {
		return Goal{}, StrategyTrial{}, err
	}
	if goal.Status != "failed" {
		return Goal{}, StrategyTrial{}, ErrTransitionConflict
	}
	req, err := normalize(Request{Action: goal.Action, PrivacyClass: goal.PrivacyClass, Metrics: goal.Metrics, Assertion: goal.Assertion})
	if err != nil {
		return Goal{}, StrategyTrial{}, err
	}
	decision, err := e.policy.Authorize(ctx, goalID, req)
	if err != nil {
		return Goal{}, StrategyTrial{}, fmt.Errorf("policy authorization unavailable: %w", err)
	}
	if decision.Code == "" || (decision.Allowed && decision.Grant == "") {
		return Goal{}, StrategyTrial{}, fmt.Errorf("policy returned an incomplete decision")
	}

	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return Goal{}, StrategyTrial{}, err
	}
	defer tx.Rollback()
	var status, jobState, baseline string
	err = tx.QueryRowContext(ctx, `SELECT COALESCE(c.state,g.status),j.state,g.policy_code FROM goal g JOIN job j ON j.goal_id=g.id
		LEFT JOIN goal_control c ON c.goal_id=g.id WHERE g.id=?`, goalID).Scan(&status, &jobState, &baseline)
	if errors.Is(err, sql.ErrNoRows) {
		return Goal{}, StrategyTrial{}, ErrNotFound
	}
	if err != nil {
		return Goal{}, StrategyTrial{}, err
	}
	if status != "failed" || jobState != "failed" {
		return Goal{}, StrategyTrial{}, ErrTransitionConflict
	}
	var trials int
	if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM strategy_trial WHERE goal_id=?`, goalID).Scan(&trials); err != nil {
		return Goal{}, StrategyTrial{}, err
	}
	if trials >= maxStrategyTrials {
		return Goal{}, StrategyTrial{}, ErrTrialLimit
	}
	now := timestamp()
	trial := StrategyTrial{Attempt: trials + 1, Strategy: strategy, Hypothesis: hypothesis, BaselineCode: baseline, Outcome: "pending", CreatedAtUTC: now}
	if !decision.Allowed {
		trial.Outcome, trial.OutcomeCode, trial.FinishedAtUTC = "blocked", decision.Code, now
	}
	var finished any
	if trial.FinishedAtUTC != "" {
		finished = trial.FinishedAtUTC
	}
	var outcomeCode any
	if trial.OutcomeCode != "" {
		outcomeCode = trial.OutcomeCode
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO strategy_trial(goal_id,attempt,strategy,hypothesis,baseline_code,outcome,outcome_code,created_at_utc,finished_at_utc)
		VALUES(?,?,?,?,?,?,?,?,?)`, goalID, trial.Attempt, strategy, hypothesis, baseline, trial.Outcome, outcomeCode, now, finished); err != nil {
		return Goal{}, StrategyTrial{}, err
	}
	if !decision.Allowed {
		if _, err = tx.ExecContext(ctx, `UPDATE goal SET status='blocked',score=NULL,grant_token=NULL,policy_code=?,updated_at_utc=? WHERE id=?`, decision.Code, now, goalID); err != nil {
			return Goal{}, StrategyTrial{}, err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE job SET state='blocked' WHERE goal_id=?`, goalID); err != nil {
			return Goal{}, StrategyTrial{}, err
		}
		if err = insertEvent(ctx, tx, goalID, "policy.denied", decision.Code, now); err != nil {
			return Goal{}, StrategyTrial{}, err
		}
	} else {
		metricsJSON, err := json.Marshal(req.Metrics)
		if err != nil {
			return Goal{}, StrategyTrial{}, err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE goal SET status='approved',metrics_json=?,score=?,grant_token=?,policy_code=?,updated_at_utc=? WHERE id=?`,
			string(metricsJSON), Score(req.Metrics), decision.Grant, decision.Code, now, goalID); err != nil {
			return Goal{}, StrategyTrial{}, err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE job SET state='queued',attempt_count=0,lease_owner=NULL,lease_token=NULL,lease_expires_at_unix=NULL WHERE goal_id=?`, goalID); err != nil {
			return Goal{}, StrategyTrial{}, err
		}
		if err = insertEvent(ctx, tx, goalID, "policy.allowed", decision.Code, now); err != nil {
			return Goal{}, StrategyTrial{}, err
		}
		if err = insertEvent(ctx, tx, goalID, "strategy.trial_started", fmt.Sprintf("TRIAL_%d", trial.Attempt), now); err != nil {
			return Goal{}, StrategyTrial{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return Goal{}, StrategyTrial{}, err
	}
	goal, err = e.Get(ctx, goalID)
	return goal, trial, err
}

// Trials lists a goal's strategy trials, oldest first.
func (e *Engine) Trials(ctx context.Context, goalID string) ([]StrategyTrial, error) {
	if _, err := e.Get(ctx, goalID); err != nil {
		return nil, err
	}
	rows, err := e.db.QueryContext(ctx, `SELECT attempt,strategy,hypothesis,baseline_code,outcome,COALESCE(outcome_code,''),created_at_utc,COALESCE(finished_at_utc,'')
		FROM strategy_trial WHERE goal_id=? ORDER BY attempt`, goalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	trials := []StrategyTrial{}
	for rows.Next() {
		var trial StrategyTrial
		if err := rows.Scan(&trial.Attempt, &trial.Strategy, &trial.Hypothesis, &trial.BaselineCode, &trial.Outcome, &trial.OutcomeCode, &trial.CreatedAtUTC, &trial.FinishedAtUTC); err != nil {
			return nil, err
		}
		trials = append(trials, trial)
	}
	return trials, rows.Err()
}

// finishTrial records the measured outcome of the goal's pending trial, if any,
// inside the transaction that records the goal's own outcome.
func finishTrial(ctx context.Context, tx *sql.Tx, goalID, outcome, code, now string) error {
	_, err := tx.ExecContext(ctx, `UPDATE strategy_trial SET outcome=?,outcome_code=?,finished_at_utc=? WHERE goal_id=? AND outcome='pending'`, outcome, code, now, goalID)
	return err
}
