use thiserror::Error;

#[derive(Error, Debug)]
pub enum RentalError {
    #[error("Property not found: {0}")]
    PropertyNotFound(String),
    
    #[error("Tenant not found: {0}")]
    TenantNotFound(String),
    
    #[error("Contract not found: {0}")]
    ContractNotFound(String),
    
    #[error("Checkout record not found: {0}")]
    CheckoutRecordNotFound(String),
    
    #[error("Property number already exists: {0}")]
    PropertyNoExists(String),
    
    #[error("Tenant phone already exists: {0}")]
    TenantPhoneExists(String),
    
    #[error("Invalid date: {0}")]
    InvalidDate(String),
    
    #[error("Contract not active: {0}")]
    ContractNotActive(String),
    
    #[error("Property is currently rented")]
    PropertyOccupied,
    
    #[error("Invalid damage cost: damage cost cannot be negative")]
    InvalidDamageCost,
    
    #[error("Internal error: {0}")]
    Internal(String),
}
