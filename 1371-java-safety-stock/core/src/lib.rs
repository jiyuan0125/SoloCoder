use chrono::{Datelike, Duration, NaiveDate, Utc};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use uuid::Uuid;

pub mod models {
    use super::*;

    #[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
    pub struct Material {
        pub id: Uuid,
        pub name: String,
        pub safety_stock: i32,
        pub lead_time_days: u32,
        pub unit: String,
        pub created_at: chrono::DateTime<Utc>,
    }

    #[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
    pub struct InboundRecord {
        pub id: Uuid,
        pub material_id: Uuid,
        pub quantity: i32,
        pub supplier: String,
        pub batch_number: String,
        pub created_at: chrono::DateTime<Utc>,
    }

    #[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
    pub struct OutboundRecord {
        pub id: Uuid,
        pub material_id: Uuid,
        pub quantity: i32,
        pub is_excess: bool,
        pub created_at: chrono::DateTime<Utc>,
    }

    #[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
    pub struct ReplenishmentOrder {
        pub id: Uuid,
        pub material_id: Uuid,
        pub current_stock: i32,
        pub safety_stock: i32,
        pub suggested_quantity: i32,
        pub suggested_arrival_date: NaiveDate,
        pub status: ReplenishmentStatus,
        pub created_at: chrono::DateTime<Utc>,
    }

    #[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
    pub enum ReplenishmentStatus {
        Pending,
        Confirmed,
        Cancelled,
    }

    #[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
    pub struct MaterialInventory {
        pub material: Material,
        pub current_stock: i32,
    }

    #[derive(Debug, Clone, Serialize, Deserialize)]
    pub struct BatchOutboundItem {
        pub material_id: Uuid,
        pub quantity: i32,
    }
}

pub mod utils {
    use super::*;

    pub fn is_weekend(date: NaiveDate) -> bool {
        let weekday = date.weekday();
        weekday == chrono::Weekday::Sat || weekday == chrono::Weekday::Sun
    }

    pub fn add_workdays(from: NaiveDate, days: u32) -> NaiveDate {
        let mut result = from;
        let mut remaining = days;

        while remaining > 0 {
            result = result + Duration::days(1);
            if !is_weekend(result) {
                remaining -= 1;
            }
        }

        result
    }

    pub fn calculate_replenishment_quantity(current_stock: i32, safety_stock: i32) -> i32 {
        if current_stock >= safety_stock {
            0
        } else {
            let base = safety_stock - current_stock;
            if current_stock < 0 {
                base + current_stock.abs()
            } else {
                base
            }
        }
    }
}

pub mod services {
    use super::*;
    use models::*;

    #[derive(Debug, Clone, Default)]
    pub struct WarehouseService {
        materials: HashMap<Uuid, Material>,
        inbound_records: Vec<InboundRecord>,
        outbound_records: Vec<OutboundRecord>,
        replenishment_orders: Vec<ReplenishmentOrder>,
    }

    impl WarehouseService {
        pub fn new() -> Self {
            Self::default()
        }

        pub fn add_material(
            &mut self,
            name: String,
            safety_stock: i32,
            lead_time_days: u32,
            unit: String,
        ) -> Material {
            let material = Material {
                id: Uuid::new_v4(),
                name,
                safety_stock,
                lead_time_days,
                unit,
                created_at: Utc::now(),
            };
            self.materials.insert(material.id, material.clone());
            material
        }

        pub fn get_material(&self, id: Uuid) -> Option<Material> {
            self.materials.get(&id).cloned()
        }

        pub fn list_materials(&self) -> Vec<Material> {
            self.materials.values().cloned().collect()
        }

        pub fn get_inventory(&self, material_id: Uuid) -> Option<i32> {
            if !self.materials.contains_key(&material_id) {
                return None;
            }

            let inbound: i32 = self
                .inbound_records
                .iter()
                .filter(|r| r.material_id == material_id)
                .map(|r| r.quantity)
                .sum();

            let outbound: i32 = self
                .outbound_records
                .iter()
                .filter(|r| r.material_id == material_id)
                .map(|r| r.quantity)
                .sum();

            Some(inbound - outbound)
        }

        pub fn list_inventories(&self) -> Vec<MaterialInventory> {
            self.materials
                .values()
                .map(|m| MaterialInventory {
                    material: m.clone(),
                    current_stock: self.get_inventory(m.id).unwrap_or(0),
                })
                .collect()
        }

        pub fn inbound(
            &mut self,
            material_id: Uuid,
            quantity: i32,
            supplier: String,
            batch_number: String,
        ) -> Result<InboundRecord, String> {
            if quantity <= 0 {
                return Err("入库数量必须大于0".to_string());
            }

            if !self.materials.contains_key(&material_id) {
                return Err("原料不存在".to_string());
            }

            let record = InboundRecord {
                id: Uuid::new_v4(),
                material_id,
                quantity,
                supplier,
                batch_number,
                created_at: Utc::now(),
            };

            self.inbound_records.push(record.clone());
            self.auto_cancel_pending_orders(material_id);

            Ok(record)
        }

