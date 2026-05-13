use actix_web::{web, App, HttpResponse, HttpServer, Responder, HttpRequest, Error};
use actix_web::http::StatusCode;
use chrono::{DateTime, Utc};
use dashmap::DashMap;
use futures::future::Shared;
use futures::FutureExt;
use lru::LruCache;
use serde::{Deserialize, Serialize};

use std::num::NonZeroUsize;
use std::sync::Arc;
use std::time::Duration;
use tokio::sync::{Mutex, RwLock, oneshot};
use tokio::task::JoinHandle;
use uuid::Uuid;

const IDEMPOTENCY_KEY_HEADER: &str = "Idempotency-Key";
const DEFAULT_CACHE_TTL: Duration = Duration::from_secs(300);
const DEFAULT_CACHE_CAPACITY: usize = 100_000;
const MAX_RETRIES: u32 = 3;
const INITIAL_BACKOFF: Duration = Duration::from_secs(1);
const PERSISTENCE_FILE: &str = "pending_requests.json";

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

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PersistedRequest {
    pub idempotency_key: Uuid,
    pub method: String,
    pub path: String,
    pub query: Option<String>,
    pub headers: Vec<(String, String)>,
    pub body: Vec<u8>,
    pub created_at: DateTime<Utc>,
}

pub type SharedResult = Shared<oneshot::Receiver<Result<CachedResponse, String>>>;

pub struct InFlightEntry {
    pub result: SharedResult,
    pub handle: Option<JoinHandle<()>>,
}

