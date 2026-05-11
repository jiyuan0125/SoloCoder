use std::sync::Arc;
use std::time::Duration;

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::{IntoResponse, Json},
    routing::{get, post},
    Router,
};
use clap::Parser;
use tracing::info;
use uuid::Uuid;

use lab_core::models::{
    ConfirmBookingRequest, CreateBookingRequest, CreateConsumableRequest, CreateEquipmentRequest,
    CreateUserRequest, UpdateConsumableStockRequest,
};
use lab_core::service::{BookingService, ConsumableService, EquipmentService, UserService};
use lab_core::{AppState, Scheduler};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "PORT", default_value_t = 3000)]
    port: u16,

    #[arg(short, long, env = "HOST", default_value = "127.0.0.1")]
    host: String,
}

struct AppStateWrapper {
    inner: AppState,
}

async fn create_user(
    State(state): State<Arc<AppStateWrapper>>,
    Json(req): Json<CreateUserRequest>,
) -> impl IntoResponse {
    match UserService::create(&state.inner, req) {
        Ok(user) => (StatusCode::CREATED, Json(user)).into_response(),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({"error": e.to_string()}))).into_response(),
    }
}

async fn list_users(State(state): State<Arc<AppStateWrapper>>) -> impl IntoResponse {
    match UserService::list(&state.inner) {
        Ok(users) => (StatusCode::OK, Json(users)).into_response(),
        Err(e) => (StatusCode::INTERNAL_SERVER_ERROR, Json(serde_json::json!({"error": e.to_string()}))).into_response(),
    }
}

async fn get_user(State(state): State<Arc<AppStateWrapper>>, Path(id): Path<Uuid>) -> impl IntoResponse {
    match UserService::get(&state.inner, id) {
        Ok(user) => (StatusCode::OK, Json(user)).into_response(),
        Err(e) => (StatusCode::NOT_FOUND, Json(serde_json::json!({"error": e.to_string()}))).into_response(),
    }
}

async fn create_equipment(
    State(state): State<Arc<AppStateWrapper>>,
    Json(req): Json<CreateEquipmentRequest>,
) -> impl IntoResponse {
    match EquipmentService::create(&state.inner, req) {
        Ok(equipment) => (StatusCode::CREATED, Json(equipment)).into_response(),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({"error": e.to_string()}))).into_response(),
    }
}

async fn list_equipment(State(state): State<Arc<AppStateWrapper>>) -> impl IntoResponse {
    match EquipmentService::list(&state.inner) {
        Ok(equipment) => (StatusCode::OK, Json(equipment)).into_response(),
        Err(e) => (StatusCode::INTERNAL_SERVER_ERROR, Json(serde_json::json!({"error": e.to_string()}))).into_response(),
    }
}

async fn get_equipment(State(state): State<Arc<AppStateWrapper>>, Path(id): Path<Uuid>) -> impl IntoResponse {
    match EquipmentService::get(&state.inner, id) {
        Ok(equipment) => (StatusCode::OK, Json(equipment)).into_response(),
        Err(e) => (StatusCode::NOT_FOUND, Json(serde_json::json!({"error": e.to_string()}))).into_response(),
    }
}

async fn create_consumable(
    State(state): State<Arc<AppStateWrapper>>,
    Json(req): Json<CreateConsumableRequest>,
) -> impl IntoResponse {
    match ConsumableService::create(&state.inner, req) {
        Ok(consumable) => (StatusCode::CREATED, Json(consumable)).into_response(),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({"error": e.to_string()}))).into_response(),
    }
}

async fn list_consumables(State(state): State<Arc<AppStateWrapper>>) -> impl IntoResponse {
    match ConsumableService::list(&state.inner) {
        Ok(consumables) => (StatusCode::OK, Json(consumables)).into_response(),
        Err(e) => (StatusCode::INTERNAL_SERVER_ERROR, Json(serde_json::json!({"error": e.to_string()}))).into_response(),
    }
}

async fn get_consumable(State(state): State<Arc<AppStateWrapper>>, Path(id): Path<Uuid>) -> impl IntoResponse {
    match ConsumableService::get(&state.inner, id) {
        Ok(consumable) => (StatusCode::OK, Json(consumable)).into_response(),
        Err(e) => (StatusCode::NOT_FOUND, Json(serde_json::json!({"error": e.to_string()}))).into_response(),
    }
}

async fn restock_consumable(
    State(state): State<Arc<AppStateWrapper>>,
    Path(id): Path<Uuid>,
    Json(req): Json<UpdateConsumableStockRequest>,
) -> impl IntoResponse {
    match ConsumableService::restock(&state.inner, id, req) {
        Ok(consumable) => (StatusCode::OK, Json(consumable)).into_response(),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({"error": e.to_string()}))).into_response(),
    }
}

