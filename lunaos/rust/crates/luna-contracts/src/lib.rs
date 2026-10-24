use serde::{Deserialize, Serialize};

pub const CONTRACT_VERSION: u16 = 1;

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct CreateGoal {
    pub action: String,
    #[serde(default)]
    pub metrics: Metrics,
    #[serde(default)]
    pub privacy_class: PrivacyClass,
    #[serde(default)]
    pub expected_cost_minor: Option<u64>,
    #[serde(default)]
    pub provenance: ProvenanceInput,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct AddDependency {
    pub depends_on_goal_id: String,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct GoalPlan {
    pub parent_goal_id: Option<String>,
    pub dependency_goal_ids: Vec<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
pub struct GoalPlanView {
    pub goal_id: String,
    pub parent_goal_id: Option<String>,
    pub dependency_goal_ids: Vec<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(rename_all = "snake_case")]
pub enum GoalLifecycleCommand {
    Pause,
    Resume,
    Cancel,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
#[serde(default)]
pub struct Metrics {
    pub impact: f64,
    pub urgency: f64,
    pub user_value: f64,
    pub confidence: f64,
    pub efficiency: f64,
}

impl Default for Metrics {
    fn default() -> Self {
        Self {
            impact: 1.0,
            urgency: 1.0,
            user_value: 1.0,
            confidence: 1.0,
            efficiency: 1.0,
        }
    }
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Default)]
#[serde(rename_all = "snake_case")]
pub enum PrivacyClass {
    #[default]
    General,
    Intimate,
    Biometric,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Default)]
#[serde(deny_unknown_fields)]
pub struct ProvenanceInput {
    pub source_id: Option<String>,
    pub claim: Option<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Default)]
#[serde(rename_all = "snake_case")]
pub enum SourceStatus {
    #[default]
    Unknown,
    Verified,
    Disputed,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
pub struct GoalView {
    pub contract_version: u16,
    pub id: String,
    pub action: String,
    pub status: String,
    pub score: Option<f64>,
    pub policy_code: String,
    pub provenance_status: SourceStatus,
    pub queued: bool,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
pub struct Evidence {
    pub goal_id: String,
    pub sha256: String,
    pub byte_count: u64,
    pub provenance: String,
    pub verified: bool,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
pub struct PolicyRequest {
    pub owner_token: String,
    pub principal: Principal,
    pub action: String,
    pub privacy_class: PrivacyClass,
    pub expected_cost_minor: Option<u64>,
    pub configured_budget_minor: Option<u64>,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(rename_all = "snake_case")]
pub enum Principal {
    Owner,
    Guest,
    System,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
pub struct PolicyDecision {
    pub allowed: bool,
    pub code: String,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
pub struct TickResult {
    pub outcome: String,
    pub goal_id: Option<String>,
    pub reason: String,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
pub struct ExecutionLease {
    pub goal_id: String,
    pub action: String,
    pub grant: String,
    pub owner: String,
    pub token: String,
    pub expires_at_unix: i64,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
pub struct GoalEventView {
    pub event_id: i64,
    pub goal_id: String,
    pub kind: String,
    pub decision_code: String,
    pub created_at_utc: String,
}

/// Versioned logical entities used by inter-service APIs. These are schemas only
/// until the corresponding Rust service and migration are implemented.
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
pub struct GoalDependency {
    pub goal_id: String,
    pub depends_on_goal_id: String,
    pub dependency_type: String,
    pub created_at_utc: String,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
pub struct GoalExecution {
    pub id: String,
    pub goal_id: String,
    pub agent_id: String,
    pub attempt: u32,
    pub state: String,
    pub lease_owner: Option<String>,
    pub lease_expires_at_utc: Option<String>,
    pub idempotency_key: String,
    pub started_at_utc: Option<String>,
    pub finished_at_utc: Option<String>,
    pub stop_reason: Option<String>,
    pub outcome_score: Option<f64>,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
pub struct GoalEvent {
    pub event_id: String,
    pub schema_version: u16,
    pub goal_id: String,
    pub actor_id: String,
    pub event_type: String,
    pub decision_code: String,
    pub correlation_id: String,
    pub occurred_at_utc: String,
    pub evidence_ids: Vec<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
pub struct PolicyGrant {
    pub grant_id: String,
    pub principal_id: String,
    pub action_family: String,
    pub target_ids: Vec<String>,
    pub data_classes: Vec<String>,
    pub recipient_ids: Vec<String>,
    pub monetary_limit_minor: Option<u64>,
    pub currency: Option<String>,
    pub valid_from_utc: String,
    pub valid_to_utc: String,
    pub revoked_at_utc: Option<String>,
    pub consent_receipt_id: Option<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
pub struct StrategyTrial {
    pub id: String,
    pub goal_id: String,
    pub version: u32,
    pub hypothesis: String,
    pub test_suite: Vec<String>,
    pub measurements: serde_json::Value,
    pub artifact_ids: Vec<String>,
    pub status: String,
    pub baseline_comparison: serde_json::Value,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
pub struct MemoryAssertion {
    pub id: String,
    pub subject_scope: String,
    pub claim: String,
    pub source_ref: Option<String>,
    pub provenance_status: SourceStatus,
    pub confidence: f64,
    pub supersedes_id: Option<String>,
    pub privacy_class: PrivacyClass,
    pub recorded_at_utc: String,
    pub reviewed_at_utc: Option<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
pub struct PersonaRevision {
    pub id: String,
    pub predecessor_id: Option<String>,
    pub change_set: serde_json::Value,
    pub evidence_ids: Vec<String>,
    pub applied_at_utc: Option<String>,
    pub evaluation_status: String,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
pub struct ExpenseIntent {
    pub id: String,
    pub goal_id: String,
    pub merchant_id: String,
    pub category: String,
    pub amount_minor: u64,
    pub currency: String,
    pub state: String,
    pub reservation_id: Option<String>,
    pub source_grant_id: String,
    pub provider_reference: Option<String>,
    pub created_at_utc: String,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
pub struct ExternalConversationGrant {
    pub id: String,
    pub platform: String,
    pub conversation_id: String,
    pub recipient_ids: Vec<String>,
    pub message_classes: Vec<String>,
    pub purpose: String,
    pub expires_at_utc: String,
    pub revoked_at_utc: Option<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
pub struct ConsentReceipt {
    pub id: String,
    pub subject_id: String,
    pub scopes: Vec<String>,
    pub purpose: String,
    pub evidence_ref: String,
    pub granted_at_utc: String,
    pub expires_at_utc: Option<String>,
    pub revoked_at_utc: Option<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
pub struct DataErasureJob {
    pub id: String,
    pub subject_id: String,
    pub source_record_ids: Vec<String>,
    pub derivative_ids: Vec<String>,
    pub status: String,
    pub requested_at_utc: String,
    pub completed_at_utc: Option<String>,
    pub audit_receipt_id: Option<String>,
}
