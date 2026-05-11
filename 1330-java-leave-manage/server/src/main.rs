use std::net::SocketAddr;
use std::sync::Arc;

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use chrono::NaiveDate;
use clap::Parser;
use serde::Deserialize;

use leave_core::{LeaveService, LeaveType};

#[derive(Parser, Debug)]
#[command(name = "leave_server")]
#[command(about = "假期管理系统服务端")]
struct Args {
    #[arg(long, env = "LEAVE_SERVER_PORT", default_value_t = 3000)]
    port: u16,

    #[arg(long, env = "LEAVE_SERVER_HOST", default_value = "127.0.0.1")]
    host: String,
}

#[derive(Clone)]
struct AppState {
    service: Arc<LeaveService>,
}

#[derive(Debug, Deserialize)]
struct CreateEmployeeRequest {
    name: String,
    hire_date: String,
}

#[derive(Debug, Deserialize)]
struct ApplyLeaveRequest {
    employee_id: String,
    leave_type: String,
    start_date: String,
    end_date: String,
    sick_leave_proof: Option<String>,
}

#[derive(Debug, Deserialize)]
struct AddBalanceRequest {
    employee_id: String,
    leave_type: String,
    days: u32,
}

#[derive(Debug, Deserialize)]
struct InitializeAnnualLeaveRequest {
    employee_id: String,
    year: i32,
}

#[derive(Debug, Deserialize)]
struct CarryoverRequest {
    employee_id: String,
    year: i32,
}

#[tokio::main]
async fn main() {
    let args = Args::parse();

    let service = Arc::new(LeaveService::new());
    let state = AppState { service };

    let app = Router::new()
        .route("/health", get(health_check))
        .route("/employees", post(create_employee).get(list_employees))
        .route("/employees/:id", get(get_employee))
        .route("/employees/:id/balance", get(get_employee_balance))
        .route("/employees/:id/leaves", get(list_employee_leaves))
        .route("/leaves", post(apply_leave).get(list_all_leaves))
        .route("/leaves/:id", get(get_leave))
        .route("/leaves/:id/cancel", post(cancel_leave))
        .route("/balance/add", post(add_balance))
        .route("/annual-leave/initialize", post(initialize_annual_leave))
        .route("/annual-leave/carryover", post(perform_carryover))
        .with_state(state);

    let addr = format!("{}:{}", args.host, args.port)
        .parse::<SocketAddr>()
        .expect("无效的地址");

    println!("假期管理服务启动于 http://{}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}

async fn health_check() -> impl IntoResponse {
    (StatusCode::OK, Json(serde_json::json!({"status": "ok"})))
}

async fn create_employee(
    State(state): State<AppState>,
    Json(req): Json<CreateEmployeeRequest>,
) -> impl IntoResponse {
    let hire_date = match NaiveDate::parse_from_str(&req.hire_date, "%Y-%m-%d") {
        Ok(d) => d,
        Err(e) => {
            return (
                StatusCode::BAD_REQUEST,
                Json(serde_json::json!({"error": format!("日期格式错误: {}", e)})),
            );
        }
    };

    let employee = state.service.add_employee(&req.name, hire_date);
    match serde_json::to_value(employee) {
        Ok(v) => (StatusCode::CREATED, Json(v)),
        Err(e) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(serde_json::json!({"error": e.to_string()})),
        ),
    }
}

async fn list_employees(State(state): State<AppState>) -> impl IntoResponse {
    let employees = state.service.list_employees();
    match serde_json::to_value(employees) {
        Ok(v) => (StatusCode::OK, Json(v)),
        Err(e) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(serde_json::json!({"error": e.to_string()})),
        ),
    }
}

async fn get_employee(State(state): State<AppState>, Path(id): Path<String>) -> impl IntoResponse {
    match state.service.get_employee(&id) {
        Ok(employee) => match serde_json::to_value(employee) {
            Ok(v) => (StatusCode::OK, Json(v)),
            Err(e) => (
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(serde_json::json!({"error": e.to_string()})),
            ),
        },
        Err(e) => (
            StatusCode::NOT_FOUND,
            Json(serde_json::json!({"error": e.to_string()})),
        ),
    }
}

async fn get_employee_balance(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match state.service.get_leave_balance(&id) {
        Ok(balance) => match serde_json::to_value(balance) {
            Ok(v) => (StatusCode::OK, Json(v)),
            Err(e) => (
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(serde_json::json!({"error": e.to_string()})),
            ),
        },
        Err(e) => (
            StatusCode::NOT_FOUND,
            Json(serde_json::json!({"error": e.to_string()})),
        ),
    }
}

