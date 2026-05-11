use std::sync::Arc;

use axum::{
    extract::{Query, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post, put, delete},
    Json, Router,
};
use clap::Parser;
use serde::Deserialize;
use tower_http::cors::{Any, CorsLayer};
use uuid::Uuid;

use lab_report_core::{Report, ReportInput, ReportStorage};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "LAB_REPORT_PORT", default_value_t = 8080)]
    port: u16,

    #[arg(long, env = "LAB_REPORT_HOST", default_value = "0.0.0.0")]
    host: String,
}

#[derive(Clone)]
struct AppState {
    storage: Arc<ReportStorage>,
}

#[derive(Debug, Deserialize)]
struct SearchQuery {
    name: Option<String>,
    start_date: Option<String>,
    end_date: Option<String>,
}

#[derive(Debug, Deserialize)]
struct UpdateDoctorRequest {
    doctor_notes: Option<String>,
    diagnosis_suggestion: Option<String>,
}

async fn create_report(
    State(state): State<AppState>,
    Json(input): Json<ReportInput>,
) -> impl IntoResponse {
    let report = state.storage.create_report(input).await;
    (StatusCode::CREATED, Json(report))
}

async fn create_reports_batch(
    State(state): State<AppState>,
    Json(inputs): Json<Vec<ReportInput>>,
) -> impl IntoResponse {
    let reports = state.storage.create_reports_batch(inputs).await;
    (StatusCode::CREATED, Json(reports))
}

async fn get_all_reports(State(state): State<AppState>) -> impl IntoResponse {
    let reports = state.storage.get_all_reports().await;
    Json(reports)
}

async fn get_report(
    State(state): State<AppState>,
    axum::extract::Path(id): axum::extract::Path<Uuid>,
) -> impl IntoResponse {
    match state.storage.get_report(&id).await {
        Some(report) => (StatusCode::OK, Json(Some(report))),
        None => (StatusCode::NOT_FOUND, Json(None)),
    }
}

async fn search_reports(
    State(state): State<AppState>,
    query: Query<SearchQuery>,
) -> impl IntoResponse {
    let start_date = query
        .start_date
        .as_deref()
        .and_then(|s| chrono::NaiveDate::parse_from_str(s, "%Y-%m-%d").ok());
    
    let end_date = query
        .end_date
        .as_deref()
        .and_then(|s| chrono::NaiveDate::parse_from_str(s, "%Y-%m-%d").ok());
    
    let reports = state
        .storage
        .search_reports(
            query.name.as_deref(),
            start_date,
            end_date,
        )
        .await;
    
    Json(reports)
}

async fn update_doctor_info(
    State(state): State<AppState>,
    axum::extract::Path(id): axum::extract::Path<Uuid>,
    Json(input): Json<UpdateDoctorRequest>,
) -> impl IntoResponse {
    match state
        .storage
        .update_doctor_info(
            &id,
            input.doctor_notes,
            input.diagnosis_suggestion,
        )
        .await
    {
        Some(report) => (StatusCode::OK, Json(Some(report))),
        None => (StatusCode::NOT_FOUND, Json(None::<Report>)),
    }
}

async fn delete_report(
    State(state): State<AppState>,
    axum::extract::Path(id): axum::extract::Path<Uuid>,
) -> impl IntoResponse {
    if state.storage.delete_report(&id).await {
        StatusCode::NO_CONTENT
    } else {
        StatusCode::NOT_FOUND
    }
}

async fn health_check() -> impl IntoResponse {
    Json(serde_json::json!({"status": "ok"}))
}

#[tokio::main]
async fn main() {
    let args = Args::parse();
    
    let storage = Arc::new(ReportStorage::new());
    let state = AppState { storage };
    
    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);
    
    let app = Router::new()
        .route("/health", get(health_check))
        .route("/reports", get(get_all_reports))
        .route("/reports", post(create_report))
        .route("/reports/batch", post(create_reports_batch))
        .route("/reports/search", get(search_reports))
        .route("/reports/:id", get(get_report))
        .route("/reports/:id/doctor", put(update_doctor_info))
        .route("/reports/:id", delete(delete_report))
        .with_state(state)
        .layer(cors);
    
    let addr = format!("{}:{}", args.host, args.port);
    println!("体检检验报告管理系统服务启动于: http://{}", addr);
    
    axum::Server::bind(&addr.parse().unwrap())
        .serve(app.into_make_service())
        .await
        .unwrap();
}
