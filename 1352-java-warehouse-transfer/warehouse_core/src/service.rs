use std::collections::HashMap;
use std::sync::{Arc, Mutex};

use chrono::{Duration, Utc};
use uuid::Uuid;

use crate::error::WarehouseError;
use crate::models::*;

#[derive(Clone)]
pub struct WarehouseService {
    inner: Arc<Mutex<InnerService>>,
}

struct InnerService {
    warehouses: HashMap<String, Warehouse>,
    inventory: HashMap<(String, String), Inventory>,
    transfers: HashMap<String, TransferOrder>,
    inventory_logs: Vec<InventoryLog>,
}

impl WarehouseService {
    pub fn new() -> Self {
        Self {
            inner: Arc::new(Mutex::new(InnerService {
                warehouses: HashMap::new(),
                inventory: HashMap::new(),
                transfers: HashMap::new(),
                inventory_logs: Vec::new(),
            })),
        }
    }

    pub fn create_warehouse(&self, name: String, city: String) -> Warehouse {
        let mut inner = self.inner.lock().unwrap();
        let warehouse = Warehouse {
            id: Uuid::new_v4().to_string(),
            name,
            city,
        };
        inner.warehouses.insert(warehouse.id.clone(), warehouse.clone());
        warehouse
    }

    pub fn list_warehouses(&self) -> Vec<Warehouse> {
        let inner = self.inner.lock().unwrap();
        inner.warehouses.values().cloned().collect()
    }

    pub fn get_warehouse(&self, id: &str) -> Option<Warehouse> {
        let inner = self.inner.lock().unwrap();
        inner.warehouses.get(id).cloned()
    }

    pub fn add_inventory(
        &self,
        warehouse_id: &str,
        product_id: &str,
        product_name: &str,
        quantity: u32,
    ) -> Result<Inventory, WarehouseError> {
        let mut inner = self.inner.lock().unwrap();

        if !inner.warehouses.contains_key(warehouse_id) {
            return Err(WarehouseError::WarehouseNotFound(warehouse_id.to_string()));
        }

        let key = (warehouse_id.to_string(), product_id.to_string());
        let prev_available;
        let prev_locked;

        {
            let inventory = inner.inventory.entry(key.clone()).or_insert_with(|| {
                Inventory::new(
                    warehouse_id.to_string(),
                    product_id.to_string(),
                    product_name.to_string(),
                )
            });
            prev_available = inventory.available_quantity;
            prev_locked = inventory.locked_quantity;
            inventory.available_quantity += quantity;
        }

        let log = InventoryLog::new(
            warehouse_id.to_string(),
            product_id.to_string(),
            InventoryChangeType::Add,
            quantity as i32,
            prev_available,
            prev_locked,
            inner.inventory[&key].available_quantity,
            inner.inventory[&key].locked_quantity,
            None,
        );
        inner.inventory_logs.push(log);

        Ok(inner.inventory[&key].clone())
    }

    pub fn list_inventory(&self, warehouse_id: Option<&str>) -> Vec<Inventory> {
        let inner = self.inner.lock().unwrap();
        inner
            .inventory
            .values()
            .filter(|inv| warehouse_id.map_or(true, |w| inv.warehouse_id == w))
            .cloned()
            .collect()
    }

    pub fn get_inventory(&self, warehouse_id: &str, product_id: &str) -> Option<Inventory> {
        let inner = self.inner.lock().unwrap();
        let key = (warehouse_id.to_string(), product_id.to_string());
        inner.inventory.get(&key).cloned()
    }

