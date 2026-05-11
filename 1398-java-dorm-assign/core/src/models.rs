use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;
use std::str::FromStr;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum RoomType {
    FourPerson,
    SixPerson,
}

impl RoomType {
    pub fn bed_count(&self) -> u32 {
        match self {
            RoomType::FourPerson => 4,
            RoomType::SixPerson => 6,
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum HygieneScore {
    A,
    B,
    C,
}

impl FromStr for HygieneScore {
    type Err = crate::error::DormError;
    
    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s.to_uppercase().as_str() {
            "A" => Ok(HygieneScore::A),
            "B" => Ok(HygieneScore::B),
            "C" => Ok(HygieneScore::C),
            _ => Err(crate::error::DormError::InvalidHygieneScore(s.to_string())),
        }
    }
}

impl std::fmt::Display for HygieneScore {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            HygieneScore::A => write!(f, "A"),
            HygieneScore::B => write!(f, "B"),
            HygieneScore::C => write!(f, "C"),
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum RoomStatus {
    Normal,
    UnderMaintenance,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum SwapRequestStatus {
    Pending,
    Confirmed,
    Rejected,
    Expired,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum StudentStatus {
    Unassigned,
    Assigned,
    PendingTransfer,
    WaitingForSwapConfirmation,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Building {
    pub id: String,
    pub name: String,
    pub floor_count: u32,
    pub rooms_per_floor: u32,
    pub created_at: DateTime<Utc>,
}

impl Building {
    pub fn new(name: String, floor_count: u32, rooms_per_floor: u32) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            name,
            floor_count,
            rooms_per_floor,
            created_at: Utc::now(),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Room {
    pub id: String,
    pub building_id: String,
    pub floor_number: u32,
    pub room_number: u32,
    pub room_type: RoomType,
    pub status: RoomStatus,
    pub hygiene_score: Option<HygieneScore>,
    pub created_at: DateTime<Utc>,
}

impl Room {
    pub fn new(
        building_id: String,
        floor_number: u32,
        room_number: u32,
        room_type: RoomType,
    ) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            building_id,
            floor_number,
            room_number,
            room_type,
            status: RoomStatus::Normal,
            hygiene_score: None,
            created_at: Utc::now(),
        }
    }
    
    pub fn bed_count(&self) -> u32 {
        self.room_type.bed_count()
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Bed {
    pub id: String,
    pub room_id: String,
    pub bed_number: u32,
    pub student_id: Option<String>,
    pub created_at: DateTime<Utc>,
}

impl Bed {
    pub fn new(room_id: String, bed_number: u32) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            room_id,
            bed_number,
            student_id: None,
            created_at: Utc::now(),
        }
    }
    
    pub fn is_occupied(&self) -> bool {
        self.student_id.is_some()
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Student {
    pub id: String,
    pub student_id: String,
    pub name: String,
    pub department: String,
    pub status: StudentStatus,
    pub bed_id: Option<String>,
    pub created_at: DateTime<Utc>,
}

impl Student {
    pub fn new(student_id: String, name: String, department: String) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            student_id,
            name,
            department,
            status: StudentStatus::Unassigned,
            bed_id: None,
            created_at: Utc::now(),
        }
    }
    
    pub fn is_assigned(&self) -> bool {
        matches!(self.status, StudentStatus::Assigned)
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SwapRequest {
    pub id: String,
    pub requester_id: String,
    pub target_id: String,
    pub requester_bed_id: String,
    pub target_bed_id: String,
    pub status: SwapRequestStatus,
    pub created_at: DateTime<Utc>,
    pub expires_at: DateTime<Utc>,
}

impl SwapRequest {
    pub fn new(
        requester_id: String,
        target_id: String,
        requester_bed_id: String,
        target_bed_id: String,
    ) -> Self {
        let now = Utc::now();
        let expires_at = now + chrono::Duration::days(7);
        
        Self {
            id: Uuid::new_v4().to_string(),
            requester_id,
            target_id,
            requester_bed_id,
            target_bed_id,
            status: SwapRequestStatus::Pending,
            created_at: now,
            expires_at,
        }
    }
    
    pub fn is_expired(&self) -> bool {
        Utc::now() > self.expires_at
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AllocationResult {
    pub department: String,
    pub total_students: usize,
    pub assigned_count: usize,
    pub unassigned_count: usize,
    pub assigned_students: Vec<String>,
    pub unassigned_students: Vec<String>,
    pub message: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateBuildingRequest {
    pub name: String,
    pub floor_count: u32,
    pub rooms_per_floor: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateRoomRequest {
    pub building_id: String,
    pub floor_number: u32,
    pub room_number: u32,
    pub room_type: RoomType,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateStudentRequest {
    pub student_id: String,
    pub name: String,
    pub department: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AllocateDepartmentRequest {
    pub department: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateSwapRequest {
    pub requester_id: String,
    pub target_id: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RespondSwapRequest {
    pub accept: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UpdateHygieneScoreRequest {
    pub score: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UpdateRoomStatusRequest {
    pub status: RoomStatus,
}