        pub fn outbound(&mut self, material_id: Uuid, quantity: i32) -> Result<OutboundRecord, String> {
            if quantity <= 0 {
                return Err("出库数量必须大于0".to_string());
            }

            if !self.materials.contains_key(&material_id) {
                return Err("原料不存在".to_string());
            }

            let current_stock = self.get_inventory(material_id).unwrap_or(0);
            let is_excess = current_stock < quantity;

            let record = OutboundRecord {
                id: Uuid::new_v4(),
                material_id,
                quantity,
                is_excess,
                created_at: Utc::now(),
            };

            self.outbound_records.push(record.clone());
            self.auto_generate_replenishment_order(material_id);

            Ok(record)
        }

        pub fn batch_outbound(
            &mut self,
            items: Vec<BatchOutboundItem>,
        ) -> Result<Vec<OutboundRecord>, String> {
            for item in &items {
                if item.quantity <= 0 {
                    return Err("出库数量必须大于0".to_string());
                }
                if !self.materials.contains_key(&item.material_id) {
                    return Err(format!("原料不存在: {}", item.material_id));
                }
            }

            let mut records = Vec::new();
            for item in items {
                let record = self.outbound(item.material_id, item.quantity)?;
                records.push(record);
            }

            Ok(records)
        }

        pub fn list_inbound_records(&self, material_id: Option<Uuid>) -> Vec<InboundRecord> {
            let mut records = self.inbound_records.clone();
            if let Some(id) = material_id {
                records.retain(|r| r.material_id == id);
            }
            records.sort_by(|a, b| b.created_at.cmp(&a.created_at));
            records
        }

        pub fn list_outbound_records(&self, material_id: Option<Uuid>) -> Vec<OutboundRecord> {
            let mut records = self.outbound_records.clone();
            if let Some(id) = material_id {
                records.retain(|r| r.material_id == id);
            }
            records.sort_by(|a, b| b.created_at.cmp(&a.created_at));
            records
        }

        fn has_pending_order(&self, material_id: Uuid) -> bool {
            self.replenishment_orders
                .iter()
                .any(|o| o.material_id == material_id && o.status == ReplenishmentStatus::Pending)
        }

        fn auto_generate_replenishment_order(&mut self, material_id: Uuid) {
            let material = match self.materials.get(&material_id) {
                Some(m) => m,
                None => return,
            };

            let current_stock = self.get_inventory(material_id).unwrap_or(0);

            if current_stock >= material.safety_stock {
                return;
            }

            if self.has_pending_order(material_id) {
                return;
            }

            let suggested_quantity =
                utils::calculate_replenishment_quantity(current_stock, material.safety_stock);

            if suggested_quantity <= 0 {
                return;
            }

            let today = Utc::now().date_naive();
            let suggested_arrival_date = utils::add_workdays(today, material.lead_time_days);

            let order = ReplenishmentOrder {
                id: Uuid::new_v4(),
                material_id,
                current_stock,
                safety_stock: material.safety_stock,
                suggested_quantity,
                suggested_arrival_date,
                status: ReplenishmentStatus::Pending,
                created_at: Utc::now(),
            };

            self.replenishment_orders.push(order);
        }

        fn auto_cancel_pending_orders(&mut self, material_id: Uuid) {
            let material = match self.materials.get(&material_id) {
                Some(m) => m,
                None => return,
            };

            let current_stock = self.get_inventory(material_id).unwrap_or(0);

            if current_stock < material.safety_stock {
                return;
            }

            for order in self.replenishment_orders.iter_mut() {
                if order.material_id == material_id && order.status == ReplenishmentStatus::Pending {
                    order.status = ReplenishmentStatus::Cancelled;
                }
            }
        }

        pub fn list_replenishment_orders(
            &self,
            material_id: Option<Uuid>,
            status: Option<ReplenishmentStatus>,
        ) -> Vec<ReplenishmentOrder> {
            let mut orders = self.replenishment_orders.clone();

            if let Some(id) = material_id {
                orders.retain(|o| o.material_id == id);
            }

            if let Some(s) = status {
                orders.retain(|o| o.status == s);
            }

            orders.sort_by(|a, b| b.created_at.cmp(&a.created_at));
            orders
        }

        pub fn confirm_replenishment_order(&mut self, order_id: Uuid) -> Result<ReplenishmentOrder, String> {
            let order = self
                .replenishment_orders
                .iter_mut()
                .find(|o| o.id == order_id)
                .ok_or_else(|| "补货建议单不存在".to_string())?;

            if order.status != ReplenishmentStatus::Pending {
                return Err("只有待确认的补货建议单可以确认".to_string());
            }

            order.status = ReplenishmentStatus::Confirmed;
            Ok(order.clone())
        }

        pub fn cancel_replenishment_order(&mut self, order_id: Uuid) -> Result<ReplenishmentOrder, String> {
            let order = self
                .replenishment_orders
                .iter_mut()
                .find(|o| o.id == order_id)
                .ok_or_else(|| "补货建议单不存在".to_string())?;

            if order.status != ReplenishmentStatus::Pending {
                return Err("只有待确认的补货建议单可以取消".to_string());
            }

            order.status = ReplenishmentStatus::Cancelled;
            Ok(order.clone())
        }
    }
}