    pub fn create_transfer(
        &self,
        source_warehouse_id: &str,
        target_warehouse_id: &str,
        product_id: &str,
        quantity: u32,
    ) -> Result<TransferOrder, WarehouseError> {
        if source_warehouse_id == target_warehouse_id {
            return Err(WarehouseError::SameWarehouse);
        }

        let mut inner = self.inner.lock().unwrap();

        if !inner.warehouses.contains_key(source_warehouse_id) {
            return Err(WarehouseError::WarehouseNotFound(
                source_warehouse_id.to_string(),
            ));
        }
        if !inner.warehouses.contains_key(target_warehouse_id) {
            return Err(WarehouseError::WarehouseNotFound(
                target_warehouse_id.to_string(),
            ));
        }

        Self::check_chain_transfer(&inner, target_warehouse_id, product_id)?;

        let src_key = (source_warehouse_id.to_string(), product_id.to_string());
        let src_inventory = inner
            .inventory
            .get_mut(&src_key)
            .ok_or_else(|| WarehouseError::ProductNotFound {
                product_id: product_id.to_string(),
            })?;

        if src_inventory.available_quantity < quantity {
            return Err(WarehouseError::InsufficientStock {
                available: src_inventory.available_quantity,
                requested: quantity,
            });
        }

        let prev_available = src_inventory.available_quantity;
        let prev_locked = src_inventory.locked_quantity;
        src_inventory.available_quantity -= quantity;
        src_inventory.locked_quantity += quantity;

        let transfer = TransferOrder::new(
            source_warehouse_id.to_string(),
            target_warehouse_id.to_string(),
            product_id.to_string(),
            quantity,
        );

        let log = InventoryLog::new(
            source_warehouse_id.to_string(),
            product_id.to_string(),
            InventoryChangeType::Lock,
            -(quantity as i32),
            prev_available,
            prev_locked,
            inner.inventory[&src_key].available_quantity,
            inner.inventory[&src_key].locked_quantity,
            Some(transfer.id.clone()),
        );
        inner.inventory_logs.push(log);

        inner
            .transfers
            .insert(transfer.id.clone(), transfer.clone());
        Ok(transfer)
    }

    fn check_chain_transfer(
        inner: &InnerService,
        warehouse_id: &str,
        product_id: &str,
    ) -> Result<(), WarehouseError> {
        for transfer in inner.transfers.values() {
            if transfer.target_warehouse_id == warehouse_id
                && transfer.product_id == product_id
                && (transfer.status == TransferStatus::Created
                    || transfer.status == TransferStatus::Shipped)
            {
                return Err(WarehouseError::ChainTransferNotAllowed {
                    warehouse_id: warehouse_id.to_string(),
                });
            }
        }
        Ok(())
    }

    pub fn ship_transfer(&self, transfer_id: &str) -> Result<TransferOrder, WarehouseError> {
        let mut inner = self.inner.lock().unwrap();
        let transfer = inner
            .transfers
            .get_mut(transfer_id)
            .ok_or_else(|| WarehouseError::TransferNotFound(transfer_id.to_string()))?;

        if transfer.status != TransferStatus::Created {
            return Err(WarehouseError::InvalidStatus {
                current: transfer.status.to_string(),
            });
        }

        transfer.status = TransferStatus::Shipped;
        transfer.shipped_at = Some(Utc::now());
        Ok(transfer.clone())
    }

