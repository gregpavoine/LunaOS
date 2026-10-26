package core

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	_ "github.com/mattn/go-sqlite3"
)

const (
	publicFixture       = "LunaOS public checkpoint v1\n"
	publicFixtureSHA256 = "07be6a1cc7b1b323fd3baca70abf34bfea13e09f9d28cf417d11aca77d4220ce"
	leaseDuration       = 30 * time.Second
	maxAttempts         = 3
)

var (
	actionPattern      = regexp.MustCompile(`^[a-z][a-z0-9_.]{0,79}$`)
	idempotencyPattern = regexp.MustCompile(`^[A-Za-z0-9_.:-]{1,128}$`)
)

type Engine struct {
	db         *sql.DB
	ownerToken string
	policy     Policy
	runner     Runner
}

func Open(path, ownerToken string, policy Policy, runner Runner) (*Engine, error) {
	if path == "" || ownerToken == "" || policy == nil || runner == nil {
		return nil, fmt.Errorf("database, owner credential, policy and runner are required")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if err := requirePrivateDirectory(filepath.Dir(absolute)); err != nil {
		return nil, err
	}
	if err := createPrivateDatabase(absolute); err != nil {
		return nil, err
	}
	uri := (&url.URL{Scheme: "file", Path: filepath.ToSlash(absolute)}).String()
	dsn := uri + "?_busy_timeout=5000&_journal_mode=WAL&_synchronous=FULL&_foreign_keys=on&_txlock=immediate"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := applyMigration(context.Background(), db, 3, goalDependencyMigration); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := applyMigration(context.Background(), db, 4, goalMetadataMigration); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Engine{db: db, ownerToken: ownerToken, policy: policy, runner: runner}, nil
}

func applyMigration(ctx context.Context, db *sql.DB, version int, migration string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var applied int
	if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM schema_migration WHERE version=?`, version).Scan(&applied); err != nil {
		return err
	}
	if applied != 0 {
		return tx.Commit()
	}
	if _, err = tx.ExecContext(ctx, migration); err != nil {
		return fmt.Errorf("apply database migration %d: %w", version, err)
	}
	return tx.Commit()
}

func (e *Engine) Close() error { return e.db.Close() }

func (e *Engine) StopIntimacy(ctx context.Context) error {
	_, err := e.db.ExecContext(ctx, `UPDATE intimacy_control SET stopped=1,updated_at_utc=? WHERE singleton=1`, timestamp())
	return err
}

func (e *Engine) IntimacyState(ctx context.Context) (IntimacyState, error) {
	var state IntimacyState
	var stopped int
	err := e.db.QueryRowContext(ctx, `SELECT stopped,privacy_mode,updated_at_utc FROM intimacy_control WHERE singleton=1`).Scan(&stopped, &state.PrivacyMode, &state.UpdatedAtUTC)
	if err != nil {
		return IntimacyState{}, err
	}
	state.Stopped = stopped == 1
	// No inference backend is configured in this checkout. Never claim capability from a contract alone.
	state.BackendAvailable = false
	return state, nil
}

func (e *Engine) Create(ctx context.Context, input Request, ownerToken, idempotencyKey string) (Goal, error) {
	if !hmac.Equal([]byte(ownerToken), []byte(e.ownerToken)) {
		return Goal{}, ErrUnauthorized
	}
	if !idempotencyPattern.MatchString(idempotencyKey) {
		return Goal{}, fmt.Errorf("%w: IDEMPOTENCY_KEY_REQUIRED", ErrInvalid)
	}
	req, err := normalize(input)
	if err != nil {
		return Goal{}, err
	}
	canonical, err := json.Marshal(req)
	if err != nil {
		return Goal{}, err
	}
	requestHash := sha256.Sum256(canonical)
	hashText := hex.EncodeToString(requestHash[:])
	if existing, oldHash, err := e.findIdempotent(ctx, idempotencyKey); err == nil {
		if oldHash != hashText {
			return Goal{}, ErrIdempotencyConflict
		}
		return existing, nil
	} else if !errors.Is(err, ErrNotFound) {
		return Goal{}, err
	}
	goalID, err := newID()
	if err != nil {
		return Goal{}, err
	}
	decision, err := e.policy.Authorize(ctx, goalID, req)
	if err != nil {
		return Goal{}, fmt.Errorf("policy authorization unavailable: %w", err)
	}
	if decision.Code == "" {
		return Goal{}, fmt.Errorf("policy returned an empty decision")
	}
	var scoreValue *float64
	status := "blocked"
	grant := ""
	if decision.Allowed {
		if decision.Grant == "" {
			return Goal{}, fmt.Errorf("policy allowed action without a scoped grant")
		}
		value := Score(req.Metrics)
		scoreValue = &value
		status = "approved"
		grant = decision.Grant
	}
	metricsJSON, err := json.Marshal(req.Metrics)
	if err != nil {
		return Goal{}, err
	}
	var assertionJSON any
	if req.Assertion != nil {
		assertionJSON, err = json.Marshal(req.Assertion)
		if err != nil {
			return Goal{}, err
		}
	}
	now := timestamp()
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return Goal{}, err
	}
	defer tx.Rollback()
	title := req.Title
	if title == "" {
		title = req.Action
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO goal
		(id,action,privacy_class,status,metrics_json,score,policy_code,provenance_status,assertion_json,idempotency_key,request_hash,grant_token,created_at_utc,updated_at_utc,title,description)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, goalID, req.Action, req.PrivacyClass, status, string(metricsJSON), scoreValue,
		decision.Code, provenanceFor(req.Assertion), nullableJSON(assertionJSON), idempotencyKey, hashText, nullableString(grant), now, now, title, req.Description)
	if err != nil {
		if existing, oldHash, lookupErr := e.findIdempotent(ctx, idempotencyKey); lookupErr == nil {
			if oldHash != hashText {
				return Goal{}, ErrIdempotencyConflict
			}
			return existing, nil
		}
		return Goal{}, err
	}
	jobState := "blocked"
	if decision.Allowed {
		jobState = "queued"
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO job(goal_id,state,max_attempts) VALUES(?,?,?)`, goalID, jobState, maxAttempts); err != nil {
		return Goal{}, err
	}
	if err = insertEvent(ctx, tx, goalID, "goal.created", "OWNER_REQUEST", now); err != nil {
		return Goal{}, err
	}
	policyKind := "policy.denied"
	if decision.Allowed {
		policyKind = "policy.allowed"
	}
	if err = insertEvent(ctx, tx, goalID, policyKind, decision.Code, now); err != nil {
		return Goal{}, err
	}
	if decision.Allowed {
		if err = insertEvent(ctx, tx, goalID, "goal.scored", "WEIGHTED_V1", now); err != nil {
			return Goal{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return Goal{}, err
	}
	return e.Get(ctx, goalID)
}

func (e *Engine) Get(ctx context.Context, goalID string) (Goal, error) {
	var g Goal
	var metricsJSON string
	var assertionJSON sql.NullString
	var parentID sql.NullString
	var scoreNumber sql.NullFloat64
	err := e.db.QueryRowContext(ctx, `SELECT g.id,g.title,g.description,g.action,g.privacy_class,COALESCE(c.state,g.status),g.metrics_json,g.score,g.policy_code,g.provenance_status,g.assertion_json,g.parent_id,g.created_at_utc,COALESCE(c.updated_at_utc,g.updated_at_utc)
		FROM goal g LEFT JOIN goal_control c ON c.goal_id=g.id WHERE g.id=?`, goalID).Scan(&g.ID, &g.Title, &g.Description, &g.Action, &g.PrivacyClass, &g.Status, &metricsJSON, &scoreNumber, &g.PolicyCode, &g.ProvenanceStatus, &assertionJSON, &parentID, &g.CreatedAtUTC, &g.UpdatedAtUTC)
	if errors.Is(err, sql.ErrNoRows) {
		return Goal{}, ErrNotFound
	}
	if err != nil {
		return Goal{}, err
	}
	if parentID.Valid {
		value := parentID.String
		g.ParentID = &value
	}
	g.DependencyIDs = []string{}
	rows, err := e.db.QueryContext(ctx, `SELECT depends_on_goal_id FROM goal_dependency WHERE goal_id=? ORDER BY depends_on_goal_id`, goalID)
	if err != nil {
		return Goal{}, err
	}
	for rows.Next() {
		var dependencyID string
		if err := rows.Scan(&dependencyID); err != nil {
			rows.Close()
			return Goal{}, err
		}
		g.DependencyIDs = append(g.DependencyIDs, dependencyID)
	}
	if err := rows.Close(); err != nil {
		return Goal{}, err
	}
	if err = json.Unmarshal([]byte(metricsJSON), &g.Metrics); err != nil {
		return Goal{}, err
	}
	if scoreNumber.Valid {
		value := scoreNumber.Float64
		g.Score = &value
	}
	if assertionJSON.Valid && assertionJSON.String != "" {
		var assertion Assertion
		if err = json.Unmarshal([]byte(assertionJSON.String), &assertion); err != nil {
			return Goal{}, err
		}
		g.Assertion = &assertion
	}
	var evidence ArtifactEvidence
	var verified int
	err = e.db.QueryRowContext(ctx, `SELECT sha256,byte_count,provenance,verified FROM evidence WHERE goal_id=?`, goalID).Scan(&evidence.SHA256, &evidence.ByteCount, &evidence.Provenance, &verified)
	if err == nil {
		evidence.Verified = verified == 1
		g.Evidence = &evidence
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Goal{}, err
	}
	return g, nil
}

// UpdateMetadata edits only owner-visible text for work that has not started.
// It cannot alter the action, grant, score, evidence, or job state.
func (e *Engine) UpdateMetadata(ctx context.Context, goalID string, title, description *string, ownerToken string) (Goal, error) {
	if !hmac.Equal([]byte(ownerToken), []byte(e.ownerToken)) {
		return Goal{}, ErrUnauthorized
	}
	if title == nil && description == nil {
		return Goal{}, fmt.Errorf("%w: GOAL_METADATA_REQUIRED", ErrInvalid)
	}
	var normalizedTitle, normalizedDescription any
	if title != nil {
		value := strings.TrimSpace(*title)
		if value == "" || utf8.RuneCountInString(value) > 256 {
			return Goal{}, fmt.Errorf("%w: INVALID_GOAL_TITLE", ErrInvalid)
		}
		normalizedTitle = value
	}
	if description != nil {
		if utf8.RuneCountInString(*description) > 4096 {
			return Goal{}, fmt.Errorf("%w: INVALID_GOAL_DESCRIPTION", ErrInvalid)
		}
		normalizedDescription = *description
	}
	if _, err := e.Get(ctx, goalID); err != nil {
		return Goal{}, err
	}
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return Goal{}, err
	}
	defer tx.Rollback()
	var status, jobState string
	err = tx.QueryRowContext(ctx, `SELECT COALESCE(c.state,g.status),j.state FROM goal g JOIN job j ON j.goal_id=g.id LEFT JOIN goal_control c ON c.goal_id=g.id WHERE g.id=?`, goalID).Scan(&status, &jobState)
	if errors.Is(err, sql.ErrNoRows) {
		return Goal{}, ErrNotFound
	}
	if err != nil {
		return Goal{}, err
	}
	if (status != "approved" && status != "blocked") || (jobState != "queued" && jobState != "blocked") {
		return Goal{}, ErrTransitionConflict
	}
	now := timestamp()
	if _, err = tx.ExecContext(ctx, `UPDATE goal SET title=COALESCE(?,title),description=COALESCE(?,description),updated_at_utc=? WHERE id=?`, normalizedTitle, normalizedDescription, now, goalID); err != nil {
		return Goal{}, err
	}
	if err = insertEvent(ctx, tx, goalID, "goal.metadata.updated", "OWNER_REQUEST", now); err != nil {
		return Goal{}, err
	}
	if err = tx.Commit(); err != nil {
		return Goal{}, err
	}
	return e.Get(ctx, goalID)
}

// SetPlan atomically replaces the owner-authenticated hierarchy parent and
// blocking prerequisites for one goal. It never creates an authorization grant.
func (e *Engine) SetPlan(ctx context.Context, goalID string, dependencyIDs []string, parentID *string, ownerToken string) (Goal, error) {
	if !hmac.Equal([]byte(ownerToken), []byte(e.ownerToken)) {
		return Goal{}, ErrUnauthorized
	}
	if goalID == "" || len(dependencyIDs) > 32 {
		return Goal{}, fmt.Errorf("%w: PLAN_TOO_LARGE_OR_INVALID", ErrInvalid)
	}
	seen := make(map[string]struct{}, len(dependencyIDs))
	for _, id := range dependencyIDs {
		if id == "" || id == goalID {
			return Goal{}, fmt.Errorf("%w: PLAN_INVALID", ErrInvalid)
		}
		if _, duplicate := seen[id]; duplicate {
			return Goal{}, fmt.Errorf("%w: PLAN_INVALID", ErrInvalid)
		}
		seen[id] = struct{}{}
	}
	if parentID != nil && (*parentID == "" || *parentID == goalID) {
		return Goal{}, fmt.Errorf("%w: PLAN_INVALID", ErrInvalid)
	}
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return Goal{}, err
	}
	defer tx.Rollback()
	var exists int
	if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM goal WHERE id=?`, goalID).Scan(&exists); err != nil {
		return Goal{}, err
	}
	if exists == 0 {
		return Goal{}, ErrNotFound
	}
	for _, id := range dependencyIDs {
		if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM goal WHERE id=?`, id).Scan(&exists); err != nil {
			return Goal{}, err
		}
		if exists == 0 {
			return Goal{}, ErrNotFound
		}
	}
	if parentID != nil {
		if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM goal WHERE id=?`, *parentID).Scan(&exists); err != nil {
			return Goal{}, err
		}
		if exists == 0 {
			return Goal{}, ErrNotFound
		}
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM goal_dependency WHERE goal_id=?`, goalID); err != nil {
		return Goal{}, err
	}
	var parent any
	if parentID != nil {
		parent = *parentID
	}
	if _, err = tx.ExecContext(ctx, `UPDATE goal SET parent_id=?,updated_at_utc=? WHERE id=?`, parent, timestamp(), goalID); err != nil {
		return Goal{}, fmt.Errorf("%w: PLAN_CYCLE_OR_INVALID", ErrInvalid)
	}
	for _, id := range dependencyIDs {
		if _, err = tx.ExecContext(ctx, `INSERT INTO goal_dependency(goal_id,depends_on_goal_id,dependency_type,created_at_utc) VALUES(?,?,'blocks',?)`, goalID, id, timestamp()); err != nil {
			return Goal{}, fmt.Errorf("%w: PLAN_CYCLE_OR_INVALID", ErrInvalid)
		}
	}
	if err = insertEvent(ctx, tx, goalID, "goal.plan.updated", "OWNER_PLAN", timestamp()); err != nil {
		return Goal{}, err
	}
	if err = tx.Commit(); err != nil {
		return Goal{}, err
	}
	return e.Get(ctx, goalID)
}

