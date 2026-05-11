use serde::{Deserialize, Serialize};
use chrono::{DateTime, Utc};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum SpaceType {
    Normal,
    Charging,
    Accessible,
}

impl SpaceType {
    pub fn as_str(&self) -> &'static str {
        match self {
            SpaceType::Normal => "normal",
            SpaceType::Charging => "charging",
            SpaceType::Accessible => "accessible",
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum VehicleType {
    Regular,
    Electric,
    Disabled,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ParkingSpace {
    pub id: String,
    pub space_type: SpaceType,
    pub is_occupied: bool,
    pub reserved_for_user: Option<String>,
}

impl ParkingSpace {
    pub fn new(id: String, space_type: SpaceType) -> Self {
        Self {
            id,
            space_type,
            is_occupied: false,
            reserved_for_user: None,
        }
    }

    pub fn is_available_for(&self, vehicle_type: VehicleType, user_id: Option<&String>) -> bool {
        if self.is_occupied {
            return false;
        }

        if let Some(reserved) = &self.reserved_for_user {
            if let Some(user) = user_id {
                if user != reserved {
                    return false;
                }
            } else {
                return false;
            }
        }

        match (self.space_type, vehicle_type) {
            (SpaceType::Normal, _) => true,
            (SpaceType::Charging, VehicleType::Electric) => true,
            (SpaceType::Accessible, VehicleType::Disabled) => true,
            _ => false,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MonthlyCard {
    pub user_id: String,
    pub vehicle_plate: String,
    pub start_time: DateTime<Utc>,
    pub end_time: DateTime<Utc>,
    pub reserved_space_id: Option<String>,
}

impl MonthlyCard {
    pub fn is_active(&self) -> bool {
        let now = Utc::now();
        now >= self.start_time && now <= self.end_time
    }

    pub fn days_until_expiry(&self) -> i64 {
        let now = Utc::now();
        let duration = self.end_time - now;
        duration.num_days()
    }

    pub fn needs_reminder(&self) -> bool {
        self.days_until_expiry() <= 7 && self.days_until_expiry() >= 0
    }

    pub fn is_expired(&self) -> bool {
        Utc::now() > self.end_time
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ParkingRecord {
    pub id: String,
    pub plate_number: String,
    pub vehicle_type: VehicleType,
    pub space_id: String,
    pub entry_time: DateTime<Utc>,
    pub exit_time: Option<DateTime<Utc>>,
    pub fee: Option<f64>,
    pub is_monthly: bool,
}

impl ParkingRecord {
    pub fn new(plate_number: String, vehicle_type: VehicleType, space_id: String, is_monthly: bool) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            plate_number,
            vehicle_type,
            space_id,
            entry_time: Utc::now(),
            exit_time: None,
            fee: None,
            is_monthly,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct User {
    pub id: String,
    pub name: String,
    pub phone: String,
}

impl User {
    pub fn new(id: String, name: String, phone: String) -> Self {
        Self { id, name, phone }
    }
}
