use std::sync::Arc;

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::{IntoResponse, Json, Response},
    routing::{get, post},
    Router,
};
use booking_core::{
    BookingError, BookingService, CreateBookingRequest, InMemoryStore, SubmitFeedbackRequest,
};
use clap::Parser;
use serde::Serialize;
use tower_http::cors::{Any, CorsLayer};
use uuid::Uuid;

#[derive(Parser, Debug, Clone)]
struct Config {
    #[arg(short, long, env = "PORT", default_value = "8080")]
    port: u16,
    #[arg(long, env = "HOST", default_value = "127.0.0.1")]
    host: String,
}

struct AppState {
    service: BookingService,
}

#[derive(Serialize)]
struct ErrorResponse {
    error: String,
}

struct AppError(BookingError);

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let (status, error) = match self.0 {
            BookingError::CarModelNotFound => (StatusCode::NOT_FOUND, self.0.to_string()),
            BookingError::NoAvailableCar => (StatusCode::BAD_REQUEST, self.0.to_string()),
            BookingError::TimeSlotUnavailable => (StatusCode::BAD_REQUEST, self.0.to_string()),
            BookingError::AlreadyBookedToday => (StatusCode::BAD_REQUEST, self.0.to_string()),
            BookingError::BookingNotFound => (StatusCode::NOT_FOUND, self.0.to_string()),
            BookingError::AdvisorNotFound => (StatusCode::NOT_FOUND, self.0.to_string()),
            BookingError::InvalidStateTransition => (StatusCode::BAD_REQUEST, self.0.to_string()),
            BookingError::BookingExpired => (StatusCode::BAD_REQUEST, self.0.to_string()),
            BookingError::Internal(msg) => (StatusCode::INTERNAL_SERVER_ERROR, msg),
        };
        (status, Json(ErrorResponse { error })).into_response()
    }
}

async fn list_models(State(state): State<Arc<AppState>>) -> Json<Vec<booking_core::CarModel>> {
    Json(state.service.store().list_car_models().await)
}

async fn list_cars(State(state): State<Arc<AppState>>) -> Json<Vec<booking_core::Car>> {
    Json(state.service.store().list_cars().await)
}

async fn list_customers(State(state): State<Arc<AppState>>) -> Json<Vec<booking_core::Customer>> {
    Json(state.service.store().list_customers().await)
}

async fn list_advisors(State(state): State<Arc<AppState>>) -> Json<Vec<booking_core::SalesAdvisor>> {
    Json(state.service.store().list_advisors().await)
}

async fn list_bookings(State(state): State<Arc<AppState>>) -> Json<Vec<booking_core::Booking>> {
    Json(state.service.store().list_bookings().await)
}

async fn get_booking(State(state): State<Arc<AppState>>, Path(id): Path<Uuid>) -> Response {
    match state.service.store().get_booking(&id).await {
        Some(booking) => Json(booking).into_response(),
        None => StatusCode::NOT_FOUND.into_response(),
    }
}

async fn create_booking(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateBookingRequest>,
) -> Response {
    match state.service.create_booking(req).await {
        Ok(booking) => (StatusCode::CREATED, Json(booking)).into_response(),
        Err(e) => AppError(e).into_response(),
    }
}

async fn start_booking(State(state): State<Arc<AppState>>, Path(id): Path<Uuid>) -> Response {
    match state.service.start_booking(&id).await {
        Ok(booking) => Json(booking).into_response(),
        Err(e) => AppError(e).into_response(),
    }
}

async fn complete_booking(State(state): State<Arc<AppState>>, Path(id): Path<Uuid>) -> Response {
    match state.service.complete_booking(&id).await {
        Ok(booking) => Json(booking).into_response(),
        Err(e) => AppError(e).into_response(),
    }
}

async fn cancel_booking(State(state): State<Arc<AppState>>, Path(id): Path<Uuid>) -> Response {
    match state.service.cancel_booking(&id).await {
        Ok(booking) => Json(booking).into_response(),
        Err(e) => AppError(e).into_response(),
    }
}

async fn submit_feedback(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
    Json(req): Json<SubmitFeedbackRequest>,
) -> Response {
    match state.service.submit_feedback(&id, req).await {
        Ok(feedback) => Json(feedback).into_response(),
        Err(e) => AppError(e).into_response(),
    }
}

#[tokio::main]
async fn main() {
    let config = Config::parse();

    let store = InMemoryStore::new();
    store.seed_sample_data().await;

    let service = BookingService::new(store);

    let app_state = Arc::new(AppState { service });

    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    let app = Router::new()
        .route("/api/models", get(list_models))
        .route("/api/cars", get(list_cars))
        .route("/api/customers", get(list_customers))
        .route("/api/advisors", get(list_advisors))
        .route("/api/bookings", get(list_bookings).post(create_booking))
        .route("/api/bookings/:id", get(get_booking))
        .route("/api/bookings/:id/start", post(start_booking))
        .route("/api/bookings/:id/complete", post(complete_booking))
        .route("/api/bookings/:id/cancel", post(cancel_booking))
        .route("/api/bookings/:id/feedback", post(submit_feedback))
        .layer(cors)
        .with_state(app_state);

    let addr = format!("{}:{}", config.host, config.port);
    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    println!("Server running on http://{}", addr);
    axum::serve(listener, app).await.unwrap();
}
