use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use serde::Serialize;
use std::sync::Arc;
use uuid::Uuid;

use recruit_core::*;

#[derive(Clone)]
struct AppState {
    service: Arc<RecruitmentService>,
}

#[derive(Debug, Serialize)]
struct ErrorResponse {
    error: String,
}

fn map_error(err: RecruitError) -> (StatusCode, Json<ErrorResponse>) {
    let status = match &err {
        RecruitError::CandidateNotFound(_) => StatusCode::NOT_FOUND,
        RecruitError::InterviewNotFound(_) => StatusCode::NOT_FOUND,
        RecruitError::OfferNotFound(_) => StatusCode::NOT_FOUND,
        RecruitError::InvalidStateTransition(_) => StatusCode::BAD_REQUEST,
        RecruitError::InterviewerTimeConflict(_) => StatusCode::CONFLICT,
        RecruitError::CandidateTimeConflict => StatusCode::CONFLICT,
        RecruitError::NoBufferTime => StatusCode::CONFLICT,
        RecruitError::InsufficientApprovalLevel(_) => StatusCode::FORBIDDEN,
        RecruitError::OfferNotPending => StatusCode::BAD_REQUEST,
        RecruitError::CandidatePending => StatusCode::BAD_REQUEST,
        RecruitError::CandidateExpired => StatusCode::BAD_REQUEST,
        RecruitError::InvalidInput(_) => StatusCode::BAD_REQUEST,
    };
    (status, Json(ErrorResponse { error: err.to_string() }))
}

async fn health_check() -> impl IntoResponse {
    (StatusCode::OK, Json(serde_json::json!({"status": "ok"})))
}

async fn create_candidate(
    State(state): State<AppState>,
    Json(req): Json<CreateCandidateRequest>,
) -> impl IntoResponse {
    match state.service.create_candidate(req) {
        Ok(candidate) => (StatusCode::CREATED, Json(candidate)).into_response(),
        Err(e) => map_error(e).into_response(),
    }
}

async fn list_candidates(State(state): State<AppState>) -> impl IntoResponse {
    match state.service.list_candidates() {
        Ok(candidates) => (StatusCode::OK, Json(candidates)).into_response(),
        Err(e) => map_error(e).into_response(),
    }
}

async fn get_candidate(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.get_candidate(id) {
        Ok(candidate) => (StatusCode::OK, Json(candidate)).into_response(),
        Err(e) => map_error(e).into_response(),
    }
}

async fn initial_screening(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(req): Json<InitialScreeningRequest>,
) -> impl IntoResponse {
    match state.service.initial_screening(id, req) {
        Ok(candidate) => (StatusCode::OK, Json(candidate)).into_response(),
        Err(e) => map_error(e).into_response(),
    }
}

async fn second_screening(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(req): Json<SecondScreeningRequest>,
) -> impl IntoResponse {
    match state.service.second_screening(id, req) {
        Ok(candidate) => (StatusCode::OK, Json(candidate)).into_response(),
        Err(e) => map_error(e).into_response(),
    }
}

async fn confirm_expired_candidate(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(req): Json<ConfirmExpiredCandidateRequest>,
) -> impl IntoResponse {
    match state.service.confirm_expired_candidate(id, req) {
        Ok(candidate) => (StatusCode::OK, Json(candidate)).into_response(),
        Err(e) => map_error(e).into_response(),
    }
}

async fn check_expired_pending(State(state): State<AppState>) -> impl IntoResponse {
    match state.service.check_and_update_expired_pending() {
        Ok(expired) => (
            StatusCode::OK,
            Json(serde_json::json!({"expired_count": expired.len(), "expired_ids": expired})),
        )
            .into_response(),
        Err(e) => map_error(e).into_response(),
    }
}

async fn schedule_interview(
    State(state): State<AppState>,
    Path(candidate_id): Path<Uuid>,
    Json(req): Json<ScheduleInterviewRequest>,
) -> impl IntoResponse {
    match state.service.schedule_interview(candidate_id, req) {
        Ok(interview) => (StatusCode::CREATED, Json(interview)).into_response(),
        Err(e) => map_error(e).into_response(),
    }
}

