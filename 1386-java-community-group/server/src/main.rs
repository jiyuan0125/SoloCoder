use std::sync::Arc;
use std::net::SocketAddr;
use std::time::Duration;
use axum::{
    extract::{Path, State},
    http::{StatusCode, Response},
    response::{IntoResponse, Json},
    routing::{get, post},
    Router,
    body::Body,
};
use clap::Parser;
use serde::{Deserialize, Serialize};
use tower_http::cors::{Any, CorsLayer};
use chrono::NaiveTime;
use uuid::Uuid;
use community_core::{
    InMemoryStore,
    CommunityService,
    BusinessError,
    CreateOrderRequest,
    CreateProductRequest,
};

#[derive(Parser, Debug, Clone)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "PORT", default_value_t = 3000)]
    port: u16,
    
    #[arg(short = 'H', long, env = "HOST", default_value = "127.0.0.1")]
    host: String,
}

#[derive(Clone)]
struct AppState {
    service: CommunityService,
}

#[derive(Serialize, Deserialize)]
struct CreatePickupPointRequest {
    name: String,
    #[serde(with = "time_serde")]
    cut_off_time: NaiveTime,
}

#[derive(Serialize, Deserialize)]
struct ErrorResponse {
    error: String,
}

struct AppError(BusinessError);

impl IntoResponse for AppError {
    fn into_response(self) -> Response<Body> {
        let status = match &self.0 {
            BusinessError::PickupPointNotFound => StatusCode::NOT_FOUND,
            BusinessError::ProductNotFound => StatusCode::NOT_FOUND,
            BusinessError::OrderNotFound => StatusCode::NOT_FOUND,
            BusinessError::SortingListNotFound => StatusCode::NOT_FOUND,
            BusinessError::InsufficientStock { .. } => StatusCode::BAD_REQUEST,
            BusinessError::OrderAlreadyPaid => StatusCode::BAD_REQUEST,
            BusinessError::OrderNotPaid => StatusCode::BAD_REQUEST,
            BusinessError::PaymentDeadlinePassed => StatusCode::BAD_REQUEST,
            BusinessError::CutOffTimePassed => StatusCode::BAD_REQUEST,
            BusinessError::RefundAfterCutOffTime => StatusCode::BAD_REQUEST,
            BusinessError::InvalidOperation(_) => StatusCode::BAD_REQUEST,
            BusinessError::ProductNotInPickupPoint => StatusCode::BAD_REQUEST,
            BusinessError::EmptyOrder => StatusCode::BAD_REQUEST,
            BusinessError::PickupPointAlreadyExists => StatusCode::BAD_REQUEST,
        };
        (status, Json(ErrorResponse { error: self.0.to_string() })).into_response()
    }
}

impl From<BusinessError> for AppError {
    fn from(err: BusinessError) -> Self {
        AppError(err)
    }
}

async fn create_pickup_point(
    State(state): State<AppState>,
    Json(req): Json<CreatePickupPointRequest>,
) -> Result<impl IntoResponse, AppError> {
    let point = state.service.create_pickup_point(req.name, req.cut_off_time);
    Ok((StatusCode::CREATED, Json(point)))
}

async fn list_pickup_points(
    State(state): State<AppState>,
) -> impl IntoResponse {
    let points = state.service.list_pickup_points();
    Json(points)
}

