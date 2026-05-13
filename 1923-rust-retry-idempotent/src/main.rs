use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, Instant};

use axum::body::Body;
use axum::extract::{Path, Query, State};
use axum::http::{HeaderMap, HeaderName, HeaderValue, StatusCode};
use axum::response::{IntoResponse, Response};
use axum::routing::{delete, get, post, put};
use axum::{Json, Router};
use http::Request;
use reqwest::Client;
use serde::{Deserialize, Serialize};
use tokio::sync::{Mutex, RwLock};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
struct RetryConfig {
    #[serde(default = "default_max_retries")]
    max_retries: u32,
    #[serde(default)]
    backoff: BackoffStrategy,
    #[serde(default = "default_total_timeout")]
    total_timeout: u64,
}

fn default_max_retries() -> u32 {
    3
}

fn default_total_timeout() -> u64 {
    30
}

impl Default for RetryConfig {
    fn default() -> Self {
        Self {
            max_retries: 3,
            backoff: BackoffStrategy::default(),
            total_timeout: 30,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type", rename_all = "lowercase")]
enum BackoffStrategy {
    Fixed { interval_ms: u64 },
    Exponential { initial_ms: u64, max_ms: u64 },
}

impl Default for BackoffStrategy {
    fn default() -> Self {
        BackoffStrategy::Fixed { interval_ms: 1000 }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
struct BackendStats {
    total_requests: u64,
    retries_by_status: HashMap<u16, u64>,
    total_retries: u64,
    success_count: u64,
    total_response_time_ms: u64,
}

impl BackendStats {
    fn new() -> Self {
        Self::default()
    }

    fn record_request(&mut self, success: bool, response_time_ms: u64) {
        self.total_requests += 1;
        if success {
            self.success_count += 1;
        }
        self.total_response_time_ms += response_time_ms;
    }

    fn record_retry(&mut self, status_code: u16) {
        self.total_retries += 1;
        *self.retries_by_status.entry(status_code).or_insert(0) += 1;
    }

    fn success_rate(&self) -> f64 {
        if self.total_requests == 0 {
            0.0
        } else {
            self.success_count as f64 / self.total_requests as f64
        }
    }

    fn avg_response_time_ms(&self) -> f64 {
        if self.total_requests == 0 {
            0.0
        } else {
            self.total_response_time_ms as f64 / self.total_requests as f64
        }
    }
}

#[derive(Debug, Clone)]
struct BackendInternal {
    id: Uuid,
    target_url: String,
    retry_config: RetryConfig,
    stats: BackendStats,
}

impl BackendInternal {
    fn new(id: Uuid, target_url: String) -> Self {
        Self {
            id,
            target_url,
            retry_config: RetryConfig::default(),
            stats: BackendStats::new(),
        }
    }
}

#[derive(Debug, Clone, Serialize)]
struct StatsResponse {
    backend_id: Uuid,
    target_url: String,
    total_requests: u64,
    retries_by_status: HashMap<u16, u64>,
    total_retries: u64,
    success_rate: f64,
    avg_response_time_ms: f64,
}

struct IdempotencyEntry {
    response: Option<CachedResponse>,
    in_flight: Option<tokio::sync::oneshot::Receiver<Result<CachedResponse, String>>>,
    created_at: Instant,
}

#[derive(Debug, Clone)]
struct CachedResponse {
    status: u16,
    headers: Vec<(String, String)>,
    body: Vec<u8>,
}

struct IdempotencyCache {
    entries: HashMap<String, IdempotencyEntry>,
    capacity: usize,
    order: std::collections::VecDeque<String>,
}

impl IdempotencyCache {
    fn new(capacity: usize) -> Self {
        Self {
            entries: HashMap::new(),
            capacity,
            order: std::collections::VecDeque::new(),
        }
    }

    fn get(&mut self, key: &str) -> Option<&mut IdempotencyEntry> {
        self.entries.get_mut(key)
    }

    fn insert(&mut self, key: String, entry: IdempotencyEntry) {
        if self.entries.contains_key(&key) {
            self.order.retain(|k| k != &key);
        } else if self.entries.len() >= self.capacity {
            if let Some(oldest) = self.order.pop_front() {
                self.entries.remove(&oldest);
            }
        }
        self.order.push_back(key.clone());
        self.entries.insert(key, entry);
    }

    fn cleanup_expired(&mut self, ttl: Duration) {
        let now = Instant::now();
        let expired: Vec<String> = self
            .entries
            .iter()
            .filter(|(_, entry)| now.duration_since(entry.created_at) > ttl)
            .map(|(k, _)| k.clone())
            .collect();
        for key in expired {
            self.entries.remove(&key);
            self.order.retain(|k| k != &key);
        }
    }
}

struct AppState {
    backends: RwLock<HashMap<Uuid, BackendInternal>>,
    idempotency_cache: Mutex<IdempotencyCache>,
    http_client: Client,
    idempotency_capacity: usize,
    idempotency_ttl: Duration,
}

#[derive(Debug, Deserialize)]
struct CreateBackendRequest {
    target_url: String,
}

#[derive(Debug, Serialize)]
struct CreateBackendResponse {
    id: Uuid,
}

#[derive(Debug, Deserialize)]
struct UpdateRetryConfigRequest {
    max_retries: Option<u32>,
    backoff: Option<BackoffStrategy>,
    total_timeout: Option<u64>,
}

#[derive(Debug, Deserialize)]
struct ProxyQuery {
    backend_id: Uuid,
    #[serde(default)]
    timeout: Option<u64>,
}

async fn create_backend(
    State(state): State<Arc<AppState>>,
    Json(payload): Json<CreateBackendRequest>,
) -> impl IntoResponse {
    let id = Uuid::new_v4();
    let backend = BackendInternal::new(id, payload.target_url);

    state.backends.write().await.insert(id, backend);

    (StatusCode::CREATED, Json(CreateBackendResponse { id }))
}

async fn delete_backend(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> StatusCode {
    if state.backends.write().await.remove(&id).is_some() {
        StatusCode::NO_CONTENT
    } else {
        StatusCode::NOT_FOUND
    }
}

async fn update_retry_config(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
    Json(payload): Json<UpdateRetryConfigRequest>,
) -> StatusCode {
    let mut backends = state.backends.write().await;
    if let Some(backend) = backends.get_mut(&id) {
        if let Some(max_retries) = payload.max_retries {
            backend.retry_config.max_retries = max_retries;
        }
        if let Some(backoff) = payload.backoff {
            backend.retry_config.backoff = backoff;
        }
        if let Some(total_timeout) = payload.total_timeout {
            backend.retry_config.total_timeout = total_timeout;
        }
        StatusCode::OK
    } else {
        StatusCode::NOT_FOUND
    }
}

async fn get_stats(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let backends = state.backends.read().await;
    let response: Vec<StatsResponse> = backends
        .values()
        .map(|b| StatsResponse {
            backend_id: b.id,
            target_url: b.target_url.clone(),
            total_requests: b.stats.total_requests,
            retries_by_status: b.stats.retries_by_status.clone(),
            total_retries: b.stats.total_retries,
            success_rate: b.stats.success_rate(),
            avg_response_time_ms: b.stats.avg_response_time_ms(),
        })
        .collect();
    Json(response)
}

fn should_retry(status: Option<reqwest::StatusCode>) -> bool {
    match status {
        Some(s) => s.is_server_error(),
        None => true,
    }
}

enum ProxyResult {
    Success {
        response: reqwest::Response,
        retries_recorded: Vec<u16>,
    },
    TimedOut,
    Error(reqwest::Error),
}

async fn execute_proxy_attempt(
    client: &Client,
    method: reqwest::Method,
    url: &str,
    headers: HeaderMap,
    body: Vec<u8>,
    timeout: Duration,
) -> Result<reqwest::Response, reqwest::Error> {
    let mut req = client
        .request(method.clone(), url)
        .timeout(timeout);

    for (name, value) in headers.iter() {
        if name.as_str().eq_ignore_ascii_case("host") {
            continue;
        }
        req = req.header(name, value);
    }

    req = req.body(body.clone());

    req.send().await
}

async fn proxy_handler(
    State(state): State<Arc<AppState>>,
    Query(query): Query<ProxyQuery>,
    req: Request<Body>,
) -> Result<Response, StatusCode> {
    let (parts, body) = req.into_parts();
    let backend_id = query.backend_id;

    let (target_url, config) = {
        let backends = state.backends.read().await;
        let backend = backends.get(&backend_id).ok_or(StatusCode::NOT_FOUND)?;
        (backend.target_url.clone(), backend.retry_config.clone())
    };

    let idempotency_key = parts
        .headers
        .get("idempotency-key")
        .and_then(|v| v.to_str().ok())
        .map(|s| s.to_string());

    let body_bytes = axum::body::to_bytes(body, usize::MAX)
        .await
        .map_err(|_| StatusCode::BAD_REQUEST)?;
    let body_vec = body_bytes.to_vec();

    if let Some(key) = idempotency_key {
        if !key.is_empty() {
            let cache_key = format!("{}:{}", backend_id, key);
            let mut cache = state.idempotency_cache.lock().await;

            cache.cleanup_expired(state.idempotency_ttl);

            if let Some(entry) = cache.get(&cache_key) {
                if let Some(ref response) = &entry.response {
                    let mut headers = HeaderMap::new();
                    for (name, value) in &response.headers {
                        if let (Ok(n), Ok(v)) = (
                            HeaderName::from_bytes(name.as_bytes()),
                            HeaderValue::from_str(value),
                        ) {
                            headers.insert(n, v);
                        }
                    }
                    return Ok(Response::builder()
                        .status(StatusCode::from_u16(response.status).unwrap())
                        .body(Body::from(response.body.clone()))
                        .unwrap());
                } else if let Some(rx) = entry.in_flight.take() {
                    drop(cache);
                    let result = rx.await.map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;
                    let cached = result.map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;
                    let mut headers = HeaderMap::new();
                    for (name, value) in &cached.headers {
                        if let (Ok(n), Ok(v)) = (
                            HeaderName::from_bytes(name.as_bytes()),
                            HeaderValue::from_str(value),
                        ) {
                            headers.insert(n, v);
                        }
                    }
                    return Ok(Response::builder()
                        .status(StatusCode::from_u16(cached.status).unwrap())
                        .body(Body::from(cached.body.clone()))
                        .unwrap());
                }
            }

            let (tx, rx) = tokio::sync::oneshot::channel();
            cache.insert(
                cache_key.clone(),
                IdempotencyEntry {
                    response: None,
                    in_flight: Some(rx),
                    created_at: Instant::now(),
                },
            );
            drop(cache);

            let result = do_proxy_request(
                &state,
                backend_id,
                &target_url,
                &config,
                &parts.method,
                &parts.headers,
                body_vec.clone(),
                query.timeout,
            )
            .await;

            let mut cache = state.idempotency_cache.lock().await;
            match &result {
                Ok((status, headers, body_vec, _)) => {
                    let cached_headers: Vec<(String, String)> = headers
                        .iter()
                        .map(|(n, v)| {
                            (
                                n.as_str().to_string(),
                                v.to_str().unwrap_or("").to_string(),
                            )
                        })
                        .collect();
                    let cached = CachedResponse {
                        status: status.as_u16(),
                        headers: cached_headers,
                        body: body_vec.clone(),
                    };
                    cache.insert(
                        cache_key.clone(),
                        IdempotencyEntry {
                            response: Some(cached.clone()),
                            in_flight: None,
                            created_at: Instant::now(),
                        },
                    );
                    let _ = tx.send(Ok(cached));
                    drop(cache);
                }
                Err(_) => {
                    cache.entries.remove(&cache_key);
                    let _ = tx.send(Err("proxy error".to_string()));
                    drop(cache);
                }
            }

            return match result {
                Ok((status, headers, body_vec, _)) => {
                    let mut response_headers = HeaderMap::new();
                    for (name, value) in headers {
                        response_headers.insert(name, value);
                    }
                    Ok(Response::builder()
                        .status(status)
                        .body(Body::from(body_vec))
                        .unwrap())
                }
                Err(e) => Err(e),
            };
        }
    }

    let (status, headers, body_vec, _) = do_proxy_request(
        &state,
        backend_id,
        &target_url,
        &config,
        &parts.method,
        &parts.headers,
        body_vec,
        query.timeout,
    )
    .await?;
    let mut response_headers = HeaderMap::new();
    for (name, value) in headers {
        response_headers.insert(name, value);
    }
    return Ok(Response::builder()
        .status(status)
        .body(Body::from(body_vec))
        .unwrap());
}

async fn do_proxy_request(
    state: &Arc<AppState>,
    backend_id: Uuid,
    target_url: &str,
    config: &RetryConfig,
    method: &axum::http::Method,
    headers: &HeaderMap,
    body: Vec<u8>,
    request_timeout: Option<u64>,
) -> Result<
    (
        StatusCode,
        Vec<(HeaderName, HeaderValue)>,
        Vec<u8>,
        bool,
    ),
    StatusCode,
> {
    let reqwest_method = match method.as_str() {
        "GET" => reqwest::Method::GET,
        "POST" => reqwest::Method::POST,
        "PUT" => reqwest::Method::PUT,
        "DELETE" => reqwest::Method::DELETE,
        "PATCH" => reqwest::Method::PATCH,
        "HEAD" => reqwest::Method::HEAD,
        "OPTIONS" => reqwest::Method::OPTIONS,
        _ => reqwest::Method::GET,
    };

    let start = Instant::now();
    let request_timeout_dur = request_timeout.map(Duration::from_secs);

    let result = tokio::time::timeout(
        Duration::from_secs(config.total_timeout),
        run_with_retries(
            &state.http_client,
            reqwest_method,
            target_url,
            headers.clone(),
            body,
            config,
            request_timeout_dur,
        ),
    )
    .await;

    let result = match result {
        Ok(r) => r,
        Err(_) => ProxyResult::TimedOut,
    };

    let elapsed_ms = start.elapsed().as_millis() as u64;

    match result {
        ProxyResult::Success { response, retries_recorded } => {
            let status_axum = StatusCode::from_u16(response.status().as_u16())
                .unwrap_or(StatusCode::INTERNAL_SERVER_ERROR);
            let success = response.status().is_success();
            let resp_headers: Vec<(HeaderName, HeaderValue)> = response
                .headers()
                .iter()
                .map(|(n, v)| (n.clone(), v.clone()))
                .collect();
            let body_vec = response
                .bytes()
                .await
                .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?
                .to_vec();

            {
                let mut backends = state.backends.write().await;
                if let Some(b) = backends.get_mut(&backend_id) {
                    for status in retries_recorded {
                        b.stats.record_retry(status);
                    }
                    b.stats.record_request(success, elapsed_ms);
                }
            }

            Ok((status_axum, resp_headers, body_vec, success))
        }
        ProxyResult::TimedOut | ProxyResult::Error(_) => {
            {
                let mut backends = state.backends.write().await;
                if let Some(b) = backends.get_mut(&backend_id) {
                    b.stats.record_request(false, elapsed_ms);
                }
            }
            Err(StatusCode::BAD_GATEWAY)
        }
    }
}

async fn run_with_retries(
    client: &Client,
    method: reqwest::Method,
    url: &str,
    headers: HeaderMap,
    body: Vec<u8>,
    config: &RetryConfig,
    request_timeout: Option<Duration>,
) -> ProxyResult {
    let mut attempt = 0;
    let default_timeout = Duration::from_secs(30);
    let request_timeout = request_timeout.unwrap_or(default_timeout);
    let mut retries_recorded = Vec::new();

    loop {
        let result = execute_proxy_attempt(
            client,
            method.clone(),
            url,
            headers.clone(),
            body.clone(),
            request_timeout,
        )
        .await;

        match result {
            Ok(response) => {
                let status = response.status();
                if !should_retry(Some(status)) {
                    return ProxyResult::Success { response, retries_recorded };
                }

                if attempt >= config.max_retries {
                    return ProxyResult::Success { response, retries_recorded };
                }

                retries_recorded.push(status.as_u16());
                attempt += 1;

                match &config.backoff {
                    BackoffStrategy::Fixed { interval_ms } => {
                        tokio::time::sleep(Duration::from_millis(*interval_ms)).await;
                    }
                    BackoffStrategy::Exponential { initial_ms, max_ms } => {
                        let delay = (*initial_ms as u64) * (2u64.pow(attempt - 1));
                        let delay = delay.min(*max_ms);
                        tokio::time::sleep(Duration::from_millis(delay)).await;
                    }
                }
            }
            Err(e) => {
                if attempt >= config.max_retries {
                    return ProxyResult::Error(e);
                }

                attempt += 1;

                match &config.backoff {
                    BackoffStrategy::Fixed { interval_ms } => {
                        tokio::time::sleep(Duration::from_millis(*interval_ms)).await;
                    }
                    BackoffStrategy::Exponential { initial_ms, max_ms } => {
                        let delay = (*initial_ms as u64) * (2u64.pow(attempt - 1));
                        let delay = delay.min(*max_ms);
                        tokio::time::sleep(Duration::from_millis(delay)).await;
                    }
                }
            }
        }
    }
}

#[tokio::main]
async fn main() {
    let port = std::env::var("PORT").unwrap_or_else(|_| "3000".to_string());
    let idempotency_capacity = std::env::var("IDEMPOTENCY_CAPACITY")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(100000);
    let idempotency_ttl = std::env::var("IDEMPOTENCY_TTL_SECONDS")
        .ok()
        .and_then(|s| s.parse::<u64>().ok())
        .map(Duration::from_secs)
        .unwrap_or(Duration::from_secs(300));

    let state = Arc::new(AppState {
        backends: RwLock::new(HashMap::new()),
        idempotency_cache: Mutex::new(IdempotencyCache::new(idempotency_capacity)),
        http_client: Client::builder()
            .timeout(Duration::from_secs(30))
            .build()
            .unwrap(),
        idempotency_capacity,
        idempotency_ttl,
    });

    let app = Router::new()
        .route("/backends", post(create_backend))
        .route("/backends/:id", delete(delete_backend))
        .route("/backends/:id/retry-config", put(update_retry_config))
        .route("/stats", get(get_stats))
        .route(
            "/proxy",
            get(proxy_handler)
                .post(proxy_handler)
                .put(proxy_handler)
                .delete(proxy_handler)
                .patch(proxy_handler),
        )
        .with_state(state);

    let addr = format!("0.0.0.0:{}", port);
    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    println!("Server listening on {}", addr);
    axum::serve(listener, app).await.unwrap();
}
