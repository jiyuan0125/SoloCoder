use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use chrono::NaiveDate;
use clap::Parser;
use profit_sharing::{
    Order, OrderType, ProfitSharingService, Promoter, Settlement,
};
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use tokio::sync::Mutex;
use tower_http::cors::{Any, CorsLayer};
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "PORT", default_value_t = 3000)]
    port: u16,
    
    #[arg(short = 'H', long, env = "HOST", default_value = "0.0.0.0")]
    host: String,
}

type AppState = Arc<Mutex<ProfitSharingService>>;

#[derive(Debug, Serialize, Deserialize)]
struct CreateOrderRequest {
    order_type: OrderType,
    amount: u64,
    merchant_id: Uuid,
    promoter_id: Option<Uuid>,
}

#[derive(Debug, Serialize, Deserialize)]
struct CompleteOrderRequest {
    order_id: Uuid,
    completed_date: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct RegisterPromoterRequest {
    name: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct BindAccountRequest {
    promoter_id: Uuid,
}

#[derive(Debug, Serialize, Deserialize)]
struct ProcessSettlementRequest {
    settlement_date: String,
}

#[derive(Debug, Serialize)]
struct ApiResponse<T> {
    success: bool,
    data: Option<T>,
    error: Option<String>,
}

impl<T> ApiResponse<T> {
    fn success(data: T) -> Self {
        Self {
            success: true,
            data: Some(data),
            error: None,
        }
    }
    
    fn error(message: String) -> Self {
        Self {
            success: false,
            data: None,
            error: Some(message),
        }
    }
}

async fn create_order(
    State(state): State<AppState>,
    Json(request): Json<CreateOrderRequest>,
) -> impl IntoResponse {
    let mut service = state.lock().await;
    let order = service.create_order(
        request.order_type,
        request.amount,
        request.merchant_id,
        request.promoter_id,
    );
    (StatusCode::OK, Json(ApiResponse::success(order)))
}

async fn complete_order(
    State(state): State<AppState>,
    Json(request): Json<CompleteOrderRequest>,
) -> impl IntoResponse {
    let completed_date = match NaiveDate::parse_from_str(&request.completed_date, "%Y-%m-%d") {
        Ok(date) => date,
        Err(e) => {
            return (
                StatusCode::BAD_REQUEST,
                Json(ApiResponse::<Order>::error(format!("Invalid date: {}", e))),
            );
        }
    };
    
    let mut service = state.lock().await;
    match service.complete_order(request.order_id, completed_date) {
        Ok(order) => (StatusCode::OK, Json(ApiResponse::success(order))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(ApiResponse::error(e))),
    }
}

async fn get_order(
    State(state): State<AppState>,
    Path(order_id): Path<Uuid>,
) -> impl IntoResponse {
    let service = state.lock().await;
    match service.get_order(order_id) {
        Some(order) => (StatusCode::OK, Json(ApiResponse::success(order))),
        None => (
            StatusCode::NOT_FOUND,
            Json(ApiResponse::<Order>::error(format!("Order not found: {}", order_id))),
        ),
    }
}

async fn list_orders(State(state): State<AppState>) -> impl IntoResponse {
    let service = state.lock().await;
    let orders = service.get_all_orders();
    (StatusCode::OK, Json(ApiResponse::success(orders)))
}

async fn register_promoter(
    State(state): State<AppState>,
    Json(request): Json<RegisterPromoterRequest>,
) -> impl IntoResponse {
    let mut service = state.lock().await;
    let promoter = service.register_promoter(request.name);
    (StatusCode::OK, Json(ApiResponse::success(promoter)))
}

async fn bind_promoter_account(
    State(state): State<AppState>,
    Json(request): Json<BindAccountRequest>,
) -> impl IntoResponse {
    let mut service = state.lock().await;
    match service.bind_promoter_account(request.promoter_id) {
        Ok(pending_amount) => (
            StatusCode::OK,
            Json(ApiResponse::success(pending_amount)),
        ),
        Err(e) => (StatusCode::BAD_REQUEST, Json(ApiResponse::error(e))),
    }
}

async fn get_promoter(
    State(state): State<AppState>,
    Path(promoter_id): Path<Uuid>,
) -> impl IntoResponse {
    let service = state.lock().await;
    match service.get_promoter(promoter_id) {
        Some(promoter) => (StatusCode::OK, Json(ApiResponse::success(promoter))),
        None => (
            StatusCode::NOT_FOUND,
            Json(ApiResponse::<Promoter>::error(format!("Promoter not found: {}", promoter_id))),
        ),
    }
}

async fn list_promoters(State(state): State<AppState>) -> impl IntoResponse {
    let service = state.lock().await;
    let promoters = service.get_all_promoters();
    (StatusCode::OK, Json(ApiResponse::success(promoters)))
}

async fn process_settlement(
    State(state): State<AppState>,
    Json(request): Json<ProcessSettlementRequest>,
) -> impl IntoResponse {
    let settlement_date = match NaiveDate::parse_from_str(&request.settlement_date, "%Y-%m-%d") {
        Ok(date) => date,
        Err(e) => {
            return (
                StatusCode::BAD_REQUEST,
                Json(ApiResponse::<Settlement>::error(format!("Invalid date: {}", e))),
            );
        }
    };
    
    let mut service = state.lock().await;
    match service.process_settlement(settlement_date) {
        Ok(settlement) => (StatusCode::OK, Json(ApiResponse::success(settlement))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(ApiResponse::error(e))),
    }
}

async fn get_settlement(
    State(state): State<AppState>,
    Path(settlement_id): Path<Uuid>,
) -> impl IntoResponse {
    let service = state.lock().await;
    match service.get_settlement(settlement_id) {
        Some(settlement) => (StatusCode::OK, Json(ApiResponse::success(settlement))),
        None => (
            StatusCode::NOT_FOUND,
            Json(ApiResponse::<Settlement>::error(format!("Settlement not found: {}", settlement_id))),
        ),
    }
}

async fn list_settlements(State(state): State<AppState>) -> impl IntoResponse {
    let service = state.lock().await;
    let settlements = service.get_all_settlements();
    (StatusCode::OK, Json(ApiResponse::success(settlements)))
}

async fn get_order_types() -> impl IntoResponse {
    let types = vec![
        ("Standard", "标准订单 - 平台10%, 商户70%, 推广者20%"),
        ("Premium", "高级订单 - 平台15%, 商户60%, 推广者25%"),
        ("VIP", "VIP订单 - 平台20%, 商户50%, 推广者30%"),
    ];
    (StatusCode::OK, Json(ApiResponse::success(types)))
}

#[tokio::main]
async fn main() {
    let args = Args::parse();
    
    let state: AppState = Arc::new(Mutex::new(ProfitSharingService::new()));
    
    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);
    
    let app = Router::new()
        .route("/api/orders", post(create_order).get(list_orders))
        .route("/api/orders/:id", get(get_order))
        .route("/api/orders/complete", post(complete_order))
        .route("/api/promoters", post(register_promoter).get(list_promoters))
        .route("/api/promoters/:id", get(get_promoter))
        .route("/api/promoters/bind-account", post(bind_promoter_account))
        .route("/api/settlements", post(process_settlement).get(list_settlements))
        .route("/api/settlements/:id", get(get_settlement))
        .route("/api/order-types", get(get_order_types))
        .layer(cors)
        .with_state(state);
    
    let addr = format!("{}:{}", args.host, args.port);
    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    println!("Server listening on {}", addr);
    
    axum::serve(listener, app).await.unwrap();
}