async fn get_pickup_point(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<impl IntoResponse, AppError> {
    let point = state.service.get_pickup_point(&id)?;
    Ok(Json(point))
}

async fn create_product(
    State(state): State<AppState>,
    Path(pickup_point_id): Path<Uuid>,
    Json(req): Json<CreateProductRequest>,
) -> Result<impl IntoResponse, AppError> {
    let product = state.service.create_product(
        pickup_point_id,
        req.name,
        req.unit_price,
        req.stock,
    )?;
    Ok((StatusCode::CREATED, Json(product)))
}

async fn list_products(
    State(state): State<AppState>,
    Path(pickup_point_id): Path<Uuid>,
) -> Result<impl IntoResponse, AppError> {
    let products = state.service.list_products_by_pickup_point(&pickup_point_id)?;
    Ok(Json(products))
}

async fn create_order(
    State(state): State<AppState>,
    Path(pickup_point_id): Path<Uuid>,
    Json(req): Json<CreateOrderRequest>,
) -> Result<impl IntoResponse, AppError> {
    let order = state.service.create_order(pickup_point_id, req.user_id, req.items)?;
    Ok((StatusCode::CREATED, Json(order)))
}

async fn get_order(
    State(state): State<AppState>,
    Path(order_id): Path<Uuid>,
) -> Result<impl IntoResponse, AppError> {
    let order = state.service.get_order(&order_id)?;
    Ok(Json(order))
}

async fn list_user_orders(
    State(state): State<AppState>,
    Path(user_id): Path<Uuid>,
) -> impl IntoResponse {
    let orders = state.service.list_orders_by_user(&user_id);
    Json(orders)
}

async fn pay_order(
    State(state): State<AppState>,
    Path(order_id): Path<Uuid>,
) -> Result<impl IntoResponse, AppError> {
    let order = state.service.pay_order(order_id)?;
    Ok(Json(order))
}

async fn cancel_order(
    State(state): State<AppState>,
    Path(order_id): Path<Uuid>,
) -> Result<impl IntoResponse, AppError> {
    let order = state.service.cancel_order(order_id)?;
    Ok(Json(order))
}

async fn refund_order(
    State(state): State<AppState>,
    Path(order_id): Path<Uuid>,
) -> Result<impl IntoResponse, AppError> {
    let order = state.service.refund_order(order_id)?;
    Ok(Json(order))
}

async fn process_cut_off(
    State(state): State<AppState>,
    Path(pickup_point_id): Path<Uuid>,
) -> Result<impl IntoResponse, AppError> {
    let list = state.service.process_cut_off(pickup_point_id)?;
    Ok((StatusCode::CREATED, Json(list)))
}

async fn list_sorting_lists(
    State(state): State<AppState>,
    Path(pickup_point_id): Path<Uuid>,
) -> Result<impl IntoResponse, AppError> {
    let lists = state.service.list_sorting_lists_by_pickup_point(&pickup_point_id)?;
    Ok(Json(lists))
}

async fn confirm_sorting_complete(
    State(state): State<AppState>,
    Path(sorting_list_id): Path<Uuid>,
) -> Result<impl IntoResponse, AppError> {
    let list = state.service.confirm_sorting_complete(sorting_list_id)?;
    Ok(Json(list))
}

async fn pickup_order(
    State(state): State<AppState>,
    Path(order_id): Path<Uuid>,
) -> Result<impl IntoResponse, AppError> {
    let order = state.service.pickup_order(order_id)?;
    Ok(Json(order))
}

#[tokio::main]
async fn main() {
    let args = Args::parse();
    
    let store = Arc::new(InMemoryStore::new());
    let service = CommunityService::new(store.clone());
    
    let service_clone = service.clone();
    tokio::spawn(async move {
        let mut interval = tokio::time::interval(Duration::from_secs(10));
        loop {
            interval.tick().await;
            let cancelled = service_clone.clean_expired_orders();
            if !cancelled.is_empty() {
                println!("Cleaned {} expired orders", cancelled.len());
            }
        }
    });
    
    let app_state = AppState { service };
    
    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);
    
    let app = Router::new()
        .route("/api/pickup-points", post(create_pickup_point).get(list_pickup_points))
        .route("/api/pickup-points/:id", get(get_pickup_point))
        .route("/api/pickup-points/:id/products", post(create_product).get(list_products))
        .route("/api/pickup-points/:id/orders", post(create_order))
        .route("/api/orders/:id", get(get_order))
        .route("/api/orders/:id/pay", post(pay_order))
        .route("/api/orders/:id/cancel", post(cancel_order))
        .route("/api/orders/:id/refund", post(refund_order))
        .route("/api/orders/:id/pickup", post(pickup_order))
        .route("/api/users/:id/orders", get(list_user_orders))
        .route("/api/pickup-points/:id/cut-off", post(process_cut_off))
        .route("/api/pickup-points/:id/sorting-lists", get(list_sorting_lists))
        .route("/api/sorting-lists/:id/confirm", post(confirm_sorting_complete))
        .with_state(app_state)
        .layer(cors);
    
    let addr: SocketAddr = format!("{}:{}", args.host, args.port)
        .parse()
        .expect("Invalid host:port");
    
    println!("Server listening on {}", addr);
    
    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}

mod time_serde {
    use serde::{self, Deserialize, Deserializer, Serializer};
    use chrono::NaiveTime;

    pub fn serialize<S>(time: &NaiveTime, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: Serializer,
    {
        serializer.serialize_str(&time.format("%H:%M").to_string())
    }

    pub fn deserialize<'de, D>(deserializer: D) -> Result<NaiveTime, D::Error>
    where
        D: Deserializer<'de>,
    {
        let s = String::deserialize(deserializer)?;
        NaiveTime::parse_from_str(&s, "%H:%M")
            .or_else(|_| NaiveTime::parse_from_str(&s, "%H:%M:%S"))
            .map_err(serde::de::Error::custom)
    }
}
