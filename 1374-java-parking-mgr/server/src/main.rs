use std::net::SocketAddr;
use std::sync::Arc;

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::{IntoResponse, Response},
    routing::{get, post},
    Json, Router,
};
use clap::Parser;
use serde::{Deserialize, Serialize};

use parking_core::{
    MonthlyCard, ParkingRecord, ParkingService, ParkingSpace, SpaceType, User, VehicleType,
};

#[derive(Parser, Debug)]
#[command(name = "parking-server", about = "Parking Management System Server")]
struct Args {
    #[arg(short, long, env = "PORT", default_value = "3000")]
    port: u16,
}

struct AppState {
    service: ParkingService,
}

#[derive(Serialize)]
struct ApiResponse<T> {
    success: bool,
    data: Option<T>,
    error: Option<String>,
}

impl<T: Serialize> ApiResponse<T> {
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

struct AppError(anyhow::Error);

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(ApiResponse::<()>::error(self.0.to_string())),
        )
            .into_response()
    }
}

impl<E> From<E> for AppError
where
    E: Into<anyhow::Error>,
{
    fn from(err: E) -> Self {
        Self(err.into())
    }
}

#[derive(Deserialize)]
struct EnterParkingRequest {
    plate_number: String,
    vehicle_type: String,
    user_id: Option<String>,
}

#[derive(Deserialize)]
struct ExitParkingRequest {
    plate_number: String,
}

#[derive(Deserialize)]
struct AddUserRequest {
    id: String,
    name: String,
    phone: String,
}

#[derive(Deserialize)]
struct AddMonthlyCardRequest {
    user_id: String,
    vehicle_plate: String,
    start_time: String,
    end_time: String,
    reserved_space_id: Option<String>,
}

#[derive(Deserialize)]
struct AddSpaceRequest {
    id: String,
    space_type: String,
}

fn parse_space_type(s: &str) -> Result<SpaceType, String> {
    match s.to_lowercase().as_str() {
        "normal" => Ok(SpaceType::Normal),
        "charging" => Ok(SpaceType::Charging),
        "accessible" => Ok(SpaceType::Accessible),
        _ => Err(format!("Invalid space type: {}", s)),
    }
}

fn parse_vehicle_type(s: &str) -> Result<VehicleType, String> {
    match s.to_lowercase().as_str() {
        "regular" => Ok(VehicleType::Regular),
        "electric" => Ok(VehicleType::Electric),
        "disabled" => Ok(VehicleType::Disabled),
        _ => Err(format!("Invalid vehicle type: {}", s)),
    }
}

async fn list_spaces(State(state): State<Arc<AppState>>) -> Response {
    let spaces = state.service.get_all_spaces();
    (StatusCode::OK, Json(ApiResponse::success(spaces))).into_response()
}

async fn get_space(State(state): State<Arc<AppState>>, Path(space_id): Path<String>) -> Response {
    match state.service.get_space(&space_id) {
        Some(space) => (StatusCode::OK, Json(ApiResponse::success(space))).into_response(),
        None => (
            StatusCode::NOT_FOUND,
            Json(ApiResponse::<()>::error(format!("Space not found: {}", space_id))),
        )
            .into_response(),
    }
}

async fn add_space(
    State(state): State<Arc<AppState>>,
    Json(req): Json<AddSpaceRequest>,
) -> Response {
    let space_type = match parse_space_type(&req.space_type) {
        Ok(t) => t,
        Err(e) => return (StatusCode::BAD_REQUEST, Json(ApiResponse::<()>::error(e))).into_response(),
    };

    let space = ParkingSpace::new(req.id, space_type);
    state.service.add_space(space.clone());

    (StatusCode::CREATED, Json(ApiResponse::success(space))).into_response()
}

async fn list_users(State(state): State<Arc<AppState>>) -> Response {
    let users = state.service.get_all_users();
    (StatusCode::OK, Json(ApiResponse::success(users))).into_response()
}

async fn get_user(State(state): State<Arc<AppState>>, Path(user_id): Path<String>) -> Response {
    match state.service.get_user(&user_id) {
        Some(user) => (StatusCode::OK, Json(ApiResponse::success(user))).into_response(),
        None => (
            StatusCode::NOT_FOUND,
            Json(ApiResponse::<()>::error(format!("User not found: {}", user_id))),
        )
            .into_response(),
    }
}

async fn add_user(
    State(state): State<Arc<AppState>>,
    Json(req): Json<AddUserRequest>,
) -> Response {
    let user = User::new(req.id, req.name, req.phone);
    state.service.add_user(user.clone());

    (StatusCode::CREATED, Json(ApiResponse::success(user))).into_response()
}

