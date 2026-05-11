use std::net::SocketAddr;
use std::sync::Arc;
use std::time::Duration;

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::{IntoResponse, Response},
    routing::{get, post},
    Json, Router,
};
use clap::Parser;
use tokio::time::interval;
use uuid::Uuid;

use complaint_core::*;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(long, env = "COMPLAINT_SERVER_HOST", default_value = "0.0.0.0")]
    host: String,

    #[arg(short, long, env = "COMPLAINT_SERVER_PORT", default_value_t = 8100)]
    port: u16,
}

#[derive(Clone)]
struct AppState {
    service: Arc<ComplaintService>,
}

fn error_response(status: StatusCode, message: impl Into<String>) -> Response {
    (status, Json(serde_json::json!({"error": message.into()}))).into_response()
}

fn handle_error(err: ComplaintError) -> Response {
    match err {
        ComplaintError::NotFound(id) => error_response(StatusCode::NOT_FOUND, format!("投诉不存在: {}", id)),
        ComplaintError::InvalidState => error_response(StatusCode::BAD_REQUEST, "投诉状态无效，无法执行该操作"),
        ComplaintError::HandlerNotFound(id) => error_response(StatusCode::BAD_REQUEST, format!("处理人不存在: {}", id)),
        ComplaintError::Internal(msg) => error_response(StatusCode::INTERNAL_SERVER_ERROR, msg),
    }
}

async fn health_check() -> &'static str {
    "OK"
}

async fn list_handlers(State(state): State<AppState>) -> Response {
    match state.service.list_handlers() {
        Ok(handlers) => (StatusCode::OK, Json(handlers)).into_response(),
        Err(e) => handle_error(e),
    }
}

async fn list_complaints(State(state): State<AppState>) -> Response {
    match state.service.list_complaints() {
        Ok(complaints) => (StatusCode::OK, Json(complaints)).into_response(),
        Err(e) => handle_error(e),
    }
}

async fn create_complaint(
    State(state): State<AppState>,
    Json(req): Json<CreateComplaintRequest>,
) -> Response {
    match state.service.create_complaint(req) {
        Ok(complaint) => (StatusCode::CREATED, Json(complaint)).into_response(),
        Err(e) => handle_error(e),
    }
}

async fn get_complaint(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Response {
    match state.service.get_complaint(id) {
        Ok(complaint) => (StatusCode::OK, Json(complaint)).into_response(),
        Err(e) => handle_error(e),
    }
}

async fn respond_to_complaint(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(req): Json<RespondToComplaintRequest>,
) -> Response {
    match state.service.respond_to_complaint(id, req) {
        Ok(complaint) => (StatusCode::OK, Json(complaint)).into_response(),
        Err(e) => handle_error(e),
    }
}

async fn propose_solution(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(req): Json<ProposeSolutionRequest>,
) -> Response {
    match state.service.propose_solution(id, req) {
        Ok(complaint) => (StatusCode::OK, Json(complaint)).into_response(),
        Err(e) => handle_error(e),
    }
}

async fn customer_feedback(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(req): Json<CustomerFeedbackRequest>,
) -> Response {
    match state.service.customer_feedback(id, req) {
        Ok(complaint) => (StatusCode::OK, Json(complaint)).into_response(),
        Err(e) => handle_error(e),
    }
}

async fn follow_up_feedback(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(req): Json<FollowUpFeedbackRequest>,
) -> Response {
    match state.service.follow_up_feedback(id, req) {
        Ok(complaint) => (StatusCode::OK, Json(complaint)).into_response(),
        Err(e) => handle_error(e),
    }
}

fn router(state: AppState) -> Router {
    Router::new()
        .route("/health", get(health_check))
        .route("/handlers", get(list_handlers))
        .route("/complaints", get(list_complaints).post(create_complaint))
        .route("/complaints/:id", get(get_complaint))
        .route("/complaints/:id/respond", post(respond_to_complaint))
        .route("/complaints/:id/solution", post(propose_solution))
        .route("/complaints/:id/feedback", post(customer_feedback))
        .route("/complaints/:id/followup", post(follow_up_feedback))
        .with_state(state)
}

fn start_escalation_checker(service: Arc<ComplaintService>) {
    tokio::spawn(async move {
        let mut ticker = interval(Duration::from_secs(60));
        loop {
            ticker.tick().await;
            let now = chrono::Utc::now();
            match service.check_and_escalate(now) {
                Ok(escalated) if !escalated.is_empty() => {
                    tracing::info!(count = escalated.len(), "检查超时投诉，已升级");
                }
                Ok(_) => {}
                Err(e) => {
                    tracing::error!(error = %e, "检查超时投诉失败");
                }
            }
        }
    });
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_max_level(tracing::Level::INFO)
        .init();

    let args = Args::parse();

    let store = InMemoryStore::new();
    let service = Arc::new(ComplaintService::new(store));
    let state = AppState { service: service.clone() };

    start_escalation_checker(service);

    let app = router(state);

    let addr = SocketAddr::from((args.host.parse::<std::net::IpAddr>().unwrap(), args.port));
    tracing::info!(%addr, "启动投诉SLA服务");

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
