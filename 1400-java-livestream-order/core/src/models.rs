use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
pub enum OrderStatus {
    PendingPayment,
    Paid,
    Shipped,
    Completed,
    Cancelled,
    Refunded,
    Returned,
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
pub enum LiveStatus {
    NotStarted,
    Live,
    Ended,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Product {
    pub id: Uuid,
    pub name: String,
    pub original_price: f64,
    pub live_price: f64,
    pub total_stock: u32,
    pub available_stock: u32,
    pub locked_stock: u32,
    pub purchase_limit: Option<u32>,
    pub is_on_sale: bool,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Order {
    pub id: Uuid,
    pub user_id: Uuid,
    pub product_id: Uuid,
    pub quantity: u32,
    pub unit_price: f64,
    pub total_amount: f64,
    pub status: OrderStatus,
    pub locked_until: DateTime<Utc>,
    pub created_at: DateTime<Utc>,
    pub paid_at: Option<DateTime<Utc>>,
    pub shipped_at: Option<DateTime<Utc>>,
    pub completed_at: Option<DateTime<Utc>>,
    pub cancelled_at: Option<DateTime<Utc>>,
    pub refunded_at: Option<DateTime<Utc>>,
    pub returned_at: Option<DateTime<Utc>>,
    pub refund_amount: Option<f64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct User {
    pub id: Uuid,
    pub name: String,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateOrderRequest {
    pub user_id: Uuid,
    pub product_id: Uuid,
    pub quantity: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateProductRequest {
    pub name: String,
    pub original_price: f64,
    pub live_price: f64,
    pub total_stock: u32,
    pub purchase_limit: Option<u32>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateUserRequest {
    pub name: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PayOrderRequest {
    pub order_id: Uuid,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CancelOrderRequest {
    pub order_id: Uuid,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RefundOrderRequest {
    pub order_id: Uuid,
    pub is_return: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UpdateProductStockRequest {
    pub product_id: Uuid,
    pub additional_stock: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SetProductSaleStatusRequest {
    pub product_id: Uuid,
    pub is_on_sale: bool,
}