// Reprioritize changes only score inputs. It reauthorizes first and never
// changes the goal action, expected spend, assertion or execution evidence.
func (e *Engine) Reprioritize(ctx context.Context, goalID string, metrics Metrics, ownerToken string) (Goal, error) {
	if !hmac.Equal([]byte(ownerToken), []byte(e.ownerToken)) {
		return Goal{}, ErrUnauthorized
	}
	if metrics.Impact == nil || metrics.Urgency == nil || metrics.UserValue == nil || metrics.Confidence == nil || metrics.Efficiency == nil {
		return Goal{}, fmt.Errorf("%w: ALL_METRICS_REQUIRED", ErrInvalid)
	}
	goal, err := e.Get(ctx, goalID)
	if err != nil {
		return Goal{}, err
	}
	req, err := normalize(Request{Action: goal.Action, PrivacyClass: goal.PrivacyClass, Metrics: metrics, Assertion: goal.Assertion})
	if err != nil {
		return Goal{}, err
	}
	decision, err := e.policy.Authorize(ctx, goalID, req)
	if err != nil {
		return Goal{}, fmt.Errorf("policy authorization unavailable: %w", err)
	}
	if decision.Code == "" {
		return Goal{}, fmt.Errorf("policy returned an empty decision")
	}
	if decision.Allowed && decision.Grant == "" {
		return Goal{}, fmt.Errorf("policy allowed action without a scoped grant")
	}
	metricsJSON, err := json.Marshal(req.Metrics)
	if err != nil {
		return Goal{}, err
	}
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return Goal{}, err
	}
	defer tx.Rollback()
	var status, jobState string
	err = tx.QueryRowContext(ctx, `SELECT COALESCE(c.state,g.status),j.state FROM goal g JOIN job j ON j.goal_id=g.id LEFT JOIN goal_control c ON c.goal_id=g.id WHERE g.id=?`, goalID).Scan(&status, &jobState)
	if errors.Is(err, sql.ErrNoRows) {
		return Goal{}, ErrNotFound
	}
	if err != nil {
		return Goal{}, err
	}
	if jobState == "running" || (status != "approved" && status != "blocked") || (jobState != "queued" && jobState != "blocked") {
		return Goal{}, ErrTransitionConflict
	}
	now := timestamp()
	newStatus, newJobState, policyCode := "blocked", "blocked", decision.Code
	var scoreValue any
	var grant any
	if decision.Allowed {
		newStatus, newJobState = "approved", "queued"
		scoreValue = Score(req.Metrics)
		grant = decision.Grant
	}
	if _, err = tx.ExecContext(ctx, `UPDATE goal SET status=?,metrics_json=?,score=?,policy_code=?,grant_token=?,updated_at_utc=? WHERE id=?`, newStatus, string(metricsJSON), scoreValue, policyCode, grant, now, goalID); err != nil {
		return Goal{}, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE job SET state=?,lease_owner=NULL,lease_token=NULL,lease_expires_at_unix=NULL WHERE goal_id=?`, newJobState, goalID); err != nil {
		return Goal{}, err
	}
	policyEvent := "policy.denied"
	if decision.Allowed {
		policyEvent = "policy.allowed"
	}
	if err = insertEvent(ctx, tx, goalID, policyEvent, decision.Code, now); err != nil {
		return Goal{}, err
	}
	if !decision.Allowed {
		if err = insertEvent(ctx, tx, goalID, "goal.blocked", decision.Code, now); err != nil {
			return Goal{}, err
		}
	} else {
		if err = insertEvent(ctx, tx, goalID, "goal.reprioritized", "OWNER_REQUEST", now); err != nil {
			return Goal{}, err
		}
		if err = insertEvent(ctx, tx, goalID, "goal.scored", "WEIGHTED_V1", now); err != nil {
			return Goal{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return Goal{}, err
	}
	return e.Get(ctx, goalID)
}

func (e *Engine) Pause(ctx context.Context, goalID, ownerToken string) (Goal, error) {
	return e.controlGoal(ctx, goalID, ownerToken, "paused")
}

func (e *Engine) Cancel(ctx context.Context, goalID, ownerToken string) (Goal, error) {
	return e.controlGoal(ctx, goalID, ownerToken, "cancelled")
}

func (e *Engine) Resume(ctx context.Context, goalID, ownerToken string) (Goal, error) {
	if !hmac.Equal([]byte(ownerToken), []byte(e.ownerToken)) {
		return Goal{}, ErrUnauthorized
	}
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return Goal{}, err
	}
	defer tx.Rollback()
	var state, jobState string
	err = tx.QueryRowContext(ctx, `SELECT c.state,j.state FROM goal_control c JOIN job j ON j.goal_id=c.goal_id WHERE c.goal_id=?`, goalID).Scan(&state, &jobState)
	if errors.Is(err, sql.ErrNoRows) {
		var exists int
		if tx.QueryRowContext(ctx, `SELECT count(*) FROM goal WHERE id=?`, goalID).Scan(&exists) != nil || exists == 0 {
			return Goal{}, ErrNotFound
		}
		return Goal{}, ErrTransitionConflict
	}
	if err != nil {
		return Goal{}, err
	}
	if state != "paused" || jobState != "queued" {
		return Goal{}, ErrTransitionConflict
	}
	now := timestamp()
	if _, err = tx.ExecContext(ctx, `DELETE FROM goal_control WHERE goal_id=?`, goalID); err != nil {
		return Goal{}, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE goal SET updated_at_utc=? WHERE id=?`, now, goalID); err != nil {
		return Goal{}, err
	}
	if err = insertEvent(ctx, tx, goalID, "goal.resumed", "OWNER_REQUEST", now); err != nil {
		return Goal{}, err
	}
	if err = tx.Commit(); err != nil {
		return Goal{}, err
	}
	return e.Get(ctx, goalID)
}

