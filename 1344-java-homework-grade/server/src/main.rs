use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use clap::Parser;
use hw_core::{
    AppError, AppState, Answer, Assignment, Course, CourseStatistics, Grading,
    Question, QuestionType, Role, ScoreDistribution, StudentTrend, Submission, TeacherWorkload,
    User,
};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::net::SocketAddr;
use tower_http::cors::{Any, CorsLayer};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "PORT", default_value_t = 8080)]
    port: u16,
}

#[derive(Clone)]
struct ServerState {
    app_state: AppState,
}

struct AppErrorResponse(AppError);

impl IntoResponse for AppErrorResponse {
    fn into_response(self) -> axum::response::Response {
        let status = match &self.0 {
            AppError::UserNotFound(_) => StatusCode::NOT_FOUND,
            AppError::CourseNotFound(_) => StatusCode::NOT_FOUND,
            AppError::AssignmentNotFound(_) => StatusCode::NOT_FOUND,
            AppError::QuestionNotFound(_) => StatusCode::NOT_FOUND,
            AppError::SubmissionNotFound(_) => StatusCode::NOT_FOUND,
            AppError::GradingNotFound(_) => StatusCode::NOT_FOUND,
            AppError::NotTeacher => StatusCode::FORBIDDEN,
            AppError::NotStudent => StatusCode::FORBIDDEN,
            AppError::NotEnrolled => StatusCode::FORBIDDEN,
            AppError::NotCourseTeacher => StatusCode::FORBIDDEN,
            AppError::ScoreExceedsMax { .. } => StatusCode::BAD_REQUEST,
            AppError::AlreadySubmitted => StatusCode::CONFLICT,
            AppError::AlreadyGraded => StatusCode::CONFLICT,
            AppError::AlreadyAppealed => StatusCode::CONFLICT,
            AppError::MustBeDifferentTeacher => StatusCode::BAD_REQUEST,
            AppError::CannotModifyAfterSubmission => StatusCode::BAD_REQUEST,
            AppError::CannotModifyAfterGrading => StatusCode::BAD_REQUEST,
            AppError::InvalidRole => StatusCode::BAD_REQUEST,
            AppError::Internal(_) => StatusCode::INTERNAL_SERVER_ERROR,
        };

        let body = Json(serde_json::json!({
            "error": self.0.to_string()
        }));

        (status, body).into_response()
    }
}

impl From<AppError> for AppErrorResponse {
    fn from(err: AppError) -> Self {
        AppErrorResponse(err)
    }
}

#[derive(Debug, Deserialize)]
struct CreateUserRequest {
    name: String,
    role: String,
}

#[derive(Debug, Deserialize)]
struct CreateCourseRequest {
    name: String,
    teacher_ids: Vec<String>,
}

#[derive(Debug, Deserialize)]
struct AddStudentRequest {
    student_id: String,
}

#[derive(Debug, Deserialize)]
struct CreateQuestion {
    question_type: String,
    title: String,
    max_score: u32,
    answer: Option<String>,
}

#[derive(Debug, Deserialize)]
struct CreateAssignmentRequest {
    title: String,
    teacher_id: String,
    questions: Vec<CreateQuestion>,
}

#[derive(Debug, Deserialize)]
struct SubmitAssignmentRequest {
    student_id: String,
    answers: Vec<AnswerDto>,
}

#[derive(Debug, Deserialize, Serialize, Clone)]
struct AnswerDto {
    question_id: String,
    answer: String,
}

#[derive(Debug, Deserialize)]
struct GradeSubjectiveRequest {
    teacher_id: String,
    scores: HashMap<String, u32>,
}

#[derive(Debug, Deserialize)]
struct AppealRequest {
    student_id: String,
}

#[derive(Debug, Serialize)]
struct SubmissionDetail {
    submission: Submission,
    gradings: Vec<Grading>,
    final_score: Option<u32>,
    max_score: u32,
}

fn parse_role(role: &str) -> Result<Role, AppError> {
    match role.to_lowercase().as_str() {
        "student" => Ok(Role::Student),
        "teacher" => Ok(Role::Teacher),
        _ => Err(AppError::InvalidRole),
    }
}

