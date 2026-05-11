use thiserror::Error;

#[derive(Error, Debug)]
pub enum WarehouseError {
    #[error("Warehouse not found: {0}")]
    WarehouseNotFound(String),
    
    #[error("Product not found in warehouse: {product_id}")]
    ProductNotFound { product_id: String },
    
    #[error("Insufficient available stock: available={available}, requested={requested}")]
    InsufficientStock { available: u32, requested: u32 },
    
    #[error("Transfer order not found: {0}")]
    TransferNotFound(String),
    
    #[error("Invalid transfer status for this operation: current={current}")]
    InvalidStatus { current: String },
    
    #[error("Cannot cancel: transfer issued more than 24 hours and arrived at destination city")]
    CannotCancel,
    
    #[error("Chain transfer not allowed: warehouse {warehouse_id} has in-transit stock")]
    ChainTransferNotAllowed { warehouse_id: String },
    
    #[error("Product {product_id} is locked by another operation")]
    ConcurrentLockConflict { product_id: String },
    
    #[error("Target warehouse must be different from source warehouse")]
    SameWarehouse,
}
