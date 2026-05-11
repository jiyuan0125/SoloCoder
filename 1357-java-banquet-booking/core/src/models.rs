use serde::{Deserialize, Serialize};
use chrono::{DateTime, Utc};
use rust_decimal::Decimal;
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum BanquetType {
    Wedding,
    Birthday,
    Graduation,
    Business,
}

impl BanquetType {
    pub fn min_charge_per_table(&self) -> Decimal {
        match self {
            BanquetType::Wedding => Decimal::new(5000, 0),
            BanquetType::Birthday => Decimal::new(3000, 0),
            BanquetType::Graduation => Decimal::new(2500, 0),
            BanquetType::Business => Decimal::new(4000, 0),
        }
    }

    pub fn as_str(&self) -> &'static str {
        match self {
            BanquetType::Wedding => "婚宴",
            BanquetType::Birthday => "寿宴",
            BanquetType::Graduation => "升学宴",
            BanquetType::Business => "商务宴",
        }
    }
}

impl std::str::FromStr for BanquetType {
    type Err = &'static str;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s.to_lowercase().as_str() {
            "wedding" | "婚宴" => Ok(BanquetType::Wedding),
            "birthday" | "寿宴" => Ok(BanquetType::Birthday),
            "graduation" | "升学宴" => Ok(BanquetType::Graduation),
            "business" | "商务宴" => Ok(BanquetType::Business),
            _ => Err("无效的宴会类型"),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MenuItem {
    pub id: Uuid,
    pub name: String,
    pub category: String,
    pub price: Decimal,
}

impl MenuItem {
    pub fn new(name: &str, category: &str, price: Decimal) -> Self {
        MenuItem {
            id: Uuid::new_v4(),
            name: name.to_string(),
            category: category.to_string(),
            price,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MenuSet {
    pub id: Uuid,
    pub name: String,
    pub banquet_type: BanquetType,
    pub items: Vec<MenuItem>,
    pub price_per_table: Decimal,
}

impl MenuSet {
    pub fn new(name: &str, banquet_type: BanquetType, items: Vec<MenuItem>, price_per_table: Decimal) -> Self {
        MenuSet {
            id: Uuid::new_v4(),
            name: name.to_string(),
            banquet_type,
            items,
            price_per_table,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Booking {
    pub id: Uuid,
    pub customer_name: String,
    pub customer_phone: String,
    pub banquet_type: BanquetType,
    pub menu_set_id: Uuid,
    pub booked_tables: u32,
    pub event_date: DateTime<Utc>,
    pub deposit: Decimal,
    pub created_at: DateTime<Utc>,
    pub status: BookingStatus,
    pub actual_tables: Option<u32>,
    pub cancelled_at: Option<DateTime<Utc>>,
    pub refund_amount: Option<Decimal>,
    pub penalty_amount: Option<Decimal>,
    pub unused_option: Option<UnusedOption>,
    pub substitutions: Vec<Substitution>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum BookingStatus {
    Confirmed,
    Completed,
    Cancelled,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum UnusedOption {
    Pack,
    Refund,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Substitution {
    pub original_item_id: Uuid,
    pub replacement_item_id: Uuid,
    pub original_name: String,
    pub replacement_name: String,
    pub price_difference: Decimal,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateBookingRequest {
    pub customer_name: String,
    pub customer_phone: String,
    pub banquet_type: BanquetType,
    pub menu_set_id: Uuid,
    pub booked_tables: u32,
    pub event_date: DateTime<Utc>,
    pub deposit: Decimal,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AdjustTablesRequest {
    pub booking_id: Uuid,
    pub new_tables: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SubstituteDishRequest {
    pub booking_id: Uuid,
    pub original_item_id: Uuid,
    pub replacement_item_id: Uuid,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CompleteBookingRequest {
    pub booking_id: Uuid,
    pub actual_tables: u32,
    pub unused_option: UnusedOption,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CancelBookingRequest {
    pub booking_id: Uuid,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BookingSummary {
    pub booking_id: Uuid,
    pub customer_name: String,
    pub banquet_type: String,
    pub booked_tables: u32,
    pub event_date: DateTime<Utc>,
    pub status: BookingStatus,
    pub total_amount: Decimal,
    pub deposit: Decimal,
    pub balance: Decimal,
    pub penalty_amount: Option<Decimal>,
    pub refund_amount: Option<Decimal>,
}
