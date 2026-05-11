use thiserror::Error;

#[derive(Debug, Error)]
pub enum AppError {
    #[error("Equipment not found: {0}")]
    EquipmentNotFound(String),
    
    #[error("Consumable not found: {0}")]
    ConsumableNotFound(String),
    
    #[error("User not found: {0}")]
    UserNotFound(String),
    
    #[error("Booking not found: {0}")]
    BookingNotFound(String),
    
    #[error("Time slot conflict")]
    TimeSlotConflict,
    
    #[error("Invalid time range")]
    InvalidTimeRange,
    
    #[error("Invalid status transition: from {from:?} to {to:?}")]
    InvalidStatusTransition { from: String, to: String },
    
    #[error("Insufficient stock for consumable {name}: required {required}, available {available}")]
    InsufficientStock {
        name: String,
        required: u32,
        available: u32,
    },
    
    #[error("Cannot cancel booking that has already started")]
    CannotCancelStartedBooking,
    
    #[error("Unauthorized action")]
    Unauthorized,
    
    #[error("Internal error: {0}")]
    Internal(String),
}

pub type AppResult<T> = Result<T, AppError>;
