use chrono::NaiveDate;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

use crate::error::{LeaveError, Result};

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum LeaveType {
    Annual,
    Personal,
    Sick,
    Compensatory,
}

impl std::fmt::Display for LeaveType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            LeaveType::Annual => write!(f, "年假"),
            LeaveType::Personal => write!(f, "事假"),
            LeaveType::Sick => write!(f, "病假"),
            LeaveType::Compensatory => write!(f, "调休"),
        }
    }
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum LeaveStatus {
    Approved,
    Cancelled,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Employee {
    pub id: String,
    pub name: String,
    pub hire_date: NaiveDate,
}

impl Employee {
    pub fn new(name: &str, hire_date: NaiveDate) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            name: name.to_string(),
            hire_date,
        }
    }

    pub fn with_id(id: &str, name: &str, hire_date: NaiveDate) -> Self {
        Self {
            id: id.to_string(),
            name: name.to_string(),
            hire_date,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AnnualLeavePool {
    pub year: i32,
    pub total_days: u32,
    pub used_days: u32,
    pub expires_on: Option<NaiveDate>,
}

impl AnnualLeavePool {
    pub fn new(year: i32, total_days: u32, expires_on: Option<NaiveDate>) -> Self {
        Self {
            year,
            total_days,
            used_days: 0,
            expires_on,
        }
    }

    pub fn remaining_days(&self) -> u32 {
        self.total_days.saturating_sub(self.used_days)
    }

    pub fn is_expired(&self, today: NaiveDate) -> bool {
        match self.expires_on {
            Some(expiry) => today > expiry,
            None => false,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LeaveBalance {
    pub annual_leave: Vec<AnnualLeavePool>,
    pub personal_leave: u32,
    pub sick_leave: u32,
    pub compensatory_leave: u32,
}

impl LeaveBalance {
    pub fn new() -> Self {
        Self {
            annual_leave: Vec::new(),
            personal_leave: 0,
            sick_leave: 0,
            compensatory_leave: 0,
        }
    }
}

impl Default for LeaveBalance {
    fn default() -> Self {
        Self::new()
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LeaveRequest {
    pub id: String,
    pub employee_id: String,
    pub leave_type: LeaveType,
    pub start_date: NaiveDate,
    pub end_date: NaiveDate,
    pub total_days: u32,
    pub status: LeaveStatus,
    pub sick_leave_proof: Option<String>,
    pub annual_leave_usage: Vec<AnnualLeaveUsage>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AnnualLeaveUsage {
    pub year: i32,
    pub is_carryover: bool,
    pub days_used: u32,
}

impl LeaveRequest {
    pub fn new(
        employee_id: &str,
        leave_type: LeaveType,
        start_date: NaiveDate,
        end_date: NaiveDate,
        sick_leave_proof: Option<String>,
    ) -> Result<Self> {
        if end_date < start_date {
            return Err(LeaveError::InvalidDate("结束日期不能早于开始日期".to_string()));
        }

        let total_days = calculate_days(start_date, end_date);

        if leave_type == LeaveType::Sick && total_days > 3 && sick_leave_proof.is_none() {
            return Err(LeaveError::SickLeaveNeedsProof);
        }

        Ok(Self {
            id: Uuid::new_v4().to_string(),
            employee_id: employee_id.to_string(),
            leave_type,
            start_date,
            end_date,
            total_days,
            status: LeaveStatus::Approved,
            sick_leave_proof,
            annual_leave_usage: Vec::new(),
        })
    }

    pub fn has_started(&self, today: NaiveDate) -> bool {
        today >= self.start_date
    }
}

pub fn calculate_days(start: NaiveDate, end: NaiveDate) -> u32 {
    if end < start {
        0
    } else {
        (end - start).num_days() as u32 + 1
    }
}

pub fn calculate_service_years(hire_date: NaiveDate, as_of: NaiveDate) -> f64 {
    let days = (as_of - hire_date).num_days();
    days as f64 / 365.0
}

pub fn calculate_annual_leave_entitlement(service_years: f64) -> u32 {
    if service_years >= 20.0 {
        15
    } else if service_years >= 10.0 {
        10
    } else if service_years >= 1.0 {
        5
    } else {
        0
    }
}

pub fn calculate_prorated_annual_leave(hire_date: NaiveDate, year: i32) -> u32 {
    let year_start = NaiveDate::from_ymd_opt(year, 1, 1).unwrap();
    let year_end = NaiveDate::from_ymd_opt(year, 12, 31).unwrap();

    let effective_start = std::cmp::max(hire_date, year_start);
    let days_in_service = (year_end - effective_start).num_days() + 1;

    let base_entitlement = calculate_annual_leave_entitlement(1.0);

    ((days_in_service as f64 / 365.0) * base_entitlement as f64).floor() as u32
}

pub fn max_carryover_days() -> u32 {
    3
}