pub struct AppState {
    pub backends: RwLock<Vec<Backend>>,
    pub stats: RwLock<StatsResponse>,
    pub cache: Mutex<LruCache<Uuid, CachedResponse>>,
    pub in_flight: DashMap<Uuid, Arc<InFlightEntry>>,
    pub cache_ttl: Duration,
    pub http_client: reqwest::Client,
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

fn build_forward_url(backend: &Backend, path: &str, query: &Option<String>) -> String {
    let query_str = query.as_ref().map(|q| format!("?{}", q)).unwrap_or_default();
    let target = backend.target_url.trim_end_matches('/');
    format!("{}{}{}", target, path, query_str)
}

fn build_headers_from_vec(
    headers_vec: &[(String, String)],
    idempotency_key: Uuid,
) -> reqwest::header::HeaderMap {
    let mut headers = reqwest::header::HeaderMap::new();
    
    for (name, value) in headers_vec {
        if name.eq_ignore_ascii_case("host") {
            continue;
        }
        if name.eq_ignore_ascii_case(IDEMPOTENCY_KEY_HEADER) {
            continue;
        }
        if let Ok(header_name) = reqwest::header::HeaderName::from_bytes(name.as_bytes()) {
            if let Ok(header_value) = reqwest::header::HeaderValue::from_str(value) {
                headers.insert(header_name, header_value);
            }
        }
    }
    
    if let Ok(header_value) = reqwest::header::HeaderValue::from_str(&idempotency_key.to_string()) {
        headers.insert(IDEMPOTENCY_KEY_HEADER, header_value);
    }
    
    headers
}

fn build_headers_vec(req: &HttpRequest) -> Vec<(String, String)> {
    let mut headers_vec = Vec::new();
    for (name, value) in req.headers() {
        if let Ok(val) = value.to_str() {
            headers_vec.push((name.as_str().to_string(), val.to_string()));
        }
    }
    headers_vec
}

fn parse_method(method_str: &str) -> Result<reqwest::Method, String> {
    match method_str {
        "GET" => Ok(reqwest::Method::GET),
        "POST" => Ok(reqwest::Method::POST),
        "PUT" => Ok(reqwest::Method::PUT),
        "DELETE" => Ok(reqwest::Method::DELETE),
        "PATCH" => Ok(reqwest::Method::PATCH),
        "HEAD" => Ok(reqwest::Method::HEAD),
        "OPTIONS" => Ok(reqwest::Method::OPTIONS),
        _ => Err(format!("Unsupported method: {}", method_str)),
    }
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

fn load_all_persisted() -> Vec<PersistedRequest> {
    match std::fs::read_to_string(PERSISTENCE_FILE) {
        Ok(content) => serde_json::from_str(&content).unwrap_or_default(),
        Err(_) => Vec::new(),
    }
}

fn save_all_persisted(requests: &[PersistedRequest]) {
    if let Ok(json) = serde_json::to_string(requests) {
        let _ = std::fs::write(PERSISTENCE_FILE, json);
    }
}

fn persist_request(req: &PersistedRequest) {
    let mut all = load_all_persisted();
    all.push(req.clone());
    save_all_persisted(&all);
}

fn remove_persisted(key: &Uuid) {
    let all = load_all_persisted();
    let filtered: Vec<_> = all.into_iter().filter(|r| r.idempotency_key != *key).collect();
    save_all_persisted(&filtered);
}

async fn execute_request_and_broadcast(
    state: Arc<AppState>,
    idempotency_key: Uuid,
    method: String,
    path: String,
    query: Option<String>,
    headers_vec: Vec<(String, String)>,
    body: Vec<u8>,
    tx: oneshot::Sender<Result<CachedResponse, String>>,
) {
    let persisted = PersistedRequest {
        idempotency_key,
        method: method.clone(),
        path: path.clone(),
        query: query.clone(),
        headers: headers_vec.clone(),
        body: body.clone(),
        created_at: Utc::now(),
    };
    persist_request(&persisted);

    let method = match parse_method(&method) {
        Ok(m) => m,
        Err(e) => {
            let _ = tx.send(Err(e));
            remove_persisted(&idempotency_key);
            state.in_flight.remove(&idempotency_key);
            return;
        }
    };

    let backend = match select_backend(&state.backends).await {
        Some(b) => b,
        None => {
            let _ = tx.send(Err("No backends available".to_string()));
            remove_persisted(&idempotency_key);
            state.in_flight.remove(&idempotency_key);
            return;
        }
    };

    let url = build_forward_url(&backend, &path, &query);
    let headers = build_headers_from_vec(&headers_vec, idempotency_key);

    let result = execute_with_retry(
        &state.http_client,
        method,
        &url,
        headers,
        body.clone(),
        MAX_RETRIES,
        INITIAL_BACKOFF,
        &state.stats,
    ).await;

    let final_result = match result {
        Ok(resp) => {
            let status = resp.status();
            let body_bytes = match resp.bytes().await {
                Ok(b) => b.to_vec(),
                Err(e) => {
                    let _ = tx.send(Err(format!("Failed to read response: {}", e)));
                    remove_persisted(&idempotency_key);
                    state.in_flight.remove(&idempotency_key);
                    return;
                }
            };

            if status.is_success() {
                let cached = CachedResponse {
                    status_code: status.as_u16(),
                    body: body_bytes,
                    cached_at: Utc::now(),
                };
                cache_insert(&state.cache, idempotency_key, cached.clone()).await;
                Ok(cached)
            } else {
                Err(format!("Backend error: {}", status))
            }
        }
        Err(e) => Err(format!("Request failed: {}", e)),
    };

    remove_persisted(&idempotency_key);
    let _ = tx.send(final_result);
    state.in_flight.remove(&idempotency_key);
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

    let (entry, is_new) = {
        if let Some(existing) = state.in_flight.get(&idempotency_key) {
            (existing.clone(), false)
        } else {
            let (tx, rx) = oneshot::channel();
            let shared_rx = rx.shared();
            
            let method = req.method().as_str().to_string();
            let path = req.uri().path().to_string();
            let query = req.uri().query().map(|s| s.to_string());
            let headers_vec = build_headers_vec(&req);
            let body_vec: Vec<u8> = body.to_vec();
            
            let state_clone = state.get_ref().clone();
            let key_clone = idempotency_key;
            
            let handle = tokio::spawn(async move {
                execute_request_and_broadcast(
                    state_clone,
                    key_clone,
                    method,
                    path,
                    query,
                    headers_vec,
                    body_vec,
                    tx,
                ).await;
            });
            
            let in_flight_entry = Arc::new(InFlightEntry {
                result: shared_rx,
                handle: Some(handle),
            });
            
            state.in_flight.insert(idempotency_key, in_flight_entry.clone());
            (in_flight_entry, true)
        }
    };

    let result_future = entry.result.clone();
    let result = result_future.await;

    match result {
        Ok(Ok(cached)) => {
            if !is_new {
                let mut s = state.stats.write().await;
                s.deduplication_hits += 1;
            }
            let status = StatusCode::from_u16(cached.status_code).unwrap_or(StatusCode::INTERNAL_SERVER_ERROR);
            Ok(HttpResponse::build(status).body(cached.body))
        }
        Ok(Err(e)) => {
            Ok(HttpResponse::BadGateway().body(e))
        }
        Err(_) => {
            Ok(HttpResponse::InternalServerError().body("Request cancelled"))
        }
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

async fn recover_pending_requests(state: Arc<AppState>) {
    let pending = load_all_persisted();
    if !pending.is_empty() {
        tracing::info!("Recovering {} pending requests from disk", pending.len());
        for req in pending {
            let (tx, rx) = oneshot::channel();
            let shared_rx = rx.shared();
            
            let in_flight_entry = Arc::new(InFlightEntry {
                result: shared_rx,
                handle: None,
            });
            state.in_flight.insert(req.idempotency_key, in_flight_entry);
            
            let state_clone = state.clone();
            tokio::spawn(async move {
                execute_request_and_broadcast(
                    state_clone,
                    req.idempotency_key,
                    req.method,
                    req.path,
                    req.query,
                    req.headers,
                    req.body,
                    tx,
                ).await;
            });
        }
    }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    tracing_subscriber::fmt::init();

    let port = std::env::var("PORT")
        .unwrap_or_else(|_| "8080".to_string())
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
        in_flight: DashMap::new(),
        cache_ttl: DEFAULT_CACHE_TTL,
        http_client: reqwest::Client::builder()
            .timeout(Duration::from_secs(30))
            .build()
            .expect("Failed to create HTTP client"),
    });

    recover_pending_requests(app_state.clone()).await;

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