fn parse_question_type(qt: &str) -> Result<QuestionType, AppError> {
    match qt.to_lowercase().as_str() {
        "objective" => Ok(QuestionType::Objective),
        "subjective" => Ok(QuestionType::Subjective),
        _ => Err(AppError::InvalidRole),
    }
}

async fn list_users(State(state): State<ServerState>) -> Result<Json<Vec<User>>, AppErrorResponse> {
    Ok(Json(state.app_state.list_users()))
}

async fn get_user(
    State(state): State<ServerState>,
    Path(id): Path<String>,
) -> Result<Json<User>, AppErrorResponse> {
    Ok(Json(state.app_state.get_user(&id)?))
}

async fn create_user(
    State(state): State<ServerState>,
    Json(req): Json<CreateUserRequest>,
) -> Result<Json<User>, AppErrorResponse> {
    let role = parse_role(&req.role)?;
    Ok(Json(state.app_state.create_user(&req.name, role)))
}

async fn list_courses(
    State(state): State<ServerState>,
) -> Result<Json<Vec<Course>>, AppErrorResponse> {
    Ok(Json(state.app_state.list_courses()))
}

async fn get_course(
    State(state): State<ServerState>,
    Path(id): Path<String>,
) -> Result<Json<Course>, AppErrorResponse> {
    Ok(Json(state.app_state.get_course(&id)?))
}

async fn create_course(
    State(state): State<ServerState>,
    Json(req): Json<CreateCourseRequest>,
) -> Result<Json<Course>, AppErrorResponse> {
    let course = state.app_state.create_course(&req.name, req.teacher_ids)?;
    Ok(Json(course))
}

async fn add_student_to_course(
    State(state): State<ServerState>,
    Path(course_id): Path<String>,
    Json(req): Json<AddStudentRequest>,
) -> Result<StatusCode, AppErrorResponse> {
    state
        .app_state
        .add_student_to_course(&course_id, &req.student_id)?;
    Ok(StatusCode::OK)
}

async fn list_assignments(
    State(state): State<ServerState>,
    Path(course_id): Path<String>,
) -> Result<Json<Vec<Assignment>>, AppErrorResponse> {
    Ok(Json(state.app_state.list_assignments(&course_id)?))
}

async fn get_assignment(
    State(state): State<ServerState>,
    Path(id): Path<String>,
) -> Result<Json<Assignment>, AppErrorResponse> {
    Ok(Json(state.app_state.get_assignment(&id)?))
}

async fn create_assignment(
    State(state): State<ServerState>,
    Path(course_id): Path<String>,
    Json(req): Json<CreateAssignmentRequest>,
) -> Result<Json<Assignment>, AppErrorResponse> {
    let mut questions = Vec::new();
    for (i, q) in req.questions.into_iter().enumerate() {
        questions.push(Question {
            id: format!("q-{}", i),
            question_type: parse_question_type(&q.question_type)?,
            title: q.title,
            max_score: q.max_score,
            answer: q.answer,
        });
    }

    let assignment =
        state
            .app_state
            .create_assignment(&course_id, &req.title, questions, &req.teacher_id)?;
    Ok(Json(assignment))
}

async fn list_submissions(
    State(state): State<ServerState>,
    Path(assignment_id): Path<String>,
) -> Result<Json<Vec<Submission>>, AppErrorResponse> {
    Ok(Json(state.app_state.list_submissions(&assignment_id)?))
}

async fn get_submission(
    State(state): State<ServerState>,
    Path(id): Path<String>,
) -> Result<Json<SubmissionDetail>, AppErrorResponse> {
    let submission = state.app_state.get_submission(&id)?;
    let assignment = state.app_state.get_assignment(&submission.assignment_id)?;
    let gradings = state.app_state.list_gradings(&id)?;
    let final_score = submission.get_final_score(&gradings);
    let max_score = submission.get_max_score(&assignment);

    Ok(Json(SubmissionDetail {
        submission,
        gradings,
        final_score,
        max_score,
    }))
}

