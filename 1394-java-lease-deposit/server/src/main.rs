use std::sync::Arc;

use axum::{
    extract::{State, Path},
    http::StatusCode,
    response::{IntoResponse, Json},
    routing::{get, post},
    Router,
};
use clap::Parser;
use serde::{Deserialize, Serialize};
use tower_http::cors::{CorsLayer, Any};
use uuid::Uuid;

use lease_core::{
    LeaseService, LeaseError,
    CreateTenantRequest, CreateRoomRequest, CreateContractRequest,
    CheckoutRequest, RenewContractRequest,
};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None, disable_help_flag = true)]
struct Args {
    #[arg(short, long, env = "LEASE_SERVER_PORT", default_value_t = 3000)]
    port: u16,

    #[arg(short = 'H', long, env = "LEASE_SERVER_HOST", default_value = "0.0.0.0")]
    host: String,
}

struct AppState {
    service: LeaseService,
}

type AppStateRef = Arc<AppState>;

#[derive(Serialize)]
struct ErrorResponse {
    error: String,
}

fn map_error(error: LeaseError) -> (StatusCode, Json<ErrorResponse>) {
    let status = match &error {
        LeaseError::RoomNotFound(_) => StatusCode::NOT_FOUND,
        LeaseError::ContractNotFound(_) => StatusCode::NOT_FOUND,
        LeaseError::TenantNotFound(_) => StatusCode::NOT_FOUND,
        LeaseError::RoomAlreadyExists(_) => StatusCode::CONFLICT,
        LeaseError::RoomHasActiveContract(_) => StatusCode::CONFLICT,
        LeaseError::ContractNotActive => StatusCode::BAD_REQUEST,
        LeaseError::ContractNotFixedTerm => StatusCode::BAD_REQUEST,
        LeaseError::ContractAlreadyMonthToMonth => StatusCode::BAD_REQUEST,
        LeaseError::InvalidDate(_) => StatusCode::BAD_REQUEST,
        LeaseError::InvalidAmount(_) => StatusCode::BAD_REQUEST,
        LeaseError::ValidationError(_) => StatusCode::BAD_REQUEST,
        LeaseError::InternalError(_) => StatusCode::INTERNAL_SERVER_ERROR,
    };

    (status, Json(ErrorResponse {
        error: error.to_string(),
    }))
}

type ApiResult<T> = Result<Json<T>, (StatusCode, Json<ErrorResponse>)>;

#[derive(Deserialize)]
struct ProcessCheckinRequest {
    contract_id: Uuid,
}

async fn create_tenant(
    State(state): State<AppStateRef>,
    Json(req): Json<CreateTenantRequest>,
) -> ApiResult<lease_core::Tenant> {
    state.service.create_tenant(req)
        .map(Json)
        .map_err(map_error)
}

async fn list_tenants(
    State(state): State<AppStateRef>,
) -> ApiResult<Vec<lease_core::Tenant>> {
    state.service.list_tenants()
        .map(Json)
        .map_err(map_error)
}

async fn get_tenant(
    State(state): State<AppStateRef>,
    Path(id): Path<Uuid>,
) -> ApiResult<lease_core::Tenant> {
    match state.service.get_tenant(id) {
        Ok(Some(tenant)) => Ok(Json(tenant)),
        Ok(None) => Err(map_error(LeaseError::ValidationError("Tenant not found".into()))),
        Err(e) => Err(map_error(e)),
    }
}

async fn create_room(
    State(state): State<AppStateRef>,
    Json(req): Json<CreateRoomRequest>,
) -> ApiResult<lease_core::Room> {
    state.service.create_room(req)
        .map(Json)
        .map_err(map_error)
}

async fn list_rooms(
    State(state): State<AppStateRef>,
) -> ApiResult<Vec<lease_core::Room>> {
    state.service.list_rooms()
        .map(Json)
        .map_err(map_error)
}

async fn get_room(
    State(state): State<AppStateRef>,
    Path(id): Path<Uuid>,
) -> ApiResult<lease_core::Room> {
    match state.service.get_room(id) {
        Ok(Some(room)) => Ok(Json(room)),
        Ok(None) => Err(map_error(LeaseError::RoomNotFound(id.to_string()))),
        Err(e) => Err(map_error(e)),
    }
}

async fn create_contract(
    State(state): State<AppStateRef>,
    Json(req): Json<CreateContractRequest>,
) -> ApiResult<lease_core::Contract> {
    state.service.create_contract(req)
        .map(Json)
        .map_err(map_error)
}

async fn list_contracts(
    State(state): State<AppStateRef>,
) -> ApiResult<Vec<lease_core::Contract>> {
    state.service.list_contracts()
        .map(Json)
        .map_err(map_error)
}

async fn get_contract(
    State(state): State<AppStateRef>,
    Path(id): Path<Uuid>,
) -> ApiResult<lease_core::Contract> {
    match state.service.get_contract(id) {
        Ok(Some(contract)) => Ok(Json(contract)),
        Ok(None) => Err(map_error(LeaseError::ContractNotFound(id.to_string()))),
        Err(e) => Err(map_error(e)),
    }
}

async fn process_checkin(
    State(state): State<AppStateRef>,
    Json(req): Json<ProcessCheckinRequest>,
) -> ApiResult<lease_core::Contract> {
    state.service.process_checkin(req.contract_id)
        .map(Json)
        .map_err(map_error)
}

async fn get_renewal_reminders(
    State(state): State<AppStateRef>,
) -> ApiResult<Vec<lease_core::RenewalReminder>> {
    state.service.get_renewal_reminders()
        .map(Json)
        .map_err(map_error)
}

async fn process_expired_contract(
    State(state): State<AppStateRef>,
    Path(id): Path<Uuid>,
) -> ApiResult<lease_core::Contract> {
    state.service.process_expired_contract(id)
        .map(Json)
        .map_err(map_error)
}

async fn renew_contract(
    State(state): State<AppStateRef>,
    Json(req): Json<RenewContractRequest>,
) -> ApiResult<lease_core::Contract> {
    state.service.renew_contract(req)
        .map(Json)
        .map_err(map_error)
}

async fn process_checkout(
    State(state): State<AppStateRef>,
    Json(req): Json<CheckoutRequest>,
) -> ApiResult<lease_core::CheckoutResult> {
    state.service.process_checkout(req)
        .map(Json)
        .map_err(map_error)
}

async fn health_check() -> impl IntoResponse {
    (StatusCode::OK, Json(serde_json::json!({"status": "ok"})))
}

fn create_router(service: LeaseService) -> Router {
    let state = Arc::new(AppState { service });

    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    Router::new()
        .route("/health", get(health_check))
        .route("/tenants", post(create_tenant).get(list_tenants))
        .route("/tenants/:id", get(get_tenant))
        .route("/rooms", post(create_room).get(list_rooms))
        .route("/rooms/:id", get(get_room))
        .route("/contracts", post(create_contract).get(list_contracts))
        .route("/contracts/:id", get(get_contract))
        .route("/contracts/checkin", post(process_checkin))
        .route("/contracts/checkout", post(process_checkout))
        .route("/contracts/renew", post(renew_contract))
        .route("/contracts/:id/process-expired", post(process_expired_contract))
        .route("/reminders/renewal", get(get_renewal_reminders))
        .with_state(state)
        .layer(cors)
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_max_level(tracing::Level::INFO)
        .init();

    let args = Args::parse();

    let service = LeaseService::new();
    let app = create_router(service);

    let addr = format!("{}:{}", args.host, args.port);
    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    
    tracing::info!("Lease management server listening on {}", addr);
    axum::serve(listener, app).await.unwrap();
}
