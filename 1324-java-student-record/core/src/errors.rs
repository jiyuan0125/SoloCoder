use thiserror::Error;

#[derive(Error, Debug)]
pub enum StudentRecordError {
    #[error("Student not found: {0}")]
    StudentNotFound(String),
    
    #[error("Major not found: {0}")]
    MajorNotFound(String),
    
    #[error("Course not found: {0}")]
    CourseNotFound(String),
    
    #[error("Student already in target major")]
    AlreadyInTargetMajor,
    
    #[error("Transfer in progress for this student")]
    TransferInProgress,
    
    #[error("Invalid input: {0}")]
    InvalidInput(String),
    
    #[error("Internal error: {0}")]
    InternalError(String),
}

pub type Result<T> = std::result::Result<T, StudentRecordError>;
