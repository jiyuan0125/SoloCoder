use std::str::FromStr;

use chrono::{DateTime, Duration, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum AuditStage {
    Submitted,
    InitialReview,
    ReReview,
    FinalReview,
    Approved,
    FinalRejected,
}

impl AuditStage {
    pub fn next(self) -> Option<Self> {
        match self {
            Self::Submitted => Some(Self::InitialReview),
            Self::InitialReview => Some(Self::ReReview),
            Self::ReReview => Some(Self::FinalReview),
            Self::FinalReview => Some(Self::Approved),
            _ => None,
        }
    }

    pub fn can_return_to(self, target: Self) -> bool {
        match self {
            Self::InitialReview => matches!(target, Self::Submitted),
            Self::ReReview => matches!(target, Self::Submitted | Self::InitialReview),
            Self::FinalReview => matches!(target, Self::Submitted | Self::InitialReview | Self::ReReview),
            _ => false,
        }
    }

    pub fn is_terminal(self) -> bool {
        matches!(self, Self::Approved | Self::FinalRejected)
    }

    pub fn deadline_hours(self) -> Option<i64> {
        match self {
            Self::InitialReview => Some(48),
            Self::ReReview => Some(72),
            Self::FinalReview => Some(96),
            _ => None,
        }
    }

    pub fn display_name(self) -> &'static str {
        match self {
            Self::Submitted => "已提交",
            Self::InitialReview => "初审",
            Self::ReReview => "复审",
            Self::FinalReview => "终审",
            Self::Approved => "已通过",
            Self::FinalRejected => "终审驳回",
        }
    }
}

