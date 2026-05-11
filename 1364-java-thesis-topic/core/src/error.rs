use thiserror::Error;

#[derive(Debug, Error)]
pub enum ThesisError {
    #[error("System phase error: {0}")]
    PhaseError(String),

    #[error("Student not found: {0}")]
    StudentNotFound(String),

    #[error("Advisor not found: {0}")]
    AdvisorNotFound(String),

    #[error("Student cannot modify preferences")]
    CannotModify,

    #[error("Invalid preferences: must have 1-3 unique advisors")]
    InvalidPreferences,

    #[error("Student already submitted")]
    AlreadySubmitted,

    #[error("Student not submitted yet")]
    NotSubmitted,

    #[error("Advisor has no capacity")]
    NoCapacity,

    #[error("Student not eligible for this advisor")]
    NotEligible,

    #[error("Student is locked or already assigned")]
    StudentLocked,

    #[error("Concurrent conflict: {0}")]
    ConcurrentConflict(String),

    #[error("Invalid operation: {0}")]
    InvalidOperation(String),

    #[error("Internal error: {0}")]
    Internal(String),
}

pub type Result<T> = std::result::Result<T, ThesisError>;
