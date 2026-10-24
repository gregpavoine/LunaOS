use axum::{
    Json, Router,
    extract::{Path, State},
    http::{HeaderMap, StatusCode},
    routing::{get, post},
};
use luna_contracts::{
    AddDependency, CreateGoal, Evidence, GoalEventView, GoalLifecycleCommand, GoalPlan,
    GoalPlanView, GoalView, TickResult,
};
use luna_goal_engine::{EngineError, GoalEngine};
use luna_policy_broker::UnixPolicyClient;
use std::sync::Arc;

#[derive(Clone)]
struct AppState {
    engine: GoalEngine,
    owner_token: Arc<String>,
}

pub fn router(engine: GoalEngine, owner_token: String) -> Router {
    Router::new()
        .route("/v1/goals", get(list_goals).post(create_goal))
        .route("/v1/goals/{id}", get(get_goal))
        .route("/v1/goals/{id}/events", get(get_goal_events))
        .route("/v1/goals/{id}/evidence", get(get_goal_evidence))
        .route("/v1/goals/{id}/dependencies", post(add_dependency))
        .route(
            "/v1/goals/{id}/plan",
            get(get_goal_plan).put(replace_goal_plan),
        )
        .route("/v1/goals/{id}/pause", post(pause_goal))
        .route("/v1/goals/{id}/resume", post(resume_goal))
        .route("/v1/goals/{id}/cancel", post(cancel_goal))
        .route("/v1/scheduler/tick", post(tick))
        .route("/v1/system/capabilities", get(capabilities))
        .with_state(AppState {
            engine,
            owner_token: Arc::new(owner_token),
        })
}

fn authenticated(headers: &HeaderMap, state: &AppState) -> bool {
    headers
        .get("authorization")
        .and_then(|v| v.to_str().ok())
        .and_then(|v| v.strip_prefix("Bearer "))
        .is_some_and(|v| subtle_token_eq(v, &state.owner_token))
}
fn subtle_token_eq(a: &str, b: &str) -> bool {
    use subtle::ConstantTimeEq;
    a.as_bytes().ct_eq(b.as_bytes()).into()
}
fn error(err: EngineError) -> (StatusCode, String) {
    match err {
        EngineError::NotFound => (StatusCode::NOT_FOUND, "GOAL_NOT_FOUND".into()),
        EngineError::Invalid(code) => (StatusCode::BAD_REQUEST, code),
        EngineError::Policy(_) => (StatusCode::FORBIDDEN, "POLICY_DENIED".into()),
        EngineError::Storage(_) => (StatusCode::INTERNAL_SERVER_ERROR, "STORAGE_ERROR".into()),
    }
}

async fn create_goal(
    State(state): State<AppState>,
    headers: HeaderMap,
    Json(body): Json<CreateGoal>,
) -> Result<Json<GoalView>, (StatusCode, String)> {
    if !authenticated(&headers, &state) {
        return Err((StatusCode::UNAUTHORIZED, "OWNER_AUTH_REQUIRED".into()));
    }
    let key = headers
        .get("idempotency-key")
        .and_then(|v| v.to_str().ok())
        .unwrap_or_default();
    state
        .engine
        .create(&body, &state.owner_token, key)
        .map(Json)
        .map_err(error)
}
async fn list_goals(
    State(state): State<AppState>,
    headers: HeaderMap,
) -> Result<Json<Vec<GoalView>>, (StatusCode, String)> {
    if !authenticated(&headers, &state) {
        return Err((StatusCode::UNAUTHORIZED, "OWNER_AUTH_REQUIRED".into()));
    }
    state.engine.list().map(Json).map_err(error)
}
async fn get_goal(
    State(state): State<AppState>,
    headers: HeaderMap,
    Path(id): Path<String>,
) -> Result<Json<GoalView>, (StatusCode, String)> {
    if !authenticated(&headers, &state) {
        return Err((StatusCode::UNAUTHORIZED, "OWNER_AUTH_REQUIRED".into()));
    }
    state.engine.get(&id).map(Json).map_err(error)
}
async fn get_goal_events(
    State(state): State<AppState>,
    headers: HeaderMap,
    Path(id): Path<String>,
) -> Result<Json<Vec<GoalEventView>>, (StatusCode, String)> {
    if !authenticated(&headers, &state) {
        return Err((StatusCode::UNAUTHORIZED, "OWNER_AUTH_REQUIRED".into()));
    }
    state.engine.events(&id, 100).map(Json).map_err(error)
}
async fn get_goal_evidence(
    State(state): State<AppState>,
    headers: HeaderMap,
    Path(id): Path<String>,
) -> Result<Json<Evidence>, (StatusCode, String)> {
    if !authenticated(&headers, &state) {
        return Err((StatusCode::UNAUTHORIZED, "OWNER_AUTH_REQUIRED".into()));
    }
    state
        .engine
        .evidence(&id)
        .map_err(error)?
        .map(Json)
        .ok_or((StatusCode::NOT_FOUND, "EVIDENCE_NOT_FOUND".into()))
}

