use thiserror::Error;

#[derive(Error, Debug)]
pub enum PurchaseReturnError {
    #[error("Product not found: {0}")]
    ProductNotFound(String),
    
    #[error("Supplier not found: {0}")]
    SupplierNotFound(String),
    
    #[error("Batch not found: {0}")]
    BatchNotFound(String),
    
    #[error("Return not found: {0}")]
    ReturnNotFound(String),
    
    #[error("Exchange not found: {0}")]
    ExchangeNotFound(String),
    
    #[error("Insufficient quantity in batch {batch_id}: requested {requested}, available {available}")]
    InsufficientQuantity {
        batch_id: String,
        requested: u32,
        available: u32,
    },
    
    #[error("Invalid price: {0}")]
    InvalidPrice(String),
    
    #[error("Invalid quantity: {0}")]
    InvalidQuantity(String),
    
    #[error("Exchange products must be different")]
    SameProductExchange,
    
    #[error("Internal error: {0}")]
    Internal(String),
}

pub type Result<T> = std::result::Result<T, PurchaseReturnError>;
