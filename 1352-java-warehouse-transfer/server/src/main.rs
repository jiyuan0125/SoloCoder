use std::net::SocketAddr;

use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    response::{IntoResponse, Json},
    routing::{get, post},
    Router,
};
use clap::Parser;
use warehouse_core::{
    AddInventoryRequest, CreateTransferRequest, ReceiveRequest, WarehouseService,
};
use serde::Deserialize;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

#[derive(Parser, Debug)]
#[command(version, about, long_about = None)]
struct Args {
    #[arg(long, env = "PORT", default_value = "3000")]
    port: u16,

    #[arg(long, env = "HOST", default_value = "127.0.0.1")]
    host: String,
}

#[derive(Debug, Deserialize)]
struct ListInventoryQuery {
    warehouse_id: Option<String>,
}

#[derive(Debug, Deserialize)]
struct CreateWarehouseQuery {
    name: String,
    city: String,
}

struct AppError(warehouse_core::WarehouseError);

impl IntoResponse for AppError {
    fn into_response(self) -> axum::response::Response {
        let (status, message) = match &self.0 {
            warehouse_core::WarehouseError::WarehouseNotFound(id) => {
                (StatusCode::NOT_FOUND, format!("Warehouse not found: {}", id))
            }
            warehouse_core::WarehouseError::ProductNotFound { product_id } => {
                (StatusCode::NOT_FOUND, format!("Product not found: {}", product_id))
            }
            warehouse_core::WarehouseError::InsufficientStock { available, requested } => (
                StatusCode::BAD_REQUEST,
                format!("Insufficient stock: available={}, requested={}", available, requested),
            ),
            warehouse_core::WarehouseError::TransferNotFound(id) => {
                (StatusCode::NOT_FOUND, format!("Transfer not found: {}", id))
            }
            warehouse_core::WarehouseError::InvalidStatus { current } => (
                StatusCode::BAD_REQUEST,
                format!("Invalid status: {}", current),
            ),
            warehouse_core::WarehouseError::CannotCancel => {
                (StatusCode::BAD_REQUEST, "Cannot cancel: transfer issued more than 24 hours and arrived at destination city".to_string())
            }
            warehouse_core::WarehouseError::ChainTransferNotAllowed { warehouse_id } => (
                StatusCode::BAD_REQUEST,
                format!(
                    "Chain transfer not allowed: warehouse {} has in-transit stock",
                    warehouse_id
                ),
            ),
            warehouse_core::WarehouseError::ConcurrentLockConflict { product_id } => (
                StatusCode::CONFLICT,
                format!("Product {} is locked by another operation", product_id),
            ),
            warehouse_core::WarehouseError::SameWarehouse => (
                StatusCode::BAD_REQUEST,
                "Target warehouse must be different from source warehouse".to_string(),
            ),
        };

        (
            status,
            Json(serde_json::json!({
                "error": message,
            })),
        )
            .into_response()
    }
}

impl From<warehouse_core::WarehouseError> for AppError {
    fn from(err: warehouse_core::WarehouseError) -> Self {
        AppError(err)
    }
}

type AppResult<T> = Result<T, AppError>;

async fn create_warehouse(
    State(service): State<WarehouseService>,
    Query(query): Query<CreateWarehouseQuery>,
) -> AppResult<impl IntoResponse> {
    let warehouse = service.create_warehouse(query.name, query.city);
    Ok((StatusCode::CREATED, Json(warehouse)))
}

async fn list_warehouses(State(service): State<WarehouseService>) -> impl IntoResponse {
    Json(service.list_warehouses())
}

async fn get_warehouse(
    State(service): State<WarehouseService>,
    Path(id): Path<String>,
) -> AppResult<impl IntoResponse> {
    match service.get_warehouse(&id) {
        Some(warehouse) => Ok(Json(warehouse)),
        None => Err(AppError(warehouse_core::WarehouseError::WarehouseNotFound(id))),
    }
}

async fn add_inventory(
    State(service): State<WarehouseService>,
    Json(req): Json<AddInventoryRequest>,
) -> AppResult<impl IntoResponse> {
    let inventory = service.add_inventory(
        &req.warehouse_id,
        &req.product_id,
        &req.product_name,
        req.quantity,
    )?;
    Ok((StatusCode::CREATED, Json(inventory)))
}