func (e *Engine) controlGoal(ctx context.Context, goalID, ownerToken, target string) (Goal, error) {
	if !hmac.Equal([]byte(ownerToken), []byte(e.ownerToken)) {
		return Goal{}, ErrUnauthorized
	}
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return Goal{}, err
	}
	defer tx.Rollback()
	var status, jobState string
	err = tx.QueryRowContext(ctx, `SELECT COALESCE(c.state,g.status),j.state FROM goal g JOIN job j ON j.goal_id=g.id LEFT JOIN goal_control c ON c.goal_id=g.id WHERE g.id=?`, goalID).Scan(&status, &jobState)
	if errors.Is(err, sql.ErrNoRows) {
		return Goal{}, ErrNotFound
	}
	if err != nil {
		return Goal{}, err
	}
	if status == target {
		if err = tx.Commit(); err != nil {
			return Goal{}, err
		}
		return e.Get(ctx, goalID)
	}
	if jobState == "running" {
		return Goal{}, ErrTransitionConflict
	}
	if target == "paused" && (status != "approved" || jobState != "queued") {
		return Goal{}, ErrTransitionConflict
	}
	if target == "cancelled" && status != "approved" && status != "paused" && status != "blocked" {
		return Goal{}, ErrTransitionConflict
	}
	now := timestamp()
	if _, err = tx.ExecContext(ctx, `INSERT INTO goal_control(goal_id,state,updated_at_utc) VALUES(?,?,?) ON CONFLICT(goal_id) DO UPDATE SET state=excluded.state,updated_at_utc=excluded.updated_at_utc`, goalID, target, now); err != nil {
		return Goal{}, err
	}
	if err = insertEvent(ctx, tx, goalID, "goal."+target, "OWNER_REQUEST", now); err != nil {
		return Goal{}, err
	}
	if err = tx.Commit(); err != nil {
		return Goal{}, err
	}
	return e.Get(ctx, goalID)
}

