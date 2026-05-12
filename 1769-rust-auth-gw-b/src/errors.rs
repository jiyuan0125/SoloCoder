use actix_web::{HttpResponse, ResponseError};
use serde::Serialize;
use std::fmt;

use crate::auth::AuthError;

#[derive(Debug, Serialize)]
pub struct ErrorResponse {
    pub success: bool,
    pub error: ErrorDetail,
}

#[derive(Debug, Serialize)]
pub struct ErrorDetail {
    pub code: String,
    pub message: String,
}

impl fmt::Display for ErrorResponse {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.error.message)
    }
}

impl ResponseError for ErrorResponse {
    fn error_response(&self) -> HttpResponse {
        HttpResponse::Unauthorized().json(self)
    }
}

impl From<AuthError> for ErrorResponse {
    fn from(err: AuthError) -> Self {
        let (code, message) = match err {
            AuthError::TokenExpired => (
                "TOKEN_EXPIRED".to_string(),
                "Access token has expired. Please use refresh token to get a new one.".to_string(),
            ),
            AuthError::InvalidToken => (
                "INVALID_TOKEN".to_string(),
                "Invalid or malformed token.".to_string(),
            ),
            AuthError::MissingToken => (
                "MISSING_TOKEN".to_string(),
                "No authorization token provided.".to_string(),
            ),
            AuthError::InvalidCredentials => (
                "INVALID_CREDENTIALS".to_string(),
                "Invalid username or password.".to_string(),
            ),
        };

        ErrorResponse {
            success: false,
            error: ErrorDetail { code, message },
        }
    }
}

pub fn unauthorized(code: &str, message: &str) -> HttpResponse {
    HttpResponse::Unauthorized().json(ErrorResponse {
        success: false,
        error: ErrorDetail {
            code: code.to_string(),
            message: message.to_string(),
        },
    })
}

pub fn internal_error(message: &str) -> HttpResponse {
    HttpResponse::InternalServerError().json(ErrorResponse {
        success: false,
        error: ErrorDetail {
            code: "INTERNAL_ERROR".to_string(),
            message: message.to_string(),
        },
    })
}
