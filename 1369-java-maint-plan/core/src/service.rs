use chrono::{DateTime, Duration, Utc};
use uuid::Uuid;

use crate::models::*;
use crate::storage::InMemoryStorage;

#[derive(Clone)]
pub struct MaintenanceService {
    storage: InMemoryStorage,
}

impl MaintenanceService {
    pub fn new(storage: InMemoryStorage) -> Self {
        MaintenanceService { storage }
    }

    pub fn storage(&self) -> &InMemoryStorage {
        &self.storage
    }

    pub fn create_device(&self, name: String, code: String, description: String) -> Device {
        let device = Device::new(name, code, description);
        self.storage.add_device(device.clone());
        device
    }

    pub fn update_device_status(&self, device_id: Uuid, status: DeviceStatus) -> Option<Device> {
        let mut device = self.storage.get_device(device_id)?;
        device.status = status;
        self.storage.update_device(device.clone());
        Some(device)
    }

    pub fn update_device_running_hours(&self, device_id: Uuid, hours: u64) -> Option<Device> {
        let mut device = self.storage.get_device(device_id)?;
        device.running_hours = hours;
        self.storage.update_device(device.clone());
        Some(device)
    }

    pub fn create_maintenance_plan(
        &self,
        device_id: Uuid,
        name: String,
        cycle_type: MaintenanceCycleType,
        cycle_value: u64,
    ) -> Option<MaintenancePlan> {
        if self.storage.get_device(device_id).is_none() {
            return None;
        }
        let plan = MaintenancePlan::new(device_id, name, cycle_type, cycle_value);
        self.storage.add_maintenance_plan(plan.clone());
        Some(plan)
    }

    pub fn create_repair_order(
        &self,
        device_id: Uuid,
        title: String,
        description: String,
    ) -> Option<RepairOrder> {
        if self.storage.get_device(device_id).is_none() {
            return None;
        }
        let order = RepairOrder::new(device_id, title, description);
        self.storage.add_repair_order(order.clone());
        
        let device = self.storage.get_device(device_id).unwrap();
        self.storage.add_notification(Notification {
            id: Uuid::new_v4(),
            notification_type: NotificationType::FaultRepair,
            title: format!("新故障报修: {}", order.title),
            message: format!(
                "设备 {} ({}) 报告故障: {}",
                device.name, device.code, order.description
            ),
            device_id: Some(device_id),
            spare_part_id: None,
            is_read: false,
            created_at: Utc::now(),
        });
        
        Some(order)
    }

    pub fn complete_repair_order(
        &self,
        order_id: Uuid,
        fault_category: String,
        spare_parts: Vec<SparePartUsage>,
        repair_notes: String,
    ) -> Result<RepairOrder, String> {
        let mut order = self
            .storage
            .get_repair_order(order_id)
            .ok_or_else(|| "维修工单不存在".to_string())?;

        if order.status == RepairOrderStatus::Completed {
            return Err("维修工单已完成".to_string());
        }

        for usage in &spare_parts {
            let mut part = self
                .storage
                .get_spare_part(usage.spare_part_id)
                .ok_or_else(|| format!("备件 {:?} 不存在", usage.spare_part_id))?;
            
            if part.quantity < usage.quantity {
                return Err(format!("备件 {} 库存不足，当前: {}, 需要: {}", 
                    part.name, part.quantity, usage.quantity));
            }
        }

        for usage in &spare_parts {
            let mut part = self.storage.get_spare_part(usage.spare_part_id).unwrap();
            part.quantity -= usage.quantity;
            self.storage.update_spare_part(part.clone());
            
            if part.is_below_safety_stock() {
                self.storage.add_notification(Notification {
                    id: Uuid::new_v4(),
                    notification_type: NotificationType::LowStock,
                    title: format!("备件库存不足: {}", part.name),
                    message: format!(
                        "备件 {} ({}) 库存低于安全线，当前: {}, 安全库存: {}",
                        part.name, part.code, part.quantity, part.safety_stock
                    ),
                    device_id: None,
                    spare_part_id: Some(part.id),
                    is_read: false,
                    created_at: Utc::now(),
                });
            }
        }

        order.fault_category = Some(fault_category.clone());
        order.spare_parts_used = spare_parts;
        order.status = RepairOrderStatus::Completed;
        order.completed_at = Some(Utc::now());
        order.repair_notes = repair_notes;
        self.storage.update_repair_order(order.clone());

        self.check_high_frequency_fault(&fault_category);

        if let Some(mut device) = self.storage.get_device(order.device_id) {
            if device.status == DeviceStatus::Fault {
                device.status = DeviceStatus::Stopped;
                self.storage.update_device(device);
            }
        }

        Ok(order)
    }

