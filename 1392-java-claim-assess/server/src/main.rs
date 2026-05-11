use std::net::SocketAddr;
use std::sync::Arc;

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::{IntoResponse, Json},
    routing::{get, post},
    Router,
};
use clap::Parser;
use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};
use tower_http::cors::{Any, CorsLayer};

use claim_core::{AppError, ClaimService, InMemoryStore, RepairItemType};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(long, env = "PORT", default_value_t = 3000)]
    port: u16,

    #[arg(long, env = "HOST", default_value = "127.0.0.1")]
    host: String,
}

type AppState = Arc<ClaimService>;

#[derive(Debug, Serialize, Deserialize)]
struct CreateVehicleRequest {
    plate_number: String,
    model: String,
    actual_value: Decimal,
}

#[derive(Debug, Serialize, Deserialize)]
struct UpdateVehicleValueRequest {
    actual_value: Decimal,
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateClaimRequest {
    vehicle_id: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct AddRepairItemRequest {
    name: String,
    item_type: RepairItemType,
    cost: Decimal,
}

#[derive(Debug, Serialize, Deserialize)]
struct SettleClaimRequest {
    salvage_value: Option<Decimal>,
}

#[derive(Debug, Serialize)]
struct ErrorResponse {
    error: String,
}

struct AppErrorResponse(AppError);

impl IntoResponse for AppErrorResponse {
    fn into_response(self) -> axum::response::Response {
        let status = match &self.0 {
            AppError::VehicleNotFound(_) => StatusCode::NOT_FOUND,
            AppError::ClaimNotFound(_) => StatusCode::NOT_FOUND,
            AppError::HasUnsettledClaim => StatusCode::BAD_REQUEST,
            AppError::SalvageValueExceedsActualValue => StatusCode::BAD_REQUEST,
            AppError::NegativePayout => StatusCode::BAD_REQUEST,
            AppError::ZeroActualValue => StatusCode::BAD_REQUEST,
            AppError::ClaimAlreadySettled => StatusCode::BAD_REQUEST,
            AppError::ClaimTotalLoss => StatusCode::BAD_REQUEST,
        };
        (
            status,
            Json(ErrorResponse {
                error: self.0.to_string(),
            }),
        )
            .into_response()
    }
}

impl From<AppError> for AppErrorResponse {
    fn from(err: AppError) -> Self {
        AppErrorResponse(err)
    }
}

async fn list_vehicles(State(service): State<AppState>) -> impl IntoResponse {
    let vehicles = service.list_vehicles();
    Json(vehicles)
}

async fn get_vehicle(
    State(service): State<AppState>,
    Path(id): Path<String>,
) -> Result<impl IntoResponse, AppErrorResponse> {
    let vehicle = service.get_vehicle(&id)?;
    Ok(Json(vehicle))
}

async fn get_vehicle_by_plate(
    State(service): State<AppState>,
    Path(plate_number): Path<String>,
) -> Result<impl IntoResponse, AppErrorResponse> {
    let vehicle = service.get_vehicle_by_plate(&plate_number)?;
    Ok(Json(vehicle))
}

async fn create_vehicle(
    State(service): State<AppState>,
    Json(req): Json<CreateVehicleRequest>,
) -> impl IntoResponse {
    let vehicle = service.create_vehicle(req.plate_number, req.model, req.actual_value);
    (StatusCode::CREATED, Json(vehicle))
}

async fn update_vehicle_value(
    State(service): State<AppState>,
    Path(id): Path<String>,
    Json(req): Json<UpdateVehicleValueRequest>,
) -> Result<impl IntoResponse, AppErrorResponse> {
    let vehicle = service.update_vehicle_actual_value(&id, req.actual_value)?;
    Ok(Json(vehicle))
}

async fn list_claims(State(service): State<AppState>) -> impl IntoResponse {
    let claims = service.list_claims();
    Json(claims)
}

async fn get_claim(
    State(service): State<AppState>,
    Path(id): Path<String>,
) -> Result<impl IntoResponse, AppErrorResponse> {
    let claim = service.get_claim(&id)?;
    Ok(Json(claim))
}

async fn list_claims_by_vehicle(
    State(service): State<AppState>,
    Path(vehicle_id): Path<String>,
) -> Result<impl IntoResponse, AppErrorResponse> {
    let claims = service.list_claims_by_vehicle(&vehicle_id)?;
    Ok(Json(claims))
}

async fn create_claim(
    State(service): State<AppState>,
    Json(req): Json<CreateClaimRequest>,
) -> Result<impl IntoResponse, AppErrorResponse> {
    let claim = service.create_claim(&req.vehicle_id)?;
    Ok((StatusCode::CREATED, Json(claim)))
}

async fn add_repair_item(
    State(service): State<AppState>,
    Path(claim_id): Path<String>,
    Json(req): Json<AddRepairItemRequest>,
) -> Result<impl IntoResponse, AppErrorResponse> {
    let claim = service.add_repair_item(&claim_id, req.name, req.item_type, req.cost)?;
    Ok(Json(claim))
}

async fn settle_claim(
    State(service): State<AppState>,
    Path(claim_id): Path<String>,
    Json(req): Json<SettleClaimRequest>,
) -> Result<impl IntoResponse, AppErrorResponse> {
    let claim = service.settle_claim(&claim_id, req.salvage_value)?;
    Ok(Json(claim))
}

fn app(state: AppState) -> Router {
    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    Router::new()
        .route("/vehicles", get(list_vehicles).post(create_vehicle))
        .route("/vehicles/:id", get(get_vehicle).post(update_vehicle_value))
        .route("/vehicles/plate/:plate_number", get(get_vehicle_by_plate))
        .route("/vehicles/:vehicle_id/claims", get(list_claims_by_vehicle))
        .route("/claims", get(list_claims).post(create_claim))
        .route("/claims/:id", get(get_claim))
        .route("/claims/:id/items", post(add_repair_item))
        .route("/claims/:id/settle", post(settle_claim))
        .layer(cors)
        .with_state(state)
}

#[tokio::main]
async fn main() {
    let args = Args::parse();

    let store = InMemoryStore::new();
    let service = ClaimService::new(store);
    let state = Arc::new(service);

    let addr = format!("{}:{}", args.host, args.port).parse::<SocketAddr>().unwrap();

    println!("Server starting on {}", addr);

    axum::Server::bind(&addr)
        .serve(app(state).into_make_service())
        .await
        .unwrap();
}
