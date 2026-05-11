use crate::{
    Advisor, AdvisorId, ChangeLog, PriorityLevel, SelectionRecord, Student, StudentId, SystemPhase,
};
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateStudentRequest {
    pub name: String,
    pub student_no: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateAdvisorRequest {
    pub name: String,
    pub capacity: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SubmitPreferencesRequest {
    pub preferences: [Option<AdvisorId>; 3],
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ModifyPreferencesRequest {
    pub preferences: [Option<AdvisorId>; 3],
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AdvisorSelectRequest {
    pub student_id: StudentId,
    pub accept: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateAppealRequest {
    pub student_id: StudentId,
    pub reason: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ManualAssignRequest {
    pub student_id: StudentId,
    pub advisor_id: AdvisorId,
    pub reason: String,
    pub operator: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApiResponse<T> {
    pub success: bool,
    pub data: Option<T>,
    pub error: Option<String>,
}

impl<T> ApiResponse<T> {
    pub fn success(data: T) -> Self {
        Self {
            success: true,
            data: Some(data),
            error: None,
        }
    }

    pub fn error(msg: String) -> Self {
        Self {
            success: false,
            data: None,
            error: Some(msg),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SystemStatus {
    pub phase: SystemPhase,
    pub student_count: usize,
    pub advisor_count: usize,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AdvisorPoolView {
    pub advisor: Advisor,
    pub eligible_students: Vec<Student>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MatchResult {
    pub student: Student,
    pub advisor: Option<Advisor>,
}
