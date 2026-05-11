use std::sync::Arc;

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::{IntoResponse, Json, Response},
    routing::{get, post, delete},
    Router,
};
use chrono::NaiveDate;
use clap::Parser;
use dialysis_core::*;
use serde::Serialize;
use tokio::sync::Mutex;
use uuid::Uuid;

#[derive(Parser, Debug, Clone)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "DIALYSIS_PORT", default_value_t = 3000)]
    port: u16,

    #[arg(long, env = "DIALYSIS_HOST", default_value = "127.0.0.1")]
    host: String,
}

type AppState = Arc<Mutex<Scheduler>>;

#[derive(Debug, Clone, Serialize)]
struct ApiError {
    error: String,
    message: String,
}

struct AppError(SchedulerError);

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let status = match &self.0 {
            SchedulerError::PatientNotFound(_)
            | SchedulerError::MachineNotFound(_)
            | SchedulerError::ScheduleNotFound(_)
            | SchedulerError::MaintenanceNotFound(_) => StatusCode::NOT_FOUND,
            _ => StatusCode::BAD_REQUEST,
        };

        (
            status,
            Json(ApiError {
                error: std::any::type_name::<SchedulerError>().to_string(),
                message: self.0.to_string(),
            }),
        )
            .into_response()
    }
}

impl<E: Into<SchedulerError>> From<E> for AppError {
    fn from(err: E) -> Self {
        AppError(err.into())
    }
}

async fn list_patients(State(state): State<AppState>) -> Json<Vec<Patient>> {
    let scheduler = state.lock().await;
    Json(scheduler.list_patients().into_iter().cloned().collect())
}

async fn create_patient(
    State(state): State<AppState>,
    Json(req): Json<CreatePatientRequest>,
) -> Json<Patient> {
    let mut scheduler = state.lock().await;
    let preferred_days = req.preferred_days.into_iter().collect();
    let preferred_slots = req.preferred_slots.into_iter().collect();
    let patient = Patient::new(
        req.name,
        req.disease,
        req.dialysis_frequency_per_week,
        preferred_days,
        preferred_slots,
    );
    Json(scheduler.add_patient(patient).clone())
}

async fn get_patient(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<Patient>, AppError> {
    let scheduler = state.lock().await;
    Ok(Json(scheduler.get_patient(&id)?.clone()))
}

async fn list_machines(State(state): State<AppState>) -> Json<Vec<Machine>> {
    let scheduler = state.lock().await;
    Json(scheduler.list_machines().into_iter().cloned().collect())
}

async fn create_machine(
    State(state): State<AppState>,
    Json(req): Json<CreateMachineRequest>,
) -> Json<Machine> {
    let mut scheduler = state.lock().await;
    let machine = Machine::new(req.name, req.is_hepatitis_b_only);
    Json(scheduler.add_machine(machine).clone())
}

async fn get_machine(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<Machine>, AppError> {
    let scheduler = state.lock().await;
    Ok(Json(scheduler.get_machine(&id)?.clone()))
}

async fn list_schedules(State(state): State<AppState>) -> Json<Vec<ScheduleEntry>> {
    let scheduler = state.lock().await;
    Json(scheduler.list_schedules().into_iter().cloned().collect())
}

async fn create_schedule(
    State(state): State<AppState>,
    Json(req): Json<CreateScheduleRequest>,
) -> Result<Json<ScheduleEntry>, AppError> {
    let mut scheduler = state.lock().await;
    Ok(Json(scheduler.create_schedule(req)?.clone()))
}

async fn bulk_create_schedules(
    State(state): State<AppState>,
    Json(req): Json<BulkCreateScheduleRequest>,
) -> Result<Json<Vec<ScheduleEntry>>, AppError> {
    let mut scheduler = state.lock().await;
    Ok(Json(
        scheduler
            .bulk_create_schedules(req)?
            .into_iter()
            .cloned()
            .collect(),
    ))
}

async fn get_schedule(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<ScheduleEntry>, AppError> {
    let scheduler = state.lock().await;
    Ok(Json(scheduler.get_schedule(&id)?.clone()))
}

async fn cancel_schedule(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<ScheduleEntry>, AppError> {
    let mut scheduler = state.lock().await;
    Ok(Json(scheduler.cancel_schedule(&id)?))
}

async fn reschedule(
    State(state): State<AppState>,
    Json(req): Json<RescheduleRequest>,
) -> Result<Json<ScheduleEntry>, AppError> {
    let mut scheduler = state.lock().await;
    Ok(Json(scheduler.reschedule(req)?.clone()))
}

async fn list_maintenances(State(state): State<AppState>) -> Json<Vec<Maintenance>> {
    let scheduler = state.lock().await;
    Json(scheduler.list_maintenances().into_iter().cloned().collect())
}

#[derive(Serialize)]
struct MaintenanceResponse {
    maintenance: Maintenance,
    affected: Vec<RescheduleResult>,
}

async fn create_maintenance(
    State(state): State<AppState>,
    Json(req): Json<CreateMaintenanceRequest>,
) -> Result<Json<MaintenanceResponse>, AppError> {
    let mut scheduler = state.lock().await;
    let (maintenance, affected) = scheduler.create_maintenance(req)?;
    Ok(Json(MaintenanceResponse {
        maintenance,
        affected,
    }))
}

async fn cancel_maintenance(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<Maintenance>, AppError> {
    let mut scheduler = state.lock().await;
    Ok(Json(scheduler.cancel_maintenance(&id)?))
}

async fn get_schedules_by_date(
    State(state): State<AppState>,
    Path(date): Path<String>,
) -> Result<Json<Vec<ScheduleEntry>>, AppError> {
    let date = NaiveDate::parse_from_str(&date, "%Y-%m-%d")
        .map_err(|_| SchedulerError::PatientNotFound("Invalid date format".to_string()))?;
    let scheduler = state.lock().await;
    Ok(Json(
        scheduler
            .get_schedules_by_date(date)
            .into_iter()
            .cloned()
            .collect(),
    ))
}

fn app(state: AppState) -> Router {
    Router::new()
        .route("/patients", get(list_patients).post(create_patient))
        .route("/patients/:id", get(get_patient))
        .route("/machines", get(list_machines).post(create_machine))
        .route("/machines/:id", get(get_machine))
        .route("/schedules", get(list_schedules).post(create_schedule))
        .route("/schedules/bulk", post(bulk_create_schedules))
        .route("/schedules/:id", get(get_schedule).delete(cancel_schedule))
        .route("/schedules/reschedule", post(reschedule))
        .route("/schedules/date/:date", get(get_schedules_by_date))
        .route("/maintenances", get(list_maintenances).post(create_maintenance))
        .route("/maintenances/:id", delete(cancel_maintenance))
        .with_state(state)
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_env_filter(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "dialysis_server=info,tower_http=info".into()),
        )
        .init();

    let args = Args::parse();

    let state = Arc::new(Mutex::new(Scheduler::new()));
    let app = app(state);

    let addr = format!("{}:{}", args.host, args.port);
    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    tracing::info!("listening on {}", addr);

    axum::serve(listener, app).await.unwrap();
}
