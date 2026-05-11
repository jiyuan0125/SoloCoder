use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::Json,
    routing::{get, post},
    Router,
};
use clap::Parser;
use delivery_dispatch_core::{
    DispatchConfig, DispatchState, Location, Order, OrderStatus, Rider, RiderStatus, Merchant,
};
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use tokio::net::TcpListener;
use tower_http::cors::{Any, CorsLayer};
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(short, long, env = "SERVER_PORT", default_value_t = 3000)]
    port: u16,

    #[arg(long, env = "MAX_ORDERS_PER_RIDER", default_value_t = 3)]
    max_orders_per_rider: usize,

    #[arg(long, env = "MAX_DETOUR_DISTANCE_KM", default_value_t = 2.0)]
    max_detour_distance_km: f64,
}

type AppState = Arc<DispatchState>;

#[derive(Debug, Serialize, Deserialize)]
struct CreateRiderRequest {
    name: String,
    location: Location,
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateMerchantRequest {
    name: String,
    location: Location,
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateOrderRequest {
    merchant_id: Uuid,
}

#[derive(Debug, Serialize, Deserialize)]
struct UpdateRiderLocationRequest {
    location: Location,
}

#[derive(Debug, Serialize)]
struct ApiResponse<T> {
    success: bool,
    data: Option<T>,
    message: Option<String>,
}

impl<T> ApiResponse<T> {
    fn success(data: T) -> Self {
        Self {
            success: true,
            data: Some(data),
            message: None,
        }
    }

    fn error(message: String) -> Self {
        Self {
            success: false,
            data: None,
            message: Some(message),
        }
    }
}

async fn create_rider(
    State(state): State<AppState>,
    Json(payload): Json<CreateRiderRequest>,
) -> Result<Json<ApiResponse<Rider>>, (StatusCode, Json<ApiResponse<Rider>>)> {
    let rider = state.add_rider(payload.name, payload.location).await;
    Ok(Json(ApiResponse::success(rider)))
}

async fn list_riders(
    State(state): State<AppState>,
) -> Result<Json<ApiResponse<Vec<Rider>>>, (StatusCode, Json<ApiResponse<Vec<Rider>>>)> {
    let riders = state.list_riders().await;
    Ok(Json(ApiResponse::success(riders)))
}

async fn get_rider(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<ApiResponse<Rider>>, (StatusCode, Json<ApiResponse<Rider>>)> {
    match state.get_rider(id).await {
        Some(rider) => Ok(Json(ApiResponse::success(rider))),
        None => Err((
            StatusCode::NOT_FOUND,
            Json(ApiResponse::error("Rider not found".to_string())),
        )),
    }
}

async fn update_rider_location(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(payload): Json<UpdateRiderLocationRequest>,
) -> Result<Json<ApiResponse<Rider>>, (StatusCode, Json<ApiResponse<Rider>>)> {
    match state.update_rider_location(id, payload.location).await {
        Some(rider) => Ok(Json(ApiResponse::success(rider))),
        None => Err((
            StatusCode::NOT_FOUND,
            Json(ApiResponse::error("Rider not found".to_string())),
        )),
    }
}

async fn create_merchant(
    State(state): State<AppState>,
    Json(payload): Json<CreateMerchantRequest>,
) -> Result<Json<ApiResponse<Merchant>>, (StatusCode, Json<ApiResponse<Merchant>>)> {
    let merchant = state.add_merchant(payload.name, payload.location).await;
    Ok(Json(ApiResponse::success(merchant)))
}

async fn list_merchants(
    State(state): State<AppState>,
) -> Result<Json<ApiResponse<Vec<Merchant>>>, (StatusCode, Json<ApiResponse<Vec<Merchant>>>)> {
    let merchants = state.list_merchants().await;
    Ok(Json(ApiResponse::success(merchants)))
}

async fn get_merchant(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<ApiResponse<Merchant>>, (StatusCode, Json<ApiResponse<Merchant>>)> {
    match state.get_merchant(id).await {
        Some(merchant) => Ok(Json(ApiResponse::success(merchant))),
        None => Err((
            StatusCode::NOT_FOUND,
            Json(ApiResponse::error("Merchant not found".to_string())),
        )),
    }
}

async fn create_order(
    State(state): State<AppState>,
    Json(payload): Json<CreateOrderRequest>,
) -> Result<Json<ApiResponse<Order>>, (StatusCode, Json<ApiResponse<Order>>)> {
    match state.create_order(payload.merchant_id).await {
        Some(order) => Ok(Json(ApiResponse::success(order))),
        None => Err((
            StatusCode::NOT_FOUND,
            Json(ApiResponse::error("Merchant not found".to_string())),
        )),
    }
}

async fn list_orders(
    State(state): State<AppState>,
) -> Result<Json<ApiResponse<Vec<Order>>>, (StatusCode, Json<ApiResponse<Vec<Order>>>)> {
    let orders = state.list_orders().await;
    Ok(Json(ApiResponse::success(orders)))
}

async fn get_order(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<ApiResponse<Order>>, (StatusCode, Json<ApiResponse<Order>>)> {
    match state.get_order(id).await {
        Some(order) => Ok(Json(ApiResponse::success(order))),
        None => Err((
            StatusCode::NOT_FOUND,
            Json(ApiResponse::error("Order not found".to_string())),
        )),
    }
}

async fn complete_order(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<ApiResponse<Order>>, (StatusCode, Json<ApiResponse<Order>>)> {
    match state.complete_order(id).await {
        Some(order) => Ok(Json(ApiResponse::success(order))),
        None => Err((
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::error(
                "Order not found or not in assignable state".to_string(),
            )),
        )),
    }
}

async fn get_waiting_queue(
    State(state): State<AppState>,
) -> Result<Json<ApiResponse<Vec<Uuid>>>, (StatusCode, Json<ApiResponse<Vec<Uuid>>>)> {
    let queue = state.get_waiting_queue().await;
    Ok(Json(ApiResponse::success(queue)))
}

async fn get_stats(
    State(state): State<AppState>,
) -> Result<Json<ApiResponse<serde_json::Value>>, (StatusCode, Json<ApiResponse<serde_json::Value>>)> {
    let riders = state.list_riders().await;
    let orders = state.list_orders().await;
    let waiting_queue = state.get_waiting_queue().await;

    let idle_riders = riders
        .iter()
        .filter(|r| r.status == RiderStatus::Idle)
        .count();
    let busy_riders = riders.len() - idle_riders;

    let pending_orders = orders
        .iter()
        .filter(|o| o.status == OrderStatus::Pending)
        .count();
    let assigned_orders = orders
        .iter()
        .filter(|o| o.status == OrderStatus::Assigned)
        .count();
    let delivered_orders = orders
        .iter()
        .filter(|o| o.status == OrderStatus::Delivered)
        .count();

    let stats = serde_json::json!({
        "riders": {
            "total": riders.len(),
            "idle": idle_riders,
            "busy": busy_riders,
        },
        "orders": {
            "total": orders.len(),
            "pending": pending_orders,
            "assigned": assigned_orders,
            "delivered": delivered_orders,
        },
        "waiting_queue": waiting_queue.len(),
    });

    Ok(Json(ApiResponse::success(stats)))
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let cli = Cli::parse();

    let config = DispatchConfig {
        max_orders_per_rider: cli.max_orders_per_rider,
        max_detour_distance_km: cli.max_detour_distance_km,
    };

    let state = Arc::new(DispatchState::with_config(config));

    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    let app = Router::new()
        .route("/riders", post(create_rider).get(list_riders))
        .route("/riders/:id", get(get_rider).put(update_rider_location))
        .route("/merchants", post(create_merchant).get(list_merchants))
        .route("/merchants/:id", get(get_merchant))
        .route("/orders", post(create_order).get(list_orders))
        .route("/orders/:id", get(get_order).post(complete_order))
        .route("/waiting-queue", get(get_waiting_queue))
        .route("/stats", get(get_stats))
        .with_state(state)
        .layer(cors);

    let addr = format!("0.0.0.0:{}", cli.port);
    println!("Server listening on {}", addr);
    let listener = TcpListener::bind(addr).await?;
    axum::serve(listener, app).await?;

    Ok(())
}
