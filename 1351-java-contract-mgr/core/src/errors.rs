use thiserror::Error;

#[derive(Debug, Error)]
pub enum ContractError {
    #[error("Contract not found: {0}")]
    NotFound(String),
    #[error("Invalid state transition: {current} -> {target}")]
    InvalidStateTransition { current: String, target: String },
    #[error("Invalid operation: {0}")]
    InvalidOperation(String),
    #[error("Validation error: {0}")]
    Validation(String),
    #[error("Framework contract has active sub-contracts")]
    HasActiveSubContracts,
    #[error("Price adjustment pending approval")]
    PriceAdjustmentPending,
    #[error("Approval request not found: {0}")]
    ApprovalNotFound(String),
}
