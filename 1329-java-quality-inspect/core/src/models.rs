use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

use crate::aql::AqlLevel;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum InspectionStatus {
    Pending,
    Inspecting,
    Completed,
    Rejected,
    ReinspectionRequested,
    Reinspecting,
    FinalRejected,
    FinalAccepted,
}

impl InspectionStatus {
    pub fn as_str(&self) -> &'static str {
        match self {
            InspectionStatus::Pending => "pending",
            InspectionStatus::Inspecting => "inspecting",
            InspectionStatus::Completed => "completed",
            InspectionStatus::Rejected => "rejected",
            InspectionStatus::ReinspectionRequested => "reinspection_requested",
            InspectionStatus::Reinspecting => "reinspecting",
            InspectionStatus::FinalRejected => "final_rejected",
            InspectionStatus::FinalAccepted => "final_accepted",
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InspectionOrder {
    pub id: Uuid,
    pub product_name: String,
    pub batch_size: u32,
    pub sample_size: u32,
    pub aql_level: AqlLevel,
    pub ac: u32,
    pub status: InspectionStatus,
    pub inspector: String,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
    pub is_reinspection: bool,
    pub parent_order_id: Option<Uuid>,
    pub defect_count: u32,
    pub inspection_count: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DefectRecord {
    pub id: Uuid,
    pub order_id: Uuid,
    pub defect_type: String,
    pub description: String,
    pub sample_item_number: u32,
    pub recorded_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateOrderRequest {
    pub product_name: String,
    pub batch_size: u32,
    pub aql_level: String,
    pub inspector: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RecordDefectRequest {
    pub order_id: Uuid,
    pub defect_type: String,
    pub description: String,
    pub sample_item_number: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UpdateDefectRequest {
    pub defect_id: Uuid,
    pub new_description: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ReinspectionRequest {
    pub order_id: Uuid,
    pub inspector: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CompleteInspectionRequest {
    pub order_id: Uuid,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InspectItemRequest {
    pub order_id: Uuid,
    pub item_number: u32,
    pub is_defective: bool,
    pub defect_type: Option<String>,
    pub defect_description: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InspectionResult {
    pub accepted: bool,
    pub defect_count: u32,
    pub ac: u32,
    pub sample_size: u32,
    pub is_final: bool,
}
