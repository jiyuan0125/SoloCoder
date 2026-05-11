use std::net::SocketAddr;
use std::sync::Arc;

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::{IntoResponse, Response},
    routing::{get, post},
    Json, Router,
};
use clap::Parser;
use serde::Serialize;
use tower_http::cors::{Any, CorsLayer};
use uuid::Uuid;

use qa_core::{
    CompleteInspectionRequest, CreateOrderRequest, InspectItemRequest, QaService,
    ReinspectionRequest, UpdateDefectRequest,
};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "PORT", default_value_t = 3000)]
    port: u16,

    #[arg(short, long, env = "HOST", default_value = "127.0.0.1")]
    host: String,
}

#[derive(Clone)]
struct AppState {
    service: QaService,
}

#[derive(Serialize)]
struct ApiError {
    error: String,
}

async fn create_order(
    State(state): State<Arc<AppState>>,
    Json(request): Json<CreateOrderRequest>,
) -> Response {
    match state.service.create_order(request) {
        Ok(order) => (StatusCode::CREATED, Json(order)).into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiError { error: e.to_string() }),
        )
            .into_response(),
    }
}

async fn get_order(
    State(state): State<Arc<AppState>>,
    Path(order_id): Path<Uuid>,
) -> Response {
    match state.service.get_order(order_id) {
        Some(order) => (StatusCode::OK, Json(order)).into_response(),
        None => (
            StatusCode::NOT_FOUND,
            Json(ApiError {
                error: format!("Order {} not found", order_id),
            }),
        )
            .into_response(),
    }
}

async fn get_all_orders(State(state): State<Arc<AppState>>) -> Response {
    let orders = state.service.get_all_orders();
    (StatusCode::OK, Json(orders)).into_response()
}

async fn start_inspection(
    State(state): State<Arc<AppState>>,
    Path(order_id): Path<Uuid>,
) -> Response {
    match state.service.start_inspection(order_id) {
        Ok(order) => (StatusCode::OK, Json(order)).into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiError { error: e.to_string() }),
        )
            .into_response(),
    }
}

async fn inspect_item(
    State(state): State<Arc<AppState>>,
    Json(request): Json<InspectItemRequest>,
) -> Response {
    match state.service.inspect_item(request) {
        Ok(defect) => (StatusCode::OK, Json(defect)).into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiError { error: e.to_string() }),
        )
            .into_response(),
    }
}

async fn get_defects_for_order(
    State(state): State<Arc<AppState>>,
    Path(order_id): Path<Uuid>,
) -> Response {
    let defects = state.service.get_defects_for_order(order_id);
    (StatusCode::OK, Json(defects)).into_response()
}

async fn update_defect(
    State(state): State<Arc<AppState>>,
    Json(request): Json<UpdateDefectRequest>,
) -> Response {
    match state.service.update_defect_description(request) {
        Ok(defect) => (StatusCode::OK, Json(defect)).into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiError { error: e.to_string() }),
        )
            .into_response(),
    }
}

async fn complete_inspection(
    State(state): State<Arc<AppState>>,
    Json(request): Json<CompleteInspectionRequest>,
) -> Response {
    match state.service.complete_inspection(request.order_id) {
        Ok(result) => (StatusCode::OK, Json(result)).into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiError { error: e.to_string() }),
        )
            .into_response(),
    }
}

async fn request_reinspection(
    State(state): State<Arc<AppState>>,
    Json(request): Json<ReinspectionRequest>,
) -> Response {
    match state.service.request_reinspection(request) {
        Ok(order) => (StatusCode::CREATED, Json(order)).into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiError { error: e.to_string() }),
        )
            .into_response(),
    }
}

#[tokio::main]
async fn main() {
    let args = Args::parse();

    let state = Arc::new(AppState {
        service: QaService::new(),
    });

    let cors = CorsLayer::new()
        .allow_methods(Any)
        .allow_origin(Any)
        .allow_headers(Any);

    let app = Router::new()
        .route("/api/orders", post(create_order).get(get_all_orders))
        .route("/api/orders/:id", get(get_order))
        .route("/api/orders/:id/start", post(start_inspection))
        .route("/api/inspect", post(inspect_item))
        .route("/api/orders/:id/defects", get(get_defects_for_order))
        .route("/api/defects/update", post(update_defect))
        .route("/api/complete", post(complete_inspection))
        .route("/api/reinspect", post(request_reinspection))
        .with_state(state)
        .layer(cors);

    let addr = format!("{}:{}", args.host, args.port);
    let socket_addr: SocketAddr = addr.parse().expect("Invalid address");

    println!("QA Inspection Server starting on {}", addr);
    println!("Available endpoints:");
    println!("  POST /api/orders - Create new inspection order");
    println!("  GET  /api/orders - List all orders");
    println!("  GET  /api/orders/:id - Get order by ID");
    println!("  POST /api/orders/:id/start - Start inspection");
    println!("  POST /api/inspect - Record item inspection");
    println!("  GET  /api/orders/:id/defects - Get defects for order");
    println!("  POST /api/defects/update - Update defect description");
    println!("  POST /api/complete - Complete inspection");
    println!("  POST /api/reinspect - Request reinspection");

    axum::Server::bind(&socket_addr)
        .serve(app.into_make_service())
        .await
        .expect("Server failed to start");
}
