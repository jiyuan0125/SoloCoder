use std::sync::Arc;

use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    response::{IntoResponse, Response},
    routing::{get, post},
    Json, Router,
};
use clap::Parser;
use serde::{Deserialize, Serialize};
use tokio::sync::Mutex;
use tower_http::cors::{Any, CorsLayer};
use uuid::Uuid;

use warehouse_core::models::*;
use warehouse_core::services::WarehouseService;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(short = 'p', long, env = "WAREHOUSE_PORT", default_value_t = 3000)]
    port: u16,

    #[arg(long, env = "WAREHOUSE_HOST", default_value = "0.0.0.0")]
    host: String,
}

type AppState = Arc<Mutex<WarehouseService>>;

#[derive(Debug, Serialize)]
struct ErrorResponse {
    error: String,
}

#[derive(Debug, Deserialize)]
struct CreateMaterialRequest {
    name: String,
    safety_stock: i32,
    lead_time_days: u32,
    unit: String,
}

#[derive(Debug, Deserialize)]
struct InboundRequest {
    material_id: Uuid,
    quantity: i32,
    supplier: String,
    batch_number: String,
}

#[derive(Debug, Deserialize)]
struct OutboundRequest {
    material_id: Uuid,
    quantity: i32,
}

#[derive(Debug, Deserialize)]
struct BatchOutboundRequest {
    items: Vec<BatchOutboundItem>,
}

#[derive(Debug, Deserialize)]
struct ListQuery {
    material_id: Option<Uuid>,
    status: Option<ReplenishmentStatus>,
}

fn internal_error(e: impl std::fmt::Display) -> Response {
    (
        StatusCode::INTERNAL_SERVER_ERROR,
        Json(ErrorResponse {
            error: e.to_string(),
        }),
    )
        .into_response()
}

fn bad_request(e: impl std::fmt::Display) -> Response {
    (
        StatusCode::BAD_REQUEST,
        Json(ErrorResponse {
            error: e.to_string(),
        }),
    )
        .into_response()
}

fn not_found(e: impl std::fmt::Display) -> Response {
    (
        StatusCode::NOT_FOUND,
        Json(ErrorResponse {
            error: e.to_string(),
        }),
    )
        .into_response()
}

async fn list_materials(State(state): State<AppState>) -> Response {
    let service = state.lock().await;
    let materials = service.list_materials();
    (StatusCode::OK, Json(materials)).into_response()
}

async fn create_material(
    State(state): State<AppState>,
    Json(req): Json<CreateMaterialRequest>,
) -> Response {
    let mut service = state.lock().await;
    let material = service.add_material(req.name, req.safety_stock, req.lead_time_days, req.unit);
    (StatusCode::CREATED, Json(material)).into_response()
}

async fn get_material(State(state): State<AppState>, Path(id): Path<Uuid>) -> Response {
    let service = state.lock().await;
    match service.get_material(id) {
        Some(material) => (StatusCode::OK, Json(material)).into_response(),
        None => not_found("原料不存在"),
    }
}

async fn list_inventories(State(state): State<AppState>) -> Response {
    let service = state.lock().await;
    let inventories = service.list_inventories();
    (StatusCode::OK, Json(inventories)).into_response()
}

async fn get_inventory(State(state): State<AppState>, Path(id): Path<Uuid>) -> Response {
    let service = state.lock().await;
    match service.get_inventory(id) {
        Some(inventory) => {
            let material = service.get_material(id).unwrap();
            (
                StatusCode::OK,
                Json(MaterialInventory {
                    material,
                    current_stock: inventory,
                }),
            )
                .into_response()
        }
        None => not_found("原料不存在"),
    }
}

async fn inbound(State(state): State<AppState>, Json(req): Json<InboundRequest>) -> Response {
    let mut service = state.lock().await;
    match service.inbound(req.material_id, req.quantity, req.supplier, req.batch_number) {
        Ok(record) => (StatusCode::CREATED, Json(record)).into_response(),
        Err(e) => bad_request(e),
    }
}

async fn outbound(State(state): State<AppState>, Json(req): Json<OutboundRequest>) -> Response {
    let mut service = state.lock().await;
    match service.outbound(req.material_id, req.quantity) {
        Ok(record) => (StatusCode::CREATED, Json(record)).into_response(),
        Err(e) => bad_request(e),
    }
}

async fn batch_outbound(
    State(state): State<AppState>,
    Json(req): Json<BatchOutboundRequest>,
) -> Response {
    let mut service = state.lock().await;
    match service.batch_outbound(req.items) {
        Ok(records) => (StatusCode::CREATED, Json(records)).into_response(),
        Err(e) => bad_request(e),
    }
}

async fn list_inbound_records(
    State(state): State<AppState>,
    Query(query): Query<ListQuery>,
) -> Response {
    let service = state.lock().await;
    let records = service.list_inbound_records(query.material_id);
    (StatusCode::OK, Json(records)).into_response()
}

async fn list_outbound_records(
    State(state): State<AppState>,
    Query(query): Query<ListQuery>,
) -> Response {
    let service = state.lock().await;
    let records = service.list_outbound_records(query.material_id);
    (StatusCode::OK, Json(records)).into_response()
}

async fn list_replenishment_orders(
    State(state): State<AppState>,
    Query(query): Query<ListQuery>,
) -> Response {
    let service = state.lock().await;
    let orders = service.list_replenishment_orders(query.material_id, query.status);
    (StatusCode::OK, Json(orders)).into_response()
}

async fn confirm_replenishment_order(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Response {
    let mut service = state.lock().await;
    match service.confirm_replenishment_order(id) {
        Ok(order) => (StatusCode::OK, Json(order)).into_response(),
        Err(e) => bad_request(e),
    }
}

async fn cancel_replenishment_order(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Response {
    let mut service = state.lock().await;
    match service.cancel_replenishment_order(id) {
        Ok(order) => (StatusCode::OK, Json(order)).into_response(),
        Err(e) => bad_request(e),
    }
}

#[tokio::main]
async fn main() {
    let cli = Cli::parse();

    let state: AppState = Arc::new(Mutex::new(WarehouseService::new()));

    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    let app = Router::new()
        .route("/materials", get(list_materials).post(create_material))
        .route("/materials/:id", get(get_material))
        .route("/inventories", get(list_inventories))
        .route("/inventories/:id", get(get_inventory))
        .route("/inbound", post(inbound))
        .route("/outbound", post(outbound))
        .route("/outbound/batch", post(batch_outbound))
        .route("/records/inbound", get(list_inbound_records))
        .route("/records/outbound", get(list_outbound_records))
        .route("/replenishment", get(list_replenishment_orders))
        .route(
            "/replenishment/:id/confirm",
            post(confirm_replenishment_order),
        )
        .route("/replenishment/:id/cancel", post(cancel_replenishment_order))
        .with_state(state)
        .layer(cors);

    let addr = format!("{}:{}", cli.host, cli.port);
    println!("仓库管理服务启动于 http://{}", addr);

    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
