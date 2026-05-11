use chrono::{DateTime, Utc};
use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum BatchState {
    InProgress,
    Completed,
    Cancelled,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum DifferenceType {
    Over,
    Shortage,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum DifferenceState {
    PendingAutoReview,
    AutoApproved,
    PendingApproval,
    Submitted,
    Approved,
    Rejected,
    Adjusted,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Material {
    pub code: String,
    pub name: String,
    pub price: Decimal,
    pub book_quantity: i64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StockBatch {
    pub batch_no: String,
    pub state: BatchState,
    pub created_at: DateTime<Utc>,
    pub completed_at: Option<DateTime<Utc>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CountRecord {
    pub id: Uuid,
    pub batch_no: String,
    pub material_code: String,
    pub actual_quantity: i64,
    pub operator: String,
    pub counted_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StockDifference {
    pub id: Uuid,
    pub batch_no: String,
    pub material_code: String,
    pub book_quantity: i64,
    pub actual_quantity: i64,
    pub difference_quantity: i64,
    pub difference_amount: Decimal,
    pub difference_type: DifferenceType,
    pub state: DifferenceState,
    pub reason: Option<String>,
    pub operator: String,
    pub approver: Option<String>,
    pub created_at: DateTime<Utc>,
    pub approved_at: Option<DateTime<Utc>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SubmitDifferenceRequest {
    pub reason: String,
    pub operator: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApproveDifferenceRequest {
    pub approved: bool,
    pub approver: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateCountRecordRequest {
    pub batch_no: String,
    pub material_code: String,
    pub actual_quantity: i64,
    pub operator: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateBatchRequest {
    pub operator: String,
}
