use thiserror::Error;

#[derive(Debug, Error)]
pub enum WorkflowError {
    #[error("Document not found")]
    DocumentNotFound,
    #[error("Invalid transition: cannot {action} from {current_stage}")]
    InvalidTransition {
        action: String,
        current_stage: String,
    },
    #[error("Return reason is required")]
    ReturnReasonRequired,
    #[error("Permission denied")]
    PermissionDenied,
    #[error("Document already archived")]
    DocumentArchived,
    #[error("Internal error: {0}")]
    Internal(String),
}
