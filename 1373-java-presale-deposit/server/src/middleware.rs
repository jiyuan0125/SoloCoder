use axum::{
    http::StatusCode,
    response::{IntoResponse, Response},
    Json,
};
use presale_core::PresaleError;
use serde::Serialize;

#[derive(Debug, Serialize)]
pub struct ErrorResponse {
    pub error: String,
    pub message: String,
}

pub fn map_error_to_response(err: PresaleError) -> Response {
    let (status, error_code) = match &err {
        PresaleError::ActivityNotFound(_) => (StatusCode::NOT_FOUND, "ACTIVITY_NOT_FOUND"),
        PresaleError::OrderNotFound(_) => (StatusCode::NOT_FOUND, "ORDER_NOT_FOUND"),
        PresaleError::ProductNotFound(_) => (StatusCode::NOT_FOUND, "PRODUCT_NOT_FOUND"),
        PresaleError::ActivityNotActive => (StatusCode::BAD_REQUEST, "ACTIVITY_NOT_ACTIVE"),
        PresaleError::ActivityFull => (StatusCode::BAD_REQUEST, "ACTIVITY_FULL"),
        PresaleError::DuplicateOrder => (StatusCode::CONFLICT, "DUPLICATE_ORDER"),
        PresaleError::CannotCancelDepositPaid => (StatusCode::BAD_REQUEST, "CANNOT_CANCEL_DEPOSIT_PAID"),
        PresaleError::InvalidOrderStatus(_) => (StatusCode::BAD_REQUEST, "INVALID_ORDER_STATUS"),
        PresaleError::InvalidActivityStatus(_) => (StatusCode::BAD_REQUEST, "INVALID_ACTIVITY_STATUS"),
        PresaleError::FinalPaymentOverdue => (StatusCode::BAD_REQUEST, "FINAL_PAYMENT_OVERDUE"),
        PresaleError::FinalPaymentAmountMismatch => (StatusCode::BAD_REQUEST, "FINAL_PAYMENT_AMOUNT_MISMATCH"),
        PresaleError::InvalidConfig(_) => (StatusCode::BAD_REQUEST, "INVALID_CONFIG"),
        PresaleError::Internal(_) => (StatusCode::INTERNAL_SERVER_ERROR, "INTERNAL_ERROR"),
    };

    (
        status,
        Json(ErrorResponse {
            error: error_code.to_string(),
            message: err.to_string(),
        }),
    )
        .into_response()
}
