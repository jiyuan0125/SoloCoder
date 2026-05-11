use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use checkin_core::{
    ApiResponse, CheckinRequest, CheckinService, CheckoutRequest, CreateActivityRequest,
    CreateEmployeeRequest,
};
use clap::Parser;
use std::net::SocketAddr;
use std::sync::Arc;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "CHECKIN_PORT", default_value_t = 3000)]
    port: u16,

    #[arg(short = 'H', long, env = "CHECKIN_HOST", default_value = "0.0.0.0")]
    host: String,
}

struct AppState {
    service: CheckinService,
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();

    let args = Args::parse();

    let service = CheckinService::new();
    let app_state = Arc::new(AppState { service });

    let app = Router::new()
        .route("/api/employees", post(create_employee))
        .route("/api/employees", get(list_employees))
        .route("/api/employees/:id", get(get_employee))
        .route("/api/activities", post(create_activity))
        .route("/api/activities", get(list_activities))
        .route("/api/activities/:id", get(get_activity))
        .route("/api/activities/:id/checkin", post(checkin))
        .route("/api/activities/:id/checkout", post(checkout))
        .route("/api/activities/:id/report", get(get_report))
        .route("/api/activities/:id/records/:employee_id", get(get_record))
        .with_state(app_state);

    let addr: SocketAddr = format!("{}:{}", args.host, args.port).parse().unwrap();
    tracing::info!("签到服务启动于 {}", addr);

    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}

async fn create_employee(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateEmployeeRequest>,
) -> impl IntoResponse {
    let employee = state.service.create_employee(req);
    (StatusCode::OK, Json(ApiResponse::success(employee)))
}

async fn list_employees(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let employees = state.service.list_employees();
    (StatusCode::OK, Json(ApiResponse::success(employees)))
}

async fn get_employee(
    State(state): State<Arc<AppState>>,
    Path(id): Path<uuid::Uuid>,
) -> impl IntoResponse {
    match state.service.get_employee(id) {
        Some(emp) => (StatusCode::OK, Json(ApiResponse::success(emp))),
        None => (
            StatusCode::NOT_FOUND,
            Json(ApiResponse::error("员工不存在".to_string())),
        ),
    }
}

async fn create_activity(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateActivityRequest>,
) -> impl IntoResponse {
    let activity = state.service.create_activity(req);
    (StatusCode::OK, Json(ApiResponse::success(activity)))
}

async fn list_activities(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let activities = state.service.list_activities();
    (StatusCode::OK, Json(ApiResponse::success(activities)))
}

async fn get_activity(
    State(state): State<Arc<AppState>>,
    Path(id): Path<uuid::Uuid>,
) -> impl IntoResponse {
    match state.service.get_activity(id) {
        Some(act) => (StatusCode::OK, Json(ApiResponse::success(act))),
        None => (
            StatusCode::NOT_FOUND,
            Json(ApiResponse::error("活动不存在".to_string())),
        ),
    }
}

async fn checkin(
    State(state): State<Arc<AppState>>,
    Path(activity_id): Path<uuid::Uuid>,
    Json(req): Json<CheckinRequest>,
) -> impl IntoResponse {
    match state.service.checkin(activity_id, req, chrono::Utc::now()) {
        Ok(record) => (
            StatusCode::OK,
            Json(ApiResponse::success_with_message("签到成功".to_string(), record)),
        ),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::error(e.to_string())),
        ),
    }
}

async fn checkout(
    State(state): State<Arc<AppState>>,
    Path(activity_id): Path<uuid::Uuid>,
    Json(req): Json<CheckoutRequest>,
) -> impl IntoResponse {
    let employee_id = req.employee_id;
    match state.service.checkout(activity_id, req, chrono::Utc::now()) {
        Ok(record) => (
            StatusCode::OK,
            Json(ApiResponse::success_with_message("签退成功".to_string(), record)),
        ),
        Err(checkin_core::CheckinError::EarlyCheckout) => {
            let record = state
                .service
                .get_record(activity_id, employee_id)
                .unwrap();
            (
                StatusCode::OK,
                Json(ApiResponse::success_with_message(
                    "签退成功（记为早退）".to_string(),
                    record,
                )),
            )
        }
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::error(e.to_string())),
        ),
    }
}

async fn get_report(
    State(state): State<Arc<AppState>>,
    Path(activity_id): Path<uuid::Uuid>,
) -> impl IntoResponse {
    match state.service.get_report(activity_id) {
        Ok(report) => (StatusCode::OK, Json(ApiResponse::success(report))),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::error(e.to_string())),
        ),
    }
}

async fn get_record(
    State(state): State<Arc<AppState>>,
    Path((activity_id, employee_id)): Path<(uuid::Uuid, uuid::Uuid)>,
) -> impl IntoResponse {
    match state.service.get_record(activity_id, employee_id) {
        Some(record) => (StatusCode::OK, Json(ApiResponse::success(record))),
        None => (
            StatusCode::NOT_FOUND,
            Json(ApiResponse::error("签到记录不存在".to_string())),
        ),
    }
}
