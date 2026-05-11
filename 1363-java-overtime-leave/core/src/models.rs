use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum OvertimeType {
    Workday,
    Weekend,
    Holiday,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum WeekendCompensation {
    Leave,
    OvertimePay,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum OvertimeStatus {
    Pending,
    Approved,
    Rejected,
    Expired,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum LeaveStatus {
    Pending,
    Approved,
    Rejected,
    Cancelled,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Employee {
    pub id: Uuid,
    pub name: String,
    pub monthly_salary: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Overtime {
    pub id: Uuid,
    pub employee_id: Uuid,
    pub start_time: DateTime<Utc>,
    pub end_time: DateTime<Utc>,
    pub overtime_type: OvertimeType,
    pub weekend_compensation: Option<WeekendCompensation>,
    pub status: OvertimeStatus,
    pub submitted_at: DateTime<Utc>,
    pub duration_hours: f64,
    pub overtime_pay: Option<f64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Leave {
    pub id: Uuid,
    pub employee_id: Uuid,
    pub start_time: DateTime<Utc>,
    pub end_time: DateTime<Utc>,
    pub duration_hours: f64,
    pub status: LeaveStatus,
    pub submitted_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Holiday {
    pub id: Uuid,
    pub date: DateTime<Utc>,
    pub name: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LeaveBalance {
    pub employee_id: Uuid,
    pub balance_hours: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateOvertimeRequest {
    pub employee_id: Uuid,
    pub start_time: DateTime<Utc>,
    pub end_time: DateTime<Utc>,
    pub weekend_compensation: Option<WeekendCompensation>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateLeaveRequest {
    pub employee_id: Uuid,
    pub start_time: DateTime<Utc>,
    pub end_time: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApproveOvertimeRequest {
    pub overtime_id: Uuid,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RejectOvertimeRequest {
    pub overtime_id: Uuid,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApproveLeaveRequest {
    pub leave_id: Uuid,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RejectLeaveRequest {
    pub leave_id: Uuid,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CancelLeaveRequest {
    pub leave_id: Uuid,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateHolidayRequest {
    pub date: DateTime<Utc>,
    pub name: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateEmployeeRequest {
    pub name: String,
    pub monthly_salary: f64,
}
