use chrono::{DateTime, Utc, NaiveDate};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum Season {
    OffSeason,
    PeakSeason,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum TicketType {
    Adult,
    Child,
    Elder,
    Student,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ChildTicketCategory {
    Free,
    HalfPrice,
    FullPrice,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Scenic {
    pub id: Uuid,
    pub name: String,
    pub base_price: u32,
    pub daily_capacity: u32,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TicketItem {
    pub id: Uuid,
    pub ticket_type: TicketType,
    pub original_price: u32,
    pub actual_price: u32,
    
    pub id_card: Option<String>,
    pub name: Option<String>,
    
    pub child_height: Option<f32>,
    pub child_category: Option<ChildTicketCategory>,
    
    pub birth_date: Option<NaiveDate>,
    pub is_elder_eligible: Option<bool>,
    
    pub student_id: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Order {
    pub id: Uuid,
    pub scenic_id: Uuid,
    pub use_date: NaiveDate,
    pub tickets: Vec<TicketItem>,
    pub total_price: u32,
    pub created_at: DateTime<Utc>,
    pub is_refunded: bool,
    pub refunded_at: Option<DateTime<Utc>>,
    pub refund_amount: Option<u32>,
    pub refund_fee: Option<u32>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateOrderRequest {
    pub scenic_id: Uuid,
    pub use_date: NaiveDate,
    pub tickets: Vec<TicketRequest>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TicketRequest {
    pub ticket_type: TicketType,
    pub id_card: Option<String>,
    pub name: Option<String>,
    pub child_height: Option<f32>,
    pub student_id: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateOrderResponse {
    pub order: Order,
    pub use_count: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RefundResponse {
    pub order: Order,
    pub refund_amount: u32,
    pub refund_fee: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ScenicDailyStats {
    pub scenic_id: Uuid,
    pub date: NaiveDate,
    pub total_sold: u32,
    pub used_id_cards: Vec<String>,
}
