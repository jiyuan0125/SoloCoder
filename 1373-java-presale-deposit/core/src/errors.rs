use thiserror::Error;
use uuid::Uuid;

#[derive(Debug, Error)]
pub enum PresaleError {
    #[error("Activity not found: {0}")]
    ActivityNotFound(Uuid),

    #[error("Order not found: {0}")]
    OrderNotFound(Uuid),

    #[error("Product not found: {0}")]
    ProductNotFound(Uuid),

    #[error("Activity is not active")]
    ActivityNotActive,

    #[error("Activity has reached maximum participants")]
    ActivityFull,

    #[error("User already has an order for this product")]
    DuplicateOrder,

    #[error("Cannot cancel order after deposit paid")]
    CannotCancelDepositPaid,

    #[error("Order is in invalid status for this operation: {0:?}")]
    InvalidOrderStatus(crate::models::OrderStatus),

    #[error("Activity is in invalid status for this operation: {0:?}")]
    InvalidActivityStatus(crate::models::ActivityStatus),

    #[error("Final payment overdue")]
    FinalPaymentOverdue,

    #[error("Final payment amount mismatch")]
    FinalPaymentAmountMismatch,

    #[error("Invalid configuration: {0}")]
    InvalidConfig(String),

    #[error("Internal error: {0}")]
    Internal(String),
}

pub type PresaleResult<T> = Result<T, PresaleError>;
