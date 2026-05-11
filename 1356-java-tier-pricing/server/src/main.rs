use std::sync::Arc;

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post, put},
    Json, Router,
};
use clap::Parser;
use serde::Deserialize;
use tracing_subscriber::EnvFilter;

use tier_pricing_core::*;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "PORT", default_value_t = 3000)]
    port: u16,

    #[arg(long, env = "HOST", default_value = "127.0.0.1")]
    host: String,
}

#[derive(Clone)]
struct AppState {
    storage: InMemoryStorage,
    pricing_service: PricingService,
    order_service: OrderService,
}

#[derive(Deserialize)]
struct CreateProductRequest {
    name: String,
    base_price: f64,
    stock: i64,
}

#[derive(Deserialize)]
struct CreateTierPriceRequest {
    product_id: String,
    min_quantity: i64,
    max_quantity: Option<i64>,
    discount_percent: f64,
    days: i64,
}

#[derive(Deserialize)]
struct CreateMemberRequest {
    name: String,
    member_type: MemberType,
}

#[derive(Deserialize)]
struct CreateMemberDiscountRequest {
    member_type: MemberType,
    product_id: Option<String>,
    discount_percent: f64,
    days: i64,
}

#[derive(Deserialize)]
struct CreateSpecialPriceRequest {
    product_id: String,
    special_price: f64,
    days: i64,
}

#[derive(Deserialize)]
struct CreateFullDiscountRequest {
    threshold: f64,
    discount: f64,
    days: i64,
}

#[derive(Deserialize)]
struct UpdateProductPriceRequest {
    new_price: f64,
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::from_default_env())
        .init();

    let args = Args::parse();

    let storage = InMemoryStorage::new();
    let pricing_service = PricingService::new(storage.clone());
    let order_service = OrderService::new(storage.clone(), pricing_service.clone());

    let state = AppState {
        storage,
        pricing_service,
        order_service,
    };

    let app = Router::new()
        .route("/products", post(create_product))
        .route("/products", get(list_products))
        .route("/products/:id", get(get_product))
        .route("/products/:id/price", put(update_product_price))
        .route("/tier-prices", post(create_tier_price))
        .route("/members", post(create_member))
        .route("/members", get(list_members))
        .route("/member-discounts", post(create_member_discount))
        .route("/special-prices", post(create_special_price))
        .route("/full-discounts", post(create_full_discount))
        .route("/price-query", post(query_price))
        .route("/orders", post(create_order))
        .route("/orders/:id", get(get_order))
        .route("/approvals/:id/approve", post(approve_approval))
        .route("/approvals", get(list_approvals))
        .with_state(Arc::new(state));

    let addr = format!("{}:{}", args.host, args.port);
    tracing::info!("Server listening on {}", addr);

    let addr = std::net::SocketAddr::from((
        args.host.parse::<std::net::IpAddr>().unwrap_or(std::net::IpAddr::V4(std::net::Ipv4Addr::LOCALHOST)),
        args.port,
    ));
    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}

async fn create_product(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateProductRequest>,
) -> impl IntoResponse {
    let product = state.pricing_service.create_product(req.name, req.base_price, req.stock);
    (StatusCode::CREATED, Json(product))
}

async fn list_products(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let products = state.storage.get_all_products();
    Json(products)
}

async fn get_product(
    State(state): State<Arc<AppState>>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match state.storage.get_product(&id) {
        Some(product) => (StatusCode::OK, Json(serde_json::to_value(product).unwrap())),
        None => (StatusCode::NOT_FOUND, Json(serde_json::json!({"error": "Product not found"}))),
    }
}

async fn update_product_price(
    State(state): State<Arc<AppState>>,
    Path(id): Path<String>,
    Json(req): Json<UpdateProductPriceRequest>,
) -> impl IntoResponse {
    match state.pricing_service.update_product_base_price(&id, req.new_price) {
        Ok(Some(approval)) => (
            StatusCode::ACCEPTED,
            Json(serde_json::json!({
                "message": "Price change requires approval",
                "approval_id": approval.id,
                "old_price": approval.old_price,
                "new_price": approval.new_price,
                "change_percent": approval.change_percent,
            })),
        ),
        Ok(None) => (
            StatusCode::OK,
            Json(serde_json::json!({"message": "Price updated successfully"})),
        ),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": e.to_string()})),
        ),
    }
}

async fn create_tier_price(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateTierPriceRequest>,
) -> impl IntoResponse {
    let tier_price = state.pricing_service.create_tier_price(
        req.product_id,
        req.min_quantity,
        req.max_quantity,
        req.discount_percent,
        req.days,
    );
    (StatusCode::CREATED, Json(tier_price))
}

async fn create_member(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateMemberRequest>,
) -> impl IntoResponse {
    let member = state.pricing_service.create_member(req.name, req.member_type);
    (StatusCode::CREATED, Json(member))
}

async fn list_members(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let members = state.storage.get_all_members();
    Json(members)
}

async fn create_member_discount(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateMemberDiscountRequest>,
) -> impl IntoResponse {
    let discount = state.pricing_service.create_member_discount(
        req.member_type,
        req.product_id,
        req.discount_percent,
        req.days,
    );
    (StatusCode::CREATED, Json(discount))
}

async fn create_special_price(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateSpecialPriceRequest>,
) -> impl IntoResponse {
    let special_price = state.pricing_service.create_special_price(
        req.product_id,
        req.special_price,
        req.days,
    );
    (StatusCode::CREATED, Json(special_price))
}

async fn create_full_discount(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateFullDiscountRequest>,
) -> impl IntoResponse {
    let discount = state.pricing_service.create_full_discount(
        req.threshold,
        req.discount,
        req.days,
    );
    (StatusCode::CREATED, Json(discount))
}

async fn query_price(
    State(state): State<Arc<AppState>>,
    Json(req): Json<PriceQueryRequest>,
) -> impl IntoResponse {
    match state.pricing_service.query_price(
        &req.product_id,
        req.quantity,
        req.member_type,
        chrono::Utc::now(),
    ) {
        Ok(result) => (StatusCode::OK, Json(serde_json::to_value(result).unwrap())),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": e.to_string()})),
        ),
    }
}

async fn create_order(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateOrderRequest>,
) -> impl IntoResponse {
    match state.order_service.create_order(req) {
        Ok(order) => (StatusCode::CREATED, Json(serde_json::to_value(order).unwrap())),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": e.to_string()})),
        ),
    }
}

async fn get_order(
    State(state): State<Arc<AppState>>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match state.order_service.get_order(&id) {
        Ok(order) => (StatusCode::OK, Json(serde_json::to_value(order).unwrap())),
        Err(e) => (
            StatusCode::NOT_FOUND,
            Json(serde_json::json!({"error": e.to_string()})),
        ),
    }
}

async fn approve_approval(
    State(state): State<Arc<AppState>>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match state.pricing_service.approve_price_change(&id) {
        Ok(_) => (
            StatusCode::OK,
            Json(serde_json::json!({"message": "Approval approved successfully"})),
        ),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": e.to_string()})),
        ),
    }
}

async fn list_approvals(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let approvals = state.storage.get_all_approvals();
    Json(approvals)
}
