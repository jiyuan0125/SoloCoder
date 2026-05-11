use std::sync::Arc;

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    Json, Router,
};
use clap::Parser;
use serde::{Deserialize, Serialize};

use asset_core::{
    AssetError, AssetService, Asset, AssetCategory, AssetDetail, BorrowApplication,
    Employee, InMemoryRepository, ReturnCondition, ReturnRecord, ResponsibilityHistory,
    TransferRequest, UserRole,
};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "SERVER_PORT", default_value_t = 8100)]
    port: u16,

    #[arg(short, long, env = "SERVER_HOST", default_value = "127.0.0.1")]
    host: String,
}

type SharedService = Arc<AssetService>;

#[derive(Serialize, Deserialize)]
struct CreateEmployeeRequest {
    name: String,
    role: String,
}

#[derive(Serialize, Deserialize)]
struct CreateCategoryRequest {
    name: String,
    useful_life_months: u32,
    max_per_employee: u32,
}

#[derive(Serialize, Deserialize)]
struct CreateAssetRequest {
    name: String,
    category_id: String,
    purchase_price: f64,
    purchase_date: String,
}

#[derive(Serialize, Deserialize)]
struct ApplyBorrowRequest {
    asset_id: String,
    applicant_id: String,
    reason: String,
}

#[derive(Serialize, Deserialize)]
struct ProcessApplicationRequest {
    application_id: String,
    approver_id: String,
}

#[derive(Serialize, Deserialize)]
struct ReturnAssetRequest {
    asset_id: String,
    returned_by: String,
    verified_by: String,
    condition: String,
    damage_note: Option<String>,
}

#[derive(Serialize, Deserialize)]
struct InitiateTransferRequest {
    asset_id: String,
    from_employee_id: String,
    to_employee_id: String,
}

#[derive(Serialize, Deserialize)]
struct ConfirmTransferRequest {
    transfer_id: String,
    employee_id: String,
}

#[derive(Serialize)]
struct ErrorResponse {
    error: String,
}

fn parse_date(s: &str) -> Result<chrono::NaiveDate, AppError> {
    chrono::NaiveDate::parse_from_str(s, "%Y-%m-%d")
        .map_err(|_| AppError::BadRequest("日期格式错误，请使用 YYYY-MM-DD".to_string()))
}

fn parse_role(s: &str) -> Result<UserRole, AppError> {
    match s.to_lowercase().as_str() {
        "admin" | "administrator" => Ok(UserRole::Admin),
        "employee" => Ok(UserRole::Employee),
        _ => Err(AppError::BadRequest(
            "角色只能是 admin 或 employee".to_string(),
        )),
    }
}

fn parse_condition(s: &str) -> Result<ReturnCondition, AppError> {
    match s.to_lowercase().as_str() {
        "good" | "完好" => Ok(ReturnCondition::Good),
        "damaged" | "损坏" => Ok(ReturnCondition::Damaged),
        _ => Err(AppError::BadRequest(
            "状态只能是 good/damaged 或 完好/损坏".to_string(),
        )),
    }
}

#[derive(Debug)]
enum AppError {
    Asset(AssetError),
    BadRequest(String),
}

impl From<AssetError> for AppError {
    fn from(err: AssetError) -> Self {
        AppError::Asset(err)
    }
}

impl IntoResponse for AppError {
    fn into_response(self) -> axum::response::Response {
        match self {
            AppError::Asset(err) => (
                StatusCode::BAD_REQUEST,
                Json(ErrorResponse {
                    error: err.to_string(),
                }),
            )
                .into_response(),
            AppError::BadRequest(msg) => (
                StatusCode::BAD_REQUEST,
                Json(ErrorResponse { error: msg }),
            )
                .into_response(),
        }
    }
}

async fn create_employee(
    State(service): State<SharedService>,
    Json(req): Json<CreateEmployeeRequest>,
) -> Result<Json<Employee>, AppError> {
    let role = parse_role(&req.role)?;
    let employee = service.create_employee(req.name, role);
    Ok(Json(employee))
}

async fn list_employees(State(service): State<SharedService>) -> Json<Vec<Employee>> {
    Json(service.list_employees())
}

async fn get_employee(
    State(service): State<SharedService>,
    Path(id): Path<String>,
) -> Result<Json<Employee>, AppError> {
    Ok(Json(service.get_employee(&id)?))
}

async fn create_category(
    State(service): State<SharedService>,
    Json(req): Json<CreateCategoryRequest>,
) -> Json<AssetCategory> {
    let category = service.create_category(req.name, req.useful_life_months, req.max_per_employee);
    Json(category)
}

async fn list_categories(State(service): State<SharedService>) -> Json<Vec<AssetCategory>> {
    Json(service.list_categories())
}

async fn get_category(
    State(service): State<SharedService>,
    Path(id): Path<String>,
) -> Result<Json<AssetCategory>, AppError> {
    Ok(Json(service.get_category(&id)?))
}

async fn create_asset(
    State(service): State<SharedService>,
    Json(req): Json<CreateAssetRequest>,
) -> Result<Json<Asset>, AppError> {
    let purchase_date = parse_date(&req.purchase_date)?;
    let asset = service.create_asset(req.name, req.category_id, req.purchase_price, purchase_date)?;
    Ok(Json(asset))
}

async fn list_assets(State(service): State<SharedService>) -> Json<Vec<Asset>> {
    Json(service.list_assets())
}

async fn list_asset_details(State(service): State<SharedService>) -> Json<Vec<AssetDetail>> {
    Json(service.list_asset_details())
}

