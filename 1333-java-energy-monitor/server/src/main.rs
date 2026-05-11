use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post, put, delete},
    Json, Router,
};
use clap::Parser;
use energy_core::{
    Alert, DailyCostBreakdown, ElectricityUsage, InMemoryStorage, PointId, TariffConfig,
    UsageRecord, calculate_daily_cost, check_all_alerts, check_alert_for_point, NaiveDate,
};
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use tower_http::cors::{Any, CorsLayer};
use chrono::Local;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "ENERGY_SERVER_PORT", default_value_t = 8080)]
    port: u16,

    #[arg(short, long, env = "ENERGY_SERVER_HOST", default_value = "127.0.0.1")]
    host: String,
}

struct AppState {
    storage: Arc<InMemoryStorage>,
    tariff: Arc<std::sync::RwLock<TariffConfig>>,
}

#[derive(Debug, Deserialize)]
struct RecordRequest {
    point_id: String,
    date: String,
    peak: f64,
    flat: f64,
    valley: f64,
}

#[derive(Debug, Serialize)]
struct RecordResponse {
    point_id: String,
    date: String,
    peak: f64,
    flat: f64,
    valley: f64,
    total_usage: f64,
    cost: DailyCostBreakdown,
}

#[derive(Debug, Deserialize)]
struct TariffUpdateRequest {
    peak_price: Option<f64>,
    flat_price: Option<f64>,
    valley_price: Option<f64>,
}

#[derive(Debug, Deserialize)]
struct ShutdownRequest {
    date: String,
}

#[derive(Debug, Serialize)]
struct ApiError {
    error: String,
}

impl IntoResponse for ApiError {
    fn into_response(self) -> axum::response::Response {
        (StatusCode::BAD_REQUEST, Json(self)).into_response()
    }
}

fn parse_date(date_str: &str) -> Result<NaiveDate, ApiError> {
    NaiveDate::parse_from_str(date_str, "%Y-%m-%d")
        .map_err(|e| ApiError { error: format!("Invalid date format: {}", e) })
}

async fn upsert_record(
    State(state): State<Arc<AppState>>,
    Json(req): Json<RecordRequest>,
) -> Result<impl IntoResponse, ApiError> {
    let date = parse_date(&req.date)?;
    let point_id = PointId::new(req.point_id);
    let usage = ElectricityUsage::new(req.peak, req.flat, req.valley);
    let record = UsageRecord::new(point_id.clone(), date, usage.clone());

    state.storage.upsert_record(record);

    let tariff = state.tariff.read().unwrap();
    let cost = calculate_daily_cost(&usage, &tariff);

    Ok(Json(RecordResponse {
        point_id: point_id.to_string(),
        date: date.to_string(),
        peak: usage.peak,
        flat: usage.flat,
        valley: usage.valley,
        total_usage: usage.total(),
        cost,
    }))
}

async fn get_record(
    State(state): State<Arc<AppState>>,
    Path((point_id, date_str)): Path<(String, String)>,
) -> Result<impl IntoResponse, ApiError> {
    let date = parse_date(&date_str)?;
    let point_id = PointId::new(point_id);

    match state.storage.get_record(&point_id, &date) {
        Some(record) => {
            let tariff = state.tariff.read().unwrap();
            let cost = calculate_daily_cost(&record.usage, &tariff);
            Ok(Json(RecordResponse {
                point_id: record.point_id.to_string(),
                date: record.date.to_string(),
                peak: record.usage.peak,
                flat: record.usage.flat,
                valley: record.usage.valley,
                total_usage: record.usage.total(),
                cost,
            }))
        }
        None => Err(ApiError { error: "Record not found".to_string() }),
    }
}

async fn get_all_records(
    State(state): State<Arc<AppState>>,
    Path(point_id): Path<String>,
) -> impl IntoResponse {
    let point_id = PointId::new(point_id);
    let records = state.storage.get_all_records_for_point(&point_id);
    let tariff = state.tariff.read().unwrap();

    let responses: Vec<RecordResponse> = records
        .into_iter()
        .map(|r| {
            let cost = calculate_daily_cost(&r.usage, &tariff);
            RecordResponse {
                point_id: r.point_id.to_string(),
                date: r.date.to_string(),
                peak: r.usage.peak,
                flat: r.usage.flat,
                valley: r.usage.valley,
                total_usage: r.usage.total(),
                cost,
            }
        })
        .collect();

    Json(responses)
}

async fn get_tariff(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let tariff = state.tariff.read().unwrap();
    Json(tariff.clone())
}

async fn update_tariff(
    State(state): State<Arc<AppState>>,
    Json(req): Json<TariffUpdateRequest>,
) -> impl IntoResponse {
    let mut tariff = state.tariff.write().unwrap();
    if let Some(p) = req.peak_price {
        tariff.peak_price = p;
    }
    if let Some(p) = req.flat_price {
        tariff.flat_price = p;
    }
    if let Some(p) = req.valley_price {
        tariff.valley_price = p;
    }
    Json(tariff.clone())
}

async fn mark_shutdown(
    State(state): State<Arc<AppState>>,
    Json(req): Json<ShutdownRequest>,
) -> Result<impl IntoResponse, ApiError> {
    let date = parse_date(&req.date)?;
    state.storage.mark_shutdown(date);
    Ok(Json(serde_json::json!({
        "status": "ok",
        "date": date.to_string(),
        "action": "marked_shutdown"
    })))
}

async fn unmark_shutdown(
    State(state): State<Arc<AppState>>,
    Path(date_str): Path<String>,
) -> Result<impl IntoResponse, ApiError> {
    let date = parse_date(&date_str)?;
    state.storage.unmark_shutdown(date);
    Ok(Json(serde_json::json!({
        "status": "ok",
        "date": date.to_string(),
        "action": "unmarked_shutdown"
    })))
}

async fn check_point_alerts(
    State(state): State<Arc<AppState>>,
    Path(point_id): Path<String>,
) -> Result<impl IntoResponse, ApiError> {
    let today = Local::now().date_naive();
    let point_id = PointId::new(point_id);

    match check_alert_for_point(&state.storage, &point_id, &today) {
        Some(alert) => Ok(Json(Some(alert))),
        None => Ok(Json(None::<Alert>)),
    }
}

async fn check_all_points_alerts(
    State(state): State<Arc<AppState>>,
) -> impl IntoResponse {
    let today = Local::now().date_naive();
    let alerts = check_all_alerts(state.storage.clone(), today);
    Json(alerts)
}

#[tokio::main]
async fn main() {
    let args = Args::parse();

    let storage = InMemoryStorage::new();
    let tariff = Arc::new(std::sync::RwLock::new(TariffConfig::standard()));

    let app_state = Arc::new(AppState { storage, tariff });

    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    let app = Router::new()
        .route("/records", post(upsert_record))
        .route("/records/:point_id/:date", get(get_record))
        .route("/records/:point_id", get(get_all_records))
        .route("/tariff", get(get_tariff))
        .route("/tariff", put(update_tariff))
        .route("/shutdown", post(mark_shutdown))
        .route("/shutdown/:date", delete(unmark_shutdown))
        .route("/alerts/:point_id", get(check_point_alerts))
        .route("/alerts", get(check_all_points_alerts))
        .with_state(app_state)
        .layer(cors);

    let addr = format!("{}:{}", args.host, args.port);
    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();

    println!("Energy Monitor Server running on {}", addr);

    axum::serve(listener, app).await.unwrap();
}
