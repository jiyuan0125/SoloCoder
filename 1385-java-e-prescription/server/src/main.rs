use std::net::SocketAddr;
use std::sync::Arc;

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use clap::Parser;
use eprescription_core::{
    services::PrescriptionError, CreatePrescriptionRequest, DispensePrescriptionRequest,
    InMemoryStorage, PrescriptionService, ReviewPrescriptionRequest,
};
use tower_http::cors::{Any, CorsLayer};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "PORT", default_value_t = 3000)]
    port: u16,
}

struct AppState {
    service: PrescriptionService,
}

#[tokio::main]
async fn main() {
    let args = Args::parse();

    let storage = InMemoryStorage::new();
    let service = PrescriptionService::new(storage);

    let app_state = Arc::new(AppState { service });

    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    let app = Router::new()
        .route("/prescriptions", post(create_prescription))
        .route("/prescriptions", get(list_prescriptions))
        .route("/prescriptions/:id", get(get_prescription))
        .route("/prescriptions/review", post(review_prescription))
        .route("/prescriptions/dispense", post(dispense_prescription))
        .route("/doctors/stats", get(get_doctor_stats))
        .route("/medicines/stats", get(get_medicine_stats))
        .route("/medicines", get(list_medicines))
        .route("/patients", get(list_patients))
        .route("/doctors", get(list_doctors))
        .with_state(app_state)
        .layer(cors);

    let addr = SocketAddr::from(([0, 0, 0, 0], args.port));
    println!("服务器监听于 http://{}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}

async fn create_prescription(
    State(state): State<Arc<AppState>>,
    Json(request): Json<CreatePrescriptionRequest>,
) -> impl IntoResponse {
    match state.service.create_prescription(request) {
        Ok((prescription, check_result)) => (
            StatusCode::CREATED,
            Json(serde_json::json!({
                "prescription": prescription,
                "check_result": check_result
            })),
        )
            .into_response(),
        Err(e) => handle_error(e),
    }
}

async fn list_prescriptions(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let prescriptions = state.service.get_all_prescriptions();
    Json(prescriptions).into_response()
}

async fn get_prescription(
    State(state): State<Arc<AppState>>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match uuid::Uuid::parse_str(&id) {
        Ok(uuid) => match state.service.get_prescription(uuid) {
            Some(prescription) => Json(prescription).into_response(),
            None => (
                StatusCode::NOT_FOUND,
                Json(serde_json::json!({"error": "处方未找到"})),
            )
                .into_response(),
        },
        Err(_) => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": "无效的处方ID"})),
        )
            .into_response(),
    }
}

async fn review_prescription(
    State(state): State<Arc<AppState>>,
    Json(request): Json<ReviewPrescriptionRequest>,
) -> impl IntoResponse {
    match state.service.review_prescription(request) {
        Ok(prescription) => Json(prescription).into_response(),
        Err(e) => handle_error(e),
    }
}

async fn dispense_prescription(
    State(state): State<Arc<AppState>>,
    Json(request): Json<DispensePrescriptionRequest>,
) -> impl IntoResponse {
    match state.service.dispense_prescription(request) {
        Ok(prescription) => Json(prescription).into_response(),
        Err(e) => handle_error(e),
    }
}

async fn get_doctor_stats(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let stats = state.service.get_all_doctors_monthly_stats();
    Json(stats).into_response()
}

async fn get_medicine_stats(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let stats = state.service.get_medicine_usage_stats();
    Json(stats).into_response()
}

async fn list_medicines(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let medicines = state.service.get_all_medicines();
    Json(medicines).into_response()
}

async fn list_patients(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let patients = state.service.get_all_patients();
    Json(patients).into_response()
}

async fn list_doctors(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let doctors = state.service.get_all_doctors();
    Json(doctors).into_response()
}

fn handle_error(e: PrescriptionError) -> axum::response::Response {
    match e {
        PrescriptionError::NotFound => (
            StatusCode::NOT_FOUND,
            Json(serde_json::json!({"error": e.to_string()})),
        ),
        PrescriptionError::InvalidState => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": e.to_string()})),
        ),
        PrescriptionError::InsufficientStock => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": e.to_string()})),
        ),
        PrescriptionError::MedicineNotFound => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": e.to_string()})),
        ),
        PrescriptionError::HasSevereWarnings => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": e.to_string()})),
        ),
    }
    .into_response()
}
