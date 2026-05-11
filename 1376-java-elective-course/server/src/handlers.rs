use std::sync::Arc;
use axum::{
    extract::{Path, State},
    http::StatusCode,
    Json,
};
use serde::Deserialize;
use course_core::{
    CourseSystemError,
    CourseService,
    CourseId,
    StudentId,
};

#[derive(Deserialize)]
pub struct EnrollRequest {
    pub student_id: String,
    pub course_id: String,
}

#[derive(Deserialize)]
pub struct WithdrawRequest {
    pub student_id: String,
    pub course_id: String,
}

fn map_error(err: CourseSystemError) -> (StatusCode, String) {
    match err {
        CourseSystemError::CourseNotFound(msg) => (StatusCode::NOT_FOUND, msg),
        CourseSystemError::StudentNotFound(msg) => (StatusCode::NOT_FOUND, msg),
        CourseSystemError::PrerequisiteNotSatisfied(msg) => (StatusCode::BAD_REQUEST, msg),
        CourseSystemError::CourseFull => (StatusCode::BAD_REQUEST, "课程已满".to_string()),
        CourseSystemError::CreditLimitExceeded { current, max, adding } => (
            StatusCode::BAD_REQUEST,
            format!("学分上限{}已达，当前{}，添加{}将超出", max, current, adding),
        ),
        CourseSystemError::AlreadyEnrolled => (StatusCode::BAD_REQUEST, "已选过此课程".to_string()),
        CourseSystemError::NotEnrolled => (StatusCode::BAD_REQUEST, "未选过此课程".to_string()),
        CourseSystemError::EnrollmentEnded => (StatusCode::BAD_REQUEST, "选课已结束".to_string()),
        CourseSystemError::ConcurrentError(msg) => (StatusCode::INTERNAL_SERVER_ERROR, msg),
    }
}

pub async fn list_courses(
    State(service): State<Arc<CourseService>>,
) -> Json<Vec<serde_json::Value>> {
    let courses = service.list_courses();
    let result: Vec<serde_json::Value> = courses
        .into_iter()
        .map(|c| {
            serde_json::json!({
                "id": c.id.0,
                "name": c.name,
                "credits": c.credits,
                "capacity": c.capacity,
                "prerequisites": c.prerequisites.into_iter().map(|p| p.0).collect::<Vec<_>>(),
            })
        })
        .collect();
    Json(result)
}

pub async fn get_course(
    State(service): State<Arc<CourseService>>,
    Path(id): Path<String>,
) -> Result<Json<serde_json::Value>, (StatusCode, Json<serde_json::Value>)> {
    match service.get_course(&CourseId(id.clone())) {
        Some(c) => Ok(Json(serde_json::json!({
            "id": c.id.0,
            "name": c.name,
            "credits": c.credits,
            "capacity": c.capacity,
            "prerequisites": c.prerequisites.into_iter().map(|p| p.0).collect::<Vec<_>>(),
        }))),
        None => {
            let (code, msg) = map_error(CourseSystemError::CourseNotFound(id));
            Err((code, Json(serde_json::json!({ "error": msg }))))
        }
    }
}

pub async fn list_students(
    State(service): State<Arc<CourseService>>,
) -> Json<Vec<serde_json::Value>> {
    let students = service.list_students();
    let result: Vec<serde_json::Value> = students
        .into_iter()
        .map(|s| {
            serde_json::json!({
                "id": s.id.0,
                "name": s.name,
                "completed_courses": s.completed_courses.into_iter().map(|c| {
                    serde_json::json!({
                        "course_id": c.course_id.0,
                        "score": c.score,
                    })
                }).collect::<Vec<_>>(),
            })
        })
        .collect();
    Json(result)
}

