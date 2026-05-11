use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::{IntoResponse, Json},
    routing::{get, post},
    Router,
};
use clap::Parser;
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use tokio::sync::Mutex;
use tracing_subscriber;
use uuid::Uuid;
use chrono::{DateTime, Utc};
use vaccination_core::{
    ContraindicationTag, InMemoryStore, Person, Vaccine, VaccineDoseSchedule,
    VaccinationService, VaccinationError,
};

#[derive(Parser, Debug)]
#[command(version, about, long_about = None)]
struct Args {
    #[arg(long, env = "VACCINATION_PORT", default_value_t = 8080)]
    port: u16,

    #[arg(long, env = "VACCINATION_HOST", default_value = "127.0.0.1")]
    host: String,
}

type AppState = Arc<Mutex<VaccinationService>>;

#[derive(Debug, Serialize, Deserialize)]
struct ErrorResponse {
    error: String,
}

struct AppError(VaccinationError);

impl IntoResponse for AppError {
    fn into_response(self) -> axum::response::Response {
        let status = match &self.0 {
            VaccinationError::VaccineNotFound => StatusCode::NOT_FOUND,
            VaccinationError::PersonNotFound => StatusCode::NOT_FOUND,
            VaccinationError::InvalidDoseNumber { .. } => StatusCode::BAD_REQUEST,
            VaccinationError::IntervalNotMet { .. } => StatusCode::BAD_REQUEST,
            VaccinationError::RecordNotFound => StatusCode::NOT_FOUND,
            VaccinationError::RecordAlreadyVoided => StatusCode::BAD_REQUEST,
        };
        (
            status,
            Json(ErrorResponse {
                error: self.0.to_string(),
            }),
        )
            .into_response()
    }
}