    pub fn check_high_frequency_fault(&self, category: &str) {
        let thirty_days_ago = Utc::now() - Duration::days(30);
        let orders = self.storage.get_repair_orders_since(thirty_days_ago);
        
        let count = orders
            .iter()
            .filter(|o| o.fault_category.as_ref().map(|c| c == category).unwrap_or(false))
            .count();

        if count >= 3 {
            self.storage.add_notification(Notification {
                id: Uuid::new_v4(),
                notification_type: NotificationType::HighFrequencyFault,
                title: format!("高频故障预警: {}", category),
                message: format!(
                    "过去30天内，故障类型 \"{}\" 已出现 {} 次，请重点关注",
                    category, count
                ),
                device_id: None,
                spare_part_id: None,
                is_read: false,
                created_at: Utc::now(),
            });
        }
    }

    pub fn create_spare_part(
        &self,
        name: String,
        code: String,
        quantity: u32,
        safety_stock: u32,
        unit: String,
    ) -> SparePart {
        let part = SparePart::new(name, code, quantity, safety_stock, unit);
        self.storage.add_spare_part(part.clone());
        part
    }

    pub fn process_daily_tasks(&self) {
        self.generate_maintenance_orders();
        self.check_delayed_orders();
    }

    pub fn generate_maintenance_orders(&self) {
        let plans = self.storage.get_all_maintenance_plans();
        
        for plan in plans {
            let device = match self.storage.get_device(plan.device_id) {
                Some(d) => d,
                None => continue,
            };

            let orders = self.storage.get_maintenance_orders_for_device(device.id);
            let has_pending = orders.iter().any(|o| {
                o.plan_id == plan.id && 
                o.status != MaintenanceOrderStatus::Completed
            });

            if has_pending {
                continue;
            }

            let next_date = plan.next_maintenance_date(&device);
            let reminder_date = next_date - Duration::days(7);
            let now = Utc::now();

            if now >= reminder_date {
                let cycle_days = match plan.cycle_type {
                    MaintenanceCycleType::CalendarTime => plan.cycle_value,
                    MaintenanceCycleType::RunningHours => 30,
                };

                let order = MaintenanceOrder::new(
                    device.id,
                    plan.id,
                    next_date,
                    cycle_days,
                );

                self.storage.add_maintenance_order(order.clone());

                self.storage.add_notification(Notification {
                    id: Uuid::new_v4(),
                    notification_type: NotificationType::MaintenanceReminder,
                    title: format!("保养提醒: {}", plan.name),
                    message: format!(
                        "设备 {} ({}) 计划于 {} 进行保养，请提前准备",
                        device.name, device.code, 
                        next_date.format("%Y-%m-%d %H:%M")
                    ),
                    device_id: Some(device.id),
                    spare_part_id: None,
                    is_read: false,
                    created_at: Utc::now(),
                });
            }
        }
    }

