use axum::{
    Json, Router,
    extract::{Path, State},
    http::{HeaderMap, StatusCode},
    routing::{get, post},
};
use luna_contracts::{CreateGoal, GoalView, TickResult};
use luna_goal_engine::{EngineError, GoalEngine};
use luna_policy_broker::PolicyBroker;
use std::sync::Arc;

#[derive(Clone)]
struct AppState {
    engine: GoalEngine,
    owner_token: Arc<String>,
}

pub fn router(engine: GoalEngine, owner_token: String) -> Router {
    Router::new()
        .route("/v1/goals", post(create_goal))
        .route("/v1/goals/{id}", get(get_goal))
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
        serde_json::json!({"contract_version":1,"goal_engine":{"state":"available","implementation":"rust-sqlite"},"scheduler":{"state":"available","implementation":"tokio-triggered-tick"},"policy_broker":{"state":"available","implementation":"rust-deny-by-default-library","separate_service":"not_yet_connected"},"llm":{"state":"blocked","reason":"llama.cpp endpoint/model not configured"},"wayland":{"state":"not_tested","reason":"Linux GTK shell requires target runtime verification"},"nvidia_gpu":{"state":"blocked","reason":"no authorized Linux target"},"comfyui":{"state":"blocked","reason":"no authorized Linux target"},"system_image":{"state":"not_built","reason":"archiso pipeline not yet run"}}),
    )
}

pub fn open_engine(database: &str, owner_token: &str) -> Result<GoalEngine, EngineError> {
    GoalEngine::open(database, PolicyBroker::new(owner_token, None))
}
