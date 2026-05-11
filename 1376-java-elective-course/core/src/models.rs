use serde::{Deserialize, Serialize};
use chrono::{DateTime, Utc};

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq, Hash)]
pub struct CourseId(pub String);

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq, Hash)]
pub struct StudentId(pub String);

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Course {
    pub id: CourseId,
    pub name: String,
    pub credits: u32,
    pub capacity: u32,
    pub prerequisites: Vec<CourseId>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CompletedCourse {
    pub course_id: CourseId,
    pub score: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Enrollment {
    pub student_id: StudentId,
    pub course_id: CourseId,
    pub enrolled_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Withdrawal {
    pub student_id: StudentId,
    pub course_id: CourseId,
    pub withdrawn_at: DateTime<Utc>,
    pub recorded: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Student {
    pub id: StudentId,
    pub name: String,
    pub completed_courses: Vec<CompletedCourse>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EnrollmentResult {
    pub student_id: StudentId,
    pub enrolled_courses: Vec<Course>,
    pub total_credits: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Transcript {
    pub student_id: StudentId,
    pub courses: Vec<TranscriptEntry>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum TranscriptEntry {
    Completed {
        course_id: CourseId,
        course_name: String,
        score: f64,
        credits: u32,
    },
    Withdrawal {
        course_id: CourseId,
        course_name: String,
    },
}