async fn get_low_stock_alerts(State(state): State<Arc<AppStateWrapper>>) -> impl IntoResponse {
    match ConsumableService::get_low_stock_alerts(&state.inner) {
        Ok(alerts) => (StatusCode::OK, Json(alerts)).into_response(),
        Err(e) => (StatusCode::INTERNAL_SERVER_ERROR, Json(serde_json::json!({"error": e.to_string()}))).into_response(),
    }
}

async fn create_booking(
    State(state): State<Arc<AppStateWrapper>>,
    Json(req): Json<CreateBookingRequest>,
) -> impl IntoResponse {
    match BookingService::create_booking(&state.inner, req) {
        Ok(booking) => (StatusCode::CREATED, Json(booking)).into_response(),
        Err(e) => {
            let status = match e {
                lab_core::errors::AppError::TimeSlotConflict => StatusCode::CONFLICT,
                _ => StatusCode::BAD_REQUEST,
            };
            (status, Json(serde_json::json!({"error": e.to_string()}))).into_response()
        }
    }
}

async fn list_bookings(State(state): State<Arc<AppStateWrapper>>) -> impl IntoResponse {
    match BookingService::list_bookings(&state.inner) {
        Ok(bookings) => (StatusCode::OK, Json(bookings)).into_response(),
        Err(e) => (StatusCode::INTERNAL_SERVER_ERROR, Json(serde_json::json!({"error": e.to_string()}))).into_response(),
    }
}

async fn get_booking(State(state): State<Arc<AppStateWrapper>>, Path(id): Path<Uuid>) -> impl IntoResponse {
    match BookingService::get_booking(&state.inner, id) {
        Ok(booking) => (StatusCode::OK, Json(booking)).into_response(),
        Err(e) => (StatusCode::NOT_FOUND, Json(serde_json::json!({"error": e.to_string()}))).into_response(),
    }
}

async fn confirm_booking(
    State(state): State<Arc<AppStateWrapper>>,
    Path(booking_id): Path<Uuid>,
    Json(req): Json<ConfirmBookingRequest>,
) -> impl IntoResponse {
    match BookingService::confirm_booking(&state.inner, booking_id, req.admin_id) {
        Ok(booking) => (StatusCode::OK, Json(booking)).into_response(),
        Err(e) => {
            let status = match e {
                lab_core::errors::AppError::Unauthorized => StatusCode::FORBIDDEN,
                _ => StatusCode::BAD_REQUEST,
            };
            (status, Json(serde_json::json!({"error": e.to_string()}))).into_response()
        }
    }
}

async fn start_booking(
    State(state): State<Arc<AppStateWrapper>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match BookingService::start_booking(&state.inner, id) {
        Ok(booking) => (StatusCode::OK, Json(booking)).into_response(),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({"error": e.to_string()}))).into_response(),
    }
}

async fn complete_booking(
    State(state): State<Arc<AppStateWrapper>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match BookingService::complete_booking(&state.inner, id) {
        Ok(booking) => (StatusCode::OK, Json(booking)).into_response(),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({"error": e.to_string()}))).into_response(),
    }
}

async fn cancel_booking(
    State(state): State<Arc<AppStateWrapper>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match BookingService::cancel_booking(&state.inner, id) {
        Ok(booking) => (StatusCode::OK, Json(booking)).into_response(),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({"error": e.to_string()}))).into_response(),
    }
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();
    let args = Args::parse();

    let app_state = AppState::new();
    let _scheduler = Scheduler::new(app_state.clone(), Duration::from_secs(1)).start();

    let wrapper = Arc::new(AppStateWrapper { inner: app_state });

    let app = Router::new()
        .route("/api/users", post(create_user).get(list_users))
        .route("/api/users/:id", get(get_user))
        .route("/api/equipment", post(create_equipment).get(list_equipment))
        .route("/api/equipment/:id", get(get_equipment))
        .route("/api/consumables", post(create_consumable).get(list_consumables))
        .route("/api/consumables/:id", get(get_consumable))
        .route("/api/consumables/:id/restock", post(restock_consumable))
        .route("/api/alerts/low-stock", get(get_low_stock_alerts))
        .route("/api/bookings", post(create_booking).get(list_bookings))
        .route("/api/bookings/:id", get(get_booking))
        .route("/api/bookings/:id/confirm", post(confirm_booking))
        .route("/api/bookings/:id/start", post(start_booking))
        .route("/api/bookings/:id/complete", post(complete_booking))
        .route("/api/bookings/:id/cancel", post(cancel_booking))
        .with_state(wrapper);

    let addr = format!("{}:{}", args.host, args.port);
    info!("Server listening on {}", addr);

    axum::Server::bind(&addr.parse().unwrap())
        .serve(app.into_make_service())
        .await
        .unwrap();
}
