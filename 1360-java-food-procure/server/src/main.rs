use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use clap::Parser;
use food_procure_core::errors::SystemError;
use food_procure_core::models::*;
use food_procure_core::services::*;
use food_procure_core::store::SharedStore;
use serde::Deserialize;
use std::net::SocketAddr;
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "SERVER_PORT", default_value_t = 3000)]
    port: u16,

    #[arg(long, env = "SERVER_HOST", default_value = "127.0.0.1")]
    host: String,
}

#[derive(Clone)]
struct AppState {
    store: SharedStore,
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();

    let args = Args::parse();
    let store = food_procure_core::store::create_shared_store();

    let app_state = AppState { store: store.clone() };

    let app = Router::new()
        .route("/health", get(health_check))
        .route("/suppliers", get(list_suppliers_handler).post(create_supplier_handler))
        .route("/suppliers/:id", get(get_supplier_handler))
        .route("/suppliers/:id/score", post(update_supplier_score_handler))
        .route("/ingredients", get(list_ingredients_handler).post(create_ingredient_handler))
        .route("/stores", get(list_stores_handler).post(create_store_handler))
        .route("/purchase-requests", get(list_purchase_requests_handler).post(create_purchase_request_handler))
        .route("/purchase-requests/:id", get(get_purchase_request_handler))
        .route("/purchase-requests/:id/quotations", get(list_quotations_handler).post(request_quotations_handler))
        .route("/orders", post(place_order_handler))
        .route("/orders/:id", get(get_order_handler).post(confirm_order_handler))
        .route("/emergency-orders", post(place_emergency_order_handler))
        .route("/stores/:id/emergency-stats", get(get_emergency_stats_handler))
        .with_state(app_state);

    let addr: SocketAddr = format!("{}:{}", args.host, args.port).parse().unwrap();
    tracing::info!("Server listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}

async fn health_check() -> &'static str {
    "OK"
}

fn error_to_response(err: SystemError) -> impl IntoResponse {
    let status = match &err {
        SystemError::SupplierNotFound(_) => StatusCode::NOT_FOUND,
        SystemError::StoreNotFound(_) => StatusCode::NOT_FOUND,
        SystemError::IngredientNotFound(_) => StatusCode::NOT_FOUND,
        SystemError::PurchaseRequestNotFound(_) => StatusCode::NOT_FOUND,
        SystemError::QuotationNotFound(_) => StatusCode::NOT_FOUND,
        SystemError::OrderNotFound(_) => StatusCode::NOT_FOUND,
        SystemError::SupplierNotActive(_) => StatusCode::BAD_REQUEST,
        SystemError::SupplierDoesNotOfferIngredient { .. } => StatusCode::BAD_REQUEST,
        SystemError::InsufficientStock { .. } => StatusCode::BAD_REQUEST,
        SystemError::MinOrderQuantityNotMet { .. } => StatusCode::BAD_REQUEST,
        SystemError::EmergencyPurchaseSingleLimitExceeded(_) => StatusCode::BAD_REQUEST,
        SystemError::EmergencyPurchaseMonthlyLimitExceeded(_) => StatusCode::BAD_REQUEST,
        SystemError::NonRecommendedSupplierWithoutReason => StatusCode::BAD_REQUEST,
        SystemError::QuotationNotAvailable => StatusCode::NOT_FOUND,
        SystemError::InvalidPurchaseStatus => StatusCode::BAD_REQUEST,
        SystemError::ConcurrentConflict => StatusCode::CONFLICT,
        SystemError::InvalidInput(_) => StatusCode::BAD_REQUEST,
        SystemError::Internal(_) => StatusCode::INTERNAL_SERVER_ERROR,
    };

    (status, Json(serde_json::json!({ "error": err.to_string() })))
}

#[derive(Deserialize)]
struct CreateSupplierRequest {
    name: String,
    contact: String,
    phone: String,
    ingredients: Vec<SupplierIngredient>,
}

async fn create_supplier_handler(
    State(state): State<AppState>,
    Json(req): Json<CreateSupplierRequest>,
) -> impl IntoResponse {
    match supplier_management::create_supplier(
        &state.store,
        req.name,
        req.contact,
        req.phone,
        req.ingredients,
    ) {
        Ok(supplier) => (StatusCode::CREATED, Json(supplier)).into_response(),
        Err(e) => error_to_response(e).into_response(),
    }
}

async fn list_suppliers_handler(State(state): State<AppState>) -> impl IntoResponse {
    match supplier_management::list_suppliers(&state.store) {
        Ok(suppliers) => (StatusCode::OK, Json(suppliers)).into_response(),
        Err(e) => error_to_response(e).into_response(),
    }
}

async fn get_supplier_handler(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match Uuid::parse_str(&id) {
        Ok(uuid) => match supplier_management::get_supplier(&state.store, uuid) {
            Ok(supplier) => (StatusCode::OK, Json(supplier)).into_response(),
            Err(e) => error_to_response(e).into_response(),
        },
        Err(_) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({ "error": "Invalid UUID" }))).into_response(),
    }
}

