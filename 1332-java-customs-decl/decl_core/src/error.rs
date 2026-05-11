use thiserror::Error;

#[derive(Debug, Error)]
pub enum AppError {
    #[error("Declaration not found: {0}")]
    DeclarationNotFound(String),

    #[error("Exchange rate not found for currency {0:?} on {1}")]
    ExchangeRateNotFound(String, String),

    #[error("Exchange rate already exists for currency {0:?} on {1}")]
    ExchangeRateAlreadyExists(String, String),

    #[error("Declaration has invalid status: expected {expected}, got {actual}")]
    InvalidStatus { expected: String, actual: String },

    #[error("Declaration no already exists: {0}")]
    DeclarationNoAlreadyExists(String),

    #[error("Cannot merge declaration with status {0:?}")]
    CannotMergeStatus(String),

    #[error("Merge requires at least 2 declarations")]
    MergeRequiresAtLeastTwo,

    #[error("HS code {0} has different names: {1} vs {2}")]
    HsCodeNameMismatch(String, String, String),

    #[error("Invalid input: {0}")]
    InvalidInput(String),

    #[error("Internal error: {0}")]
    Internal(String),
}

impl AppError {
    pub fn http_status(&self) -> u16 {
        match self {
            AppError::DeclarationNotFound(_) => 404,
            AppError::ExchangeRateNotFound(_, _) => 400,
            AppError::ExchangeRateAlreadyExists(_, _) => 409,
            AppError::InvalidStatus { .. } => 400,
            AppError::DeclarationNoAlreadyExists(_) => 409,
            AppError::CannotMergeStatus(_) => 400,
            AppError::MergeRequiresAtLeastTwo => 400,
            AppError::HsCodeNameMismatch(..) => 400,
            AppError::InvalidInput(_) => 400,
            AppError::Internal(_) => 500,
        }
    }
}
