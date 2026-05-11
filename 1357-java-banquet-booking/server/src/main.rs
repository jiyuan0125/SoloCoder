extern crate alloc;

use std::sync::Arc;

use axum::{
    extract::{Path, State},
    http::StatusCode,
    routing::{get, post, put, delete},
    Json, Router,
};
use clap::Parser;
use booking_core::{
    BookingService, BookingError,
    AdjustTablesRequest, CancelBookingRequest, CompleteBookingRequest,
    CreateBookingRequest, SubstituteDishRequest,
};
use tower_http::cors::{Any, CorsLayer};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(long, env = "PORT", default_value_t = 3000)]
    port: u16,
    
    #[arg(long, env = "HOST", default_value = "127.0.0.1")]
    host: String,
}

type AppState = Arc<BookingService>;

#[tokio::main]
async fn main() {
    let args = Args::parse();
    
    let service = BookingService::new();
    let state = Arc::new(service);

    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    let app = Router::new()
        .route("/menu-sets", get(list_menu_sets))
        .route("/menu-sets/:id", get(get_menu_set))
        .route("/menu-items", get(list_menu_items))
        .route("/bookings", get(list_bookings).post(create_booking))
        .route("/bookings/:id", get(get_booking))
        .route("/bookings/:id/adjust-tables", put(adjust_tables))
        .route("/bookings/:id/substitute-dish", put(substitute_dish))
        .route("/bookings/:id/complete", put(complete_booking))
        .route("/bookings/:id/cancel", delete(cancel_booking))
        .route("/bookings/:id/summary", get(get_booking_summary))
        .with_state(state)
        .layer(cors);

    let addr = format!("{}:{}", args.host, args.port);
    println!("宴会预订管理系统服务端运行在: http://{}", addr);
    
    axum::Server::bind(&addr.parse().unwrap())
        .serve(app.into_make_service())
        .await
        .unwrap();
}

async fn list_menu_sets(State(state): State<AppState>) -> Json<serde_json::Value> {
    let sets = state.list_menu_sets();
    Json(serde_json::to_value(sets).unwrap())
}

async fn get_menu_set(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<Json<serde_json::Value>, StatusCode> {
    let uuid = uuid::Uuid::parse_str(&id).map_err(|_| StatusCode::BAD_REQUEST)?;
    match state.get_menu_set(uuid) {
        Some(set) => Ok(Json(serde_json::to_value(set).unwrap())),
        None => Err(StatusCode::NOT_FOUND),
    }
}

async fn list_menu_items(State(state): State<AppState>) -> Json<serde_json::Value> {
    let items = state.list_menu_items();
    Json(serde_json::to_value(items).unwrap())
}

async fn create_booking(
    State(state): State<AppState>,
    Json(req): Json<CreateBookingRequest>,
) -> Result<Json<serde_json::Value>, (StatusCode, String)> {
    match state.create_booking(req) {
        Ok(booking) => Ok(Json(serde_json::to_value(booking).unwrap())),
        Err(e) => Err(map_error(e)),
    }
}

async fn list_bookings(State(state): State<AppState>) -> Json<serde_json::Value> {
    let bookings = state.list_bookings();
    Json(serde_json::to_value(bookings).unwrap())
}

async fn get_booking(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<Json<serde_json::Value>, StatusCode> {
    let uuid = uuid::Uuid::parse_str(&id).map_err(|_| StatusCode::BAD_REQUEST)?;
    match state.get_booking(uuid) {
        Some(booking) => Ok(Json(serde_json::to_value(booking).unwrap())),
        None => Err(StatusCode::NOT_FOUND),
    }
}

async fn adjust_tables(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(req): Json<AdjustTablesRequest>,
) -> Result<Json<serde_json::Value>, (StatusCode, String)> {
    let uuid = uuid::Uuid::parse_str(&id).map_err(|_| (StatusCode::BAD_REQUEST, "无效的ID".to_string()))?;
    let mut req = req;
    req.booking_id = uuid;
    
    match state.adjust_tables(req) {
        Ok(booking) => Ok(Json(serde_json::to_value(booking).unwrap())),
        Err(e) => Err(map_error(e)),
    }
}

async fn substitute_dish(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(req): Json<SubstituteDishRequest>,
) -> Result<Json<serde_json::Value>, (StatusCode, String)> {
    let uuid = uuid::Uuid::parse_str(&id).map_err(|_| (StatusCode::BAD_REQUEST, "无效的ID".to_string()))?;
    let mut req = req;
    req.booking_id = uuid;
    
    match state.substitute_dish(req) {
        Ok(booking) => Ok(Json(serde_json::to_value(booking).unwrap())),
        Err(e) => Err(map_error(e)),
    }
}

async fn complete_booking(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(req): Json<CompleteBookingRequest>,
) -> Result<Json<serde_json::Value>, (StatusCode, String)> {
    let uuid = uuid::Uuid::parse_str(&id).map_err(|_| (StatusCode::BAD_REQUEST, "无效的ID".to_string()))?;
    let mut req = req;
    req.booking_id = uuid;
    
    match state.complete_booking(req) {
        Ok(booking) => Ok(Json(serde_json::to_value(booking).unwrap())),
        Err(e) => Err(map_error(e)),
    }
}

async fn cancel_booking(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<Json<serde_json::Value>, (StatusCode, String)> {
    let uuid = uuid::Uuid::parse_str(&id).map_err(|_| (StatusCode::BAD_REQUEST, "无效的ID".to_string()))?;
    let req = CancelBookingRequest { booking_id: uuid };
    
    match state.cancel_booking(req) {
        Ok(booking) => Ok(Json(serde_json::to_value(booking).unwrap())),
        Err(e) => Err(map_error(e)),
    }
}

async fn get_booking_summary(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<Json<serde_json::Value>, (StatusCode, String)> {
    let uuid = uuid::Uuid::parse_str(&id).map_err(|_| (StatusCode::BAD_REQUEST, "无效的ID".to_string()))?;
    
    match state.get_booking_summary(uuid) {
        Ok(summary) => Ok(Json(serde_json::to_value(summary).unwrap())),
        Err(e) => Err(map_error(e)),
    }
}

fn map_error(err: BookingError) -> (StatusCode, String) {
    match err {
        BookingError::BookingNotFound => (StatusCode::NOT_FOUND, err.to_string()),
        BookingError::MenuSetNotFound => (StatusCode::NOT_FOUND, err.to_string()),
        BookingError::MenuItemNotFound => (StatusCode::NOT_FOUND, err.to_string()),
        BookingError::BookingInactive => (StatusCode::BAD_REQUEST, err.to_string()),
        BookingError::InvalidEventDate => (StatusCode::BAD_REQUEST, err.to_string()),
        BookingError::InvalidTableCount => (StatusCode::BAD_REQUEST, err.to_string()),
        BookingError::InvalidDeposit => (StatusCode::BAD_REQUEST, err.to_string()),
        BookingError::PriceDifferenceExceeded => (StatusCode::BAD_REQUEST, err.to_string()),
        BookingError::DifferentCategory => (StatusCode::BAD_REQUEST, err.to_string()),
        BookingError::MixedUnusedOption => (StatusCode::BAD_REQUEST, err.to_string()),
        BookingError::ActualTablesExceedBooked => (StatusCode::BAD_REQUEST, err.to_string()),
        BookingError::OperationTooLate => (StatusCode::BAD_REQUEST, err.to_string()),
        BookingError::CalculationError => (StatusCode::INTERNAL_SERVER_ERROR, err.to_string()),
    }
}
