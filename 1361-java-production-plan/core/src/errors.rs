use thiserror::Error;

#[derive(Debug, Error)]
pub enum SchedulerError {
    #[error("Production line not found: {0}")]
    LineNotFound(String),
    
    #[error("Order not found: {0}")]
    OrderNotFound(String),
    
    #[error("Device not found: {0}")]
    DeviceNotFound(String),
    
    #[error("Maintenance window not found: {0}")]
    MaintenanceWindowNotFound(String),
    
    #[error("Order already assigned to a line")]
    OrderAlreadyAssigned,
    
    #[error("Order is already scheduled")]
    OrderAlreadyScheduled,
    
    #[error("Invalid maintenance window: end time must be after start time")]
    InvalidMaintenanceWindow,
    
    #[error("No available production lines")]
    NoAvailableLines,
    
    #[error("Internal error: {0}")]
    InternalError(String),
}
