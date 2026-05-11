use std::sync::Arc;
use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::{IntoResponse, Json, Response},
    routing::{get, post, put},
    Router,
};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

use grading_core::*;
use crate::state::AppState;
use std::fmt;

pub fn routes() -> Router<Arc<AppState>> {
    Router::new()
        .route("/health", get(health_check))
        .route("/teachers", post(create_teacher))
        .route("/teachers/:id", get(get_teacher))
        .route("/students", post(create_student))
        .route("/students/:id", get(get_student))
        .route("/exams", post(create_exam))
        .route("/exams/:id", get(get_exam))
        .route("/student-exams", post(submit_student_exam))
        .route("/student-exams/:id", get(get_student_exam))
        .route("/student-exams/:id/grade-objective", post(grade_objective))
        .route("/exams/:id/assign-tasks", post(assign_tasks))
        .route("/teachers/:id/tasks", get(get_teacher_tasks))
        .route("/tasks/:id/submit", post(submit_task))
        .route("/student-exams/:id/anomaly", post(mark_anomaly))
        .route("/exams/:id/publish", post(publish_exam))
        .route("/reviews", post(request_review))
        .route("/reviews/:id/process", post(process_review))
}

async fn health_check() -> Json<serde_json::Value> {
    Json(serde_json::json!({ "status": "ok" }))
}

#[derive(Debug, Deserialize)]
struct CreateTeacherRequest {
    name: String,
    mentor_students: Option<Vec<Uuid>>,
}

#[derive(Debug, Serialize)]
struct IdResponse {
    id: Uuid,
}

async fn create_teacher(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateTeacherRequest>,
) -> Result<Json<IdResponse>, AppError> {
    let teacher = if let Some(students) = req.mentor_students {
        Teacher::with_students(req.name, students)
    } else {
        Teacher::new(req.name)
    };
    let id = teacher.id;
    state.storage.add_teacher(teacher);
    Ok(Json(IdResponse { id }))
}

