use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MaintenanceRule {
    pub hours_interval: Option<u64>,
    pub days_interval: Option<u32>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Equipment {
    pub id: String,
    pub name: String,
    pub install_date: DateTime<Utc>,
    pub total_runtime_hours: u64,
    pub maintenance_rule: MaintenanceRule,
    pub last_maintenance_date: Option<DateTime<Utc>>,
    pub last_maintenance_runtime: Option<u64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MaterialUsage {
    pub name: String,
    pub quantity: f64,
    pub unit_cost: f64,
}

impl MaterialUsage {
    pub fn total_cost(&self) -> f64 {
        self.quantity * self.unit_cost
    }
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum MaintenanceType {
    Hourly,
    Calendar,
    Both,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MaintenanceRecord {
    pub id: String,
    pub equipment_id: String,
    pub maintenance_type: MaintenanceType,
    pub date: DateTime<Utc>,
    pub labor_hours: f64,
    pub materials: Vec<MaterialUsage>,
    pub personnel: String,
    pub notes: Option<String>,
    pub runtime_at_maintenance: u64,
}

impl MaintenanceRecord {
    pub fn material_cost(&self) -> f64 {
        self.materials.iter().map(|m| m.total_cost()).sum()
    }

    pub fn total_cost(&self) -> f64 {
        self.labor_hours * 50.0 + self.material_cost()
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateEquipmentRequest {
    pub name: String,
    pub install_date: DateTime<Utc>,
    pub hours_interval: Option<u64>,
    pub days_interval: Option<u32>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UpdateRuntimeRequest {
    pub additional_hours: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PerformMaintenanceRequest {
    pub maintenance_type: MaintenanceType,
    pub date: DateTime<Utc>,
    pub labor_hours: f64,
    pub materials: Vec<MaterialUsage>,
    pub personnel: String,
    pub notes: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MaintenanceReminder {
    pub equipment_id: String,
    pub equipment_name: String,
    pub trigger_type: MaintenanceType,
    pub hours_since_last: Option<u64>,
    pub days_since_last: Option<u32>,
    pub hours_remaining: Option<i64>,
    pub days_remaining: Option<i32>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StatisticsResult {
    pub total_labor_hours: f64,
    pub total_material_cost: f64,
    pub total_cost: f64,
    pub record_count: usize,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TimeRange {
    pub start: DateTime<Utc>,
    pub end: DateTime<Utc>,
}
