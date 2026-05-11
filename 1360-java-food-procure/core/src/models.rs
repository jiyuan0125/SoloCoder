use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;
use std::collections::HashMap;

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum SupplierStatus {
    Active,
    Observation,
    Disqualified,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Supplier {
    pub id: Uuid,
    pub name: String,
    pub contact: String,
    pub phone: String,
    pub status: SupplierStatus,
    pub fulfillment_rate: f64,
    pub monthly_scores: Vec<MonthlyScore>,
    pub ingredients: Vec<SupplierIngredient>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MonthlyScore {
    pub month: String,
    pub score: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SupplierIngredient {
    pub ingredient_id: Uuid,
    pub unit_price: f64,
    pub stock_quantity: f64,
    pub min_order_quantity: f64,
    pub quality_level: QualityLevel,
    pub delivery_time_hours: u32,
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, PartialOrd)]
pub enum QualityLevel {
    Premium = 5,
    High = 4,
    Standard = 3,
    Basic = 2,
    Low = 1,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Ingredient {
    pub id: Uuid,
    pub name: String,
    pub unit: String,
    pub category: String,
    pub weight_per_unit: f64,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Store {
    pub id: Uuid,
    pub name: String,
    pub address: String,
    pub contact: String,
    pub phone: String,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum PurchaseType {
    Normal,
    Emergency,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum PurchaseStatus {
    Draft,
    InQuotation,
    QuotationCompleted,
    OrderPlaced,
    InTransit,
    Delivered,
    Cancelled,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PurchaseRequest {
    pub id: Uuid,
    pub store_id: Uuid,
    pub purchase_type: PurchaseType,
    pub items: Vec<PurchaseItem>,
    pub status: PurchaseStatus,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PurchaseItem {
    pub id: Uuid,
    pub ingredient_id: Uuid,
    pub requested_quantity: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Quotation {
    pub id: Uuid,
    pub purchase_request_id: Uuid,
    pub supplier_id: Uuid,
    pub items: Vec<QuotationItem>,
    pub total_price: f64,
    pub total_weight: f64,
    pub shipping_fee: f64,
    pub delivery_time_hours: u32,
    pub composite_score: f64,
    pub is_recommended: bool,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct QuotationItem {
    pub ingredient_id: Uuid,
    pub unit_price: f64,
    pub available_quantity: f64,
    pub quality_level: QualityLevel,
    pub subtotal: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CompositeScoreBreakdown {
    pub price_score: f64,
    pub quality_score: f64,
    pub timeliness_score: f64,
    pub fulfillment_score: f64,
    pub total_score: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PurchaseOrder {
    pub id: Uuid,
    pub purchase_request_id: Uuid,
    pub quotation_id: Uuid,
    pub store_id: Uuid,
    pub supplier_id: Uuid,
    pub items: Vec<OrderItem>,
    pub total_price: f64,
    pub shipping_fee: f64,
    pub is_merged: bool,
    pub merged_with: Option<Vec<Uuid>>,
    pub selected_non_recommended: bool,
    pub non_recommended_reason: Option<String>,
    pub status: OrderStatus,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum OrderStatus {
    Pending,
    Confirmed,
    Shipped,
    Delivered,
    Cancelled,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OrderItem {
    pub ingredient_id: Uuid,
    pub quantity: f64,
    pub unit_price: f64,
    pub subtotal: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StoreMonthlyStats {
    pub store_id: Uuid,
    pub month: String,
    pub total_purchase_amount: f64,
    pub emergency_purchase_amount: f64,
    pub emergency_purchase_count: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InProgressLock {
    pub supplier_id: Uuid,
    pub ingredient_id: Uuid,
    pub locked_quantity: f64,
    pub lock_expires_at: DateTime<Utc>,
    pub order_ids: Vec<Uuid>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MergedOrderPool {
    pub pool_id: Uuid,
    pub supplier_id: Uuid,
    pub ingredient_id: Uuid,
    pub total_quantity: f64,
    pub store_orders: HashMap<Uuid, f64>,
    pub created_at: DateTime<Utc>,
    pub expires_at: DateTime<Utc>,
}
