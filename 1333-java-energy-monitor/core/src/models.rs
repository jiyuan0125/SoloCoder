use chrono::NaiveDate;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq, Hash)]
pub struct PointId(pub String);

impl PointId {
    pub fn new(id: impl Into<String>) -> Self {
        PointId(id.into())
    }
}

impl std::fmt::Display for PointId {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.0)
    }
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct ElectricityUsage {
    pub peak: f64,
    pub flat: f64,
    pub valley: f64,
}

impl ElectricityUsage {
    pub fn new(peak: f64, flat: f64, valley: f64) -> Self {
        Self { peak, flat, valley }
    }

    pub fn total(&self) -> f64 {
        self.peak + self.flat + self.valley
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UsageRecord {
    pub point_id: PointId,
    pub date: NaiveDate,
    pub usage: ElectricityUsage,
}

impl UsageRecord {
    pub fn new(point_id: PointId, date: NaiveDate, usage: ElectricityUsage) -> Self {
        Self {
            point_id,
            date,
            usage,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DailyCostBreakdown {
    pub peak_cost: f64,
    pub flat_cost: f64,
    pub valley_cost: f64,
    pub total_cost: f64,
}

impl DailyCostBreakdown {
    pub fn total(&self) -> f64 {
        self.total_cost
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Alert {
    pub point_id: PointId,
    pub date: NaiveDate,
    pub current_usage: f64,
    pub previous_usage: f64,
    pub increase_percent: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WorkDayFlags {
    days: HashMap<NaiveDate, bool>,
}

impl Default for WorkDayFlags {
    fn default() -> Self {
        Self::new()
    }
}

impl WorkDayFlags {
    pub fn new() -> Self {
        Self {
            days: HashMap::new(),
        }
    }

    pub fn mark_shutdown(&mut self, date: NaiveDate) {
        self.days.insert(date, false);
    }

    pub fn unmark_shutdown(&mut self, date: NaiveDate) {
        self.days.insert(date, true);
    }

    pub fn is_shutdown_day(&self, date: &NaiveDate) -> bool {
        matches!(self.days.get(date), Some(false))
    }

    pub fn is_work_day(&self, date: &NaiveDate) -> bool {
        self.days.get(date).copied().unwrap_or(true)
    }

    pub fn should_skip_alert(&self, date: &NaiveDate) -> bool {
        let prev_date = *date - chrono::Duration::days(1);
        self.is_shutdown_day(date) || self.is_shutdown_day(&prev_date)
    }
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
pub enum TariffPeriod {
    Peak,
    Flat,
    Valley,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TimeRange {
    pub start_hour: u32,
    pub end_hour: u32,
}

impl TimeRange {
    pub fn new(start_hour: u32, end_hour: u32) -> Self {
        Self {
            start_hour,
            end_hour,
        }
    }

    pub fn contains_hour(&self, hour: u32) -> bool {
        if self.start_hour <= self.end_hour {
            hour >= self.start_hour && hour < self.end_hour
        } else {
            hour >= self.start_hour || hour < self.end_hour
        }
    }
}
