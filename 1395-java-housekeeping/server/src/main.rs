use std::sync::Arc;

use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    response::{IntoResponse, Json},
    routing::{get, post},
    Router,
};
use clap::Parser;
use serde::Deserialize;
use tower_http::cors::{Any, CorsLayer};

use housekeeping_core::{
    ApproveRefundRequest, BookingService, CreateAuntRequest, CreateCustomerRequest,
    CreateOrderRequest, InMemoryStore, RejectRefundRequest, RequestRefundRequest,
    SkillType, TimeSlot,
};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "HOUSEKEEPING_PORT", default_value_t = 8080)]
    port: u16,
}

#[derive(Clone)]
struct AppState {
    service: Arc<BookingService>,
}

#[derive(Debug, Deserialize)]
struct RecommendQuery {
    skill_type: SkillType,
    date: String,
    start_time: String,
    end_time: String,
}

#[derive(Debug, Deserialize)]
struct EarningsQuery {
    year: i32,
    month: u32,
}

async fn create_aunt(
    State(state): State<AppState>,
    Json(req): Json<CreateAuntRequest>,
) -> impl IntoResponse {
    match state.service.create_aunt(req).await {
        Ok(aunt) => (StatusCode::CREATED, Json(serde_json::json!({
            "success": true,
            "data": aunt
        }))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": e.to_string()
        }))),
    }
}

async fn get_aunt(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    let uuid = match uuid::Uuid::parse_str(&id) {
        Ok(u) => u,
        Err(_) => return (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": "无效的UUID格式"
        }))),
    };

    match state.service.get_aunt(&uuid).await {
        Ok(aunt) => (StatusCode::OK, Json(serde_json::json!({
            "success": true,
            "data": aunt
        }))),
        Err(e) => (StatusCode::NOT_FOUND, Json(serde_json::json!({
            "success": false,
            "error": e.to_string()
        }))),
    }
}

async fn get_all_aunts(State(state): State<AppState>) -> impl IntoResponse {
    match state.service.get_all_aunts().await {
        Ok(aunts) => (StatusCode::OK, Json(serde_json::json!({
            "success": true,
            "data": aunts
        }))),
        Err(e) => (StatusCode::INTERNAL_SERVER_ERROR, Json(serde_json::json!({
            "success": false,
            "error": e.to_string()
        }))),
    }
}

async fn create_customer(
    State(state): State<AppState>,
    Json(req): Json<CreateCustomerRequest>,
) -> impl IntoResponse {
    match state.service.create_customer(req).await {
        Ok(customer) => (StatusCode::CREATED, Json(serde_json::json!({
            "success": true,
            "data": customer
        }))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": e.to_string()
        }))),
    }
}

async fn get_customer(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    let uuid = match uuid::Uuid::parse_str(&id) {
        Ok(u) => u,
        Err(_) => return (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": "无效的UUID格式"
        }))),
    };

    match state.service.get_customer(&uuid).await {
        Ok(customer) => (StatusCode::OK, Json(serde_json::json!({
            "success": true,
            "data": customer
        }))),
        Err(e) => (StatusCode::NOT_FOUND, Json(serde_json::json!({
            "success": false,
            "error": e.to_string()
        }))),
    }
}

async fn get_all_customers(State(state): State<AppState>) -> impl IntoResponse {
    match state.service.get_all_customers().await {
        Ok(customers) => (StatusCode::OK, Json(serde_json::json!({
            "success": true,
            "data": customers
        }))),
        Err(e) => (StatusCode::INTERNAL_SERVER_ERROR, Json(serde_json::json!({
            "success": false,
            "error": e.to_string()
        }))),
    }
}

