use std::sync::Arc;
use clap::Parser;
use tokio::sync::Mutex;
use axum::{
    routing::{get, post, delete},
    Router,
    Json,
    extract::{State, Path},
    http::StatusCode,
    response::IntoResponse,
};
use tower_http::cors::{CorsLayer, Any};
use checkup_core::*;
use chrono::NaiveDate;
use uuid::Uuid;
use serde::{Deserialize, Serialize};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "CHECKUP_PORT", default_value_t = 3000)]
    port: u16,
}

type AppState = Arc<Mutex<CheckupService>>;

#[derive(Debug, Serialize, Deserialize)]
struct CreateItemRequest {
    name: String,
    price: f64,
    gender_restriction: GenderRestriction,
    has_contrast_agent: bool,
    is_gastroscopy: bool,
    is_colonoscopy: bool,
    is_xray: bool,
    is_pregnant_forbidden: bool,
}

#[derive(Debug, Serialize, Deserialize)]
struct CreatePackageRequest {
    name: String,
    items: Vec<Uuid>,
    package_price: f64,
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateCustomerRequest {
    name: String,
    gender: Gender,
    birth_date: NaiveDate,
    is_pregnant: bool,
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateMutexRuleRequest {
    rule_type: MutexRuleType,
    item1: Uuid,
    item2: Uuid,
    description: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateReservationRequest {
    customer_id: Uuid,
    appointment_date: NaiveDate,
    base_package: Option<Uuid>,
    additional_items: Vec<Uuid>,
    removed_items: Vec<Uuid>,
}

struct AppError(CheckupError);

impl IntoResponse for AppError {
    fn into_response(self) -> axum::response::Response {
        let (status, error_message) = match self.0 {
            CheckupError::ItemNotFound(msg) => (StatusCode::NOT_FOUND, msg),
            CheckupError::PackageNotFound(msg) => (StatusCode::NOT_FOUND, msg),
            CheckupError::CustomerNotFound(msg) => (StatusCode::NOT_FOUND, msg),
            CheckupError::HardMutexConflict(msg) => (StatusCode::BAD_REQUEST, msg),
            CheckupError::GenderRestrictionViolation(msg) => (StatusCode::BAD_REQUEST, msg),
            CheckupError::AgeRestrictionViolation(msg) => (StatusCode::BAD_REQUEST, msg),
            CheckupError::PregnantForbidden(msg) => (StatusCode::BAD_REQUEST, msg),
        };
        
        (status, Json(serde_json::json!({ "error": error_message }))).into_response()
    }
}

impl From<CheckupError> for AppError {
    fn from(err: CheckupError) -> Self {
        AppError(err)
    }
}

async fn create_item(
    State(state): State<AppState>,
    Json(req): Json<CreateItemRequest>,
) -> Json<CheckupItem> {
    let mut service = state.lock().await;
    let item = CheckupItem::new(
        req.name,
        req.price,
        req.gender_restriction,
        req.has_contrast_agent,
        req.is_gastroscopy,
        req.is_colonoscopy,
        req.is_xray,
        req.is_pregnant_forbidden,
    );
    let item_clone = item.clone();
    service.add_item(item);
    Json(item_clone)
}

async fn list_items(State(state): State<AppState>) -> Json<Vec<CheckupItem>> {
    let service = state.lock().await;
    Json(service.list_items().into_iter().cloned().collect())
}

async fn get_item(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<CheckupItem>, AppError> {
    let service = state.lock().await;
    match service.get_item(&id) {
        Some(item) => Ok(Json(item.clone())),
        None => Err(CheckupError::ItemNotFound(format!("{}", id)).into()),
    }
}

async fn create_package(
    State(state): State<AppState>,
    Json(req): Json<CreatePackageRequest>,
) -> Json<CheckupPackage> {
    let mut service = state.lock().await;
    let package = CheckupPackage::new(req.name, req.items, req.package_price);
    let pkg_clone = package.clone();
    service.add_package(package);
    Json(pkg_clone)
}

async fn list_packages(State(state): State<AppState>) -> Json<Vec<CheckupPackage>> {
    let service = state.lock().await;
    Json(service.list_packages().into_iter().cloned().collect())
}

async fn get_package(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<CheckupPackage>, AppError> {
    let service = state.lock().await;
    match service.get_package(&id) {
        Some(pkg) => Ok(Json(pkg.clone())),
        None => Err(CheckupError::PackageNotFound(format!("{}", id)).into()),
    }
}

async fn create_customer(
    State(state): State<AppState>,
    Json(req): Json<CreateCustomerRequest>,
) -> Json<Customer> {
    let mut service = state.lock().await;
    let customer = Customer::new(req.name, req.gender, req.birth_date, req.is_pregnant);
    let cust_clone = customer.clone();
    service.add_customer(customer);
    Json(cust_clone)
}

async fn list_customers(State(state): State<AppState>) -> Json<Vec<Customer>> {
    let service = state.lock().await;
    Json(service.list_customers().into_iter().cloned().collect())
}

async fn get_customer(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<Customer>, AppError> {
    let service = state.lock().await;
    match service.get_customer(&id) {
        Some(cust) => Ok(Json(cust.clone())),
        None => Err(CheckupError::CustomerNotFound(format!("{}", id)).into()),
    }
}

async fn create_mutex_rule(
    State(state): State<AppState>,
    Json(req): Json<CreateMutexRuleRequest>,
) -> Json<MutexRule> {
    let mut service = state.lock().await;
    let rule = MutexRule::new(req.rule_type, req.item1, req.item2, req.description);
    let rule_clone = rule.clone();
    service.add_mutex_rule(rule);
    Json(rule_clone)
}

async fn list_mutex_rules(State(state): State<AppState>) -> Json<Vec<MutexRule>> {
    let service = state.lock().await;
    Json(service.list_mutex_rules().clone())
}

async fn delete_mutex_rule(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> StatusCode {
    let mut service = state.lock().await;
    service.remove_mutex_rule(&id);
    StatusCode::NO_CONTENT
}

async fn create_reservation(
    State(state): State<AppState>,
    Json(req): Json<CreateReservationRequest>,
) -> Result<Json<Reservation>, AppError> {
    let mut service = state.lock().await;
    let reservation = service.create_reservation(
        req.customer_id,
        req.appointment_date,
        req.base_package,
        req.additional_items,
        req.removed_items,
    )?;
    Ok(Json(reservation))
}

async fn list_reservations(State(state): State<AppState>) -> Json<Vec<Reservation>> {
    let service = state.lock().await;
    Json(service.list_reservations().into_iter().cloned().collect())
}

async fn get_reservation(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<Reservation>, AppError> {
    let service = state.lock().await;
    match service.get_reservation(&id) {
        Some(res) => Ok(Json(res.clone())),
        None => Err(CheckupError::CustomerNotFound(format!("{}", id)).into()),
    }
}

#[tokio::main]
async fn main() {
    let args = Args::parse();
    
    let service = Arc::new(Mutex::new(CheckupService::new()));
    
    let app = Router::new()
        .route("/items", post(create_item).get(list_items))
        .route("/items/:id", get(get_item))
        .route("/packages", post(create_package).get(list_packages))
        .route("/packages/:id", get(get_package))
        .route("/customers", post(create_customer).get(list_customers))
        .route("/customers/:id", get(get_customer))
        .route("/mutex-rules", post(create_mutex_rule).get(list_mutex_rules))
        .route("/mutex-rules/:id", delete(delete_mutex_rule))
        .route("/reservations", post(create_reservation).get(list_reservations))
        .route("/reservations/:id", get(get_reservation))
        .layer(CorsLayer::new().allow_origin(Any).allow_methods(Any).allow_headers(Any))
        .with_state(service);
    
    let addr = std::net::SocketAddr::from(([127, 0, 0, 1], args.port));
    println!("体检套餐管理系统服务器运行在 http://{}", addr);
    
    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}
