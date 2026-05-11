use std::sync::Arc;
use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post, put, delete},
    Json, Router,
};
use serde::Deserialize;
use dorm_core::*;
use crate::state::AppState;

pub fn create_router(state: Arc<AppState>) -> Router {
    Router::new()
        .route("/api/buildings", get(list_buildings).post(create_building))
        .route("/api/buildings/:id", get(get_building))
        .route("/api/rooms", get(list_rooms).post(create_room))
        .route("/api/rooms/:id", get(get_room))
        .route("/api/rooms/:id/maintenance", put(set_room_maintenance))
        .route("/api/rooms/:id/hygiene", put(update_hygiene_score))
        .route("/api/beds", get(list_beds))
        .route("/api/beds/:id", get(get_bed))
        .route("/api/students", get(list_students).post(create_student))
        .route("/api/students/:id", get(get_student).delete(check_out_student))
        .route("/api/allocate", post(allocate_department))
        .route("/api/swap-requests", get(list_swap_requests).post(create_swap_request))
        .route("/api/swap-requests/:id", get(get_swap_request))
        .route("/api/swap-requests/:id/respond", put(respond_swap_request))
        .route("/api/stats/available-beds", get(get_available_beds_count))
        .with_state(state)
}

#[derive(Deserialize)]
pub struct BuildingQuery {
    pub building_id: Option<String>,
}

#[derive(Deserialize)]
pub struct StudentQuery {
    pub department: Option<String>,
    pub status: Option<String>,
}

#[derive(Deserialize)]
pub struct SwapRequestQuery {
    pub status: Option<String>,
}

pub async fn list_buildings(
    State(state): State<Arc<AppState>>,
) -> impl IntoResponse {
    let buildings = state.service.list_buildings();
    Json(buildings)
}

pub async fn create_building(
    State(state): State<Arc<AppState>>,
    Json(payload): Json<CreateBuildingRequest>,
) -> impl IntoResponse {
    let building = state.service.create_building(
        payload.name,
        payload.floor_count,
        payload.rooms_per_floor,
    );
    (StatusCode::CREATED, Json(building))
}

pub async fn get_building(
    State(state): State<Arc<AppState>>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match state.service.get_building(&id) {
        Ok(building) => (StatusCode::OK, Json(Some(building))).into_response(),
        Err(e) => (
            StatusCode::NOT_FOUND,
            Json(serde_json::json!({ "error": e.to_string() })),
        ).into_response(),
    }
}

pub async fn list_rooms(
    State(state): State<Arc<AppState>>,
    Query(query): Query<BuildingQuery>,
) -> impl IntoResponse {
    let rooms = state.service.list_rooms(query.building_id.as_deref());
    Json(rooms)
}

pub async fn create_room(
    State(state): State<Arc<AppState>>,
    Json(payload): Json<CreateRoomRequest>,
) -> impl IntoResponse {
    match state.service.create_room(
        payload.building_id,
        payload.floor_number,
        payload.room_number,
        payload.room_type,
    ) {
        Ok(room) => (StatusCode::CREATED, Json(room)).into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({ "error": e.to_string() })),
        ).into_response(),
    }
}

pub async fn get_room(
    State(state): State<Arc<AppState>>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match state.service.get_room(&id) {
        Ok(room) => (StatusCode::OK, Json(Some(room))).into_response(),
        Err(e) => (
            StatusCode::NOT_FOUND,
            Json(serde_json::json!({ "error": e.to_string() })),
        ).into_response(),
    }
}

pub async fn set_room_maintenance(
    State(state): State<Arc<AppState>>,
    Path(id): Path<String>,
    Json(payload): Json<UpdateRoomStatusRequest>,
) -> impl IntoResponse {
    let under_maintenance = payload.status == RoomStatus::UnderMaintenance;
    match state.service.set_room_maintenance(&id, under_maintenance) {
        Ok(room) => (StatusCode::OK, Json(room)).into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({ "error": e.to_string() })),
        ).into_response(),
    }
}

pub async fn update_hygiene_score(
    State(state): State<Arc<AppState>>,
    Path(id): Path<String>,
    Json(payload): Json<UpdateHygieneScoreRequest>,
) -> impl IntoResponse {
    match state.service.update_hygiene_score(&id, &payload.score) {
        Ok(room) => (StatusCode::OK, Json(room)).into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({ "error": e.to_string() })),
        ).into_response(),
    }
}

pub async fn list_beds(
    State(state): State<Arc<AppState>>,
    Query(query): Query<BuildingQuery>,
) -> impl IntoResponse {
    let beds = state.service.list_beds(query.building_id.as_deref());
    Json(beds)
}

