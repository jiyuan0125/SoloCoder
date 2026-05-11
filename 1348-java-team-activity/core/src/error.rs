use thiserror::Error;

#[derive(Debug, Error, Clone)]
pub enum ActivityError {
    #[error("Department not found")]
    DepartmentNotFound,
    #[error("Activity not found")]
    ActivityNotFound,
    #[error("User already registered for this activity")]
    UserAlreadyRegistered,
    #[error("Activity is full and waiting list is enabled")]
    ActivityFull,
    #[error("Insufficient budget")]
    InsufficientBudget,
    #[error("Registration deadline has passed")]
    RegistrationClosed,
    #[error("Cannot increase registration limit")]
    CannotIncreaseLimit,
    #[error("Activity is not in a valid state for this operation")]
    InvalidState,
    #[error("Participant count must match registered count")]
    ParticipantCountMismatch,
    #[error("User not registered for this activity")]
    UserNotRegistered,
    #[error("Cannot cancel activity after it has started")]
    CannotCancelStartedActivity,
    #[error("Concurrent registration conflict, please retry")]
    ConcurrentConflict,
}
