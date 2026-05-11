use chrono::{DateTime, Duration, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Department {
    pub id: Uuid,
    pub name: String,
    pub annual_budget: f64,
    pub remaining_budget: f64,
}

impl Department {
    pub fn new(name: String, annual_budget: f64) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            annual_budget,
            remaining_budget: annual_budget,
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ActivityStatus {
    Draft,
    Published,
    RegistrationClosed,
    InProgress,
    Completed,
    Cancelled,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum RegistrationStatus {
    Registered,
    Waitlisted,
    TemporaryExit,
    Cancelled,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Registration {
    pub user_id: String,
    pub status: RegistrationStatus,
    pub registered_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Activity {
    pub id: Uuid,
    pub department_id: Uuid,
    pub name: String,
    pub description: String,
    pub cost_per_person: f64,
    pub max_participants: usize,
    pub start_time: DateTime<Utc>,
    pub end_time: DateTime<Utc>,
    pub status: ActivityStatus,
    pub registrations: Vec<Registration>,
    pub waitlist: Vec<String>,
    pub actual_participants: Option<Vec<String>>,
    pub created_at: DateTime<Utc>,
}

impl Activity {
    pub fn new(
        department_id: Uuid,
        name: String,
        description: String,
        cost_per_person: f64,
        max_participants: usize,
        start_time: DateTime<Utc>,
        end_time: DateTime<Utc>,
    ) -> Self {
        Self {
            id: Uuid::new_v4(),
            department_id,
            name,
            description,
            cost_per_person,
            max_participants,
            start_time,
            end_time,
            status: ActivityStatus::Draft,
            registrations: Vec::new(),
            waitlist: Vec::new(),
            actual_participants: None,
            created_at: Utc::now(),
        }
    }

    pub fn registration_deadline(&self) -> DateTime<Utc> {
        self.start_time - Duration::hours(48)
    }

    pub fn is_registration_open(&self, now: DateTime<Utc>) -> bool {
        if self.status != ActivityStatus::Published {
            return false;
        }
        now < self.registration_deadline()
    }

    pub fn is_before_24_hours(&self, now: DateTime<Utc>) -> bool {
        now < self.start_time - Duration::hours(24)
    }

    pub fn estimated_total_cost(&self) -> f64 {
        self.cost_per_person * self.max_participants as f64
    }

    pub fn registered_count(&self) -> usize {
        self.registrations
            .iter()
            .filter(|r| r.status == RegistrationStatus::Registered)
            .count()
    }

    pub fn is_full(&self) -> bool {
        self.registered_count() >= self.max_participants
    }

    pub fn has_registration(&self, user_id: &str) -> bool {
        self.registrations.iter().any(|r| r.user_id == user_id)
    }

    pub fn get_registration(&self, user_id: &str) -> Option<&Registration> {
        self.registrations.iter().find(|r| r.user_id == user_id)
    }

    pub fn has_waitlist_entry(&self, user_id: &str) -> bool {
        self.waitlist.contains(&user_id.to_string())
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateActivityRequest {
    pub department_id: Uuid,
    pub name: String,
    pub description: String,
    pub cost_per_person: f64,
    pub max_participants: usize,
    pub start_time: DateTime<Utc>,
    pub end_time: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UpdateLimitRequest {
    pub new_limit: usize,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SettlingRequest {
    pub actual_participants: Vec<String>,
}
