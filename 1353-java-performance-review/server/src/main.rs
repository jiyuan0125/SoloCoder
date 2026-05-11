use std::sync::Arc;

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use clap::Parser;
use performance_core::*;
use serde::{Deserialize, Serialize};
use tokio::sync::Mutex;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "PERFORMANCE_SERVER_PORT", default_value_t = 3000)]
    port: u16,

    #[arg(long, env = "PERFORMANCE_SERVER_HOST", default_value = "127.0.0.1")]
    host: String,
}

type AppState = Arc<Mutex<PerformanceData>>;

#[tokio::main]
async fn main() {
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "performance_server=info,tower_http=info".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    let args = Args::parse();

    let state: AppState = Arc::new(Mutex::new(PerformanceData::new()));

    let app = Router::new()
        .route("/api/health", get(health_check))
        .route("/api/departments", post(create_department).get(list_departments))
        .route("/api/departments/:id", get(get_department))
        .route("/api/employees", post(create_employee).get(list_employees))
        .route("/api/employees/:id", get(get_employee))
        .route("/api/cycles", post(create_cycle).get(list_cycles))
        .route("/api/cycles/:id", get(get_cycle))
        .route("/api/self-assessment", post(submit_self_assessment))
        .route("/api/manager-review", post(submit_manager_review))
        .route("/api/peer-review", post(submit_peer_review))
        .route("/api/cycles/:id/results", get(calculate_results))
        .route("/api/cycles/:id/departments/:dept_id/results", get(get_department_results))
        .with_state(state);

    let addr = format!("{}:{}", args.host, args.port);

    tracing::info!("listening on {}", addr);

    axum::Server::bind(&addr.parse().unwrap())
        .serve(app.into_make_service())
        .await
        .unwrap();
}

async fn health_check() -> &'static str {
    "OK"
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateDepartmentRequest {
    id: String,
    name: String,
}

async fn create_department(
    State(state): State<AppState>,
    Json(payload): Json<CreateDepartmentRequest>,
) -> impl IntoResponse {
    let mut data = state.lock().await;
    let dept = Department {
        id: payload.id,
        name: payload.name,
    };
    data.add_department(dept.clone());
    (StatusCode::CREATED, Json(dept))
}

async fn list_departments(State(state): State<AppState>) -> impl IntoResponse {
    let data = state.lock().await;
    let depts: Vec<Department> = data.departments.values().cloned().collect();
    Json(depts)
}

async fn get_department(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    let data = state.lock().await;
    match data.departments.get(&id) {
        Some(dept) => (StatusCode::OK, Json(Some(dept.clone()))),
        None => (StatusCode::NOT_FOUND, Json(None)),
    }
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateEmployeeRequest {
    id: String,
    name: String,
    manager_id: Option<String>,
    base_performance: f64,
    department_assignments: Vec<DepartmentAssignment>,
}

async fn create_employee(
    State(state): State<AppState>,
    Json(payload): Json<CreateEmployeeRequest>,
) -> impl IntoResponse {
    let mut data = state.lock().await;
    let emp = Employee {
        id: payload.id,
        name: payload.name,
        manager_id: payload.manager_id,
        base_performance: payload.base_performance,
        department_assignments: payload.department_assignments,
    };
    data.add_employee(emp.clone());
    (StatusCode::CREATED, Json(emp))
}

async fn list_employees(State(state): State<AppState>) -> impl IntoResponse {
    let data = state.lock().await;
    let emps: Vec<Employee> = data.employees.values().cloned().collect();
    Json(emps)
}

async fn get_employee(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    let data = state.lock().await;
    match data.employees.get(&id) {
        Some(emp) => (StatusCode::OK, Json(Some(emp.clone()))),
        None => (StatusCode::NOT_FOUND, Json(None)),
    }
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateCycleRequest {
    id: String,
    name: String,
    start_date: String,
    end_date: String,
}

async fn create_cycle(
    State(state): State<AppState>,
    Json(payload): Json<CreateCycleRequest>,
) -> impl IntoResponse {
    let mut data = state.lock().await;
    let cycle = ReviewCycle {
        id: payload.id,
        name: payload.name,
        start_date: payload.start_date,
        end_date: payload.end_date,
    };
    data.add_review_cycle(cycle.clone());
    (StatusCode::CREATED, Json(cycle))
}

async fn list_cycles(State(state): State<AppState>) -> impl IntoResponse {
    let data = state.lock().await;
    let cycles: Vec<ReviewCycle> = data.review_cycles.values().cloned().collect();
    Json(cycles)
}

async fn get_cycle(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    let data = state.lock().await;
    match data.review_cycles.get(&id) {
        Some(cycle) => (StatusCode::OK, Json(Some(cycle.clone()))),
        None => (StatusCode::NOT_FOUND, Json(None)),
    }
}

#[derive(Debug, Serialize, Deserialize)]
struct SubmitSelfAssessmentRequest {
    employee_id: String,
    cycle_id: String,
    score: f64,
}

async fn submit_self_assessment(
    State(state): State<AppState>,
    Json(payload): Json<SubmitSelfAssessmentRequest>,
) -> impl IntoResponse {
    let mut data = state.lock().await;
    let sa = SelfAssessment {
        employee_id: payload.employee_id,
        cycle_id: payload.cycle_id,
        score: payload.score,
    };
    data.add_self_assessment(sa.clone());
    (StatusCode::CREATED, Json(sa))
}

#[derive(Debug, Serialize, Deserialize)]
struct SubmitManagerReviewRequest {
    employee_id: String,
    cycle_id: String,
    manager_id: String,
    score: f64,
}

async fn submit_manager_review(
    State(state): State<AppState>,
    Json(payload): Json<SubmitManagerReviewRequest>,
) -> impl IntoResponse {
    let mut data = state.lock().await;
    let mr = ManagerReview {
        employee_id: payload.employee_id,
        cycle_id: payload.cycle_id,
        manager_id: payload.manager_id,
        score: payload.score,
    };
    data.add_manager_review(mr.clone());
    (StatusCode::CREATED, Json(mr))
}

#[derive(Debug, Serialize, Deserialize)]
struct SubmitPeerReviewRequest {
    reviewer_id: String,
    reviewee_id: String,
    cycle_id: String,
    score: f64,
}

async fn submit_peer_review(
    State(state): State<AppState>,
    Json(payload): Json<SubmitPeerReviewRequest>,
) -> impl IntoResponse {
    let mut data = state.lock().await;
    let pr = PeerReview {
        reviewer_id: payload.reviewer_id,
        reviewee_id: payload.reviewee_id,
        cycle_id: payload.cycle_id,
        score: payload.score,
    };
    data.add_peer_review(pr.clone());
    (StatusCode::CREATED, Json(pr))
}

async fn calculate_results(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    let data = state.lock().await;
    let results = calculate_cycle_results(&data, &id);
    Json(results)
}

async fn get_department_results(
    State(state): State<AppState>,
    Path((cycle_id, dept_id)): Path<(String, String)>,
) -> impl IntoResponse {
    let data = state.lock().await;
    let all_results = calculate_cycle_results(&data, &cycle_id);
    let dept_result = all_results.into_iter().find(|r| r.department_id == dept_id);
    match dept_result {
        Some(r) => (StatusCode::OK, Json(Some(r))),
        None => (StatusCode::NOT_FOUND, Json(None)),
    }
}
