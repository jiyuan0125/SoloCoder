use std::net::SocketAddr;
use std::sync::Arc;
use axum::{
    extract::{Json, Path, State},
    http::StatusCode,
    response::{IntoResponse, Response},
    routing::{get, post},
    Router,
};
use clap::Parser;
use serde::Serialize;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};
use uuid::Uuid;
use livestream_core::{
    InMemoryStore, LiveStreamService,
    CreateOrderRequest, CreateProductRequest, CreateUserRequest,
    PayOrderRequest, CancelOrderRequest, RefundOrderRequest,
    UpdateProductStockRequest, SetProductSaleStatusRequest,
    SystemError,
};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short = 'p', long, env = "PORT", default_value_t = 3000)]
    port: u16,
    
    #[arg(short = 'H', long, env = "HOST", default_value = "127.0.0.1")]
    host: String,
}

#[derive(Debug, Clone)]
struct AppState {
    service: Arc<LiveStreamService>,
}

#[derive(Debug, Serialize)]
struct ErrorResponse {
    error: String,
}

struct AppError(SystemError);

impl From<SystemError> for AppError {
    fn from(err: SystemError) -> Self {
        AppError(err)
    }
}

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let status = match &self.0 {
            SystemError::ProductNotFound(_) => StatusCode::NOT_FOUND,
            SystemError::OrderNotFound(_) => StatusCode::NOT_FOUND,
            SystemError::UserNotFound(_) => StatusCode::NOT_FOUND,
            SystemError::InsufficientStock { .. } => StatusCode::BAD_REQUEST,
            SystemError::OrderLockExpired => StatusCode::GONE,
            SystemError::OrderAlreadyPaid => StatusCode::BAD_REQUEST,
            SystemError::OrderAlreadyCancelled => StatusCode::BAD_REQUEST,
            SystemError::OrderNotPaid => StatusCode::BAD_REQUEST,
            SystemError::OrderAlreadyShipped => StatusCode::BAD_REQUEST,
            SystemError::OrderAlreadyCompleted => StatusCode::BAD_REQUEST,
            SystemError::OrderAlreadyRefunded => StatusCode::BAD_REQUEST,
            SystemError::InvalidStatusTransition { .. } => StatusCode::BAD_REQUEST,
            SystemError::ProductNotOnSale => StatusCode::BAD_REQUEST,
            SystemError::LiveStreamNotActive => StatusCode::BAD_REQUEST,
            SystemError::PurchaseLimitExceeded { .. } => StatusCode::BAD_REQUEST,
            SystemError::Internal(_) => StatusCode::INTERNAL_SERVER_ERROR,
        };
        
        (status, Json(ErrorResponse { error: self.0.to_string() })).into_response()
    }
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "server=debug,tower_http=debug,axum=debug".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    let args = Args::parse();

    let store = InMemoryStore::new();
    let service = Arc::new(LiveStreamService::new(store));
    let state = AppState { service };

    let app = Router::new()
        .route("/health", get(health_check))
        .route("/live/status", get(get_live_status))
        .route("/live/start", post(start_live_stream))
        .route("/live/end", post(end_live_stream))
        .route("/users", post(create_user))
        .route("/users", get(list_users))
        .route("/users/:id", get(get_user))
        .route("/products", post(create_product))
        .route("/products", get(list_products))
        .route("/products/:id", get(get_product))
        .route("/products/stock", post(update_product_stock))
        .route("/products/sale", post(set_product_sale_status))
        .route("/orders", post(create_order))
        .route("/orders", get(list_orders))
        .route("/orders/:id", get(get_order))
        .route("/orders/pay", post(pay_order))
        .route("/orders/cancel", post(cancel_order))
        .route("/orders/:id/ship", post(ship_order))
        .route("/orders/:id/complete", post(complete_order))
        .route("/orders/refund", post(refund_order))
        .with_state(state);

    let addr: SocketAddr = format!("{}:{}", args.host, args.port).parse()?;
    tracing::info!("Server listening on {}", addr);
    
    let listener = tokio::net::TcpListener::bind(addr).await?;
    axum::serve(listener, app).await?;

    Ok(())
}