    pub fn check_delayed_orders(&self) {
        let orders = self.storage.get_all_maintenance_orders();
        
        for mut order in orders {
            if order.status == MaintenanceOrderStatus::Completed {
                continue;
            }

            let device = match self.storage.get_device(order.device_id) {
                Some(d) => d,
                None => continue,
            };

            let now = Utc::now();
            let is_overdue = now > order.scheduled_date;

            if !is_overdue {
                continue;
            }

            if device.status == DeviceStatus::Running {
                if order.status != MaintenanceOrderStatus::Delayed && 
                   order.status != MaintenanceOrderStatus::Urgent {
                    order.status = MaintenanceOrderStatus::Delayed;
                    self.storage.add_notification(Notification {
                        id: Uuid::new_v4(),
                        notification_type: NotificationType::MaintenanceDelayed,
                        title: format!("保养工单延期: {}", order.id),
                        message: format!(
                            "设备 {} ({}) 正在运行，保养任务无法执行，已自动延期",
                            device.name, device.code
                        ),
                        device_id: Some(device.id),
                        spare_part_id: None,
                        is_read: false,
                        created_at: Utc::now(),
                    });
                }
                
                order.delay_days += 1;

                if order.delay_days >= order.max_delay_days && 
                   order.status != MaintenanceOrderStatus::Urgent {
                    order.status = MaintenanceOrderStatus::Urgent;
                    self.storage.add_notification(Notification {
                        id: Uuid::new_v4(),
                        notification_type: NotificationType::MaintenanceUrgent,
                        title: format!("紧急通知: 保养严重延期"),
                        message: format!(
                            "设备 {} ({}) 的保养已延期 {} 天，超过保养周期的50%，请立即安排处理",
                            device.name, device.code, order.delay_days
                        ),
                        device_id: Some(device.id),
                        spare_part_id: None,
                        is_read: false,
                        created_at: Utc::now(),
                    });
                }

                self.storage.update_maintenance_order(order);
            }
        }
    }

    pub fn complete_maintenance_order(
        &self,
        order_id: Uuid,
        notes: String,
    ) -> Result<MaintenanceOrder, String> {
        let mut order = self
            .storage
            .get_maintenance_order(order_id)
            .ok_or_else(|| "保养工单不存在".to_string())?;

        if order.status == MaintenanceOrderStatus::Completed {
            return Err("保养工单已完成".to_string());
        }

        let mut plan = self
            .storage
            .get_maintenance_plan(order.plan_id)
            .ok_or_else(|| "保养计划不存在".to_string())?;

        let device = self
            .storage
            .get_device(order.device_id)
            .ok_or_else(|| "设备不存在".to_string())?;

        plan.last_maintenance_at = Some(Utc::now());
        plan.last_running_hours = match plan.cycle_type {
            MaintenanceCycleType::RunningHours => Some(device.running_hours),
            MaintenanceCycleType::CalendarTime => plan.last_running_hours,
        };
        self.storage.update_maintenance_plan(plan);

        order.status = MaintenanceOrderStatus::Completed;
        order.completed_at = Some(Utc::now());
        order.notes = notes;
        self.storage.update_maintenance_order(order.clone());

        Ok(order)
    }

    pub fn get_device_history(&self, device_id: Uuid) -> Vec<MaintenanceHistoryItem> {
        let mut items = Vec::new();

        let maintenance_orders = self.storage.get_maintenance_orders_for_device(device_id);
        for order in maintenance_orders {
            items.push(MaintenanceHistoryItem {
                id: order.id,
                device_id,
                item_type: "maintenance".to_string(),
                title: format!("保养 - {:?}", order.status),
                description: order.notes.clone(),
                occurred_at: order.completed_at.unwrap_or(order.created_at),
            });
        }

        let repair_orders = self.storage.get_repair_orders_for_device(device_id);
        for order in repair_orders {
            items.push(MaintenanceHistoryItem {
                id: order.id,
                device_id,
                item_type: "repair".to_string(),
                title: format!(
                    "维修 - {}{}",
                    order.title,
                    order.fault_category
                        .as_ref()
                        .map(|c| format!(" ({})", c))
                        .unwrap_or_default()
                ),
                description: order.description.clone(),
                occurred_at: order.created_at,
            });
        }

        items.sort_by(|a, b| b.occurred_at.cmp(&a.occurred_at));
        items
    }

    pub fn mark_notification_read(&self, notification_id: Uuid) -> Option<Notification> {
        let mut notification = self.storage.get_notification(notification_id)?;
        notification.is_read = true;
        self.storage.update_notification(notification.clone());
        Some(notification)
    }
}
