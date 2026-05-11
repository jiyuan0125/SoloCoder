use thiserror::Error;

#[derive(Error, Debug)]
pub enum LeaseError {
    #[error("Room not found: {0}")]
    RoomNotFound(String),
    #[error("Room already exists: {0}")]
    RoomAlreadyExists(String),
    #[error("Contract not found: {0}")]
    ContractNotFound(String),
    #[error("Tenant not found: {0}")]
    TenantNotFound(String),
    #[error("Room has active contract: {0}")]
    RoomHasActiveContract(String),
    #[error("Contract is not active")]
    ContractNotActive,
    #[error("Contract is not in fixed-term mode")]
    ContractNotFixedTerm,
    #[error("Contract is already in month-to-month mode")]
    ContractAlreadyMonthToMonth,
    #[error("Invalid date: {0}")]
    InvalidDate(String),
    #[error("Invalid amount: {0}")]
    InvalidAmount(String),
    #[error("Validation error: {0}")]
    ValidationError(String),
    #[error("Internal error: {0}")]
    InternalError(String),
}
