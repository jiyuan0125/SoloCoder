use thiserror::Error;

#[derive(Debug, Error)]
pub enum ConfigError {
    #[error("namespace not found: {0}")]
    NamespaceNotFound(String),
    
    #[error("change request not found: {0}")]
    ChangeRequestNotFound(String),
    
    #[error("invalid status transition: {0}")]
    InvalidStatusTransition(String),
    
    #[error("permission denied: {0}")]
    PermissionDenied(String),
    
    #[error("type validation failed: {0}")]
    TypeValidationError(String),
    
    #[error("no approved version to rollback")]
    NoApprovedVersion,
    
    #[error("internal error: {0}")]
    Internal(String),
}