impl From<VaccinationError> for AppError {
    fn from(err: VaccinationError) -> Self {
        AppError(err)
    }
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateVaccineRequest {
    name: String,
    total_doses: u32,
    dose_schedules: Vec<VaccineDoseSchedule>,
    contraindications: Vec<String>,
}

#[derive(Debug, Serialize, Deserialize)]
struct CreatePersonRequest {
    name: String,
    contraindications: Vec<String>,
}

#[derive(Debug, Serialize, Deserialize)]
struct RecordVaccinationRequest {
    person_id: Uuid,
    vaccine_id: Uuid,
    vaccination_date: DateTime<Utc>,
}

#[derive(Debug, Serialize, Deserialize)]
struct ValidateVaccinationRequest {
    person_id: Uuid,
    vaccine_id: Uuid,
    vaccination_date: DateTime<Utc>,
}

fn init_sample_data(store: &mut InMemoryStore) {
    let hpv = Vaccine {
        id: Uuid::new_v4(),
        name: "HPV疫苗".to_string(),
        total_doses: 3,
        dose_schedules: vec![
            VaccineDoseSchedule { dose_number: 1, min_interval_days: 0 },
            VaccineDoseSchedule { dose_number: 2, min_interval_days: 30 },
            VaccineDoseSchedule { dose_number: 3, min_interval_days: 180 },
        ],
        contraindications: vec![ContraindicationTag("过敏体质".to_string())],
    };

    let covid = Vaccine {
        id: Uuid::new_v4(),
        name: "新冠疫苗".to_string(),
        total_doses: 2,
        dose_schedules: vec![
            VaccineDoseSchedule { dose_number: 1, min_interval_days: 0 },
            VaccineDoseSchedule { dose_number: 2, min_interval_days: 21 },
        ],
        contraindications: vec![
            ContraindicationTag("发热".to_string()),
            ContraindicationTag("严重慢性病".to_string()),
        ],
    };

    store.add_vaccine(hpv);
    store.add_vaccine(covid);

    let person1 = Person {
        id: Uuid::new_v4(),
        name: "张三".to_string(),
        contraindications: vec![],
    };

    let person2 = Person {
        id: Uuid::new_v4(),
        name: "李四".to_string(),
        contraindications: vec![ContraindicationTag("过敏体质".to_string())],
    };

    store.add_person(person1);
    store.add_person(person2);
}

async fn create_vaccine(
    State(state): State<AppState>,
    Json(req): Json<CreateVaccineRequest>,
) -> impl IntoResponse {
    let mut service = state.lock().await;
    let vaccine = Vaccine {
        id: Uuid::new_v4(),
        name: req.name,
        total_doses: req.total_doses,
        dose_schedules: req.dose_schedules,
        contraindications: req
            .contraindications
            .into_iter()
            .map(ContraindicationTag)
            .collect(),
    };
    service.store_mut().add_vaccine(vaccine.clone());
    (StatusCode::CREATED, Json(vaccine))
}

async fn list_vaccines(State(state): State<AppState>) -> impl IntoResponse {
    let service = state.lock().await;
    Json(service.store().list_vaccines())
}

async fn get_vaccine(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<impl IntoResponse, StatusCode> {
    let service = state.lock().await;
    service
        .store()
        .get_vaccine(&id)
        .map(|v| Json(v.clone()))
        .ok_or(StatusCode::NOT_FOUND)
}

async fn create_person(
    State(state): State<AppState>,
    Json(req): Json<CreatePersonRequest>,
) -> impl IntoResponse {
    let mut service = state.lock().await;
    let person = Person {
        id: Uuid::new_v4(),
        name: req.name,
        contraindications: req
            .contraindications
            .into_iter()
            .map(ContraindicationTag)
            .collect(),
    };
    service.store_mut().add_person(person.clone());
    (StatusCode::CREATED, Json(person))
}

async fn list_persons(State(state): State<AppState>) -> impl IntoResponse {
    let service = state.lock().await;
    Json(service.store().list_persons())
}

async fn get_person(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<impl IntoResponse, StatusCode> {
    let service = state.lock().await;
    service
        .store()
        .get_person(&id)
        .map(|p| Json(p.clone()))
        .ok_or(StatusCode::NOT_FOUND)
}

async fn validate_vaccination(
    State(state): State<AppState>,
    Json(req): Json<ValidateVaccinationRequest>,
) -> Result<impl IntoResponse, AppError> {
    let service = state.lock().await;
    let result = service.validate_vaccination(
        &req.person_id,
        &req.vaccine_id,
        &req.vaccination_date,
    )?;
    Ok(Json(result))
}

async fn record_vaccination(
    State(state): State<AppState>,
    Json(req): Json<RecordVaccinationRequest>,
) -> Result<impl IntoResponse, AppError> {
    let mut service = state.lock().await;
    let record = service.record_vaccination(
        req.person_id,
        req.vaccine_id,
        req.vaccination_date,
    )?;
    Ok((StatusCode::CREATED, Json(record)))
}

async fn list_records(State(state): State<AppState>) -> impl IntoResponse {
    let service = state.lock().await;
    Json(service.store().list_records())
}

async fn get_record(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<impl IntoResponse, StatusCode> {
    let service = state.lock().await;
    service
        .store()
        .get_record(&id)
        .map(|r| Json(r.clone()))
        .ok_or(StatusCode::NOT_FOUND)
}

async fn void_record(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<impl IntoResponse, AppError> {
    let mut service = state.lock().await;
    service.store_mut().void_record(&id)?;
    Ok(Json(serde_json::json!({ "status": "ok" })))
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_env_filter(tracing_subscriber::EnvFilter::from_default_env())
        .init();

    let args = Args::parse();

    let mut service = VaccinationService::new();
    init_sample_data(service.store_mut());

    let app_state = Arc::new(Mutex::new(service));

    let app = Router::new()
        .route("/api/vaccines", post(create_vaccine).get(list_vaccines))
        .route("/api/vaccines/:id", get(get_vaccine))
        .route("/api/persons", post(create_person).get(list_persons))
        .route("/api/persons/:id", get(get_person))
        .route("/api/vaccinations/validate", post(validate_vaccination))
        .route("/api/vaccinations", post(record_vaccination).get(list_records))
        .route("/api/vaccinations/:id", get(get_record))
        .route("/api/vaccinations/:id/void", post(void_record))
        .with_state(app_state);

    let addr = format!("{}:{}", args.host, args.port);
    tracing::info!("Server listening on {}", addr);
    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