async fn recommend_aunts(
    State(state): State<AppState>,
    Query(query): Query<RecommendQuery>,
) -> impl IntoResponse {
    let date = match chrono::NaiveDate::parse_from_str(&query.date, "%Y-%m-%d") {
        Ok(d) => d,
        Err(_) => return (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": "无效的日期格式，应为YYYY-MM-DD"
        }))),
    };

    let start_time = match chrono::NaiveTime::parse_from_str(&query.start_time, "%H:%M:%S") {
        Ok(t) => t,
        Err(_) => return (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": "无效的开始时间格式，应为HH:MM:SS"
        }))),
    };

    let end_time = match chrono::NaiveTime::parse_from_str(&query.end_time, "%H:%M:%S") {
        Ok(t) => t,
        Err(_) => return (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": "无效的结束时间格式，应为HH:MM:SS"
        }))),
    };

    let time_slot = TimeSlot::new(date, start_time, end_time);

    match state.service.recommend_aunts(query.skill_type, &time_slot).await {
        Ok(recommendations) => (StatusCode::OK, Json(serde_json::json!({
            "success": true,
            "data": recommendations
        }))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": e.to_string()
        }))),
    }
}

async fn create_order(
    State(state): State<AppState>,
    Json(req): Json<CreateOrderRequest>,
) -> impl IntoResponse {
    match state.service.create_order(req).await {
        Ok(order) => (StatusCode::CREATED, Json(serde_json::json!({
            "success": true,
            "data": order
        }))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": e.to_string()
        }))),
    }
}

async fn get_order(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    let uuid = match uuid::Uuid::parse_str(&id) {
        Ok(u) => u,
        Err(_) => return (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": "无效的UUID格式"
        }))),
    };

    match state.service.get_order(&uuid).await {
        Ok(order) => (StatusCode::OK, Json(serde_json::json!({
            "success": true,
            "data": order
        }))),
        Err(e) => (StatusCode::NOT_FOUND, Json(serde_json::json!({
            "success": false,
            "error": e.to_string()
        }))),
    }
}

async fn get_all_orders(State(state): State<AppState>) -> impl IntoResponse {
    match state.service.get_all_orders().await {
        Ok(orders) => (StatusCode::OK, Json(serde_json::json!({
            "success": true,
            "data": orders
        }))),
        Err(e) => (StatusCode::INTERNAL_SERVER_ERROR, Json(serde_json::json!({
            "success": false,
            "error": e.to_string()
        }))),
    }
}

async fn confirm_order(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    let uuid = match uuid::Uuid::parse_str(&id) {
        Ok(u) => u,
        Err(_) => return (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": "无效的UUID格式"
        }))),
    };

    match state.service.confirm_order(&uuid).await {
        Ok(order) => (StatusCode::OK, Json(serde_json::json!({
            "success": true,
            "data": order
        }))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": e.to_string()
        }))),
    }
}

async fn start_service(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    let uuid = match uuid::Uuid::parse_str(&id) {
        Ok(u) => u,
        Err(_) => return (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": "无效的UUID格式"
        }))),
    };

    match state.service.start_service(&uuid).await {
        Ok(order) => (StatusCode::OK, Json(serde_json::json!({
            "success": true,
            "data": order
        }))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": e.to_string()
        }))),
    }
}

async fn complete_order(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    let uuid = match uuid::Uuid::parse_str(&id) {
        Ok(u) => u,
        Err(_) => return (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": "无效的UUID格式"
        }))),
    };

    match state.service.complete_order(&uuid).await {
        Ok(order) => (StatusCode::OK, Json(serde_json::json!({
            "success": true,
            "data": order
        }))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": e.to_string()
        }))),
    }
}

async fn cancel_order(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    let uuid = match uuid::Uuid::parse_str(&id) {
        Ok(u) => u,
        Err(_) => return (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": "无效的UUID格式"
        }))),
    };

    match state.service.cancel_order(&uuid).await {
        Ok(order) => (StatusCode::OK, Json(serde_json::json!({
            "success": true,
            "data": order
        }))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": e.to_string()
        }))),
    }
}

async fn request_refund(
    State(state): State<AppState>,
    Json(req): Json<RequestRefundRequest>,
) -> impl IntoResponse {
    match state.service.request_refund(req).await {
        Ok(order) => (StatusCode::CREATED, Json(serde_json::json!({
            "success": true,
            "data": order
        }))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": e.to_string()
        }))),
    }
}

