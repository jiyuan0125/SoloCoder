use std::sync::Arc;
use std::time::Duration;

use axum::{
    extract::State,
    http::StatusCode,
    response::{IntoResponse, Json, Response},
    routing::{get, post, put},
    Router,
};
use clap::Parser;
use chrono::Utc;
use serde::{Deserialize, Serialize};
use tracing::{info};
use uuid::Uuid;

use seal_borrow_core::{
    SealBorrowService, InMemoryStore,
    CreateSealRequest, CreateBorrowRequest, ApproveRequest, RejectRequest,
    RenewRequest, ReturnRequest, UpdateSealStatusRequest, CreateEmployeeRequest,
    SealBorrowError,
};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Config {
    #[arg(short, long, env = "SERVER_PORT", default_value_t = 8080)]
    port: u16,
    
    #[arg(short, long, env = "REMINDER_INTERVAL_DAYS", default_value_t = 3)]
    reminder_interval_days: i64,
}

struct AppState {
    service: SealBorrowService,
}

#[derive(Debug, Serialize, Deserialize)]
struct ErrorResponse {
    error: String,
}

fn error_to_response(err: SealBorrowError) -> Response {
    let status = match &err {
        SealBorrowError::SealNotFound => StatusCode::NOT_FOUND,
        SealBorrowError::EmployeeNotFound => StatusCode::NOT_FOUND,
        SealBorrowError::BorrowRequestNotFound => StatusCode::NOT_FOUND,
        SealBorrowError::SealInMaintenance => StatusCode::BAD_REQUEST,
        SealBorrowError::SealAlreadyBorrowed => StatusCode::BAD_REQUEST,
        SealBorrowError::AlreadyBorrowingThisSeal => StatusCode::BAD_REQUEST,
        SealBorrowError::ApproverCannotBeBorrower => StatusCode::BAD_REQUEST,
        SealBorrowError::OnlyCustodianCanApprove => StatusCode::FORBIDDEN,
        SealBorrowError::InvalidStatusForOperation => StatusCode::BAD_REQUEST,
        SealBorrowError::NewReturnDateMustBeAfterOriginal => StatusCode::BAD_REQUEST,
        SealBorrowError::OnlyBorrowerCanRenew => StatusCode::FORBIDDEN,
        SealBorrowError::InternalError(_) => StatusCode::INTERNAL_SERVER_ERROR,
    };
    
    (status, Json(ErrorResponse { error: err.to_string() })).into_response()
}

type HandlerResult<T> = std::result::Result<Json<T>, Response>;

fn map_handler_result<T>(result: seal_borrow_core::Result<T>) -> HandlerResult<T> {
    result.map(Json).map_err(error_to_response)
}

#[derive(Debug, Serialize, Deserialize)]
struct ProcessRenewalRequest {
    renewal_request_id: Uuid,
    approved: bool,
    approver_id: Uuid,
    reject_reason: Option<String>,
}

async fn create_employee(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateEmployeeRequest>,
) -> HandlerResult<seal_borrow_core::Employee> {
    let employee = state.service.create_employee(req).await;
    Ok(Json(employee))
}

async fn list_employees(
    State(state): State<Arc<AppState>>,
) -> HandlerResult<Vec<seal_borrow_core::Employee>> {
    let employees = state.service.list_employees().await;
    Ok(Json(employees))
}

async fn get_employee(
    State(state): State<Arc<AppState>>,
    axum::extract::Path(id): axum::extract::Path<Uuid>,
) -> HandlerResult<seal_borrow_core::Employee> {
    map_handler_result(state.service.get_employee(id).await)
}

async fn create_seal(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateSealRequest>,
) -> HandlerResult<seal_borrow_core::Seal> {
    map_handler_result(state.service.create_seal(req).await)
}

async fn list_seals(
    State(state): State<Arc<AppState>>,
) -> HandlerResult<Vec<seal_borrow_core::Seal>> {
    let seals = state.service.list_seals().await;
    Ok(Json(seals))
}

async fn get_seal(
    State(state): State<Arc<AppState>>,
    axum::extract::Path(id): axum::extract::Path<Uuid>,
) -> HandlerResult<seal_borrow_core::Seal> {
    map_handler_result(state.service.get_seal(id).await)
}

async fn update_seal_status(
    State(state): State<Arc<AppState>>,
    Json(req): Json<UpdateSealStatusRequest>,
) -> HandlerResult<seal_borrow_core::Seal> {
    map_handler_result(state.service.update_seal_status(req).await)
}

async fn create_borrow_request(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateBorrowRequest>,
) -> HandlerResult<seal_borrow_core::BorrowRequest> {
    map_handler_result(state.service.create_borrow_request(req).await)
}

async fn list_borrow_requests(
    State(state): State<Arc<AppState>>,
) -> HandlerResult<Vec<seal_borrow_core::BorrowRequest>> {
    let requests = state.service.list_borrow_requests().await;
    Ok(Json(requests))
}

async fn get_borrow_request(
    State(state): State<Arc<AppState>>,
    axum::extract::Path(id): axum::extract::Path<Uuid>,
) -> HandlerResult<seal_borrow_core::BorrowRequest> {
    map_handler_result(state.service.get_borrow_request(id).await)
}