func (e *Engine) List(ctx context.Context) ([]Goal, error) {
	rows, err := e.db.QueryContext(ctx, `SELECT id FROM goal ORDER BY COALESCE(score,-1) DESC,created_at_utc,id LIMIT 100`)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, 100)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	out := make([]Goal, 0, len(ids))
	for _, id := range ids {
		goal, err := e.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		out = append(out, goal)
	}
	return out, nil
}

func (e *Engine) Tick(ctx context.Context) (TickResult, error) {
	goalID, req, grant, leaseToken, err := e.claim(ctx)
	if err != nil {
		return TickResult{}, err
	}
	if goalID == "" {
		return TickResult{Outcome: "idle_success", Reason: "NO_AUTHORIZED_USEFUL_WORK"}, nil
	}
	decision, err := e.policy.Evaluate(ctx, goalID, req, grant)
	if err != nil {
		_ = e.releaseClaim(ctx, goalID, leaseToken, "POLICY_UNAVAILABLE")
		return TickResult{}, fmt.Errorf("policy revalidation unavailable: %w", err)
	}
	if !decision.Allowed {
		if err = e.finishBlocked(ctx, goalID, leaseToken, decision.Code); err != nil {
			return TickResult{}, err
		}
		return TickResult{Outcome: "blocked", GoalID: goalID, Reason: decision.Code}, nil
	}
	result, runErr := e.runner.Execute(ctx, req.Action)
	if runErr != nil {
		if err = e.finishFailed(ctx, goalID, leaseToken, "ACTION_FAILED"); err != nil {
			return TickResult{}, err
		}
		return TickResult{Outcome: "failed", GoalID: goalID, Reason: "ACTION_FAILED"}, nil
	}
	expectedBytes := int64(len(publicFixture))
	verified := result.SHA256 == publicFixtureSHA256 && result.ByteCount == expectedBytes && result.Provenance == "local_public_fixture_v1"
	if !verified {
		if err = e.finishFailed(ctx, goalID, leaseToken, "ARTIFACT_MISMATCH"); err != nil {
			return TickResult{}, err
		}
		return TickResult{Outcome: "failed", GoalID: goalID, Reason: "ARTIFACT_MISMATCH"}, nil
	}
	result.Verified = true
	if err = e.finishSucceeded(ctx, goalID, leaseToken, result); err != nil {
		return TickResult{}, err
	}
	return TickResult{Outcome: "succeeded", GoalID: goalID}, nil
}

