use chrono::{DateTime, Datelike, Local, NaiveDate};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum DeskType {
    Fixed,
    Shared,
    Visitor,
}

impl DeskType {
    pub fn as_str(&self) -> &'static str {
        match self {
            DeskType::Fixed => "fixed",
            DeskType::Shared => "shared",
            DeskType::Visitor => "visitor",
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Department {
    pub id: Uuid,
    pub name: String,
    pub floor: i32,
    pub manager_id: Option<Uuid>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Employee {
    pub id: Uuid,
    pub name: String,
    pub department_id: Uuid,
    pub is_active: bool,
    pub join_date: NaiveDate,
    pub leave_date: Option<NaiveDate>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Desk {
    pub id: Uuid,
    pub code: String,
    pub desk_type: DeskType,
    pub floor: i32,
    pub status: DeskStatus,
    pub current_employee_id: Option<Uuid>,
    pub reservation_expiry_date: Option<NaiveDate>,
    pub reserved_for_department: Option<Uuid>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum DeskStatus {
    Available,
    Occupied,
    Reserved,
    PendingRelease,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Reservation {
    pub id: Uuid,
    pub desk_id: Uuid,
    pub employee_id: Uuid,
    pub date: NaiveDate,
    pub status: ReservationStatus,
    pub check_in_time: Option<DateTime<Local>>,
    pub created_at: DateTime<Local>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ReservationStatus {
    Confirmed,
    CheckedIn,
    Cancelled,
    Expired,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Notification {
    pub id: Uuid,
    pub recipient_id: Uuid,
    pub message: String,
    pub created_at: DateTime<Local>,
    pub is_read: bool,
}

impl Department {
    pub fn new(name: String, floor: i32) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            floor,
            manager_id: None,
        }
    }
}

impl Employee {
    pub fn new(name: String, department_id: Uuid) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            department_id,
            is_active: true,
            join_date: Local::now().date_naive(),
            leave_date: None,
        }
    }
}

impl Desk {
    pub fn new(code: String, desk_type: DeskType, floor: i32) -> Self {
        Self {
            id: Uuid::new_v4(),
            code,
            desk_type,
            floor,
            status: DeskStatus::Available,
            current_employee_id: None,
            reservation_expiry_date: None,
            reserved_for_department: None,
        }
    }
}

impl Reservation {
    pub fn new(desk_id: Uuid, employee_id: Uuid, date: NaiveDate) -> Self {
        Self {
            id: Uuid::new_v4(),
            desk_id,
            employee_id,
            date,
            status: ReservationStatus::Confirmed,
            check_in_time: None,
            created_at: Local::now(),
        }
    }
}

impl Notification {
    pub fn new(recipient_id: Uuid, message: String) -> Self {
        Self {
            id: Uuid::new_v4(),
            recipient_id,
            message,
            created_at: Local::now(),
            is_read: false,
        }
    }
}

pub fn now_naive() -> NaiveDate {
    Local::now().date_naive()
}

pub fn is_weekend(date: NaiveDate) -> bool {
    use chrono::Weekday::*;
    matches!(date.weekday(), Sat | Sun)
}

pub fn add_working_days(mut date: NaiveDate, days: i32) -> NaiveDate {
    let mut added = 0;
    while added < days {
        date = date.succ_opt().unwrap_or(date);
        if !is_weekend(date) {
            added += 1;
        }
    }
    date
}
