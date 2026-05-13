use crate::cache::{CacheEngine, NamespaceStats};
use axum::{
    extract::{Path, Query, State},
    http::{HeaderMap, HeaderValue, StatusCode},
    response::{IntoResponse, Json, Response},
    routing::{delete, get, put},
    Router,
};
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use tokio::sync::RwLock;

#[derive(Debug, Clone)]
pub struct AppState {
    pub cache: Arc<RwLock<CacheEngine>>,
}

#[derive(Debug, Deserialize)]
pub struct SetQuery {
    #[serde(default)]
    pub ttl: Option<u64>,
}

#[derive(Debug, Deserialize)]
pub struct ScanQuery {
    pub prefix: String,
}

#[derive(Debug, Deserialize)]
pub struct DeletePrefixQuery {
    pub prefix: String,
}

#[derive(Debug, Deserialize)]
pub struct ConfigRequest {
    pub capacity: Option<usize>,
    pub default_ttl: Option<u64>,
}

#[derive(Debug, Serialize)]
pub struct StatsResponse {
    pub hit_rate: f64,
    pub entry_count: usize,
    pub total_queries: u64,
    pub evictions: u64,
    pub expired_cleanups: u64,
}

#[derive(Debug, Serialize)]
pub struct NamespaceStatsResponse {
    pub namespace: String,
    pub total_queries: u64,
    pub cache_hits: u64,
    pub hit_rate: f64,
}

#[derive(Debug, Serialize)]
pub struct ConfigResponse {
    pub capacity: usize,
    pub default_ttl: u64,
}

#[derive(Debug, Serialize)]
pub struct ScanEntry {
    pub key: String,
    pub value: serde_json::Value,
}

#[derive(Debug, Serialize)]
pub struct DeleteCountResponse {
    pub deleted: usize,
}

pub fn create_router(cache: Arc<RwLock<CacheEngine>>) -> Router {
    let state = AppState { cache };

    Router::new()
        .route("/cache/:key", get(get_handler))
        .route("/cache/:key", put(set_handler))
        .route("/cache/:key", delete(delete_handler))
        .route("/admin/stats", get(stats_handler))
        .route("/admin/stats/namespace/:namespace", get(namespace_stats_handler))
        .route("/admin/config", get(config_handler))
        .route("/admin/config", put(update_config_handler))
        .route("/admin/prefix", delete(delete_prefix_handler))
        .route("/admin/scan", get(scan_handler))
        .route("/admin/namespace/:namespace", delete(clear_namespace_handler))
        .with_state(state)
}

async fn get_handler(
    State(state): State<AppState>,
    Path(key): Path<String>,
) -> Response {
    let mut cache = state.cache.write().await;
    match cache.get(&key) {
        Some(value) => {
            let mut headers = HeaderMap::new();
            headers.insert(
                "Content-Type",
                HeaderValue::from_static("application/octet-stream"),
            );
            (StatusCode::OK, headers, value.clone()).into_response()
        }
        None => (StatusCode::NOT_FOUND, "Not Found").into_response(),
    }
}

async fn set_handler(
    State(state): State<AppState>,
    Path(key): Path<String>,
    Query(query): Query<SetQuery>,
    body: axum::body::Bytes,
) -> Response {
    let value = body.to_vec();

    if value.len() > 1_048_576 {
        return (StatusCode::UNPROCESSABLE_ENTITY, "Value too large (max 1MB)").into_response();
    }

    if key.len() > 256 {
        return (StatusCode::UNPROCESSABLE_ENTITY, "Key too long (max 256 bytes)").into_response();
    }

    let mut cache = state.cache.write().await;
    if cache.set(key, value, query.ttl) {
        StatusCode::NO_CONTENT.into_response()
    } else {
        (StatusCode::UNPROCESSABLE_ENTITY, "Failed to set value").into_response()
    }
}

async fn delete_handler(
    State(state): State<AppState>,
    Path(key): Path<String>,
) -> Response {
    let mut cache = state.cache.write().await;
    if cache.delete(&key) {
        StatusCode::NO_CONTENT.into_response()
    } else {
        (StatusCode::NOT_FOUND, "Not Found").into_response()
    }
}

async fn scan_handler(
    State(state): State<AppState>,
    Query(query): Query<ScanQuery>,
) -> Response {
    let cache = state.cache.read().await;
    let entries = cache.scan_by_prefix(&query.prefix);
    let mut result = Vec::new();

    for (key, value) in entries {
        let value_json = match serde_json::from_slice::<serde_json::Value>(&value) {
            Ok(v) => v,
            Err(_) => serde_json::json!(value),
        };
        result.push(ScanEntry {
            key,
            value: value_json,
        });
    }

    (StatusCode::OK, Json(result)).into_response()
}

async fn delete_prefix_handler(
    State(state): State<AppState>,
    Query(query): Query<DeletePrefixQuery>,
) -> Response {
    let mut cache = state.cache.write().await;
    let count = cache.delete_by_prefix(&query.prefix);
    (StatusCode::OK, Json(DeleteCountResponse { deleted: count })).into_response()
}

async fn stats_handler(State(state): State<AppState>) -> Response {
    let cache = state.cache.read().await;
    let stats = cache.stats();
    let response = StatsResponse {
        hit_rate: stats.hit_rate(),
        entry_count: cache.entry_count(),
        total_queries: stats.total_queries,
        evictions: stats.evictions,
        expired_cleanups: stats.expired_cleanups,
    };
    (StatusCode::OK, Json(response)).into_response()
}

async fn namespace_stats_handler(
    State(state): State<AppState>,
    Path(namespace): Path<String>,
) -> Response {
    let cache = state.cache.read().await;
    let ns_stats: NamespaceStats = cache.namespace_stats(&namespace);
    let hit_rate = if ns_stats.total_queries > 0 {
        ns_stats.cache_hits as f64 / ns_stats.total_queries as f64
    } else {
        0.0
    };
    let response = NamespaceStatsResponse {
        namespace,
        total_queries: ns_stats.total_queries,
        cache_hits: ns_stats.cache_hits,
        hit_rate,
    };
    (StatusCode::OK, Json(response)).into_response()
}

async fn config_handler(State(state): State<AppState>) -> Response {
    let cache = state.cache.read().await;
    let response = ConfigResponse {
        capacity: cache.capacity(),
        default_ttl: cache.default_ttl(),
    };
    (StatusCode::OK, Json(response)).into_response()
}

async fn update_config_handler(
    State(state): State<AppState>,
    Json(payload): Json<ConfigRequest>,
) -> Response {
    let mut cache = state.cache.write().await;
    if let Some(cap) = payload.capacity {
        cache.set_capacity(cap);
    }
    if let Some(ttl) = payload.default_ttl {
        cache.set_default_ttl(ttl);
    }
    let response = ConfigResponse {
        capacity: cache.capacity(),
        default_ttl: cache.default_ttl(),
    };
    (StatusCode::OK, Json(response)).into_response()
}

async fn clear_namespace_handler(
    State(state): State<AppState>,
    Path(namespace): Path<String>,
) -> Response {
    let mut cache = state.cache.write().await;
    let count = cache.clear_namespace(&namespace);
    (StatusCode::OK, Json(DeleteCountResponse { deleted: count })).into_response()
}
