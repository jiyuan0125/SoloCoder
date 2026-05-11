use std::sync::Arc;
use axum::{
    extract::{Path, State},
    http::StatusCode,
    Json,
    response::IntoResponse,
    routing::{get, post},
    Router,
};
use clap::Parser;
use core_lib::{
    models::{ConsumeBreakdown, Member, MemberLevel, Transaction},
    services::StoredValueService,
    storage::InMemoryStorage,
    StoredValueError,
};
use serde::{Deserialize, Serialize};
use tower_http::cors::CorsLayer;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "SV_PORT", default_value_t = 3000)]
    port: u16,
    
    #[arg(short, long, env = "SV_HOST", default_value = "0.0.0.0")]
    host: String,
}

#[derive(Clone)]
struct AppState {
    service: Arc<StoredValueService>,
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateMemberRequest {
    name: String,
    level: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct RechargeRequest {
    amount: u64,
}

#[derive(Debug, Serialize, Deserialize)]
struct ConsumeRequest {
    original_price: u64,
    description: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct RefundRequest {
    consume_transaction_id: String,
}

#[derive(Debug, Serialize)]
struct ApiResponse {
    success: bool,
    data: Option<serde_json::Value>,
    error: Option<String>,
}

impl ApiResponse {
    fn success<T: Serialize>(data: T) -> Self {
        Self {
            success: true,
            data: Some(serde_json::to_value(data).unwrap()),
            error: None,
        }
    }
    
    fn error(err: &str) -> Self {
        Self {
            success: false,
            data: None,
            error: Some(err.to_string()),
        }
    }
}

#[derive(Debug, Serialize)]
struct ConsumeResponse {
    member: Member,
    transaction: Transaction,
    breakdown: ConsumeBreakdown,
}

#[derive(Debug, Serialize)]
struct RechargeResponse {
    member: Member,
    transactions: Vec<Transaction>,
}

fn parse_level(level_str: &str) -> Result<MemberLevel, String> {
    match level_str.to_lowercase().as_str() {
        "regular" | "普通" => Ok(MemberLevel::Regular),
        "silver" | "银卡" => Ok(MemberLevel::Silver),
        "gold" | "金卡" => Ok(MemberLevel::Gold),
        "diamond" | "钻石" => Ok(MemberLevel::Diamond),
        _ => Err(format!("无效的会员等级: {}", level_str)),
    }
}

async fn create_member(
    State(state): State<AppState>,
    Json(req): Json<CreateMemberRequest>,
) -> impl IntoResponse {
    let level = match parse_level(&req.level) {
        Ok(l) => l,
        Err(e) => return (StatusCode::BAD_REQUEST, Json(ApiResponse::error(&e))),
    };
    
    let member = state.service.create_member(req.name, level).await;
    (StatusCode::OK, Json(ApiResponse::success(member)))
}

async fn get_member(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match state.service.get_member(&id).await {
        Ok(member) => (StatusCode::OK, Json(ApiResponse::success(member))),
        Err(e) => {
            let status = match &e {
                StoredValueError::MemberNotFound(_) => StatusCode::NOT_FOUND,
                StoredValueError::TransactionNotFound(_) => StatusCode::NOT_FOUND,
                StoredValueError::InsufficientBalance(..) => StatusCode::BAD_REQUEST,
                StoredValueError::RefundAmountExceeds(..) => StatusCode::BAD_REQUEST,
                StoredValueError::InvalidRechargeAmount => StatusCode::BAD_REQUEST,
                StoredValueError::InvalidConsumeAmount => StatusCode::BAD_REQUEST,
                StoredValueError::InternalError(_) => StatusCode::INTERNAL_SERVER_ERROR,
            };
            (status, Json(ApiResponse::error(&e.to_string())))
        }
    }
}

async fn list_members(State(state): State<AppState>) -> impl IntoResponse {
    let members = state.service.list_members().await;
    (StatusCode::OK, Json(ApiResponse::success(members)))
}

async fn recharge(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(req): Json<RechargeRequest>,
) -> impl IntoResponse {
    match state.service.recharge(&id, req.amount).await {
        Ok((member, transactions)) => {
            let response = RechargeResponse { member, transactions };
            (StatusCode::OK, Json(ApiResponse::success(response)))
        }
        Err(e) => {
            let status = match &e {
                StoredValueError::MemberNotFound(_) => StatusCode::NOT_FOUND,
                StoredValueError::TransactionNotFound(_) => StatusCode::NOT_FOUND,
                StoredValueError::InsufficientBalance(..) => StatusCode::BAD_REQUEST,
                StoredValueError::RefundAmountExceeds(..) => StatusCode::BAD_REQUEST,
                StoredValueError::InvalidRechargeAmount => StatusCode::BAD_REQUEST,
                StoredValueError::InvalidConsumeAmount => StatusCode::BAD_REQUEST,
                StoredValueError::InternalError(_) => StatusCode::INTERNAL_SERVER_ERROR,
            };
            (status, Json(ApiResponse::error(&e.to_string())))
        }
    }
}

async fn consume(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(req): Json<ConsumeRequest>,
) -> impl IntoResponse {
    match state.service.consume(&id, req.original_price, req.description).await {
        Ok((member, transaction, breakdown)) => {
            let response = ConsumeResponse { member, transaction, breakdown };
            (StatusCode::OK, Json(ApiResponse::success(response)))
        }
        Err(e) => {
            let status = match &e {
                StoredValueError::MemberNotFound(_) => StatusCode::NOT_FOUND,
                StoredValueError::TransactionNotFound(_) => StatusCode::NOT_FOUND,
                StoredValueError::InsufficientBalance(..) => StatusCode::BAD_REQUEST,
                StoredValueError::RefundAmountExceeds(..) => StatusCode::BAD_REQUEST,
                StoredValueError::InvalidRechargeAmount => StatusCode::BAD_REQUEST,
                StoredValueError::InvalidConsumeAmount => StatusCode::BAD_REQUEST,
                StoredValueError::InternalError(_) => StatusCode::INTERNAL_SERVER_ERROR,
            };
            (status, Json(ApiResponse::error(&e.to_string())))
        }
    }
}

async fn refund(
    State(state): State<AppState>,
    Path(_id): Path<String>,
    Json(req): Json<RefundRequest>,
) -> impl IntoResponse {
    match state.service.refund(&req.consume_transaction_id).await {
        Ok((member, transaction)) => {
            let response = (member, transaction);
            (StatusCode::OK, Json(ApiResponse::success(response)))
        }
        Err(e) => {
            let status = match &e {
                StoredValueError::MemberNotFound(_) => StatusCode::NOT_FOUND,
                StoredValueError::TransactionNotFound(_) => StatusCode::NOT_FOUND,
                StoredValueError::InsufficientBalance(..) => StatusCode::BAD_REQUEST,
                StoredValueError::RefundAmountExceeds(..) => StatusCode::BAD_REQUEST,
                StoredValueError::InvalidRechargeAmount => StatusCode::BAD_REQUEST,
                StoredValueError::InvalidConsumeAmount => StatusCode::BAD_REQUEST,
                StoredValueError::InternalError(_) => StatusCode::INTERNAL_SERVER_ERROR,
            };
            (status, Json(ApiResponse::error(&e.to_string())))
        }
    }
}

async fn list_transactions(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match state.service.list_member_transactions(&id).await {
        Ok(transactions) => (StatusCode::OK, Json(ApiResponse::success(transactions))),
        Err(e) => {
            let status = match &e {
                StoredValueError::MemberNotFound(_) => StatusCode::NOT_FOUND,
                StoredValueError::TransactionNotFound(_) => StatusCode::NOT_FOUND,
                StoredValueError::InsufficientBalance(..) => StatusCode::BAD_REQUEST,
                StoredValueError::RefundAmountExceeds(..) => StatusCode::BAD_REQUEST,
                StoredValueError::InvalidRechargeAmount => StatusCode::BAD_REQUEST,
                StoredValueError::InvalidConsumeAmount => StatusCode::BAD_REQUEST,
                StoredValueError::InternalError(_) => StatusCode::INTERNAL_SERVER_ERROR,
            };
            (status, Json(ApiResponse::error(&e.to_string())))
        }
    }
}

#[tokio::main]
async fn main() {
    let args = Args::parse();
    
    let storage = Arc::new(InMemoryStorage::new());
    let service = Arc::new(StoredValueService::new(storage));
    let state = AppState { service };
    
    let app = Router::new()
        .route("/members", get(list_members).post(create_member))
        .route("/members/:id", get(get_member))
        .route("/members/:id/recharge", post(recharge))
        .route("/members/:id/consume", post(consume))
        .route("/members/:id/refund", post(refund))
        .route("/members/:id/transactions", get(list_transactions))
        .layer(CorsLayer::permissive())
        .with_state(state);
    
    let addr = format!("{}:{}", args.host, args.port);
    println!("会员储值卡系统服务器启动中...");
    println!("监听地址: {}", addr);
    println!("端口配置可以通过 --port 参数或 SV_PORT 环境变量指定");
    
    axum::Server::bind(&addr.parse().unwrap())
        .serve(app.into_make_service())
        .await
        .unwrap();
}
