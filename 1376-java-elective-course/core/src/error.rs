use thiserror::Error;

#[derive(Error, Debug)]
pub enum CourseSystemError {
    #[error("Course not found: {0}")]
    CourseNotFound(String),

    #[error("Student not found: {0}")]
    StudentNotFound(String),

    #[error("Prerequisite not satisfied: {0}")]
    PrerequisiteNotSatisfied(String),

    #[error("Course is full")]
    CourseFull,

    #[error("Credit limit exceeded. Current: {current}, Max: {max}, Adding: {adding}")]
    CreditLimitExceeded {
        current: u32,
        max: u32,
        adding: u32,
    },

    #[error("Student already enrolled in this course")]
    AlreadyEnrolled,

    #[error("Student not enrolled in this course")]
    NotEnrolled,

    #[error("Enrollment period has ended")]
    EnrollmentEnded,

    #[error("Concurrent access error: {0}")]
    ConcurrentError(String),
}
