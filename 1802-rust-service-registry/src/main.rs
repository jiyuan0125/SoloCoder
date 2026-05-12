use std::collections::{HashMap, HashSet};
use std::sync::Arc;
use std::time::{Duration, Instant};

use axum::{
    extract::{Query, State},
    http::StatusCode,
    response::{IntoResponse, Response, Sse},
    routing::{get, post},
    Json, Router,
};
use axum::response::sse::Event;
use chrono::{DateTime, Utc};
use futures::stream::Stream;
use futures::StreamExt;
use regex::Regex;
use serde::{Deserialize, Serialize};
use tokio::sync::{broadcast, RwLock};
use tokio::time::interval;
use tracing::{info, warn};

const HEARTBEAT_TIMEOUT_SECS: u64 = 60;
const DEFAULT_PAGE_SIZE: usize = 20;
const MAX_TAGS: usize = 10;

#[derive(Debug, Clone)]
struct ServiceInstance {
    name: String,
    address: String,
    tags: Vec<String>,
    weight: u32,
    registered_at: DateTime<Utc>,
    last_heartbeat: Instant,
    callback_url: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct RegisterRequest {
    name: String,
    address: String,
    tags: Vec<String>,
    weight: u32,
    callback_url: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct HeartbeatRequest {
    name: String,
    address: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct QueryParams {
    name: Option<String>,
    tags: Option<Vec<String>>,
    page: Option<usize>,
    page_size: Option<usize>,
}

#[derive(Debug, Clone, Serialize)]
struct ServiceInstanceResponse {
    name: String,
    address: String,
    tags: Vec<String>,
    weight: u32,
    registered_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize)]
struct QueryResponse {
    total: usize,
    page: usize,
    page_size: usize,
    instances: Vec<ServiceInstanceResponse>,
}

#[derive(Debug, Clone, Serialize)]
#[serde(tag = "type", content = "instance")]
enum InstanceEvent {
    Registered(ServiceInstanceResponse),
    Updated(ServiceInstanceResponse),
    Deregistered(ServiceInstanceResponse),
}

type InstanceKey = (String, String);

struct RegistryState {
    instances: HashMap<InstanceKey, ServiceInstance>,
    name_index: HashMap<String, HashSet<InstanceKey>>,
    tag_index: HashMap<String, HashSet<InstanceKey>>,
}

struct AppState {
    registry: RwLock<RegistryState>,
    event_tx: broadcast::Sender<InstanceEvent>,
}

fn validate_name(name: &str) -> Result<(), String> {
    if name.trim().is_empty() {
        return Err("服务名称不能为空".to_string());
    }
    Ok(())
}

fn validate_address(address: &str) -> Result<(), String> {
    let re = Regex::new(r"^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?):([0-9]{1,5})$").unwrap();
    if !re.is_match(address) {
        return Err("地址格式不合法，应为 IP:Port 格式（如 192.168.1.1:8080）".to_string());
    }
    let port = address.split(':').last().unwrap();
    if let Ok(p) = port.parse::<u16>() {
        if p == 0 {
            return Err("端口号不能为 0".to_string());
        }
    } else {
        return Err("端口号不合法".to_string());
    }
    Ok(())
}

fn validate_tags(tags: &[String]) -> Result<(), String> {
    if tags.len() > MAX_TAGS {
        return Err(format!("标签数量不能超过 {} 个", MAX_TAGS));
    }
    for tag in tags {
        if tag.trim().is_empty() {
            return Err("标签不能为空字符串".to_string());
        }
    }
    Ok(())
}

fn instance_to_response(instance: &ServiceInstance) -> ServiceInstanceResponse {
    ServiceInstanceResponse {
        name: instance.name.clone(),
        address: instance.address.clone(),
        tags: instance.tags.clone(),
        weight: instance.weight,
        registered_at: instance.registered_at,
    }
}

async fn register(
    State(state): State<Arc<AppState>>,
    Json(req): Json<RegisterRequest>,
) -> Result<Response, Response> {
    let name = req.name.trim().to_string();
    let address = req.address.trim().to_string();

    validate_name(&name).map_err(|e| {
        (StatusCode::BAD_REQUEST, e).into_response()
    })?;
    validate_address(&address).map_err(|e| {
        (StatusCode::BAD_REQUEST, e).into_response()
    })?;
    validate_tags(&req.tags).map_err(|e| {
        (StatusCode::BAD_REQUEST, e).into_response()
    })?;

    let key: InstanceKey = (name.clone(), address.clone());
    let mut registry = state.registry.write().await;

    let old_instance = registry.instances.remove(&key);
    if let Some(ref old) = old_instance {
        registry.name_index.get_mut(&old.name).map(|s| s.remove(&key));
        for tag in &old.tags {
            registry.tag_index.get_mut(tag).map(|s| s.remove(&key));
        }
    }

    let tags: Vec<String> = req.tags.iter().map(|t| t.trim().to_string()).collect();
    let instance = ServiceInstance {
        name: name.clone(),
        address: address.clone(),
        tags: tags.clone(),
        weight: req.weight,
        registered_at: old_instance.as_ref().map(|i| i.registered_at).unwrap_or_else(Utc::now),
        last_heartbeat: Instant::now(),
        callback_url: req.callback_url,
    };

    registry.instances.insert(key.clone(), instance.clone());

    registry.name_index.entry(name).or_default().insert(key.clone());
    for tag in &tags {
        registry.tag_index.entry(tag.clone()).or_default().insert(key.clone());
    }

    let response = instance_to_response(&instance);
    let event = if old_instance.is_some() {
        InstanceEvent::Updated(response.clone())
    } else {
        InstanceEvent::Registered(response.clone())
    };
    let _ = state.event_tx.send(event);

    Ok((StatusCode::OK, Json(response)).into_response())
}

async fn deregister(
    State(state): State<Arc<AppState>>,
    Json(req): Json<HeartbeatRequest>,
) -> Result<Response, Response> {
    let key: InstanceKey = (req.name.trim().to_string(), req.address.trim().to_string());
    let mut registry = state.registry.write().await;

    if let Some(instance) = registry.instances.remove(&key) {
        registry.name_index.get_mut(&instance.name).map(|s| s.remove(&key));
        for tag in &instance.tags {
            registry.tag_index.get_mut(tag).map(|s| s.remove(&key));
        }

        let response = instance_to_response(&instance);
        if let Some(ref url) = instance.callback_url {
            let event = InstanceEvent::Deregistered(response.clone());
            let url = url.clone();
            tokio::spawn(async move {
                let client = reqwest::Client::builder()
                    .timeout(Duration::from_secs(5))
                    .build();
                if let Ok(client) = client {
                    let result = client.post(&url).json(&event).send().await;
                    match result {
                        Ok(resp) if resp.status().is_success() => {
                            info!("回调通知成功: {}", url);
                        }
                        Ok(resp) => {
                            warn!("回调通知返回非成功状态码: {}, status: {}", url, resp.status());
                        }
                        Err(e) => {
                            warn!("回调通知失败: {}, error: {}", url, e);
                        }
                    }
                }
            });
        }

        let _ = state.event_tx.send(InstanceEvent::Deregistered(response.clone()));
        Ok((StatusCode::OK, Json(response)).into_response())
    } else {
        Ok((StatusCode::NOT_FOUND, Json(serde_json::json!({"error": "服务实例不存在"}))).into_response())
    }
}

async fn heartbeat(
    State(state): State<Arc<AppState>>,
    Json(req): Json<HeartbeatRequest>,
) -> Result<Response, Response> {
    let key: InstanceKey = (req.name.trim().to_string(), req.address.trim().to_string());
    let mut registry = state.registry.write().await;

    if let Some(instance) = registry.instances.get_mut(&key) {
        instance.last_heartbeat = Instant::now();
        Ok((StatusCode::OK, Json(serde_json::json!({"status": "ok"}))).into_response())
    } else {
        Ok((StatusCode::NOT_FOUND, Json(serde_json::json!({"error": "服务实例不存在"}))).into_response())
    }
}

async fn query(
    State(state): State<Arc<AppState>>,
    Query(params): Query<QueryParams>,
) -> Response {
    let registry = state.registry.read().await;

    let page = params.page.unwrap_or(1).max(1);
    let page_size = params.page_size.unwrap_or(DEFAULT_PAGE_SIZE).max(1);

    let matching_keys: HashSet<InstanceKey> = match (params.name, params.tags) {
        (Some(name), Some(tags)) if !tags.is_empty() => {
            let name_keys = registry.name_index.get(&name).cloned().unwrap_or_default();
            let mut tag_keys: Option<HashSet<InstanceKey>> = None;
            for tag in tags {
                let keys = registry.tag_index.get(&tag).cloned().unwrap_or_default();
                if let Some(prev) = tag_keys {
                    tag_keys = Some(prev.intersection(&keys).cloned().collect());
                } else {
                    tag_keys = Some(keys);
                }
            }
            name_keys.intersection(&tag_keys.unwrap_or_default()).cloned().collect()
        }
        (Some(name), _) => {
            registry.name_index.get(&name).cloned().unwrap_or_default()
        }
        (None, Some(tags)) if !tags.is_empty() => {
            let mut tag_keys: Option<HashSet<InstanceKey>> = None;
            for tag in tags {
                let keys = registry.tag_index.get(&tag).cloned().unwrap_or_default();
                if let Some(prev) = tag_keys {
                    tag_keys = Some(prev.intersection(&keys).cloned().collect());
                } else {
                    tag_keys = Some(keys);
                }
            }
            tag_keys.unwrap_or_default()
        }
        (None, _) => {
            registry.instances.keys().cloned().collect()
        }
    };

    let mut instances: Vec<&ServiceInstance> = matching_keys
        .iter()
        .filter_map(|k| registry.instances.get(k))
        .filter(|i| i.weight > 0)
        .collect();

    instances.sort_by(|a, b| b.registered_at.cmp(&a.registered_at));

    let total = instances.len();
    let start = (page - 1) * page_size;
    let paged: Vec<ServiceInstanceResponse> = instances
        .into_iter()
        .skip(start)
        .take(page_size)
        .map(instance_to_response)
        .collect();

    (StatusCode::OK, Json(QueryResponse {
        total,
        page,
        page_size,
        instances: paged,
    })).into_response()
}

async fn subscribe(
    State(state): State<Arc<AppState>>,
) -> Sse<impl Stream<Item = Result<Event, axum::BoxError>>> {
    let rx = state.event_tx.subscribe();
    let stream = tokio_stream::wrappers::BroadcastStream::new(rx)
        .filter_map(|result| async move {
            match result {
                Ok(event) => {
                    let data = serde_json::to_string(&event).ok()?;
                    Some(Ok(Event::default().data(data)))
                }
                Err(e) => {
                    warn!("SSE 接收错误: {}", e);
                    None
                }
            }
        });
    Sse::new(stream).keep_alive(axum::response::sse::KeepAlive::default())
}

async fn heartbeat_checker(state: Arc<AppState>) {
    let mut interval = interval(Duration::from_secs(1));
    loop {
        interval.tick().await;

        let now = Instant::now();
        let mut registry = state.registry.write().await;
        let mut expired: Vec<(InstanceKey, ServiceInstance)> = Vec::new();

        registry.instances.retain(|key, instance| {
            if now.duration_since(instance.last_heartbeat) > Duration::from_secs(HEARTBEAT_TIMEOUT_SECS) {
                expired.push((key.clone(), instance.clone()));
                false
            } else {
                true
            }
        });

        for (key, instance) in expired {
            registry.name_index.get_mut(&instance.name).map(|s| s.remove(&key));
            for tag in &instance.tags {
                registry.tag_index.get_mut(tag).map(|s| s.remove(&key));
            }

            let response = instance_to_response(&instance);
            if let Some(ref url) = instance.callback_url {
                let event = InstanceEvent::Deregistered(response.clone());
                let url = url.clone();
                tokio::spawn(async move {
                    let client = reqwest::Client::builder()
                        .timeout(Duration::from_secs(5))
                        .build();
                    if let Ok(client) = client {
                        let result = client.post(&url).json(&event).send().await;
                        match result {
                            Ok(resp) if resp.status().is_success() => {
                                info!("回调通知成功: {}", url);
                            }
                            Ok(resp) => {
                                warn!("回调通知返回非成功状态码: {}, status: {}", url, resp.status());
                            }
                            Err(e) => {
                                warn!("回调通知失败: {}, error: {}", url, e);
                            }
                        }
                    }
                });
            }

            let _ = state.event_tx.send(InstanceEvent::Deregistered(response));
            info!("实例因心跳超时而注销: {}@{}", instance.name, instance.address);
        }
    }
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_env_filter(tracing_subscriber::EnvFilter::from_default_env())
        .init();

    let registry = RegistryState {
        instances: HashMap::new(),
        name_index: HashMap::new(),
        tag_index: HashMap::new(),
    };

    let (event_tx, _) = broadcast::channel(100);

    let state = Arc::new(AppState {
        registry: RwLock::new(registry),
        event_tx,
    });

    let checker_state = state.clone();
    tokio::spawn(async move {
        heartbeat_checker(checker_state).await;
    });

    let app = Router::new()
        .route("/register", post(register))
        .route("/deregister", post(deregister))
        .route("/heartbeat", post(heartbeat))
        .route("/query", get(query))
        .route("/subscribe", get(subscribe))
        .with_state(state);

    let port = std::env::var("PORT")
        .unwrap_or_else(|_| "3000".to_string())
        .parse::<u16>()
        .expect("PORT 环境变量必须是有效的端口号");

    let addr = format!("0.0.0.0:{}", port).parse().unwrap();
    info!("服务注册中心正在监听: {}", addr);

    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}