    pub fn receive_transfer(
        &self,
        transfer_id: &str,
        received_quantity: u32,
        discrepancy_reason: Option<DiscrepancyReason>,
    ) -> Result<TransferOrder, WarehouseError> {
        if received_quantity == 0 {
            return self.return_transfer(transfer_id);
        }

        let mut inner = self.inner.lock().unwrap();

        let (
            source_warehouse_id,
            target_warehouse_id,
            product_id,
            transfer_quantity,
            transfer_status,
        ) = {
            let transfer = inner
                .transfers
                .get(transfer_id)
                .ok_or_else(|| WarehouseError::TransferNotFound(transfer_id.to_string()))?;
            (
                transfer.source_warehouse_id.clone(),
                transfer.target_warehouse_id.clone(),
                transfer.product_id.clone(),
                transfer.quantity,
                transfer.status,
            )
        };

        if transfer_status != TransferStatus::Created && transfer_status != TransferStatus::Shipped {
            return Err(WarehouseError::InvalidStatus {
                current: transfer_status.to_string(),
            });
        }

        if received_quantity > transfer_quantity {
            return Err(WarehouseError::InsufficientStock {
                available: transfer_quantity,
                requested: received_quantity,
            });
        }

        let src_key = (source_warehouse_id.clone(), product_id.clone());
        let target_key = (target_warehouse_id.clone(), product_id.clone());

        let product_name = inner
            .inventory
            .get(&src_key)
            .ok_or_else(|| WarehouseError::ProductNotFound {
                product_id: product_id.clone(),
            })?
            .product_name
            .clone();

        let (new_status, discrepancy) = if received_quantity == transfer_quantity {
            (TransferStatus::Received, None)
        } else {
            let unreturned = transfer_quantity - received_quantity;
            (
                TransferStatus::PartialReceived,
                Some(Discrepancy {
                    reason: discrepancy_reason.unwrap_or(DiscrepancyReason::Lost),
                    quantity: unreturned,
                }),
            )
        };

        let is_rejected = discrepancy
            .as_ref()
            .map(|d| d.reason == DiscrepancyReason::Rejected)
            .unwrap_or(false);

        let (
            prev_src_available,
            prev_src_locked,
            new_src_available,
            new_src_locked,
            prev_target_available,
            prev_target_locked,
            new_target_available,
            new_target_locked,
        ) = {
            let src_inventory = inner
                .inventory
                .get_mut(&src_key)
                .ok_or_else(|| WarehouseError::ProductNotFound {
                    product_id: product_id.clone(),
                })?;
            let prev_src_available = src_inventory.available_quantity;
            let prev_src_locked = src_inventory.locked_quantity;

            if received_quantity == transfer_quantity {
                src_inventory.locked_quantity -= transfer_quantity;
            } else if is_rejected {
                let unreturned = transfer_quantity - received_quantity;
                src_inventory.locked_quantity -= transfer_quantity;
                src_inventory.available_quantity += unreturned;
            } else {
                src_inventory.locked_quantity -= transfer_quantity;
            }

            let new_src_available = src_inventory.available_quantity;
            let new_src_locked = src_inventory.locked_quantity;

            let target_inventory =
                inner
                    .inventory
                    .entry(target_key.clone())
                    .or_insert_with(|| {
                        Inventory::new(
                            target_warehouse_id.clone(),
                            product_id.clone(),
                            product_name.clone(),
                        )
                    });
            let prev_target_available = target_inventory.available_quantity;
            let prev_target_locked = target_inventory.locked_quantity;
            target_inventory.available_quantity += received_quantity;

            (
                prev_src_available,
                prev_src_locked,
                new_src_available,
                new_src_locked,
                prev_target_available,
                prev_target_locked,
                target_inventory.available_quantity,
                target_inventory.locked_quantity,
            )
        };

        let src_log = InventoryLog::new(
            source_warehouse_id.clone(),
            product_id.clone(),
            InventoryChangeType::TransferOut,
            -(transfer_quantity as i32),
            prev_src_available,
            prev_src_locked,
            new_src_available,
            new_src_locked,
            Some(transfer_id.to_string()),
        );
        inner.inventory_logs.push(src_log);

        let target_log = InventoryLog::new(
            target_warehouse_id,
            product_id,
            InventoryChangeType::TransferIn,
            received_quantity as i32,
            prev_target_available,
            prev_target_locked,
            new_target_available,
            new_target_locked,
            Some(transfer_id.to_string()),
        );
        inner.inventory_logs.push(target_log);

        let transfer = inner
            .transfers
            .get_mut(transfer_id)
            .ok_or_else(|| WarehouseError::TransferNotFound(transfer_id.to_string()))?;
        transfer.status = new_status;
        transfer.discrepancy = discrepancy;
        transfer.received_quantity = Some(received_quantity);
        transfer.received_at = Some(Utc::now());

        Ok(transfer.clone())
    }

    pub fn return_transfer(&self, transfer_id: &str) -> Result<TransferOrder, WarehouseError> {
        let mut inner = self.inner.lock().unwrap();

        let (source_warehouse_id, product_id, quantity, transfer_status) = {
            let transfer = inner
                .transfers
                .get(transfer_id)
                .ok_or_else(|| WarehouseError::TransferNotFound(transfer_id.to_string()))?;
            (
                transfer.source_warehouse_id.clone(),
                transfer.product_id.clone(),
                transfer.quantity,
                transfer.status,
            )
        };

        if transfer_status != TransferStatus::Created && transfer_status != TransferStatus::Shipped {
            return Err(WarehouseError::InvalidStatus {
                current: transfer_status.to_string(),
            });
        }

        let src_key = (source_warehouse_id.clone(), product_id.clone());

        let (prev_available, prev_locked, new_available, new_locked) = {
            let src_inventory = inner
                .inventory
                .get_mut(&src_key)
                .ok_or_else(|| WarehouseError::ProductNotFound {
                    product_id: product_id.clone(),
                })?;
            let prev_available = src_inventory.available_quantity;
            let prev_locked = src_inventory.locked_quantity;
            src_inventory.locked_quantity -= quantity;
            src_inventory.available_quantity += quantity;
            (
                prev_available,
                prev_locked,
                src_inventory.available_quantity,
                src_inventory.locked_quantity,
            )
        };

        let log = InventoryLog::new(
            source_warehouse_id,
            product_id,
            InventoryChangeType::Unlock,
            quantity as i32,
            prev_available,
            prev_locked,
            new_available,
            new_locked,
            Some(transfer_id.to_string()),
        );
        inner.inventory_logs.push(log);

        let transfer = inner
            .transfers
            .get_mut(transfer_id)
            .ok_or_else(|| WarehouseError::TransferNotFound(transfer_id.to_string()))?;
        transfer.status = TransferStatus::Returned;
        transfer.received_at = Some(Utc::now());
        transfer.received_quantity = Some(0);

        Ok(transfer.clone())
    }

