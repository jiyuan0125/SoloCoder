use chrono::{Local, NaiveDate};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum TimeSlot {
    Morning,
    Afternoon,
}

impl TimeSlot {
    pub fn max_duration_hours(&self) -> u32 {
        match self {
            TimeSlot::Morning => 3,
            TimeSlot::Afternoon => 4,
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum BookingStatus {
    Reserved,
    InProgress,
    Completed,
    Cancelled,
    Expired,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CarModel {
    pub id: Uuid,
    pub name: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Car {
    pub id: Uuid,
    pub model_id: Uuid,
    pub plate_number: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Customer {
    pub id: Uuid,
    pub name: String,
    pub phone: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SalesAdvisor {
    pub id: Uuid,
    pub name: String,
    pub on_duty: bool,
    pub booking_count: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Booking {
    pub id: Uuid,
    pub customer_id: Uuid,
    pub car_id: Uuid,
    pub model_id: Uuid,
    pub advisor_id: Uuid,
    pub date: NaiveDate,
    pub time_slot: TimeSlot,
    pub status: BookingStatus,
    pub created_at: chrono::DateTime<Local>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Feedback {
    pub booking_id: Uuid,
    pub satisfaction: u8,
    pub purchase_intent: String,
    pub created_at: chrono::DateTime<Local>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateBookingRequest {
    pub customer_id: Uuid,
    pub model_id: Uuid,
    pub date: NaiveDate,
    pub time_slot: TimeSlot,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SubmitFeedbackRequest {
    pub satisfaction: u8,
    pub purchase_intent: String,
}