async fn approve_refund(
    State(state): State<AppState>,
    Json(req): Json<ApproveRefundRequest>,
) -> impl IntoResponse {
    match state.service.approve_refund(req).await {
        Ok(order) => (StatusCode::OK, Json(serde_json::json!({
            "success": true,
            "data": order
        }))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": e.to_string()
        }))),
    }
}

async fn reject_refund(
    State(state): State<AppState>,
    Json(req): Json<RejectRefundRequest>,
) -> impl IntoResponse {
    match state.service.reject_refund(req).await {
        Ok(order) => (StatusCode::OK, Json(serde_json::json!({
            "success": true,
            "data": order
        }))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": e.to_string()
        }))),
    }
}

async fn get_monthly_earnings(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Query(query): Query<EarningsQuery>,
) -> impl IntoResponse {
    let uuid = match uuid::Uuid::parse_str(&id) {
        Ok(u) => u,
        Err(_) => return (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": "无效的UUID格式"
        }))),
    };

    match state.service.get_monthly_earnings(&uuid, query.year, query.month).await {
        Ok(earnings) => (StatusCode::OK, Json(serde_json::json!({
            "success": true,
            "data": earnings
        }))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": e.to_string()
        }))),
    }
}

async fn get_skills() -> impl IntoResponse {
    let skills = SkillType::all()
        .into_iter()
        .map(|s| serde_json::json!({
            "type": s,
            "name": s.as_str()
        }))
        .collect::<Vec<_>>();

    (StatusCode::OK, Json(serde_json::json!({
        "success": true,
        "data": skills
    })))
}

fn create_router(state: AppState) -> Router {
    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    Router::new()
        .route("/api/skills", get(get_skills))
        .route("/api/aunts", post(create_aunt).get(get_all_aunts))
        .route("/api/aunts/:id", get(get_aunt))
        .route("/api/aunts/:id/earnings", get(get_monthly_earnings))
        .route("/api/aunts/recommend", get(recommend_aunts))
        .route("/api/customers", post(create_customer).get(get_all_customers))
        .route("/api/customers/:id", get(get_customer))
        .route("/api/orders", post(create_order).get(get_all_orders))
        .route("/api/orders/:id", get(get_order))
        .route("/api/orders/:id/confirm", post(confirm_order))
        .route("/api/orders/:id/start", post(start_service))
        .route("/api/orders/:id/complete", post(complete_order))
        .route("/api/orders/:id/cancel", post(cancel_order))
        .route("/api/refunds/request", post(request_refund))
        .route("/api/refunds/approve", post(approve_refund))
        .route("/api/refunds/reject", post(reject_refund))
        .layer(cors)
        .with_state(state)
}

#[tokio::main]
async fn main() {
    let args = Args::parse();
    
    let store = InMemoryStore::new();
    let service = BookingService::new(store);
    let state = AppState {
        service: Arc::new(service),
    };

    let router = create_router(state);
    let addr = std::net::SocketAddr::from(([0, 0, 0, 0], args.port));
    
    println!("家政服务平台管理系统启动中...");
    println!("服务端口: {}", args.port);
    println!("API文档:");
    println!("  - GET  /api/skills                    获取技能列表");
    println!("  - POST /api/aunts                     创建阿姨");
    println!("  - GET  /api/aunts                     获取所有阿姨");
    println!("  - GET  /api/aunts/:id                 获取阿姨详情");
    println!("  - GET  /api/aunts/recommend           推荐阿姨");
    println!("  - POST /api/customers                 创建客户");
    println!("  - GET  /api/customers                 获取所有客户");
    println!("  - POST /api/orders                    创建订单");
    println!("  - GET  /api/orders                    获取所有订单");
    println!("  - POST /api/orders/:id/confirm        确认订单");
    println!("  - POST /api/orders/:id/start          开始服务");
    println!("  - POST /api/orders/:id/complete       完成订单");
    println!("  - POST /api/orders/:id/cancel         取消订单");
    println!("  - POST /api/refunds/request           申请退款");
    println!("  - POST /api/refunds/approve           批准退款");
    println!("  - POST /api/refunds/reject            拒绝退款");
    println!("  - GET  /api/aunts/:id/earnings        月度收入");
    println!();

    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    axum::serve(listener, router).await.unwrap();
}
