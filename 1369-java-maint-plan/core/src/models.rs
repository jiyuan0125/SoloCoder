use chrono::{DateTime, Utc, Duration};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub enum MaintenanceCycleType {
    CalendarTime,
    RunningHours,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub enum DeviceStatus {
    Running,
    Stopped,
    Maintenance,
    Fault,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub enum MaintenanceOrderStatus {
    Pending,
    InProgress,
    Completed,
    Delayed,
    Urgent,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub enum RepairOrderStatus {
    Pending,
    InProgress,
    Completed,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub enum NotificationType {
    MaintenanceReminder,
    MaintenanceDelayed,
    MaintenanceUrgent,
    FaultRepair,
    HighFrequencyFault,
    LowStock,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Device {
    pub id: Uuid,
    pub name: String,
    pub code: String,
    pub description: String,
    pub status: DeviceStatus,
    pub running_hours: u64,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MaintenancePlan {
    pub id: Uuid,
    pub device_id: Uuid,
    pub name: String,
    pub cycle_type: MaintenanceCycleType,
    pub cycle_value: u64,
    pub last_maintenance_at: Option<DateTime<Utc>>,
    pub last_running_hours: Option<u64>,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MaintenanceOrder {
    pub id: Uuid,
    pub device_id: Uuid,
    pub plan_id: Uuid,
    pub scheduled_date: DateTime<Utc>,
    pub status: MaintenanceOrderStatus,
    pub delay_days: u32,
    pub max_delay_days: u32,
    pub completed_at: Option<DateTime<Utc>>,
    pub notes: String,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RepairOrder {
    pub id: Uuid,
    pub device_id: Uuid,
    pub title: String,
    pub description: String,
    pub fault_category: Option<String>,
    pub status: RepairOrderStatus,
    pub spare_parts_used: Vec<SparePartUsage>,
    pub completed_at: Option<DateTime<Utc>>,
    pub repair_notes: String,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SparePart {
    pub id: Uuid,
    pub name: String,
    pub code: String,
    pub quantity: u32,
    pub safety_stock: u32,
    pub unit: String,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SparePartUsage {
    pub spare_part_id: Uuid,
    pub quantity: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FaultCategory {
    pub id: Uuid,
    pub name: String,
    pub description: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Notification {
    pub id: Uuid,
    pub notification_type: NotificationType,
    pub title: String,
    pub message: String,
    pub device_id: Option<Uuid>,
    pub spare_part_id: Option<Uuid>,
    pub is_read: bool,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MaintenanceHistoryItem {
    pub id: Uuid,
    pub device_id: Uuid,
    pub item_type: String,
    pub title: String,
    pub description: String,
    pub occurred_at: DateTime<Utc>,
}

impl Device {
    pub fn new(name: String, code: String, description: String) -> Self {
        Device {
            id: Uuid::new_v4(),
            name,
            code,
            description,
            status: DeviceStatus::Stopped,
            running_hours: 0,
            created_at: Utc::now(),
        }
    }
}

impl MaintenancePlan {
    pub fn new(
        device_id: Uuid,
        name: String,
        cycle_type: MaintenanceCycleType,
        cycle_value: u64,
    ) -> Self {
        MaintenancePlan {
            id: Uuid::new_v4(),
            device_id,
            name,
            cycle_type,
            cycle_value,
            last_maintenance_at: None,
            last_running_hours: None,
            created_at: Utc::now(),
        }
    }

    pub fn next_maintenance_date(&self, device: &Device) -> DateTime<Utc> {
        match self.cycle_type {
            MaintenanceCycleType::CalendarTime => {
                let base = self.last_maintenance_at.unwrap_or(device.created_at);
                base + Duration::days(self.cycle_value as i64)
            }
            MaintenanceCycleType::RunningHours => {
                let current_hours = device.running_hours;
                let last_hours = self.last_running_hours.unwrap_or(0);
                let remaining_hours = if current_hours >= last_hours {
                    current_hours - last_hours
                } else {
                    0
                };
                
                if remaining_hours >= self.cycle_value {
                    Utc::now()
                } else {
                    let hours_needed = self.cycle_value - remaining_hours;
                    Utc::now() + Duration::hours(hours_needed as i64)
                }
            }
        }
    }

    pub fn needs_maintenance(&self, device: &Device) -> bool {
        match self.cycle_type {
            MaintenanceCycleType::CalendarTime => {
                let next_date = self.next_maintenance_date(device);
                Utc::now() >= next_date
            }
            MaintenanceCycleType::RunningHours => {
                let last_hours = self.last_running_hours.unwrap_or(0);
                let elapsed = device.running_hours.saturating_sub(last_hours);
                elapsed >= self.cycle_value
            }
        }
    }
}

impl MaintenanceOrder {
    pub fn new(
        device_id: Uuid,
        plan_id: Uuid,
        scheduled_date: DateTime<Utc>,
        plan_cycle_days: u64,
    ) -> Self {
        MaintenanceOrder {
            id: Uuid::new_v4(),
            device_id,
            plan_id,
            scheduled_date,
            status: MaintenanceOrderStatus::Pending,
            delay_days: 0,
            max_delay_days: (plan_cycle_days as f64 * 0.5) as u32,
            completed_at: None,
            notes: String::new(),
            created_at: Utc::now(),
        }
    }
}

impl RepairOrder {
    pub fn new(device_id: Uuid, title: String, description: String) -> Self {
        RepairOrder {
            id: Uuid::new_v4(),
            device_id,
            title,
            description,
            fault_category: None,
            status: RepairOrderStatus::Pending,
            spare_parts_used: Vec::new(),
            completed_at: None,
            repair_notes: String::new(),
            created_at: Utc::now(),
        }
    }
}

impl SparePart {
    pub fn new(name: String, code: String, quantity: u32, safety_stock: u32, unit: String) -> Self {
        SparePart {
            id: Uuid::new_v4(),
            name,
            code,
            quantity,
            safety_stock,
            unit,
            created_at: Utc::now(),
        }
    }

    pub fn is_below_safety_stock(&self) -> bool {
        self.quantity < self.safety_stock
    }
}

impl FaultCategory {
    pub fn new(name: String, description: String) -> Self {
        FaultCategory {
            id: Uuid::new_v4(),
            name,
            description,
        }
    }
}
