use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Vehicle {
    pub id: String,
    pub plate_number: String,
    pub model: String,
    pub actual_value: Decimal,
}

impl Vehicle {
    pub fn new(plate_number: String, model: String, actual_value: Decimal) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            plate_number,
            model,
            actual_value,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub enum RepairItemType {
    PartReplacement,
    Labor,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RepairItem {
    pub id: String,
    pub name: String,
    pub item_type: RepairItemType,
    pub cost: Decimal,
}

impl RepairItem {
    pub fn new(name: String, item_type: RepairItemType, cost: Decimal) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            name,
            item_type,
            cost,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub enum ClaimStatus {
    Pending,
    TotalLoss,
    Settled,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Claim {
    pub id: String,
    pub vehicle_id: String,
    pub status: ClaimStatus,
    pub repair_items: Vec<RepairItem>,
    pub salvage_value: Option<Decimal>,
    pub payout: Option<Decimal>,
}

impl Claim {
    pub fn new(vehicle_id: String) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            vehicle_id,
            status: ClaimStatus::Pending,
            repair_items: Vec::new(),
            salvage_value: None,
            payout: None,
        }
    }

    pub fn total_repair_cost(&self) -> Decimal {
        self.repair_items
            .iter()
            .fold(Decimal::ZERO, |acc, item| acc + item.cost)
    }
}

pub const TOTAL_LOSS_THRESHOLD_PERCENT: u64 = 80;
