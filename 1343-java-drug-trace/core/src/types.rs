use serde::{Serialize, Deserialize};
use chrono::DateTime;
use chrono::Utc;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InboundRequest {
    pub drug_name: String,
    pub batch_no: String,
    pub production_date: DateTime<Utc>,
    pub expiry_date: DateTime<Utc>,
    pub supplier: String,
    pub quantity: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OutboundRequest {
    pub drug_name: String,
    pub customer: String,
    pub quantity: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RecallRequest {
    pub batch_no: String,
}
