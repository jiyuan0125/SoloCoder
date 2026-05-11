use std::sync::Arc;
use axum::{
    extract::{State, Path},
    http::StatusCode,
    response::{IntoResponse, Response, Json},
    routing::{get, post},
    Router,
    Server,
};
use clap::Parser;
use serde::{Deserialize, Serialize};
use uuid::Uuid;
use chrono::NaiveDate;

use ticket_core::{
    errors::TicketError,
    models::CreateOrderRequest,
    create_in_memory_service, InMemoryTicketService,
};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "PORT", default_value = "3000")]
    port: u16,
    
    #[arg(short, long, env = "HOST", default_value = "0.0.0.0")]
    host: String,
}

#[derive(Debug, Serialize, Deserialize)]
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
    
    fn error(err: &TicketError) -> Self {
        Self {
            success: false,
            data: None,
            error: Some(err.to_string()),
        }
    }
}

fn error_to_response(err: TicketError) -> Response {
    let status = match &err {
        TicketError::ScenicNotFound => StatusCode::NOT_FOUND,
        TicketError::OrderNotFound => StatusCode::NOT_FOUND,
        TicketError::InvalidTicketType => StatusCode::BAD_REQUEST,
        TicketError::InvalidIdCard => StatusCode::BAD_REQUEST,
        TicketError::IdCardAlreadyPurchased => StatusCode::CONFLICT,
        TicketError::CapacityExceeded => StatusCode::SERVICE_UNAVAILABLE,
        TicketError::FreeChildrenExceeded => StatusCode::BAD_REQUEST,
        TicketError::InvalidDate => StatusCode::BAD_REQUEST,
        TicketError::RefundTimeExceeded => StatusCode::BAD_REQUEST,
        TicketError::OrderAlreadyRefunded => StatusCode::CONFLICT,
        TicketError::ConcurrencyConflict => StatusCode::CONFLICT,
        TicketError::Internal(_) => StatusCode::INTERNAL_SERVER_ERROR,
        TicketError::InvalidParam(_) => StatusCode::BAD_REQUEST,
    };
    
    (status, Json(ApiResponse::<()>::error(&err))).into_response()
}

type AppState = Arc<InMemoryTicketService>;

#[derive(Debug, Deserialize)]
struct CreateScenicRequest {
    name: String,
    base_price: u32,
    daily_capacity: u32,
}

async fn create_scenic(
    State(state): State<AppState>,
    Json(req): Json<CreateScenicRequest>,
) -> impl IntoResponse {
    match state.create_scenic(req.name, req.base_price, req.daily_capacity).await {
        Ok(scenic) => (StatusCode::OK, Json(ApiResponse::success(scenic))).into_response(),
        Err(err) => error_to_response(err),
    }
}

async fn list_scenics(
    State(state): State<AppState>,
) -> impl IntoResponse {
    match state.get_all_scenics().await {
        Ok(scenics) => (StatusCode::OK, Json(ApiResponse::success(scenics))).into_response(),
        Err(err) => error_to_response(err),
    }
}

async fn get_scenic(
    State(state): State<AppState>,
    Path(scenic_id): Path<Uuid>,
) -> impl IntoResponse {
    match state.get_scenic(&scenic_id).await {
        Ok(scenic) => (StatusCode::OK, Json(ApiResponse::success(scenic))).into_response(),
        Err(err) => error_to_response(err),
    }
}

async fn create_order(
    State(state): State<AppState>,
    Json(req): Json<CreateOrderRequest>,
) -> impl IntoResponse {
    match state.create_order(req).await {
        Ok(response) => (StatusCode::OK, Json(ApiResponse::success(response))).into_response(),
        Err(err) => error_to_response(err),
    }
}

async fn get_order(
    State(state): State<AppState>,
    Path(order_id): Path<Uuid>,
) -> impl IntoResponse {
    match state.get_order(&order_id).await {
        Ok(order) => (StatusCode::OK, Json(ApiResponse::success(order))).into_response(),
        Err(err) => error_to_response(err),
    }
}

async fn refund_order(
    State(state): State<AppState>,
    Path(order_id): Path<Uuid>,
) -> impl IntoResponse {
    match state.refund_order(&order_id).await {
        Ok(response) => (StatusCode::OK, Json(ApiResponse::success(response))).into_response(),
        Err(err) => error_to_response(err),
    }
}

async fn get_daily_stats(
    State(state): State<AppState>,
    Path((scenic_id, date)): Path<(Uuid, String)>,
) -> impl IntoResponse {
    let date = match NaiveDate::parse_from_str(&date, "%Y-%m-%d") {
        Ok(d) => d,
        Err(_) => return error_to_response(TicketError::InvalidDate),
    };
    
    match state.get_daily_stats(&scenic_id, &date).await {
        Ok(stats) => (StatusCode::OK, Json(ApiResponse::success(stats))).into_response(),
        Err(err) => error_to_response(err),
    }
}

async fn health_check() -> impl IntoResponse {
    (StatusCode::OK, Json(serde_json::json!({"status": "ok"}))).into_response()
}

fn create_router(service: AppState) -> Router {
    Router::new()
        .route("/health", get(health_check))
        .route("/scenics", post(create_scenic).get(list_scenics))
        .route("/scenics/:id", get(get_scenic))
        .route("/orders", post(create_order))
        .route("/orders/:id", get(get_order))
        .route("/orders/:id/refund", post(refund_order))
        .route("/stats/:scenic_id/:date", get(get_daily_stats))
        .with_state(service)
}

async fn seed_data(service: &InMemoryTicketService) {
    let _ = service.create_scenic("故宫博物院".into(), 100, 1000).await;
    let _ = service.create_scenic("颐和园".into(), 60, 2000).await;
    let _ = service.create_scenic("长城".into(), 80, 500).await;
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_env_filter(tracing_subscriber::EnvFilter::from_default_env())
        .init();

    let args = Args::parse();
    
    let service = create_in_memory_service();
    seed_data(&service).await;
    
    let app = create_router(service);
    
    let addr = format!("{}:{}", args.host, args.port);
    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    
    tracing::info!("Ticket booking server listening on {}", addr);
    
    Server::from_tcp(listener.into_std().unwrap()).unwrap()
        .serve(app.into_make_service())
        .await
        .unwrap();
}