func (e *Engine) claim(ctx context.Context) (string, Request, string, string, error) {
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return "", Request{}, "", "", err
	}
	defer tx.Rollback()
	nowUnix := time.Now().Unix()
	now := timestamp()
	if _, err = tx.ExecContext(ctx, `UPDATE job SET state=CASE WHEN attempt_count>=max_attempts THEN 'failed' ELSE 'queued' END,
		lease_owner=NULL,lease_token=NULL,lease_expires_at_unix=NULL
		WHERE state='running' AND lease_expires_at_unix<=?`, nowUnix); err != nil {
		return "", Request{}, "", "", err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE goal SET status='failed',updated_at_utc=? WHERE id IN
		(SELECT goal_id FROM job WHERE state='failed') AND status='approved'`, now); err != nil {
		return "", Request{}, "", "", err
	}
	var goalID, action, privacyClass, metricsJSON string
	var assertionJSON, grant sql.NullString
	var score sql.NullFloat64
	err = tx.QueryRowContext(ctx, `SELECT g.id,g.action,g.privacy_class,g.metrics_json,g.assertion_json,g.grant_token,g.score
		FROM goal g JOIN job j ON j.goal_id=g.id
		WHERE j.state='queued' AND g.status='approved' AND g.score>0 AND NOT EXISTS (SELECT 1 FROM goal_control c WHERE c.goal_id=g.id)
		AND NOT EXISTS (SELECT 1 FROM goal_dependency d JOIN goal prerequisite ON prerequisite.id=d.depends_on_goal_id WHERE d.goal_id=g.id AND prerequisite.status!='completed')
		ORDER BY g.score DESC,g.created_at_utc,g.id LIMIT 1`).Scan(&goalID, &action, &privacyClass, &metricsJSON, &assertionJSON, &grant, &score)
	if errors.Is(err, sql.ErrNoRows) {
		if err = tx.Commit(); err != nil {
			return "", Request{}, "", "", err
		}
		return "", Request{}, "", "", nil
	}
	if err != nil {
		return "", Request{}, "", "", err
	}
	token, err := newID()
	if err != nil {
		return "", Request{}, "", "", err
	}
	worker, err := newID()
	if err != nil {
		return "", Request{}, "", "", err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE job SET state='running',attempt_count=attempt_count+1,lease_owner=?,lease_token=?,lease_expires_at_unix=? WHERE goal_id=? AND state='queued'`, worker, token, nowUnix+int64(leaseDuration.Seconds()), goalID); err != nil {
		return "", Request{}, "", "", err
	}
	if err = insertEvent(ctx, tx, goalID, "goal.started", "POLICY_RECHECK_PENDING", now); err != nil {
		return "", Request{}, "", "", err
	}
	if err = tx.Commit(); err != nil {
		return "", Request{}, "", "", err
	}
	req := Request{Action: action, PrivacyClass: privacyClass}
	if err = json.Unmarshal([]byte(metricsJSON), &req.Metrics); err != nil {
		return "", Request{}, "", "", err
	}
	if assertionJSON.Valid {
		var assertion Assertion
		if err = json.Unmarshal([]byte(assertionJSON.String), &assertion); err != nil {
			return "", Request{}, "", "", err
		}
		req.Assertion = &assertion
	}
	return goalID, req, grant.String, token, nil
}

