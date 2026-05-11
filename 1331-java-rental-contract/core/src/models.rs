use chrono::{NaiveDate, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub enum ContractStatus {
    Active,
    Expired,
    Terminated,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Property {
    pub id: Uuid,
    pub property_no: String,
    pub area: f64,
    pub monthly_rent: f64,
    pub created_at: chrono::DateTime<Utc>,
    pub updated_at: chrono::DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Tenant {
    pub id: Uuid,
    pub name: String,
    pub phone: String,
    pub created_at: chrono::DateTime<Utc>,
    pub updated_at: chrono::DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Contract {
    pub id: Uuid,
    pub tenant_id: Uuid,
    pub property_id: Uuid,
    pub start_date: NaiveDate,
    pub end_date: NaiveDate,
    pub monthly_rent: f64,
    pub deposit_amount: f64,
    pub status: ContractStatus,
    pub original_contract_id: Option<Uuid>,
    pub created_at: chrono::DateTime<Utc>,
    pub updated_at: chrono::DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CheckoutRecord {
    pub id: Uuid,
    pub contract_id: Uuid,
    pub actual_end_date: NaiveDate,
    pub damage_cost: f64,
    pub refund_amount: f64,
    pub outstanding_debt: f64,
    pub created_at: chrono::DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreatePropertyRequest {
    pub property_no: String,
    pub area: f64,
    pub monthly_rent: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateTenantRequest {
    pub name: String,
    pub phone: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateContractRequest {
    pub tenant_id: Uuid,
    pub property_id: Uuid,
    pub start_date: NaiveDate,
    pub end_date: NaiveDate,
    pub monthly_rent: Option<f64>,
    pub deposit_amount: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RenewContractRequest {
    pub contract_id: Uuid,
    pub new_monthly_rent: Option<f64>,
    pub new_deposit_amount: Option<f64>,
    pub extend_years: Option<i32>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CheckoutRequest {
    pub contract_id: Uuid,
    pub actual_end_date: Option<NaiveDate>,
    pub damage_cost: f64,
}

impl Property {
    pub fn new(property_no: String, area: f64, monthly_rent: f64) -> Self {
        let now = Utc::now();
        Self {
            id: Uuid::new_v4(),
            property_no,
            area,
            monthly_rent,
            created_at: now,
            updated_at: now,
        }
    }
}

impl Tenant {
    pub fn new(name: String, phone: String) -> Self {
        let now = Utc::now();
        Self {
            id: Uuid::new_v4(),
            name,
            phone,
            created_at: now,
            updated_at: now,
        }
    }
}

impl Contract {
    pub fn new(
        tenant_id: Uuid,
        property_id: Uuid,
        start_date: NaiveDate,
        end_date: NaiveDate,
        monthly_rent: f64,
        deposit_amount: f64,
        original_contract_id: Option<Uuid>,
    ) -> Self {
        let now = Utc::now();
        Self {
            id: Uuid::new_v4(),
            tenant_id,
            property_id,
            start_date,
            end_date,
            monthly_rent,
            deposit_amount,
            status: ContractStatus::Active,
            original_contract_id,
            created_at: now,
            updated_at: now,
        }
    }
}
