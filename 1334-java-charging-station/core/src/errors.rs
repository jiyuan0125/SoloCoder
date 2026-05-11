use thiserror::Error;

#[derive(Debug, Error)]
pub enum ChargingError {
    #[error("Charger not found: {0}")]
    ChargerNotFound(String),
    
    #[error("Charger is busy: {0}")]
    ChargerBusy(String),
    
    #[error("Order not found: {0}")]
    OrderNotFound(String),
    
    #[error("Order already completed: {0}")]
    OrderAlreadyCompleted(String),
    
    #[error("Invalid request: {0}")]
    InvalidRequest(String),
}
