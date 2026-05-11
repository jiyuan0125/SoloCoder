use chrono::{DateTime, Duration, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum CandidateStatus {
    InitialScreening,
    InitialScreeningPassed,
    InitialScreeningRejected,
    SecondScreening,
    RecommendedForInterview,
    Pending,
    PendingExpired,
    NotSuitable,
    Interviewing,
    InterviewPassed,
    InterviewRejected,
    OfferPending,
    OfferApproved,
    OfferRejected,
    Hired,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum InterviewType {
    Technical,
    HR,
    Director,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum InterviewStatus {
    Scheduled,
    Completed,
    Passed,
    Rejected,
    Cancelled,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ApprovalLevel {
    HRDirector,
    VP,
    CEO,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum OfferStatus {
    Draft,
    PendingApproval,
    Approved,
    Rejected,
    Revised,
    Sent,
    Accepted,
    Declined,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Candidate {
    pub id: Uuid,
    pub name: String,
    pub email: String,
    pub phone: Option<String>,
    pub resume: Option<String>,
    pub status: CandidateStatus,
    pub pending_since: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Interview {
    pub id: Uuid,
    pub candidate_id: Uuid,
    pub interview_type: InterviewType,
    pub round: u32,
    pub interviewer: String,
    pub start_time: DateTime<Utc>,
    pub duration_minutes: u32,
    pub status: InterviewStatus,
    pub notes: Option<String>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OfferVersion {
    pub version: u32,
    pub salary: u32,
    pub position: String,
    pub department: String,
    pub start_date: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Offer {
    pub id: Uuid,
    pub candidate_id: Uuid,
    pub current_version: u32,
    pub versions: Vec<OfferVersion>,
    pub status: OfferStatus,
    pub approval_level: ApprovalLevel,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

impl InterviewType {
    pub fn default_duration(&self) -> u32 {
        match self {
            InterviewType::Technical => 60,
            InterviewType::HR => 30,
            InterviewType::Director => 45,
        }
    }
}

impl ApprovalLevel {
    pub fn for_salary(salary: u32) -> Self {
        if salary < 30000 {
            ApprovalLevel::HRDirector
        } else if salary <= 50000 {
            ApprovalLevel::VP
        } else {
            ApprovalLevel::CEO
        }
    }

    pub fn can_approve(&self, required: &ApprovalLevel) -> bool {
        match (self, required) {
            (ApprovalLevel::CEO, _) => true,
            (ApprovalLevel::VP, ApprovalLevel::CEO) => false,
            (ApprovalLevel::VP, _) => true,
            (ApprovalLevel::HRDirector, ApprovalLevel::HRDirector) => true,
            _ => false,
        }
    }
}

impl Candidate {
    pub fn new(name: String, email: String) -> Self {
        let now = Utc::now();
        Self {
            id: Uuid::new_v4(),
            name,
            email,
            phone: None,
            resume: None,
            status: CandidateStatus::InitialScreening,
            pending_since: None,
            created_at: now,
            updated_at: now,
        }
    }

    pub fn is_pending_expired(&self, now: DateTime<Utc>) -> bool {
        if self.status != CandidateStatus::Pending {
            return false;
        }
        if let Some(pending_since) = self.pending_since {
            now.signed_duration_since(pending_since) > Duration::days(14)
        } else {
            false
        }
    }
}

impl Interview {
    pub fn end_time(&self) -> DateTime<Utc> {
        self.start_time + Duration::minutes(self.duration_minutes as i64)
    }

    pub fn conflicts_with(&self, other: &Interview) -> bool {
        let self_end = self.end_time();
        let other_end = other.end_time();
        !(self_end <= other.start_time || self.start_time >= other_end)
    }
}

impl Offer {
    pub fn new(candidate_id: Uuid, salary: u32, position: String, department: String) -> Self {
        let now = Utc::now();
        let version = OfferVersion {
            version: 1,
            salary,
            position,
            department,
            start_date: None,
            created_at: now,
        };
        Self {
            id: Uuid::new_v4(),
            candidate_id,
            current_version: 1,
            versions: vec![version],
            status: OfferStatus::Draft,
            approval_level: ApprovalLevel::for_salary(salary),
            created_at: now,
            updated_at: now,
        }
    }

    pub fn current_version_data(&self) -> Option<&OfferVersion> {
        self.versions
            .iter()
            .find(|v| v.version == self.current_version)
    }

    pub fn revise(&mut self, salary: u32, position: String, department: String) {
        let new_version_num = self.current_version + 1;
        let now = Utc::now();
        let version = OfferVersion {
            version: new_version_num,
            salary,
            position,
            department,
            start_date: None,
            created_at: now,
        };
        self.versions.push(version);
        self.current_version = new_version_num;
        self.status = OfferStatus::Revised;
        self.approval_level = ApprovalLevel::for_salary(salary);
        self.updated_at = now;
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateCandidateRequest {
    pub name: String,
    pub email: String,
    pub phone: Option<String>,
    pub resume: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InitialScreeningRequest {
    pub passed: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SecondScreeningRequest {
    pub decision: SecondScreeningDecision,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum SecondScreeningDecision {
    RecommendInterview,
    Pending,
    NotSuitable,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ScheduleInterviewRequest {
    pub interview_type: InterviewType,
    pub round: u32,
    pub interviewer: String,
    pub start_time: DateTime<Utc>,
    pub duration_minutes: Option<u32>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InterviewResultRequest {
    pub passed: bool,
    pub notes: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateOfferRequest {
    pub salary: u32,
    pub position: String,
    pub department: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApproveOfferRequest {
    pub approver_level: ApprovalLevel,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ReviseOfferRequest {
    pub salary: u32,
    pub position: String,
    pub department: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ConfirmExpiredCandidateRequest {
    pub confirmed: bool,
}