async fn list_inventory(
    State(service): State<WarehouseService>,
    Query(query): Query<ListInventoryQuery>,
) -> impl IntoResponse {
    Json(service.list_inventory(query.warehouse_id.as_deref()))
}

async fn get_inventory(
    State(service): State<WarehouseService>,
    Path((warehouse_id, product_id)): Path<(String, String)>,
) -> AppResult<impl IntoResponse> {
    match service.get_inventory(&warehouse_id, &product_id) {
        Some(inventory) => Ok(Json(inventory)),
        None => Err(AppError(warehouse_core::WarehouseError::ProductNotFound { product_id })),
    }
}

async fn create_transfer(
    State(service): State<WarehouseService>,
    Json(req): Json<CreateTransferRequest>,
) -> AppResult<impl IntoResponse> {
    let transfer = service.create_transfer(
        &req.source_warehouse_id,
        &req.target_warehouse_id,
        &req.product_id,
        req.quantity,
    )?;
    Ok((StatusCode::CREATED, Json(transfer)))
}

async fn list_transfers(State(service): State<WarehouseService>) -> impl IntoResponse {
    Json(service.list_transfers())
}

async fn get_transfer(
    State(service): State<WarehouseService>,
    Path(id): Path<String>,
) -> AppResult<impl IntoResponse> {
    match service.get_transfer(&id) {
        Some(transfer) => Ok(Json(transfer)),
        None => Err(AppError(warehouse_core::WarehouseError::TransferNotFound(id))),
    }
}

async fn ship_transfer(
    State(service): State<WarehouseService>,
    Path(id): Path<String>,
) -> AppResult<impl IntoResponse> {
    let transfer = service.ship_transfer(&id)?;
    Ok(Json(transfer))
}

async fn receive_transfer(
    State(service): State<WarehouseService>,
    Json(req): Json<ReceiveRequest>,
) -> AppResult<impl IntoResponse> {
    let transfer = service.receive_transfer(
        &req.transfer_id,
        req.received_quantity,
        req.discrepancy_reason,
    )?;
    Ok(Json(transfer))
}

async fn return_transfer(
    State(service): State<WarehouseService>,
    Path(id): Path<String>,
) -> AppResult<impl IntoResponse> {
    let transfer = service.return_transfer(&id)?;
    Ok(Json(transfer))
}

async fn cancel_transfer(
    State(service): State<WarehouseService>,
    Path(id): Path<String>,
) -> AppResult<impl IntoResponse> {
    let transfer = service.cancel_transfer(&id)?;
    Ok(Json(transfer))
}

async fn list_inventory_logs(State(service): State<WarehouseService>) -> impl IntoResponse {
    Json(service.list_inventory_logs())
}

async fn list_inventory_logs_by_transfer(
    State(service): State<WarehouseService>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    Json(service.list_inventory_logs_by_transfer(&id))
}

#[tokio::main]
async fn main() {
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "server=debug,tower_http=debug".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    let args = Args::parse();

    let service = WarehouseService::new();

    let app = Router::new()
        .route("/api/warehouses", post(create_warehouse).get(list_warehouses))
        .route("/api/warehouses/:id", get(get_warehouse))
        .route("/api/inventory", post(add_inventory).get(list_inventory))
        .route(
            "/api/inventory/:warehouse_id/:product_id",
            get(get_inventory),
        )
        .route("/api/transfers", post(create_transfer).get(list_transfers))
        .route("/api/transfers/:id", get(get_transfer))
        .route("/api/transfers/:id/ship", post(ship_transfer))
        .route("/api/transfers/receive", post(receive_transfer))
        .route("/api/transfers/:id/return", post(return_transfer))
        .route("/api/transfers/:id/cancel", post(cancel_transfer))
        .route("/api/logs", get(list_inventory_logs))
        .route("/api/logs/transfer/:id", get(list_inventory_logs_by_transfer))
        .with_state(service);

    let addr: SocketAddr = format!("{}:{}", args.host, args.port)
        .parse()
        .expect("Invalid address");

    tracing::info!("Listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
