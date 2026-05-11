use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::{IntoResponse, Json},
    routing::{delete, get, post, put},
    Router,
};
use serde::{Deserialize, Serialize};

use archive_core::*;

use crate::state::AppState;

#[derive(Debug, Serialize)]
struct ApiResponse<T> {
    success: bool,
    data: Option<T>,
    error: Option<String>,
}

impl<T> ApiResponse<T> {
    fn success(data: T) -> Self {
        Self {
            success: true,
            data: Some(data),
            error: None,
        }
    }

    fn error(msg: String) -> Self {
        Self {
            success: false,
            data: None,
            error: Some(msg),
        }
    }
}

struct ApiError(ArchiveError);

impl IntoResponse for ApiError {
    fn into_response(self) -> axum::response::Response {
        let status = match &self.0 {
            ArchiveError::ArchiveNotFound(_) => StatusCode::NOT_FOUND,
            ArchiveError::UserNotFound(_) => StatusCode::NOT_FOUND,
            ArchiveError::BorrowNotFound(_) => StatusCode::NOT_FOUND,
            ArchiveError::ReservationNotFound(_) => StatusCode::NOT_FOUND,
            ArchiveError::ApprovalNotFound => StatusCode::NOT_FOUND,
            ArchiveError::ArchiveAlreadyBorrowed => StatusCode::CONFLICT,
            ArchiveError::ArchiveAlreadyReserved => StatusCode::CONFLICT,
            ArchiveError::UserAlreadyReserved => StatusCode::CONFLICT,
            ArchiveError::UserAlreadyBorrowed => StatusCode::CONFLICT,
            ArchiveError::InsufficientPermission => StatusCode::FORBIDDEN,
            ArchiveError::UserSuspended => StatusCode::FORBIDDEN,
            ArchiveError::ConfidentialRenewalNotAllowed => StatusCode::FORBIDDEN,
            ArchiveError::RenewalLimitExceeded => StatusCode::BAD_REQUEST,
            ArchiveError::ReservationExpired => StatusCode::GONE,
            ArchiveError::ReservationNotClaimed => StatusCode::GONE,
            ArchiveError::ApprovalAlreadyCompleted => StatusCode::CONFLICT,
            ArchiveError::ApproverMismatch => StatusCode::FORBIDDEN,
            ArchiveError::ApprovalOrderError => StatusCode::BAD_REQUEST,
            _ => StatusCode::INTERNAL_SERVER_ERROR,
        };

        let body = Json(ApiResponse::<()>::error(self.0.to_string()));
        (status, body).into_response()
    }
}

#[derive(Debug, Deserialize)]
struct CreateUserRequest {
    id: String,
    name: String,
    email: String,
    department: String,
    role: UserRole,
}

#[derive(Debug, Deserialize)]
struct CreateArchiveRequest {
    id: String,
    title: String,
    description: String,
    classification: ClassificationLevel,
}

#[derive(Debug, Deserialize)]
struct BorrowRequest {
    user_id: String,
    archive_id: String,
}

#[derive(Debug, Deserialize)]
struct ApproveRequest {
    approver_id: String,
    step: u8,
    comment: Option<String>,
}

#[derive(Debug, Deserialize)]
struct ReturnRequest {
    is_damaged: bool,
    damage_description: Option<String>,
    compensation_amount: Option<f64>,
    checked_by: String,
}

#[derive(Debug, Deserialize)]
struct ReserveRequest {
    user_id: String,
    archive_id: String,
}

async fn list_users(State(state): State<AppState>) -> impl IntoResponse {
    let service = state.service.lock().await;
    let users = service.get_all_users();
    Json(ApiResponse::success(users))
}

async fn get_user(State(state): State<AppState>, Path(id): Path<String>) -> Result<impl IntoResponse, ApiError> {
    let service = state.service.lock().await;
    let user = service.get_user(&id).map_err(ApiError)?;
    Ok(Json(ApiResponse::success(user)))
}

async fn create_user(
    State(state): State<AppState>,
    Json(req): Json<CreateUserRequest>,
) -> impl IntoResponse {
    let service = state.service.lock().await;
    let user = User::new(&req.id, &req.name, &req.email, &req.department, req.role);
    service.create_user(user.clone());
    (StatusCode::CREATED, Json(ApiResponse::success(user)))
}

async fn list_archives(State(state): State<AppState>) -> impl IntoResponse {
    let service = state.service.lock().await;
    let archives = service.get_all_archives();
    Json(ApiResponse::success(archives))
}

async fn get_archive(State(state): State<AppState>, Path(id): Path<String>) -> Result<impl IntoResponse, ApiError> {
    let service = state.service.lock().await;
    let archive = service.get_archive(&id).map_err(ApiError)?;
    Ok(Json(ApiResponse::success(archive)))
}

async fn create_archive(
    State(state): State<AppState>,
    Json(req): Json<CreateArchiveRequest>,
) -> impl IntoResponse {
    let service = state.service.lock().await;
    let archive = Archive::new(&req.id, &req.title, &req.description, req.classification);
    service.create_archive(archive.clone());
    (StatusCode::CREATED, Json(ApiResponse::success(archive)))
}

async fn list_borrows(State(state): State<AppState>) -> impl IntoResponse {
    let service = state.service.lock().await;
    let borrows = service.get_all_borrows();
    Json(ApiResponse::success(borrows))
}

