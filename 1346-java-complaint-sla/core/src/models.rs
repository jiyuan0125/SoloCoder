use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum Severity {
    #[serde(rename = "general")]
    General,
    #[serde(rename = "serious")]
    Serious,
    #[serde(rename = "urgent")]
    Urgent,
}

impl Severity {
    pub fn response_hours(&self) -> u32 {
        match self {
            Severity::General => 24,
            Severity::Serious => 8,
            Severity::Urgent => 2,
        }
    }

    pub fn upgrade(&self) -> Option<Self> {
        match self {
            Severity::General => Some(Severity::Serious),
            Severity::Serious => Some(Severity::Urgent),
            Severity::Urgent => None,
        }
    }

    pub fn as_str(&self) -> &'static str {
        match self {
            Severity::General => "一般",
            Severity::Serious => "严重",
            Severity::Urgent => "紧急",
        }
    }
}

impl std::fmt::Display for Severity {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.as_str())
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum Status {
    #[serde(rename = "pending")]
    Pending,
    #[serde(rename = "processing")]
    Processing,
    #[serde(rename = "awaiting_feedback")]
    AwaitingFeedback,
    #[serde(rename = "reopened")]
    Reopened,
    #[serde(rename = "escalated")]
    Escalated,
    #[serde(rename = "closed")]
    Closed,
}

impl Status {
    pub fn as_str(&self) -> &'static str {
        match self {
            Status::Pending => "待处理",
            Status::Processing => "处理中",
            Status::AwaitingFeedback => "待客户反馈",
            Status::Reopened => "重新打开",
            Status::Escalated => "已升级",
            Status::Closed => "已结案",
        }
    }
}

impl std::fmt::Display for Status {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.as_str())
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Handler {
    pub id: String,
    pub name: String,
    pub level: u32,
}

impl Handler {
    pub fn new(id: impl Into<String>, name: impl Into<String>, level: u32) -> Self {
        Self {
            id: id.into(),
            name: name.into(),
            level,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EscalationRecord {
    pub id: Uuid,
    pub from_severity: Option<Severity>,
    pub to_severity: Option<Severity>,
    pub reason: String,
    pub escalated_at: DateTime<Utc>,
    pub new_handler_id: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FollowUp {
    pub id: Uuid,
    pub complaint_id: Uuid,
    pub satisfaction: bool,
    pub comments: Option<String>,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ComplaintHistory {
    pub id: Uuid,
    pub action: String,
    pub details: Option<String>,
    pub timestamp: DateTime<Utc>,
    pub actor: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Complaint {
    pub id: Uuid,
    pub title: String,
    pub description: String,
    pub customer_name: String,
    pub customer_contact: String,
    pub severity: Severity,
    pub original_severity: Severity,
    pub status: Status,
    pub handler_id: Option<String>,
    pub created_at: DateTime<Utc>,
    pub deadline: DateTime<Utc>,
    pub response_time: Option<DateTime<Utc>>,
    pub closed_at: Option<DateTime<Utc>>,
    pub reopen_count: u32,
    pub escalations: Vec<EscalationRecord>,
    pub follow_ups: Vec<FollowUp>,
    pub history: Vec<ComplaintHistory>,
}

impl Complaint {
    pub fn new(
        title: impl Into<String>,
        description: impl Into<String>,
        customer_name: impl Into<String>,
        customer_contact: impl Into<String>,
        severity: Severity,
        handler_id: Option<String>,
        now: DateTime<Utc>,
        deadline: DateTime<Utc>,
    ) -> Self {
        let id = Uuid::new_v4();
        let history = vec![ComplaintHistory {
            id: Uuid::new_v4(),
            action: "创建投诉".to_string(),
            details: Some(format!("严重程度: {}", severity.as_str())),
            timestamp: now,
            actor: "系统".to_string(),
        }];

        Self {
            id,
            title: title.into(),
            description: description.into(),
            customer_name: customer_name.into(),
            customer_contact: customer_contact.into(),
            severity,
            original_severity: severity,
            status: Status::Pending,
            handler_id,
            created_at: now,
            deadline,
            response_time: None,
            closed_at: None,
            reopen_count: 0,
            escalations: Vec::new(),
            follow_ups: Vec::new(),
            history,
        }
    }

    pub fn add_history(&mut self, action: impl Into<String>, details: Option<String>, actor: impl Into<String>) {
        self.history.push(ComplaintHistory {
            id: Uuid::new_v4(),
            action: action.into(),
            details,
            timestamp: Utc::now(),
            actor: actor.into(),
        });
    }

    pub fn is_overdue(&self, now: DateTime<Utc>) -> bool {
        now > self.deadline && self.status != Status::Closed
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateComplaintRequest {
    pub title: String,
    pub description: String,
    pub customer_name: String,
    pub customer_contact: String,
    pub severity: Severity,
    pub handler_id: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RespondToComplaintRequest {
    pub handler_id: String,
    pub response: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProposeSolutionRequest {
    pub handler_id: String,
    pub solution: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CustomerFeedbackRequest {
    pub accepted: bool,
    pub comments: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FollowUpFeedbackRequest {
    pub satisfied: bool,
    pub comments: Option<String>,
}
