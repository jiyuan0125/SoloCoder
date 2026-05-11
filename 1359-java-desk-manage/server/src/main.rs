use std::net::SocketAddr;
use std::sync::Arc;

use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    response::{IntoResponse, Json},
    routing::{get, post, put},
    Router,
};
use clap::Parser;
use serde::{Deserialize, Serialize};
use tokio::net::TcpListener;
use uuid::Uuid;
use chrono::NaiveDate;
use desk_core::{DeskService, DeskError, DeskType};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long)]
    port: Option<u16>,
}

struct AppState {
    service: Arc<DeskService>,
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateDepartmentRequest {
    name: String,
    floor: i32,
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateDeskRequest {
    code: String,
    desk_type: DeskType,
    floor: i32,
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateEmployeeRequest {
    name: String,
    department_id: Uuid,
}

#[derive(Debug, Serialize, Deserialize)]
struct BatchCreateEmployeesRequest {
    employees: Vec<(String, Uuid)>,
}

#[derive(Debug, Serialize, Deserialize)]
struct ReserveDeskRequest {
    employee_id: Uuid,
    date: NaiveDate,
}

#[derive(Debug, Serialize, Deserialize)]
struct BatchRelocateRequest {
    department_id: Uuid,
    target_floor: i32,
}

#[derive(Debug, Serialize)]
struct ApiResponse<T> {
    success: bool,
    data: Option<T>,
    error: Option<String>,
}

impl<T> ApiResponse<T> {
    fn success(data: T) -> Self {
        Self {
            success: true,
            data: Some(data),
            error: None,
        }
    }
    
    fn error(message: String) -> Self {
        Self {
            success: false,
            data: None,
            error: Some(message),
        }
    }
}

fn error_to_response(err: DeskError) -> (StatusCode, Json<ApiResponse<()>>) {
    let status = match &err {
        DeskError::DeskNotFound(_) | DeskError::EmployeeNotFound(_) | 
        DeskError::DepartmentNotFound(_) | DeskError::ReservationNotFound(_) => {
            StatusCode::NOT_FOUND
        }
        DeskError::NoAvailableFixedDesk | DeskError::DeskAlreadyOccupied |
        DeskError::EmployeeAlreadyHasReservation | DeskError::ReservationTimeOutOfRange |
        DeskError::ReservationExpired | DeskError::ReservationAlreadyCheckedIn |
        DeskError::BatchConflict(_) | DeskError::ConcurrentAllocationConflict => {
            StatusCode::CONFLICT
        }
        DeskError::InvalidOperation(_) => StatusCode::BAD_REQUEST,
    };
    
    (status, Json(ApiResponse::error(err.to_string())))
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();
    
    let args = Args::parse();
    
    let port = args.port
        .or_else(|| std::env::var("PORT").ok().and_then(|p| p.parse().ok()))
        .unwrap_or(8080);
    
    let service = Arc::new(DeskService::new());
    let state = Arc::new(AppState { service });
    
    let app = Router::new()
        .route("/api/health", get(health_check))
        
        .route("/api/departments", post(create_department))
        .route("/api/departments", get(list_departments))
        .route("/api/departments/:id", get(get_department))
        
        .route("/api/desks", post(create_desk))
        .route("/api/desks", get(list_desks))
        .route("/api/desks/:id", get(get_desk))
        
        .route("/api/employees", post(create_employee))
        .route("/api/employees/batch", post(batch_create_employees))
        .route("/api/employees", get(list_employees))
        .route("/api/employees/:id", get(get_employee))
        .route("/api/employees/:id/leave", put(employee_leave))
        .route("/api/employees/:id/desk", get(get_employee_desk))
        
        .route("/api/reservations", post(reserve_desk))
        .route("/api/reservations", get(list_reservations))
        .route("/api/reservations/:id", get(get_reservation))
        .route("/api/reservations/:id/checkin", put(check_in_reservation))
        .route("/api/reservations/:id/cancel", put(cancel_reservation))
        
        .route("/api/desks/shared/available", get(list_available_shared_desks))
        .route("/api/departments/relocate", post(batch_relocate_department))
        .route("/api/maintenance/expired", post(process_expired_reservations))
        .route("/api/maintenance/pending", post(process_pending_desks))
        
        .with_state(state);
    
    let addr = SocketAddr::from(([0, 0, 0, 0], port));
    let listener = TcpListener::bind(addr).await.unwrap();
    tracing::info!("Server listening on {}", addr);
    
    axum::serve(listener, app).await.unwrap();
}

async fn health_check() -> Json<ApiResponse<String>> {
    Json(ApiResponse::success("OK".to_string()))
}

async fn create_department(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateDepartmentRequest>,
) -> impl IntoResponse {
    let dept = state.service.create_department(req.name, req.floor);
    (StatusCode::CREATED, Json(ApiResponse::success(dept)))
}

async fn list_departments(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let depts = state.service.list_departments();
    Json(ApiResponse::success(depts))
}

async fn get_department(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.get_department(id) {
        Ok(dept) => Json(ApiResponse::success(dept)).into_response(),
        Err(err) => {
            let (status, resp) = error_to_response(err);
            (status, resp).into_response()
        }
    }
}

async fn create_desk(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateDeskRequest>,
) -> impl IntoResponse {
    let desk = state.service.create_desk(req.code, req.desk_type, req.floor);
    (StatusCode::CREATED, Json(ApiResponse::success(desk)))
}

async fn list_desks(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let desks = state.service.list_desks();
    Json(ApiResponse::success(desks))
}

async fn get_desk(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.get_desk(id) {
        Ok(desk) => Json(ApiResponse::success(desk)).into_response(),
        Err(err) => {
            let (status, resp) = error_to_response(err);
            (status, resp).into_response()
        }
    }
}

async fn create_employee(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateEmployeeRequest>,
) -> impl IntoResponse {
    match state.service.create_employee(req.name, req.department_id) {
        Ok((emp, desk)) => {
            let result = serde_json::json!({
                "employee": emp,
                "desk": desk
            });
            (StatusCode::CREATED, Json(ApiResponse::success(result))).into_response()
        }
        Err(err) => {
            let (status, resp) = error_to_response(err);
            (status, resp).into_response()
        }
    }
}

async fn batch_create_employees(
    State(state): State<Arc<AppState>>,
    Json(req): Json<BatchCreateEmployeesRequest>,
) -> impl IntoResponse {
    match state.service.batch_create_employees(req.employees) {
        Ok(results) => (StatusCode::CREATED, Json(ApiResponse::success(results))).into_response(),
        Err(err) => {
            let (status, resp) = error_to_response(err);
            (status, resp).into_response()
        }
    }
}

async fn list_employees(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let employees = state.service.list_employees();
    Json(ApiResponse::success(employees))
}

async fn get_employee(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.get_employee(id) {
        Ok(emp) => Json(ApiResponse::success(emp)).into_response(),
        Err(err) => {
            let (status, resp) = error_to_response(err);
            (status, resp).into_response()
        }
    }
}

async fn employee_leave(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.employee_leave(id) {
        Ok(()) => Json(ApiResponse::success(())).into_response(),
        Err(err) => {
            let (status, resp) = error_to_response(err);
            (status, resp).into_response()
        }
    }
}

async fn get_employee_desk(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    let desk = state.service.get_employee_desk(id);
    Json(ApiResponse::success(desk))
}

async fn reserve_desk(
    State(state): State<Arc<AppState>>,
    Path(desk_id): Path<Uuid>,
    Json(req): Json<ReserveDeskRequest>,
) -> impl IntoResponse {
    match state.service.reserve_shared_desk(req.employee_id, desk_id, req.date) {
        Ok(reservation) => (StatusCode::CREATED, Json(ApiResponse::success(reservation))).into_response(),
        Err(err) => {
            let (status, resp) = error_to_response(err);
            (status, resp).into_response()
        }
    }
}

async fn list_reservations(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let reservations = state.service.list_reservations();
    Json(ApiResponse::success(reservations))
}

async fn get_reservation(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.get_reservation(id) {
        Ok(reservation) => Json(ApiResponse::success(reservation)).into_response(),
        Err(err) => {
            let (status, resp) = error_to_response(err);
            (status, resp).into_response()
        }
    }
}

async fn check_in_reservation(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.check_in_reservation(id) {
        Ok(()) => Json(ApiResponse::success(())).into_response(),
        Err(err) => {
            let (status, resp) = error_to_response(err);
            (status, resp).into_response()
        }
    }
}

async fn cancel_reservation(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.cancel_reservation(id) {
        Ok(()) => Json(ApiResponse::success(())).into_response(),
        Err(err) => {
            let (status, resp) = error_to_response(err);
            (status, resp).into_response()
        }
    }
}

#[derive(Debug, Deserialize)]
struct DateQuery {
    date: Option<NaiveDate>,
}

async fn list_available_shared_desks(
    State(state): State<Arc<AppState>>,
    Query(query): Query<DateQuery>,
) -> impl IntoResponse {
    let date = query.date.unwrap_or_else(desk_core::models::now_naive);
    let desks = state.service.list_available_shared_desks(date);
    Json(ApiResponse::success(desks))
}

async fn batch_relocate_department(
    State(state): State<Arc<AppState>>,
    Json(req): Json<BatchRelocateRequest>,
) -> impl IntoResponse {
    match state.service.batch_relocate_department(req.department_id, req.target_floor) {
        Ok(mapping) => Json(ApiResponse::success(mapping)).into_response(),
        Err(err) => {
            let (status, resp) = error_to_response(err);
            (status, resp).into_response()
        }
    }
}

async fn process_expired_reservations(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let expired = state.service.process_expired_reservations();
    Json(ApiResponse::success(expired))
}

async fn process_pending_desks(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let (released, notifications) = state.service.process_pending_release_desks();
    let result = serde_json::json!({
        "released_desks": released,
        "notifications": notifications
    });
    Json(ApiResponse::success(result))
}
