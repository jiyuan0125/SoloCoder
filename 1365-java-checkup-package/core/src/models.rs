use serde::{Serialize, Deserialize};
use uuid::Uuid;
use chrono::{NaiveDate, Datelike};

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum Gender {
    Male,
    Female,
    Unspecified,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum GenderRestriction {
    None,
    MaleOnly,
    FemaleOnly,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct CheckupItem {
    pub id: Uuid,
    pub name: String,
    pub price: f64,
    pub gender_restriction: GenderRestriction,
    pub has_contrast_agent: bool,
    pub is_gastroscopy: bool,
    pub is_colonoscopy: bool,
    pub is_xray: bool,
    pub is_pregnant_forbidden: bool,
}

impl CheckupItem {
    pub fn new(
        name: String,
        price: f64,
        gender_restriction: GenderRestriction,
        has_contrast_agent: bool,
        is_gastroscopy: bool,
        is_colonoscopy: bool,
        is_xray: bool,
        is_pregnant_forbidden: bool,
    ) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            price,
            gender_restriction,
            has_contrast_agent,
            is_gastroscopy,
            is_colonoscopy,
            is_xray,
            is_pregnant_forbidden,
        }
    }
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct CheckupPackage {
    pub id: Uuid,
    pub name: String,
    pub items: Vec<Uuid>,
    pub package_price: f64,
}

impl CheckupPackage {
    pub fn new(name: String, items: Vec<Uuid>, package_price: f64) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            items,
            package_price,
        }
    }
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Customer {
    pub id: Uuid,
    pub name: String,
    pub gender: Gender,
    pub birth_date: NaiveDate,
    pub is_pregnant: bool,
}

impl Customer {
    pub fn new(
        name: String,
        gender: Gender,
        birth_date: NaiveDate,
        is_pregnant: bool,
    ) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            gender,
            birth_date,
            is_pregnant,
        }
    }

    pub fn calculate_age(&self, appointment_date: NaiveDate) -> i32 {
        let mut age = appointment_date.year() - self.birth_date.year();
        
        if appointment_date.month() < self.birth_date.month() ||
           (appointment_date.month() == self.birth_date.month() && 
            appointment_date.day() < self.birth_date.day()) {
            age -= 1;
        }
        
        age
    }
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Reservation {
    pub id: Uuid,
    pub customer_id: Uuid,
    pub appointment_date: NaiveDate,
    pub base_package: Option<Uuid>,
    pub selected_items: Vec<Uuid>,
    pub removed_items: Vec<Uuid>,
    pub warnings: Vec<String>,
    pub requires_doctor_approval: Vec<Uuid>,
    pub total_price: f64,
    pub refund_amount: f64,
}

impl Reservation {
    pub fn new(
        customer_id: Uuid,
        appointment_date: NaiveDate,
        base_package: Option<Uuid>,
        selected_items: Vec<Uuid>,
        removed_items: Vec<Uuid>,
    ) -> Self {
        Self {
            id: Uuid::new_v4(),
            customer_id,
            appointment_date,
            base_package,
            selected_items,
            removed_items,
            warnings: Vec::new(),
            requires_doctor_approval: Vec::new(),
            total_price: 0.0,
            refund_amount: 0.0,
        }
    }
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum MutexRuleType {
    Warning,
    Hard,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct MutexRule {
    pub id: Uuid,
    pub rule_type: MutexRuleType,
    pub item1: Uuid,
    pub item2: Uuid,
    pub description: String,
}

impl MutexRule {
    pub fn new(
        rule_type: MutexRuleType,
        item1: Uuid,
        item2: Uuid,
        description: String,
    ) -> Self {
        Self {
            id: Uuid::new_v4(),
            rule_type,
            item1,
            item2,
            description,
        }
    }
}
