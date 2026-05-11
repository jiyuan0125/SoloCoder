use chrono::NaiveDate;
use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum Currency {
    USD,
    EUR,
    JPY,
    GBP,
    CNY,
}

impl Currency {
    pub fn all() -> &'static [Self] {
        &[Self::USD, Self::EUR, Self::JPY, Self::GBP, Self::CNY]
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExchangeRate {
    pub date: NaiveDate,
    pub currency: Currency,
    pub rate_to_cny: Decimal,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum DeclarationStatus {
    Draft,
    Merged,
    Declared,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DeclarationItem {
    pub hs_code: String,
    pub name: String,
    pub quantity: Decimal,
    pub unit_price: Decimal,
    pub currency: Currency,
}

impl DeclarationItem {
    pub fn amount(&self) -> Decimal {
        self.quantity * self.unit_price
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Declaration {
    pub id: Uuid,
    pub declaration_no: String,
    pub items: Vec<DeclarationItem>,
    pub status: DeclarationStatus,
    pub declare_date: Option<NaiveDate>,
    pub total_cny: Option<Decimal>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateDeclarationRequest {
    pub declaration_no: String,
    pub items: Vec<DeclarationItem>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UpdateDeclarationRequest {
    pub declaration_no: Option<String>,
    pub items: Option<Vec<DeclarationItem>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateExchangeRateRequest {
    pub date: NaiveDate,
    pub currency: Currency,
    pub rate_to_cny: Decimal,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MergeDeclarationRequest {
    pub declaration_ids: Vec<Uuid>,
    pub new_declaration_no: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DeclareRequest {
    pub declaration_id: Uuid,
    pub declare_date: NaiveDate,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DeclarationSummary {
    pub id: Uuid,
    pub declaration_no: String,
    pub item_count: usize,
    pub status: DeclarationStatus,
    pub declare_date: Option<NaiveDate>,
    pub total_cny: Option<Decimal>,
}

impl Declaration {
    pub fn summary(&self) -> DeclarationSummary {
        DeclarationSummary {
            id: self.id,
            declaration_no: self.declaration_no.clone(),
            item_count: self.items.len(),
            status: self.status,
            declare_date: self.declare_date,
            total_cny: self.total_cny,
        }
    }

    pub fn currencies(&self) -> Vec<Currency> {
        let mut map: HashMap<Currency, ()> = HashMap::new();
        for item in &self.items {
            map.insert(item.currency, ());
        }
        map.into_keys().collect()
    }
}
