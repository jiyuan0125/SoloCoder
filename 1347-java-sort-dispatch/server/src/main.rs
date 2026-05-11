extern crate core as logistics_core;

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use clap::Parser;
use logistics_core::{
    models::{
        Exception, Order, OrderRequest, Package, PackageStatus, Route, TransferStation, Warehouse,
    },
    LogisticsSystem,
};
use std::net::SocketAddr;
use std::sync::Arc;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "SERVER_PORT", default_value_t = 3000)]
    port: u16,

    #[arg(short, long, env = "SERVER_HOST", default_value = "0.0.0.0")]
    host: String,
}

type AppState = Arc<LogisticsSystem>;

async fn create_order(
    State(state): State<AppState>,
    Json(request): Json<OrderRequest>,
) -> impl IntoResponse {
    match state.create_order(request) {
        Ok(order) => (StatusCode::CREATED, Json(order)).into_response(),
        Err(err) => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": err})),
        )
            .into_response(),
    }
}

async fn get_order(
    State(state): State<AppState>,
    Path(order_id): Path<String>,
) -> impl IntoResponse {
    match state.get_order(&order_id) {
        Some(order) => (StatusCode::OK, Json(order)).into_response(),
        None => (
            StatusCode::NOT_FOUND,
            Json(serde_json::json!({"error": "Order not found"})),
        )
            .into_response(),
    }
}

async fn get_package(
    State(state): State<AppState>,
    Path(package_id): Path<String>,
) -> impl IntoResponse {
    match state.get_package(&package_id) {
        Some(package) => (StatusCode::OK, Json(package)).into_response(),
        None => (
            StatusCode::NOT_FOUND,
            Json(serde_json::json!({"error": "Package not found"})),
        )
            .into_response(),
    }
}

async fn update_package_status(
    State(state): State<AppState>,
    Path(package_id): Path<String>,
    Json(status): Json<PackageStatus>,
) -> impl IntoResponse {
    match state.update_package_status(&package_id, status) {
        Ok(_) => (
            StatusCode::OK,
            Json(serde_json::json!({"message": "Package status updated"})),
        )
            .into_response(),
        Err(err) => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": err})),
        )
            .into_response(),
    }
}

async fn get_exceptions(State(state): State<AppState>) -> impl IntoResponse {
    let exceptions = state.check_exceptions();
    (StatusCode::OK, Json(exceptions))
}

async fn get_warehouses(State(state): State<AppState>) -> impl IntoResponse {
    let warehouses = state.get_all_warehouses();
    (StatusCode::OK, Json(warehouses))
}

async fn get_transfer_stations(State(state): State<AppState>) -> impl IntoResponse {
    let stations = state.get_all_transfer_stations();
    (StatusCode::OK, Json(stations))
}

async fn process_packages(State(state): State<AppState>) -> impl IntoResponse {
    state.process_packages();
    (
        StatusCode::OK,
        Json(serde_json::json!({"message": "Packages processed"})),
    )
}

fn initialize_demo_data(system: &LogisticsSystem) {
    let warehouse1 = Warehouse::new("wh-001", "北京仓库", "北京");
    let warehouse2 = Warehouse::new("wh-002", "上海仓库", "上海");
    let warehouse3 = Warehouse::new("wh-003", "广州仓库", "广州");

    system.add_warehouse(warehouse1);
    system.add_warehouse(warehouse2);
    system.add_warehouse(warehouse3);

    let station1 = TransferStation::new("ts-001", "天津中转站", "天津", 100);
    let station2 = TransferStation::new("ts-002", "南京中转站", "南京", 100);
    let station3 = TransferStation::new("ts-003", "深圳中转站", "深圳", 50);

    system.add_transfer_station(station1);
    system.add_transfer_station(station2);
    system.add_transfer_station(station3);

    system.add_route(Route::new("wh-001", "北京", vec![], 1));
    system.add_route(Route::new("wh-001", "天津", vec!["ts-001"], 1));
    system.add_route(Route::new("wh-002", "上海", vec![], 1));
    system.add_route(Route::new("wh-002", "南京", vec!["ts-002"], 1));
    system.add_route(Route::new("wh-003", "广州", vec![], 1));
    system.add_route(Route::new("wh-003", "深圳", vec!["ts-003"], 1));
    system.add_route(Route::new("wh-001", "上海", vec!["ts-002"], 2));
    system.add_route(Route::new("wh-002", "北京", vec!["ts-001"], 2));

    system.add_inventory("wh-001", "product-001", 100);
    system.add_inventory("wh-002", "product-001", 50);
    system.add_inventory("wh-003", "product-001", 200);
    system.add_inventory("wh-001", "product-002", 30);
    system.add_inventory("wh-002", "product-002", 10);
}

#[tokio::main]
async fn main() {
    let args = Args::parse();

    let system = LogisticsSystem::new();
    initialize_demo_data(&system);

    let app_state = Arc::new(system);

    let app = Router::new()
        .route("/orders", post(create_order))
        .route("/orders/:id", get(get_order))
        .route("/packages/:id", get(get_package))
        .route("/packages/:id/status", post(update_package_status))
        .route("/exceptions", get(get_exceptions))
        .route("/warehouses", get(get_warehouses))
        .route("/transfer-stations", get(get_transfer_stations))
        .route("/process-packages", post(process_packages))
        .with_state(app_state);

    let addr = SocketAddr::new(args.host.parse().unwrap(), args.port);
    println!("Server running on http://{}", addr);

    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}
