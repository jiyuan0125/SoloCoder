use std::net::SocketAddr;
use std::sync::Arc;
use std::time::Duration;

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post, put},
    Json, Router,
};
use clap::Parser;
use serde::{Deserialize, Serialize};
use tokio::time::interval;
use uuid::Uuid;

use maint_core::{
    Device, DeviceStatus, InMemoryStorage, MaintenanceCycleType, MaintenanceOrder,
    MaintenancePlan, MaintenanceService, Notification, RepairOrder, SparePartUsage,
};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "SERVER_PORT", default_value_t = 3000)]
    port: u16,

    #[arg(long, env = "SERVER_HOST", default_value = "0.0.0.0")]
    host: String,
}

struct AppState {
    service: MaintenanceService,
}

#[derive(Debug, Deserialize)]
struct CreateDeviceRequest {
    name: String,
    code: String,
    description: String,
}

#[derive(Debug, Deserialize)]
struct UpdateDeviceStatusRequest {
    status: DeviceStatus,
}

#[derive(Debug, Deserialize)]
struct UpdateDeviceHoursRequest {
    running_hours: u64,
}

#[derive(Debug, Deserialize)]
struct CreatePlanRequest {
    device_id: Uuid,
    name: String,
    cycle_type: MaintenanceCycleType,
    cycle_value: u64,
}

#[derive(Debug, Deserialize)]
struct CreateRepairOrderRequest {
    device_id: Uuid,
    title: String,
    description: String,
}

#[derive(Debug, Deserialize)]
struct CompleteRepairOrderRequest {
    fault_category: String,
    spare_parts: Vec<SparePartUsage>,
    repair_notes: String,
}

#[derive(Debug, Deserialize)]
struct CreateSparePartRequest {
    name: String,
    code: String,
    quantity: u32,
    safety_stock: u32,
    unit: String,
}

#[derive(Debug, Deserialize)]
struct CompleteMaintenanceOrderRequest {
    notes: String,
}

#[derive(Debug, Serialize)]
struct ApiResponse<T> {
    success: bool,
    data: Option<T>,
    message: Option<String>,
}

impl<T> ApiResponse<T> {
    fn success(data: T) -> Self {
        ApiResponse {
            success: true,
            data: Some(data),
            message: None,
        }
    }

    fn error(message: String) -> Self {
        ApiResponse {
            success: false,
            data: None,
            message: Some(message),
        }
    }
}

async fn create_device(
    State(state): State<Arc<AppState>>,
    Json(payload): Json<CreateDeviceRequest>,
) -> impl IntoResponse {
    let device = state.service.create_device(
        payload.name,
        payload.code,
        payload.description,
    );
    (StatusCode::CREATED, Json(ApiResponse::success(device)))
}

async fn get_devices(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let devices = state.service.storage().get_all_devices();
    Json(ApiResponse::success(devices))
}

async fn get_device(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.storage().get_device(id) {
        Some(device) => Json(ApiResponse::success(device)).into_response(),
        None => (
            StatusCode::NOT_FOUND,
            Json(ApiResponse::<Device>::error("设备不存在".to_string())),
        )
            .into_response(),
    }
}

async fn update_device_status(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
    Json(payload): Json<UpdateDeviceStatusRequest>,
) -> impl IntoResponse {
    match state.service.update_device_status(id, payload.status) {
        Some(device) => Json(ApiResponse::success(device)).into_response(),
        None => (
            StatusCode::NOT_FOUND,
            Json(ApiResponse::<Device>::error("设备不存在".to_string())),
        )
            .into_response(),
    }
}

async fn update_device_hours(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
    Json(payload): Json<UpdateDeviceHoursRequest>,
) -> impl IntoResponse {
    match state.service.update_device_running_hours(id, payload.running_hours) {
        Some(device) => Json(ApiResponse::success(device)).into_response(),
        None => (
            StatusCode::NOT_FOUND,
            Json(ApiResponse::<Device>::error("设备不存在".to_string())),
        )
            .into_response(),
    }
}

async fn get_device_history(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    let history = state.service.get_device_history(id);
    Json(ApiResponse::success(history))
}

async fn create_plan(
    State(state): State<Arc<AppState>>,
    Json(payload): Json<CreatePlanRequest>,
) -> impl IntoResponse {
    match state.service.create_maintenance_plan(
        payload.device_id,
        payload.name,
        payload.cycle_type,
        payload.cycle_value,
    ) {
        Some(plan) => (StatusCode::CREATED, Json(ApiResponse::success(plan))).into_response(),
        None => (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::<MaintenancePlan>::error("设备不存在".to_string())),
        )
            .into_response(),
    }
}

async fn get_plans(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let plans = state.service.storage().get_all_maintenance_plans();
    Json(ApiResponse::success(plans))
}

async fn get_device_plans(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    let plans = state.service.storage().get_maintenance_plans_for_device(id);
    Json(ApiResponse::success(plans))
}

