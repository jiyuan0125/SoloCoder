use std::sync::Arc;
use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    Json,
};
use serde::{Deserialize, Serialize};
use crate::service::RetryIdempotentService;

#[derive(Debug, Deserialize)]
pub struct ExecuteRequest {
    pub idempotency_key: String,
    pub callback_url: String,
    pub payload: serde_json::Value,
}

#[derive(Debug, Serialize)]
pub struct ErrorResponse {
    pub error: String,
}

pub async fn health() -> impl IntoResponse {
    StatusCode::OK
}

pub async fn execute_request(
    State(service): State<Arc<RetryIdempotentService>>,
    Json(req): Json<ExecuteRequest>,
) -> impl IntoResponse {
    if req.idempotency_key.is_empty() {
        return (
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                error: "idempotency_key is required".to_string(),
            }),
        )
            .into_response();
    }

    if req.callback_url.is_empty() {
        return (
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                error: "callback_url is required".to_string(),
            }),
        )
            .into_response();
    }

    let response = service.process_request(&req.idempotency_key, &req.callback_url, &req.payload).await;
    (StatusCode::OK, Json(response)).into_response()
}

pub async fn query_request(
    State(service): State<Arc<RetryIdempotentService>>,
    Path(key): Path<String>,
) -> impl IntoResponse {
    match service.query_request(&key) {
        Some(query_response) => (StatusCode::OK, Json(query_response)).into_response(),
        None => (
            StatusCode::NOT_FOUND,
            Json(ErrorResponse {
                error: format!("Request with idempotency_key '{}' not found", key),
            }),
        )
            .into_response(),
    }
}

pub async fn get_statistics(
    State(service): State<Arc<RetryIdempotentService>>,
) -> impl IntoResponse {
    let stats = service.get_statistics();
    (StatusCode::OK, Json(stats)).into_response()
}