async fn add_dependency(
    State(state): State<AppState>,
    headers: HeaderMap,
    Path(id): Path<String>,
    Json(body): Json<AddDependency>,
) -> Result<StatusCode, (StatusCode, String)> {
    if !authenticated(&headers, &state) {
        return Err((StatusCode::UNAUTHORIZED, "OWNER_AUTH_REQUIRED".into()));
    }
    state
        .engine
        .add_dependency(&id, &body.depends_on_goal_id)
        .map(|()| StatusCode::NO_CONTENT)
        .map_err(error)
}
async fn get_goal_plan(
    State(state): State<AppState>,
    headers: HeaderMap,
    Path(id): Path<String>,
) -> Result<Json<GoalPlanView>, (StatusCode, String)> {
    if !authenticated(&headers, &state) {
        return Err((StatusCode::UNAUTHORIZED, "OWNER_AUTH_REQUIRED".into()));
    }
    state.engine.get_goal_plan(&id).map(Json).map_err(error)
}
async fn replace_goal_plan(
    State(state): State<AppState>,
    headers: HeaderMap,
    Path(id): Path<String>,
    Json(body): Json<GoalPlan>,
) -> Result<Json<GoalPlanView>, (StatusCode, String)> {
    if !authenticated(&headers, &state) {
        return Err((StatusCode::UNAUTHORIZED, "OWNER_AUTH_REQUIRED".into()));
    }
    state
        .engine
        .set_goal_plan(&id, &body)
        .map(Json)
        .map_err(error)
}
async fn pause_goal(
    State(state): State<AppState>,
    headers: HeaderMap,
    Path(id): Path<String>,
) -> Result<Json<GoalView>, (StatusCode, String)> {
    lifecycle(state, headers, id, GoalLifecycleCommand::Pause)
}
async fn resume_goal(
    State(state): State<AppState>,
    headers: HeaderMap,
    Path(id): Path<String>,
) -> Result<Json<GoalView>, (StatusCode, String)> {
    lifecycle(state, headers, id, GoalLifecycleCommand::Resume)
}
async fn cancel_goal(
    State(state): State<AppState>,
    headers: HeaderMap,
    Path(id): Path<String>,
) -> Result<Json<GoalView>, (StatusCode, String)> {
    lifecycle(state, headers, id, GoalLifecycleCommand::Cancel)
}
fn lifecycle(
    state: AppState,
    headers: HeaderMap,
    id: String,
    command: GoalLifecycleCommand,
) -> Result<Json<GoalView>, (StatusCode, String)> {
    if !authenticated(&headers, &state) {
        return Err((StatusCode::UNAUTHORIZED, "OWNER_AUTH_REQUIRED".into()));
    }
    state
        .engine
        .transition(&id, command)
        .map(Json)
        .map_err(error)
}
async fn tick(
    State(state): State<AppState>,
    headers: HeaderMap,
) -> Result<Json<TickResult>, (StatusCode, String)> {
    if !authenticated(&headers, &state) {
        return Err((StatusCode::UNAUTHORIZED, "OWNER_AUTH_REQUIRED".into()));
    }
    state.engine.tick().map(Json).map_err(error)
}
async fn capabilities() -> Json<serde_json::Value> {
    Json(
        serde_json::json!({"contract_version":1,"goal_engine":{"state":"available","implementation":"rust-sqlite"},"scheduler":{"state":"available","implementation":"tokio-triggered-tick"},"policy_broker":{"state":"separate_rust_process","transport":"private-unix-socket","privileged_system_actions":"not_enabled"},"llm":{"state":"blocked","reason":"llama.cpp endpoint/model not configured"},"wayland":{"state":"not_tested","reason":"Linux GTK shell requires target runtime verification"},"nvidia_gpu":{"state":"blocked","reason":"no authorized Linux target"},"comfyui":{"state":"blocked","reason":"no authorized Linux target"},"system_image":{"state":"not_built","reason":"archiso pipeline not yet run"}}),
    )
}

