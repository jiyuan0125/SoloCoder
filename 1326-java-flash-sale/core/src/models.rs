use serde::{Deserialize, Serialize};
use chrono::{DateTime, Utc};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum OrderStatus {
    Created,
    Paid,
    Cancelled,
    Ended,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FlashSaleActivity {
    pub id: String,
    pub product_name: String,
    pub flash_price: f64,
    pub stock: u32,
    pub start_time: DateTime<Utc>,
    pub end_time: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Order {
    pub id: String,
    pub activity_id: String,
    pub user_id: String,
    pub status: OrderStatus,
    pub created_at: DateTime<Utc>,
    pub expires_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ActivityStatistics {
    pub activity_id: String,
    pub product_name: String,
    pub total_orders: u32,
    pub sold_count: u32,
    pub available_stock: u32,
}

impl FlashSaleActivity {
    pub fn new(
        product_name: String,
        flash_price: f64,
        stock: u32,
        start_time: DateTime<Utc>,
        end_time: DateTime<Utc>,
    ) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            product_name,
            flash_price,
            stock,
            start_time,
            end_time,
        }
    }
}

impl Order {
    pub fn new(activity_id: String, user_id: String, timeout_seconds: u32) -> Self {
        let created_at = Utc::now();
        let expires_at = created_at + chrono::Duration::seconds(timeout_seconds as i64);
        Self {
            id: Uuid::new_v4().to_string(),
            activity_id,
            user_id,
            status: OrderStatus::Created,
            created_at,
            expires_at,
        }
    }
}