async fn health_check() -> StatusCode {
    StatusCode::OK
}

async fn get_live_status(
    State(state): State<AppState>,
) -> Result<impl IntoResponse, AppError> {
    let status = state.service.get_live_status()?;
    Ok(Json(status))
}

async fn start_live_stream(
    State(state): State<AppState>,
) -> Result<impl IntoResponse, AppError> {
    state.service.start_live_stream()?;
    Ok(StatusCode::OK)
}

async fn end_live_stream(
    State(state): State<AppState>,
) -> Result<impl IntoResponse, AppError> {
    state.service.end_live_stream()?;
    Ok(StatusCode::OK)
}

async fn create_user(
    State(state): State<AppState>,
    Json(request): Json<CreateUserRequest>,
) -> Result<impl IntoResponse, AppError> {
    let user = state.service.create_user(request)?;
    Ok((StatusCode::CREATED, Json(user)))
}

async fn list_users(
    State(state): State<AppState>,
) -> Result<impl IntoResponse, AppError> {
    let users = state.service.list_users()?;
    Ok(Json(users))
}

async fn get_user(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<impl IntoResponse, AppError> {
    let user = state.service.get_user(id)?;
    Ok(Json(user))
}

async fn create_product(
    State(state): State<AppState>,
    Json(request): Json<CreateProductRequest>,
) -> Result<impl IntoResponse, AppError> {
    let product = state.service.create_product(request)?;
    Ok((StatusCode::CREATED, Json(product)))
}

async fn list_products(
    State(state): State<AppState>,
) -> Result<impl IntoResponse, AppError> {
    let products = state.service.list_products()?;
    Ok(Json(products))
}

async fn get_product(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<impl IntoResponse, AppError> {
    let product = state.service.get_product(id)?;
    Ok(Json(product))
}

async fn update_product_stock(
    State(state): State<AppState>,
    Json(request): Json<UpdateProductStockRequest>,
) -> Result<impl IntoResponse, AppError> {
    let product = state.service.update_product_stock(request)?;
    Ok(Json(product))
}

async fn set_product_sale_status(
    State(state): State<AppState>,
    Json(request): Json<SetProductSaleStatusRequest>,
) -> Result<impl IntoResponse, AppError> {
    let product = state.service.set_product_sale_status(request)?;
    Ok(Json(product))
}

async fn create_order(
    State(state): State<AppState>,
    Json(request): Json<CreateOrderRequest>,
) -> Result<impl IntoResponse, AppError> {
    let order = state.service.create_order(request)?;
    Ok((StatusCode::CREATED, Json(order)))
}

async fn list_orders(
    State(state): State<AppState>,
) -> Result<impl IntoResponse, AppError> {
    let orders = state.service.list_orders()?;
    Ok(Json(orders))
}

async fn get_order(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<impl IntoResponse, AppError> {
    let order = state.service.get_order(id)?;
    Ok(Json(order))
}

async fn pay_order(
    State(state): State<AppState>,
    Json(request): Json<PayOrderRequest>,
) -> Result<impl IntoResponse, AppError> {
    let order = state.service.pay_order(request)?;
    Ok(Json(order))
}

async fn cancel_order(
    State(state): State<AppState>,
    Json(request): Json<CancelOrderRequest>,
) -> Result<impl IntoResponse, AppError> {
    let order = state.service.cancel_order(request)?;
    Ok(Json(order))
}

async fn ship_order(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<impl IntoResponse, AppError> {
    let order = state.service.ship_order(id)?;
    Ok(Json(order))
}

async fn complete_order(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<impl IntoResponse, AppError> {
    let order = state.service.complete_order(id)?;
    Ok(Json(order))
}

async fn refund_order(
    State(state): State<AppState>,
    Json(request): Json<RefundOrderRequest>,
) -> Result<impl IntoResponse, AppError> {
    let order = state.service.refund_order(request)?;
    Ok(Json(order))
}
