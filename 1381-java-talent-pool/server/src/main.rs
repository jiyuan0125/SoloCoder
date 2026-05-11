use std::net::SocketAddr;
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
use talent_core::{
    AppError, GridZone,
    TalentService,
};
use tower_http::cors::CorsLayer;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "TALENT_PORT", default_value_t = 3000)]
    port: u16,
    
    #[arg(short, long, env = "TALENT_HOST", default_value = "127.0.0.1")]
    host: String,
}

struct AppState {
    service: TalentService,
}

impl AppState {
    fn new() -> Self {
        Self {
            service: TalentService::default(),
        }
    }
}

#[tokio::main]
async fn main() {
    let args = Args::parse();
    
    let state = Arc::new(AppState::new());
    
    let app = Router::new()
        .route("/api/departments", get(list_departments).post(create_department))
        .route("/api/employees", get(list_employees).post(create_employee))
        .route("/api/employees/:id/assessments", get(list_assessments).post(create_assessment))
        .route("/api/succession-plans", get(list_succession_plans).post(create_succession_plan))
        .route("/api/succession-plans/:id/candidates", post(add_candidate))
        .route("/api/grid/distribution", get(grid_distribution))
        .route("/api/grid/employees", get(employees_by_zone))
        .route("/api/warnings", get(warnings))
        .with_state(state)
        .layer(CorsLayer::permissive());
    
    let addr: SocketAddr = format!("{}:{}", args.host, args.port)
        .parse()
        .expect("Invalid host/port");
    
    println!("Talent Pool Server listening on {}", addr);
    
    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}

#[derive(Debug, Deserialize)]
struct CreateDepartmentRequest {
    name: String,
}

async fn create_department(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateDepartmentRequest>,
) -> impl IntoResponse {
    let dept = state.service.create_department(req.name).await;
    (StatusCode::CREATED, Json(dept))
}

async fn list_departments(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let depts = state.service.get_departments().await;
    Json(depts)
}

#[derive(Debug, Deserialize)]
struct CreateEmployeeRequest {
    name: String,
    employee_number: String,
    department_id: uuid::Uuid,
    position: String,
}

async fn create_employee(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateEmployeeRequest>,
) -> Result<impl IntoResponse, AppErrorResponse> {
    let emp = state.service.create_employee(
        req.name,
        req.employee_number,
        req.department_id,
        req.position,
    ).await?;
    Ok((StatusCode::CREATED, Json(emp)))
}

#[derive(Debug, Deserialize)]
struct ListEmployeesQuery {
    department_id: Option<uuid::Uuid>,
}

async fn list_employees(
    State(state): State<Arc<AppState>>,
    Query(query): Query<ListEmployeesQuery>,
) -> impl IntoResponse {
    let employees = state.service.get_employees(query.department_id).await;
    Json(employees)
}

#[derive(Debug, Deserialize)]
struct CreateAssessmentRequest {
    year: i32,
    performance: u8,
    potential: u8,
}

async fn create_assessment(
    State(state): State<Arc<AppState>>,
    Path(employee_id): Path<uuid::Uuid>,
    Json(req): Json<CreateAssessmentRequest>,
) -> Result<impl IntoResponse, AppErrorResponse> {
    let assessment = state.service.create_assessment(
        employee_id,
        req.year,
        req.performance,
        req.potential,
    ).await?;
    Ok((StatusCode::CREATED, Json(assessment)))
}

async fn list_assessments(
    State(state): State<Arc<AppState>>,
    Path(employee_id): Path<uuid::Uuid>,
) -> impl IntoResponse {
    let assessments = state.service.get_assessments(employee_id).await;
    Json(assessments)
}

#[derive(Debug, Deserialize)]
struct CreateSuccessionPlanRequest {
    position: String,
    department_id: uuid::Uuid,
    is_key_position: bool,
}

