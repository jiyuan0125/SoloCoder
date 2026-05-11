use serde::{Deserialize, Serialize};
use chrono::{DateTime, Utc, Duration};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ContractType {
    Purchase,
    Sales,
    Service,
}

impl ContractType {
    pub fn default_validity_days(&self) -> i64 {
        match self {
            ContractType::Purchase => 365,
            ContractType::Sales => 180,
            ContractType::Service => 730,
        }
    }

    pub fn name(&self) -> &'static str {
        match self {
            ContractType::Purchase => "采购合同",
            ContractType::Sales => "销售合同",
            ContractType::Service => "服务合同",
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ContractStatus {
    Draft,
    Active,
    RenewalPending,
    Expiring,
    Terminated,
    Closed,
}

impl ContractStatus {
    pub fn name(&self) -> &'static str {
        match self {
            ContractStatus::Draft => "草稿",
            ContractStatus::Active => "生效中",
            ContractStatus::RenewalPending => "续签审批中",
            ContractStatus::Expiring => "即将到期",
            ContractStatus::Terminated => "已终止",
            ContractStatus::Closed => "已关闭",
        }
    }

    pub fn can_transition_to(&self, next: ContractStatus) -> bool {
        match (self, next) {
            (ContractStatus::Draft, ContractStatus::Active) => true,
            (ContractStatus::Active, ContractStatus::RenewalPending) => true,
            (ContractStatus::Active, ContractStatus::Expiring) => true,
            (ContractStatus::Active, ContractStatus::Terminated) => true,
            (ContractStatus::RenewalPending, ContractStatus::Active) => true,
            (ContractStatus::RenewalPending, ContractStatus::Terminated) => true,
            (ContractStatus::Expiring, ContractStatus::Active) => true,
            (ContractStatus::Expiring, ContractStatus::Closed) => true,
            (ContractStatus::Active, ContractStatus::Closed) => true,
            _ => false,
        }
    }

    pub fn can_renew(&self) -> bool {
        match self {
            ContractStatus::Active | ContractStatus::Expiring | ContractStatus::RenewalPending => true,
            _ => false,
        }
    }

    pub fn can_close(&self) -> bool {
        match self {
            ContractStatus::Active | ContractStatus::Expiring => true,
            _ => false,
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ApprovalLevel {
    DepartmentHead,
    DivisionLeader,
    GeneralManagerMeeting,
}

impl ApprovalLevel {
    pub fn name(&self) -> &'static str {
        match self {
            ApprovalLevel::DepartmentHead => "部门负责人",
            ApprovalLevel::DivisionLeader => "分管领导",
            ApprovalLevel::GeneralManagerMeeting => "总经理办公会",
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ApprovalType {
    Renewal,
    PriceIncrease,
    PriceDecrease,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ApprovalStatus {
    Pending,
    Approved,
    Rejected,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApprovalRequest {
    pub id: Uuid,
    pub contract_id: Uuid,
    pub approval_type: ApprovalType,
    pub level: ApprovalLevel,
    pub status: ApprovalStatus,
    pub old_amount: f64,
    pub new_amount: f64,
    pub created_at: DateTime<Utc>,
    pub decided_at: Option<DateTime<Utc>>,
    pub reason: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ReminderLevel {
    pub days_before: i64,
    pub recipients: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ReminderConfig {
    pub level1: ReminderLevel,
    pub level2: ReminderLevel,
    pub level3: ReminderLevel,
}

impl Default for ReminderConfig {
    fn default() -> Self {
        ReminderConfig {
            level1: ReminderLevel {
                days_before: 60,
                recipients: vec!["合同管理员".to_string()],
            },
            level2: ReminderLevel {
                days_before: 30,
                recipients: vec!["合同管理员".to_string(), "部门负责人".to_string()],
            },
            level3: ReminderLevel {
                days_before: 7,
                recipients: vec!["合同管理员".to_string(), "部门负责人".to_string(), "分管领导".to_string()],
            },
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Contract {
    pub id: Uuid,
    pub name: String,
    pub contract_type: ContractType,
    pub status: ContractStatus,
    pub amount: f64,
    pub effective_amount: f64,
    pub start_date: DateTime<Utc>,
    pub end_date: DateTime<Utc>,
    pub is_framework: bool,
    pub parent_id: Option<Uuid>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
    pub pending_adjustment: Option<PendingPriceAdjustment>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PendingPriceAdjustment {
    pub approval_id: Uuid,
    pub new_amount: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateContractRequest {
    pub name: String,
    pub contract_type: ContractType,
    pub amount: f64,
    pub start_date: DateTime<Utc>,
    pub end_date: Option<DateTime<Utc>>,
    pub is_framework: bool,
    pub parent_id: Option<Uuid>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RenewalRequest {
    pub new_amount: f64,
    pub end_date: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PriceAdjustmentRequest {
    pub new_amount: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ReminderInfo {
    pub contract_id: Uuid,
    pub contract_name: String,
    pub days_until_expiry: i64,
    pub level: u32,
    pub recipients: Vec<String>,
    pub end_date: DateTime<Utc>,
}

impl Contract {
    pub fn new(req: CreateContractRequest) -> Self {
        let now = Utc::now();
        let end_date = req.end_date.unwrap_or_else(|| {
            req.start_date + Duration::days(req.contract_type.default_validity_days())
        });

        Contract {
            id: Uuid::new_v4(),
            name: req.name,
            contract_type: req.contract_type,
            status: ContractStatus::Draft,
            amount: req.amount,
            effective_amount: req.amount,
            start_date: req.start_date,
            end_date,
            is_framework: req.is_framework,
            parent_id: req.parent_id,
            created_at: now,
            updated_at: now,
            pending_adjustment: None,
        }
    }

    pub fn days_until_expiry(&self, now: DateTime<Utc>) -> i64 {
        (self.end_date - now).num_days()
    }

    pub fn get_reminder_level(&self, config: &ReminderConfig, now: DateTime<Utc>) -> Option<u32> {
        let days = self.days_until_expiry(now);
        if days <= config.level3.days_before {
            Some(3)
        } else if days <= config.level2.days_before {
            Some(2)
        } else if days <= config.level1.days_before {
            Some(1)
        } else {
            None
        }
    }

    pub fn get_reminder_recipients(&self, config: &ReminderConfig, level: u32) -> Vec<String> {
        match level {
            1 => config.level1.recipients.clone(),
            2 => config.level2.recipients.clone(),
            3 => config.level3.recipients.clone(),
            _ => vec![],
        }
    }

    pub fn is_expired(&self, now: DateTime<Utc>) -> bool {
        self.end_date <= now
    }
}

pub fn calculate_approval_level_for_renewal(old_amount: f64, new_amount: f64) -> ApprovalLevel {
    if old_amount <= 0.0 {
        return ApprovalLevel::GeneralManagerMeeting;
    }
    
    let change_percent = (new_amount - old_amount) / old_amount * 100.0;
    
    if change_percent.abs() <= 10.0 {
        ApprovalLevel::DepartmentHead
    } else if change_percent.abs() <= 50.0 {
        ApprovalLevel::DivisionLeader
    } else {
        ApprovalLevel::GeneralManagerMeeting
    }
}

pub fn needs_approval_for_price_change(old_amount: f64, new_amount: f64) -> Option<ApprovalLevel> {
    if old_amount <= 0.0 {
        return Some(ApprovalLevel::GeneralManagerMeeting);
    }
    
    let change_percent = (new_amount - old_amount) / old_amount * 100.0;
    
    if change_percent > 20.0 || change_percent < -30.0 {
        if change_percent > 50.0 || change_percent < -50.0 {
            Some(ApprovalLevel::GeneralManagerMeeting)
        } else if change_percent > 30.0 || change_percent < -40.0 {
            Some(ApprovalLevel::DivisionLeader)
        } else {
            Some(ApprovalLevel::DepartmentHead)
        }
    } else {
        None
    }
}