func (e *Engine) finishSucceeded(ctx context.Context, goalID, leaseToken string, evidence ArtifactEvidence) error {
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	now := timestamp()
	result, err := tx.ExecContext(ctx, `UPDATE job SET state='succeeded',lease_owner=NULL,lease_token=NULL,lease_expires_at_unix=NULL
		WHERE goal_id=? AND state='running' AND lease_token=?`, goalID, leaseToken)
	if err != nil {
		return err
	}
	if changed, _ := result.RowsAffected(); changed != 1 {
		return fmt.Errorf("stale execution lease")
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO evidence(goal_id,sha256,byte_count,provenance,verified,created_at_utc) VALUES(?,?,?,?,1,?)`,
		goalID, evidence.SHA256, evidence.ByteCount, evidence.Provenance, now); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE goal SET status='completed',updated_at_utc=? WHERE id=?`, now, goalID); err != nil {
		return err
	}
	if err = insertEvent(ctx, tx, goalID, "goal.completed", "ARTIFACT_VERIFIED", now); err != nil {
		return err
	}
	return tx.Commit()
}

func (e *Engine) finishBlocked(ctx context.Context, goalID, leaseToken, code string) error {
	if code == "" {
		code = "POLICY_DENIED"
	}
	return e.finishTerminal(ctx, goalID, leaseToken, "blocked", "blocked", "policy.denied", code)
}

