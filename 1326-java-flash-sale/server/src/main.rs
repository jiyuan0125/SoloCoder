use std::sync::Arc;

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use clap::Parser;
use serde::{Deserialize, Serialize};
use tokio::time::{sleep, Duration};

use flash_sale_core::error::FlashSaleError;
use flash_sale_core::models::{ActivityStatistics, FlashSaleActivity, Order};
use flash_sale_core::service::FlashSaleService;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "PORT", default_value_t = 3000)]
    port: u16,
}

#[derive(Clone)]
struct AppState {
    service: Arc<FlashSaleService>,
}

#[derive(Debug, Deserialize)]
struct CreateActivityRequest {
    product_name: String,
    flash_price: f64,
    stock: u32,
    start_time: chrono::DateTime<chrono::Utc>,
    end_time: chrono::DateTime<chrono::Utc>,
}

#[derive(Debug, Deserialize)]
struct CreateOrderRequest {
    user_id: String,
}

#[derive(Debug, Serialize)]
struct ErrorResponse {
    error: String,
}

struct AppError(FlashSaleError);

impl IntoResponse for AppError {
    fn into_response(self) -> axum::response::Response {
        let (status, message) = match self.0 {
            FlashSaleError::ActivityNotFound => (StatusCode::NOT_FOUND, "活动不存在".to_string()),
            FlashSaleError::ActivityNotStarted => (StatusCode::BAD_REQUEST, "活动还未开始".to_string()),
            FlashSaleError::ActivityEnded => (StatusCode::BAD_REQUEST, "活动已结束".to_string()),
            FlashSaleError::OutOfStock => (StatusCode::BAD_REQUEST, "库存不足".to_string()),
            FlashSaleError::UserAlreadyPurchased => (StatusCode::BAD_REQUEST, "用户已成功购买过该活动商品".to_string()),
            FlashSaleError::OrderNotFound => (StatusCode::NOT_FOUND, "订单不存在".to_string()),
            FlashSaleError::InvalidOrderStatus => (StatusCode::BAD_REQUEST, "订单状态不允许此操作".to_string()),
            FlashSaleError::InternalError(msg) => (StatusCode::INTERNAL_SERVER_ERROR, format!("系统错误: {}", msg)),
        };

        (status, Json(ErrorResponse { error: message })).into_response()
    }
}

impl From<FlashSaleError> for AppError {
    fn from(err: FlashSaleError) -> Self {
        AppError(err)
    }
}

async fn create_activity(
    State(state): State<AppState>,
    Json(req): Json<CreateActivityRequest>,
) -> Json<FlashSaleActivity> {
    let activity = state
        .service
        .create_activity(
            req.product_name,
            req.flash_price,
            req.stock,
            req.start_time,
            req.end_time,
        )
        .await;
    Json(activity)
}

async fn get_activity(
    State(state): State<AppState>,
    Path(activity_id): Path<String>,
) -> Result<Json<FlashSaleActivity>, AppError> {
    let activity = state.service.get_activity(&activity_id).await?;
    Ok(Json(activity))
}

async fn list_activities(State(state): State<AppState>) -> Json<Vec<FlashSaleActivity>> {
    let activities = state.service.list_activities().await;
    Json(activities)
}

async fn create_order(
    State(state): State<AppState>,
    Path(activity_id): Path<String>,
    Json(req): Json<CreateOrderRequest>,
) -> Result<Json<Order>, AppError> {
    let order = state
        .service
        .create_order(activity_id, req.user_id)
        .await?;
    Ok(Json(order))
}

async fn get_order(
    State(state): State<AppState>,
    Path(order_id): Path<String>,
) -> Result<Json<Order>, AppError> {
    let order = state.service.get_order(&order_id).await?;
    Ok(Json(order))
}

async fn list_orders(State(state): State<AppState>) -> Json<Vec<Order>> {
    let orders = state.service.list_orders().await;
    Json(orders)
}

async fn pay_order(
    State(state): State<AppState>,
    Path(order_id): Path<String>,
) -> Result<Json<Order>, AppError> {
    let order = state.service.pay_order(order_id).await?;
    Ok(Json(order))
}

async fn cancel_order(
    State(state): State<AppState>,
    Path(order_id): Path<String>,
) -> Result<Json<Order>, AppError> {
    let order = state.service.cancel_order(order_id).await?;
    Ok(Json(order))
}

async fn get_statistics(
    State(state): State<AppState>,
    Path(activity_id): Path<String>,
) -> Result<Json<ActivityStatistics>, AppError> {
    let stats = state.service.get_statistics(&activity_id).await?;
    Ok(Json(stats))
}

fn create_router(state: AppState) -> Router {
    Router::new()
        .route("/api/activities", post(create_activity).get(list_activities))
        .route("/api/activities/:id", get(get_activity))
        .route("/api/activities/:id/orders", post(create_order))
        .route("/api/activities/:id/statistics", get(get_statistics))
        .route("/api/orders", get(list_orders))
        .route("/api/orders/:id", get(get_order))
        .route("/api/orders/:id/pay", post(pay_order))
        .route("/api/orders/:id/cancel", post(cancel_order))
        .with_state(state)
}

#[tokio::main]
async fn main() {
    let args = Args::parse();
    let service = FlashSaleService::new();
    let state = AppState {
        service: service.clone(),
    };

    let router = create_router(state);

    let service_clone = service.clone();
    tokio::spawn(async move {
        loop {
            sleep(Duration::from_secs(10)).await;
            service_clone.process_expired_orders().await;
            service_clone.process_ended_activities().await;
        }
    });

    let addr = format!("0.0.0.0:{}", args.port);
    println!("Server listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    axum::Server::from_tcp(listener.into_std().unwrap())
        .unwrap()
        .serve(router.into_make_service())
        .await
        .unwrap();
}
