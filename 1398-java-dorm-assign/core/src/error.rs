use thiserror::Error;

#[derive(Error, Debug)]
pub enum DormError {
    #[error("Building not found: {0}")]
    BuildingNotFound(String),
    
    #[error("Room not found: {0}")]
    RoomNotFound(String),
    
    #[error("Bed not found: {0}")]
    BedNotFound(String),
    
    #[error("Student not found: {0}")]
    StudentNotFound(String),
    
    #[error("Swap request not found: {0}")]
    SwapRequestNotFound(String),
    
    #[error("Student {0} already has a bed assigned")]
    StudentAlreadyAssigned(String),
    
    #[error("Bed {0} is already occupied")]
    BedAlreadyOccupied(String),
    
    #[error("Room {0} is under maintenance")]
    RoomUnderMaintenance(String),
    
    #[error("Not enough beds available. Requested: {requested}, Available: {available}")]
    InsufficientBeds { requested: usize, available: usize },
    
    #[error("Cannot swap with the same student")]
    CannotSwapWithSelf,
    
    #[error("Swap request {0} has expired")]
    SwapRequestExpired(String),
    
    #[error("Swap request {0} is not pending")]
    SwapRequestNotPending(String),
    
    #[error("Student {0} is not the target of this swap request")]
    NotSwapTarget(String),
    
    #[error("Student {0} has a pending swap request")]
    HasPendingSwapRequest(String),
    
    #[error("Room number {room_number} already exists in building {building_id}")]
    RoomNumberExists { building_id: String, room_number: u32 },
    
    #[error("Floor number {floor} exceeds building's floor count {max_floors}")]
    FloorExceedsMax { floor: u32, max_floors: u32 },
    
    #[error("Invalid hygiene score: {0}. Must be A, B, or C.")]
    InvalidHygieneScore(String),
    
    #[error("Student {0} has no bed assigned")]
    NoBedAssigned(String),
    
    #[error("Concurrent allocation conflict")]
    ConcurrentConflict,
}
