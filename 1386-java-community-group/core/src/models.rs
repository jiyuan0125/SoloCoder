use chrono::{DateTime, Utc, NaiveTime};
use serde::{Serialize, Deserialize};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PickupPoint {
    pub id: Uuid,
    pub name: String,
    pub cut_off_time: NaiveTime,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Product {
    pub id: Uuid,
    pub pickup_point_id: Uuid,
    pub name: String,
    pub unit_price: i64,
    pub stock: i64,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum OrderStatus {
    Created,
    Paid,
    Cancelled,
    Refunded,
    ReadyForPickup,
    PickedUp,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OrderItem {
    pub product_id: Uuid,
    pub product_name: String,
    pub unit_price: i64,
    pub quantity: i64,
    pub subtotal: i64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Order {
    pub id: Uuid,
    pub pickup_point_id: Uuid,
    pub user_id: Uuid,
    pub items: Vec<OrderItem>,
    pub total_amount: i64,
    pub status: OrderStatus,
    pub created_at: DateTime<Utc>,
    pub payment_deadline: DateTime<Utc>,
    pub cut_off_time: DateTime<Utc>,
    pub paid_at: Option<DateTime<Utc>>,
    pub picked_up_at: Option<DateTime<Utc>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SortingList {
    pub id: Uuid,
    pub pickup_point_id: Uuid,
    pub batch_date: DateTime<Utc>,
    pub items: Vec<SortingItem>,
    pub created_at: DateTime<Utc>,
    pub confirmed_at: Option<DateTime<Utc>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SortingItem {
    pub product_id: Uuid,
    pub product_name: String,
    pub total_quantity: i64,
    pub order_ids: Vec<Uuid>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateProductRequest {
    pub name: String,
    pub unit_price: i64,
    pub stock: i64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateOrderRequest {
    pub user_id: Uuid,
    pub items: Vec<CreateOrderItemRequest>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateOrderItemRequest {
    pub product_id: Uuid,
    pub quantity: i64,
}
