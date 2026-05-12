use serde::{Serialize, Deserialize};
use chrono::{DateTime, Utc};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
pub enum RequestStatus {
    Initial,
    Retrying,
    Success,
    Failed,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RetryHistory {
    pub attempt: u32,
    pub timestamp: DateTime<Utc>,
    pub duration_ms: u64,
    pub success: bool,
    pub error_message: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RequestRecord {
    pub idempotency_key: String,
    pub request_id: Uuid,
    pub status: RequestStatus,
    pub created_at: DateTime<Utc>,
    pub completed_at: Option<DateTime<Utc>>,
    pub retry_history: Vec<RetryHistory>,
    pub total_retries: u32,
    pub response: Option<ResponseData>,
    pub expires_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ResponseData {
    pub status_code: u16,
    pub body: serde_json::Value,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RequestState {
    pub status: RequestStatus,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ApiRequest {
    pub idempotency_key: String,
    pub payload: serde_json::Value,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ApiResponse {
    pub request_id: Uuid,
    pub idempotency_key: String,
    pub status: RequestStatus,
    pub from_cache: bool,
    pub response: Option<ResponseData>,
    pub retry_count: u32,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct QueryResponse {
    pub request: RequestRecord,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct Statistics {
    pub total_requests: u64,
    pub total_with_retry: u64,
    pub success_requests: u64,
    pub retry_success_rate: f64,
    pub average_retries: f64,
}
