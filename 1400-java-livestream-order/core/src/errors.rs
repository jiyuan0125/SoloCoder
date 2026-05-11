use thiserror::Error;

#[derive(Error, Debug)]
pub enum SystemError {
    #[error("Product not found: {0}")]
    ProductNotFound(String),
    
    #[error("Order not found: {0}")]
    OrderNotFound(String),
    
    #[error("User not found: {0}")]
    UserNotFound(String),
    
    #[error("Insufficient stock: available {available}, requested {requested}")]
    InsufficientStock {
        available: u32,
        requested: u32,
    },
    
    #[error("Order lock expired")]
    OrderLockExpired,
    
    #[error("Order already paid")]
    OrderAlreadyPaid,
    
    #[error("Order already cancelled")]
    OrderAlreadyCancelled,
    
    #[error("Order not paid yet")]
    OrderNotPaid,
    
    #[error("Order already shipped")]
    OrderAlreadyShipped,
    
    #[error("Order already completed")]
    OrderAlreadyCompleted,
    
    #[error("Order already refunded")]
    OrderAlreadyRefunded,
    
    #[error("Invalid order status transition: from {from} to {to}")]
    InvalidStatusTransition {
        from: String,
        to: String,
    },
    
    #[error("Product not on sale")]
    ProductNotOnSale,
    
    #[error("Live stream not active")]
    LiveStreamNotActive,
    
    #[error("Purchase limit exceeded: user has {purchased}, limit is {limit}, requested {requested}")]
    PurchaseLimitExceeded {
        purchased: u32,
        limit: u32,
        requested: u32,
    },
    
    #[error("Internal error: {0}")]
    Internal(String),
}
