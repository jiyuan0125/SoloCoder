use thiserror::Error;

#[derive(Error, Debug)]
pub enum SystemError {
    #[error("Supplier not found: {0}")]
    SupplierNotFound(String),
    
    #[error("Store not found: {0}")]
    StoreNotFound(String),
    
    #[error("Ingredient not found: {0}")]
    IngredientNotFound(String),
    
    #[error("Purchase request not found: {0}")]
    PurchaseRequestNotFound(String),
    
    #[error("Quotation not found: {0}")]
    QuotationNotFound(String),
    
    #[error("Order not found: {0}")]
    OrderNotFound(String),
    
    #[error("Supplier is not active: {0}")]
    SupplierNotActive(String),
    
    #[error("Supplier does not offer ingredient: {supplier}, {ingredient}")]
    SupplierDoesNotOfferIngredient { supplier: String, ingredient: String },
    
    #[error("Insufficient stock: {available} < {requested}")]
    InsufficientStock { available: f64, requested: f64 },
    
    #[error("Minimum order quantity not met: {min} > {requested}")]
    MinOrderQuantityNotMet { min: f64, requested: f64 },
    
    #[error("Emergency purchase limit exceeded: single {0} > 2000")]
    EmergencyPurchaseSingleLimitExceeded(f64),
    
    #[error("Emergency purchase monthly limit exceeded: {0}% > 10%")]
    EmergencyPurchaseMonthlyLimitExceeded(f64),
    
    #[error("Non-recommended supplier selected without reason")]
    NonRecommendedSupplierWithoutReason,
    
    #[error("Quotation not available for this purchase request")]
    QuotationNotAvailable,
    
    #[error("Invalid purchase status transition")]
    InvalidPurchaseStatus,
    
    #[error("Concurrent operation conflict, please retry")]
    ConcurrentConflict,
    
    #[error("Invalid input: {0}")]
    InvalidInput(String),
    
    #[error("Internal error: {0}")]
    Internal(String),
}
