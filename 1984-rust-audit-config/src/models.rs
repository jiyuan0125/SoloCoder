use serde::{Deserialize, Serialize};
use time::OffsetDateTime;
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
#[serde(rename_all = "lowercase")]
pub enum ValueType {
    String,
    Number,
    Json,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct ConfigValue {
    pub value_type: ValueType,
    pub value: serde_json::Value,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Configuration {
    pub namespace: String,
    pub key: String,
    pub config: ConfigValue,
    pub version: u64,
    pub updated_at: OffsetDateTime,
    pub updated_by: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
#[serde(rename_all = "snake_case")]
pub enum ChangeRequestStatus {
    Pending,
    Approved,
    Rejected,
    Canary,
    RolledBack,
    Deployed,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DiffItem {
    pub key: String,
    pub before: Option<ConfigValue>,
    pub after: Option<ConfigValue>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChangeRequest {
    pub id: Uuid,
    pub namespace: String,
    pub submitter: String,
    pub submitted_at: OffsetDateTime,
    pub diffs: Vec<DiffItem>,
    pub status: ChangeRequestStatus,
    pub approver: Option<String>,
    pub approved_at: Option<OffsetDateTime>,
    pub canary_percent: Option<u8>,
    pub canary_started_at: Option<OffsetDateTime>,
    pub canary_duration_secs: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NamespaceConfig {
    pub name: String,
    pub approvers: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
#[serde(rename_all = "snake_case")]
pub enum OperationType {
    Submit,
    Approve,
    Reject,
    Canary,
    Rollback,
    Deploy,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AuditLog {
    pub id: Uuid,
    pub timestamp: OffsetDateTime,
    pub operator: String,
    pub operation_type: OperationType,
    pub namespace: String,
    pub change_request_id: Option<Uuid>,
    pub diffs: Vec<DiffItem>,
    pub description: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SubmitRequest {
    pub namespace: String,
    pub submitter: String,
    pub changes: Vec<ConfigChange>,
    #[serde(default = "default_canary_percent")]
    pub canary_percent: u8,
    #[serde(default = "default_canary_duration")]
    pub canary_duration_secs: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ConfigChange {
    pub key: String,
    pub value_type: ValueType,
    pub value: serde_json::Value,
}

fn default_canary_percent() -> u8 {
    10
}

fn default_canary_duration() -> u64 {
    600
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApproveRequest {
    pub approver: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RollbackRequest {
    pub operator: String,
}
