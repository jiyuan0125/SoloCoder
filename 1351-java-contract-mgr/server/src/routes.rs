use axum::{
    extract::{Path, State},
    http::StatusCode,
    Json,
    Router,
    routing::{get, post, put},
};
use uuid::Uuid;
use contract_core::{
    Contract, ContractError, CreateContractRequest, RenewalRequest,
    PriceAdjustmentRequest, ApprovalRequest, ReminderInfo, SystemConfig,
};
use crate::AppState;

pub fn routes() -> Router<AppState> {
    Router::new()
        .route("/api/contracts", post(create_contract))
        .route("/api/contracts", get(list_contracts))
        .route("/api/contracts/:id", get(get_contract))
        .route("/api/contracts/:id/activate", put(activate_contract))
        .route("/api/contracts/:id/renew", put(request_renewal))
        .route("/api/contracts/:id/terminate", put(terminate_contract))
        .route("/api/contracts/:id/close", put(close_contract))
        .route("/api/contracts/:id/price", put(adjust_price))
        .route("/api/approvals/:id/approve-renewal", put(approve_renewal))
        .route("/api/approvals/:id/reject-renewal", put(reject_renewal))
        .route("/api/approvals/:id/approve-price", put(approve_price))
        .route("/api/approvals/:id/reject-price", put(reject_price))
        .route("/api/approvals/:id", get(get_approval))
        .route("/api/reminders", get(get_reminders))
        .route("/api/config", get(get_config))
        .route("/api/config", put(update_config))
}

async fn create_contract(
    State(state): State<AppState>,
    Json(req): Json<CreateContractRequest>,
) -> Result<(StatusCode, Json<Contract>), (StatusCode, Json<serde_json::Value>)> {
    state.service.create_contract(req)
        .map(|c| (StatusCode::CREATED, Json(c)))
        .map_err(handle_error)
}

async fn list_contracts(
    State(state): State<AppState>,
) -> Json<Vec<Contract>> {
    Json(state.service.get_all_contracts())
}

async fn get_contract(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<Contract>, (StatusCode, Json<serde_json::Value>)> {
    state.service.get_contract(&id)
        .map(Json)
        .map_err(handle_error)
}

async fn activate_contract(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<Contract>, (StatusCode, Json<serde_json::Value>)> {
    state.service.activate_contract(&id)
        .map(Json)
        .map_err(handle_error)
}

#[derive(serde::Deserialize)]
struct RenewalRequestWrapper {
    new_amount: f64,
    end_date: chrono::DateTime<chrono::Utc>,
}

async fn request_renewal(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(req): Json<RenewalRequestWrapper>,
) -> Result<Json<ApprovalRequest>, (StatusCode, Json<serde_json::Value>)> {
    state.service.request_renewal(&id, RenewalRequest {
        new_amount: req.new_amount,
        end_date: req.end_date,
    })
        .map(Json)
        .map_err(handle_error)
}

#[derive(serde::Deserialize)]
struct ApproveRenewalRequest {
    end_date: chrono::DateTime<chrono::Utc>,
}

async fn approve_renewal(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(req): Json<ApproveRenewalRequest>,
) -> Result<Json<Contract>, (StatusCode, Json<serde_json::Value>)> {
    state.service.approve_renewal(&id, req.end_date)
        .map(Json)
        .map_err(handle_error)
}

#[derive(serde::Deserialize)]
struct RejectRequest {
    reason: String,
}

async fn reject_renewal(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(req): Json<RejectRequest>,
) -> Result<StatusCode, (StatusCode, Json<serde_json::Value>)> {
    state.service.reject_renewal(&id, req.reason)
        .map(|_| StatusCode::OK)
        .map_err(handle_error)
}

async fn adjust_price(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(req): Json<PriceAdjustmentRequest>,
) -> Result<Json<Option<ApprovalRequest>>, (StatusCode, Json<serde_json::Value>)> {
    state.service.adjust_price(&id, req)
        .map(Json)
        .map_err(handle_error)
}

async fn approve_price(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<Contract>, (StatusCode, Json<serde_json::Value>)> {
    state.service.approve_price_adjustment(&id)
        .map(Json)
        .map_err(handle_error)
}

async fn reject_price(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(req): Json<RejectRequest>,
) -> Result<StatusCode, (StatusCode, Json<serde_json::Value>)> {
    state.service.reject_price_adjustment(&id, req.reason)
        .map(|_| StatusCode::OK)
        .map_err(handle_error)
}

async fn terminate_contract(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<Contract>, (StatusCode, Json<serde_json::Value>)> {
    state.service.terminate_contract(&id)
        .map(Json)
        .map_err(handle_error)
}

async fn close_contract(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<Contract>, (StatusCode, Json<serde_json::Value>)> {
    state.service.close_contract(&id)
        .map(Json)
        .map_err(handle_error)
}

async fn get_reminders(
    State(state): State<AppState>,
) -> Json<Vec<ReminderInfo>> {
    Json(state.service.get_expiring_contracts_reminders())
}

async fn get_config(
    State(state): State<AppState>,
) -> Json<SystemConfig> {
    Json(state.service.get_config())
}

async fn update_config(
    State(state): State<AppState>,
    Json(config): Json<SystemConfig>,
) -> Json<SystemConfig> {
    state.service.update_config(config.clone());
    Json(config)
}

async fn get_approval(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<ApprovalRequest>, (StatusCode, Json<serde_json::Value>)> {
    state.service.get_approval(&id)
        .map(Json)
        .map_err(handle_error)
}

fn handle_error(err: ContractError) -> (StatusCode, Json<serde_json::Value>) {
    let (status, message) = match err {
        ContractError::NotFound(_) => (StatusCode::NOT_FOUND, err.to_string()),
        ContractError::InvalidStateTransition { .. } => (StatusCode::BAD_REQUEST, err.to_string()),
        ContractError::InvalidOperation(_) => (StatusCode::BAD_REQUEST, err.to_string()),
        ContractError::Validation(_) => (StatusCode::BAD_REQUEST, err.to_string()),
        ContractError::HasActiveSubContracts => (StatusCode::BAD_REQUEST, err.to_string()),
        ContractError::PriceAdjustmentPending => (StatusCode::BAD_REQUEST, err.to_string()),
        ContractError::ApprovalNotFound(_) => (StatusCode::NOT_FOUND, err.to_string()),
    };
    (status, Json(serde_json::json!({ "error": message })))
}