async fn list_employee_leaves(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match state.service.list_employee_leaves(&id) {
        Ok(leaves) => match serde_json::to_value(leaves) {
            Ok(v) => (StatusCode::OK, Json(v)),
            Err(e) => (
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(serde_json::json!({"error": e.to_string()})),
            ),
        },
        Err(e) => (
            StatusCode::NOT_FOUND,
            Json(serde_json::json!({"error": e.to_string()})),
        ),
    }
}

async fn apply_leave(
    State(state): State<AppState>,
    Json(req): Json<ApplyLeaveRequest>,
) -> impl IntoResponse {
    let leave_type = match parse_leave_type(&req.leave_type) {
        Ok(t) => t,
        Err(e) => {
            return (
                StatusCode::BAD_REQUEST,
                Json(serde_json::json!({"error": e.to_string()})),
            );
        }
    };

    let start_date = match NaiveDate::parse_from_str(&req.start_date, "%Y-%m-%d") {
        Ok(d) => d,
        Err(e) => {
            return (
                StatusCode::BAD_REQUEST,
                Json(serde_json::json!({"error": format!("开始日期格式错误: {}", e)})),
            );
        }
    };

    let end_date = match NaiveDate::parse_from_str(&req.end_date, "%Y-%m-%d") {
        Ok(d) => d,
        Err(e) => {
            return (
                StatusCode::BAD_REQUEST,
                Json(serde_json::json!({"error": format!("结束日期格式错误: {}", e)})),
            );
        }
    };

    match state.service.apply_leave(
        &req.employee_id,
        leave_type,
        start_date,
        end_date,
        req.sick_leave_proof,
    ) {
        Ok(leave) => match serde_json::to_value(leave) {
            Ok(v) => (StatusCode::CREATED, Json(v)),
            Err(e) => (
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(serde_json::json!({"error": e.to_string()})),
            ),
        },
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": e.to_string()})),
        ),
    }
}

async fn get_leave(State(state): State<AppState>, Path(id): Path<String>) -> impl IntoResponse {
    match state.service.get_leave_request(&id) {
        Ok(leave) => match serde_json::to_value(leave) {
            Ok(v) => (StatusCode::OK, Json(v)),
            Err(e) => (
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(serde_json::json!({"error": e.to_string()})),
            ),
        },
        Err(e) => (
            StatusCode::NOT_FOUND,
            Json(serde_json::json!({"error": e.to_string()})),
        ),
    }
}

async fn cancel_leave(State(state): State<AppState>, Path(id): Path<String>) -> impl IntoResponse {
    match state.service.cancel_leave(&id) {
        Ok(leave) => match serde_json::to_value(leave) {
            Ok(v) => (StatusCode::OK, Json(v)),
            Err(e) => (
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(serde_json::json!({"error": e.to_string()})),
            ),
        },
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": e.to_string()})),
        ),
    }
}

async fn list_all_leaves(State(state): State<AppState>) -> impl IntoResponse {
    let leaves = state.service.list_all_leaves();
    match serde_json::to_value(leaves) {
        Ok(v) => (StatusCode::OK, Json(v)),
        Err(e) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(serde_json::json!({"error": e.to_string()})),
        ),
    }
}

async fn add_balance(
    State(state): State<AppState>,
    Json(req): Json<AddBalanceRequest>,
) -> impl IntoResponse {
    let leave_type = match parse_leave_type(&req.leave_type) {
        Ok(t) => t,
        Err(e) => {
            return (
                StatusCode::BAD_REQUEST,
                Json(serde_json::json!({"error": e.to_string()})),
            );
        }
    };

    match state.service.add_leave_balance(&req.employee_id, leave_type, req.days) {
        Ok(()) => (StatusCode::OK, Json(serde_json::json!({"status": "ok"}))),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": e.to_string()})),
        ),
    }
}

async fn initialize_annual_leave(
    State(state): State<AppState>,
    Json(req): Json<InitializeAnnualLeaveRequest>,
) -> impl IntoResponse {
    match state.service.initialize_annual_leave(&req.employee_id, req.year) {
        Ok(days) => (StatusCode::OK, Json(serde_json::json!({"entitlement": days}))),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": e.to_string()})),
        ),
    }
}

async fn perform_carryover(
    State(state): State<AppState>,
    Json(req): Json<CarryoverRequest>,
) -> impl IntoResponse {
    match state.service.perform_year_end_carryover(&req.employee_id, req.year) {
        Ok(days) => (StatusCode::OK, Json(serde_json::json!({"carryover_days": days}))),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": e.to_string()})),
        ),
    }
}

fn parse_leave_type(s: &str) -> Result<LeaveType, String> {
    match s.to_lowercase().as_str() {
        "annual" | "年假" => Ok(LeaveType::Annual),
        "personal" | "事假" => Ok(LeaveType::Personal),
        "sick" | "病假" => Ok(LeaveType::Sick),
        "compensatory" | "调休" => Ok(LeaveType::Compensatory),
        _ => Err(format!("未知的假期类型: {}", s)),
    }
}