async fn create_succession_plan(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateSuccessionPlanRequest>,
) -> Result<impl IntoResponse, AppErrorResponse> {
    let plan = state.service.create_succession_plan(
        req.position,
        req.department_id,
        req.is_key_position,
    ).await?;
    Ok((StatusCode::CREATED, Json(plan)))
}

#[derive(Debug, Deserialize)]
struct ListSuccessionPlansQuery {
    department_id: Option<uuid::Uuid>,
}

async fn list_succession_plans(
    State(state): State<Arc<AppState>>,
    Query(query): Query<ListSuccessionPlansQuery>,
) -> impl IntoResponse {
    let plans = state.service.get_succession_plans(query.department_id).await;
    Json(plans)
}

#[derive(Debug, Deserialize)]
struct AddCandidateRequest {
    employee_id: uuid::Uuid,
    assessment_year: Option<i32>,
    notes: Option<String>,
}

async fn add_candidate(
    State(state): State<Arc<AppState>>,
    Path(plan_id): Path<uuid::Uuid>,
    Json(req): Json<AddCandidateRequest>,
) -> Result<impl IntoResponse, AppErrorResponse> {
    let plan = state.service.add_candidate_to_succession_plan(
        plan_id,
        req.employee_id,
        req.assessment_year,
        req.notes,
    ).await?;
    Ok((StatusCode::OK, Json(plan)))
}

#[derive(Debug, Deserialize)]
struct GridDistributionQuery {
    year: i32,
    department_id: Option<uuid::Uuid>,
}

async fn grid_distribution(
    State(state): State<Arc<AppState>>,
    Query(query): Query<GridDistributionQuery>,
) -> impl IntoResponse {
    let distribution = state.service.get_grid_distribution(
        query.year,
        query.department_id,
    ).await;
    Json(distribution)
}

#[derive(Debug, Deserialize)]
struct EmployeesByZoneQuery {
    year: i32,
    zone: GridZone,
    department_id: Option<uuid::Uuid>,
}

#[derive(Debug, serde::Serialize)]
struct EmployeeWithAssessment {
    employee: talent_core::Employee,
    assessment: talent_core::Assessment,
}

async fn employees_by_zone(
    State(state): State<Arc<AppState>>,
    Query(query): Query<EmployeesByZoneQuery>,
) -> impl IntoResponse {
    let results: Vec<(talent_core::Employee, talent_core::Assessment)> = state.service.get_employees_by_zone(
        query.year,
        query.zone,
        query.department_id,
    ).await;
    
    let response: Vec<EmployeeWithAssessment> = results
        .into_iter()
        .map(|(e, a)| EmployeeWithAssessment {
            employee: e,
            assessment: a,
        })
        .collect();
    
    Json(response)
}

async fn warnings(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let warnings = state.service.get_succession_warnings().await;
    Json(warnings)
}

struct AppErrorResponse(AppError);

impl From<AppError> for AppErrorResponse {
    fn from(err: AppError) -> Self {
        AppErrorResponse(err)
    }
}

impl IntoResponse for AppErrorResponse {
    fn into_response(self) -> axum::response::Response {
        let status = match &self.0 {
            AppError::DepartmentNotFound(_) => StatusCode::NOT_FOUND,
            AppError::EmployeeNotFound(_) => StatusCode::NOT_FOUND,
            AppError::EmployeeAlreadyExists(_) => StatusCode::CONFLICT,
            AppError::PositionNotFound(_) => StatusCode::NOT_FOUND,
            AppError::AssessmentNotFound(_) => StatusCode::NOT_FOUND,
            AppError::AssessmentAlreadyExistsForYear(_) => StatusCode::CONFLICT,
            AppError::InvalidRating => StatusCode::BAD_REQUEST,
            AppError::DataIsolationViolation => StatusCode::FORBIDDEN,
            AppError::Internal(_) => StatusCode::INTERNAL_SERVER_ERROR,
        };
        
        let body = serde_json::json!({
            "error": self.0.to_string(),
        });
        
        (status, Json(body)).into_response()
    }
}
