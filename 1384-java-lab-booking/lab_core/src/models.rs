use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum UserRole {
    Researcher,
    Admin,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct User {
    pub id: Uuid,
    pub name: String,
    pub role: UserRole,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Equipment {
    pub id: Uuid,
    pub name: String,
    pub description: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Consumable {
    pub id: Uuid,
    pub name: String,
    pub stock: u32,
    pub safety_stock: u32,
}

impl Consumable {
    pub fn is_below_safety(&self) -> bool {
        self.stock < self.safety_stock
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ConsumableUsage {
    pub consumable_id: Uuid,
    pub quantity: u32,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum BookingStatus {
    Pending,
    Confirmed,
    InProgress,
    Completed,
    Cancelled,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Booking {
    pub id: Uuid,
    pub equipment_id: Uuid,
    pub researcher_id: Uuid,
    pub start_time: DateTime<Utc>,
    pub end_time: DateTime<Utc>,
    pub status: BookingStatus,
    pub consumables: Vec<ConsumableUsage>,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateBookingRequest {
    pub equipment_id: Uuid,
    pub researcher_id: Uuid,
    pub start_time: DateTime<Utc>,
    pub end_time: DateTime<Utc>,
    pub consumables: Vec<ConsumableUsage>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateEquipmentRequest {
    pub name: String,
    pub description: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateConsumableRequest {
    pub name: String,
    pub initial_stock: u32,
    pub safety_stock: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UpdateConsumableStockRequest {
    pub quantity: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateUserRequest {
    pub name: String,
    pub role: UserRole,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LowStockAlert {
    pub consumable_id: Uuid,
    pub consumable_name: String,
    pub current_stock: u32,
    pub safety_stock: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ConfirmBookingRequest {
    pub admin_id: Uuid,
}
