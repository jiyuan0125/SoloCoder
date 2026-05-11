use chrono::{Datelike, DateTime, Local, NaiveDate};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub enum UserRole {
    Employee,
    Admin,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Employee {
    pub id: String,
    pub name: String,
    pub role: UserRole,
    pub created_at: DateTime<Local>,
}

impl Employee {
    pub fn new(name: String, role: UserRole) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            name,
            role,
            created_at: Local::now(),
        }
    }

    pub fn is_admin(&self) -> bool {
        matches!(self.role, UserRole::Admin)
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AssetCategory {
    pub id: String,
    pub name: String,
    pub useful_life_months: u32,
    pub max_per_employee: u32,
    pub created_at: DateTime<Local>,
}

impl AssetCategory {
    pub fn new(name: String, useful_life_months: u32, max_per_employee: u32) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            name,
            useful_life_months,
            max_per_employee,
            created_at: Local::now(),
        }
    }
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
pub enum AssetStatus {
    Available,
    InUse,
    UnderRepair,
    Scrapped,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Asset {
    pub id: String,
    pub name: String,
    pub category_id: String,
    pub purchase_price: f64,
    pub purchase_date: NaiveDate,
    pub status: AssetStatus,
    pub current_holder_id: Option<String>,
    pub created_at: DateTime<Local>,
}

impl Asset {
    pub fn new(name: String, category_id: String, purchase_price: f64, purchase_date: NaiveDate) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            name,
            category_id,
            purchase_price,
            purchase_date,
            status: AssetStatus::Available,
            current_holder_id: None,
            created_at: Local::now(),
        }
    }

    pub fn is_available(&self) -> bool {
        self.status == AssetStatus::Available && self.current_holder_id.is_none()
    }

    pub fn calculate_depreciation(
        &self,
        useful_life_months: u32,
        current_date: NaiveDate,
    ) -> (f64, f64) {
        let salvage_value = self.purchase_price * 0.05;
        let depreciable_base = self.purchase_price - salvage_value;
        let monthly_depreciation = if useful_life_months > 0 {
            depreciable_base / useful_life_months as f64
        } else {
            0.0
        };

        let start_month = if let Some(next_month) = self.purchase_date
            .with_month(self.purchase_date.month() + 1)
            .or_else(|| {
                self.purchase_date
                    .with_year(self.purchase_date.year() + 1)
                    .and_then(|d| d.with_month(1))
            }) {
            next_month
        } else {
            self.purchase_date.clone()
        };

        let mut months_depreciated = 0i32;
        if current_date >= start_month {
            let year_diff = current_date.year() - start_month.year();
            let month_diff = current_date.month() as i32 - start_month.month() as i32;
            months_depreciated = year_diff * 12 + month_diff;
            months_depreciated = months_depreciated.max(0);
            months_depreciated = months_depreciated.min(useful_life_months as i32);
        }

        let accumulated_depreciation = monthly_depreciation * months_depreciated as f64;
        let current_net_value = self.purchase_price - accumulated_depreciation;
        let current_net_value = current_net_value.max(salvage_value);

        (accumulated_depreciation, current_net_value)
    }
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
pub enum ApplicationStatus {
    Pending,
    Approved,
    Rejected,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BorrowApplication {
    pub id: String,
    pub asset_id: String,
    pub applicant_id: String,
    pub status: ApplicationStatus,
    pub approved_by: Option<String>,
    pub reason: String,
    pub created_at: DateTime<Local>,
    pub processed_at: Option<DateTime<Local>>,
}

impl BorrowApplication {
    pub fn new(asset_id: String, applicant_id: String, reason: String) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            asset_id,
            applicant_id,
            status: ApplicationStatus::Pending,
            approved_by: None,
            reason,
            created_at: Local::now(),
            processed_at: None,
        }
    }
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
pub enum ReturnCondition {
    Good,
    Damaged,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ReturnRecord {
    pub id: String,
    pub asset_id: String,
    pub returned_by: String,
    pub verified_by: String,
    pub condition: ReturnCondition,
    pub damage_note: Option<String>,
    pub returned_at: DateTime<Local>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TransferRequest {
    pub id: String,
    pub asset_id: String,
    pub from_employee_id: String,
    pub to_employee_id: String,
    pub from_confirmed: bool,
    pub to_confirmed: bool,
    pub completed_at: Option<DateTime<Local>>,
    pub created_at: DateTime<Local>,
}

impl TransferRequest {
    pub fn new(asset_id: String, from_employee_id: String, to_employee_id: String) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            asset_id,
            from_employee_id,
            to_employee_id,
            from_confirmed: false,
            to_confirmed: false,
            completed_at: None,
            created_at: Local::now(),
        }
    }

    pub fn is_completed(&self) -> bool {
        self.from_confirmed && self.to_confirmed
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ResponsibilityHistory {
    pub id: String,
    pub asset_id: String,
    pub employee_id: String,
    pub action: String,
    pub action_at: DateTime<Local>,
    pub notes: Option<String>,
}

impl ResponsibilityHistory {
    pub fn new(asset_id: String, employee_id: String, action: String, notes: Option<String>) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            asset_id,
            employee_id,
            action,
            action_at: Local::now(),
            notes,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AssetDetail {
    pub asset: Asset,
    pub category_name: String,
    pub accumulated_depreciation: f64,
    pub current_net_value: f64,
    pub current_holder_name: Option<String>,
}
