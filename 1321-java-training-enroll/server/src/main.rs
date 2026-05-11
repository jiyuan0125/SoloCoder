use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post, put, delete},
    Json, Router,
};
use clap::Parser;
use app_core::models::*;
use app_core::service::TrainingService;
use serde::{Deserialize, Serialize};
use std::sync::{Arc, Mutex};
use tower_http::cors::{Any, CorsLayer};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "PORT", default_value_t = 3000)]
    port: u16,

    #[arg(short, long, env = "HOST", default_value = "127.0.0.1")]
    host: String,
}

type AppState = Arc<Mutex<TrainingService>>;

#[derive(Debug, Deserialize)]
struct CreateCourseRequest {
    id: String,
    name: String,
    capacity: u32,
    registration_start: u64,
    registration_end: u64,
}

#[derive(Debug, Deserialize)]
struct RegisterRequest {
    employee_id: String,
}

#[derive(Debug, Serialize)]
struct ApiResponse<T> {
    success: bool,
    data: Option<T>,
    error: Option<String>,
}

impl<T> ApiResponse<T> {
    fn success(data: T) -> Self {
        ApiResponse {
            success: true,
            data: Some(data),
            error: None,
        }
    }

    fn error(message: String) -> Self {
        ApiResponse {
            success: false,
            data: None,
            error: Some(message),
        }
    }
}

fn get_current_time() -> u64 {
    use std::time::{SystemTime, UNIX_EPOCH};
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .expect("Time went backwards")
        .as_secs()
}

async fn list_courses(State(state): State<AppState>) -> impl IntoResponse {
    let mut service = state.lock().unwrap();
    service.set_time(get_current_time());
    let courses = service.list_courses();
    (StatusCode::OK, Json(ApiResponse::success(courses)))
}

async fn get_course(State(state): State<AppState>, Path(course_id): Path<String>) -> impl IntoResponse {
    let mut service = state.lock().unwrap();
    service.set_time(get_current_time());
    
    match service.get_course(&CourseId(course_id)) {
        Some(course) => (StatusCode::OK, Json(ApiResponse::success(course.clone()))),
        None => (StatusCode::NOT_FOUND, Json(ApiResponse::error("Course not found".to_string()))),
    }
}

async fn create_course(
    State(state): State<AppState>,
    Json(req): Json<CreateCourseRequest>,
) -> impl IntoResponse {
    let mut service = state.lock().unwrap();
    service.set_time(get_current_time());
    
    match service.create_course(
        CourseId(req.id),
        req.name,
        req.capacity,
        req.registration_start,
        req.registration_end,
    ) {
        Ok(course) => (StatusCode::CREATED, Json(ApiResponse::success(course))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(ApiResponse::error(e))),
    }
}

async fn register_for_course(
    State(state): State<AppState>,
    Path(course_id): Path<String>,
    Json(req): Json<RegisterRequest>,
) -> impl IntoResponse {
    let mut service = state.lock().unwrap();
    service.set_time(get_current_time());
    service.check_expired_confirmations();
    
    match service.register(EmployeeId(req.employee_id), CourseId(course_id)) {
        Ok(registration) => (StatusCode::OK, Json(ApiResponse::success(registration))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(ApiResponse::error(e))),
    }
}

async fn cancel_registration(
    State(state): State<AppState>,
    Path((course_id, employee_id)): Path<(String, String)>,
) -> impl IntoResponse {
    let mut service = state.lock().unwrap();
    service.set_time(get_current_time());
    service.check_expired_confirmations();
    
    match service.cancel_registration(EmployeeId(employee_id), CourseId(course_id)) {
        Ok(_) => (StatusCode::OK, Json(ApiResponse::success(()))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(ApiResponse::error(e))),
    }
}

async fn confirm_registration(
    State(state): State<AppState>,
    Path((course_id, employee_id)): Path<(String, String)>,
) -> impl IntoResponse {
    let mut service = state.lock().unwrap();
    service.set_time(get_current_time());
    service.check_expired_confirmations();
    
    match service.confirm_registration(EmployeeId(employee_id), CourseId(course_id)) {
        Ok(registration) => (StatusCode::OK, Json(ApiResponse::success(registration))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(ApiResponse::error(e))),
    }
}

async fn get_course_registrations(
    State(state): State<AppState>,
    Path(course_id): Path<String>,
) -> impl IntoResponse {
    let mut service = state.lock().unwrap();
    service.set_time(get_current_time());
    service.check_expired_confirmations();
    
    let registrations = service.get_registrations(&CourseId(course_id));
    (StatusCode::OK, Json(ApiResponse::success(registrations)))
}

async fn get_employee_registrations(
    State(state): State<AppState>,
    Path(employee_id): Path<String>,
) -> impl IntoResponse {
    let mut service = state.lock().unwrap();
    service.set_time(get_current_time());
    service.check_expired_confirmations();
    
    let registrations = service.get_employee_registrations(&EmployeeId(employee_id));
    (StatusCode::OK, Json(ApiResponse::success(registrations)))
}

async fn get_waiting_list(
    State(state): State<AppState>,
    Path(course_id): Path<String>,
) -> impl IntoResponse {
    let mut service = state.lock().unwrap();
    service.set_time(get_current_time());
    service.check_expired_confirmations();
    
    let waiting_list = service.get_waiting_list(&CourseId(course_id));
    (StatusCode::OK, Json(ApiResponse::success(waiting_list)))
}

async fn get_notifications(
    State(state): State<AppState>,
    Path(employee_id): Path<String>,
) -> impl IntoResponse {
    let mut service = state.lock().unwrap();
    service.set_time(get_current_time());
    
    let notifications = service.get_notifications(&EmployeeId(employee_id));
    (StatusCode::OK, Json(ApiResponse::success(notifications)))
}

async fn mark_notification_read(
    State(state): State<AppState>,
    Path(notification_id): Path<String>,
) -> impl IntoResponse {
    let mut service = state.lock().unwrap();
    service.set_time(get_current_time());
    
    match service.mark_notification_read(&notification_id) {
        Ok(_) => (StatusCode::OK, Json(ApiResponse::success(()))),
        Err(e) => (StatusCode::NOT_FOUND, Json(ApiResponse::error(e))),
    }
}

#[tokio::main]
async fn main() {
    let args = Args::parse();
    
    let service = Arc::new(Mutex::new(TrainingService::new()));
    
    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    let app = Router::new()
        .route("/courses", get(list_courses).post(create_course))
        .route("/courses/:id", get(get_course))
        .route("/courses/:id/register", post(register_for_course))
        .route("/courses/:id/registrations", get(get_course_registrations))
        .route("/courses/:id/waiting-list", get(get_waiting_list))
        .route("/courses/:course_id/registrations/:employee_id", delete(cancel_registration))
        .route("/courses/:course_id/registrations/:employee_id/confirm", put(confirm_registration))
        .route("/employees/:id/registrations", get(get_employee_registrations))
        .route("/employees/:id/notifications", get(get_notifications))
        .route("/notifications/:id/read", put(mark_notification_read))
        .layer(cors)
        .with_state(service);

    let addr = format!("{}:{}", args.host, args.port);
    println!("Server listening on http://{}", addr);
    
    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
