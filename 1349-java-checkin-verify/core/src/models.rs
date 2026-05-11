use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GpsCoordinate {
    pub latitude: f64,
    pub longitude: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Activity {
    pub id: Uuid,
    pub name: String,
    pub checkin_start: DateTime<Utc>,
    pub checkin_end: DateTime<Utc>,
    pub checkout_deadline: DateTime<Utc>,
    pub location: GpsCoordinate,
    pub allowed_distance_meters: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Employee {
    pub id: Uuid,
    pub name: String,
    pub employee_number: String,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum CheckinStatus {
    NotCheckedIn,
    Normal,
    EarlyCheckout,
    NotCheckedOut,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CheckinRecord {
    pub activity_id: Uuid,
    pub employee_id: Uuid,
    pub device_id: String,
    pub checkin_time: Option<DateTime<Utc>>,
    pub checkout_time: Option<DateTime<Utc>>,
    pub status: CheckinStatus,
    pub checkin_location: Option<GpsCoordinate>,
    pub checkout_location: Option<GpsCoordinate>,
}

impl CheckinRecord {
    pub fn new(activity_id: Uuid, employee_id: Uuid, device_id: String) -> Self {
        Self {
            activity_id,
            employee_id,
            device_id,
            checkin_time: None,
            checkout_time: None,
            status: CheckinStatus::NotCheckedIn,
            checkin_location: None,
            checkout_location: None,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CheckinReport {
    pub activity_id: Uuid,
    pub total_count: usize,
    pub checked_in_count: usize,
    pub checkin_rate: f64,
    pub normal_count: usize,
    pub not_checked_out_count: usize,
    pub early_checkout_count: usize,
    pub not_checked_in_count: usize,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CheckinRequest {
    pub employee_id: Uuid,
    pub device_id: String,
    pub location: GpsCoordinate,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CheckoutRequest {
    pub employee_id: Uuid,
    pub location: GpsCoordinate,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateActivityRequest {
    pub name: String,
    pub checkin_start: DateTime<Utc>,
    pub checkin_end: DateTime<Utc>,
    pub checkout_deadline: DateTime<Utc>,
    pub location: GpsCoordinate,
    pub allowed_distance_meters: f64,
    pub employee_ids: Vec<Uuid>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateEmployeeRequest {
    pub name: String,
    pub employee_number: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApiResponse<T> {
    pub success: bool,
    pub message: String,
    pub data: Option<T>,
}

impl<T> ApiResponse<T> {
    pub fn success(data: T) -> Self {
        Self {
            success: true,
            message: "操作成功".to_string(),
            data: Some(data),
        }
    }

    pub fn success_with_message(message: String, data: T) -> Self {
        Self {
            success: true,
            message,
            data: Some(data),
        }
    }

    pub fn error(message: String) -> Self {
        Self {
            success: false,
            message,
            data: None,
        }
    }
}
