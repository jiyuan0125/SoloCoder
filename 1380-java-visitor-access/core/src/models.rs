use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
pub enum VisitorStatus {
    PendingApproval,
    Approved,
    Rejected,
    Visiting,
    Completed,
    NotSignedOut,
    PasscodeExpired,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct VisitorRegistration {
    pub name: String,
    pub phone: String,
    pub id_card: String,
    pub purpose: String,
    pub host: String,
    pub expected_arrival_time: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct VisitorRecord {
    pub id: Uuid,
    pub name: String,
    pub phone: String,
    pub id_card: String,
    pub purpose: String,
    pub host: String,
    pub expected_arrival_time: DateTime<Utc>,
    pub status: VisitorStatus,
    pub passcode: Option<String>,
    pub passcode_expires_at: Option<DateTime<Utc>>,
    pub passcode_used: bool,
    pub arrival_time: Option<DateTime<Utc>>,
    pub departure_time: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

impl VisitorRecord {
    pub fn new(registration: VisitorRegistration) -> Self {
        let now = Utc::now();
        Self {
            id: Uuid::new_v4(),
            name: registration.name,
            phone: registration.phone,
            id_card: registration.id_card,
            purpose: registration.purpose,
            host: registration.host,
            expected_arrival_time: registration.expected_arrival_time,
            status: VisitorStatus::PendingApproval,
            passcode: None,
            passcode_expires_at: None,
            passcode_used: false,
            arrival_time: None,
            departure_time: None,
            created_at: now,
            updated_at: now,
        }
    }

    pub fn update_from_registration(&mut self, registration: VisitorRegistration) {
        self.name = registration.name;
        self.phone = registration.phone;
        self.id_card = registration.id_card;
        self.purpose = registration.purpose;
        self.host = registration.host;
        self.expected_arrival_time = registration.expected_arrival_time;
        self.status = VisitorStatus::PendingApproval;
        self.passcode = None;
        self.passcode_expires_at = None;
        self.passcode_used = false;
        self.updated_at = Utc::now();
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApprovalDecision {
    pub approved: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PasscodeVerification {
    pub passcode: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct VisitorUpdate {
    pub id: Uuid,
    pub registration: VisitorRegistration,
}