async fn list_borrow_requests_by_seal(
    State(state): State<Arc<AppState>>,
    axum::extract::Path(seal_id): axum::extract::Path<Uuid>,
) -> HandlerResult<Vec<seal_borrow_core::BorrowRequest>> {
    let requests = state.service.list_borrow_requests_by_seal(seal_id).await;
    Ok(Json(requests))
}

async fn list_borrow_requests_by_borrower(
    State(state): State<Arc<AppState>>,
    axum::extract::Path(borrower_id): axum::extract::Path<Uuid>,
) -> HandlerResult<Vec<seal_borrow_core::BorrowRequest>> {
    let requests = state.service.list_borrow_requests_by_borrower(borrower_id).await;
    Ok(Json(requests))
}

async fn approve_request(
    State(state): State<Arc<AppState>>,
    Json(req): Json<ApproveRequest>,
) -> HandlerResult<seal_borrow_core::BorrowRequest> {
    map_handler_result(state.service.approve_request(req).await)
}

async fn reject_request(
    State(state): State<Arc<AppState>>,
    Json(req): Json<RejectRequest>,
) -> HandlerResult<seal_borrow_core::BorrowRequest> {
    map_handler_result(state.service.reject_request(req).await)
}

async fn renew_request(
    State(state): State<Arc<AppState>>,
    Json(req): Json<RenewRequest>,
) -> HandlerResult<seal_borrow_core::BorrowRequest> {
    map_handler_result(state.service.renew_request(req).await)
}

async fn process_renewal_approval(
    State(state): State<Arc<AppState>>,
    Json(req): Json<ProcessRenewalRequest>,
) -> HandlerResult<seal_borrow_core::BorrowRequest> {
    map_handler_result(
        state.service.process_renewal_approval(
            req.renewal_request_id,
            req.approved,
            req.approver_id,
            req.reject_reason,
        ).await
    )
}

async fn return_seal(
    State(state): State<Arc<AppState>>,
    Json(req): Json<ReturnRequest>,
) -> HandlerResult<seal_borrow_core::BorrowRequest> {
    map_handler_result(state.service.return_seal(req).await)
}

async fn list_reminders(
    State(state): State<Arc<AppState>>,
) -> HandlerResult<Vec<seal_borrow_core::ReminderRecord>> {
    let reminders = state.service.list_reminders().await;
    Ok(Json(reminders))
}

async fn list_reminders_by_request(
    State(state): State<Arc<AppState>>,
    axum::extract::Path(request_id): axum::extract::Path<Uuid>,
) -> HandlerResult<Vec<seal_borrow_core::ReminderRecord>> {
    let reminders = state.service.list_reminders_by_request(request_id).await;
    Ok(Json(reminders))
}

async fn health_check() -> &'static str {
    "OK"
}

fn create_router(state: Arc<AppState>) -> Router {
    Router::new()
        .route("/health", get(health_check))
        .route("/employees", post(create_employee).get(list_employees))
        .route("/employees/:id", get(get_employee))
        .route("/seals", post(create_seal).get(list_seals))
        .route("/seals/:id", get(get_seal))
        .route("/seals/status", put(update_seal_status))
        .route("/borrow-requests", post(create_borrow_request).get(list_borrow_requests))
        .route("/borrow-requests/:id", get(get_borrow_request))
        .route("/borrow-requests/seal/:seal_id", get(list_borrow_requests_by_seal))
        .route("/borrow-requests/borrower/:borrower_id", get(list_borrow_requests_by_borrower))
        .route("/borrow-requests/approve", post(approve_request))
        .route("/borrow-requests/reject", post(reject_request))
        .route("/borrow-requests/renew", post(renew_request))
        .route("/borrow-requests/process-renewal", post(process_renewal_approval))
        .route("/borrow-requests/return", post(return_seal))
        .route("/reminders", get(list_reminders))
        .route("/reminders/request/:request_id", get(list_reminders_by_request))
        .with_state(state)
}

async fn start_reminder_checker(service: SealBorrowService, interval_hours: u64) {
    let mut interval = tokio::time::interval(Duration::from_secs(interval_hours * 3600));
    loop {
        interval.tick().await;
        info!("检查逾期借用并发送催还通知...");
        let now = Utc::now();
        let reminders = service.check_and_create_reminders(now).await;
        if !reminders.is_empty() {
            info!("已生成 {} 条催还通知", reminders.len());
        }
    }
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();
    
    let config = Config::parse();
    
    let store = Arc::new(InMemoryStore::new());
    let service = SealBorrowService::new(store.clone())
        .with_reminder_interval(config.reminder_interval_days);
    
    let state = Arc::new(AppState {
        service: service.clone(),
    });
    
    let app = create_router(state);
    
    let reminder_service = service.clone();
    tokio::spawn(async move {
        start_reminder_checker(reminder_service, 1).await;
    });
    
    let addr = std::net::SocketAddr::from(([0, 0, 0, 0], config.port));
    info!("印章借用管理系统服务启动在 {}", addr);
    
    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
