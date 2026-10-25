package intimacy

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrDiscoveryExpansionUnconfirmed = errors.New("DISCOVERY_EXPANSION_UNCONFIRMED")
	ErrDiscoveryPolicyUnconfirmed    = errors.New("DISCOVERY_POLICY_UNCONFIRMED")
	ErrInvalidDiscoveryAllowlist     = errors.New("INVALID_DISCOVERY_ALLOWLIST")
)

// DiscoveryAllowlist is an owner-configured metadata policy. It does not prove
// media legality, subject age, consent, or content safety.
type DiscoveryAllowlist struct {
	Enabled       bool     `json:"enabled"`
	Sources       []string `json:"sources"`
	Creators      []string `json:"creators"`
	Categories    []string `json:"categories"`
	ModelFamilies []string `json:"model_families"`
	Tags          []string `json:"tags"`
}

type DiscoveryCandidate struct {
	Source      string   `json:"source"`
	Creator     string   `json:"creator"`
	Category    string   `json:"category"`
	ModelFamily string   `json:"model_family"`
	Tags        []string `json:"tags"`
}

type DiscoveryDecision struct {
	Allowed bool   `json:"allowed"`
	Code    string `json:"code"`
}

type DiscoveryAllowlistExpansion struct {
	Sources       []string
	Creators      []string
	Categories    []string
	ModelFamilies []string
	Tags          []string
}

func EvaluateDiscoveryCandidate(policy DiscoveryAllowlist, candidate DiscoveryCandidate) DiscoveryDecision {
	if !policy.Enabled {
		return DiscoveryDecision{Code: "POLICY_DISABLED"}
	}
	if !validAllowlist(policy) {
		return DiscoveryDecision{Code: "POLICY_UNCONFIGURED"}
	}
	if !validCandidate(candidate) {
		return DiscoveryDecision{Code: "ALLOWLIST_MISS"}
	}
	if !contains(policy.Sources, candidate.Source) ||
		!contains(policy.Creators, candidate.Creator) ||
		!contains(policy.Categories, candidate.Category) ||
		!contains(policy.ModelFamilies, candidate.ModelFamily) {
		return DiscoveryDecision{Code: "ALLOWLIST_MISS"}
	}
	for _, tag := range candidate.Tags {
		if !contains(policy.Tags, tag) {
			return DiscoveryDecision{Code: "ALLOWLIST_MISS"}
		}
	}
	return DiscoveryDecision{Allowed: true, Code: "ALLOWLIST_MATCH"}
}

// ExpandDiscoveryAllowlist returns a copy and only applies additions after an
// explicit owner confirmation. Callers must persist the returned policy through
// their separately authenticated owner settings path.
func ExpandDiscoveryAllowlist(current DiscoveryAllowlist, expansion DiscoveryAllowlistExpansion, ownerConfirmed bool) (DiscoveryAllowlist, error) {
	if !ownerConfirmed {
		return DiscoveryAllowlist{}, ErrDiscoveryExpansionUnconfirmed
	}
	if !validAllowlist(current) || !validExpansion(expansion) {
		return DiscoveryAllowlist{}, ErrInvalidDiscoveryAllowlist
	}
	updated := cloneAllowlist(current)
	updated.Sources = mergeValues(updated.Sources, expansion.Sources)
	updated.Creators = mergeValues(updated.Creators, expansion.Creators)
	updated.Categories = mergeValues(updated.Categories, expansion.Categories)
	updated.ModelFamilies = mergeValues(updated.ModelFamilies, expansion.ModelFamilies)
	updated.Tags = mergeValues(updated.Tags, expansion.Tags)
	if !validAllowlist(updated) {
		return DiscoveryAllowlist{}, ErrInvalidDiscoveryAllowlist
	}
	return updated, nil
}

func validAllowlist(policy DiscoveryAllowlist) bool {
	return validValues(policy.Sources) && validValues(policy.Creators) && validValues(policy.Categories) && validValues(policy.ModelFamilies) && validValues(policy.Tags)
}

func validCandidate(candidate DiscoveryCandidate) bool {
	return validValue(candidate.Source) && validValue(candidate.Creator) && validValue(candidate.Category) && validValue(candidate.ModelFamily) && validValues(candidate.Tags)
}

func validExpansion(expansion DiscoveryAllowlistExpansion) bool {
	values := [][]string{expansion.Sources, expansion.Creators, expansion.Categories, expansion.ModelFamilies, expansion.Tags}
	any := false
	for _, list := range values {
		if len(list) != 0 {
			any = true
		}
		if len(list) > 0 && !validValues(list) {
			return false
		}
	}
	return any
}

func validValues(values []string) bool {
	if len(values) == 0 || len(values) > 256 {
		return false
	}
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		normalized, ok := normalizedValue(value)
		if !ok || seen[normalized] {
			return false
		}
		seen[normalized] = true
	}
	return true
}

func validValue(value string) bool {
	_, ok := normalizedValue(value)
	return ok
}

func normalizedValue(value string) (string, bool) {
	value = strings.ToLower(strings.TrimSpace(value))
	if len(value) == 0 || len(value) > 128 {
		return "", false
	}
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return "", false
		}
	}
	return value, true
}

func contains(values []string, value string) bool {
	normalized, ok := normalizedValue(value)
	if !ok {
		return false
	}
	for _, candidate := range values {
		if allowed, valid := normalizedValue(candidate); valid && allowed == normalized {
			return true
		}
	}
	return false
}

func mergeValues(current, additions []string) []string {
	merged := append([]string(nil), current...)
	for _, addition := range additions {
		if !contains(merged, addition) {
			merged = append(merged, strings.TrimSpace(addition))
		}
	}
	return merged
}

