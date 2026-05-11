use thiserror::Error;

#[derive(Debug, Error)]
pub enum ColdChainError {
    #[error("Shipment not found: {0}")]
    ShipmentNotFound(String),

    #[error("Invalid state transition: {0}")]
    InvalidStateTransition(String),

    #[error("Shipment is not in transit")]
    ShipmentNotInTransit,

    #[error("Shipment has broken chain, cannot be accepted")]
    BrokenChainCannotAccept,

    #[error("Internal error: {0}")]
    Internal(String),
}
