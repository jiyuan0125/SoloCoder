use chrono::{DateTime, Utc};
use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct Product {
    pub id: String,
    pub name: String,
    pub created_at: DateTime<Utc>,
}

impl Product {
    pub fn new(name: String) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            name,
            created_at: Utc::now(),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct Supplier {
    pub id: String,
    pub name: String,
    pub created_at: DateTime<Utc>,
}

impl Supplier {
    pub fn new(name: String) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            name,
            created_at: Utc::now(),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct PurchaseBatch {
    pub id: String,
    pub product_id: String,
    pub supplier_id: String,
    pub batch_number: String,
    pub purchase_price: Decimal,
    pub total_quantity: u32,
    pub remaining_quantity: u32,
    pub created_at: DateTime<Utc>,
}

impl PurchaseBatch {
    pub fn new(
        product_id: String,
        supplier_id: String,
        batch_number: String,
        purchase_price: Decimal,
        quantity: u32,
    ) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            product_id,
            supplier_id,
            batch_number,
            purchase_price,
            total_quantity: quantity,
            remaining_quantity: quantity,
            created_at: Utc::now(),
        }
    }

    pub fn can_return(&self, quantity: u32) -> bool {
        quantity > 0 && quantity <= self.remaining_quantity
    }

    pub fn return_quantity(&mut self, quantity: u32) {
        self.remaining_quantity = self.remaining_quantity.saturating_sub(quantity);
    }
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct ReturnRecord {
    pub id: String,
    pub batch_id: String,
    pub product_id: String,
    pub supplier_id: String,
    pub batch_number: String,
    pub quantity: u32,
    pub unit_price: Decimal,
    pub total_amount: Decimal,
    pub created_at: DateTime<Utc>,
}

impl ReturnRecord {
    pub fn new(batch: &PurchaseBatch, quantity: u32) -> Self {
        let total_amount = batch.purchase_price * Decimal::from(quantity);
        Self {
            id: Uuid::new_v4().to_string(),
            batch_id: batch.id.clone(),
            product_id: batch.product_id.clone(),
            supplier_id: batch.supplier_id.clone(),
            batch_number: batch.batch_number.clone(),
            quantity,
            unit_price: batch.purchase_price,
            total_amount,
            created_at: Utc::now(),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum InventoryItemStatus {
    Normal,
    ExchangeReturned,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct InventoryItem {
    pub id: String,
    pub product_id: String,
    pub quantity: u32,
    pub status: InventoryItemStatus,
    pub created_at: DateTime<Utc>,
}

impl InventoryItem {
    pub fn new(product_id: String, quantity: u32, status: InventoryItemStatus) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            product_id,
            quantity,
            status,
            created_at: Utc::now(),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct ExchangeRecord {
    pub id: String,
    pub old_product_id: String,
    pub old_quantity: u32,
    pub old_unit_price: Decimal,
    pub old_total: Decimal,
    pub new_product_id: String,
    pub new_quantity: u32,
    pub new_unit_price: Decimal,
    pub new_total: Decimal,
    pub price_difference: Decimal,
    pub created_at: DateTime<Utc>,
}

impl ExchangeRecord {
    pub fn new(
        old_product_id: String,
        old_quantity: u32,
        old_unit_price: Decimal,
        new_product_id: String,
        new_quantity: u32,
        new_unit_price: Decimal,
    ) -> Self {
        let old_total = old_unit_price * Decimal::from(old_quantity);
        let new_total = new_unit_price * Decimal::from(new_quantity);
        let price_difference = new_total - old_total;

        Self {
            id: Uuid::new_v4().to_string(),
            old_product_id,
            old_quantity,
            old_unit_price,
            old_total,
            new_product_id,
            new_quantity,
            new_unit_price,
            new_total,
            price_difference,
            created_at: Utc::now(),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateProductRequest {
    pub name: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateSupplierRequest {
    pub name: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreatePurchaseRequest {
    pub product_id: String,
    pub supplier_id: String,
    pub batch_number: String,
    pub price: String,
    pub quantity: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateReturnRequest {
    pub batch_id: String,
    pub quantity: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateExchangeRequest {
    pub old_product_id: String,
    pub old_quantity: u32,
    pub new_product_id: String,
    pub new_quantity: u32,
}
