use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Product {
    pub id: Uuid,
    pub name: String,
    pub original_price: f64,
    pub floor_price: f64,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub enum BargainStatus {
    Active,
    Success,
    Failed,
    Purchased,
    Abandoned,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BargainActivity {
    pub id: Uuid,
    pub product_id: Uuid,
    pub initiator_id: Uuid,
    pub current_price: f64,
    pub original_price: f64,
    pub floor_price: f64,
    pub status: BargainStatus,
    pub created_at: DateTime<Utc>,
    pub expires_at: DateTime<Utc>,
    pub bargain_count: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BargainRecord {
    pub id: Uuid,
    pub activity_id: Uuid,
    pub user_id: Uuid,
    pub bargain_amount: f64,
    pub price_before: f64,
    pub price_after: f64,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct User {
    pub id: Uuid,
    pub name: String,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateProductRequest {
    pub name: String,
    pub original_price: f64,
    pub floor_price: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StartBargainRequest {
    pub product_id: Uuid,
    pub user_id: Uuid,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BargainRequest {
    pub activity_id: Uuid,
    pub user_id: Uuid,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ActivityDetail {
    pub activity: BargainActivity,
    pub product: Product,
    pub records: Vec<BargainRecord>,
}
