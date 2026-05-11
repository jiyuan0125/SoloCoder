use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Product {
    pub id: String,
    pub name: String,
    pub base_retail_price: f64,
    pub stock: i64,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

impl Product {
    pub fn new(name: String, base_retail_price: f64, stock: i64) -> Self {
        let now = Utc::now();
        Self {
            id: Uuid::new_v4().to_string(),
            name,
            base_retail_price,
            stock,
            created_at: now,
            updated_at: now,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TierPrice {
    pub id: String,
    pub product_id: String,
    pub min_quantity: i64,
    pub max_quantity: Option<i64>,
    pub discount_percent: f64,
    pub start_time: DateTime<Utc>,
    pub end_time: DateTime<Utc>,
    pub is_active: bool,
}

impl TierPrice {
    pub fn new(
        product_id: String,
        min_quantity: i64,
        max_quantity: Option<i64>,
        discount_percent: f64,
        start_time: DateTime<Utc>,
        end_time: DateTime<Utc>,
    ) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            product_id,
            min_quantity,
            max_quantity,
            discount_percent,
            start_time,
            end_time,
            is_active: true,
        }
    }

    pub fn is_valid_at(&self, time: DateTime<Utc>) -> bool {
        self.is_active && time >= self.start_time && time <= self.end_time
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Member {
    pub id: String,
    pub name: String,
    pub member_type: MemberType,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
pub enum MemberType {
    Normal,
    Gold,
    Platinum,
}

impl Member {
    pub fn new(name: String, member_type: MemberType) -> Self {
        let now = Utc::now();
        Self {
            id: Uuid::new_v4().to_string(),
            name,
            member_type,
            created_at: now,
            updated_at: now,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MemberDiscount {
    pub id: String,
    pub member_type: MemberType,
    pub product_id: Option<String>,
    pub discount_percent: f64,
    pub start_time: DateTime<Utc>,
    pub end_time: DateTime<Utc>,
    pub is_active: bool,
}

impl MemberDiscount {
    pub fn new(
        member_type: MemberType,
        product_id: Option<String>,
        discount_percent: f64,
        start_time: DateTime<Utc>,
        end_time: DateTime<Utc>,
    ) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            member_type,
            product_id,
            discount_percent,
            start_time,
            end_time,
            is_active: true,
        }
    }

    pub fn is_valid_at(&self, time: DateTime<Utc>) -> bool {
        self.is_active && time >= self.start_time && time <= self.end_time
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SpecialPrice {
    pub id: String,
    pub product_id: String,
    pub special_price: f64,
    pub start_time: DateTime<Utc>,
    pub end_time: DateTime<Utc>,
    pub is_active: bool,
}

impl SpecialPrice {
    pub fn new(
        product_id: String,
        special_price: f64,
        start_time: DateTime<Utc>,
        end_time: DateTime<Utc>,
    ) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            product_id,
            special_price,
            start_time,
            end_time,
            is_active: true,
        }
    }

    pub fn is_valid_at(&self, time: DateTime<Utc>) -> bool {
        self.is_active && time >= self.start_time && time <= self.end_time
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FullDiscount {
    pub id: String,
    pub threshold: f64,
    pub discount: f64,
    pub start_time: DateTime<Utc>,
    pub end_time: DateTime<Utc>,
    pub is_active: bool,
}

impl FullDiscount {
    pub fn new(
        threshold: f64,
        discount: f64,
        start_time: DateTime<Utc>,
        end_time: DateTime<Utc>,
    ) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            threshold,
            discount,
            start_time,
            end_time,
            is_active: true,
        }
    }

    pub fn is_valid_at(&self, time: DateTime<Utc>) -> bool {
        self.is_active && time >= self.start_time && time <= self.end_time
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ApprovalStatus {
    Pending,
    Approved,
    Rejected,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum PriceChangeType {
    Increase,
    Decrease,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApprovalRequest {
    pub id: String,
    pub target_id: String,
    pub target_type: ApprovalTargetType,
    pub change_type: PriceChangeType,
    pub change_percent: f64,
    pub old_price: f64,
    pub new_price: f64,
    pub status: ApprovalStatus,
    pub created_at: DateTime<Utc>,
    pub approved_at: Option<DateTime<Utc>>,
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize)]
pub enum ApprovalTargetType {
    ProductBasePrice,
    TierPrice,
    SpecialPrice,
}

impl ApprovalRequest {
    pub fn new(
        target_id: String,
        target_type: ApprovalTargetType,
        change_type: PriceChangeType,
        change_percent: f64,
        old_price: f64,
        new_price: f64,
    ) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            target_id,
            target_type,
            change_type,
            change_percent,
            old_price,
            new_price,
            status: ApprovalStatus::Pending,
            created_at: Utc::now(),
            approved_at: None,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum PriceType {
    Special,
    Tier,
    Member,
    Base,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PriceResult {
    pub product_id: String,
    pub quantity: i64,
    pub unit_price: f64,
    pub total_price: f64,
    pub price_type: PriceType,
    pub rule_id: Option<String>,
    pub is_special: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OrderItem {
    pub product_id: String,
    pub quantity: i64,
    pub unit_price: f64,
    pub total_price: f64,
    pub price_type: PriceType,
    pub is_special: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum OrderStatus {
    Created,
    Paid,
    Shipped,
    Completed,
    Cancelled,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Order {
    pub id: String,
    pub member_id: Option<String>,
    pub items: Vec<OrderItem>,
    pub subtotal: f64,
    pub full_discount_amount: f64,
    pub total_amount: f64,
    pub status: OrderStatus,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

impl Order {
    pub fn new(
        member_id: Option<String>,
        items: Vec<OrderItem>,
        subtotal: f64,
        full_discount_amount: f64,
        total_amount: f64,
    ) -> Self {
        let now = Utc::now();
        Self {
            id: Uuid::new_v4().to_string(),
            member_id,
            items,
            subtotal,
            full_discount_amount,
            total_amount,
            status: OrderStatus::Created,
            created_at: now,
            updated_at: now,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateOrderRequest {
    pub member_id: Option<String>,
    pub items: Vec<OrderItemRequest>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OrderItemRequest {
    pub product_id: String,
    pub quantity: i64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PriceQueryRequest {
    pub product_id: String,
    pub quantity: i64,
    pub member_type: Option<MemberType>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FullDiscountResult {
    pub applied: bool,
    pub threshold: f64,
    pub discount: f64,
    pub eligible_amount: f64,
}
