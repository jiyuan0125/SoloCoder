use axum::{
    body::Bytes,
    extract::{Path, Query, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{delete, get, post},
    Json, Router,
};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::Mutex;

use crate::cache::{validate_key, validate_ttl, CacheService};
use crate::diag::DiagLogger;

#[derive(Debug, Deserialize)]
pub struct SetQuery {
    pub ttl: Option<i64>,
}

#[derive(Debug, Deserialize)]
pub struct BatchSetItem {
    pub key: String,
    pub value: String,
    pub ttl: Option<i64>,
}

#[derive(Debug, Deserialize)]
pub struct BatchSetRequest {
    pub items: Vec<BatchSetItem>,
}

#[derive(Debug, Deserialize)]
pub struct BatchGetQuery {
    pub keys: String,
}

#[derive(Debug, Serialize)]
pub struct BatchGetResponse {
    pub hits: HashMap<String, String>,
    pub misses: Vec<String>,
}

#[derive(Debug, Serialize)]
pub struct ErrorResponse {
    pub error: String,
}

type AppState = (Arc<Mutex<CacheService>>, Arc<Mutex<DiagLogger>>);

pub fn create_router(
    cache: Arc<Mutex<CacheService>>,
    logger: Arc<Mutex<DiagLogger>>,
) -> Router {
    Router::new()
        .route("/cache/:key", get(get_handler))
        .route("/cache/:key", post(set_handler))
        .route("/cache/:key", delete(delete_handler))
        .route("/cache/batch/set", post(batch_set_handler))
        .route("/cache/batch/get", get(batch_get_handler))
        .route("/health", get(health_handler))
        .with_state((cache, logger))
}

async fn health_handler() -> impl IntoResponse {
    (StatusCode::OK, "OK")
}

async fn get_handler(
    Path(key): Path<String>,
    State((cache, _logger)): State<AppState>,
) -> impl IntoResponse {
    if let Err(err) = validate_key(&key) {
        return (
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse { error: err }),
        ).into_response();
    }

    let mut cache = cache.lock().await;
    match cache.get(&key) {
        Some(entry) => {
            (StatusCode::OK, Bytes::from(entry.value.clone())).into_response()
        }
        None => {
            (StatusCode::NOT_FOUND).into_response()
        }
    }
}

async fn set_handler(
    Path(key): Path<String>,
    Query(query): Query<SetQuery>,
    State((cache, logger)): State<AppState>,
    body: Bytes,
) -> impl IntoResponse {
    if let Err(err) = validate_key(&key) {
        return (
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse { error: err }),
        ).into_response();
    }

    let ttl = query.ttl.unwrap_or(0);
    if let Err(err) = validate_ttl(ttl) {
        return (
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse { error: err }),
        ).into_response();
    }

    let value: Vec<u8> = body.to_vec();
    let mut cache = cache.lock().await;
    let eviction_failed = cache.set(key, value, ttl);

    if eviction_failed {
        let mut logger = logger.lock().await;
        logger.log_eviction_failure("LRU eviction failed during set");
    }

    (StatusCode::OK).into_response()
}

async fn delete_handler(
    Path(key): Path<String>,
    State((cache, _logger)): State<AppState>,
) -> impl IntoResponse {
    if let Err(err) = validate_key(&key) {
        return (
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse { error: err }),
        ).into_response();
    }

    let mut cache = cache.lock().await;
    if cache.delete(&key) {
        (StatusCode::OK).into_response()
    } else {
        (StatusCode::NOT_FOUND).into_response()
    }
}

async fn batch_set_handler(
    State((cache, logger)): State<AppState>,
    Json(body): Json<BatchSetRequest>,
) -> impl IntoResponse {
    let mut entries: Vec<(String, Vec<u8>, i64)> = Vec::new();

    for item in &body.items {
        if let Err(err) = validate_key(&item.key) {
            return (
                StatusCode::BAD_REQUEST,
                Json(ErrorResponse { error: format!("key 校验失败: {}", err) }),
            ).into_response();
        }
        let ttl = item.ttl.unwrap_or(0);
        if let Err(err) = validate_ttl(ttl) {
            return (
                StatusCode::BAD_REQUEST,
                Json(ErrorResponse { error: format!("ttl 校验失败: {}", err) }),
            ).into_response();
        }
        entries.push((item.key.clone(), item.value.as_bytes().to_vec(), ttl));
    }

    let mut cache = cache.lock().await;
    let any_eviction_failed = cache.batch_set(entries);

    if any_eviction_failed {
        let mut logger = logger.lock().await;
        logger.log_eviction_failure("LRU eviction failed during batch set");
    }

    (StatusCode::OK).into_response()
}

async fn batch_get_handler(
    Query(query): Query<BatchGetQuery>,
    State((cache, _logger)): State<AppState>,
) -> impl IntoResponse {
    let keys: Vec<&str> = query.keys.split(',').collect();
    let mut validated_keys: Vec<&str> = Vec::new();

    for key in &keys {
        if let Err(err) = validate_key(key) {
            return (
                StatusCode::BAD_REQUEST,
                Json(ErrorResponse { error: format!("key 校验失败: {}", err) }),
            ).into_response();
        }
        validated_keys.push(*key);
    }

    let mut cache = cache.lock().await;
    let (hits_bytes, misses) = cache.batch_get(validated_keys);

    let mut hits: HashMap<String, String> = HashMap::new();
    for (key, value) in hits_bytes {
        hits.insert(key, String::from_utf8_lossy(&value).to_string());
    }

    (
        StatusCode::OK,
        Json(BatchGetResponse { hits, misses }),
    ).into_response()
}
