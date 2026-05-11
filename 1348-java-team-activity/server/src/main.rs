use std::net::SocketAddr;
use std::sync::Arc;

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::{IntoResponse, Json},
    routing::{get, post, put},
    Router,
};
use clap::Parser;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

use activity_core::{
    ActivityService, ActivityError, CreateActivityRequest, InMemoryStorage, SettlingRequest, UpdateLimitRequest,
};

#[derive(Parser, Debug)]
#[command(name = "team-activity-server")]
struct Args {
    #[arg(short, long, env = "SERVER_PORT", default_value_t = 3000)]
    port: u16,

    #[arg(long, env = "SERVER_HOST", default_value = "0.0.0.0")]
    host: String,
}

#[derive(Clone)]
struct AppState {
    service: Arc<ActivityService>,
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateDepartmentRequest {
    name: String,
    annual_budget: f64,
}

#[derive(Debug, Serialize, Deserialize)]
struct RegisterRequest {
    user_id: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct ApiErrorResponse {
    error: String,
}

struct AppError(ActivityError);

impl IntoResponse for AppError {
    fn into_response(self) -> axum::response::Response {
        let status = match &self.0 {
            ActivityError::DepartmentNotFound => StatusCode::NOT_FOUND,
            ActivityError::ActivityNotFound => StatusCode::NOT_FOUND,
            ActivityError::UserAlreadyRegistered => StatusCode::CONFLICT,
            ActivityError::ActivityFull => StatusCode::CONFLICT,
            ActivityError::InsufficientBudget => StatusCode::BAD_REQUEST,
            ActivityError::RegistrationClosed => StatusCode::BAD_REQUEST,
            ActivityError::CannotIncreaseLimit => StatusCode::BAD_REQUEST,
            ActivityError::InvalidState => StatusCode::BAD_REQUEST,
            ActivityError::ParticipantCountMismatch => StatusCode::BAD_REQUEST,
            ActivityError::UserNotRegistered => StatusCode::NOT_FOUND,
            ActivityError::CannotCancelStartedActivity => StatusCode::BAD_REQUEST,
            ActivityError::ConcurrentConflict => StatusCode::CONFLICT,
        };

        let body = Json(ApiErrorResponse {
            error: self.0.to_string(),
        });

        (status, body).into_response()
    }
}

impl From<ActivityError> for AppError {
    fn from(err: ActivityError) -> Self {
        Self(err)
    }
}

async fn create_department(
    State(state): State<AppState>,
    Json(req): Json<CreateDepartmentRequest>,
) -> Json<activity_core::Department> {
    let dept = state
        .service
        .create_department(req.name, req.annual_budget)
        .await;
    Json(dept)
}

async fn get_all_departments(
    State(state): State<AppState>,
) -> Json<Vec<activity_core::Department>> {
    let depts = state.service.get_all_departments().await;
    Json(depts)
}

async fn get_department(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<activity_core::Department>, AppError> {
    state
        .service
        .get_department(id)
        .await
        .map(Json)
        .ok_or(ActivityError::DepartmentNotFound.into())
}

async fn create_activity(
    State(state): State<AppState>,
    Json(req): Json<CreateActivityRequest>,
) -> Result<Json<activity_core::Activity>, AppError> {
    let activity = state.service.create_activity(req).await?;
    Ok(Json(activity))
}

async fn get_all_activities(
    State(state): State<AppState>,
) -> Json<Vec<activity_core::Activity>> {
    let activities = state.service.get_all_activities().await;
    Json(activities)
}

async fn get_activity(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<activity_core::Activity>, AppError> {
    state
        .service
        .get_activity(id)
        .await
        .map(Json)
        .ok_or(ActivityError::ActivityNotFound.into())
}

async fn publish_activity(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<activity_core::Activity>, AppError> {
    let activity = state.service.publish_activity(id).await?;
    Ok(Json(activity))
}

async fn cancel_activity(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<activity_core::Activity>, AppError> {
    let activity = state.service.cancel_activity(id).await?;
    Ok(Json(activity))
}

async fn update_registration_limit(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(req): Json<UpdateLimitRequest>,
) -> Result<Json<activity_core::Activity>, AppError> {
    let activity = state.service.update_registration_limit(id, req).await?;
    Ok(Json(activity))
}

async fn register_for_activity(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(req): Json<RegisterRequest>,
) -> Result<Json<activity_core::RegistrationStatus>, AppError> {
    let status = state.service.register_for_activity(id, req.user_id).await?;
    Ok(Json(status))
}

async fn cancel_registration(
    State(state): State<AppState>,
    Path((activity_id, user_id)): Path<(Uuid, String)>,
) -> Result<(), AppError> {
    state
        .service
        .cancel_registration(activity_id, user_id)
        .await?;
    Ok(())
}

async fn close_registration(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<activity_core::Activity>, AppError> {
    let activity = state.service.close_registration(id).await?;
    Ok(Json(activity))
}

async fn settle_activity(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(req): Json<SettlingRequest>,
) -> Result<Json<activity_core::Activity>, AppError> {
    let activity = state.service.settle_activity(id, req).await?;
    Ok(Json(activity))
}

fn create_router(service: Arc<ActivityService>) -> Router {
    let state = AppState { service };

    Router::new()
        .route("/api/departments", get(get_all_departments).post(create_department))
        .route("/api/departments/:id", get(get_department))
        .route("/api/activities", get(get_all_activities).post(create_activity))
        .route("/api/activities/:id", get(get_activity))
        .route("/api/activities/:id/publish", post(publish_activity))
        .route("/api/activities/:id/cancel", post(cancel_activity))
        .route(
            "/api/activities/:id/limit",
            put(update_registration_limit),
        )
        .route("/api/activities/:id/register", post(register_for_activity))
        .route(
            "/api/activities/:activity_id/register/:user_id",
            axum::routing::delete(cancel_registration),
        )
        .route(
            "/api/activities/:id/close-registration",
            post(close_registration),
        )
        .route("/api/activities/:id/settle", post(settle_activity))
        .with_state(state)
}

#[tokio::main]
async fn main() {
    let args = Args::parse();

    let storage = Arc::new(InMemoryStorage::new());
    let service = Arc::new(ActivityService::new(storage));

    let app = create_router(service);

    let addr: SocketAddr = format!("{}:{}", args.host, args.port).parse().unwrap();
    println!("Server listening on {}", addr);

    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}