impl FromStr for AuditStage {
    type Err = String;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s {
            "Submitted" => Ok(Self::Submitted),
            "InitialReview" => Ok(Self::InitialReview),
            "ReReview" => Ok(Self::ReReview),
            "FinalReview" => Ok(Self::FinalReview),
            "Approved" => Ok(Self::Approved),
            "FinalRejected" => Ok(Self::FinalRejected),
            _ => Err(format!("Invalid audit stage: {}", s)),
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum PriorityLevel {
    Normal,
    Serious,
    Urgent,
}

impl PriorityLevel {
    pub fn upgrade(self) -> Option<Self> {
        match self {
            Self::Normal => Some(Self::Serious),
            Self::Serious => Some(Self::Urgent),
            Self::Urgent => None,
        }
    }

    pub fn display_name(self) -> &'static str {
        match self {
            Self::Normal => "一般",
            Self::Serious => "严重",
            Self::Urgent => "紧急",
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum TimeoutStatus {
    OnTime,
    Timeout,
    Escalated,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Claim {
    pub id: Uuid,
    pub employee_id: String,
    pub title: String,
    pub description: String,
    pub amount: f64,
    pub current_stage: AuditStage,
    pub priority: PriorityLevel,
    pub timeout_status: TimeoutStatus,
    pub stage_start_time: DateTime<Utc>,
    pub stage_deadline: Option<DateTime<Utc>>,
    pub submitted_at: DateTime<Utc>,
    pub operation_history: Vec<OperationRecord>,
    pub escalated_to_manager: bool,
}

impl Claim {
    pub fn new(employee_id: String, title: String, description: String, amount: f64) -> Self {
        let now = Utc::now();
        let operator_id = employee_id.clone();
        Self {
            id: Uuid::new_v4(),
            employee_id,
            title,
            description,
            amount,
            current_stage: AuditStage::Submitted,
            priority: PriorityLevel::Normal,
            timeout_status: TimeoutStatus::OnTime,
            stage_start_time: now,
            stage_deadline: None,
            submitted_at: now,
            operation_history: vec![OperationRecord {
                stage: AuditStage::Submitted,
                action: "提交申请".to_string(),
                operator: operator_id,
                reason: None,
                timestamp: now,
            }],
            escalated_to_manager: false,
        }
    }

    pub fn update_content(&mut self, title: String, description: String, amount: f64) {
        self.title = title;
        self.description = description;
        self.amount = amount;
    }

    pub fn start_audit(&mut self) {
        if self.current_stage == AuditStage::Submitted {
            self.transition_to(AuditStage::InitialReview, "系统自动进入初审", "SYSTEM", None);
        }
    }

    fn transition_to(&mut self, new_stage: AuditStage, action: &str, operator: &str, reason: Option<String>) {
        let now = Utc::now();
        self.current_stage = new_stage;
        self.stage_start_time = now;
        self.stage_deadline = new_stage.deadline_hours().map(|h| now + Duration::hours(h));
        self.timeout_status = TimeoutStatus::OnTime;
        self.operation_history.push(OperationRecord {
            stage: new_stage,
            action: action.to_string(),
            operator: operator.to_string(),
            reason,
            timestamp: now,
        });
    }

    pub fn approve_current(&mut self, operator: &str) -> Result<(), String> {
        if self.current_stage.is_terminal() {
            return Err("申请已处于终态".to_string());
        }
        if let Some(next) = self.current_stage.next() {
            self.transition_to(next, "通过审核", operator, None);
            Ok(())
        } else {
            Err("无法通过当前环节".to_string())
        }
    }

    pub fn return_to(&mut self, target_stage: AuditStage, reason: String, operator: &str) -> Result<(), String> {
        if self.current_stage == AuditStage::FinalReview {
            return Err("终审环节只能终审驳回，不能退回".to_string());
        }
        if self.current_stage.is_terminal() {
            return Err("申请已处于终态".to_string());
        }
        if !self.current_stage.can_return_to(target_stage) {
            return Err(format!(
                "无法从{}退回到{}",
                self.current_stage.display_name(),
                target_stage.display_name()
            ));
        }
        self.transition_to(target_stage, "退回修改", operator, Some(reason));
        self.priority = PriorityLevel::Normal;
        self.escalated_to_manager = false;
        Ok(())
    }

    pub fn final_reject(&mut self, reason: String, operator: &str) -> Result<(), String> {
        if self.current_stage != AuditStage::FinalReview {
            return Err("只有终审环节可以终审驳回".to_string());
        }
        self.transition_to(AuditStage::FinalRejected, "终审驳回", operator, Some(reason));
        Ok(())
    }

    pub fn resubmit(&mut self, employee_id: &str) -> Result<(), String> {
        if self.current_stage == AuditStage::Submitted {
            return Err("申请已在待提交状态".to_string());
        }
        if self.current_stage.is_terminal() {
            return Err("申请已处于终态，无法重新提交".to_string());
        }
        self.transition_to(AuditStage::InitialReview, "重新提交", employee_id, None);
        self.priority = PriorityLevel::Normal;
        self.escalated_to_manager = false;
        Ok(())
    }

    pub fn check_timeout(&mut self) -> TimeoutCheckResult {
        let now = Utc::now();
        let deadline = match self.stage_deadline {
            Some(d) => d,
            None => return TimeoutCheckResult::NoDeadline,
        };

        if now < deadline || self.current_stage.is_terminal() {
            return TimeoutCheckResult::NoChange;
        }

        if self.timeout_status == TimeoutStatus::OnTime {
            self.timeout_status = TimeoutStatus::Timeout;
            self.operation_history.push(OperationRecord {
                stage: self.current_stage,
                action: "超时".to_string(),
                operator: "SYSTEM".to_string(),
                reason: None,
                timestamp: now,
            });
            return TimeoutCheckResult::NewTimeout;
        }

        if self.priority == PriorityLevel::Urgent && !self.escalated_to_manager {
            self.escalated_to_manager = true;
            self.timeout_status = TimeoutStatus::Escalated;
            self.operation_history.push(OperationRecord {
                stage: self.current_stage,
                action: "升级至上级主管".to_string(),
                operator: "SYSTEM".to_string(),
                reason: Some("紧急投诉超时".to_string()),
                timestamp: now,
            });
            return TimeoutCheckResult::Escalated;
        }

        if let Some(new_priority) = self.priority.upgrade() {
            self.priority = new_priority;
            self.stage_deadline = Some(now + Duration::hours(self.current_stage.deadline_hours().unwrap_or(48)));
            self.operation_history.push(OperationRecord {
                stage: self.current_stage,
                action: format!("优先级升级至{}", new_priority.display_name()),
                operator: "SYSTEM".to_string(),
                reason: Some("超时升级".to_string()),
                timestamp: now,
            });
            return TimeoutCheckResult::PriorityUpgraded(new_priority);
        }

        TimeoutCheckResult::NoChange
    }
}

pub enum TimeoutCheckResult {
    NoDeadline,
    NoChange,
    NewTimeout,
    PriorityUpgraded(PriorityLevel),
    Escalated,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OperationRecord {
    pub stage: AuditStage,
    pub action: String,
    pub operator: String,
    pub reason: Option<String>,
    pub timestamp: DateTime<Utc>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct CreateClaimRequest {
    pub employee_id: String,
    pub title: String,
    pub description: String,
    pub amount: f64,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct UpdateClaimRequest {
    pub title: String,
    pub description: String,
    pub amount: f64,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ApproveRequest {
    pub operator: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ReturnRequest {
    pub target_stage: AuditStage,
    pub reason: String,
    pub operator: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct FinalRejectRequest {
    pub reason: String,
    pub operator: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ResubmitRequest {
    pub employee_id: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ApiResponse<T> {
    pub success: bool,
    pub message: String,
    pub data: Option<T>,
}

impl<T> ApiResponse<T> {
    pub fn success(data: T) -> Self {
        Self {
            success: true,
            message: "操作成功".to_string(),
            data: Some(data),
        }
    }

    pub fn error(message: String) -> Self {
        Self {
            success: false,
            message,
            data: None,
        }
    }
}