async fn get_asset(
    State(service): State<SharedService>,
    Path(id): Path<String>,
) -> Result<Json<Asset>, AppError> {
    Ok(Json(service.get_asset(&id)?))
}

async fn get_asset_detail(
    State(service): State<SharedService>,
    Path(id): Path<String>,
) -> Result<Json<AssetDetail>, AppError> {
    Ok(Json(service.get_asset_detail(&id)?))
}

async fn apply_for_borrow(
    State(service): State<SharedService>,
    Json(req): Json<ApplyBorrowRequest>,
) -> Result<Json<BorrowApplication>, AppError> {
    let application = service.apply_for_borrow(req.asset_id, req.applicant_id, req.reason)?;
    Ok(Json(application))
}

async fn approve_application(
    State(service): State<SharedService>,
    Json(req): Json<ProcessApplicationRequest>,
) -> Result<Json<BorrowApplication>, AppError> {
    let application = service.approve_application(req.application_id, req.approver_id)?;
    Ok(Json(application))
}

async fn reject_application(
    State(service): State<SharedService>,
    Json(req): Json<ProcessApplicationRequest>,
) -> Result<Json<BorrowApplication>, AppError> {
    let application = service.reject_application(req.application_id, req.approver_id)?;
    Ok(Json(application))
}

async fn list_applications(State(service): State<SharedService>) -> Json<Vec<BorrowApplication>> {
    Json(service.list_applications())
}

async fn get_application(
    State(service): State<SharedService>,
    Path(id): Path<String>,
) -> Result<Json<BorrowApplication>, AppError> {
    Ok(Json(service.get_application(&id)?))
}

async fn return_asset(
    State(service): State<SharedService>,
    Json(req): Json<ReturnAssetRequest>,
) -> Result<Json<Asset>, AppError> {
    let condition = parse_condition(&req.condition)?;
    let asset = service.return_asset(
        req.asset_id,
        req.returned_by,
        req.verified_by,
        condition,
        req.damage_note,
    )?;
    Ok(Json(asset))
}

async fn list_return_records(
    State(service): State<SharedService>,
    Path(asset_id): Path<String>,
) -> Json<Vec<ReturnRecord>> {
    Json(service.list_return_records(&asset_id))
}

async fn initiate_transfer(
    State(service): State<SharedService>,
    Json(req): Json<InitiateTransferRequest>,
) -> Result<Json<TransferRequest>, AppError> {
    let transfer = service.initiate_transfer(req.asset_id, req.from_employee_id, req.to_employee_id)?;
    Ok(Json(transfer))
}

async fn confirm_transfer_from(
    State(service): State<SharedService>,
    Json(req): Json<ConfirmTransferRequest>,
) -> Result<Json<TransferRequest>, AppError> {
    let transfer = service.confirm_transfer_by_from(req.transfer_id, req.employee_id)?;
    Ok(Json(transfer))
}

async fn confirm_transfer_to(
    State(service): State<SharedService>,
    Json(req): Json<ConfirmTransferRequest>,
) -> Result<Json<TransferRequest>, AppError> {
    let transfer = service.confirm_transfer_by_to(req.transfer_id, req.employee_id)?;
    Ok(Json(transfer))
}

async fn list_transfers(State(service): State<SharedService>) -> Json<Vec<TransferRequest>> {
    Json(service.list_transfers())
}

async fn get_transfer(
    State(service): State<SharedService>,
    Path(id): Path<String>,
) -> Result<Json<TransferRequest>, AppError> {
    Ok(Json(service.get_transfer(&id)?))
}

async fn list_history(
    State(service): State<SharedService>,
    Path(asset_id): Path<String>,
) -> Json<Vec<ResponsibilityHistory>> {
    Json(service.list_history_by_asset(&asset_id))
}

fn app(service: SharedService) -> Router {
    Router::new()
        .route("/employees", axum::routing::get(list_employees).post(create_employee))
        .route("/employees/:id", axum::routing::get(get_employee))
        .route("/categories", axum::routing::get(list_categories).post(create_category))
        .route("/categories/:id", axum::routing::get(get_category))
        .route("/assets", axum::routing::get(list_assets).post(create_asset))
        .route("/assets/details", axum::routing::get(list_asset_details))
        .route("/assets/:id", axum::routing::get(get_asset))
        .route("/assets/:id/detail", axum::routing::get(get_asset_detail))
        .route("/applications", axum::routing::get(list_applications).post(apply_for_borrow))
        .route("/applications/:id", axum::routing::get(get_application))
        .route("/applications/approve", axum::routing::post(approve_application))
        .route("/applications/reject", axum::routing::post(reject_application))
        .route("/returns", axum::routing::post(return_asset))
        .route("/returns/:asset_id", axum::routing::get(list_return_records))
        .route("/transfers", axum::routing::get(list_transfers).post(initiate_transfer))
        .route("/transfers/:id", axum::routing::get(get_transfer))
        .route("/transfers/confirm-from", axum::routing::post(confirm_transfer_from))
        .route("/transfers/confirm-to", axum::routing::post(confirm_transfer_to))
        .route("/history/:asset_id", axum::routing::get(list_history))
        .with_state(service)
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();

    let args = Args::parse();
    let repo = InMemoryRepository::new();
    let service = Arc::new(AssetService::new(repo));

    let app = app(service);
    let addr = format!("{}:{}", args.host, args.port).parse().unwrap();
    println!("Server starting on {}", addr);

    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}
