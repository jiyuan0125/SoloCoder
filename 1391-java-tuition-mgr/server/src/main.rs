use std::sync::Arc;

use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    response::{IntoResponse, Response},
    routing::{get, post, delete},
    Json, Router,
};
use clap::Parser;
use serde::{Deserialize, Serialize};
use tower_http::cors::{Any, CorsLayer};
use uuid::Uuid;

use tuition_core::{
    errors::TuitionError,
    models::{Enrollment, Semester, Student, StudentOwingInfo, TuitionRecord, Course, CourseWithEnrollment},
    service::TuitionService,
    store::InMemoryStore,
};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "PORT", default_value_t = 3000)]
    port: u16,
}

type AppState = Arc<TuitionService>;

#[derive(Debug, Serialize, Deserialize)]
struct CreateStudentRequest {
    name: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateCourseRequest {
    name: String,
    capacity: u32,
    semester: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct SetTuitionRequest {
    student_id: Uuid,
    semester: String,
    amount: f64,
}

#[derive(Debug, Serialize, Deserialize)]
struct PayRequest {
    student_id: Uuid,
    semester: String,
    amount: f64,
}

#[derive(Debug, Serialize, Deserialize)]
struct EnrollRequest {
    student_id: Uuid,
    course_id: Uuid,
    semester: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct SetEnrollmentPeriodRequest {
    semester: String,
    start_time: chrono::DateTime<chrono::Utc>,
    end_time: chrono::DateTime<chrono::Utc>,
}

#[derive(Debug, Serialize)]
struct ErrorResponse {
    error: String,
}

struct AppError(TuitionError);

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let status = match &self.0 {
            TuitionError::StudentNotFound => StatusCode::NOT_FOUND,
            TuitionError::CourseNotFound => StatusCode::NOT_FOUND,
            TuitionError::TuitionRecordNotFound(_) => StatusCode::NOT_FOUND,
            TuitionError::EnrollmentPeriodNotSet(_) => StatusCode::BAD_REQUEST,
            TuitionError::EnrollmentPeriodNotActive(_) => StatusCode::FORBIDDEN,
            TuitionError::CourseFull(_) => StatusCode::CONFLICT,
            TuitionError::AlreadyEnrolled(_) => StatusCode::CONFLICT,
            TuitionError::NotEnrolled(_) => StatusCode::NOT_FOUND,
            TuitionError::UnpaidTuition(_) => StatusCode::FORBIDDEN,
            TuitionError::InvalidPaymentAmount(_) => StatusCode::BAD_REQUEST,
            TuitionError::FirstInstallmentAlreadyPaid => StatusCode::CONFLICT,
            TuitionError::SecondInstallmentAlreadyPaid => StatusCode::CONFLICT,
            TuitionError::MustPayFirstInstallmentFirst => StatusCode::BAD_REQUEST,
            TuitionError::Internal(_) => StatusCode::INTERNAL_SERVER_ERROR,
        };

        let body = Json(ErrorResponse {
            error: self.0.to_string(),
        });

        (status, body).into_response()
    }
}

impl From<TuitionError> for AppError {
    fn from(err: TuitionError) -> Self {
        AppError(err)
    }
}

async fn create_student(
    State(service): State<AppState>,
    Json(req): Json<CreateStudentRequest>,
) -> Json<Student> {
    let student = service.create_student(req.name);
    Json(student)
}

async fn list_students(State(service): State<AppState>) -> Json<Vec<Student>> {
    let students = service.list_students();
    Json(students)
}

async fn get_student(
    State(service): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<Student>, AppError> {
    service
        .get_student(id)
        .map(Json)
        .ok_or(TuitionError::StudentNotFound.into())
}

async fn create_course(
    State(service): State<AppState>,
    Json(req): Json<CreateCourseRequest>,
) -> Json<Course> {
    let course = service.create_course(req.name, req.capacity, Semester(req.semester));
    Json(course)
}

async fn list_courses(
    State(service): State<AppState>,
    Query(params): Query<std::collections::HashMap<String, String>>,
) -> Json<Vec<CourseWithEnrollment>> {
    let semester = params.get("semester").cloned().unwrap_or_default();
    let courses = service.list_courses(&Semester(semester));
    Json(courses)
}

async fn get_course(
    State(service): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<Course>, AppError> {
    service
        .get_course(id)
        .map(Json)
        .ok_or(TuitionError::CourseNotFound.into())
}

async fn set_tuition(
    State(service): State<AppState>,
    Json(req): Json<SetTuitionRequest>,
) -> Result<Json<TuitionRecord>, AppError> {
    let record = service.set_tuition(req.student_id, Semester(req.semester), req.amount)?;
    Ok(Json(record))
}

async fn get_tuition(
    State(service): State<AppState>,
    Path((student_id, semester)): Path<(Uuid, String)>,
) -> Result<Json<TuitionRecord>, AppError> {
    let record = service.get_tuition_record(student_id, &Semester(semester))?;
    Ok(Json(record))
}

async fn pay_first_installment(
    State(service): State<AppState>,
    Json(req): Json<PayRequest>,
) -> Result<Json<TuitionRecord>, AppError> {
    let record = service.pay_first_installment(req.student_id, &Semester(req.semester), req.amount)?;
    Ok(Json(record))
}

async fn pay_second_installment(
    State(service): State<AppState>,
    Json(req): Json<PayRequest>,
) -> Result<Json<TuitionRecord>, AppError> {
    let record = service.pay_second_installment(req.student_id, &Semester(req.semester), req.amount)?;
    Ok(Json(record))
}

async fn set_enrollment_period(
    State(service): State<AppState>,
    Json(req): Json<SetEnrollmentPeriodRequest>,
) {
    service.set_enrollment_period(
        Semester(req.semester),
        req.start_time,
        req.end_time,
    );
}

async fn enroll(
    State(service): State<AppState>,
    Json(req): Json<EnrollRequest>,
) -> Result<Json<Enrollment>, AppError> {
    let enrollment = service.enroll(req.student_id, req.course_id, &Semester(req.semester))?;
    Ok(Json(enrollment))
}

async fn drop_course(
    State(service): State<AppState>,
    Path((student_id, course_id, semester)): Path<(Uuid, Uuid, String)>,
) -> Result<StatusCode, AppError> {
    service.drop_course(student_id, course_id, &Semester(semester))?;
    Ok(StatusCode::NO_CONTENT)
}

async fn list_student_enrollments(
    State(service): State<AppState>,
    Path((student_id, semester)): Path<(Uuid, String)>,
) -> Result<Json<Vec<Course>>, AppError> {
    let courses = service.list_student_enrollments(student_id, &Semester(semester))?;
    Ok(Json(courses))
}

async fn list_owing_students(
    State(service): State<AppState>,
    Path(semester): Path<String>,
) -> Json<Vec<StudentOwingInfo>> {
    let students = service.list_owing_students(&Semester(semester));
    Json(students)
}

#[tokio::main]
async fn main() {
    let args = Args::parse();
    
    let store = Arc::new(InMemoryStore::new());
    let service = Arc::new(TuitionService::new(store));

    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    let app = Router::new()
        .route("/students", post(create_student).get(list_students))
        .route("/students/:id", get(get_student))
        .route("/courses", post(create_course).get(list_courses))
        .route("/courses/:id", get(get_course))
        .route("/tuition", post(set_tuition))
        .route("/tuition/:student_id/:semester", get(get_tuition))
        .route("/pay/first", post(pay_first_installment))
        .route("/pay/second", post(pay_second_installment))
        .route("/enrollment-period", post(set_enrollment_period))
        .route("/enroll", post(enroll))
        .route("/enroll/:student_id/:course_id/:semester", delete(drop_course))
        .route("/enrollments/:student_id/:semester", get(list_student_enrollments))
        .route("/owing/:semester", get(list_owing_students))
        .layer(cors)
        .with_state(service);

    let addr = std::net::SocketAddr::from(([127, 0, 0, 1], args.port));
    println!("Server listening on {}", addr);
    
    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}
