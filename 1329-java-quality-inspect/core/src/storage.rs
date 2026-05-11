use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use uuid::Uuid;

use crate::models::{DefectRecord, InspectionOrder};

#[derive(Clone, Default)]
pub struct InMemoryStorage {
    orders: Arc<Mutex<HashMap<Uuid, InspectionOrder>>>,
    defects: Arc<Mutex<HashMap<Uuid, DefectRecord>>>,
}

impl InMemoryStorage {
    pub fn new() -> Self {
        InMemoryStorage::default()
    }

    pub fn save_order(&self, order: InspectionOrder) {
        let mut orders = self.orders.lock().unwrap();
        orders.insert(order.id, order);
    }

    pub fn get_order(&self, id: Uuid) -> Option<InspectionOrder> {
        let orders = self.orders.lock().unwrap();
        orders.get(&id).cloned()
    }

    pub fn get_all_orders(&self) -> Vec<InspectionOrder> {
        let orders = self.orders.lock().unwrap();
        orders.values().cloned().collect()
    }

    pub fn save_defect(&self, defect: DefectRecord) {
        let mut defects = self.defects.lock().unwrap();
        defects.insert(defect.id, defect);
    }

    pub fn get_defect(&self, id: Uuid) -> Option<DefectRecord> {
        let defects = self.defects.lock().unwrap();
        defects.get(&id).cloned()
    }

    pub fn get_defects_for_order(&self, order_id: Uuid) -> Vec<DefectRecord> {
        let defects = self.defects.lock().unwrap();
        defects
            .values()
            .filter(|d| d.order_id == order_id)
            .cloned()
            .collect()
    }
}