async fn get_teacher(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> Result<Json<Teacher>, AppError> {
    let teacher = state.storage.get_teacher(id)?;
    Ok(Json(teacher))
}

#[derive(Debug, Deserialize)]
struct CreateStudentRequest {
    name: String,
    mentor_id: Option<Uuid>,
}

async fn create_student(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateStudentRequest>,
) -> Result<Json<IdResponse>, AppError> {
    let student = if let Some(mentor_id) = req.mentor_id {
        Student::with_mentor(req.name, mentor_id)
    } else {
        Student::new(req.name)
    };
    let id = student.id;
    state.storage.add_student(student);
    Ok(Json(IdResponse { id }))
}

async fn get_student(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> Result<Json<Student>, AppError> {
    let student = state.storage.get_student(id)?;
    Ok(Json(student))
}

#[derive(Debug, Deserialize)]
struct QuestionCreate {
    question_type: String,
    content: String,
    options: Option<Vec<String>>,
    correct_answer: Option<Vec<String>>,
    max_score: u32,
    author_id: Uuid,
}

#[derive(Debug, Deserialize)]
struct CreateExamRequest {
    name: String,
    questions: Vec<QuestionCreate>,
}

async fn create_exam(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateExamRequest>,
) -> Result<Json<IdResponse>, AppError> {
    let mut questions = Vec::new();
    
    for q in req.questions {
        let question = match q.question_type.as_str() {
            "single_choice" => Question::new_single_choice(
                q.content,
                q.options.ok_or(AppError::bad_request("options required".into()))?,
                q.correct_answer
                    .ok_or(AppError::bad_request("correct_answer required".into()))?
                    .into_iter()
                    .next()
                    .ok_or(AppError::bad_request("correct_answer cannot be empty".into()))?,
                q.max_score,
                q.author_id,
            ),
            "multiple_choice" => Question::new_multiple_choice(
                q.content,
                q.options.ok_or(AppError::bad_request("options required".into()))?,
                q.correct_answer.ok_or(AppError::bad_request("correct_answer required".into()))?,
                q.max_score,
                q.author_id,
            ),
            "subjective" => Question::new_subjective(
                q.content,
                q.max_score,
                q.author_id,
            ),
            _ => return Err(AppError::bad_request("invalid question_type".into())),
        };
        questions.push(question);
    }
    
    let exam = ExamPaper::new(req.name, questions);
    let id = exam.id;
    state.storage.add_exam(exam);
    Ok(Json(IdResponse { id }))
}

async fn get_exam(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> Result<Json<ExamPaper>, AppError> {
    let exam = state.storage.get_exam(id)?;
    Ok(Json(exam))
}

#[derive(Debug, Deserialize)]
struct AnswerCreate {
    question_id: Uuid,
    content: Option<String>,
    selected_options: Option<Vec<String>>,
}

#[derive(Debug, Deserialize)]
struct SubmitExamRequest {
    student_id: Uuid,
    exam_id: Uuid,
    answers: Vec<AnswerCreate>,
}

async fn submit_student_exam(
    State(state): State<Arc<AppState>>,
    Json(req): Json<SubmitExamRequest>,
) -> Result<Json<IdResponse>, AppError> {
    let mut answers_map = std::collections::HashMap::new();
    
    for a in req.answers {
        let answer = Answer {
            question_id: a.question_id,
            content: a.content,
            selected_options: a.selected_options,
        };
        answers_map.insert(a.question_id, answer);
    }
    
    let id = state.storage.submit_student_exam(
        req.student_id,
        req.exam_id,
        answers_map,
    )?;
    
    Ok(Json(IdResponse { id }))
}

async fn get_student_exam(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> Result<Json<StudentExam>, AppError> {
    let exam = state.storage.get_student_exam(id)?;
    Ok(Json(exam))
}

async fn grade_objective(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> Result<Json<serde_json::Value>, AppError> {
    state.storage.grade_all_objective(id)?;
    Ok(Json(serde_json::json!({ "status": "success" })))
}

async fn assign_tasks(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> Result<Json<Vec<Uuid>>, AppError> {
    let task_ids = state.storage.assign_all_subjective_tasks(id)?;
    Ok(Json(task_ids))
}

async fn get_teacher_tasks(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> Json<Vec<GradingTask>> {
    let tasks = state.storage.get_teacher_tasks(id);
    Json(tasks)
}

#[derive(Debug, Deserialize)]
struct SubmitTaskRequest {
    score: u32,
    comments: Option<String>,
}

async fn submit_task(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
    Json(req): Json<SubmitTaskRequest>,
) -> Result<Json<serde_json::Value>, AppError> {
    state.storage.submit_grade(id, req.score, req.comments)?;
    Ok(Json(serde_json::json!({ "status": "success" })))
}

#[derive(Debug, Deserialize)]
struct MarkAnomalyRequest {
    anomaly_type: String,
    reported_by: Uuid,
    description: String,
    related_student_exam_id: Option<Uuid>,
}

async fn mark_anomaly(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
    Json(req): Json<MarkAnomalyRequest>,
) -> Result<Json<IdResponse>, AppError> {
    let anomaly_type = match req.anomaly_type.as_str() {
        "plagiarism" => AnomalyType::Plagiarism,
        "illegible" => AnomalyType::Illegible,
        _ => AnomalyType::Other,
    };
    
    let anomaly_id = state.storage.mark_anomaly(
        id,
        anomaly_type,
        req.reported_by,
        req.description,
        req.related_student_exam_id,
    )?;
    
    Ok(Json(IdResponse { id: anomaly_id }))
}

async fn publish_exam(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> Result<Json<serde_json::Value>, AppError> {
    state.storage.publish_scores(id)?;
    Ok(Json(serde_json::json!({ "status": "success" })))
}

#[derive(Debug, Deserialize)]
struct RequestReviewRequest {
    student_exam_id: Uuid,
    requested_by: Uuid,
    reason: String,
}

async fn request_review(
    State(state): State<Arc<AppState>>,
    Json(req): Json<RequestReviewRequest>,
) -> Result<Json<IdResponse>, AppError> {
    let id = state.storage.request_review(
        req.student_exam_id,
        req.requested_by,
        req.reason,
    )?;
    Ok(Json(IdResponse { id }))
}

async fn process_review(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> Result<Json<ReviewResult>, AppError> {
    let result = state.storage.process_review(id)?;
    Ok(Json(result))
}

#[derive(Debug)]
enum AppErrorKind {
    Grading(GradingError),
    BadRequest(String),
    Other(anyhow::Error),
}

#[derive(Debug)]
struct AppError(AppErrorKind);

impl fmt::Display for AppError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match &self.0 {
            AppErrorKind::Grading(e) => write!(f, "{}", e),
            AppErrorKind::BadRequest(msg) => write!(f, "{}", msg),
            AppErrorKind::Other(e) => write!(f, "{}", e),
        }
    }
}

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        tracing::error!("Error: {}", self);
        
        match self.0 {
            AppErrorKind::Grading(e) => {
                let (status, message) = match e {
                    GradingError::ExamNotFound(_) => (StatusCode::NOT_FOUND, e.to_string()),
                    GradingError::StudentExamNotFound(_) => (StatusCode::NOT_FOUND, e.to_string()),
                    GradingError::QuestionNotFound(_) => (StatusCode::NOT_FOUND, e.to_string()),
                    GradingError::TeacherNotFound(_) => (StatusCode::NOT_FOUND, e.to_string()),
                    GradingError::StudentNotFound(_) => (StatusCode::NOT_FOUND, e.to_string()),
                    GradingError::TaskNotFound(_) => (StatusCode::NOT_FOUND, e.to_string()),
                    GradingError::ReviewRequestNotFound(_) => (StatusCode::NOT_FOUND, e.to_string()),
                    GradingError::NoEligibleTeachers => (StatusCode::BAD_REQUEST, e.to_string()),
                    GradingError::ScoreOutOfRange(_, _) => (StatusCode::BAD_REQUEST, e.to_string()),
                    GradingError::TaskAlreadyCompleted => (StatusCode::BAD_REQUEST, e.to_string()),
                    GradingError::ExamAlreadyPublished => (StatusCode::BAD_REQUEST, e.to_string()),
                    GradingError::ExamNotPublished => (StatusCode::BAD_REQUEST, e.to_string()),
                    GradingError::InvalidStateTransition => (StatusCode::BAD_REQUEST, e.to_string()),
                    _ => (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()),
                };
                (status, Json(serde_json::json!({ "error": message }))).into_response()
            }
            AppErrorKind::BadRequest(msg) => {
                (StatusCode::BAD_REQUEST, Json(serde_json::json!({ "error": msg }))).into_response()
            }
            AppErrorKind::Other(e) => {
                tracing::error!("Internal error: {:?}", e);
                (
                    StatusCode::INTERNAL_SERVER_ERROR,
                    Json(serde_json::json!({ "error": "internal server error" })),
                ).into_response()
            }
        }
    }
}

impl From<GradingError> for AppError {
    fn from(err: GradingError) -> Self {
        Self(AppErrorKind::Grading(err))
    }
}

impl From<anyhow::Error> for AppError {
    fn from(err: anyhow::Error) -> Self {
        if let Some(grading_err) = err.downcast_ref::<GradingError>() {
            Self(AppErrorKind::Grading(grading_err.clone()))
        } else {
            Self(AppErrorKind::Other(err))
        }
    }
}

impl From<serde_json::Error> for AppError {
    fn from(err: serde_json::Error) -> Self {
        Self(AppErrorKind::Other(err.into()))
    }
}

impl AppError {
    fn bad_request(msg: String) -> Self {
        Self(AppErrorKind::BadRequest(msg))
    }
}
