use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use std::cmp::Ordering;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum OrderPriority {
    Normal,
    Urgent,
}

impl Ord for OrderPriority {
    fn cmp(&self, other: &Self) -> Ordering {
        match (self, other) {
            (OrderPriority::Urgent, OrderPriority::Normal) => Ordering::Greater,
            (OrderPriority::Normal, OrderPriority::Urgent) => Ordering::Less,
            _ => Ordering::Equal,
        }
    }
}

impl PartialOrd for OrderPriority {
    fn partial_cmp(&self, other: &Self) -> Option<Ordering> {
        Some(self.cmp(other))
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Order {
    pub id: String,
    pub name: String,
    pub priority: OrderPriority,
    pub duration_minutes: u32,
    pub due_date: DateTime<Utc>,
    pub assigned_line_id: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Device {
    pub id: String,
    pub name: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MaintenanceWindow {
    pub id: String,
    pub device_id: String,
    pub start_time: DateTime<Utc>,
    pub end_time: DateTime<Utc>,
    pub description: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProductionLine {
    pub id: String,
    pub name: String,
    pub devices: Vec<Device>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ScheduledTask {
    pub order_id: String,
    pub line_id: String,
    pub start_time: DateTime<Utc>,
    pub end_time: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Warning {
    pub order_id: String,
    pub message: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ScheduleResult {
    pub tasks: Vec<ScheduledTask>,
    pub warnings: Vec<Warning>,
}