async fn get_borrow(State(state): State<AppState>, Path(id): Path<String>) -> Result<impl IntoResponse, ApiError> {
    let service = state.service.lock().await;
    let borrow = service.get_borrow(&id).map_err(ApiError)?;
    Ok(Json(ApiResponse::success(borrow)))
}

async fn request_borrow(
    State(state): State<AppState>,
    Json(req): Json<BorrowRequest>,
) -> Result<impl IntoResponse, ApiError> {
    let service = state.service.lock().await;
    let borrow = service
        .request_borrow(&req.user_id, &req.archive_id)
        .map_err(ApiError)?;
    Ok((StatusCode::CREATED, Json(ApiResponse::success(borrow))))
}

async fn renew_borrow(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<impl IntoResponse, ApiError> {
    let service = state.service.lock().await;
    let borrow = service.renew_borrow(&id).map_err(ApiError)?;
    Ok(Json(ApiResponse::success(borrow)))
}

async fn return_borrow(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(req): Json<ReturnRequest>,
) -> Result<impl IntoResponse, ApiError> {
    use chrono::Utc;
    let check_result = ReturnCheckResult {
        is_damaged: req.is_damaged,
        damage_description: req.damage_description,
        compensation_amount: req.compensation_amount,
        checked_by: req.checked_by,
        checked_at: Utc::now(),
    };

    let service = state.service.lock().await;
    let borrow = service.return_archive(&id, check_result).map_err(ApiError)?;
    Ok(Json(ApiResponse::success(borrow)))
}

async fn get_approval_process(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<impl IntoResponse, ApiError> {
    let service = state.service.lock().await;
    let process = service.get_approval_process(&id).map_err(ApiError)?;
    Ok(Json(ApiResponse::success(process)))
}

async fn approve_borrow(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(req): Json<ApproveRequest>,
) -> Result<impl IntoResponse, ApiError> {
    let service = state.service.lock().await;
    let borrow = service
        .approve_borrow(&id, &req.approver_id, req.step, req.comment)
        .map_err(ApiError)?;
    Ok(Json(ApiResponse::success(borrow)))
}

async fn reject_borrow(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(req): Json<ApproveRequest>,
) -> Result<impl IntoResponse, ApiError> {
    let service = state.service.lock().await;
    let borrow = service
        .reject_borrow(&id, &req.approver_id, req.step, req.comment)
        .map_err(ApiError)?;
    Ok(Json(ApiResponse::success(borrow)))
}

async fn list_reservations(State(state): State<AppState>) -> impl IntoResponse {
    let service = state.service.lock().await;
    let reservations = service.get_all_reservations();
    Json(ApiResponse::success(reservations))
}

async fn get_reservation(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<impl IntoResponse, ApiError> {
    let service = state.service.lock().await;
    let reservation = service.get_reservation(&id).map_err(ApiError)?;
    Ok(Json(ApiResponse::success(reservation)))
}

async fn reserve_archive(
    State(state): State<AppState>,
    Json(req): Json<ReserveRequest>,
) -> Result<impl IntoResponse, ApiError> {
    let service = state.service.lock().await;
    let reservation = service
        .reserve_archive(&req.user_id, &req.archive_id)
        .map_err(ApiError)?;
    Ok((StatusCode::CREATED, Json(ApiResponse::success(reservation))))
}

async fn claim_reservation(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<impl IntoResponse, ApiError> {
    let service = state.service.lock().await;
    let borrow = service.claim_reservation(&id).map_err(ApiError)?;
    Ok(Json(ApiResponse::success(borrow)))
}

async fn cancel_reservation(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<impl IntoResponse, ApiError> {
    let service = state.service.lock().await;
    let reservation = service.cancel_reservation(&id).map_err(ApiError)?;
    Ok(Json(ApiResponse::success(reservation)))
}

async fn get_due_date_reminders(State(state): State<AppState>) -> impl IntoResponse {
    let service = state.service.lock().await;
    let reminders = service.get_due_date_reminders();
    Json(ApiResponse::success(reminders))
}

async fn process_overdue(State(state): State<AppState>) -> impl IntoResponse {
    let service = state.service.lock().await;
    let overdue = service.process_overdue();
    Json(ApiResponse::success(overdue))
}

async fn check_expired_reservations(State(state): State<AppState>) -> impl IntoResponse {
    let service = state.service.lock().await;
    let expired = service.check_expired_reservations();
    Json(ApiResponse::success(expired))
}

pub fn create_router(state: AppState) -> Router {
    Router::new()
        .route("/users", get(list_users).post(create_user))
        .route("/users/:id", get(get_user))
        .route("/archives", get(list_archives).post(create_archive))
        .route("/archives/:id", get(get_archive))
        .route("/borrows", get(list_borrows).post(request_borrow))
        .route("/borrows/:id", get(get_borrow))
        .route("/borrows/:id/renew", put(renew_borrow))
        .route("/borrows/:id/return", post(return_borrow))
        .route("/approvals/:id", get(get_approval_process))
        .route("/approvals/:id/approve", post(approve_borrow))
        .route("/approvals/:id/reject", post(reject_borrow))
        .route("/reservations", get(list_reservations).post(reserve_archive))
        .route("/reservations/:id", get(get_reservation))
        .route("/reservations/:id/claim", post(claim_reservation))
        .route("/reservations/:id/cancel", delete(cancel_reservation))
        .route("/reminders/due", get(get_due_date_reminders))
        .route("/overdue/process", post(process_overdue))
        .route("/reservations/check-expired", post(check_expired_reservations))
        .with_state(state)
}
