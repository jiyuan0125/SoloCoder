use std::sync::Arc;
use axum::{
    extract::State,
    http::StatusCode,
    response::{IntoResponse, Response},
    routing::{get, post},
    Json, Router,
};
use clap::Parser;
use serde::{Deserialize, Serialize};
use forex_core::{
    ForexSystem, CurrencyPair, OrderSide, ForexError,
};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "PORT", default_value_t = 3000)]
    port: u16,
    
    #[arg(long, env = "HOST", default_value = "127.0.0.1")]
    host: String,
}

type AppState = Arc<ForexSystem>;

#[derive(Debug, Serialize, Deserialize)]
struct CreateAccountRequest {
    name: String,
    is_reviewer: bool,
}

#[derive(Debug, Serialize, Deserialize)]
struct PlaceOrderRequest {
    account_id: String,
    pair: String,
    side: String,
    amount: f64,
}

#[derive(Debug, Serialize, Deserialize)]
struct ApproveOrderRequest {
    reviewer_id: String,
    order_id: String,
    comment: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
struct RejectOrderRequest {
    reviewer_id: String,
    order_id: String,
    comment: Option<String>,
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

    fn error(msg: String) -> Self {
        Self {
            success: false,
            data: None,
            error: Some(msg),
        }
    }
}

fn error_response(err: ForexError) -> Response {
    let status = match &err {
        ForexError::AccountNotFound(_) => StatusCode::NOT_FOUND,
        ForexError::CurrencyPairNotFound(_) => StatusCode::NOT_FOUND,
        ForexError::OrderNotFound(_) => StatusCode::NOT_FOUND,
        ForexError::InsufficientBalance => StatusCode::BAD_REQUEST,
        ForexError::InvalidAmount => StatusCode::BAD_REQUEST,
        ForexError::InvalidCurrency => StatusCode::BAD_REQUEST,
        ForexError::InvalidOrderStatus => StatusCode::BAD_REQUEST,
        ForexError::ApprovalRequired => StatusCode::BAD_REQUEST,
        ForexError::OrderRejected => StatusCode::BAD_REQUEST,
        ForexError::ConcurrentConflict => StatusCode::CONFLICT,
    };
    (status, Json(ApiResponse::<()>::error(err.to_string()))).into_response()
}

async fn create_account(
    State(state): State<AppState>,
    Json(req): Json<CreateAccountRequest>,
) -> Response {
    let account = state.create_account(req.name, req.is_reviewer);
    (StatusCode::OK, Json(ApiResponse::success(account))).into_response()
}

async fn get_account(
    State(state): State<AppState>,
    axum::extract::Path(account_id): axum::extract::Path<String>,
) -> Response {
    match state.get_account(&account_id) {
        Ok(account) => (StatusCode::OK, Json(ApiResponse::success(account))).into_response(),
        Err(err) => error_response(err),
    }
}

async fn get_balances(
    State(state): State<AppState>,
    axum::extract::Path(account_id): axum::extract::Path<String>,
) -> Response {
    match state.get_all_balances(&account_id) {
        Ok(balances) => (StatusCode::OK, Json(ApiResponse::success(balances))).into_response(),
        Err(err) => error_response(err),
    }
}

async fn get_rates(
    State(state): State<AppState>,
) -> Response {
    let rates = state.get_all_rates();
    (StatusCode::OK, Json(ApiResponse::success(rates))).into_response()
}

async fn place_order(
    State(state): State<AppState>,
    Json(req): Json<PlaceOrderRequest>,
) -> Response {
    let pair = match CurrencyPair::from_string(&req.pair) {
        Some(p) => p,
        None => return (StatusCode::BAD_REQUEST, Json(ApiResponse::<()>::error("无效的货币对格式".to_string()))).into_response(),
    };
    
    let side = match req.side.to_lowercase().as_str() {
        "buy" => OrderSide::Buy,
        "sell" => OrderSide::Sell,
        _ => return (StatusCode::BAD_REQUEST, Json(ApiResponse::<()>::error("无效的订单方向".to_string()))).into_response(),
    };
    
    match state.place_order(&req.account_id, pair, side, req.amount) {
        Ok(order) => (StatusCode::OK, Json(ApiResponse::success(order))).into_response(),
        Err(err) => error_response(err),
    }
}

async fn get_order(
    State(state): State<AppState>,
    axum::extract::Path(order_id): axum::extract::Path<String>,
) -> Response {
    match state.get_order(&order_id) {
        Ok(order) => (StatusCode::OK, Json(ApiResponse::success(order))).into_response(),
        Err(err) => error_response(err),
    }
}

async fn get_account_orders(
    State(state): State<AppState>,
    axum::extract::Path(account_id): axum::extract::Path<String>,
) -> Response {
    match state.get_orders_by_account(&account_id) {
        Ok(orders) => (StatusCode::OK, Json(ApiResponse::success(orders))).into_response(),
        Err(err) => error_response(err),
    }
}

async fn get_pending_approval(
    State(state): State<AppState>,
) -> Response {
    let orders = state.get_pending_approval_orders();
    (StatusCode::OK, Json(ApiResponse::success(orders))).into_response()
}

async fn approve_order(
    State(state): State<AppState>,
    Json(req): Json<ApproveOrderRequest>,
) -> Response {
    match state.approve_order(&req.reviewer_id, &req.order_id, req.comment) {
        Ok(order) => (StatusCode::OK, Json(ApiResponse::success(order))).into_response(),
        Err(err) => error_response(err),
    }
}

async fn reject_order(
    State(state): State<AppState>,
    Json(req): Json<RejectOrderRequest>,
) -> Response {
    match state.reject_order(&req.reviewer_id, &req.order_id, req.comment) {
        Ok(order) => (StatusCode::OK, Json(ApiResponse::success(order))).into_response(),
        Err(err) => error_response(err),
    }
}

async fn process_settlement(
    State(state): State<AppState>,
) -> Response {
    let count = state.process_settlement();
    (StatusCode::OK, Json(ApiResponse::success(serde_json::json!({"settled": count})))).into_response()
}

#[tokio::main]
async fn main() {
    let args = Args::parse();
    
    tracing_subscriber::fmt::init();
    
    let system = Arc::new(ForexSystem::new());
    
    let app = Router::new()
        .route("/health", get(|| async { "OK" }))
        .route("/api/accounts", post(create_account))
        .route("/api/accounts/:id", get(get_account))
        .route("/api/accounts/:id/balances", get(get_balances))
        .route("/api/accounts/:id/orders", get(get_account_orders))
        .route("/api/rates", get(get_rates))
        .route("/api/orders", post(place_order))
        .route("/api/orders/:id", get(get_order))
        .route("/api/orders/pending-approval", get(get_pending_approval))
        .route("/api/orders/approve", post(approve_order))
        .route("/api/orders/reject", post(reject_order))
        .route("/api/settlement/process", post(process_settlement))
        .with_state(system);
    
    let addr = format!("{}:{}", args.host, args.port);
    
    println!("外汇交易系统服务器已启动: http://{}", addr);
    println!("支持的货币对: USD/CNY, EUR/CNY, USD/JPY, EUR/USD");
    
    axum::Server::bind(&addr.parse().unwrap())
        .serve(app.into_make_service())
        .await
        .unwrap();
}