pub fn open_engine(
    database: &str,
    owner_token: &str,
    policy_socket: &str,
) -> Result<GoalEngine, EngineError> {
    GoalEngine::open(database, UnixPolicyClient::new(policy_socket, owner_token))
}

#[cfg(test)]
mod tests {
    use super::*;
    use axum::{body::Body, http::Request};
    use luna_contracts::{CreateGoal, Metrics, PrivacyClass, ProvenanceInput};
    use luna_policy_broker::PolicyBroker;
    use tower::ServiceExt;
    use uuid::Uuid;

    #[tokio::test]
    async fn owner_can_add_dependency_and_non_owner_is_rejected() {
        let database =
            std::env::temp_dir().join(format!("luna-core-api-{}.sqlite", Uuid::new_v4()));
        let owner_token = "owner-test-token";
        let engine = GoalEngine::open(&database, PolicyBroker::new(owner_token, None)).unwrap();
        let body = CreateGoal {
            action: "local.checksum".into(),
            metrics: Metrics::default(),
            privacy_class: PrivacyClass::General,
            expected_cost_minor: None,
            provenance: ProvenanceInput::default(),
        };
        let first = engine
            .create(&body, owner_token, "api-prerequisite")
            .unwrap();
        let second = engine.create(&body, owner_token, "api-dependent").unwrap();
        let app = router(engine, owner_token.into());
        let path = format!("/v1/goals/{}/dependencies", second.id);
        let request = Request::builder()
            .method("POST")
            .uri(&path)
            .header("authorization", format!("Bearer {owner_token}"))
            .header("content-type", "application/json")
            .body(Body::from(format!(
                r#"{{"depends_on_goal_id":"{}"}}"#,
                first.id
            )))
            .unwrap();
        let response = app.clone().oneshot(request).await.unwrap();
        assert_eq!(response.status(), StatusCode::NO_CONTENT);

        let request = Request::builder()
            .method("POST")
            .uri(&path)
            .header("content-type", "application/json")
            .body(Body::from(format!(
                r#"{{"depends_on_goal_id":"{}"}}"#,
                first.id
            )))
            .unwrap();
        let response = app.oneshot(request).await.unwrap();
        assert_eq!(response.status(), StatusCode::UNAUTHORIZED);
        let _ = std::fs::remove_file(database);
    }

