use serde::{Deserialize, Serialize};
use uuid::Uuid;
use chrono::{DateTime, Utc};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Warehouse {
    pub id: String,
    pub name: String,
    pub city: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Inventory {
    pub warehouse_id: String,
    pub product_id: String,
    pub product_name: String,
    pub available_quantity: u32,
    pub locked_quantity: u32,
}

impl Inventory {
    pub fn new(warehouse_id: String, product_id: String, product_name: String) -> Self {
        Self {
            warehouse_id,
            product_id,
            product_name,
            available_quantity: 0,
            locked_quantity: 0,
        }
    }
    
    pub fn total_quantity(&self) -> u32 {
        self.available_quantity + self.locked_quantity
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum TransferStatus {
    Created,
    Shipped,
    Received,
    PartialReceived,
    Returned,
    Cancelled,
}

impl std::fmt::Display for TransferStatus {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            TransferStatus::Created => write!(f, "Created"),
            TransferStatus::Shipped => write!(f, "Shipped"),
            TransferStatus::Received => write!(f, "Received"),
            TransferStatus::PartialReceived => write!(f, "PartialReceived"),
            TransferStatus::Returned => write!(f, "Returned"),
            TransferStatus::Cancelled => write!(f, "Cancelled"),
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum DiscrepancyReason {
    Lost,
    Damaged,
    Rejected,
}

impl std::fmt::Display for DiscrepancyReason {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            DiscrepancyReason::Lost => write!(f, "Lost"),
            DiscrepancyReason::Damaged => write!(f, "Damaged"),
            DiscrepancyReason::Rejected => write!(f, "Rejected"),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Discrepancy {
    pub reason: DiscrepancyReason,
    pub quantity: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TransferOrder {
    pub id: String,
    pub source_warehouse_id: String,
    pub target_warehouse_id: String,
    pub product_id: String,
    pub quantity: u32,
    pub received_quantity: Option<u32>,
    pub discrepancy: Option<Discrepancy>,
    pub status: TransferStatus,
    pub created_at: DateTime<Utc>,
    pub shipped_at: Option<DateTime<Utc>>,
    pub received_at: Option<DateTime<Utc>>,
}

impl TransferOrder {
    pub fn new(
        source_warehouse_id: String,
        target_warehouse_id: String,
        product_id: String,
        quantity: u32,
    ) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            source_warehouse_id,
            target_warehouse_id,
            product_id,
            quantity,
            received_quantity: None,
            discrepancy: None,
            status: TransferStatus::Created,
            created_at: Utc::now(),
            shipped_at: None,
            received_at: None,
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum InventoryChangeType {
    Add,
    Remove,
    Lock,
    Unlock,
    TransferIn,
    TransferOut,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InventoryLog {
    pub id: String,
    pub warehouse_id: String,
    pub product_id: String,
    pub change_type: InventoryChangeType,
    pub quantity: i32,
    pub previous_available: u32,
    pub previous_locked: u32,
    pub new_available: u32,
    pub new_locked: u32,
    pub reference_transfer_id: Option<String>,
    pub created_at: DateTime<Utc>,
}

impl InventoryLog {
    pub fn new(
        warehouse_id: String,
        product_id: String,
        change_type: InventoryChangeType,
        quantity: i32,
        previous_available: u32,
        previous_locked: u32,
        new_available: u32,
        new_locked: u32,
        reference_transfer_id: Option<String>,
    ) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            warehouse_id,
            product_id,
            change_type,
            quantity,
            previous_available,
            previous_locked,
            new_available,
            new_locked,
            reference_transfer_id,
            created_at: Utc::now(),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateTransferRequest {
    pub source_warehouse_id: String,
    pub target_warehouse_id: String,
    pub product_id: String,
    pub quantity: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ReceiveRequest {
    pub transfer_id: String,
    pub received_quantity: u32,
    pub discrepancy_reason: Option<DiscrepancyReason>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AddInventoryRequest {
    pub warehouse_id: String,
    pub product_id: String,
    pub product_name: String,
    pub quantity: u32,
}