#[derive(Deserialize)]
struct UpdateScoreRequest {
    month: String,
    score: f64,
}

async fn update_supplier_score_handler(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(req): Json<UpdateScoreRequest>,
) -> impl IntoResponse {
    match Uuid::parse_str(&id) {
        Ok(uuid) => match supplier_management::update_supplier_score(&state.store, uuid, req.month, req.score) {
            Ok(supplier) => (StatusCode::OK, Json(supplier)).into_response(),
            Err(e) => error_to_response(e).into_response(),
        },
        Err(_) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({ "error": "Invalid UUID" }))).into_response(),
    }
}

#[derive(Deserialize)]
struct CreateIngredientRequest {
    name: String,
    unit: String,
    category: String,
    weight_per_unit: f64,
}

async fn create_ingredient_handler(
    State(state): State<AppState>,
    Json(req): Json<CreateIngredientRequest>,
) -> impl IntoResponse {
    match supplier_management::create_ingredient(
        &state.store,
        req.name,
        req.unit,
        req.category,
        req.weight_per_unit,
    ) {
        Ok(ingredient) => (StatusCode::CREATED, Json(ingredient)).into_response(),
        Err(e) => error_to_response(e).into_response(),
    }
}

async fn list_ingredients_handler(State(state): State<AppState>) -> impl IntoResponse {
    match supplier_management::list_ingredients(&state.store) {
        Ok(ingredients) => (StatusCode::OK, Json(ingredients)).into_response(),
        Err(e) => error_to_response(e).into_response(),
    }
}

#[derive(Deserialize)]
struct CreateStoreRequest {
    name: String,
    address: String,
    contact: String,
    phone: String,
}

async fn create_store_handler(
    State(state): State<AppState>,
    Json(req): Json<CreateStoreRequest>,
) -> impl IntoResponse {
    match supplier_management::create_store(&state.store, req.name, req.address, req.contact, req.phone) {
        Ok(store) => (StatusCode::CREATED, Json(store)).into_response(),
        Err(e) => error_to_response(e).into_response(),
    }
}

async fn list_stores_handler(State(state): State<AppState>) -> impl IntoResponse {
    match supplier_management::list_stores(&state.store) {
        Ok(stores) => (StatusCode::OK, Json(stores)).into_response(),
        Err(e) => error_to_response(e).into_response(),
    }
}

#[derive(Deserialize)]
struct CreatePurchaseRequestRequest {
    store_id: Uuid,
    purchase_type: PurchaseType,
    items: Vec<PurchaseItem>,
}

async fn create_purchase_request_handler(
    State(state): State<AppState>,
    Json(req): Json<CreatePurchaseRequestRequest>,
) -> impl IntoResponse {
    match supplier_management::create_purchase_request(
        &state.store,
        req.store_id,
        req.purchase_type,
        req.items,
    ) {
        Ok(pr) => (StatusCode::CREATED, Json(pr)).into_response(),
        Err(e) => error_to_response(e).into_response(),
    }
}

