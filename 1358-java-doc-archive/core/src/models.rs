use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ClassificationLevel {
    Public,
    Internal,
    Confidential,
}

impl ClassificationLevel {
    pub fn max_borrow_days(&self) -> i64 {
        match self {
            ClassificationLevel::Public => 30,
            ClassificationLevel::Internal => 14,
            ClassificationLevel::Confidential => 7,
        }
    }

    pub fn can_take_out(&self) -> bool {
        match self {
            ClassificationLevel::Public => true,
            ClassificationLevel::Internal => true,
            ClassificationLevel::Confidential => false,
        }
    }

    pub fn requires_second_approval(&self) -> bool {
        match self {
            ClassificationLevel::Public => false,
            ClassificationLevel::Internal => false,
            ClassificationLevel::Confidential => true,
        }
    }

    pub fn can_renew(&self) -> bool {
        match self {
            ClassificationLevel::Public => true,
            ClassificationLevel::Internal => true,
            ClassificationLevel::Confidential => false,
        }
    }
}

impl std::fmt::Display for ClassificationLevel {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            ClassificationLevel::Public => write!(f, "公开"),
            ClassificationLevel::Internal => write!(f, "内部"),
            ClassificationLevel::Confidential => write!(f, "机密"),
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum UserRole {
    User,
    Approver,
    Admin,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct User {
    pub id: String,
    pub name: String,
    pub email: String,
    pub department: String,
    pub role: UserRole,
    pub max_borrow_count: u32,
    pub is_suspended: bool,
    pub created_at: DateTime<Utc>,
}

impl User {
    pub fn new(id: &str, name: &str, email: &str, department: &str, role: UserRole) -> Self {
        Self {
            id: id.to_string(),
            name: name.to_string(),
            email: email.to_string(),
            department: department.to_string(),
            role,
            max_borrow_count: 5,
            is_suspended: false,
            created_at: Utc::now(),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Archive {
    pub id: String,
    pub title: String,
    pub description: String,
    pub classification: ClassificationLevel,
    pub is_borrowed: bool,
    pub current_borrow_id: Option<String>,
    pub created_at: DateTime<Utc>,
}

impl Archive {
    pub fn new(id: &str, title: &str, description: &str, classification: ClassificationLevel) -> Self {
        Self {
            id: id.to_string(),
            title: title.to_string(),
            description: description.to_string(),
            classification,
            is_borrowed: false,
            current_borrow_id: None,
            created_at: Utc::now(),
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ApprovalStatus {
    Pending,
    Approved,
    Rejected,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApprovalStep {
    pub approver_id: String,
    pub step: u8,
    pub status: ApprovalStatus,
    pub comment: Option<String>,
    pub approved_at: Option<DateTime<Utc>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApprovalProcess {
    pub id: String,
    pub borrow_id: String,
    pub archive_id: String,
    pub requester_id: String,
    pub steps: Vec<ApprovalStep>,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum BorrowStatus {
    PendingApproval,
    Approved,
    Active,
    Returned,
    Renewed,
    Overdue,
    Cancelled,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Borrow {
    pub id: String,
    pub archive_id: String,
    pub user_id: String,
    pub status: BorrowStatus,
    pub borrow_date: DateTime<Utc>,
    pub due_date: DateTime<Utc>,
    pub return_date: Option<DateTime<Utc>>,
    pub renewal_count: u32,
    pub approval_process_id: Option<String>,
    pub is_reading_room: bool,
    pub damage_report: Option<DamageReport>,
    pub created_at: DateTime<Utc>,
}

impl Borrow {
    pub fn new(
        archive_id: &str,
        user_id: &str,
        classification: ClassificationLevel,
    ) -> Self {
        let now = Utc::now();
        let due_date = now + chrono::Duration::days(classification.max_borrow_days());
        let status = if classification.requires_second_approval() {
            BorrowStatus::PendingApproval
        } else {
            BorrowStatus::Approved
        };

        Self {
            id: Uuid::new_v4().to_string(),
            archive_id: archive_id.to_string(),
            user_id: user_id.to_string(),
            status,
            borrow_date: now,
            due_date,
            return_date: None,
            renewal_count: 0,
            approval_process_id: None,
            is_reading_room: !classification.can_take_out(),
            damage_report: None,
            created_at: now,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DamageReport {
    pub id: String,
    pub borrow_id: String,
    pub description: String,
    pub compensation_amount: Option<f64>,
    pub reported_by: String,
    pub reported_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ReservationStatus {
    Pending,
    Notified,
    Claimed,
    Cancelled,
    Expired,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Reservation {
    pub id: String,
    pub archive_id: String,
    pub user_id: String,
    pub status: ReservationStatus,
    pub queue_position: u32,
    pub notified_at: Option<DateTime<Utc>>,
    pub claimed_at: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
}

impl Reservation {
    pub fn new(archive_id: &str, user_id: &str, queue_position: u32) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            archive_id: archive_id.to_string(),
            user_id: user_id.to_string(),
            status: ReservationStatus::Pending,
            queue_position,
            notified_at: None,
            claimed_at: None,
            created_at: Utc::now(),
        }
    }

    pub fn is_claim_deadline_passed(&self) -> bool {
        if let Some(notified_at) = self.notified_at {
            let now = Utc::now();
            let deadline = notified_at + chrono::Duration::days(3);
            now > deadline
        } else {
            false
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum OverdueStage {
    Warning,
    DepartmentNotify,
    Suspension,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OverdueInfo {
    pub borrow_id: String,
    pub days_overdue: i64,
    pub stage: OverdueStage,
    pub notified_at: Option<DateTime<Utc>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ReturnCheckResult {
    pub is_damaged: bool,
    pub damage_description: Option<String>,
    pub compensation_amount: Option<f64>,
    pub checked_by: String,
    pub checked_at: DateTime<Utc>,
}
