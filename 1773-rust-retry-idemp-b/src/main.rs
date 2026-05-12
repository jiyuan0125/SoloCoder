use std::{
    collections::HashMap,
    sync::{
        atomic::{AtomicU64, Ordering},
        Arc,
    },
    time::Duration,
};

use axum::{
    body::Body,
    extract::{Query, Request, State},
    http::StatusCode,
    middleware::{self, Next},
    response::{IntoResponse, Response},
    routing::{get, post},
    Json, Router,
};
use tokio::net::TcpListener;
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use thiserror::Error;
use tokio::sync::Mutex;

pub const IDEMPOTENCY_KEY_HEADER: &str = "x-idempotency-key";

#[derive(Debug, Clone, Copy, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum RetryStatus {
    Pending,
    Success,
    Failed,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RetryAttempt {
    pub attempt: u32,
    pub at: DateTime<Utc>,
    pub success: bool,
    pub response_status: Option<u16>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct IdempotentRecord {
    pub idempotency_key: String,
    pub request_id: String,
    pub method: String,
    pub path: String,
    pub created_at: DateTime<Utc>,
    pub status: RetryStatus,
    pub max_attempts: u32,
    pub attempts: Vec<RetryAttempt>,
    pub response_cache: Option<CachedResponse>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CachedResponse {
    pub status: u16,
    pub body: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RetryConfig {
    pub max_attempts: u32,
    pub initial_delay_ms: u64,
    pub backoff_multiplier: f64,
    pub jitter_ms: u64,
}

impl Default for RetryConfig {
    fn default() -> Self {
        Self {
            max_attempts: 3,
            initial_delay_ms: 1000,
            backoff_multiplier: 2.0,
            jitter_ms: 100,
        }
    }
}

impl RetryConfig {
    pub fn delay_for_attempt(&self, attempt: u32) -> Duration {
        let base = self.initial_delay_ms as f64 * self.backoff_multiplier.powi((attempt - 1) as i32);
        let jitter = if self.jitter_ms > 0 {
            (rand::random::<u64>() % self.jitter_ms) as f64
        } else {
            0.0
        };
        Duration::from_millis((base + jitter) as u64)
    }
}

#[derive(Debug, Error)]
pub enum RetryError {
    #[error("max attempts reached")]
    MaxAttemptsReached,
}

#[derive(Clone)]
pub struct AppState {
    pub records: Arc<Mutex<HashMap<String, IdempotentRecord>>>,
    pub retry_config: RetryConfig,
    request_counter: Arc<AtomicU64>,
}

impl AppState {
    pub fn new() -> Self {
        Self {
            records: Arc::new(Mutex::new(HashMap::new())),
            retry_config: RetryConfig::default(),
            request_counter: Arc::new(AtomicU64::new(1)),
        }
    }

    pub fn next_request_id(&self) -> String {
        let n = self.request_counter.fetch_add(1, Ordering::SeqCst);
        format!("req-{}", n)
    }
}

pub fn should_retry(status: StatusCode) -> bool {
    status.is_server_error()
}

async fn clone_request_body(body: Body) -> Option<Vec<u8>> {
    let bytes = axum::body::to_bytes(body, usize::MAX).await.ok()?;
    Some(bytes.to_vec())
}

pub async fn idempotent_middleware(
    State(state): State<AppState>,
    request: Request,
    next: Next,
) -> Response {
    let idempotency_key = match request
        .headers()
        .get(IDEMPOTENCY_KEY_HEADER)
        .and_then(|v| v.to_str().ok())
    {
        Some(k) if !k.trim().is_empty() => Some(k.trim().to_string()),
        _ => None,
    };

    let key = match idempotency_key {
        Some(k) => k,
        None => return next.run(request).await,
    };

    let method = request.method().as_str().to_string();
    let path = request.uri().path().to_string();

    {
        let records = state.records.lock().await;
        if let Some(record) = records.get(&key) {
            if let Some(cached) = &record.response_cache {
                tracing::info!("returning cached response for key: {}", key);
                return Response::builder()
                    .status(cached.status)
                    .body(Body::from(cached.body.clone()))
                    .unwrap();
            }
        }
    }

    let (parts, body) = request.into_parts();
    let body_bytes = clone_request_body(body).await.unwrap_or_default();

    let mut new_record = IdempotentRecord {
        idempotency_key: key.clone(),
        request_id: state.next_request_id(),
        method: method.clone(),
        path: path.clone(),
        created_at: Utc::now(),
        status: RetryStatus::Pending,
        max_attempts: state.retry_config.max_attempts,
        attempts: Vec::new(),
        response_cache: None,
    };

    {
        let mut records = state.records.lock().await;
        if records.contains_key(&key) {
            let record = records.get(&key).unwrap();
            if let Some(cached) = &record.response_cache {
                return Response::builder()
                    .status(cached.status)
                    .body(Body::from(cached.body.clone()))
                    .unwrap();
            }
        }
        records.insert(key.clone(), new_record.clone());
    }

    let retry_config = state.retry_config.clone();
    let mut last_response: Option<Response> = None;

    for attempt in 1..=retry_config.max_attempts {
        let attempt_at = Utc::now();
        tracing::info!("attempt {} of {}", attempt, retry_config.max_attempts);

        let cloned_body = Body::from(body_bytes.clone());
        let cloned_parts = parts.clone();
        let req_to_use = Request::from_parts(cloned_parts, cloned_body);

        let response = next.clone().run(req_to_use).await;
        let status = response.status();

        let response_body_bytes = axum::body::to_bytes(response.into_body(), usize::MAX)
            .await
            .unwrap_or_default();
        let response_str = String::from_utf8_lossy(&response_body_bytes).to_string();

        let cache_attempt = RetryAttempt {
            attempt,
            at: attempt_at,
            success: !should_retry(status),
            response_status: Some(status.as_u16()),
        };
        new_record.attempts.push(cache_attempt);

        let rebuild = Response::builder()
            .status(status)
            .body(Body::from(response_str.clone()))
            .unwrap();

        if !should_retry(status) {
            new_record.status = RetryStatus::Success;
            new_record.response_cache = Some(CachedResponse {
                status: status.as_u16(),
                body: response_str,
            });
            last_response = Some(rebuild);
            break;
        } else {
            last_response = Some(rebuild);
        }

        if attempt < retry_config.max_attempts {
            let delay = retry_config.delay_for_attempt(attempt);
            tracing::info!("retry after {}ms", delay.as_millis());
            tokio::time::sleep(delay).await;
        }
    }

    if new_record.attempts.iter().all(|a| !a.success) {
        new_record.status = RetryStatus::Failed;
    }

    {
        let mut records = state.records.lock().await;
        records.insert(key, new_record);
    }

    match last_response {
        Some(response) => response,
        None => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(serde_json::json!({
                "error": "retry_failed",
                "message": "max attempts reached"
            })),
        )
            .into_response(),
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HistoryQueryParams {
    pub key: Option<String>,
    pub status: Option<String>,
    pub limit: Option<usize>,
    pub offset: Option<usize>,
}

pub async fn get_retry_history(
    State(state): State<AppState>,
    query: Query<HistoryQueryParams>,
) -> impl IntoResponse {
    let records = state.records.lock().await;
    let mut results: Vec<IdempotentRecord> = records.values().cloned().collect();

    if let Some(key) = &query.key {
        results.retain(|r| r.idempotency_key.contains(key));
    }

    if let Some(status) = &query.status {
        let status_lower = status.to_lowercase();
        results.retain(|r| {
            let r_str = serde_json::to_string(&r.status).unwrap_or_default();
            r_str.contains(&status_lower)
        });
    }

    results.sort_by(|a, b| b.created_at.cmp(&a.created_at));

    let offset = query.offset.unwrap_or(0);
    let limit = query.limit.unwrap_or(100);
    let paginated: Vec<IdempotentRecord> = results.into_iter().skip(offset).take(limit).collect();

    Json(serde_json::json!({
        "total": records.len(),
        "count": paginated.len(),
        "offset": offset,
        "limit": limit,
        "records": paginated
    }))
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TestRequestBody {
    pub fail_count: Option<u32>,
    pub message: String,
}

static FAIL_COUNTERS: once_cell::sync::Lazy<tokio::sync::Mutex<HashMap<String, u32>>> =
    once_cell::sync::Lazy::new(|| tokio::sync::Mutex::new(HashMap::new()));

pub async fn test_retry_handler(State(_state): State<AppState>, Json(body): Json<TestRequestBody>) -> impl IntoResponse {
    let fail_count = body.fail_count.unwrap_or(0);

    let mut counters = FAIL_COUNTERS.lock().await;
    let request_id = body.message.clone();
    let counter = counters.entry(request_id).or_insert(0);

    *counter += 1;
    let current_count = *counter;

    drop(counters);

    if current_count <= fail_count {
        tracing::info!("simulating failure at attempt {}", current_count);
        return (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(serde_json::json!({
                "error": "simulated_server_error",
                "attempt": current_count,
                "message": body.message
            })),
        );
    }

    (
        StatusCode::OK,
        Json(serde_json::json!({
            "status": "success",
            "attempt": current_count,
            "message": body.message
        })),
    )
}

pub async fn reset_handler() -> impl IntoResponse {
    let mut counters = FAIL_COUNTERS.lock().await;
    counters.clear();
    Json(serde_json::json!({ "status": "ok", "message": "counters reset" }))
}

pub async fn health_handler() -> impl IntoResponse {
    Json(serde_json::json!({ "status": "ok" }))
}

pub fn create_router(state: AppState) -> Router {
    Router::new()
        .route("/health", get(health_handler))
        .route("/reset", post(reset_handler))
        .route("/history", get(get_retry_history))
        .route("/test-retry", post(test_retry_handler))
        .route_layer(middleware::from_fn_with_state(
            state.clone(),
            idempotent_middleware,
        ))
        .with_state(state)
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_max_level(tracing::Level::INFO)
        .init();

    let port: u16 = std::env::var("PORT")
        .ok()
        .and_then(|p| p.parse().ok())
        .unwrap_or(3000);

    let state = AppState::new();
    let app = create_router(state);

    let addr = std::net::SocketAddr::from(([0, 0, 0, 0], port));
    tracing::info!("server listening on {}", addr);

    let listener = TcpListener::bind(&addr).await.expect("failed to bind");
    axum::serve(listener, app)
        .await
        .expect("failed to start server");
}