    pub fn cancel_transfer(&self, transfer_id: &str) -> Result<TransferOrder, WarehouseError> {
        let mut inner = self.inner.lock().unwrap();

        let (
            source_warehouse_id,
            target_warehouse_id,
            product_id,
            quantity,
            transfer_status,
            shipped_at,
        ) = {
            let transfer = inner
                .transfers
                .get(transfer_id)
                .ok_or_else(|| WarehouseError::TransferNotFound(transfer_id.to_string()))?;
            (
                transfer.source_warehouse_id.clone(),
                transfer.target_warehouse_id.clone(),
                transfer.product_id.clone(),
                transfer.quantity,
                transfer.status,
                transfer.shipped_at,
            )
        };

        if transfer_status != TransferStatus::Created {
            return Err(WarehouseError::InvalidStatus {
                current: transfer_status.to_string(),
            });
        }

        let target_warehouse = inner
            .warehouses
            .get(&target_warehouse_id)
            .ok_or_else(|| WarehouseError::WarehouseNotFound(target_warehouse_id.clone()))?;
        let source_warehouse = inner
            .warehouses
            .get(&source_warehouse_id)
            .ok_or_else(|| WarehouseError::WarehouseNotFound(source_warehouse_id.clone()))?;

        let shipped_over_24h = shipped_at
            .map(|t| Utc::now() - t > Duration::hours(24))
            .unwrap_or(false);
        let arrived_at_destination_city = target_warehouse.city == source_warehouse.city;

        if shipped_over_24h && arrived_at_destination_city {
            return Err(WarehouseError::CannotCancel);
        }

        let src_key = (source_warehouse_id.clone(), product_id.clone());

        let (prev_available, prev_locked, new_available, new_locked) = {
            let src_inventory = inner
                .inventory
                .get_mut(&src_key)
                .ok_or_else(|| WarehouseError::ProductNotFound {
                    product_id: product_id.clone(),
                })?;
            let prev_available = src_inventory.available_quantity;
            let prev_locked = src_inventory.locked_quantity;
            src_inventory.locked_quantity -= quantity;
            src_inventory.available_quantity += quantity;
            (
                prev_available,
                prev_locked,
                src_inventory.available_quantity,
                src_inventory.locked_quantity,
            )
        };

        let log = InventoryLog::new(
            source_warehouse_id,
            product_id,
            InventoryChangeType::Unlock,
            quantity as i32,
            prev_available,
            prev_locked,
            new_available,
            new_locked,
            Some(transfer_id.to_string()),
        );
        inner.inventory_logs.push(log);

        let transfer = inner
            .transfers
            .get_mut(transfer_id)
            .ok_or_else(|| WarehouseError::TransferNotFound(transfer_id.to_string()))?;
        transfer.status = TransferStatus::Cancelled;

        Ok(transfer.clone())
    }

    pub fn list_transfers(&self) -> Vec<TransferOrder> {
        let inner = self.inner.lock().unwrap();
        inner.transfers.values().cloned().collect()
    }

    pub fn get_transfer(&self, id: &str) -> Option<TransferOrder> {
        let inner = self.inner.lock().unwrap();
        inner.transfers.get(id).cloned()
    }

    pub fn list_inventory_logs(&self) -> Vec<InventoryLog> {
        let inner = self.inner.lock().unwrap();
        inner.inventory_logs.clone()
    }

    pub fn list_inventory_logs_by_transfer(&self, transfer_id: &str) -> Vec<InventoryLog> {
        let inner = self.inner.lock().unwrap();
        inner
            .inventory_logs
            .iter()
            .filter(|log| log.reference_transfer_id.as_deref() == Some(transfer_id))
            .cloned()
            .collect()
    }
}

impl Default for WarehouseService {
    fn default() -> Self {
        Self::new()
    }
}
