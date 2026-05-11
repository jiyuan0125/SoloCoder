use std::sync::Arc;

use axum::{
    extract::State,
    http::StatusCode,
    routing::{get, post},
    Json, Router,
};
use clap::Parser;
use serde::{Deserialize, Serialize};
use tower_http::cors::{Any, CorsLayer};

use drug_trace_core::{
    DrugTraceService, DrugTraceError, InboundRequest, OutboundRequest, RecallRequest,
};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "DRUG_TRACE_PORT", default_value_t = 3000)]
    port: u16,

    #[arg(long, env = "DRUG_TRACE_HOST", default_value = "0.0.0.0")]
    host: String,
}

type AppState = Arc<DrugTraceService>;

#[derive(Debug, Serialize, Deserialize)]
struct ErrorResponse {
    error: String,
}

fn map_error(err: DrugTraceError) -> (StatusCode, Json<ErrorResponse>) {
    let status = match &err {
        DrugTraceError::DrugNotFound(_) => StatusCode::NOT_FOUND,
        DrugTraceError::BatchNotFound(_) => StatusCode::NOT_FOUND,
        DrugTraceError::InsufficientStock => StatusCode::BAD_REQUEST,
        DrugTraceError::InvalidQuantity => StatusCode::BAD_REQUEST,
        DrugTraceError::Expired => StatusCode::BAD_REQUEST,
        DrugTraceError::BatchFrozen => StatusCode::BAD_REQUEST,
        DrugTraceError::BatchRecalled => StatusCode::BAD_REQUEST,
        DrugTraceError::EmptyDrugName => StatusCode::BAD_REQUEST,
        DrugTraceError::EmptyBatchNo => StatusCode::BAD_REQUEST,
        DrugTraceError::EmptySupplier => StatusCode::BAD_REQUEST,
        DrugTraceError::InvalidDate => StatusCode::BAD_REQUEST,
    };

    (
        status,
        Json(ErrorResponse {
            error: err.to_string(),
        }),
    )
}

async fn inbound(
    State(state): State<AppState>,
    Json(req): Json<InboundRequest>,
) -> Result<Json<impl Serialize>, (StatusCode, Json<ErrorResponse>)> {
    state
        .inbound(req)
        .map(|batch| Json(batch))
        .map_err(map_error)
}

async fn outbound(
    State(state): State<AppState>,
    Json(req): Json<OutboundRequest>,
) -> Result<Json<impl Serialize>, (StatusCode, Json<ErrorResponse>)> {
    state
        .outbound(req)
        .map(|result| Json(result))
        .map_err(map_error)
}

async fn recall(
    State(state): State<AppState>,
    Json(req): Json<RecallRequest>,
) -> Result<Json<impl Serialize>, (StatusCode, Json<ErrorResponse>)> {
    state
        .recall(req)
        .map(|result| Json(result))
        .map_err(map_error)
}

async fn get_batches(
    State(state): State<AppState>,
) -> Result<Json<impl Serialize>, (StatusCode, Json<ErrorResponse>)> {
    Ok(Json(state.get_all_batches()))
}

async fn get_alerts(
    State(state): State<AppState>,
) -> Result<Json<impl Serialize>, (StatusCode, Json<ErrorResponse>)> {
    Ok(Json(state.get_expiry_alerts()))
}

async fn get_inbound_records(
    State(state): State<AppState>,
) -> Result<Json<impl Serialize>, (StatusCode, Json<ErrorResponse>)> {
    Ok(Json(state.get_inbound_records()))
}

async fn get_outbound_records(
    State(state): State<AppState>,
) -> Result<Json<impl Serialize>, (StatusCode, Json<ErrorResponse>)> {
    Ok(Json(state.get_outbound_records()))
}

async fn health() -> &'static str {
    "OK"
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();

    let args = Args::parse();

    let service = Arc::new(DrugTraceService::new());

    let cors = CorsLayer::new()
        .allow_methods(Any)
        .allow_origin(Any)
        .allow_headers(Any);

    let app = Router::new()
        .route("/health", get(health))
        .route("/api/batches", get(get_batches))
        .route("/api/inbound", post(inbound))
        .route("/api/outbound", post(outbound))
        .route("/api/recall", post(recall))
        .route("/api/alerts", get(get_alerts))
        .route("/api/records/inbound", get(get_inbound_records))
        .route("/api/records/outbound", get(get_outbound_records))
        .layer(cors)
        .with_state(service);

    let addr = format!("{}:{}", args.host, args.port).parse().unwrap();
    tracing::info!("Server listening on {}", addr);

    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}
