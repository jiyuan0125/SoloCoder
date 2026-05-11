use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
pub enum DocumentType {
    Notice,
    Request,
    Report,
    Letter,
}

impl DocumentType {
    pub fn to_str(&self) -> &'static str {
        match self {
            DocumentType::Notice => "通知",
            DocumentType::Request => "请示",
            DocumentType::Report => "报告",
            DocumentType::Letter => "函",
        }
    }

    pub fn from_str(s: &str) -> Option<Self> {
        match s {
            "通知" => Some(DocumentType::Notice),
            "请示" => Some(DocumentType::Request),
            "报告" => Some(DocumentType::Report),
            "函" => Some(DocumentType::Letter),
            _ => None,
        }
    }

    pub fn needs_countersignature(&self) -> bool {
        matches!(self, DocumentType::Request)
    }
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
pub enum UrgencyLevel {
    Normal,
    Urgent,
    ExtraUrgent,
}

impl UrgencyLevel {
    pub fn to_str(&self) -> &'static str {
        match self {
            UrgencyLevel::Normal => "普通",
            UrgencyLevel::Urgent => "加急",
            UrgencyLevel::ExtraUrgent => "特急",
        }
    }

    pub fn from_str(s: &str) -> Option<Self> {
        match s {
            "普通" => Some(UrgencyLevel::Normal),
            "加急" => Some(UrgencyLevel::Urgent),
            "特急" => Some(UrgencyLevel::ExtraUrgent),
            _ => None,
        }
    }

    pub fn timeout_hours(&self) -> u32 {
        match self {
            UrgencyLevel::Normal => 72,
            UrgencyLevel::Urgent => 24,
            UrgencyLevel::ExtraUrgent => 4,
        }
    }
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
pub enum WorkflowStage {
    Draft,
    DepartmentReview,
    Countersignature,
    Issuance,
    Archived,
}

impl WorkflowStage {
    pub fn to_str(&self) -> &'static str {
        match self {
            WorkflowStage::Draft => "拟稿",
            WorkflowStage::DepartmentReview => "部门审核",
            WorkflowStage::Countersignature => "会签",
            WorkflowStage::Issuance => "签发",
            WorkflowStage::Archived => "归档",
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Document {
    pub id: Uuid,
    pub number: String,
    pub title: String,
    pub content: String,
    pub doc_type: DocumentType,
    pub urgency: UrgencyLevel,
    pub author: String,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
    pub current_stage: WorkflowStage,
    pub is_overdue: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum HistoryAction {
    Create,
    Submit { from: WorkflowStage, to: WorkflowStage },
    ReviewApprove,
    CountersignApprove { comment: Option<String> },
    Issue,
    Archive,
    Return { from: WorkflowStage, reason: String },
    Resubmit,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WorkflowHistory {
    pub id: Uuid,
    pub document_id: Uuid,
    pub action: HistoryAction,
    pub operator: String,
    pub timestamp: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StageStatus {
    pub document_id: Uuid,
    pub stage: WorkflowStage,
    pub entered_at: DateTime<Utc>,
    pub assignee: String,
}
