use chrono::{DateTime, Utc};
use serde::{Serialize, Deserialize};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum BatchStatus {
    InStock,
    Frozen,
    PartiallyOut,
    FullyOut,
    Recalled,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DrugBatch {
    pub id: Uuid,
    pub drug_name: String,
    pub batch_no: String,
    pub production_date: DateTime<Utc>,
    pub expiry_date: DateTime<Utc>,
    pub supplier: String,
    pub total_quantity: u32,
    pub remaining_quantity: u32,
    pub status: BatchStatus,
    pub created_at: DateTime<Utc>,
    pub recall_flag: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InboundRecord {
    pub id: Uuid,
    pub batch_id: Uuid,
    pub drug_name: String,
    pub batch_no: String,
    pub quantity: u32,
    pub supplier: String,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OutboundDetail {
    pub batch_id: Uuid,
    pub batch_no: String,
    pub quantity: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OutboundRecord {
    pub id: Uuid,
    pub drug_name: String,
    pub customer: String,
    pub details: Vec<OutboundDetail>,
    pub total_quantity: u32,
    pub has_warning: bool,
    pub warning_message: Option<String>,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RecallRecord {
    pub id: Uuid,
    pub batch_no: String,
    pub drug_name: String,
    pub supplier: String,
    pub inbound: Option<InboundRecord>,
    pub in_stock_batches: Vec<DrugBatch>,
    pub outbound_records: Vec<OutboundRecord>,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExpiryAlert {
    pub batch_no: String,
    pub drug_name: String,
    pub remaining_days: i64,
    pub remaining_quantity: u32,
    pub is_expired: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OutboundResult {
    pub record: OutboundRecord,
    pub alerts: Vec<ExpiryAlert>,
}