async fn list_monthly_cards(State(state): State<Arc<AppState>>) -> Response {
    let cards = state.service.get_all_monthly_cards();
    (StatusCode::OK, Json(ApiResponse::success(cards))).into_response()
}

async fn get_reminders(State(state): State<Arc<AppState>>) -> Response {
    let cards = state.service.get_cards_needing_reminder();
    (StatusCode::OK, Json(ApiResponse::success(cards))).into_response()
}

async fn add_monthly_card(
    State(state): State<Arc<AppState>>,
    Json(req): Json<AddMonthlyCardRequest>,
) -> Response {
    let start_time = match chrono::DateTime::parse_from_rfc3339(&req.start_time) {
        Ok(t) => t.with_timezone(&chrono::Utc),
        Err(e) => {
            return (
                StatusCode::BAD_REQUEST,
                Json(ApiResponse::<()>::error(format!("Invalid start time: {}", e))),
            )
                .into_response()
        }
    };

    let end_time = match chrono::DateTime::parse_from_rfc3339(&req.end_time) {
        Ok(t) => t.with_timezone(&chrono::Utc),
        Err(e) => {
            return (
                StatusCode::BAD_REQUEST,
                Json(ApiResponse::<()>::error(format!("Invalid end time: {}", e))),
            )
                .into_response()
        }
    };

    let card = MonthlyCard {
        user_id: req.user_id,
        vehicle_plate: req.vehicle_plate,
        start_time,
        end_time,
        reserved_space_id: req.reserved_space_id,
    };

    match state.service.add_monthly_card(card.clone()) {
        Ok(_) => (StatusCode::CREATED, Json(ApiResponse::success(card))).into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::<()>::error(e.to_string())),
        )
            .into_response(),
    }
}

async fn enter_parking(
    State(state): State<Arc<AppState>>,
    Json(req): Json<EnterParkingRequest>,
) -> Response {
    let vehicle_type = match parse_vehicle_type(&req.vehicle_type) {
        Ok(t) => t,
        Err(e) => return (StatusCode::BAD_REQUEST, Json(ApiResponse::<()>::error(e))).into_response(),
    };

    match state
        .service
        .enter_parking(&req.plate_number, vehicle_type, req.user_id)
    {
        Ok(record) => (StatusCode::OK, Json(ApiResponse::success(record))).into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::<()>::error(e.to_string())),
        )
            .into_response(),
    }
}

async fn exit_parking(
    State(state): State<Arc<AppState>>,
    Json(req): Json<ExitParkingRequest>,
) -> Response {
    match state.service.exit_parking(&req.plate_number) {
        Ok(record) => (StatusCode::OK, Json(ApiResponse::success(record))).into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::<()>::error(e.to_string())),
        )
            .into_response(),
    }
}

async fn get_active_record(
    State(state): State<Arc<AppState>>,
    Path(plate): Path<String>,
) -> Response {
    match state.service.get_active_record(&plate) {
        Some(record) => (StatusCode::OK, Json(ApiResponse::success(record))).into_response(),
        None => (
            StatusCode::NOT_FOUND,
            Json(ApiResponse::<()>::error(format!("No active record for plate: {}", plate))),
        )
            .into_response(),
    }
}

async fn list_records(State(state): State<Arc<AppState>>) -> Response {
    let records = state.service.get_all_records();
    (StatusCode::OK, Json(ApiResponse::success(records))).into_response()
}

async fn get_record(State(state): State<Arc<AppState>>, Path(record_id): Path<String>) -> Response {
    match state.service.get_record(&record_id) {
        Some(record) => (StatusCode::OK, Json(ApiResponse::success(record))).into_response(),
        None => (
            StatusCode::NOT_FOUND,
            Json(ApiResponse::<()>::error(format!("Record not found: {}", record_id))),
        )
            .into_response(),
    }
}

#[tokio::main]
async fn main() {
    let args = Args::parse();

    let service = ParkingService::new();

    let state = Arc::new(AppState { service });

    let app = Router::new()
        .route("/spaces", get(list_spaces).post(add_space))
        .route("/spaces/:id", get(get_space))
        .route("/users", get(list_users).post(add_user))
        .route("/users/:id", get(get_user))
        .route("/monthly-cards", get(list_monthly_cards).post(add_monthly_card))
        .route("/monthly-cards/reminders", get(get_reminders))
        .route("/parking/enter", post(enter_parking))
        .route("/parking/exit", post(exit_parking))
        .route("/parking/active/:plate", get(get_active_record))
        .route("/records", get(list_records))
        .route("/records/:id", get(get_record))
        .with_state(state);

    let addr = SocketAddr::from(([127, 0, 0, 1], args.port));
    println!("Parking Management Server listening on {}", addr);

    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}
