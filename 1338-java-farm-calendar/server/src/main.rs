use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::{IntoResponse, Json},
    routing::{get, post},
    Router,
};
use clap::Parser;
use agri_core as core;
use core::{
    AppError, AgriculturalService, CreateCropRequest,
    CreatePlotRequest, CreateRegionRequest, CreateSowingPlanRequest, InMemoryStore,
};
use serde_json::json;
use std::net::SocketAddr;
use std::sync::Arc;
use tower_http::cors::{Any, CorsLayer};
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "SERVER_PORT", default_value_t = 3000)]
    port: u16,

    #[arg(short, long, env = "SERVER_HOST", default_value = "127.0.0.1")]
    host: String,
}

type AppState = Arc<AgriculturalService<InMemoryStore>>;

fn error_to_response(error: AppError) -> (StatusCode, Json<serde_json::Value>) {
    match error {
        AppError::Validation(msg) => (
            StatusCode::BAD_REQUEST,
            Json(json!({"error": "Validation", "message": msg})),
        ),
        AppError::Conflict(msg) => (
            StatusCode::CONFLICT,
            Json(json!({"error": "Conflict", "message": msg})),
        ),
        AppError::NotFound(msg) => (
            StatusCode::NOT_FOUND,
            Json(json!({"error": "NotFound", "message": msg})),
        ),
        AppError::InvalidDate(msg) => (
            StatusCode::BAD_REQUEST,
            Json(json!({"error": "InvalidDate", "message": msg})),
        ),
        AppError::Internal(msg) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(json!({"error": "Internal", "message": msg})),
        ),
    }
}

struct ApiErrorWrapper(AppError);

impl IntoResponse for ApiErrorWrapper {
    fn into_response(self) -> axum::response::Response {
        let (status, body) = error_to_response(self.0);
        (status, body).into_response()
    }
}

impl From<AppError> for ApiErrorWrapper {
    fn from(err: AppError) -> Self {
        ApiErrorWrapper(err)
    }
}

async fn create_crop(
    State(state): State<AppState>,
    Json(request): Json<CreateCropRequest>,
) -> Result<impl IntoResponse, ApiErrorWrapper> {
    let crop = state.create_crop(request).map_err(ApiErrorWrapper::from)?;
    Ok((StatusCode::CREATED, Json(crop)))
}

async fn get_crop(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<impl IntoResponse, ApiErrorWrapper> {
    let crop = state.get_crop(id).map_err(ApiErrorWrapper::from)?;
    Ok(Json(crop))
}

async fn get_all_crops(State(state): State<AppState>) -> impl IntoResponse {
    let crops = state.get_all_crops();
    Json(crops)
}

async fn create_region(
    State(state): State<AppState>,
    Json(request): Json<CreateRegionRequest>,
) -> Result<impl IntoResponse, ApiErrorWrapper> {
    let region = state.create_region(request).map_err(ApiErrorWrapper::from)?;
    Ok((StatusCode::CREATED, Json(region)))
}

async fn get_region(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<impl IntoResponse, ApiErrorWrapper> {
    let region = state.get_region(id).map_err(ApiErrorWrapper::from)?;
    Ok(Json(region))
}

async fn get_all_regions(State(state): State<AppState>) -> impl IntoResponse {
    let regions = state.get_all_regions();
    Json(regions)
}

async fn create_plot(
    State(state): State<AppState>,
    Json(request): Json<CreatePlotRequest>,
) -> Result<impl IntoResponse, ApiErrorWrapper> {
    let plot = state.create_plot(request).map_err(ApiErrorWrapper::from)?;
    Ok((StatusCode::CREATED, Json(plot)))
}

async fn get_plot(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<impl IntoResponse, ApiErrorWrapper> {
    let plot = state.get_plot(id).map_err(ApiErrorWrapper::from)?;
    Ok(Json(plot))
}

async fn get_all_plots(State(state): State<AppState>) -> impl IntoResponse {
    let plots = state.get_all_plots();
    Json(plots)
}

async fn create_sowing_plan(
    State(state): State<AppState>,
    Json(request): Json<CreateSowingPlanRequest>,
) -> Result<impl IntoResponse, ApiErrorWrapper> {
    let plan = state.create_sowing_plan(request).map_err(ApiErrorWrapper::from)?;
    Ok((StatusCode::CREATED, Json(plan)))
}

async fn batch_create_sowing_plans(
    State(state): State<AppState>,
    Json(request): Json<core::BatchCreateSowingPlanRequest>,
) -> Result<impl IntoResponse, ApiErrorWrapper> {
    let result = state.batch_create_sowing_plans(request).map_err(ApiErrorWrapper::from)?;
    if result.has_conflicts {
        Ok((StatusCode::MULTI_STATUS, Json(result)))
    } else {
        Ok((StatusCode::CREATED, Json(result)))
    }
}

async fn get_sowing_plan(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<impl IntoResponse, ApiErrorWrapper> {
    let plan = state.get_sowing_plan(id).map_err(ApiErrorWrapper::from)?;
    Ok(Json(plan))
}

async fn get_all_sowing_plans(State(state): State<AppState>) -> impl IntoResponse {
    let plans = state.get_all_sowing_plans();
    Json(plans)
}

async fn get_sowing_plans_by_plot(
    State(state): State<AppState>,
    Path(plot_id): Path<Uuid>,
) -> impl IntoResponse {
    let plans = state.get_sowing_plans_by_plot(plot_id);
    Json(plans)
}

async fn health_check() -> impl IntoResponse {
    Json(json!({"status": "ok"}))
}

fn create_router(state: AppState) -> Router {
    let cors = CorsLayer::new()
        .allow_methods(Any)
        .allow_headers(Any)
        .allow_origin(Any);

    Router::new()
        .route("/health", get(health_check))
        .route("/crops", post(create_crop).get(get_all_crops))
        .route("/crops/:id", get(get_crop))
        .route("/regions", post(create_region).get(get_all_regions))
        .route("/regions/:id", get(get_region))
        .route("/plots", post(create_plot).get(get_all_plots))
        .route("/plots/:id", get(get_plot))
        .route("/plans", post(create_sowing_plan).get(get_all_sowing_plans))
        .route("/plans/batch", post(batch_create_sowing_plans))
        .route("/plans/:id", get(get_sowing_plan))
        .route("/plans/plot/:plot_id", get(get_sowing_plans_by_plot))
        .with_state(state)
        .layer(cors)
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_env_filter(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "server=info,tower_http=debug".into()),
        )
        .init();

    let args = Args::parse();

    let store = InMemoryStore::new();
    let service = AgriculturalService::new(store);
    let state = Arc::new(service);

    let app = create_router(state);

    let addr: SocketAddr = format!("{}:{}", args.host, args.port)
        .parse()
        .expect("Invalid address");

    tracing::info!("Server listening on {}", addr);

    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .expect("Server failed to start");
}
