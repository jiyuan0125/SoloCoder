use axum::{
    extract::State,
    http::StatusCode,
    response::{IntoResponse, Json},
    routing::get,
    Router,
};
use purchase_return_core::{
    CreateExchangeRequest, CreateProductRequest, CreatePurchaseRequest,
    CreateReturnRequest, CreateSupplierRequest, InMemoryStore, PurchaseReturnService,
};
use serde_json::json;
use std::net::SocketAddr;
use std::sync::Arc;
use tower_http::cors::CorsLayer;

type AppState = Arc<PurchaseReturnService>;

pub async fn run_server(addr: SocketAddr) {
    let store = InMemoryStore::new();
    let service = PurchaseReturnService::new(store);
    let state: AppState = Arc::new(service);

    let app = Router::new()
        .route("/products", get(list_products).post(create_product))
        .route("/suppliers", get(list_suppliers).post(create_supplier))
        .route("/batches", get(list_batches).post(create_purchase))
        .route("/returns", get(list_returns).post(create_return))
        .route("/exchanges", get(list_exchanges).post(create_exchange))
        .layer(CorsLayer::permissive())
        .with_state(state);

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    println!("Server running on {}", listener.local_addr().unwrap());
    axum::serve(listener, app).await.unwrap();
}

async fn list_products(State(state): State<AppState>) -> impl IntoResponse {
    let products = state.list_products();
    Json(products).into_response()
}

async fn create_product(
    State(state): State<AppState>,
    Json(req): Json<CreateProductRequest>,
) -> impl IntoResponse {
    match state.create_product(req.name) {
        Ok(product) => Json(product).into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(json!({ "error": e.to_string() })),
        )
            .into_response(),
    }
}

async fn list_suppliers(State(state): State<AppState>) -> impl IntoResponse {
    let suppliers = state.list_suppliers();
    Json(suppliers).into_response()
}

async fn create_supplier(
    State(state): State<AppState>,
    Json(req): Json<CreateSupplierRequest>,
) -> impl IntoResponse {
    match state.create_supplier(req.name) {
        Ok(supplier) => Json(supplier).into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(json!({ "error": e.to_string() })),
        )
            .into_response(),
    }
}

async fn list_batches(State(state): State<AppState>) -> impl IntoResponse {
    let batches = state.list_batches();
    Json(batches).into_response()
}

async fn create_purchase(
    State(state): State<AppState>,
    Json(req): Json<CreatePurchaseRequest>,
) -> impl IntoResponse {
    match state.create_purchase(
        req.product_id,
        req.supplier_id,
        req.batch_number,
        req.price,
        req.quantity,
    ) {
        Ok(batch) => Json(batch).into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(json!({ "error": e.to_string() })),
        )
            .into_response(),
    }
}

async fn list_returns(State(state): State<AppState>) -> impl IntoResponse {
    let returns = state.list_returns();
    Json(returns).into_response()
}

async fn create_return(
    State(state): State<AppState>,
    Json(req): Json<CreateReturnRequest>,
) -> impl IntoResponse {
    match state.create_return(req.batch_id, req.quantity) {
        Ok(return_record) => Json(return_record).into_response(),
        Err(e) => {
            let status = match e {
                purchase_return_core::PurchaseReturnError::InsufficientQuantity { .. } => {
                    StatusCode::BAD_REQUEST
                }
                _ => StatusCode::BAD_REQUEST,
            };
            (status, Json(json!({ "error": e.to_string() }))).into_response()
        }
    }
}

async fn list_exchanges(State(state): State<AppState>) -> impl IntoResponse {
    let exchanges = state.list_exchanges();
    Json(exchanges).into_response()
}

async fn create_exchange(
    State(state): State<AppState>,
    Json(req): Json<CreateExchangeRequest>,
) -> impl IntoResponse {
    match state.create_exchange(
        req.old_product_id,
        req.old_quantity,
        req.new_product_id,
        req.new_quantity,
    ) {
        Ok(exchange) => Json(exchange).into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(json!({ "error": e.to_string() })),
        )
            .into_response(),
    }
}
