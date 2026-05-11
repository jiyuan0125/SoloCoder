use std::sync::Arc;
use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    response::{IntoResponse, Json},
    routing::{get, post},
    Router,
};
use clap::Parser;
use serde::Deserialize;
use tower_http::cors::{Any, CorsLayer};
use uuid::Uuid;

use otl_core::{
    InMemoryStore,
    OvertimeLeaveService,
    models::*,
    errors::SystemError,
};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "PORT", default_value_t = 3000)]
    port: u16,
}

#[derive(Clone)]
struct AppState {
    service: Arc<OvertimeLeaveService>,
}

#[derive(Debug, Deserialize)]
struct ListQuery {
    employee_id: Option<Uuid>,
}

async fn create_employee(
    State(state): State<AppState>,
    Json(req): Json<CreateEmployeeRequest>,
) -> impl IntoResponse {
    let employee = state.service.create_employee(req);
    (StatusCode::CREATED, Json(employee))
}

async fn list_employees(State(state): State<AppState>) -> impl IntoResponse {
    let employees = state.service.list_employees();
    Json(employees)
}

async fn get_employee(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.get_employee(id) {
        Ok(employee) => Json(employee).into_response(),
        Err(e) => error_response(e),
    }
}

async fn create_overtime(
    State(state): State<AppState>,
    Json(req): Json<CreateOvertimeRequest>,
) -> impl IntoResponse {
    match state.service.create_overtime(req) {
        Ok(overtime) => (StatusCode::CREATED, Json(overtime)).into_response(),
        Err(e) => error_response(e),
    }
}

async fn approve_overtime(
    State(state): State<AppState>,
    Json(req): Json<ApproveOvertimeRequest>,
) -> impl IntoResponse {
    match state.service.approve_overtime(req) {
        Ok(overtime) => Json(overtime).into_response(),
        Err(e) => error_response(e),
    }
}

async fn reject_overtime(
    State(state): State<AppState>,
    Json(req): Json<RejectOvertimeRequest>,
) -> impl IntoResponse {
    match state.service.reject_overtime(req) {
        Ok(overtime) => Json(overtime).into_response(),
        Err(e) => error_response(e),
    }
}

async fn list_overtimes(
    State(state): State<AppState>,
    Query(query): Query<ListQuery>,
) -> impl IntoResponse {
    let overtimes = state.service.list_overtimes(query.employee_id);
    Json(overtimes)
}

async fn create_leave(
    State(state): State<AppState>,
    Json(req): Json<CreateLeaveRequest>,
) -> impl IntoResponse {
    match state.service.create_leave(req) {
        Ok(leave) => (StatusCode::CREATED, Json(leave)).into_response(),
        Err(e) => error_response(e),
    }
}

async fn approve_leave(
    State(state): State<AppState>,
    Json(req): Json<ApproveLeaveRequest>,
) -> impl IntoResponse {
    match state.service.approve_leave(req) {
        Ok(leave) => Json(leave).into_response(),
        Err(e) => error_response(e),
    }
}

async fn reject_leave(
    State(state): State<AppState>,
    Json(req): Json<RejectLeaveRequest>,
) -> impl IntoResponse {
    match state.service.reject_leave(req) {
        Ok(leave) => Json(leave).into_response(),
        Err(e) => error_response(e),
    }
}

async fn cancel_leave(
    State(state): State<AppState>,
    Json(req): Json<CancelLeaveRequest>,
) -> impl IntoResponse {
    match state.service.cancel_leave(req) {
        Ok(leave) => Json(leave).into_response(),
        Err(e) => error_response(e),
    }
}

async fn list_leaves(
    State(state): State<AppState>,
    Query(query): Query<ListQuery>,
) -> impl IntoResponse {
    let leaves = state.service.list_leaves(query.employee_id);
    Json(leaves)
}

async fn create_holiday(
    State(state): State<AppState>,
    Json(req): Json<CreateHolidayRequest>,
) -> impl IntoResponse {
    let holiday = state.service.create_holiday(req);
    (StatusCode::CREATED, Json(holiday))
}

async fn list_holidays(State(state): State<AppState>) -> impl IntoResponse {
    let holidays = state.service.list_holidays();
    Json(holidays)
}

async fn get_leave_balance(
    State(state): State<AppState>,
    Path(employee_id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.get_leave_balance(employee_id) {
        Ok(balance) => Json(serde_json::json!({
            "employee_id": employee_id,
            "balance_hours": balance
        })).into_response(),
        Err(e) => error_response(e),
    }
}

fn error_response(e: SystemError) -> axum::response::Response {
    let status = match &e {
        SystemError::EmployeeNotFound(_) => StatusCode::NOT_FOUND,
        SystemError::OvertimeNotFound(_) => StatusCode::NOT_FOUND,
        SystemError::LeaveNotFound(_) => StatusCode::NOT_FOUND,
        SystemError::OvertimeExpired(_) => StatusCode::BAD_REQUEST,
        SystemError::OvertimeAlreadyProcessed(_) => StatusCode::BAD_REQUEST,
        SystemError::LeaveAlreadyProcessed(_) => StatusCode::BAD_REQUEST,
        SystemError::InsufficientLeaveBalance(_, _, _) => StatusCode::BAD_REQUEST,
        SystemError::CannotCancelPastLeave(_) => StatusCode::BAD_REQUEST,
        SystemError::WeekendCompensationRequired => StatusCode::BAD_REQUEST,
        SystemError::WeekendCompensationCannotChange => StatusCode::BAD_REQUEST,
        SystemError::HolidayCannotBeLeave => StatusCode::BAD_REQUEST,
        SystemError::InvalidOvertimePeriod => StatusCode::BAD_REQUEST,
        SystemError::InvalidLeavePeriod => StatusCode::BAD_REQUEST,
        SystemError::InternalError(_) => StatusCode::INTERNAL_SERVER_ERROR,
    };
    (status, Json(serde_json::json!({
        "error": e.to_string()
    }))).into_response()
}

#[tokio::main]
async fn main() {
    let args = Args::parse();

    let store = Arc::new(InMemoryStore::new());
    let service = Arc::new(OvertimeLeaveService::new(store));
    let state = AppState { service };

    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    let app = Router::new()
        .route("/api/employees", post(create_employee).get(list_employees))
        .route("/api/employees/:id", get(get_employee))
        .route("/api/employees/:id/balance", get(get_leave_balance))
        .route("/api/overtimes", post(create_overtime).get(list_overtimes))
        .route("/api/overtimes/approve", post(approve_overtime))
        .route("/api/overtimes/reject", post(reject_overtime))
        .route("/api/leaves", post(create_leave).get(list_leaves))
        .route("/api/leaves/approve", post(approve_leave))
        .route("/api/leaves/reject", post(reject_leave))
        .route("/api/leaves/cancel", post(cancel_leave))
        .route("/api/holidays", post(create_holiday).get(list_holidays))
        .layer(cors)
        .with_state(state);

    let addr = std::net::SocketAddr::from(([127, 0, 0, 1], args.port));
    println!("Server listening on http://{}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
