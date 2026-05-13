use actix_web::{web, App, HttpResponse, HttpServer, Responder, HttpRequest, Error};
use actix_web::http::StatusCode;
use chrono::{DateTime, Utc};
use dashmap::DashMap;
use lru::LruCache;
use serde::{Deserialize, Serialize};
use std::collections::VecDeque;
use std::num::NonZeroUsize;
use std::sync::Arc;
use std::time::Duration;
use tokio::sync::{Mutex, RwLock, oneshot};
use uuid::Uuid;

const IDEMPOTENCY_KEY_HEADER: &str = "Idempotency-Key";
const DEFAULT_CACHE_TTL: Duration = Duration::from_secs(300);
const DEFAULT_CACHE_CAPACITY: usize = 100_000;
const MAX_RETRIES: u32 = 3;
const INITIAL_BACKOFF: Duration = Duration::from_secs(1);

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Backend {
    pub id: Uuid,
    pub target_url: String,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RegisterBackendRequest {
    pub target_url: String,
}

#[derive(Debug, Clone, Serialize)]
pub struct StatsResponse {
    pub total_requests: u64,
    pub retry_count: u64,
    pub deduplication_hits: u64,
}

#[derive(Debug, Clone, Serialize)]
pub struct CachedResponse {
    pub status_code: u16,
    pub body: Vec<u8>,
    pub cached_at: DateTime<Utc>,
}

pub type PendingMap = DashMap<Uuid, Arc<Mutex<VecDeque<oneshot::Sender<Result<CachedResponse, String>>>>>>;

pub struct AppState {
    pub backends: RwLock<Vec<Backend>>,
    pub stats: RwLock<StatsResponse>,
    pub cache: Mutex<LruCache<Uuid, CachedResponse>>,
    pub pending: PendingMap,
    pub cache_ttl: Duration,
    pub http_client: reqwest::Client,
}

pub struct PendingRequest {
    pub idempotency_key: Uuid,
    pub method: String,
    pub uri: String,
    pub headers: Vec<(String, String)>,
    pub body: Vec<u8>,
    pub backend_url: String,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct PersistedPendingRequest {
    pub idempotency_key: Uuid,
    pub method: String,
    pub uri: String,
    pub headers: Vec<(String, String)>,
    pub body: Vec<u8>,
    pub backend_url: String,
    pub created_at: DateTime<Utc>,
}

fn is_retriable_status(status: StatusCode) -> bool {
    status.is_server_error() || status == StatusCode::TOO_MANY_REQUESTS
}

async fn should_retry(result: &Result<reqwest::Response, reqwest::Error>) -> bool {
    match result {
        Ok(response) => is_retriable_status(response.status()),
        Err(e) => e.is_timeout() || e.is_connect() || e.is_request(),
    }
}

async fn execute_with_retry(
    client: &reqwest::Client,
    method: reqwest::Method,
    url: &str,
    headers: reqwest::header::HeaderMap,
    body: Vec<u8>,
    max_retries: u32,
    initial_backoff: Duration,
    stats: &RwLock<StatsResponse>,
) -> Result<reqwest::Response, reqwest::Error> {
    let mut attempt = 0;
    let mut backoff = initial_backoff;

    loop {
        let req_builder = client
            .request(method.clone(), url)
            .headers(headers.clone())
            .body(body.clone());

        let result = req_builder.send().await;

        if attempt >= max_retries {
            return result;
        }

        if should_retry(&result).await {
            {
                let mut s = stats.write().await;
                s.retry_count += 1;
            }
            tokio::time::sleep(backoff).await;
            backoff = backoff.saturating_mul(2);
            attempt += 1;
        } else {
            return result;
        }
    }
}

async fn select_backend(backends: &RwLock<Vec<Backend>>) -> Option<Backend> {
    let backends_guard = backends.read().await;
    if backends_guard.is_empty() {
        None
    } else {
        use rand::seq::SliceRandom;
        let mut rng = rand::thread_rng();
        backends_guard.choose(&mut rng).cloned()
    }
}

async fn build_forward_url(backend: &Backend, req: &HttpRequest) -> String {
    let path = req.uri().path();
    let query = req.uri().query().map(|q| format!("?{}", q)).unwrap_or_default();
    let target = backend.target_url.trim_end_matches('/');
    format!("{}{}{}", target, path, query)
}

async fn build_headers(req: &HttpRequest, idempotency_key: Uuid) -> reqwest::header::HeaderMap {
    let mut headers = reqwest::header::HeaderMap::new();
    
    for (name, value) in req.headers() {
        if name.as_str().eq_ignore_ascii_case("host") {
            continue;
        }
        if name.as_str().eq_ignore_ascii_case(IDEMPOTENCY_KEY_HEADER) {
            continue;
        }
        if let Ok(val) = value.to_str() {
            if let Ok(header_name) = reqwest::header::HeaderName::from_bytes(name.as_str().as_bytes()) {
                if let Ok(header_value) = reqwest::header::HeaderValue::from_str(val) {
                    headers.insert(header_name, header_value);
                }
            }
        }
    }
    
    headers.insert(
        IDEMPOTENCY_KEY_HEADER,
        reqwest::header::HeaderValue::from_str(&idempotency_key.to_string()).unwrap(),
    );
    
    headers
}

fn parse_idempotency_key(req: &HttpRequest) -> Result<Uuid, String> {
    let header_value = req
        .headers()
        .get(IDEMPOTENCY_KEY_HEADER)
        .ok_or_else(|| format!("Missing {} header", IDEMPOTENCY_KEY_HEADER))?;
    
    let key_str = header_value
        .to_str()
        .map_err(|e| format!("Invalid header encoding: {}", e))?;
    
    Uuid::parse_str(key_str).map_err(|e| format!("Invalid UUID format: {}", e))
}

async fn cache_lookup(
    cache: &Mutex<LruCache<Uuid, CachedResponse>>,
    key: &Uuid,
    ttl: Duration,
) -> Option<CachedResponse> {
    let mut cache_guard = cache.lock().await;
    if let Some(cached) = cache_guard.get(key) {
        let elapsed = Utc::now().signed_duration_since(cached.cached_at);
        if elapsed.num_seconds() < ttl.as_secs() as i64 {
            return Some(cached.clone());
        } else {
            cache_guard.pop(key);
        }
    }
    None
}

async fn cache_insert(
    cache: &Mutex<LruCache<Uuid, CachedResponse>>,
    key: Uuid,
    response: CachedResponse,
) {
    let mut cache_guard = cache.lock().await;
    cache_guard.put(key, response);
}

async fn proxy_handler(
    req: HttpRequest,
    body: web::Bytes,
    state: web::Data<Arc<AppState>>,
) -> Result<HttpResponse, Error> {
    let idempotency_key = match parse_idempotency_key(&req) {
        Ok(k) => k,
        Err(e) => {
            return Ok(HttpResponse::BadRequest().body(e));
        }
    };

    {
        let mut s = state.stats.write().await;
        s.total_requests += 1;
    }

    if let Some(cached) = cache_lookup(&state.cache, &idempotency_key, state.cache_ttl).await {
        {
            let mut s = state.stats.write().await;
            s.deduplication_hits += 1;
        }
        let status = StatusCode::from_u16(cached.status_code).unwrap_or(StatusCode::INTERNAL_SERVER_ERROR);
        return Ok(HttpResponse::build(status).body(cached.body));
    }

    let (tx, rx) = oneshot::channel();
    
    let waiters = state.pending.entry(idempotency_key).or_insert_with(|| Arc::new(Mutex::new(VecDeque::new())));
    let is_first = {
        let mut w = waiters.value().lock().await;
        let was_empty = w.is_empty();
        w.push_back(tx);
        was_empty
    };

    if !is_first {
        drop(waiters);
        match rx.await {
            Ok(Ok(cached)) => {
                let status = StatusCode::from_u16(cached.status_code).unwrap_or(StatusCode::INTERNAL_SERVER_ERROR);
                return Ok(HttpResponse::build(status).body(cached.body));
            }
            Ok(Err(e)) => {
                return Ok(HttpResponse::BadGateway().body(e));
            }
            Err(_) => {
                return Ok(HttpResponse::InternalServerError().body("Request cancelled"));
            }
        }
    }

    let method = match req.method().as_str() {
        "GET" => reqwest::Method::GET,
        "POST" => reqwest::Method::POST,
        "PUT" => reqwest::Method::PUT,
        "DELETE" => reqwest::Method::DELETE,
        "PATCH" => reqwest::Method::PATCH,
        "HEAD" => reqwest::Method::HEAD,
        "OPTIONS" => reqwest::Method::OPTIONS,
        _ => {
            notify_waiters(&state.pending, &idempotency_key, Err("Unsupported method".to_string()));
            return Ok(HttpResponse::MethodNotAllowed().body("Method not allowed"));
        }
    };

    let backend = match select_backend(&state.backends).await {
        Some(b) => b,
        None => {
            notify_waiters(&state.pending, &idempotency_key, Err("No backends available".to_string()));
            return Ok(HttpResponse::ServiceUnavailable().body("No backends registered"));
        }
    };

    let url = build_forward_url(&backend, &req).await;
    let headers = build_headers(&req, idempotency_key).await;
    let body_vec: Vec<u8> = body.to_vec();

    let result = execute_with_retry(
        &state.http_client,
        method,
        &url,
        headers,
        body_vec,
        MAX_RETRIES,
        INITIAL_BACKOFF,
        &state.stats,
    ).await;

    let response_result = match result {
        Ok(resp) => {
            let status = resp.status();
            let body_bytes = match resp.bytes().await {
                Ok(b) => b.to_vec(),
                Err(e) => {
                    notify_waiters(&state.pending, &idempotency_key, Err(format!("Failed to read response: {}", e)));
                    return Ok(HttpResponse::BadGateway().body(format!("Failed to read response: {}", e)));
                }
            };

            if status.is_success() {
                let cached = CachedResponse {
                    status_code: status.as_u16(),
                    body: body_bytes.clone(),
                    cached_at: Utc::now(),
                };
                cache_insert(&state.cache, idempotency_key, cached.clone()).await;
                notify_waiters(&state.pending, &idempotency_key, Ok(cached));
            } else {
                notify_waiters(&state.pending, &idempotency_key, Err(format!("Backend error: {}", status)));
            }

            Ok(HttpResponse::build(status).body(body_bytes))
        }
        Err(e) => {
            notify_waiters(&state.pending, &idempotency_key, Err(format!("Request failed: {}", e)));
            Ok(HttpResponse::BadGateway().body(format!("Request failed: {}", e)))
        }
    };

    state.pending.remove(&idempotency_key);
    response_result
}

fn notify_waiters(
    pending: &PendingMap,
    key: &Uuid,
    result: Result<CachedResponse, String>,
) {
    if let Some((_, waiters)) = pending.remove(key) {
        tokio::spawn(async move {
            let mut w = waiters.lock().await;
            while let Some(tx) = w.pop_front() {
                let _ = tx.send(result.clone());
            }
        });
    }
}

async fn register_backend(
    req: web::Json<RegisterBackendRequest>,
    state: web::Data<Arc<AppState>>,
) -> impl Responder {
    let backend = Backend {
        id: Uuid::new_v4(),
        target_url: req.target_url.clone(),
        created_at: Utc::now(),
    };
    
    {
        let mut backends = state.backends.write().await;
        backends.push(backend.clone());
    }
    
    HttpResponse::Created().json(backend)
}

async fn list_backends(state: web::Data<Arc<AppState>>) -> impl Responder {
    let backends = state.backends.read().await.clone();
    HttpResponse::Ok().json(backends)
}

async fn get_stats(state: web::Data<Arc<AppState>>) -> impl Responder {
    let stats = state.stats.read().await.clone();
    HttpResponse::Ok().json(stats)
}

async fn health_check() -> impl Responder {
    HttpResponse::Ok().body("OK")
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    tracing_subscriber::fmt::init();

    let port = std::env::var("PORT")
        .unwrap_or_else(|_| "9101".to_string())
        .parse::<u16>()
        .expect("PORT must be a valid port number");

    let cache_capacity = NonZeroUsize::new(DEFAULT_CACHE_CAPACITY).unwrap();
    
    let app_state = Arc::new(AppState {
        backends: RwLock::new(Vec::new()),
        stats: RwLock::new(StatsResponse {
            total_requests: 0,
            retry_count: 0,
            deduplication_hits: 0,
        }),
        cache: Mutex::new(LruCache::new(cache_capacity)),
        pending: DashMap::new(),
        cache_ttl: DEFAULT_CACHE_TTL,
        http_client: reqwest::Client::builder()
            .timeout(Duration::from_secs(30))
            .build()
            .expect("Failed to create HTTP client"),
    });

    tracing::info!("Starting retry-idempotent-proxy on port {}", port);

    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(app_state.clone()))
            .route("/health", web::get().to(health_check))
            .route("/backends", web::post().to(register_backend))
            .route("/backends", web::get().to(list_backends))
            .route("/stats", web::get().to(get_stats))
            .default_service(web::route().to(proxy_handler))
    })
    .bind(("0.0.0.0", port))?
    .run()
    .await
}
