use std::sync::Arc;

use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post, put},
    Json, Router,
};
use chrono::{DateTime, Utc};
use clap::Parser;
use maintenance_core::*;
use serde::Deserialize;
use tower_http::trace::TraceLayer;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

#[derive(Debug, Parser)]
#[command(author, version, about, long_about = None)]
struct Config {
    #[arg(long, env = "MAINTENANCE_PORT", default_value_t = 3000)]
    port: u16,

    #[arg(long, env = "MAINTENANCE_HOST", default_value = "127.0.0.1")]
    host: String,
}

type AppState = Arc<MaintenanceService>;

#[derive(Debug, Deserialize)]
struct StatsQuery {
    start: Option<String>,
    end: Option<String>,
    equipment: Option<String>,
    personnel: Option<String>,
}

#[tokio::main]
async fn main() {
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "maintenance_server=debug,tower_http=debug".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    let config = Config::parse();
    let service = Arc::new(MaintenanceService::new());

    let app = Router::new()
        .route("/health", get(health_check))
        .route("/equipment", post(create_equipment))
        .route("/equipment", get(list_equipments))
        .route("/equipment/:id", get(get_equipment))
        .route("/equipment/:id/runtime", put(update_runtime))
        .route("/equipment/:id/maintenance", post(perform_maintenance))
        .route("/equipment/:id/records", get(get_equipment_records))
        .route("/reminders", get(get_reminders))
        .route("/records", get(list_records))
        .route("/records/:id", get(get_record))
        .route("/personnel", get(list_personnel))
        .route("/stats", get(get_stats))
        .layer(TraceLayer::new_for_http())
        .with_state(service);

    let addr = format!("{}:{}", config.host, config.port);
    tracing::info!("Server listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}

async fn health_check() -> impl IntoResponse {
    Json(serde_json::json!({"status": "ok"}))
}

async fn create_equipment(
    State(state): State<AppState>,
    Json(req): Json<CreateEquipmentRequest>,
) -> impl IntoResponse {
    let equipment = state.create_equipment(req);
    (StatusCode::CREATED, Json(equipment))
}

async fn list_equipments(State(state): State<AppState>) -> impl IntoResponse {
    let equipments = state.list_equipments();
    Json(equipments)
}

async fn get_equipment(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match state.get_equipment(&id) {
        Some(eq) => Json(serde_json::json!(eq)).into_response(),
        None => (StatusCode::NOT_FOUND, Json(serde_json::json!({"error": "Equipment not found"}))).into_response(),
    }
}

async fn update_runtime(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(req): Json<UpdateRuntimeRequest>,
) -> impl IntoResponse {
    match state.update_runtime(&id, req.additional_hours) {
        Some(eq) => Json(eq).into_response(),
        None => (StatusCode::NOT_FOUND, Json(serde_json::json!({"error": "Equipment not found"}))).into_response(),
    }
}

async fn perform_maintenance(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(req): Json<PerformMaintenanceRequest>,
) -> impl IntoResponse {
    match state.perform_maintenance(&id, req) {
        Some(record) => (StatusCode::CREATED, Json(record)).into_response(),
        None => (StatusCode::NOT_FOUND, Json(serde_json::json!({"error": "Equipment not found"}))).into_response(),
    }
}

async fn get_equipment_records(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    if state.get_equipment(&id).is_none() {
        return (StatusCode::NOT_FOUND, Json(serde_json::json!({"error": "Equipment not found"}))).into_response();
    }
    let records = state.get_records_by_equipment(&id);
    Json(records).into_response()
}

async fn get_reminders(State(state): State<AppState>) -> impl IntoResponse {
    let now = Utc::now();
    let reminders = state.get_all_reminders(now);
    Json(reminders)
}

async fn list_records(State(state): State<AppState>) -> impl IntoResponse {
    let records = state.list_all_records();
    Json(records)
}

async fn get_record(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    let all = state.list_all_records();
    match all.into_iter().find(|r| r.id == id) {
        Some(record) => Json(serde_json::json!(record)).into_response(),
        None => (StatusCode::NOT_FOUND, Json(serde_json::json!({"error": "Record not found"}))).into_response(),
    }
}

async fn list_personnel(State(state): State<AppState>) -> impl IntoResponse {
    let personnel = state.list_personnel();
    Json(personnel)
}

async fn get_stats(
    State(state): State<AppState>,
    Query(query): Query<StatsQuery>,
) -> impl IntoResponse {
    if let Some(personnel) = query.personnel {
        let stats = state.stats_by_personnel(&personnel);
        return Json(stats).into_response();
    }

    if let Some(equipment) = query.equipment {
        if state.get_equipment(&equipment).is_none() {
            return (StatusCode::NOT_FOUND, Json(serde_json::json!({"error": "Equipment not found"}))).into_response();
        }
        let stats = state.stats_by_equipment(&equipment);
        return Json(stats).into_response();
    }

    if let (Some(start_str), Some(end_str)) = (query.start, query.end) {
        let start = match DateTime::parse_from_rfc3339(&start_str) {
            Ok(dt) => dt.with_timezone(&Utc),
            Err(_) => {
                return (
                    StatusCode::BAD_REQUEST,
                    Json(serde_json::json!({"error": "Invalid start date format"})),
                ).into_response();
            }
        };
        let end = match DateTime::parse_from_rfc3339(&end_str) {
            Ok(dt) => dt.with_timezone(&Utc),
            Err(_) => {
                return (
                    StatusCode::BAD_REQUEST,
                    Json(serde_json::json!({"error": "Invalid end date format"})),
                ).into_response();
            }
        };
        let stats = state.stats_by_time_range(start, end);
        return Json(stats).into_response();
    }

    (
        StatusCode::BAD_REQUEST,
        Json(serde_json::json!({"error": "Must specify equipment, personnel, or time range (start and end)"})),
    ).into_response()
}