async fn get_maintenance_orders(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let orders = state.service.storage().get_all_maintenance_orders();
    Json(ApiResponse::success(orders))
}

async fn get_device_maintenance_orders(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    let orders = state.service.storage().get_maintenance_orders_for_device(id);
    Json(ApiResponse::success(orders))
}

async fn complete_maintenance_order(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
    Json(payload): Json<CompleteMaintenanceOrderRequest>,
) -> impl IntoResponse {
    match state.service.complete_maintenance_order(id, payload.notes) {
        Ok(order) => Json(ApiResponse::success(order)).into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::<MaintenanceOrder>::error(e)),
        )
            .into_response(),
    }
}

async fn create_repair_order(
    State(state): State<Arc<AppState>>,
    Json(payload): Json<CreateRepairOrderRequest>,
) -> impl IntoResponse {
    match state.service.create_repair_order(payload.device_id, payload.title, payload.description) {
        Some(order) => (StatusCode::CREATED, Json(ApiResponse::success(order))).into_response(),
        None => (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::<RepairOrder>::error("设备不存在".to_string())),
        )
            .into_response(),
    }
}

async fn get_repair_orders(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let orders = state.service.storage().get_all_repair_orders();
    Json(ApiResponse::success(orders))
}

async fn get_device_repair_orders(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    let orders = state.service.storage().get_repair_orders_for_device(id);
    Json(ApiResponse::success(orders))
}

async fn complete_repair_order(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
    Json(payload): Json<CompleteRepairOrderRequest>,
) -> impl IntoResponse {
    match state.service.complete_repair_order(
        id,
        payload.fault_category,
        payload.spare_parts,
        payload.repair_notes,
    ) {
        Ok(order) => Json(ApiResponse::success(order)).into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::<RepairOrder>::error(e)),
        )
            .into_response(),
    }
}

async fn create_spare_part(
    State(state): State<Arc<AppState>>,
    Json(payload): Json<CreateSparePartRequest>,
) -> impl IntoResponse {
    let part = state.service.create_spare_part(
        payload.name,
        payload.code,
        payload.quantity,
        payload.safety_stock,
        payload.unit,
    );
    (StatusCode::CREATED, Json(ApiResponse::success(part)))
}

async fn get_spare_parts(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let parts = state.service.storage().get_all_spare_parts();
    Json(ApiResponse::success(parts))
}

async fn get_notifications(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let notifications = state.service.storage().get_all_notifications();
    Json(ApiResponse::success(notifications))
}

async fn mark_notification_read(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.mark_notification_read(id) {
        Some(n) => Json(ApiResponse::success(n)).into_response(),
        None => (
            StatusCode::NOT_FOUND,
            Json(ApiResponse::<Notification>::error("通知不存在".to_string())),
        )
            .into_response(),
    }
}

async fn trigger_daily_tasks(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    state.service.process_daily_tasks();
    Json(ApiResponse::success("Daily tasks processed"))
}

fn create_router(state: Arc<AppState>) -> Router {
    Router::new()
        .route("/api/devices", post(create_device).get(get_devices))
        .route("/api/devices/:id", get(get_device))
        .route("/api/devices/:id/status", put(update_device_status))
        .route("/api/devices/:id/hours", put(update_device_hours))
        .route("/api/devices/:id/history", get(get_device_history))
        .route(
            "/api/devices/:id/maintenance-plans",
            get(get_device_plans),
        )
        .route(
            "/api/devices/:id/maintenance-orders",
            get(get_device_maintenance_orders),
        )
        .route("/api/devices/:id/repair-orders", get(get_device_repair_orders))
        .route("/api/maintenance-plans", post(create_plan).get(get_plans))
        .route(
            "/api/maintenance-orders",
            get(get_maintenance_orders),
        )
        .route(
            "/api/maintenance-orders/:id/complete",
            post(complete_maintenance_order),
        )
        .route(
            "/api/repair-orders",
            post(create_repair_order).get(get_repair_orders),
        )
        .route(
            "/api/repair-orders/:id/complete",
            post(complete_repair_order),
        )
        .route("/api/spare-parts", post(create_spare_part).get(get_spare_parts))
        .route("/api/notifications", get(get_notifications))
        .route("/api/notifications/:id/read", post(mark_notification_read))
        .route("/api/daily-tasks", post(trigger_daily_tasks))
        .with_state(state)
}

#[tokio::main]
async fn main() {
    let args = Args::parse();

    let storage = InMemoryStorage::new();
    let service = MaintenanceService::new(storage);

    let state = Arc::new(AppState { service: service.clone() });

    let app = create_router(state);

    let addr: SocketAddr = format!("{}:{}", args.host, args.port)
        .parse()
        .expect("Invalid address");

    let daily_service = service.clone();
    tokio::spawn(async move {
        let mut interval = interval(Duration::from_secs(60));
        loop {
            interval.tick().await;
            daily_service.process_daily_tasks();
        }
    });

    println!("Server listening on {}", addr);

    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .expect("Failed to start server");
}