async fn submit_assignment(
    State(state): State<ServerState>,
    Path(assignment_id): Path<String>,
    Json(req): Json<SubmitAssignmentRequest>,
) -> Result<Json<Submission>, AppErrorResponse> {
    let answers: Vec<Answer> = req
        .answers
        .into_iter()
        .map(|a| Answer {
            question_id: a.question_id,
            answer: a.answer,
        })
        .collect();

    let submission = state
        .app_state
        .submit_assignment(&assignment_id, &req.student_id, answers)?;
    Ok(Json(submission))
}

async fn grade_subjective(
    State(state): State<ServerState>,
    Path(submission_id): Path<String>,
    Json(req): Json<GradeSubjectiveRequest>,
) -> Result<Json<Grading>, AppErrorResponse> {
    let scores: HashMap<String, u32> = req.scores;
    let grading = state
        .app_state
        .grade_subjective(&submission_id, &req.teacher_id, scores)?;
    Ok(Json(grading))
}

async fn list_gradings(
    State(state): State<ServerState>,
    Path(submission_id): Path<String>,
) -> Result<Json<Vec<Grading>>, AppErrorResponse> {
    Ok(Json(state.app_state.list_gradings(&submission_id)?))
}

async fn appeal_submission(
    State(state): State<ServerState>,
    Path(submission_id): Path<String>,
    Json(req): Json<AppealRequest>,
) -> Result<StatusCode, AppErrorResponse> {
    state
        .app_state
        .appeal_submission(&submission_id, &req.student_id)?;
    Ok(StatusCode::OK)
}

async fn grade_appeal(
    State(state): State<ServerState>,
    Path(submission_id): Path<String>,
    Json(req): Json<GradeSubjectiveRequest>,
) -> Result<Json<Grading>, AppErrorResponse> {
    let scores: HashMap<String, u32> = req.scores;
    let grading = state
        .app_state
        .grade_appeal(&submission_id, &req.teacher_id, scores)?;
    Ok(Json(grading))
}

async fn course_statistics(
    State(state): State<ServerState>,
    Path(course_id): Path<String>,
) -> Result<Json<CourseStatistics>, AppErrorResponse> {
    Ok(Json(state.app_state.get_course_statistics(&course_id)?))
}

async fn score_distribution(
    State(state): State<ServerState>,
    Path(course_id): Path<String>,
) -> Result<Json<ScoreDistribution>, AppErrorResponse> {
    Ok(Json(state.app_state.get_score_distribution(&course_id)?))
}

async fn student_trends(
    State(state): State<ServerState>,
    Path(course_id): Path<String>,
) -> Result<Json<Vec<StudentTrend>>, AppErrorResponse> {
    Ok(Json(state.app_state.get_student_trends(&course_id)?))
}

async fn teacher_workload(
    State(state): State<ServerState>,
) -> Result<Json<Vec<TeacherWorkload>>, AppErrorResponse> {
    Ok(Json(state.app_state.get_teacher_workload()?))
}

fn create_router(state: ServerState) -> Router {
    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    Router::new()
        .route("/users", get(list_users).post(create_user))
        .route("/users/:id", get(get_user))
        .route("/courses", get(list_courses).post(create_course))
        .route("/courses/:id", get(get_course))
        .route("/courses/:id/students", post(add_student_to_course))
        .route(
            "/courses/:id/assignments",
            get(list_assignments).post(create_assignment),
        )
        .route("/assignments/:id", get(get_assignment))
        .route(
            "/assignments/:id/submissions",
            get(list_submissions).post(submit_assignment),
        )
        .route("/submissions/:id", get(get_submission))
        .route("/submissions/:id/gradings", get(list_gradings))
        .route("/submissions/:id/grade", post(grade_subjective))
        .route("/submissions/:id/appeal", post(appeal_submission))
        .route("/submissions/:id/grade-appeal", post(grade_appeal))
        .route("/courses/:id/statistics", get(course_statistics))
        .route("/courses/:id/distribution", get(score_distribution))
        .route("/courses/:id/trends", get(student_trends))
        .route("/teachers/workload", get(teacher_workload))
        .layer(cors)
        .with_state(state)
}

#[tokio::main]
async fn main() {
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "server=info,tower_http=debug".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    let args = Args::parse();

    let app_state = AppState::new();
    let state = ServerState { app_state };
    let app = create_router(state);

    let addr = SocketAddr::from(([0, 0, 0, 0], args.port));
    tracing::info!("listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
