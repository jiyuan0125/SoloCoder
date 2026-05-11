use std::collections::HashMap;
use std::sync::{Arc, Mutex};

use crate::models::{Device, MaintenanceWindow, Order, ProductionLine, ScheduledTask};
use crate::errors::SchedulerError;

#[derive(Debug, Clone, Default)]
pub struct InMemoryStore {
    inner: Arc<Mutex<StoreInner>>,
}

#[derive(Debug, Default)]
struct StoreInner {
    lines: HashMap<String, ProductionLine>,
    devices: HashMap<String, Device>,
    orders: HashMap<String, Order>,
    maintenance_windows: HashMap<String, MaintenanceWindow>,
    scheduled_tasks: HashMap<String, Vec<ScheduledTask>>,
}

impl InMemoryStore {
    pub fn new() -> Self {
        Self {
            inner: Arc::new(Mutex::new(StoreInner::default())),
        }
    }

    pub fn add_line(&self, line: ProductionLine) {
        let mut inner = self.inner.lock().unwrap();
        for device in &line.devices {
            inner.devices.insert(device.id.clone(), device.clone());
        }
        inner.lines.insert(line.id.clone(), line);
    }

    pub fn get_line(&self, line_id: &str) -> Option<ProductionLine> {
        let inner = self.inner.lock().unwrap();
        inner.lines.get(line_id).cloned()
    }

    pub fn get_all_lines(&self) -> Vec<ProductionLine> {
        let inner = self.inner.lock().unwrap();
        inner.lines.values().cloned().collect()
    }

    pub fn get_devices_for_line(&self, line_id: &str) -> Vec<Device> {
        let inner = self.inner.lock().unwrap();
        inner.lines.get(line_id)
            .map(|line| line.devices.clone())
            .unwrap_or_default()
    }

    pub fn add_order(&self, order: Order) {
        let mut inner = self.inner.lock().unwrap();
        inner.orders.insert(order.id.clone(), order);
    }

    pub fn get_order(&self, order_id: &str) -> Option<Order> {
        let inner = self.inner.lock().unwrap();
        inner.orders.get(order_id).cloned()
    }

    pub fn get_all_orders(&self) -> Vec<Order> {
        let inner = self.inner.lock().unwrap();
        inner.orders.values().cloned().collect()
    }

    pub fn update_order(&self, order: Order) -> Result<(), SchedulerError> {
        let mut inner = self.inner.lock().unwrap();
        if !inner.orders.contains_key(&order.id) {
            return Err(SchedulerError::OrderNotFound(order.id.clone()));
        }
        inner.orders.insert(order.id.clone(), order);
        Ok(())
    }

    pub fn add_maintenance_window(&self, window: MaintenanceWindow) -> Result<(), SchedulerError> {
        if window.end_time <= window.start_time {
            return Err(SchedulerError::InvalidMaintenanceWindow);
        }
        let mut inner = self.inner.lock().unwrap();
        inner.maintenance_windows.insert(window.id.clone(), window);
        Ok(())
    }

    pub fn get_maintenance_window(&self, window_id: &str) -> Option<MaintenanceWindow> {
        let inner = self.inner.lock().unwrap();
        inner.maintenance_windows.get(window_id).cloned()
    }

    pub fn get_all_maintenance_windows(&self) -> Vec<MaintenanceWindow> {
        let inner = self.inner.lock().unwrap();
        inner.maintenance_windows.values().cloned().collect()
    }

    pub fn update_maintenance_window(&self, window: MaintenanceWindow) -> Result<(), SchedulerError> {
        if window.end_time <= window.start_time {
            return Err(SchedulerError::InvalidMaintenanceWindow);
        }
        let mut inner = self.inner.lock().unwrap();
        if !inner.maintenance_windows.contains_key(&window.id) {
            return Err(SchedulerError::MaintenanceWindowNotFound(window.id.clone()));
        }
        inner.maintenance_windows.insert(window.id.clone(), window);
        Ok(())
    }

    pub fn delete_maintenance_window(&self, window_id: &str) -> Result<(), SchedulerError> {
        let mut inner = self.inner.lock().unwrap();
        if !inner.maintenance_windows.remove(window_id).is_some() {
            return Err(SchedulerError::MaintenanceWindowNotFound(window_id.to_string()));
        }
        Ok(())
    }

    pub fn get_maintenance_windows_for_line(&self, line_id: &str) -> Vec<MaintenanceWindow> {
        let inner = self.inner.lock().unwrap();
        let device_ids: Vec<String> = inner.lines.get(line_id)
            .map(|line| line.devices.iter().map(|d| d.id.clone()).collect())
            .unwrap_or_default();
        
        inner.maintenance_windows.values()
            .filter(|w| device_ids.contains(&w.device_id))
            .cloned()
            .collect()
    }

    pub fn get_scheduled_tasks(&self, line_id: &str) -> Vec<ScheduledTask> {
        let inner = self.inner.lock().unwrap();
        inner.scheduled_tasks.get(line_id).cloned().unwrap_or_default()
    }

    pub fn get_all_scheduled_tasks(&self) -> Vec<ScheduledTask> {
        let inner = self.inner.lock().unwrap();
        inner.scheduled_tasks.values().flatten().cloned().collect()
    }

    pub fn set_scheduled_tasks(&self, line_id: String, tasks: Vec<ScheduledTask>) {
        let mut inner = self.inner.lock().unwrap();
        inner.scheduled_tasks.insert(line_id, tasks);
    }

    pub fn clear_all_scheduled_tasks(&self) {
        let mut inner = self.inner.lock().unwrap();
        inner.scheduled_tasks.clear();
    }
}