func cloneAllowlist(policy DiscoveryAllowlist) DiscoveryAllowlist {
	policy.Sources = append([]string(nil), policy.Sources...)
	policy.Creators = append([]string(nil), policy.Creators...)
	policy.Categories = append([]string(nil), policy.Categories...)
	policy.ModelFamilies = append([]string(nil), policy.ModelFamilies...)
	policy.Tags = append([]string(nil), policy.Tags...)
	return policy
}

const discoveryPolicyID = "owner"

// SetDiscoveryPolicy stores owner-authorized metadata encrypted separately
// from the general Core database. No discovery candidate or media is persisted.
func (v *Vault) SetDiscoveryPolicy(ctx context.Context, consentID string, policy DiscoveryAllowlist, ownerConfirmed bool) error {
	if !ownerConfirmed {
		return ErrDiscoveryPolicyUnconfirmed
	}
	if !validPersistedPolicy(policy) {
		return ErrInvalidDiscoveryAllowlist
	}
	tx, err := v.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = authorizeQuery(ctx, tx, consentID, "owner", ScopeDiscoveryPolicyWrite); err != nil {
		return err
	}
	ciphertext, err := v.sealDiscoveryPolicy(policy)
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err = tx.ExecContext(ctx, `INSERT INTO private_discovery_policy(subject_id,ciphertext,updated_at_utc) VALUES('owner',?,?) ON CONFLICT(subject_id) DO UPDATE SET ciphertext=excluded.ciphertext,updated_at_utc=excluded.updated_at_utc`, ciphertext, now); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO discovery_audit(event_code,subject_id,created_at_utc) VALUES('DISCOVERY_POLICY_UPDATED','owner',?)`, now); err != nil {
		return err
	}
	return tx.Commit()
}

func (v *Vault) ReadDiscoveryPolicy(ctx context.Context, consentID string) (DiscoveryAllowlist, error) {
	tx, err := v.db.BeginTx(ctx, nil)
	if err != nil {
		return DiscoveryAllowlist{}, err
	}
	defer tx.Rollback()
	if err = authorizeQuery(ctx, tx, consentID, "owner", ScopeDiscoveryPolicyRead); err != nil {
		return DiscoveryAllowlist{}, err
	}
	policy, err := v.readDiscoveryPolicyQuery(ctx, tx)
	if err != nil {
		return DiscoveryAllowlist{}, err
	}
	if err = tx.Commit(); err != nil {
		return DiscoveryAllowlist{}, err
	}
	return policy, nil
}

func (v *Vault) EvaluateDiscovery(ctx context.Context, consentID string, candidate DiscoveryCandidate) (DiscoveryDecision, error) {
	tx, err := v.db.BeginTx(ctx, nil)
	if err != nil {
		return DiscoveryDecision{}, err
	}
	defer tx.Rollback()
	if err = authorizeQuery(ctx, tx, consentID, "owner", ScopeDiscoveryRun); err != nil {
		return DiscoveryDecision{}, err
	}
	policy, err := v.readDiscoveryPolicyQuery(ctx, tx)
	if err != nil {
		return DiscoveryDecision{}, err
	}
	if err = tx.Commit(); err != nil {
		return DiscoveryDecision{}, err
	}
	return EvaluateDiscoveryCandidate(policy, candidate), nil
}

func (v *Vault) readDiscoveryPolicyQuery(ctx context.Context, queryer rowQueryer) (DiscoveryAllowlist, error) {
	var ciphertext []byte
	err := queryer.QueryRowContext(ctx, `SELECT ciphertext FROM private_discovery_policy WHERE subject_id='owner'`).Scan(&ciphertext)
	if errors.Is(err, sql.ErrNoRows) {
		return DiscoveryAllowlist{}, nil
	}
	if err != nil {
		return DiscoveryAllowlist{}, err
	}
	plain, err := v.openDiscoveryPolicy(ciphertext)
	if err != nil {
		return DiscoveryAllowlist{}, ErrInvalidDiscoveryAllowlist
	}
	var policy DiscoveryAllowlist
	if json.Unmarshal(plain, &policy) != nil || !validPersistedPolicy(policy) {
		return DiscoveryAllowlist{}, ErrInvalidDiscoveryAllowlist
	}
	return policy, nil
}

func (v *Vault) sealDiscoveryPolicy(policy DiscoveryAllowlist) ([]byte, error) {
	plain, err := json.Marshal(policy)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, v.aead.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}
	return v.aead.Seal(nonce, nonce, plain, discoveryPolicyAAD()), nil
}

func (v *Vault) openDiscoveryPolicy(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < v.aead.NonceSize() {
		return nil, ErrInvalidDiscoveryAllowlist
	}
	nonce, body := ciphertext[:v.aead.NonceSize()], ciphertext[v.aead.NonceSize():]
	return v.aead.Open(nil, nonce, body, discoveryPolicyAAD())
}

func discoveryPolicyAAD() []byte {
	return []byte(discoveryPolicyID + "\x00owner\x00discovery-policy.v1")
}

func validPersistedPolicy(policy DiscoveryAllowlist) bool {
	if policy.Enabled {
		return validAllowlist(policy)
	}
	return validOptionalValues(policy.Sources) && validOptionalValues(policy.Creators) && validOptionalValues(policy.Categories) && validOptionalValues(policy.ModelFamilies) && validOptionalValues(policy.Tags)
}

func validOptionalValues(values []string) bool {
	return len(values) == 0 || validValues(values)
}
