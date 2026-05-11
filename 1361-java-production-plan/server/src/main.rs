use std::sync::Arc;

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post, put},
    Json, Router,
};
use clap::Parser;
use scheduler_core::{
    Device, InMemoryStore, MaintenanceWindow, Order, ProductionLine, Scheduler,
};
use serde::Deserialize;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(long, env = "PORT", default_value = "8080")]
    port: u16,

    #[arg(long, env = "HOST", default_value = "0.0.0.0")]
    host: String,
}

struct AppState {
    scheduler: Scheduler,
    store: InMemoryStore,
}

#[derive(Debug, Deserialize)]
struct CreateLineRequest {
    name: String,
    devices: Vec<CreateDeviceRequest>,
}

#[derive(Debug, Deserialize)]
struct CreateDeviceRequest {
    id: String,
    name: String,
}

#[derive(Debug, Deserialize)]
struct CreateOrderRequest {
    id: String,
    name: String,
    priority: String,
    duration_minutes: u32,
    due_date: String,
}

#[derive(Debug, Deserialize)]
struct CreateMaintenanceRequest {
    id: String,
    device_id: String,
    start_time: String,
    end_time: String,
    description: Option<String>,
}

#[derive(Debug, Deserialize)]
struct UpdateMaintenanceRequest {
    device_id: String,
    start_time: String,
    end_time: String,
    description: Option<String>,
}

fn parse_datetime(s: &str) -> Result<chrono::DateTime<chrono::Utc>, String> {
    chrono::DateTime::parse_from_rfc3339(s)
        .map(|dt| dt.with_timezone(&chrono::Utc))
        .or_else(|_| {
            chrono::NaiveDateTime::parse_from_str(s, "%Y-%m-%d %H:%M:%S")
                .map(|nd| chrono::DateTime::<chrono::Utc>::from_naive_utc_and_offset(nd, chrono::Utc))
        })
        .map_err(|e| format!("Invalid datetime: {}", e))
}

async fn create_line(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateLineRequest>,
) -> impl IntoResponse {
    let line_id = uuid::Uuid::new_v4().to_string();
    let devices: Vec<Device> = req.devices
        .into_iter()
        .map(|d| Device {
            id: d.id,
            name: d.name,
        })
        .collect();

    let line = ProductionLine {
        id: line_id.clone(),
        name: req.name,
        devices,
    };

    state.store.add_line(line.clone());

    (StatusCode::CREATED, Json(line))
}

async fn get_lines(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let lines = state.store.get_all_lines();
    Json(lines)
}

async fn get_line(
    State(state): State<Arc<AppState>>,
    Path(line_id): Path<String>,
) -> impl IntoResponse {
    match state.store.get_line(&line_id) {
        Some(line) => (StatusCode::OK, Json(serde_json::json!(line))).into_response(),
        None => (StatusCode::NOT_FOUND, Json(serde_json::json!({"error": "Line not found"}))).into_response(),
    }
}

async fn create_order(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateOrderRequest>,
) -> impl IntoResponse {
    let due_date = match parse_datetime(&req.due_date) {
        Ok(d) => d,
        Err(e) => return (StatusCode::BAD_REQUEST, Json(serde_json::json!({"error": e}))).into_response(),
    };

    let priority = match req.priority.to_lowercase().as_str() {
        "urgent" => scheduler_core::OrderPriority::Urgent,
        "normal" | _ => scheduler_core::OrderPriority::Normal,
    };

    let order = Order {
        id: req.id,
        name: req.name,
        priority,
        duration_minutes: req.duration_minutes,
        due_date,
        assigned_line_id: None,
    };

    let result = match priority {
        scheduler_core::OrderPriority::Urgent => state.scheduler.add_urgent_order(order),
        scheduler_core::OrderPriority::Normal => state.scheduler.add_normal_order(order),
    };

    match result {
        Ok(schedule_result) => (StatusCode::CREATED, Json(serde_json::json!(schedule_result))).into_response(),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({"error": e.to_string()}))).into_response(),
    }
}

async fn get_orders(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let orders = state.store.get_all_orders();
    Json(orders)
}

async fn get_order(
    State(state): State<Arc<AppState>>,
    Path(order_id): Path<String>,
) -> impl IntoResponse {
    match state.store.get_order(&order_id) {
        Some(order) => (StatusCode::OK, Json(serde_json::json!(order))).into_response(),
        None => (StatusCode::NOT_FOUND, Json(serde_json::json!({"error": "Order not found"}))).into_response(),
    }
}