pub async fn get_bed(
    State(state): State<Arc<AppState>>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match state.service.get_bed(&id) {
        Ok(bed) => (StatusCode::OK, Json(Some(bed))).into_response(),
        Err(e) => (
            StatusCode::NOT_FOUND,
            Json(serde_json::json!({ "error": e.to_string() })),
        ).into_response(),
    }
}

pub async fn list_students(
    State(state): State<Arc<AppState>>,
    Query(query): Query<StudentQuery>,
) -> impl IntoResponse {
    let status = query.status.as_ref().and_then(|s| match s.to_uppercase().as_str() {
        "UNASSIGNED" => Some(StudentStatus::Unassigned),
        "ASSIGNED" => Some(StudentStatus::Assigned),
        "PENDING_TRANSFER" => Some(StudentStatus::PendingTransfer),
        "WAITING_FOR_SWAP_CONFIRMATION" => Some(StudentStatus::WaitingForSwapConfirmation),
        _ => None,
    });
    
    let students = state.service.list_students(query.department.as_deref(), status);
    Json(students)
}

pub async fn create_student(
    State(state): State<Arc<AppState>>,
    Json(payload): Json<CreateStudentRequest>,
) -> impl IntoResponse {
    let student = state.service.create_student(
        payload.student_id,
        payload.name,
        payload.department,
    );
    (StatusCode::CREATED, Json(student))
}

pub async fn get_student(
    State(state): State<Arc<AppState>>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match state.service.get_student(&id) {
        Ok(student) => (StatusCode::OK, Json(Some(student))).into_response(),
        Err(e) => (
            StatusCode::NOT_FOUND,
            Json(serde_json::json!({ "error": e.to_string() })),
        ).into_response(),
    }
}

pub async fn check_out_student(
    State(state): State<Arc<AppState>>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match state.service.check_out(&id) {
        Ok(()) => (StatusCode::OK, Json(serde_json::json!({ "message": "Check out successful" }))).into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({ "error": e.to_string() })),
        ).into_response(),
    }
}

pub async fn allocate_department(
    State(state): State<Arc<AppState>>,
    Json(payload): Json<AllocateDepartmentRequest>,
) -> impl IntoResponse {
    match state.service.allocate_department(&payload.department) {
        Ok(result) => (StatusCode::OK, Json(result)).into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({ "error": e.to_string() })),
        ).into_response(),
    }
}

pub async fn list_swap_requests(
    State(state): State<Arc<AppState>>,
    Query(query): Query<SwapRequestQuery>,
) -> impl IntoResponse {
    let status = query.status.as_ref().and_then(|s| match s.to_uppercase().as_str() {
        "PENDING" => Some(SwapRequestStatus::Pending),
        "CONFIRMED" => Some(SwapRequestStatus::Confirmed),
        "REJECTED" => Some(SwapRequestStatus::Rejected),
        "EXPIRED" => Some(SwapRequestStatus::Expired),
        _ => None,
    });
    
    let requests = state.service.list_swap_requests(status);
    Json(requests)
}

pub async fn create_swap_request(
    State(state): State<Arc<AppState>>,
    Json(payload): Json<CreateSwapRequest>,
) -> impl IntoResponse {
    match state.service.create_swap_request(&payload.requester_id, &payload.target_id) {
        Ok(request) => (StatusCode::CREATED, Json(request)).into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({ "error": e.to_string() })),
        ).into_response(),
    }
}

pub async fn get_swap_request(
    State(state): State<Arc<AppState>>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match state.service.get_swap_request(&id) {
        Ok(request) => (StatusCode::OK, Json(Some(request))).into_response(),
        Err(e) => (
            StatusCode::NOT_FOUND,
            Json(serde_json::json!({ "error": e.to_string() })),
        ).into_response(),
    }
}

pub async fn respond_swap_request(
    State(state): State<Arc<AppState>>,
    Path(id): Path<String>,
    Json(payload): Json<RespondSwapRequest>,
) -> impl IntoResponse {
    let target_id = match state.service.get_swap_request(&id) {
        Ok(req) => req.target_id.clone(),
        Err(e) => {
            return (
                StatusCode::NOT_FOUND,
                Json(serde_json::json!({ "error": e.to_string() })),
            ).into_response();
        }
    };

    match state.service.respond_swap_request(&id, &target_id, payload.accept) {
        Ok(request) => (StatusCode::OK, Json(request)).into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({ "error": e.to_string() })),
        ).into_response(),
    }
}

pub async fn get_available_beds_count(
    State(state): State<Arc<AppState>>,
) -> impl IntoResponse {
    let count = state.service.get_available_beds_count();
    Json(serde_json::json!({ "available_beds": count }))
}
