use axum::{
    extract::{Path, State},
    http::StatusCode,
    Json,
    Router,
    routing::{get, post, put},
};
use uuid::Uuid;

use claim_audit_core::{
    ApiResponse, ApproveRequest, Claim, ClaimService, CreateClaimRequest, FinalRejectRequest,
    ReturnRequest, ResubmitRequest, UpdateClaimRequest,
};

pub fn claim_routes(service: ClaimService) -> Router {
    Router::new()
        .route("/claims", post(create_claim))
        .route("/claims", get(list_claims))
        .route("/claims/:id", get(get_claim))
        .route("/claims/:id", put(update_claim))
        .route("/claims/:id/approve", post(approve_claim))
        .route("/claims/:id/return", post(return_claim))
        .route("/claims/:id/final-reject", post(final_reject_claim))
        .route("/claims/:id/resubmit", post(resubmit_claim))
        .with_state(service)
}

async fn create_claim(
    State(service): State<ClaimService>,
    Json(req): Json<CreateClaimRequest>,
) -> (StatusCode, Json<ApiResponse<Claim>>) {
    let response = service.create_claim(req);
    if response.success {
        (StatusCode::CREATED, Json(response))
    } else {
        (StatusCode::BAD_REQUEST, Json(response))
    }
}

async fn list_claims(State(service): State<ClaimService>) -> Json<ApiResponse<Vec<Claim>>> {
    Json(service.list_claims())
}

async fn get_claim(
    State(service): State<ClaimService>,
    Path(id): Path<Uuid>,
) -> (StatusCode, Json<ApiResponse<Claim>>) {
    let response = service.get_claim(id);
    if response.success {
        (StatusCode::OK, Json(response))
    } else {
        (StatusCode::NOT_FOUND, Json(response))
    }
}

async fn update_claim(
    State(service): State<ClaimService>,
    Path(id): Path<Uuid>,
    Json(req): Json<UpdateClaimRequest>,
) -> (StatusCode, Json<ApiResponse<Claim>>) {
    let response = service.update_claim(id, req);
    if response.success {
        (StatusCode::OK, Json(response))
    } else {
        (StatusCode::BAD_REQUEST, Json(response))
    }
}

async fn approve_claim(
    State(service): State<ClaimService>,
    Path(id): Path<Uuid>,
    Json(req): Json<ApproveRequest>,
) -> (StatusCode, Json<ApiResponse<Claim>>) {
    let response = service.approve(id, req.operator);
    if response.success {
        (StatusCode::OK, Json(response))
    } else {
        (StatusCode::BAD_REQUEST, Json(response))
    }
}

async fn return_claim(
    State(service): State<ClaimService>,
    Path(id): Path<Uuid>,
    Json(req): Json<ReturnRequest>,
) -> (StatusCode, Json<ApiResponse<Claim>>) {
    let response = service.return_to(id, req);
    if response.success {
        (StatusCode::OK, Json(response))
    } else {
        (StatusCode::BAD_REQUEST, Json(response))
    }
}

async fn final_reject_claim(
    State(service): State<ClaimService>,
    Path(id): Path<Uuid>,
    Json(req): Json<FinalRejectRequest>,
) -> (StatusCode, Json<ApiResponse<Claim>>) {
    let response = service.final_reject(id, req);
    if response.success {
        (StatusCode::OK, Json(response))
    } else {
        (StatusCode::BAD_REQUEST, Json(response))
    }
}

async fn resubmit_claim(
    State(service): State<ClaimService>,
    Path(id): Path<Uuid>,
    Json(req): Json<ResubmitRequest>,
) -> (StatusCode, Json<ApiResponse<Claim>>) {
    let response = service.resubmit(id, req);
    if response.success {
        (StatusCode::OK, Json(response))
    } else {
        (StatusCode::BAD_REQUEST, Json(response))
    }
}