async fn create_maintenance(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateMaintenanceRequest>,
) -> impl IntoResponse {
    let start_time = match parse_datetime(&req.start_time) {
        Ok(t) => t,
        Err(e) => return (StatusCode::BAD_REQUEST, Json(serde_json::json!({"error": e}))).into_response(),
    };

    let end_time = match parse_datetime(&req.end_time) {
        Ok(t) => t,
        Err(e) => return (StatusCode::BAD_REQUEST, Json(serde_json::json!({"error": e}))).into_response(),
    };

    let window = MaintenanceWindow {
        id: req.id,
        device_id: req.device_id,
        start_time,
        end_time,
        description: req.description,
    };

    match state.store.add_maintenance_window(window.clone()) {
        Ok(_) => {
            if let Err(e) = state.scheduler.reschedule_all() {
                return (StatusCode::INTERNAL_SERVER_ERROR, Json(serde_json::json!({"error": e.to_string()}))).into_response();
            }
            (StatusCode::CREATED, Json(window)).into_response()
        }
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({"error": e.to_string()}))).into_response(),
    }
}

async fn update_maintenance(
    State(state): State<Arc<AppState>>,
    Path(window_id): Path<String>,
    Json(req): Json<UpdateMaintenanceRequest>,
) -> impl IntoResponse {
    let start_time = match parse_datetime(&req.start_time) {
        Ok(t) => t,
        Err(e) => return (StatusCode::BAD_REQUEST, Json(serde_json::json!({"error": e}))).into_response(),
    };

    let end_time = match parse_datetime(&req.end_time) {
        Ok(t) => t,
        Err(e) => return (StatusCode::BAD_REQUEST, Json(serde_json::json!({"error": e}))).into_response(),
    };

    let window = MaintenanceWindow {
        id: window_id,
        device_id: req.device_id,
        start_time,
        end_time,
        description: req.description,
    };

    match state.store.update_maintenance_window(window.clone()) {
        Ok(_) => {
            match state.scheduler.reschedule_all() {
                Ok(result) => (StatusCode::OK, Json(serde_json::json!({
                    "window": window,
                    "affected_schedule": result
                }))).into_response(),
                Err(e) => (StatusCode::INTERNAL_SERVER_ERROR, Json(serde_json::json!({"error": e.to_string()}))).into_response(),
            }
        }
        Err(e) => (StatusCode::NOT_FOUND, Json(serde_json::json!({"error": e.to_string()}))).into_response(),
    }
}

async fn delete_maintenance(
    State(state): State<Arc<AppState>>,
    Path(window_id): Path<String>,
) -> impl IntoResponse {
    match state.store.delete_maintenance_window(&window_id) {
        Ok(_) => {
            match state.scheduler.reschedule_all() {
                Ok(result) => (StatusCode::OK, Json(serde_json::json!({
                    "message": "Maintenance window deleted",
                    "affected_schedule": result
                }))).into_response(),
                Err(e) => (StatusCode::INTERNAL_SERVER_ERROR, Json(serde_json::json!({"error": e.to_string()}))).into_response(),
            }
        }
        Err(e) => (StatusCode::NOT_FOUND, Json(serde_json::json!({"error": e.to_string()}))).into_response(),
    }
}

async fn get_maintenances(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let windows = state.store.get_all_maintenance_windows();
    Json(windows)
}

async fn get_schedule(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let schedule = state.scheduler.get_schedule();
    Json(schedule)
}

async fn get_line_schedule(
    State(state): State<Arc<AppState>>,
    Path(line_id): Path<String>,
) -> impl IntoResponse {
    let schedule = state.scheduler.get_schedule_for_line(&line_id);
    Json(schedule)
}

async fn reschedule_all(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    match state.scheduler.reschedule_all() {
        Ok(result) => (StatusCode::OK, Json(result)).into_response(),
        Err(e) => (StatusCode::INTERNAL_SERVER_ERROR, Json(serde_json::json!({"error": e.to_string()}))).into_response(),
    }
}

#[tokio::main]
async fn main() {
    let args = Args::parse();

    let store = InMemoryStore::new();
    let scheduler = Scheduler::new(store.clone());

    let state = Arc::new(AppState { scheduler, store });

    let app = Router::new()
        .route("/lines", post(create_line).get(get_lines))
        .route("/lines/:line_id", get(get_line))
        .route("/lines/:line_id/schedule", get(get_line_schedule))
        .route("/orders", post(create_order).get(get_orders))
        .route("/orders/:order_id", get(get_order))
        .route("/maintenances", post(create_maintenance).get(get_maintenances))
        .route("/maintenances/:window_id", put(update_maintenance).delete(delete_maintenance))
        .route("/schedule", get(get_schedule).post(reschedule_all))
        .with_state(state);

    let addr = format!("{}:{}", args.host, args.port);
    println!("Scheduler server starting on {}", addr);

    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
