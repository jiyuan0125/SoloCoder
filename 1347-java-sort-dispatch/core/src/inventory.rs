use std::collections::HashMap;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InventoryRecord {
    pub warehouse_id: String,
    pub product_id: String,
    pub available_stock: u32,
    pub reserved_stock: u32,
}

#[derive(Debug, Clone)]
pub struct InventoryManager {
    inventory: HashMap<String, HashMap<String, InventoryRecord>>,
}

impl InventoryManager {
    pub fn new() -> Self {
        Self {
            inventory: HashMap::new(),
        }
    }

    pub fn add_stock(&mut self, warehouse_id: &str, product_id: &str, quantity: u32) {
        let warehouse_inventory = self
            .inventory
            .entry(warehouse_id.to_string())
            .or_default();

        let record = warehouse_inventory
            .entry(product_id.to_string())
            .or_insert_with(|| InventoryRecord {
                warehouse_id: warehouse_id.to_string(),
                product_id: product_id.to_string(),
                available_stock: 0,
                reserved_stock: 0,
            });

        record.available_stock += quantity;
    }

    pub fn has_stock(&self, warehouse_id: &str, product_id: &str, quantity: u32) -> bool {
        self.inventory
            .get(warehouse_id)
            .and_then(|w| w.get(product_id))
            .map(|r| r.available_stock >= quantity)
            .unwrap_or(false)
    }

    pub fn reserve_stock(
        &mut self,
        warehouse_id: &str,
        product_id: &str,
        quantity: u32,
    ) -> Result<(), String> {
        let warehouse_inventory = self
            .inventory
            .get_mut(warehouse_id)
            .ok_or_else(|| format!("Warehouse {} not found", warehouse_id))?;

        let record = warehouse_inventory
            .get_mut(product_id)
            .ok_or_else(|| format!("Product {} not found in warehouse {}", product_id, warehouse_id))?;

        if record.available_stock < quantity {
            return Err(format!(
                "Insufficient stock for product {} in warehouse {}: available {}, required {}",
                product_id, warehouse_id, record.available_stock, quantity
            ));
        }

        record.available_stock -= quantity;
        record.reserved_stock += quantity;

        Ok(())
    }

    pub fn confirm_shipment(
        &mut self,
        warehouse_id: &str,
        product_id: &str,
        quantity: u32,
    ) -> Result<(), String> {
        let warehouse_inventory = self
            .inventory
            .get_mut(warehouse_id)
            .ok_or_else(|| format!("Warehouse {} not found", warehouse_id))?;

        let record = warehouse_inventory
            .get_mut(product_id)
            .ok_or_else(|| format!("Product {} not found in warehouse {}", product_id, warehouse_id))?;

        if record.reserved_stock < quantity {
            return Err(format!(
                "Insufficient reserved stock for product {} in warehouse {}: reserved {}, required {}",
                product_id, warehouse_id, record.reserved_stock, quantity
            ));
        }

        record.reserved_stock -= quantity;

        Ok(())
    }

    pub fn release_reserved(
        &mut self,
        warehouse_id: &str,
        product_id: &str,
        quantity: u32,
    ) -> Result<(), String> {
        let warehouse_inventory = self
            .inventory
            .get_mut(warehouse_id)
            .ok_or_else(|| format!("Warehouse {} not found", warehouse_id))?;

        let record = warehouse_inventory
            .get_mut(product_id)
            .ok_or_else(|| format!("Product {} not found in warehouse {}", product_id, warehouse_id))?;

        if record.reserved_stock < quantity {
            return Err(format!(
                "Insufficient reserved stock for product {} in warehouse {}: reserved {}, required {}",
                product_id, warehouse_id, record.reserved_stock, quantity
            ));
        }

        record.reserved_stock -= quantity;
        record.available_stock += quantity;

        Ok(())
    }

    pub fn get_inventory(&self, warehouse_id: &str, product_id: &str) -> Option<InventoryRecord> {
        self.inventory
            .get(warehouse_id)
            .and_then(|w| w.get(product_id))
            .cloned()
    }

    pub fn get_all_inventory(&self) -> Vec<InventoryRecord> {
        self.inventory
            .values()
            .flat_map(|w| w.values().cloned())
            .collect()
    }
}

impl Default for InventoryManager {
    fn default() -> Self {
        Self::new()
    }
}