func (e *Engine) finishFailed(ctx context.Context, goalID, leaseToken, code string) error {
	return e.finishTerminal(ctx, goalID, leaseToken, "failed", "failed", "goal.failed", code)
}

func (e *Engine) finishTerminal(ctx context.Context, goalID, leaseToken, goalStatus, jobStatus, eventKind, code string) error {
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	now := timestamp()
	result, err := tx.ExecContext(ctx, `UPDATE job SET state=?,lease_owner=NULL,lease_token=NULL,lease_expires_at_unix=NULL
		WHERE goal_id=? AND state='running' AND lease_token=?`, jobStatus, goalID, leaseToken)
	if err != nil {
		return err
	}
	if changed, _ := result.RowsAffected(); changed != 1 {
		return fmt.Errorf("stale execution lease")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE goal SET status=?,policy_code=?,updated_at_utc=? WHERE id=?`, goalStatus, code, now, goalID); err != nil {
		return err
	}
	if err = insertEvent(ctx, tx, goalID, eventKind, code, now); err != nil {
		return err
	}
	return tx.Commit()
}

func (e *Engine) releaseClaim(ctx context.Context, goalID, leaseToken, code string) error {
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `UPDATE job SET state='queued',lease_owner=NULL,lease_token=NULL,lease_expires_at_unix=NULL
		WHERE goal_id=? AND state='running' AND lease_token=?`, goalID, leaseToken); err != nil {
		return err
	}
	if err = insertEvent(ctx, tx, goalID, "policy.deferred", code, timestamp()); err != nil {
		return err
	}
	return tx.Commit()
}

func (e *Engine) findIdempotent(ctx context.Context, key string) (Goal, string, error) {
	var id, requestHash string
	err := e.db.QueryRowContext(ctx, `SELECT id,request_hash FROM goal WHERE idempotency_key=?`, key).Scan(&id, &requestHash)
	if errors.Is(err, sql.ErrNoRows) {
		return Goal{}, "", ErrNotFound
	}
	if err != nil {
		return Goal{}, "", err
	}
	g, err := e.Get(ctx, id)
	return g, requestHash, err
}

// Decisions returns up to the newest 100 safe audit events in chronological order.
func (e *Engine) Decisions(ctx context.Context, goalID string) ([]GoalDecision, error) {
	var exists int
	if err := e.db.QueryRowContext(ctx, `SELECT 1 FROM goal WHERE id=?`, goalID).Scan(&exists); errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	rows, err := e.db.QueryContext(ctx, `SELECT id,kind,code,created_at_utc FROM (
		SELECT id,kind,code,created_at_utc FROM goal_event WHERE goal_id=?
		ORDER BY created_at_utc DESC,id DESC LIMIT 100
	) ORDER BY created_at_utc,id`, goalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]GoalDecision, 0, 100)
	for rows.Next() {
		var decision GoalDecision
		if err := rows.Scan(&decision.ID, &decision.Kind, &decision.Code, &decision.CreatedAtUTC); err != nil {
			return nil, err
		}
		out = append(out, decision)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func normalize(input Request) (Request, error) {
	if !actionPattern.MatchString(input.Action) {
		return Request{}, fmt.Errorf("%w: INVALID_ACTION", ErrInvalid)
	}
	if input.PrivacyClass == "" {
		input.PrivacyClass = "general"
	}
	if input.PrivacyClass != "general" && input.PrivacyClass != "intimate" && input.PrivacyClass != "biometric" {
		return Request{}, fmt.Errorf("%w: INVALID_PRIVACY_CLASS", ErrInvalid)
	}
	if input.ExpectedCostMinor < 0 {
		return Request{}, fmt.Errorf("%w: INVALID_EXPECTED_COST", ErrInvalid)
	}
	if input.Title != "" {
		input.Title = strings.TrimSpace(input.Title)
		if input.Title == "" || utf8.RuneCountInString(input.Title) > 256 {
			return Request{}, fmt.Errorf("%w: INVALID_GOAL_TITLE", ErrInvalid)
		}
	}
	if utf8.RuneCountInString(input.Description) > 4096 {
		return Request{}, fmt.Errorf("%w: INVALID_GOAL_DESCRIPTION", ErrInvalid)
	}
	var err error
	input.Metrics.Impact, err = metric(input.Metrics.Impact)
	if err != nil {
		return Request{}, err
	}
	input.Metrics.Urgency, err = metric(input.Metrics.Urgency)
	if err != nil {
		return Request{}, err
	}
	input.Metrics.UserValue, err = metric(input.Metrics.UserValue)
	if err != nil {
		return Request{}, err
	}
	input.Metrics.Confidence, err = metric(input.Metrics.Confidence)
	if err != nil {
		return Request{}, err
	}
	input.Metrics.Efficiency, err = metric(input.Metrics.Efficiency)
	if err != nil {
		return Request{}, err
	}
	if input.Assertion != nil {
		if len(input.Assertion.Claim) < 1 || len(input.Assertion.Claim) > 512 || input.Assertion.Source != nil && len(*input.Assertion.Source) > 256 || math.IsNaN(input.Assertion.Confidence) || math.IsInf(input.Assertion.Confidence, 0) || input.Assertion.Confidence < 0 || input.Assertion.Confidence > 1 {
			return Request{}, fmt.Errorf("%w: INVALID_ASSERTION", ErrInvalid)
		}
		input.Assertion.ProvenanceStatus = "unknown"
		input.Assertion.EpistemicStatus = "hypothesis"
	}
	if input.PrivacyClass != "general" {
		input.Assertion = nil
	}
	return input, nil
}

func metric(value *float64) (*float64, error) {
	if value == nil {
		v := 1.0
		return &v, nil
	}
	if math.IsNaN(*value) || math.IsInf(*value, 0) || *value < 0 || *value > 1 {
		return nil, fmt.Errorf("%w: METRICS_MUST_BE_FINITE_0_TO_1", ErrInvalid)
	}
	v := *value
	return &v, nil
}

func Score(m Metrics) float64 {
	value := .30**m.Impact + .20**m.Urgency + .20**m.UserValue + .15**m.Confidence + .15**m.Efficiency
	return math.Round(value*100*1_000_000) / 1_000_000
}

func provenanceFor(assertion *Assertion) string {
	if assertion == nil {
		return "unknown"
	}
	return assertion.ProvenanceStatus
}

func nullableJSON(value any) any {
	if value == nil {
		return nil
	}
	return string(value.([]byte))
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func insertEvent(ctx context.Context, tx *sql.Tx, goalID, kind, code, at string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO goal_event(goal_id,kind,code,created_at_utc) VALUES(?,?,?,?)`, goalID, kind, code, at)
	return err
}

func timestamp() string { return time.Now().UTC().Format("2006-01-02T15:04:05.000000Z") }

func newID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	encoded := hex.EncodeToString(value[:])
	return encoded[:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:], nil
}

func createPrivateDatabase(path string) error {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		file, createErr := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if createErr != nil {
			return createErr
		}
		return file.Close()
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 {
		return fmt.Errorf("private regular database file required")
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); ok && uint64(stat.Uid) != uint64(os.Getuid()) {
		return fmt.Errorf("database owner mismatch")
	}
	return nil
}

func requirePrivateDirectory(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() || info.Mode().Perm()&0077 != 0 {
		return fmt.Errorf("owner-only database directory required")
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); ok && uint64(stat.Uid) != uint64(os.Getuid()) {
		return fmt.Errorf("database directory owner mismatch")
	}
	return nil
}