pub async fn get_student(
    State(service): State<Arc<CourseService>>,
    Path(id): Path<String>,
) -> Result<Json<serde_json::Value>, (StatusCode, Json<serde_json::Value>)> {
    match service.get_student(&StudentId(id.clone())) {
        Some(s) => Ok(Json(serde_json::json!({
            "id": s.id.0,
            "name": s.name,
            "completed_courses": s.completed_courses.into_iter().map(|c| {
                serde_json::json!({
                    "course_id": c.course_id.0,
                    "score": c.score,
                })
            }).collect::<Vec<_>>(),
        }))),
        None => {
            let (code, msg) = map_error(CourseSystemError::StudentNotFound(id));
            Err((code, Json(serde_json::json!({ "error": msg }))))
        }
    }
}

pub async fn enroll(
    State(service): State<Arc<CourseService>>,
    Json(req): Json<EnrollRequest>,
) -> Result<Json<serde_json::Value>, (StatusCode, Json<serde_json::Value>)> {
    let student_id = StudentId(req.student_id);
    let course_id = CourseId(req.course_id);

    match service.enroll(&student_id, &course_id) {
        Ok(()) => Ok(Json(serde_json::json!({
            "success": true,
            "message": "选课成功",
        }))),
        Err(err) => {
            let (code, msg) = map_error(err);
            Err((code, Json(serde_json::json!({ "error": msg }))))
        }
    }
}

pub async fn withdraw(
    State(service): State<Arc<CourseService>>,
    Json(req): Json<WithdrawRequest>,
) -> Result<Json<serde_json::Value>, (StatusCode, Json<serde_json::Value>)> {
    let student_id = StudentId(req.student_id);
    let course_id = CourseId(req.course_id);

    match service.withdraw(&student_id, &course_id) {
        Ok(recorded) => Ok(Json(serde_json::json!({
            "success": true,
            "message": if recorded { "退课成功，成绩单将显示W标记" } else { "退课成功" },
            "recorded": recorded,
        }))),
        Err(err) => {
            let (code, msg) = map_error(err);
            Err((code, Json(serde_json::json!({ "error": msg }))))
        }
    }
}

pub async fn get_enrollments(
    State(service): State<Arc<CourseService>>,
    Path(id): Path<String>,
) -> Result<Json<serde_json::Value>, (StatusCode, Json<serde_json::Value>)> {
    let student_id = StudentId(id);

    match service.get_enrollment_result(&student_id) {
        Ok(result) => Ok(Json(serde_json::json!({
            "student_id": result.student_id.0,
            "enrolled_courses": result.enrolled_courses.into_iter().map(|c| {
                serde_json::json!({
                    "id": c.id.0,
                    "name": c.name,
                    "credits": c.credits,
                })
            }).collect::<Vec<_>>(),
            "total_credits": result.total_credits,
        }))),
        Err(err) => {
            let (code, msg) = map_error(err);
            Err((code, Json(serde_json::json!({ "error": msg }))))
        }
    }
}

pub async fn get_transcript(
    State(service): State<Arc<CourseService>>,
    Path(id): Path<String>,
) -> Result<Json<serde_json::Value>, (StatusCode, Json<serde_json::Value>)> {
    let student_id = StudentId(id);

    match service.get_transcript(&student_id) {
        Ok(transcript) => {
            let entries: Vec<serde_json::Value> = transcript.courses
                .into_iter()
                .map(|entry| match entry {
                    course_core::TranscriptEntry::Completed { course_id, course_name, score, credits } => {
                        serde_json::json!({
                            "type": "completed",
                            "course_id": course_id.0,
                            "course_name": course_name,
                            "score": score,
                            "credits": credits,
                        })
                    }
                    course_core::TranscriptEntry::Withdrawal { course_id, course_name } => {
                        serde_json::json!({
                            "type": "withdrawal",
                            "course_id": course_id.0,
                            "course_name": course_name,
                        })
                    }
                })
                .collect();

            Ok(Json(serde_json::json!({
                "student_id": transcript.student_id.0,
                "courses": entries,
            })))
        }
        Err(err) => {
            let (code, msg) = map_error(err);
            Err((code, Json(serde_json::json!({ "error": msg }))))
        }
    }
}
