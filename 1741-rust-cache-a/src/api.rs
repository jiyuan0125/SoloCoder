use std::collections::HashMap;
use std::sync::Arc;
use std::time::Duration;

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{delete, get, post},
    Json, Router,
};
use serde::{Deserialize, Serialize};
use tokio::sync::RwLock;

use crate::cache::{Cache, CacheEntry};

type AppState = Arc<RwLock<Cache>>;

#[derive(Debug, Deserialize)]
pub struct SetRequest {
    pub key: String,
    pub value: String,
    #[serde(default)]
    pub ttl_seconds: Option<u64>,
}

#[derive(Debug, Deserialize)]
pub struct BatchSetRequest {
    pub entries: Vec<BatchEntry>,
}

#[derive(Debug, Deserialize)]
pub struct BatchEntry {
    pub key: String,
    pub value: String,
    #[serde(default)]
    pub ttl_seconds: Option<u64>,
}

#[derive(Debug, Deserialize)]
pub struct BatchGetRequest {
    pub keys: Vec<String>,
}

#[derive(Debug, Serialize)]
pub struct BatchSetResponse {
    pub success: Vec<String>,
    pub failed: Vec<String>,
}

#[derive(Debug, Serialize)]
pub struct BatchGetResponse {
    pub found: HashMap<String, String>,
    pub not_found: Vec<String>,
}

#[derive(Debug, Serialize)]
pub struct StatsResponse {
    pub total_entries: usize,
    pub hits: u64,
    pub misses: u64,
    pub evictions: u64,
    pub hit_rate: Option<f64>,
}

pub fn create_router(state: AppState) -> Router {
    Router::new()
        .route("/cache/:key", get(get_handler))
        .route("/cache", post(set_handler))
        .route("/cache/:key", delete(delete_handler))
        .route("/cache/batch/set", post(batch_set_handler))
        .route("/cache/batch/get", post(batch_get_handler))
        .route("/cache/import", post(import_handler))
        .route("/cache/stats", get(stats_handler))
        .with_state(state)
}

async fn get_handler(
    State(state): State<AppState>,
    Path(key): Path<String>,
) -> impl IntoResponse {
    let mut cache = state.write().await;
    match cache.get(&key) {
        Some(entry) => (StatusCode::OK, Json(serde_json::json!({
            "key": key,
            "value": entry.value
        }))).into_response(),
        None => (StatusCode::NOT_FOUND, Json(serde_json::json!({
            "error": "not found"
        }))).into_response(),
    }
}

async fn set_handler(
    State(state): State<AppState>,
    Json(req): Json<SetRequest>,
) -> impl IntoResponse {
    let entry = CacheEntry {
        value: req.value,
        ttl: req.ttl_seconds.map(Duration::from_secs),
        created_at: None,
    };

    let mut cache = state.write().await;
    cache.set(req.key.clone(), entry);

    (StatusCode::OK, Json(serde_json::json!({
        "status": "ok",
        "key": req.key
    }))).into_response()
}

async fn delete_handler(
    State(state): State<AppState>,
    Path(key): Path<String>,
) -> impl IntoResponse {
    let mut cache = state.write().await;
    if cache.delete(&key) {
        (StatusCode::OK, Json(serde_json::json!({
            "status": "deleted"
        }))).into_response()
    } else {
        (StatusCode::NOT_FOUND, Json(serde_json::json!({
            "error": "not found"
        }))).into_response()
    }
}

async fn batch_set_handler(
    State(state): State<AppState>,
    Json(req): Json<BatchSetRequest>,
) -> impl IntoResponse {
    let mut entries = HashMap::new();
    for entry in req.entries {
        entries.insert(entry.key, CacheEntry {
            value: entry.value,
            ttl: entry.ttl_seconds.map(Duration::from_secs),
            created_at: None,
        });
    }

    let mut cache = state.write().await;
    let (success, failed) = cache.batch_set(entries);

    (StatusCode::OK, Json(BatchSetResponse { success, failed })).into_response()
}

async fn batch_get_handler(
    State(state): State<AppState>,
    Json(req): Json<BatchGetRequest>,
) -> impl IntoResponse {
    let mut cache = state.write().await;
    let (found, not_found) = cache.batch_get(&req.keys);

    (StatusCode::OK, Json(BatchGetResponse { found, not_found })).into_response()
}

async fn import_handler(
    State(state): State<AppState>,
    Json(data): Json<HashMap<String, CacheEntry>>,
) -> impl IntoResponse {
    let mut cache = state.write().await;
    let count = cache.import(data);

    (StatusCode::OK, Json(serde_json::json!({
        "imported": count
    }))).into_response()
}

async fn stats_handler(
    State(state): State<AppState>,
) -> impl IntoResponse {
    let cache = state.read().await;
    let stats = cache.stats();

    let total = stats.hits + stats.misses;
    let hit_rate = if total > 0 {
        Some(stats.hits as f64 / total as f64)
    } else {
        None
    };

    (StatusCode::OK, Json(StatsResponse {
        total_entries: stats.total_entries,
        hits: stats.hits,
        misses: stats.misses,
        evictions: stats.evictions,
        hit_rate,
    })).into_response()
}
