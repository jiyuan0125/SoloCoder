use thiserror::Error;

#[derive(Debug, Error)]
pub enum BargainError {
    #[error("Product not found")]
    ProductNotFound,
    
    #[error("Activity not found")]
    ActivityNotFound,
    
    #[error("User already has an active activity for this product")]
    DuplicateActivity,
    
    #[error("User has already helped bargain in this activity")]
    AlreadyBargained,
    
    #[error("Activity has expired")]
    ActivityExpired,
    
    #[error("Activity is not active")]
    ActivityNotActive,
    
    #[error("Invalid price: floor price must be less than original price")]
    InvalidPrice,
    
    #[error("Internal error: {0}")]
    Internal(String),
}

pub type Result<T> = std::result::Result<T, BargainError>;