async fn list_purchase_requests_handler(State(state): State<AppState>) -> impl IntoResponse {
    match supplier_management::list_purchase_requests(&state.store) {
        Ok(prs) => (StatusCode::OK, Json(prs)).into_response(),
        Err(e) => error_to_response(e).into_response(),
    }
}

async fn get_purchase_request_handler(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match Uuid::parse_str(&id) {
        Ok(uuid) => match supplier_management::get_purchase_request(&state.store, uuid) {
            Ok(pr) => (StatusCode::OK, Json(pr)).into_response(),
            Err(e) => error_to_response(e).into_response(),
        },
        Err(_) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({ "error": "Invalid UUID" }))).into_response(),
    }
}

async fn request_quotations_handler(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match Uuid::parse_str(&id) {
        Ok(uuid) => match quotation::request_quotations(&state.store, uuid) {
            Ok(quotations) => (StatusCode::CREATED, Json(quotations)).into_response(),
            Err(e) => error_to_response(e).into_response(),
        },
        Err(_) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({ "error": "Invalid UUID" }))).into_response(),
    }
}

async fn list_quotations_handler(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match Uuid::parse_str(&id) {
        Ok(uuid) => match quotation::list_quotations(&state.store, uuid) {
            Ok(quotations) => (StatusCode::OK, Json(quotations)).into_response(),
            Err(e) => error_to_response(e).into_response(),
        },
        Err(_) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({ "error": "Invalid UUID" }))).into_response(),
    }
}

#[derive(Deserialize)]
struct PlaceOrderRequest {
    quotation_id: Uuid,
    non_recommended_reason: Option<String>,
}

async fn place_order_handler(
    State(state): State<AppState>,
    Json(req): Json<PlaceOrderRequest>,
) -> impl IntoResponse {
    match ordering::place_order(&state.store, req.quotation_id, req.non_recommended_reason) {
        Ok(order) => (StatusCode::CREATED, Json(order)).into_response(),
        Err(e) => error_to_response(e).into_response(),
    }
}

async fn get_order_handler(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match Uuid::parse_str(&id) {
        Ok(uuid) => match ordering::get_order(&state.store, uuid) {
            Ok(order) => (StatusCode::OK, Json(order)).into_response(),
            Err(e) => error_to_response(e).into_response(),
        },
        Err(_) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({ "error": "Invalid UUID" }))).into_response(),
    }
}

async fn confirm_order_handler(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match Uuid::parse_str(&id) {
        Ok(uuid) => match ordering::confirm_order(&state.store, uuid) {
            Ok(order) => (StatusCode::OK, Json(order)).into_response(),
            Err(e) => error_to_response(e).into_response(),
        },
        Err(_) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({ "error": "Invalid UUID" }))).into_response(),
    }
}

#[derive(Deserialize)]
struct PlaceEmergencyOrderRequest {
    store_id: Uuid,
    supplier_id: Uuid,
    items: Vec<PurchaseItem>,
}

async fn place_emergency_order_handler(
    State(state): State<AppState>,
    Json(req): Json<PlaceEmergencyOrderRequest>,
) -> impl IntoResponse {
    match emergency_purchase::place_emergency_order(
        &state.store,
        req.store_id,
        req.supplier_id,
        req.items,
    ) {
        Ok(order) => (StatusCode::CREATED, Json(order)).into_response(),
        Err(e) => error_to_response(e).into_response(),
    }
}

#[derive(Deserialize)]
struct EmergencyStatsQuery {
    month: Option<String>,
}

async fn get_emergency_stats_handler(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match Uuid::parse_str(&id) {
        Ok(uuid) => match emergency_purchase::get_store_emergency_stats(&state.store, uuid, None) {
            Ok(stats) => (StatusCode::OK, Json(stats)).into_response(),
            Err(e) => error_to_response(e).into_response(),
        },
        Err(_) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({ "error": "Invalid UUID" }))).into_response(),
    }
}