    #[tokio::test]
    async fn goal_lifecycle_routes_require_owner_and_persist_transitions() {
        let database =
            std::env::temp_dir().join(format!("luna-core-lifecycle-{}.sqlite", Uuid::new_v4()));
        let owner_token = "owner-test-token";
        let engine = GoalEngine::open(&database, PolicyBroker::new(owner_token, None)).unwrap();
        let goal = engine
            .create(
                &CreateGoal {
                    action: "local.checksum".into(),
                    metrics: Metrics::default(),
                    privacy_class: PrivacyClass::General,
                    expected_cost_minor: None,
                    provenance: ProvenanceInput::default(),
                },
                owner_token,
                "lifecycle-api",
            )
            .unwrap();
        let app = router(engine, owner_token.into());
        let list_unauthorized = Request::builder()
            .uri("/v1/goals")
            .body(Body::empty())
            .unwrap();
        assert_eq!(
            app.clone()
                .oneshot(list_unauthorized)
                .await
                .unwrap()
                .status(),
            StatusCode::UNAUTHORIZED
        );
        let list_owner = Request::builder()
            .uri("/v1/goals")
            .header("authorization", format!("Bearer {owner_token}"))
            .body(Body::empty())
            .unwrap();
        let listed = app.clone().oneshot(list_owner).await.unwrap();
        assert_eq!(listed.status(), StatusCode::OK);
        let listed: Vec<GoalView> = serde_json::from_slice(
            &axum::body::to_bytes(listed.into_body(), usize::MAX)
                .await
                .unwrap(),
        )
        .unwrap();
        assert_eq!(listed.len(), 1);
        let events_unauthorized = Request::builder()
            .uri(format!("/v1/goals/{}/events", goal.id))
            .body(Body::empty())
            .unwrap();
        assert_eq!(
            app.clone()
                .oneshot(events_unauthorized)
                .await
                .unwrap()
                .status(),
            StatusCode::UNAUTHORIZED
        );
        let events_owner = Request::builder()
            .uri(format!("/v1/goals/{}/events", goal.id))
            .header("authorization", format!("Bearer {owner_token}"))
            .body(Body::empty())
            .unwrap();
        let events = app.clone().oneshot(events_owner).await.unwrap();
        assert_eq!(events.status(), StatusCode::OK);
        let events: Vec<GoalEventView> = serde_json::from_slice(
            &axum::body::to_bytes(events.into_body(), usize::MAX)
                .await
                .unwrap(),
        )
        .unwrap();
        assert_eq!(events.len(), 1);
        assert_eq!(events[0].kind, "goal.created");
        let path = format!("/v1/goals/{}/pause", goal.id);
        let unauthorized = Request::builder()
            .method("POST")
            .uri(&path)
            .body(Body::empty())
            .unwrap();
        assert_eq!(
            app.clone().oneshot(unauthorized).await.unwrap().status(),
            StatusCode::UNAUTHORIZED
        );

        let owner_request = |action: &str| {
            Request::builder()
                .method("POST")
                .uri(format!("/v1/goals/{}/{action}", goal.id))
                .header("authorization", format!("Bearer {owner_token}"))
                .body(Body::empty())
                .unwrap()
        };
        let paused = app.clone().oneshot(owner_request("pause")).await.unwrap();
        assert_eq!(paused.status(), StatusCode::OK);
        let paused: GoalView = serde_json::from_slice(
            &axum::body::to_bytes(paused.into_body(), usize::MAX)
                .await
                .unwrap(),
        )
        .unwrap();
        assert_eq!(paused.status, "paused");

        let resumed = app.clone().oneshot(owner_request("resume")).await.unwrap();
        let resumed: GoalView = serde_json::from_slice(
            &axum::body::to_bytes(resumed.into_body(), usize::MAX)
                .await
                .unwrap(),
        )
        .unwrap();
        assert_eq!(resumed.status, "active");
        let cancelled = app.oneshot(owner_request("cancel")).await.unwrap();
        let cancelled: GoalView = serde_json::from_slice(
            &axum::body::to_bytes(cancelled.into_body(), usize::MAX)
                .await
                .unwrap(),
        )
        .unwrap();
        assert_eq!(cancelled.status, "cancelled");
        let _ = std::fs::remove_file(database);
    }

