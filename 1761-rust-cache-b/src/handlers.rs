use std::sync::Arc;

use axum::extract::{Path, State};
use axum::http::StatusCode;
use axum::response::{IntoResponse, Json};
use axum::response::Response;
use serde_json::Value as JsonValue;

use crate::cache::{CacheEntry, CacheManager, CacheValue};
use crate::models::{
    ApiResponse, BatchDeleteRequest, BatchDeleteResponse, BatchGetRequest, BatchGetResponse,
    BatchGetResponseItem, BatchResultItem, BatchSetRequest, BatchSetResponse, CacheEntryResponse,
    HealthResponse, SetCacheRequest, StatsResponse,
};

type CacheState = Arc<CacheManager>;

fn value_to_json(value: CacheValue) -> JsonValue {
    match value {
        CacheValue::String(s) => JsonValue::String(s),
        CacheValue::Integer(n) => JsonValue::Number(serde_json::Number::from(n)),
        CacheValue::Float(f) => {
            if let Some(n) = serde_json::Number::from_f64(f) {
                JsonValue::Number(n)
            } else {
                JsonValue::Null
            }
        }
        CacheValue::Boolean(b) => JsonValue::Bool(b),
        CacheValue::Object(v) => v,
    }
}

fn json_to_value(json: JsonValue) -> CacheValue {
    match json {
        JsonValue::String(s) => CacheValue::String(s),
        JsonValue::Number(n) => {
            if let Some(i) = n.as_i64() {
                CacheValue::Integer(i)
            } else if let Some(f) = n.as_f64() {
                CacheValue::Float(f)
            } else {
                CacheValue::Object(JsonValue::Number(n))
            }
        }
        JsonValue::Bool(b) => CacheValue::Boolean(b),
        JsonValue::Null => CacheValue::Object(JsonValue::Null),
        other => CacheValue::Object(other),
    }
}

fn entry_to_response(entry: CacheEntry) -> CacheEntryResponse {
    let remaining_ttl = entry.remaining_ttl();
    CacheEntryResponse {
        key: entry.key.clone(),
        value: value_to_json(entry.value),
        remaining_ttl,
    }
}

pub async fn health() -> impl IntoResponse {
    (
        StatusCode::OK,
        Json(HealthResponse {
            status: "healthy".to_string(),
            service: "rust-cache".to_string(),
            version: env!("CARGO_PKG_VERSION").to_string(),
        }),
    )
}

pub async fn get_cache(
    Path(key): Path<String>,
    State(cache): State<CacheState>,
) -> Response {
    match cache.get_with_meta(&key) {
        Some(entry) => (
            StatusCode::OK,
            Json(ApiResponse::ok(entry_to_response(entry))),
        )
            .into_response(),
        None => (
            StatusCode::NOT_FOUND,
            Json(ApiResponse::<()>::err(format!("Key '{}' not found", key))),
        )
            .into_response(),
    }
}

pub async fn set_cache(
    Path(key): Path<String>,
    State(cache): State<CacheState>,
    Json(body): Json<SetCacheRequest>,
) -> Response {
    let cache_value = json_to_value(body.value);
    cache.set(key.clone(), cache_value, body.ttl_seconds);
    
    (
        StatusCode::OK,
        Json(ApiResponse::<()>::empty_ok()),
    )
        .into_response()
}

pub async fn delete_cache(
    Path(key): Path<String>,
    State(cache): State<CacheState>,
) -> Response {
    let deleted = cache.delete(&key);
    
    if deleted {
        (
            StatusCode::OK,
            Json(ApiResponse::<()>::empty_ok()),
        )
            .into_response()
    } else {
        (
            StatusCode::NOT_FOUND,
            Json(ApiResponse::<()>::err(format!("Key '{}' not found", key))),
        )
            .into_response()
    }
}

pub async fn batch_get(
    State(cache): State<CacheState>,
    Json(body): Json<BatchGetRequest>,
) -> Response {
    let mut results = Vec::with_capacity(body.keys.len());
    let mut total_found = 0;

    for key in body.keys.iter() {
        match cache.get_with_meta(key) {
            Some(entry) => {
                total_found += 1;
                let remaining_ttl = entry.remaining_ttl();
                results.push(BatchGetResponseItem {
                    key: key.clone(),
                    found: true,
                    value: Some(value_to_json(entry.value)),
                    remaining_ttl,
                });
            }
            None => {
                results.push(BatchGetResponseItem {
                    key: key.clone(),
                    found: false,
                    value: None,
                    remaining_ttl: None,
                });
            }
        }
    }

    (
        StatusCode::OK,
        Json(ApiResponse::ok(BatchGetResponse {
            results,
            total_found,
            total_keys: body.keys.len(),
        })),
    )
        .into_response()
}

pub async fn batch_set(
    State(cache): State<CacheState>,
    Json(body): Json<BatchSetRequest>,
) -> Response {
    let mut results = Vec::with_capacity(body.items.len());
    let mut total_success = 0;

    for item in body.items.iter() {
        let cache_value = json_to_value(item.value.clone());
        cache.set(item.key.clone(), cache_value, item.ttl_seconds);
        
        total_success += 1;
        results.push(BatchResultItem {
            key: item.key.clone(),
            success: true,
            error: None,
        });
    }

    let response = BatchSetResponse {
        results,
        total_success,
        total_failed: body.items.len() - total_success,
        total_items: body.items.len(),
    };

    (
        StatusCode::OK,
        Json(ApiResponse::ok(response)),
    )
        .into_response()
}

pub async fn batch_delete(
    State(cache): State<CacheState>,
    Json(body): Json<BatchDeleteRequest>,
) -> Response {
    let mut results = Vec::with_capacity(body.keys.len());
    let mut total_deleted = 0;

    for key in body.keys.iter() {
        let deleted = cache.delete(key);
        
        if deleted {
            total_deleted += 1;
        }

        results.push(BatchResultItem {
            key: key.clone(),
            success: deleted,
            error: if deleted { None } else { Some("Key not found".to_string()) },
        });
    }

    let response = BatchDeleteResponse {
        results,
        total_deleted,
        total_keys: body.keys.len(),
    };

    (
        StatusCode::OK,
        Json(ApiResponse::ok(response)),
    )
        .into_response()
}

pub async fn get_stats(State(cache): State<CacheState>) -> Response {
    let stats = cache.stats();
    
    let response = StatsResponse {
        total_requests: stats.get_total_requests(),
        cache_hits: stats.get_hits(),
        cache_misses: stats.get_misses(),
        hit_rate: stats.get_hit_rate(),
        total_sets: stats.get_sets(),
        total_deletes: stats.get_deletes(),
        total_evictions: stats.get_evictions(),
        expired_removed: stats.get_expires_removed(),
        current_size: cache.len(),
        capacity: cache.capacity(),
    };

    (
        StatusCode::OK,
        Json(ApiResponse::ok(response)),
    )
        .into_response()
}

pub async fn clear_cache(State(cache): State<CacheState>) -> Response {
    cache.clear();
    
    (
        StatusCode::OK,
        Json(ApiResponse::<()>::empty_ok()),
    )
        .into_response()
}
