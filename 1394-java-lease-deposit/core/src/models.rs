use serde::{Deserialize, Serialize};
use uuid::Uuid;
use chrono::{DateTime, Local, NaiveDate};
use rust_decimal::Decimal;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ContractStatus {
    Pending,
    Active,
    MonthToMonth,
    Terminated,
    Expired,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ContractType {
    FixedTerm,
    MonthToMonth,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Tenant {
    pub id: Uuid,
    pub name: String,
    pub phone: String,
    pub id_card: Option<String>,
    pub created_at: DateTime<Local>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Room {
    pub id: Uuid,
    pub room_number: String,
    pub area: Decimal,
    pub default_monthly_rent: Decimal,
    pub is_available: bool,
    pub created_at: DateTime<Local>,
    pub updated_at: DateTime<Local>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Contract {
    pub id: Uuid,
    pub room_id: Uuid,
    pub tenant_id: Uuid,
    pub tenant_name: String,
    pub start_date: NaiveDate,
    pub end_date: NaiveDate,
    pub monthly_rent: Decimal,
    pub deposit_amount: Decimal,
    pub deposit_paid: bool,
    pub first_month_rent_paid: bool,
    pub status: ContractStatus,
    pub contract_type: ContractType,
    pub created_at: DateTime<Local>,
    pub updated_at: DateTime<Local>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CheckoutResult {
    pub contract_id: Uuid,
    pub room_id: Uuid,
    pub tenant_id: Uuid,
    pub checkout_date: NaiveDate,
    pub total_damage_fee: Decimal,
    pub early_termination_fee: Decimal,
    pub deposit_refund: Decimal,
    pub additional_payment: Decimal,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RenewalReminder {
    pub contract_id: Uuid,
    pub room_number: String,
    pub tenant_name: String,
    pub end_date: NaiveDate,
    pub days_remaining: i64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateTenantRequest {
    pub name: String,
    pub phone: String,
    pub id_card: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateRoomRequest {
    pub room_number: String,
    pub area: Decimal,
    pub default_monthly_rent: Decimal,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateContractRequest {
    pub room_id: Uuid,
    pub tenant_id: Uuid,
    pub start_date: NaiveDate,
    pub end_date: NaiveDate,
    pub monthly_rent: Option<Decimal>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CheckoutRequest {
    pub contract_id: Uuid,
    pub checkout_date: NaiveDate,
    pub damage_fee: Option<Decimal>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RenewContractRequest {
    pub contract_id: Uuid,
    pub new_end_date: NaiveDate,
    pub new_monthly_rent: Option<Decimal>,
}
