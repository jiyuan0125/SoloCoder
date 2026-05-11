use chrono::{NaiveDate, Weekday};
use serde::{Deserialize, Serialize};
use uuid::Uuid;
use std::collections::HashSet;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum TimeSlot {
    Morning,
    Afternoon,
    Evening,
}

impl TimeSlot {
    pub fn all() -> [Self; 3] {
        [TimeSlot::Morning, TimeSlot::Afternoon, TimeSlot::Evening]
    }

    pub fn description(&self) -> &'static str {
        match self {
            TimeSlot::Morning => "上午 08:00-12:00",
            TimeSlot::Afternoon => "下午 13:00-17:00",
            TimeSlot::Evening => "晚上 18:00-22:00",
        }
    }
}

impl std::fmt::Display for TimeSlot {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.description())
    }
}

#[derive(Debug, Clone, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum DiseaseType {
    None,
    HepatitisB,
    HepatitisC,
    Other(String),
}

impl DiseaseType {
    pub fn can_share_machine(&self, other: &Self) -> bool {
        match (self, other) {
            (DiseaseType::None, DiseaseType::None) => true,
            (DiseaseType::HepatitisB, DiseaseType::HepatitisB) => true,
            (DiseaseType::HepatitisC, DiseaseType::HepatitisC) => true,
            (DiseaseType::Other(a), DiseaseType::Other(b)) => a == b,
            _ => false,
        }
    }

    pub fn requires_special_machine(&self) -> bool {
        !matches!(self, DiseaseType::None)
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Patient {
    pub id: Uuid,
    pub name: String,
    pub disease: DiseaseType,
    pub dialysis_frequency_per_week: u8,
    pub preferred_days: HashSet<Weekday>,
    pub preferred_slots: HashSet<TimeSlot>,
}

impl Patient {
    pub fn new(
        name: String,
        disease: DiseaseType,
        dialysis_frequency_per_week: u8,
        preferred_days: HashSet<Weekday>,
        preferred_slots: HashSet<TimeSlot>,
    ) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            disease,
            dialysis_frequency_per_week,
            preferred_days,
            preferred_slots,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Machine {
    pub id: Uuid,
    pub name: String,
    pub is_hepatitis_b_only: bool,
}

impl Machine {
    pub fn new(name: String, is_hepatitis_b_only: bool) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            is_hepatitis_b_only,
        }
    }

    pub fn can_treat(&self, disease: &DiseaseType) -> bool {
        if self.is_hepatitis_b_only {
            matches!(disease, DiseaseType::HepatitisB)
        } else {
            !matches!(disease, DiseaseType::HepatitisB)
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ScheduleEntry {
    pub id: Uuid,
    pub patient_id: Uuid,
    pub machine_id: Uuid,
    pub date: NaiveDate,
    pub time_slot: TimeSlot,
}

impl ScheduleEntry {
    pub fn new(
        patient_id: Uuid,
        machine_id: Uuid,
        date: NaiveDate,
        time_slot: TimeSlot,
    ) -> Self {
        Self {
            id: Uuid::new_v4(),
            patient_id,
            machine_id,
            date,
            time_slot,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Maintenance {
    pub id: Uuid,
    pub machine_id: Uuid,
    pub date: NaiveDate,
    pub time_slot: TimeSlot,
    pub is_emergency: bool,
    pub description: String,
}

impl Maintenance {
    pub fn new(
        machine_id: Uuid,
        date: NaiveDate,
        time_slot: TimeSlot,
        is_emergency: bool,
        description: String,
    ) -> Self {
        Self {
            id: Uuid::new_v4(),
            machine_id,
            date,
            time_slot,
            is_emergency,
            description,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreatePatientRequest {
    pub name: String,
    pub disease: DiseaseType,
    pub dialysis_frequency_per_week: u8,
    pub preferred_days: Vec<Weekday>,
    pub preferred_slots: Vec<TimeSlot>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateMachineRequest {
    pub name: String,
    pub is_hepatitis_b_only: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateScheduleRequest {
    pub patient_id: Uuid,
    pub machine_id: Uuid,
    pub date: NaiveDate,
    pub time_slot: TimeSlot,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BulkCreateScheduleRequest {
    pub schedules: Vec<CreateScheduleRequest>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RescheduleRequest {
    pub schedule_id: Uuid,
    pub new_date: NaiveDate,
    pub new_time_slot: TimeSlot,
    pub new_machine_id: Option<Uuid>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateMaintenanceRequest {
    pub machine_id: Uuid,
    pub date: NaiveDate,
    pub time_slot: TimeSlot,
    pub is_emergency: bool,
    pub description: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RescheduleResult {
    pub original_schedule_id: Uuid,
    pub new_schedule: Option<ScheduleEntry>,
    pub canceled: bool,
    pub message: String,
}