    #[tokio::test]
    async fn goal_plan_routes_require_owner_and_roundtrip_hierarchy() {
        let database =
            std::env::temp_dir().join(format!("luna-core-plan-{}.sqlite", Uuid::new_v4()));
        let owner_token = "owner-plan-token";
        let engine = GoalEngine::open(&database, PolicyBroker::new(owner_token, None)).unwrap();
        let create = |key: &str| {
            engine
                .create(
                    &CreateGoal {
                        action: "local.checksum".into(),
                        metrics: Metrics::default(),
                        privacy_class: PrivacyClass::General,
                        expected_cost_minor: None,
                        provenance: ProvenanceInput::default(),
                    },
                    owner_token,
                    key,
                )
                .unwrap()
        };
        let parent = create("plan-parent");
        let child = create("plan-child");
        let app = router(engine, owner_token.into());
        let uri = format!("/v1/goals/{}/plan", child.id);
        let unauthorized = Request::builder()
            .method("PUT")
            .uri(&uri)
            .header("content-type", "application/json")
            .body(Body::from(format!(
                r#"{{"parent_goal_id":"{}","dependency_goal_ids":[]}}"#,
                parent.id
            )))
            .unwrap();
        assert_eq!(
            app.clone().oneshot(unauthorized).await.unwrap().status(),
            StatusCode::UNAUTHORIZED
        );
        let owner = Request::builder()
            .method("PUT")
            .uri(&uri)
            .header("authorization", format!("Bearer {owner_token}"))
            .header("content-type", "application/json")
            .body(Body::from(format!(
                r#"{{"parent_goal_id":"{}","dependency_goal_ids":[]}}"#,
                parent.id
            )))
            .unwrap();
        let response = app.clone().oneshot(owner).await.unwrap();
        assert_eq!(response.status(), StatusCode::OK);
        let plan: GoalPlanView = serde_json::from_slice(
            &axum::body::to_bytes(response.into_body(), usize::MAX)
                .await
                .unwrap(),
        )
        .unwrap();
        assert_eq!(plan.parent_goal_id.as_deref(), Some(parent.id.as_str()));
        let get = Request::builder()
            .uri(&uri)
            .header("authorization", format!("Bearer {owner_token}"))
            .body(Body::empty())
            .unwrap();
        assert_eq!(
            app.clone().oneshot(get).await.unwrap().status(),
            StatusCode::OK
        );
        let _ = std::fs::remove_file(database);
    }

    #[tokio::test]
    async fn evidence_route_requires_owner_and_returns_verified_artifact() {
        let database =
            std::env::temp_dir().join(format!("luna-core-evidence-{}.sqlite", Uuid::new_v4()));
        let owner_token = "owner-evidence-token";
        let engine = GoalEngine::open(&database, PolicyBroker::new(owner_token, None)).unwrap();
        let goal = engine
            .create(
                &CreateGoal {
                    action: "local.checksum".into(),
                    metrics: Metrics::default(),
                    privacy_class: PrivacyClass::General,
                    expected_cost_minor: None,
                    provenance: ProvenanceInput::default(),
                },
                owner_token,
                "evidence-route",
            )
            .unwrap();
        engine.tick().unwrap();
        let app = router(engine, owner_token.into());
        let uri = format!("/v1/goals/{}/evidence", goal.id);
        let unauthorized = Request::builder().uri(&uri).body(Body::empty()).unwrap();
        assert_eq!(
            app.clone().oneshot(unauthorized).await.unwrap().status(),
            StatusCode::UNAUTHORIZED
        );
        let owner = Request::builder()
            .uri(&uri)
            .header("authorization", format!("Bearer {owner_token}"))
            .body(Body::empty())
            .unwrap();
        let response = app.oneshot(owner).await.unwrap();
        assert_eq!(response.status(), StatusCode::OK);
        let evidence: luna_contracts::Evidence = serde_json::from_slice(
            &axum::body::to_bytes(response.into_body(), usize::MAX)
                .await
                .unwrap(),
        )
        .unwrap();
        assert!(evidence.verified);
        assert_eq!(evidence.goal_id, goal.id);
        assert_eq!(evidence.byte_count, 28);
        let _ = std::fs::remove_file(database);
    }
}
