use axum::extract::{Path, State};
use axum::http::StatusCode;
use axum::response::IntoResponse;
use axum::Json;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

use crate::errors::ConfigError;
use crate::models::{ApproveRequest, RollbackRequest, SubmitRequest};
use crate::service::ConfigService;

#[derive(Clone)]
pub struct AppState {
    pub service: ConfigService,
}

#[derive(Serialize)]
struct ErrorResponse {
    error: String,
    message: String,
}

impl IntoResponse for ConfigError {
    fn into_response(self) -> axum::response::Response {
        let (status, error_type) = match &self {
            ConfigError::NamespaceNotFound(_) => (StatusCode::NOT_FOUND, "namespace_not_found"),
            ConfigError::ChangeRequestNotFound(_) => (StatusCode::NOT_FOUND, "change_request_not_found"),
            ConfigError::InvalidStatusTransition(_) => (StatusCode::BAD_REQUEST, "invalid_status_transition"),
            ConfigError::PermissionDenied(_) => (StatusCode::FORBIDDEN, "permission_denied"),
            ConfigError::TypeValidationError(_) => (StatusCode::UNPROCESSABLE_ENTITY, "type_validation_error"),
            ConfigError::NoApprovedVersion => (StatusCode::BAD_REQUEST, "no_approved_version"),
            ConfigError::Internal(_) => (StatusCode::INTERNAL_SERVER_ERROR, "internal_error"),
        };

        let body = Json(ErrorResponse {
            error: error_type.to_string(),
            message: self.to_string(),
        });

        (status, body).into_response()
    }
}

#[derive(Deserialize)]
pub struct NamespaceQuery {
    namespace: Option<String>,
}

pub async fn submit_change(
    State(state): State<AppState>,
    Json(req): Json<SubmitRequest>,
) -> Result<impl IntoResponse, ConfigError> {
    let change = state.service.submit_change(req)?;
    Ok((StatusCode::CREATED, Json(change)))
}

pub async fn approve_change(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(req): Json<ApproveRequest>,
) -> Result<impl IntoResponse, ConfigError> {
    let change = state.service.approve_change(&id, &req.approver)?;
    Ok(Json(change))
}

pub async fn reject_change(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(req): Json<ApproveRequest>,
) -> Result<impl IntoResponse, ConfigError> {
    let change = state.service.reject_change(&id, &req.approver)?;
    Ok(Json(change))
}

pub async fn start_canary(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(req): Json<RollbackRequest>,
) -> Result<impl IntoResponse, ConfigError> {
    let change = state.service.start_canary(&id, &req.operator)?;
    Ok(Json(change))
}

pub async fn deploy_full(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(req): Json<RollbackRequest>,
) -> Result<impl IntoResponse, ConfigError> {
    let change = state.service.deploy_full(&id, &req.operator)?;
    Ok(Json(change))
}

pub async fn rollback_change(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(req): Json<RollbackRequest>,
) -> Result<impl IntoResponse, ConfigError> {
    let (change, diffs) = state.service.rollback_change(&id, &req.operator)?;
    #[derive(Serialize)]
    struct RollbackResponse {
        change_request: crate::models::ChangeRequest,
        rollback_diffs: Vec<crate::models::DiffItem>,
    }
    Ok(Json(RollbackResponse {
        change_request: change,
        rollback_diffs: diffs,
    }))
}

pub async fn get_change_request(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<impl IntoResponse, ConfigError> {
    match state.service.get_change_request(&id) {
        Some(change) => Ok(Json(change)),
        None => Err(ConfigError::ChangeRequestNotFound(id.to_string())),
    }
}

pub async fn get_configs(
    State(state): State<AppState>,
    Path(namespace): Path<String>,
) -> Result<impl IntoResponse, ConfigError> {
    let configs = state.service.get_configs(&namespace)?;
    Ok(Json(configs))
}

pub async fn get_audit_logs(
    State(state): State<AppState>,
    query: axum::extract::Query<NamespaceQuery>,
) -> impl IntoResponse {
    let logs = state.service.get_audit_logs(query.namespace.as_deref());
    Json(logs)
}

pub async fn health_check() -> impl IntoResponse {
    #[derive(Serialize)]
    struct HealthResponse {
        status: &'static str,
    }
    Json(HealthResponse { status: "ok" })
}
