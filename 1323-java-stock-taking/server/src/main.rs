use std::sync::Arc;

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    Json, Router,
    routing::{get, post, put},
};
use clap::Parser;
use rust_decimal::Decimal;
use rust_decimal::prelude::FromPrimitive;
use stock_core::{
    ApproveDifferenceRequest, CreateCountRecordRequest, CreateBatchRequest, InventoryService,
    Material, ServiceConfig, StockError, SubmitDifferenceRequest,
};
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "PORT", default_value_t = 3000)]
    port: u16,

    #[arg(short, long, env = "HOST", default_value = "127.0.0.1")]
    host: String,

    #[arg(long, env = "AUTO_APPROVE_THRESHOLD", default_value_t = 100)]
    auto_approve_threshold: u64,
}

struct AppState {
    service: Arc<InventoryService>,
}

async fn health() -> impl IntoResponse {
    Json(serde_json::json!({"status": "ok"}))
}

async fn list_materials(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let materials = state.service.list_materials().await;
    Json(materials)
}

async fn create_material(
    State(state): State<Arc<AppState>>,
    Json(material): Json<Material>,
) -> impl IntoResponse {
    state.service.add_material(material).await;
    StatusCode::CREATED
}

async fn get_material(
    State(state): State<Arc<AppState>>,
    Path(code): Path<String>,
) -> impl IntoResponse {
    match state.service.get_material(&code).await {
        Some(m) => (StatusCode::OK, Json(serde_json::to_value(m).unwrap())),
        None => (
            StatusCode::NOT_FOUND,
            Json(serde_json::json!({"error": "物料不存在"})),
        ),
    }
}

async fn create_batch(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateBatchRequest>,
) -> impl IntoResponse {
    match state.service.create_batch(req).await {
        Ok(batch) => (StatusCode::CREATED, Json(serde_json::to_value(batch).unwrap())),
        Err(e) => handle_error(e),
    }
}

async fn list_batches(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let batches = state.service.list_batches().await;
    Json(batches)
}

async fn get_batch(
    State(state): State<Arc<AppState>>,
    Path(batch_no): Path<String>,
) -> impl IntoResponse {
    match state.service.get_batch(&batch_no).await {
        Some(b) => (StatusCode::OK, Json(serde_json::to_value(b).unwrap())),
        None => (
            StatusCode::NOT_FOUND,
            Json(serde_json::json!({"error": "批次不存在"})),
        ),
    }
}

async fn cancel_batch(
    State(state): State<Arc<AppState>>,
    Path(batch_no): Path<String>,
) -> impl IntoResponse {
    match state.service.cancel_batch(&batch_no).await {
        Ok(batch) => (StatusCode::OK, Json(serde_json::to_value(batch).unwrap())),
        Err(e) => handle_error(e),
    }
}

async fn complete_batch(
    State(state): State<Arc<AppState>>,
    Path(batch_no): Path<String>,
) -> impl IntoResponse {
    match state.service.complete_batch(&batch_no).await {
        Ok(batch) => (StatusCode::OK, Json(serde_json::to_value(batch).unwrap())),
        Err(e) => handle_error(e),
    }
}

async fn submit_count(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateCountRecordRequest>,
) -> impl IntoResponse {
    match state.service.submit_count(req).await {
        Ok(diff) => (StatusCode::CREATED, Json(serde_json::to_value(diff).unwrap())),
        Err(e) => handle_error(e),
    }
}

async fn list_batch_records(
    State(state): State<Arc<AppState>>,
    Path(batch_no): Path<String>,
) -> impl IntoResponse {
    let records = state.service.list_batch_records(&batch_no).await;
    Json(records)
}

async fn list_batch_differences(
    State(state): State<Arc<AppState>>,
    Path(batch_no): Path<String>,
) -> impl IntoResponse {
    let diffs = state.service.list_batch_differences(&batch_no).await;
    Json(diffs)
}

async fn get_difference(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.get_difference(&id).await {
        Some(d) => (StatusCode::OK, Json(serde_json::to_value(d).unwrap())),
        None => (
            StatusCode::NOT_FOUND,
            Json(serde_json::json!({"error": "差异不存在"})),
        ),
    }
}

async fn submit_difference(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
    Json(req): Json<SubmitDifferenceRequest>,
) -> impl IntoResponse {
    match state.service.submit_difference(&id, req).await {
        Ok(diff) => (StatusCode::OK, Json(serde_json::to_value(diff).unwrap())),
        Err(e) => handle_error(e),
    }
}

async fn approve_difference(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
    Json(req): Json<ApproveDifferenceRequest>,
) -> impl IntoResponse {
    match state.service.approve_difference(&id, req).await {
        Ok(diff) => (StatusCode::OK, Json(serde_json::to_value(diff).unwrap())),
        Err(e) => handle_error(e),
    }
}

fn handle_error(e: StockError) -> (StatusCode, Json<serde_json::Value>) {
    let status = match &e {
        StockError::MaterialNotFound(_) => StatusCode::NOT_FOUND,
        StockError::BatchNotFound(_) => StatusCode::NOT_FOUND,
        StockError::DifferenceNotFound(_) => StatusCode::NOT_FOUND,
        StockError::InvalidBatchState(_) => StatusCode::BAD_REQUEST,
        StockError::InvalidDifferenceState(_) => StatusCode::BAD_REQUEST,
        StockError::ReasonRequired => StatusCode::BAD_REQUEST,
        StockError::ConcurrencyConflict => StatusCode::CONFLICT,
        StockError::Internal(_) => StatusCode::INTERNAL_SERVER_ERROR,
    };
    (status, Json(serde_json::json!({"error": e.to_string()})))
}

fn router(state: Arc<AppState>) -> Router {
    Router::new()
        .route("/health", get(health))
        .route("/materials", get(list_materials).post(create_material))
        .route("/materials/:code", get(get_material))
        .route("/batches", get(list_batches).post(create_batch))
        .route("/batches/:batch_no", get(get_batch))
        .route("/batches/:batch_no/cancel", put(cancel_batch))
        .route("/batches/:batch_no/complete", put(complete_batch))
        .route("/batches/:batch_no/records", get(list_batch_records))
        .route("/batches/:batch_no/differences", get(list_batch_differences))
        .route("/counts", post(submit_count))
        .route("/differences/:id", get(get_difference))
        .route("/differences/:id/submit", put(submit_difference))
        .route("/differences/:id/approve", put(approve_difference))
        .with_state(state)
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();

    let args = Args::parse();

    let threshold = Decimal::from_u64(args.auto_approve_threshold)
        .expect("Invalid threshold");

    let config = ServiceConfig {
        auto_approve_threshold: threshold,
    };

    let service = InventoryService::new(config);

    let state = Arc::new(AppState { service });

    let app = router(state);

    let addr = format!("{}:{}", args.host, args.port);
    let addr = addr.parse().unwrap();

    tracing::info!("服务器启动，监听地址: {}", addr);

    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}
