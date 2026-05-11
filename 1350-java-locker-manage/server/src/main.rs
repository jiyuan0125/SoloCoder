use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::{IntoResponse, Json, Response},
    routing::{get, post},
    Router,
};
use clap::Parser;
use locker_core::*;
use serde::Serialize;
use std::net::SocketAddr;
use std::sync::Arc;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "LOCKER_PORT", default_value_t = 3000)]
    port: u16,

    #[arg(long, env = "LOCKER_HOST", default_value = "127.0.0.1")]
    host: String,
}

#[derive(Debug, Serialize)]
struct ErrorResponse {
    error: String,
    message: String,
}

struct AppError(LockerError);

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let status = match &self.0 {
            LockerError::LockerNotFound(_) => StatusCode::NOT_FOUND,
            LockerError::NoAvailableCompartment(_) => StatusCode::CONFLICT,
            LockerError::InvalidPickupCode => StatusCode::BAD_REQUEST,
            LockerError::PickupLocked(_) => StatusCode::LOCKED,
            LockerError::PickupCodeGenerationFailed => StatusCode::INTERNAL_SERVER_ERROR,
            LockerError::PackageNotFound => StatusCode::NOT_FOUND,
        };

        let body = Json(ErrorResponse {
            error: status.to_string(),
            message: self.0.to_string(),
        });

        (status, body).into_response()
    }
}

impl From<LockerError> for AppError {
    fn from(err: LockerError) -> Self {
        AppError(err)
    }
}

type AppState = Arc<LockerSystem>;
type AppResult<T> = std::result::Result<T, AppError>;

async fn list_lockers(State(state): State<AppState>) -> Json<Vec<LockerInfo>> {
    Json(state.list_lockers())
}

async fn get_locker(
    State(state): State<AppState>,
    Path(locker_id): Path<String>,
) -> AppResult<Json<LockerDetails>> {
    let details = state.get_locker_details(&locker_id)?;
    Ok(Json(details))
}

async fn get_locker_info(
    State(state): State<AppState>,
    Path(locker_id): Path<String>,
) -> AppResult<Json<LockerInfo>> {
    let info = state.get_locker_info(&locker_id)?;
    Ok(Json(info))
}

async fn deposit_package(
    State(state): State<AppState>,
    Json(request): Json<DepositRequest>,
) -> AppResult<Json<DepositResponse>> {
    let response = state.deposit_package(&request)?;
    Ok(Json(response))
}

async fn pickup_package(
    State(state): State<AppState>,
    Json(request): Json<PickupRequest>,
) -> AppResult<Json<PickupResponse>> {
    let response = state.pickup_package(&request)?;
    Ok(Json(response))
}

async fn list_packages_by_phone(
    State(state): State<AppState>,
    Path(phone): Path<String>,
) -> Json<Vec<PackageInfo>> {
    let packages = state.list_packages_by_phone(&phone);
    Json(packages)
}

async fn health_check() -> Json<serde_json::Value> {
    Json(serde_json::json!({
        "status": "ok",
        "version": env!("CARGO_PKG_VERSION"),
    }))
}

#[tokio::main]
async fn main() {
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "locker_server=debug,tower_http=debug,axum=info".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    let args = Args::parse();

    let system = create_sample_locker_system();
    let app = Router::new()
        .route("/health", get(health_check))
        .route("/lockers", get(list_lockers))
        .route("/lockers/:id", get(get_locker))
        .route("/lockers/:id/info", get(get_locker_info))
        .route("/deposit", post(deposit_package))
        .route("/pickup", post(pickup_package))
        .route("/packages/phone/:phone", get(list_packages_by_phone))
        .with_state(system);

    let addr: SocketAddr = format!("{}:{}", args.host, args.port).parse().unwrap();
    tracing::info!("Server listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
