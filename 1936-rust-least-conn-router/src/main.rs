use axum::{
    body::Body,
    extract::{Path, State},
    http::{Request, StatusCode},
    response::{IntoResponse, Response},
    routing::{delete, get, post, put},
    Json, Router,
};
use dashmap::DashMap;
use rand::seq::SliceRandom;
use rand::Rng;
use serde::{Deserialize, Serialize};
use std::collections::hash_map::DefaultHasher;
use std::hash::{Hash, Hasher};
use std::net::SocketAddr;
use std::sync::Arc;
use std::time::Duration;
use tokio::sync::RwLock;
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Backend {
    pub id: Uuid,
    pub url: String,
    pub active_connections: u32,
    pub total_requests: u64,
    pub failure_count: u32,
    pub consecutive_failures: u32,
    pub healthy: bool,
}

impl Backend {
    pub fn new(url: String) -> Self {
        Self {
            id: Uuid::new_v4(),
            url,
            active_connections: 0,
            total_requests: 0,
            failure_count: 0,
            consecutive_failures: 0,
            healthy: true,
        }
    }
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
pub enum RoutingMode {
    LeastConnections,
    ConsistentHashing,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Config {
    pub routing_mode: RoutingMode,
    pub request_timeout_seconds: u64,
    pub max_consecutive_failures: u32,
    pub health_check_interval_seconds: u64,
}

impl Default for Config {
    fn default() -> Self {
        Self {
            routing_mode: RoutingMode::LeastConnections,
            request_timeout_seconds: 5,
            max_consecutive_failures: 5,
            health_check_interval_seconds: 10,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AddBackendRequest {
    pub url: String,
}

#[derive(Debug, Clone)]
pub struct RouterState {
    pub backends: Arc<DashMap<Uuid, Backend>>,
    pub config: Arc<RwLock<Config>>,
    pub http_client: Arc<RwLock<reqwest::Client>>,
}

impl RouterState {
    pub fn new() -> Self {
        let config = Config::default();
        let http_client = reqwest::Client::builder()
            .timeout(Duration::from_secs(config.request_timeout_seconds))
            .build()
            .unwrap_or_default();

        Self {
            backends: Arc::new(DashMap::new()),
            config: Arc::new(RwLock::new(config)),
            http_client: Arc::new(RwLock::new(http_client)),
        }
    }

    pub async fn update_http_client_timeout(&self, timeout_secs: u64) {
        if let Ok(client) = reqwest::Client::builder()
            .timeout(Duration::from_secs(timeout_secs))
            .build()
        {
            let mut http_client = self.http_client.write().await;
            *http_client = client;
        }
    }
}

fn calculate_hash<T: Hash>(t: &T) -> u64 {
    let mut s = DefaultHasher::new();
    t.hash(&mut s);
    s.finish()
}

pub async fn select_backend(
    state: &RouterState,
    client_ip: Option<String>,
) -> Option<Uuid> {
    let config = state.config.read().await;

    let healthy_backends: Vec<Uuid> = state
        .backends
        .iter()
        .filter(|entry| entry.value().healthy)
        .map(|entry| *entry.key())
        .collect();

    if healthy_backends.is_empty() {
        return None;
    }

    match config.routing_mode {
        RoutingMode::LeastConnections => select_least_connections(state, &healthy_backends),
        RoutingMode::ConsistentHashing => select_consistent_hash(state, &healthy_backends, client_ip),
    }
}

fn select_least_connections(state: &RouterState, healthy_backends: &[Uuid]) -> Option<Uuid> {
    let min_conn = healthy_backends
        .iter()
        .filter_map(|id| state.backends.get(id).map(|b| b.active_connections))
        .min()?;

    let candidates: Vec<Uuid> = healthy_backends
        .iter()
        .filter(|id| {
            state
                .backends
                .get(id)
                .map(|b| b.active_connections == min_conn)
                .unwrap_or(false)
        })
        .cloned()
        .collect();

    let mut rng = rand::thread_rng();
    candidates.choose(&mut rng).cloned()
}

fn select_consistent_hash(
    _state: &RouterState,
    healthy_backends: &[Uuid],
    client_ip: Option<String>,
) -> Option<Uuid> {
    let key = client_ip.unwrap_or_else(|| {
        let mut rng = rand::thread_rng();
        rng.gen::<u64>().to_string()
    });

    let hash = calculate_hash(&key);
    let mut min_diff = u64::MAX;
    let mut selected: Option<Uuid> = None;

    for id in healthy_backends {
        let backend_hash = calculate_hash(id);
        let diff = if backend_hash >= hash {
            backend_hash - hash
        } else {
            u64::MAX - hash + backend_hash
        };

        if diff < min_diff {
            min_diff = diff;
            selected = Some(*id);
        }
    }

    selected
}

pub async fn get_backends(State(state): State<RouterState>) -> impl IntoResponse {
    let backends: Vec<Backend> = state.backends.iter().map(|entry| entry.value().clone()).collect();
    Json(backends).into_response()
}

pub async fn add_backend(
    State(state): State<RouterState>,
    Json(payload): Json<AddBackendRequest>,
) -> impl IntoResponse {
    let backend = Backend::new(payload.url);
    let id = backend.id;
    state.backends.insert(id, backend);
    (StatusCode::CREATED, Json(id)).into_response()
}

pub async fn remove_backend(
    State(state): State<RouterState>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    if state.backends.remove(&id).is_some() {
        StatusCode::NO_CONTENT.into_response()
    } else {
        StatusCode::NOT_FOUND.into_response()
    }
}

pub async fn get_config(State(state): State<RouterState>) -> impl IntoResponse {
    let config = state.config.read().await;
    Json(config.clone()).into_response()
}

pub async fn update_config(
    State(state): State<RouterState>,
    Json(new_config): Json<Config>,
) -> impl IntoResponse {
    let mut config = state.config.write().await;
    let old_timeout = config.request_timeout_seconds;
    *config = new_config.clone();

    if new_config.request_timeout_seconds != old_timeout {
        drop(config);
        state.update_http_client_timeout(new_config.request_timeout_seconds).await;
    }

    Json(new_config).into_response()
}

pub async fn health_check_task(state: RouterState) {
    loop {
        let interval = {
            let config = state.config.read().await;
            config.health_check_interval_seconds
        };
        tokio::time::sleep(Duration::from_secs(interval)).await;

        let unhealthy_ids: Vec<Uuid> = state
            .backends
            .iter()
            .filter(|entry| !entry.value().healthy)
            .map(|entry| *entry.key())
            .collect();

        for id in unhealthy_ids {
            if let Some(backend) = state.backends.get(&id) {
                let url = backend.url.clone();
                let client = {
                    let client = state.http_client.read().await;
                    client.clone()
                };

                let result = tokio::time::timeout(
                    Duration::from_secs(3),
                    client.get(&url).send(),
                )
                .await;

                drop(backend);

                let healthy = matches!(result, Ok(Ok(resp)) if resp.status().is_success());

                if healthy {
                    if let Some(mut backend) = state.backends.get_mut(&id) {
                        backend.healthy = true;
                        backend.consecutive_failures = 0;
                        println!("Backend {} recovered: {}", id, url);
                    }
                }
            }
        }
    }
}

pub async fn proxy_handler(
    State(state): State<RouterState>,
    client_addr: Option<axum::extract::ConnectInfo<SocketAddr>>,
    req: Request<Body>,
) -> Response {
    let client_ip = client_addr.map(|info| info.ip().to_string());

    let backend_id = match select_backend(&state, client_ip).await {
        Some(id) => id,
        None => {
            return Response::builder()
                .status(StatusCode::SERVICE_UNAVAILABLE)
                .body(Body::from("No healthy backends available"))
                .unwrap();
        }
    };

    let (backend_url, request_timeout) = {
        let mut backend = state.backends.get_mut(&backend_id).unwrap();
        backend.active_connections += 1;
        backend.total_requests += 1;
        let config = state.config.read().await;
        (backend.url.clone(), config.request_timeout_seconds)
    };

    let result = forward_request(&state, &backend_url, req, request_timeout).await;

    let success = result.is_ok()
        && result
            .as_ref()
            .map(|r| r.status().is_success())
            .unwrap_or(false);

    let max_consecutive_failures = {
        let config = state.config.read().await;
        config.max_consecutive_failures
    };

    if let Some(mut backend) = state.backends.get_mut(&backend_id) {
        backend.active_connections -= 1;

        if success {
            backend.consecutive_failures = 0;
        } else {
            backend.failure_count += 1;
            backend.consecutive_failures += 1;

            if backend.consecutive_failures >= max_consecutive_failures && backend.healthy {
                backend.healthy = false;
                println!("Backend {} marked as unhealthy: {}", backend_id, backend.url);
            }
        }
    }

    match result {
        Ok(response) => response.into_response(),
        Err(e) => Response::builder()
            .status(StatusCode::BAD_GATEWAY)
            .body(Body::from(format!("Proxy error: {}", e)))
            .unwrap(),
    }
}

async fn forward_request(
    state: &RouterState,
    backend_url: &str,
    req: Request<Body>,
    timeout_secs: u64,
) -> Result<Response, Box<dyn std::error::Error + Send + Sync>> {
    let method = req.method().clone();
    let uri = req.uri().clone();
    let headers = req.headers().clone();
    let body_bytes = axum::body::to_bytes(req.into_body(), usize::MAX)
        .await
        .unwrap_or_default();

    let target_url = format!("{}{}", backend_url.trim_end_matches('/'), uri.path_and_query().map(|p| p.as_str()).unwrap_or(""));

    let request_builder = {
        let client = state.http_client.read().await;
        client
            .request(method.clone(), &target_url)
            .timeout(Duration::from_secs(timeout_secs))
    };

    let mut request_builder = request_builder.body(body_bytes);
    for (key, value) in headers.iter() {
        if let Ok(v) = value.to_str() {
            request_builder = request_builder.header(key, v);
        }
    }

    let response = request_builder.send().await?;

    let status = response.status();
    let mut builder = Response::builder().status(status);

    for (key, value) in response.headers().iter() {
        builder = builder.header(key, value);
    }

    let body_bytes = response.bytes().await?;
    Ok(builder.body(Body::from(body_bytes))?)
}

#[tokio::main]
async fn main() {
    let state = RouterState::new();

    tokio::spawn(health_check_task(state.clone()));

    let app = Router::new()
        .route("/backends", get(get_backends))
        .route("/backends", post(add_backend))
        .route("/backends/:id", delete(remove_backend))
        .route("/config", get(get_config))
        .route("/config", put(update_config))
        .fallback(proxy_handler)
        .with_state(state);

    let port: u16 = std::env::var("PORT")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(8080);

    let addr = SocketAddr::from(([0, 0, 0, 0], port));
    println!("Router listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(
        listener,
        app.into_make_service_with_connect_info::<SocketAddr>(),
    )
    .await
    .unwrap();
}
