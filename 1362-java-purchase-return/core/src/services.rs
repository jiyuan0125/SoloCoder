use std::collections::HashMap;
use std::sync::{Arc, Mutex};

use rust_decimal::Decimal;
use rust_decimal_macros::dec;

use crate::errors::*;
use crate::models::*;

#[derive(Clone, Default)]
pub struct InMemoryStore {
    inner: Arc<Mutex<StoreInner>>,
}

#[derive(Default)]
struct StoreInner {
    products: HashMap<String, Product>,
    suppliers: HashMap<String, Supplier>,
    batches: HashMap<String, PurchaseBatch>,
    returns: HashMap<String, ReturnRecord>,
    exchanges: HashMap<String, ExchangeRecord>,
    inventory: HashMap<String, InventoryItem>,
    product_prices: HashMap<String, Decimal>,
}

impl InMemoryStore {
    pub fn new() -> Self {
        Self::default()
    }
}

pub struct PurchaseReturnService {
    store: InMemoryStore,
}

impl PurchaseReturnService {
    pub fn new(store: InMemoryStore) -> Self {
        Self { store }
    }

    pub fn create_product(&self, name: String) -> Result<Product> {
        let product = Product::new(name);
        let mut inner = self.store.inner.lock().unwrap();
        inner.products.insert(product.id.clone(), product.clone());
        Ok(product)
    }

    pub fn list_products(&self) -> Vec<Product> {
        let inner = self.store.inner.lock().unwrap();
        inner.products.values().cloned().collect()
    }

    pub fn get_product(&self, id: &str) -> Result<Product> {
        let inner = self.store.inner.lock().unwrap();
        inner
            .products
            .get(id)
            .cloned()
            .ok_or_else(|| PurchaseReturnError::ProductNotFound(id.to_string()))
    }

    pub fn create_supplier(&self, name: String) -> Result<Supplier> {
        let supplier = Supplier::new(name);
        let mut inner = self.store.inner.lock().unwrap();
        inner.suppliers.insert(supplier.id.clone(), supplier.clone());
        Ok(supplier)
    }

    pub fn list_suppliers(&self) -> Vec<Supplier> {
        let inner = self.store.inner.lock().unwrap();
        inner.suppliers.values().cloned().collect()
    }

    pub fn get_supplier(&self, id: &str) -> Result<Supplier> {
        let inner = self.store.inner.lock().unwrap();
        inner
            .suppliers
            .get(id)
            .cloned()
            .ok_or_else(|| PurchaseReturnError::SupplierNotFound(id.to_string()))
    }

    pub fn create_purchase(
        &self,
        product_id: String,
        supplier_id: String,
        batch_number: String,
        price_str: String,
        quantity: u32,
    ) -> Result<PurchaseBatch> {
        if quantity == 0 {
            return Err(PurchaseReturnError::InvalidQuantity(
                "Quantity must be greater than 0".to_string(),
            ));
        }

        let price: Decimal = price_str
            .parse()
            .map_err(|_| PurchaseReturnError::InvalidPrice(price_str.clone()))?;

        if price <= dec!(0) {
            return Err(PurchaseReturnError::InvalidPrice(
                "Price must be greater than 0".to_string(),
            ));
        }

        let _ = self.get_product(&product_id)?;
        let _ = self.get_supplier(&supplier_id)?;

        let batch = PurchaseBatch::new(
            product_id.clone(),
            supplier_id,
            batch_number,
            price,
            quantity,
        );

        let mut inner = self.store.inner.lock().unwrap();
        inner.product_prices.insert(product_id, price);
        inner.batches.insert(batch.id.clone(), batch.clone());
        Ok(batch)
    }

    pub fn list_batches(&self) -> Vec<PurchaseBatch> {
        let inner = self.store.inner.lock().unwrap();
        inner.batches.values().cloned().collect()
    }

    pub fn get_batch(&self, id: &str) -> Result<PurchaseBatch> {
        let inner = self.store.inner.lock().unwrap();
        inner
            .batches
            .get(id)
            .cloned()
            .ok_or_else(|| PurchaseReturnError::BatchNotFound(id.to_string()))
    }

    pub fn create_return(&self, batch_id: String, quantity: u32) -> Result<ReturnRecord> {
        if quantity == 0 {
            return Err(PurchaseReturnError::InvalidQuantity(
                "Quantity must be greater than 0".to_string(),
            ));
        }

        let mut inner = self.store.inner.lock().unwrap();

        let batch = inner
            .batches
            .get_mut(&batch_id)
            .ok_or_else(|| PurchaseReturnError::BatchNotFound(batch_id.clone()))?;

        if !batch.can_return(quantity) {
            return Err(PurchaseReturnError::InsufficientQuantity {
                batch_id: batch.id.clone(),
                requested: quantity,
                available: batch.remaining_quantity,
            });
        }

        batch.return_quantity(quantity);
        let batch_clone = batch.clone();
        let return_record = ReturnRecord::new(&batch_clone, quantity);
        inner
            .returns
            .insert(return_record.id.clone(), return_record.clone());

        Ok(return_record)
    }

    pub fn list_returns(&self) -> Vec<ReturnRecord> {
        let inner = self.store.inner.lock().unwrap();
        inner.returns.values().cloned().collect()
    }

    pub fn get_return(&self, id: &str) -> Result<ReturnRecord> {
        let inner = self.store.inner.lock().unwrap();
        inner
            .returns
            .get(id)
            .cloned()
            .ok_or_else(|| PurchaseReturnError::ReturnNotFound(id.to_string()))
    }

    pub fn create_exchange(
        &self,
        old_product_id: String,
        old_quantity: u32,
        new_product_id: String,
        new_quantity: u32,
    ) -> Result<ExchangeRecord> {
        if old_product_id == new_product_id {
            return Err(PurchaseReturnError::SameProductExchange);
        }

        if old_quantity == 0 || new_quantity == 0 {
            return Err(PurchaseReturnError::InvalidQuantity(
                "Quantity must be greater than 0".to_string(),
            ));
        }

        let _ = self.get_product(&old_product_id)?;
        let _ = self.get_product(&new_product_id)?;

        let mut inner = self.store.inner.lock().unwrap();

        let old_price = inner
            .product_prices
            .get(&old_product_id)
            .cloned()
            .unwrap_or(dec!(0));

        let new_price = inner
            .product_prices
            .get(&new_product_id)
            .cloned()
            .unwrap_or(dec!(0));

        let exchange = ExchangeRecord::new(
            old_product_id.clone(),
            old_quantity,
            old_price,
            new_product_id.clone(),
            new_quantity,
            new_price,
        );

        let old_inventory = InventoryItem::new(old_product_id, old_quantity, InventoryItemStatus::ExchangeReturned);
        inner.inventory.insert(old_inventory.id.clone(), old_inventory);

        inner.exchanges.insert(exchange.id.clone(), exchange.clone());

        Ok(exchange)
    }

    pub fn list_exchanges(&self) -> Vec<ExchangeRecord> {
        let inner = self.store.inner.lock().unwrap();
        inner.exchanges.values().cloned().collect()
    }

    pub fn get_exchange(&self, id: &str) -> Result<ExchangeRecord> {
        let inner = self.store.inner.lock().unwrap();
        inner
            .exchanges
            .get(id)
            .cloned()
            .ok_or_else(|| PurchaseReturnError::ExchangeNotFound(id.to_string()))
    }
}