async fn list_interviews(State(state): State<AppState>) -> impl IntoResponse {
    match state.service.list_interviews(None) {
        Ok(interviews) => (StatusCode::OK, Json(interviews)).into_response(),
        Err(e) => map_error(e).into_response(),
    }
}

async fn get_interview(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.get_interview(id) {
        Ok(interview) => (StatusCode::OK, Json(interview)).into_response(),
        Err(e) => map_error(e).into_response(),
    }
}

async fn submit_interview_result(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(req): Json<InterviewResultRequest>,
) -> impl IntoResponse {
    match state.service.submit_interview_result(id, req) {
        Ok(interview) => (StatusCode::OK, Json(interview)).into_response(),
        Err(e) => map_error(e).into_response(),
    }
}

async fn create_offer(
    State(state): State<AppState>,
    Path(candidate_id): Path<Uuid>,
    Json(req): Json<CreateOfferRequest>,
) -> impl IntoResponse {
    match state.service.create_offer(candidate_id, req) {
        Ok(offer) => (StatusCode::CREATED, Json(offer)).into_response(),
        Err(e) => map_error(e).into_response(),
    }
}

async fn list_offers(State(state): State<AppState>) -> impl IntoResponse {
    match state.service.list_offers(None) {
        Ok(offers) => (StatusCode::OK, Json(offers)).into_response(),
        Err(e) => map_error(e).into_response(),
    }
}

async fn get_offer(State(state): State<AppState>, Path(id): Path<Uuid>) -> impl IntoResponse {
    match state.service.get_offer(id) {
        Ok(offer) => (StatusCode::OK, Json(offer)).into_response(),
        Err(e) => map_error(e).into_response(),
    }
}

async fn submit_offer_for_approval(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.submit_offer_for_approval(id) {
        Ok(offer) => (StatusCode::OK, Json(offer)).into_response(),
        Err(e) => map_error(e).into_response(),
    }
}

async fn approve_offer(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(req): Json<ApproveOfferRequest>,
) -> impl IntoResponse {
    match state.service.approve_offer(id, req) {
        Ok(offer) => (StatusCode::OK, Json(offer)).into_response(),
        Err(e) => map_error(e).into_response(),
    }
}

async fn reject_offer(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(req): Json<ApproveOfferRequest>,
) -> impl IntoResponse {
    match state.service.reject_offer(id, req) {
        Ok(offer) => (StatusCode::OK, Json(offer)).into_response(),
        Err(e) => map_error(e).into_response(),
    }
}

async fn revise_offer(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(req): Json<ReviseOfferRequest>,
) -> impl IntoResponse {
    match state.service.revise_offer(id, req) {
        Ok(offer) => (StatusCode::OK, Json(offer)).into_response(),
        Err(e) => map_error(e).into_response(),
    }
}

pub fn router() -> Router {
    let state = AppState {
        service: Arc::new(RecruitmentService::new()),
    };

    Router::new()
        .route("/health", get(health_check))
        .route("/candidates", post(create_candidate).get(list_candidates))
        .route("/candidates/:id", get(get_candidate))
        .route("/candidates/:id/initial-screening", post(initial_screening))
        .route("/candidates/:id/second-screening", post(second_screening))
        .route(
            "/candidates/:id/confirm-expired",
            post(confirm_expired_candidate),
        )
        .route("/candidates/check-expired", post(check_expired_pending))
        .route("/candidates/:id/interviews", post(schedule_interview))
        .route("/interviews", get(list_interviews))
        .route("/interviews/:id", get(get_interview))
        .route("/interviews/:id/result", post(submit_interview_result))
        .route("/candidates/:id/offers", post(create_offer))
        .route("/offers", get(list_offers))
        .route("/offers/:id", get(get_offer))
        .route("/offers/:id/submit", post(submit_offer_for_approval))
        .route("/offers/:id/approve", post(approve_offer))
        .route("/offers/:id/reject", post(reject_offer))
        .route("/offers/:id/revise", post(revise_offer))
        .with_state(state)
}
