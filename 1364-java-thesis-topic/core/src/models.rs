use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

pub type StudentId = Uuid;
pub type AdvisorId = Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum SystemPhase {
    StudentSubmission,
    AdvisorSelection,
    Appeal,
    Finalized,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum StudentStatus {
    Draft,
    Submitted,
    WaitingAdvisor(PriorityLevel),
    AcceptedByAdvisor(AdvisorId),
    RejectedByCurrent,
    ToBeAssigned,
    Assigned(AdvisorId),
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum PriorityLevel {
    First,
    Second,
    Third,
}

impl PriorityLevel {
    pub fn next(self) -> Option<Self> {
        match self {
            PriorityLevel::First => Some(PriorityLevel::Second),
            PriorityLevel::Second => Some(PriorityLevel::Third),
            PriorityLevel::Third => None,
        }
    }

    pub fn index(self) -> usize {
        match self {
            PriorityLevel::First => 0,
            PriorityLevel::Second => 1,
            PriorityLevel::Third => 2,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Student {
    pub id: StudentId,
    pub name: String,
    pub student_no: String,
    pub status: StudentStatus,
    pub preferences: [Option<AdvisorId>; 3],
    pub can_modify_once: bool,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Advisor {
    pub id: AdvisorId,
    pub name: String,
    pub capacity: u32,
    pub accepted: Vec<StudentId>,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SelectionRecord {
    pub id: Uuid,
    pub advisor_id: AdvisorId,
    pub student_id: StudentId,
    pub priority: PriorityLevel,
    pub accepted: bool,
    pub timestamp: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AppealRecord {
    pub id: Uuid,
    pub student_id: StudentId,
    pub reason: String,
    pub resolved: bool,
    pub created_at: DateTime<Utc>,
    pub resolved_at: Option<DateTime<Utc>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ChangeType {
    StatusChange { from: StudentStatus, to: StudentStatus },
    Assignment { advisor_id: AdvisorId },
    Unassignment { advisor_id: AdvisorId },
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChangeLog {
    pub id: Uuid,
    pub student_id: StudentId,
    pub change_type: ChangeType,
    pub operator: String,
    pub reason: Option<String>,
    pub timestamp: DateTime<Utc>,
}

impl Student {
    pub fn new(name: String, student_no: String) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            student_no,
            status: StudentStatus::Draft,
            preferences: [None, None, None],
            can_modify_once: false,
            created_at: Utc::now(),
        }
    }

    pub fn active_preference(&self) -> Option<AdvisorId> {
        match self.status {
            StudentStatus::WaitingAdvisor(p) => self.preferences[p.index()],
            _ => None,
        }
    }
}

impl Advisor {
    pub fn new(name: String, capacity: u32) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            capacity,
            accepted: Vec::new(),
            created_at: Utc::now(),
        }
    }

    pub fn has_capacity(&self) -> bool {
        self.accepted.len() < self.capacity as usize
    }
}
