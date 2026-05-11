use std::collections::HashMap;
use std::sync::{Arc, Mutex};

use chrono::{DateTime, Utc};
use uuid::Uuid;

use crate::models::*;

#[derive(Clone, Default)]
pub struct InMemoryStorage {
    devices: Arc<Mutex<HashMap<Uuid, Device>>>,
    maintenance_plans: Arc<Mutex<HashMap<Uuid, MaintenancePlan>>>,
    maintenance_orders: Arc<Mutex<HashMap<Uuid, MaintenanceOrder>>>,
    repair_orders: Arc<Mutex<HashMap<Uuid, RepairOrder>>>,
    spare_parts: Arc<Mutex<HashMap<Uuid, SparePart>>>,
    fault_categories: Arc<Mutex<HashMap<Uuid, FaultCategory>>>,
    notifications: Arc<Mutex<HashMap<Uuid, Notification>>>,
}

impl InMemoryStorage {
    pub fn new() -> Self {
        InMemoryStorage {
            devices: Arc::new(Mutex::new(HashMap::new())),
            maintenance_plans: Arc::new(Mutex::new(HashMap::new())),
            maintenance_orders: Arc::new(Mutex::new(HashMap::new())),
            repair_orders: Arc::new(Mutex::new(HashMap::new())),
            spare_parts: Arc::new(Mutex::new(HashMap::new())),
            fault_categories: Arc::new(Mutex::new(HashMap::new())),
            notifications: Arc::new(Mutex::new(HashMap::new())),
        }
    }

    pub fn add_device(&self, device: Device) {
        self.devices.lock().unwrap().insert(device.id, device);
    }

    pub fn get_device(&self, id: Uuid) -> Option<Device> {
        self.devices.lock().unwrap().get(&id).cloned()
    }

    pub fn get_all_devices(&self) -> Vec<Device> {
        self.devices.lock().unwrap().values().cloned().collect()
    }

    pub fn update_device(&self, device: Device) {
        self.devices.lock().unwrap().insert(device.id, device);
    }

    pub fn add_maintenance_plan(&self, plan: MaintenancePlan) {
        self.maintenance_plans.lock().unwrap().insert(plan.id, plan);
    }

    pub fn get_maintenance_plan(&self, id: Uuid) -> Option<MaintenancePlan> {
        self.maintenance_plans.lock().unwrap().get(&id).cloned()
    }

    pub fn get_maintenance_plans_for_device(&self, device_id: Uuid) -> Vec<MaintenancePlan> {
        self.maintenance_plans
            .lock()
            .unwrap()
            .values()
            .filter(|p| p.device_id == device_id)
            .cloned()
            .collect()
    }

    pub fn get_all_maintenance_plans(&self) -> Vec<MaintenancePlan> {
        self.maintenance_plans
            .lock()
            .unwrap()
            .values()
            .cloned()
            .collect()
    }

    pub fn update_maintenance_plan(&self, plan: MaintenancePlan) {
        self.maintenance_plans.lock().unwrap().insert(plan.id, plan);
    }

    pub fn add_maintenance_order(&self, order: MaintenanceOrder) {
        self.maintenance_orders.lock().unwrap().insert(order.id, order);
    }

    pub fn get_maintenance_order(&self, id: Uuid) -> Option<MaintenanceOrder> {
        self.maintenance_orders.lock().unwrap().get(&id).cloned()
    }

    pub fn get_maintenance_orders_for_device(&self, device_id: Uuid) -> Vec<MaintenanceOrder> {
        self.maintenance_orders
            .lock()
            .unwrap()
            .values()
            .filter(|o| o.device_id == device_id)
            .cloned()
            .collect()
    }

    pub fn get_all_maintenance_orders(&self) -> Vec<MaintenanceOrder> {
        self.maintenance_orders
            .lock()
            .unwrap()
            .values()
            .cloned()
            .collect()
    }

    pub fn update_maintenance_order(&self, order: MaintenanceOrder) {
        self.maintenance_orders.lock().unwrap().insert(order.id, order);
    }

    pub fn add_repair_order(&self, order: RepairOrder) {
        self.repair_orders.lock().unwrap().insert(order.id, order);
    }

    pub fn get_repair_order(&self, id: Uuid) -> Option<RepairOrder> {
        self.repair_orders.lock().unwrap().get(&id).cloned()
    }

    pub fn get_repair_orders_for_device(&self, device_id: Uuid) -> Vec<RepairOrder> {
        self.repair_orders
            .lock()
            .unwrap()
            .values()
            .filter(|o| o.device_id == device_id)
            .cloned()
            .collect()
    }

    pub fn get_all_repair_orders(&self) -> Vec<RepairOrder> {
        self.repair_orders
            .lock()
            .unwrap()
            .values()
            .cloned()
            .collect()
    }

    pub fn update_repair_order(&self, order: RepairOrder) {
        self.repair_orders.lock().unwrap().insert(order.id, order);
    }

    pub fn add_spare_part(&self, part: SparePart) {
        self.spare_parts.lock().unwrap().insert(part.id, part);
    }

    pub fn get_spare_part(&self, id: Uuid) -> Option<SparePart> {
        self.spare_parts.lock().unwrap().get(&id).cloned()
    }

    pub fn get_all_spare_parts(&self) -> Vec<SparePart> {
        self.spare_parts.lock().unwrap().values().cloned().collect()
    }

    pub fn update_spare_part(&self, part: SparePart) {
        self.spare_parts.lock().unwrap().insert(part.id, part);
    }

    pub fn add_fault_category(&self, category: FaultCategory) {
        self.fault_categories
            .lock()
            .unwrap()
            .insert(category.id, category);
    }

    pub fn get_fault_category(&self, id: Uuid) -> Option<FaultCategory> {
        self.fault_categories.lock().unwrap().get(&id).cloned()
    }

    pub fn get_all_fault_categories(&self) -> Vec<FaultCategory> {
        self.fault_categories
            .lock()
            .unwrap()
            .values()
            .cloned()
            .collect()
    }

    pub fn add_notification(&self, notification: Notification) {
        self.notifications
            .lock()
            .unwrap()
            .insert(notification.id, notification);
    }

    pub fn get_notification(&self, id: Uuid) -> Option<Notification> {
        self.notifications.lock().unwrap().get(&id).cloned()
    }

    pub fn get_all_notifications(&self) -> Vec<Notification> {
        self.notifications
            .lock()
            .unwrap()
            .values()
            .cloned()
            .collect()
    }

    pub fn update_notification(&self, notification: Notification) {
        self.notifications
            .lock()
            .unwrap()
            .insert(notification.id, notification);
    }

    pub fn get_repair_orders_by_category(&self, category: &str) -> Vec<RepairOrder> {
        self.repair_orders
            .lock()
            .unwrap()
            .values()
            .filter(|o| o.fault_category.as_ref().map(|c| c == category).unwrap_or(false))
            .cloned()
            .collect()
    }

    pub fn get_repair_orders_since(&self, since: DateTime<Utc>) -> Vec<RepairOrder> {
        self.repair_orders
            .lock()
            .unwrap()
            .values()
            .filter(|o| o.created_at >= since)
            .cloned()
            .collect()
    }
}
