use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum SealStatus {
    InStock,
    Borrowed,
    Maintenance,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum SealType {
    Official,
    Finance,
    Contract,
    Legal,
    Other,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Seal {
    pub id: Uuid,
    pub name: String,
    pub seal_type: SealType,
    pub custodian_id: Uuid,
    pub status: SealStatus,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum BorrowRequestStatus {
    Pending,
    Approved,
    Rejected,
    Canceled,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BorrowRequest {
    pub id: Uuid,
    pub seal_id: Uuid,
    pub borrower_id: Uuid,
    pub reason: String,
    pub expected_return_date: DateTime<Utc>,
    pub actual_return_date: Option<DateTime<Utc>>,
    pub status: BorrowRequestStatus,
    pub approver_id: Option<Uuid>,
    pub reject_reason: Option<String>,
    pub is_renewal: bool,
    pub original_request_id: Option<Uuid>,
    pub created_at: DateTime<Utc>,
    pub approved_at: Option<DateTime<Utc>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ReminderRecord {
    pub id: Uuid,
    pub borrow_request_id: Uuid,
    pub reminder_date: DateTime<Utc>,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Employee {
    pub id: Uuid,
    pub name: String,
    pub email: String,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateSealRequest {
    pub name: String,
    pub seal_type: SealType,
    pub custodian_id: Uuid,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateBorrowRequest {
    pub seal_id: Uuid,
    pub borrower_id: Uuid,
    pub reason: String,
    pub expected_return_date: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApproveRequest {
    pub request_id: Uuid,
    pub approver_id: Uuid,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RejectRequest {
    pub request_id: Uuid,
    pub approver_id: Uuid,
    pub reason: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RenewRequest {
    pub request_id: Uuid,
    pub borrower_id: Uuid,
    pub reason: String,
    pub new_expected_return_date: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ReturnRequest {
    pub request_id: Uuid,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UpdateSealStatusRequest {
    pub seal_id: Uuid,
    pub status: SealStatus,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateEmployeeRequest {
    pub name: String,
    pub email: String,
}
