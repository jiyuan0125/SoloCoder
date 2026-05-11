use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum CourseType {
    Public,
    Professional,
    Elective,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Course {
    pub id: String,
    pub name: String,
    pub credit: f64,
    pub course_type: CourseType,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreditRequirements {
    pub public: f64,
    pub professional: f64,
    pub elective: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Major {
    pub id: String,
    pub name: String,
    pub requirements: CreditRequirements,
    pub curriculum: HashMap<String, Course>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Student {
    pub id: String,
    pub name: String,
    pub enrollment_year: u32,
    pub current_major_id: String,
    pub completed_courses: Vec<CompletedCourse>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CompletedCourse {
    pub course_id: String,
    pub course_name: String,
    pub credit: f64,
    pub original_course_type: CourseType,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreditSummary {
    pub public: f64,
    pub professional: f64,
    pub elective: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GraduationGap {
    pub public: f64,
    pub professional: f64,
    pub elective: f64,
    pub total: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TransferResult {
    pub student_id: String,
    pub from_major_id: String,
    pub to_major_id: String,
    pub transferred_credits: CreditSummary,
    pub converted_to_elective: f64,
    pub excess_elective: f64,
    pub graduation_gap: GraduationGap,
    pub success: bool,
    pub message: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TransferRequest {
    pub student_id: String,
    pub target_major_id: String,
}
