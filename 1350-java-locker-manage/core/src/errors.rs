use thiserror::Error;

#[derive(Debug, Error)]
pub enum LockerError {
    #[error("Locker not found: {0}")]
    LockerNotFound(String),

    #[error("No available compartment for package size: {0:?}")]
    NoAvailableCompartment(crate::PackageSize),

    #[error("Invalid pickup code")]
    InvalidPickupCode,

    #[error("Pickup locked. Please try again after {0} minutes")]
    PickupLocked(u64),

    #[error("Pickup code generation failed after multiple attempts")]
    PickupCodeGenerationFailed,

    #[error("Package not found")]
    PackageNotFound,
}
