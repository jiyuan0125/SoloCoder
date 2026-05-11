use thiserror::Error;
use uuid::Uuid;

#[derive(Debug, Error)]
pub enum SystemError {
    #[error("Employee not found: {0}")]
    EmployeeNotFound(Uuid),
    #[error("Overtime not found: {0}")]
    OvertimeNotFound(Uuid),
    #[error("Leave not found: {0}")]
    LeaveNotFound(Uuid),
    #[error("Overtime expired: {0}")]
    OvertimeExpired(Uuid),
    #[error("Overtime already processed: {0}")]
    OvertimeAlreadyProcessed(Uuid),
    #[error("Leave already processed: {0}")]
    LeaveAlreadyProcessed(Uuid),
    #[error("Insufficient leave balance: employee {0}, required {1}h, available {2}h")]
    InsufficientLeaveBalance(Uuid, f64, f64),
    #[error("Cannot cancel past leave: {0}")]
    CannotCancelPastLeave(Uuid),
    #[error("Weekend compensation type required for weekend overtime")]
    WeekendCompensationRequired,
    #[error("Weekend compensation type cannot be changed")]
    WeekendCompensationCannotChange,
    #[error("Holiday overtime cannot be compensated with leave")]
    HolidayCannotBeLeave,
    #[error("Invalid overtime period: start time must be before end time")]
    InvalidOvertimePeriod,
    #[error("Invalid leave period: start time must be before end time")]
    InvalidLeavePeriod,
    #[error("Internal error: {0}")]
    InternalError(String),
}
