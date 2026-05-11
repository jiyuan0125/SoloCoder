use thiserror::Error;
use std::fmt;

#[derive(Debug, Error)]
pub enum BusinessError {
    #[error("Pickup point not found")]
    PickupPointNotFound,
    
    #[error("Product not found")]
    ProductNotFound,
    
    #[error("Order not found")]
    OrderNotFound,
    
    #[error("Sorting list not found")]
    SortingListNotFound,
    
    #[error("Insufficient stock: available {available}, requested {requested}")]
    InsufficientStock { available: i64, requested: i64 },
    
    #[error("Order already paid")]
    OrderAlreadyPaid,
    
    #[error("Order not paid")]
    OrderNotPaid,
    
    #[error("Payment deadline passed")]
    PaymentDeadlinePassed,
    
    #[error("Cut off time passed, operation not allowed")]
    CutOffTimePassed,
    
    #[error("Order cannot be refunded after cut off time")]
    RefundAfterCutOffTime,
    
    #[error("Invalid operation: {0}")]
    InvalidOperation(String),
    
    #[error("Product does not belong to pickup point")]
    ProductNotInPickupPoint,
    
    #[error("Empty order")]
    EmptyOrder,
    
    #[error("Pickup point already exists")]
    PickupPointAlreadyExists,
}

#[derive(Debug)]
pub struct ApiError(pub BusinessError);

impl fmt::Display for ApiError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.0)
    }
}

impl std::error::Error for ApiError {}

impl From<BusinessError> for ApiError {
    fn from(err: BusinessError) -> Self {
        ApiError(err)
    }
}

impl BusinessError {
    pub fn http_status_code(&self) -> u16 {
        match self {
            BusinessError::PickupPointNotFound => 404,
            BusinessError::ProductNotFound => 404,
            BusinessError::OrderNotFound => 404,
            BusinessError::SortingListNotFound => 404,
            _ => 400,
        }
    }
}
