use thiserror::Error;

#[derive(Error, Debug)]
pub enum ParkingError {
    #[error("No available parking space")]
    NoSpaceAvailable,
    
    #[error("Vehicle already parked: {0}")]
    VehicleAlreadyParked(String),
    
    #[error("Vehicle not found: {0}")]
    VehicleNotFound(String),
    
    #[error("User not found: {0}")]
    UserNotFound(String),
    
    #[error("Monthly card not found for user: {0}")]
    MonthlyCardNotFound(String),
    
    #[error("Invalid vehicle type for space")]
    InvalidVehicleType,
    
    #[error("Space not found: {0}")]
    SpaceNotFound(String),
    
    #[error("Space already reserved")]
    SpaceAlreadyReserved,
    
    #[error("Parking record not found: {0}")]
    ParkingRecordNotFound(String),
    
    #[error("Database error: {0}")]
    DatabaseError(String),
    
    #[error("Internal error: {0}")]
    InternalError(String),
}

pub type Result<T> = std::result::Result<T, ParkingError>;
