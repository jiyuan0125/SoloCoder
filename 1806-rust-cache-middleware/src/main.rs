mod cache;
mod data_source;
mod stats;

use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    response::{IntoResponse, Json},
    routing::{delete, get, post, put},
    Router,
};
use cache::CacheManager;
use data_source::DataSourceManager;
use serde::{Deserialize, Serialize};
use std::{collections::HashMap, sync::Arc, time::Duration};
use stats::CacheStats;

#[derive(Clone)]
struct AppState {
    cache: Arc<CacheManager>,
    sources: Arc<DataSourceManager>,
    stats: Arc<CacheStats>,
}

#[derive(Debug, Deserialize)]
struct SetRequest {
    value: String,
    ttl: Option<u64>,
}

#[derive(Debug, Deserialize)]
struct SourceCreateRequest {
    name: String,
    url_template: String,
    method: Option<String>,
}

#[derive(Debug, Deserialize)]
struct ReadThroughQuery {
    #[serde(flatten)]
    params: HashMap<String, String>,
}

#[derive(Debug, Serialize)]
struct ReadThroughResponse {
    data: serde_json::Value,
    from_cache: bool,
}

async fn manual_get(
    State(state): State<AppState>,
    Path(key): Path<String>,
) -> impl IntoResponse {
    state.stats.increment_get_requests();
    match state.cache.get(&key).await {
        Some(val) => {
            state.stats.increment_hits();
            (StatusCode::OK, Json(serde_json::json!({ "value": val }))).into_response()
        }
        None => {
            state.stats.increment_misses();
            (StatusCode::NOT_FOUND, Json(serde_json::json!({ "error": "key not found" }))).into_response()
        }
    }
}

async fn manual_set(
    State(state): State<AppState>,
    Path(key): Path<String>,
    Json(req): Json<SetRequest>,
) -> impl IntoResponse {
    let ttl = req.ttl.unwrap_or(300);
    state.cache.set(key, req.value, Duration::from_secs(ttl)).await;
    (StatusCode::OK, Json(serde_json::json!({ "status": "ok" }))).into_response()
}

async fn manual_delete(
    State(state): State<AppState>,
    Path(key): Path<String>,
) -> impl IntoResponse {
    state.cache.delete(&key).await;
    (StatusCode::OK, Json(serde_json::json!({ "status": "ok" }))).into_response()
}

async fn read_through_get(
    State(state): State<AppState>,
    Path((source_name, key)): Path<(String, String)>,
    Query(query): Query<ReadThroughQuery>,
) -> impl IntoResponse {
    let source = match state.sources.get(&source_name) {
        Some(s) => s,
        None => {
            return (
                StatusCode::NOT_FOUND,
                Json(serde_json::json!({ "error": "data source not found" })),
            )
                .into_response();
        }
    };

    state.stats.increment_get_requests();

    if let Some(cached) = state.cache.get(&key).await {
        state.stats.increment_hits();
        let data: serde_json::Value = match serde_json::from_str(&cached) {
            Ok(v) => v,
            Err(_) => serde_json::Value::String(cached),
        };
        return (
            StatusCode::OK,
            Json(ReadThroughResponse {
                data,
                from_cache: true,
            }),
        )
            .into_response();
    }

    state.stats.increment_misses();
    state.stats.increment_source_calls();

    let mut params = query.params.clone();
    params.insert("key".to_string(), key.clone());

    match state.sources.fetch(&source, &params).await {
        Ok((data, status)) => {
            if status.is_success() {
                let json_str = serde_json::to_string(&data).unwrap_or_default();
                state
                    .cache
                    .set(key.clone(), json_str, Duration::from_secs(300))
                    .await;
                (
                    status,
                    Json(ReadThroughResponse {
                        data,
                        from_cache: false,
                    }),
                )
                    .into_response()
            } else {
                (
                    status,
                    Json(serde_json::json!({ "error": data, "source": source_name })),
                )
                    .into_response()
            }
        }
        Err(e) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(serde_json::json!({ "error": e.to_string() })),
        )
            .into_response(),
    }
}

async fn source_list(State(state): State<AppState>) -> impl IntoResponse {
    let sources: Vec<_> = state
        .sources
        .list()
        .into_iter()
        .map(|s| serde_json::json!({ "name": s.name, "url_template": s.url_template, "method": s.method }))
        .collect();
    (StatusCode::OK, Json(sources)).into_response()
}

async fn source_get(State(state): State<AppState>, Path(name): Path<String>) -> impl IntoResponse {
    match state.sources.get(&name) {
        Some(s) => (
            StatusCode::OK,
            Json(serde_json::json!({ "name": s.name, "url_template": s.url_template, "method": s.method })),
        )
            .into_response(),
        None => (
            StatusCode::NOT_FOUND,
            Json(serde_json::json!({ "error": "data source not found" })),
        )
            .into_response(),
    }
}

async fn source_create(
    State(state): State<AppState>,
    Json(req): Json<SourceCreateRequest>,
) -> impl IntoResponse {
    match state
        .sources
        .create(req.name, req.url_template, req.method.unwrap_or_else(|| "GET".to_string()))
    {
        Ok(_) => (StatusCode::CREATED, Json(serde_json::json!({ "status": "created" }))).into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({ "error": e.to_string() })),
        )
            .into_response(),
    }
}

async fn source_update(
    State(state): State<AppState>,
    Path(name): Path<String>,
    Json(req): Json<SourceCreateRequest>,
) -> impl IntoResponse {
    match state.sources.update(&name, req.url_template, req.method) {
        Ok(_) => (StatusCode::OK, Json(serde_json::json!({ "status": "updated" }))).into_response(),
        Err(e) => (
            StatusCode::NOT_FOUND,
            Json(serde_json::json!({ "error": e.to_string() })),
        )
            .into_response(),
    }
}

async fn source_delete(State(state): State<AppState>, Path(name): Path<String>) -> impl IntoResponse {
    match state.sources.delete(&name) {
        Ok(_) => (StatusCode::OK, Json(serde_json::json!({ "status": "deleted" }))).into_response(),
        Err(e) => (
            StatusCode::NOT_FOUND,
            Json(serde_json::json!({ "error": e.to_string() })),
        )
            .into_response(),
    }
}

async fn get_stats(State(state): State<AppState>) -> impl IntoResponse {
    let stats = state.stats.snapshot();
    let current_keys = state.cache.len().await;
    (
        StatusCode::OK,
        Json(serde_json::json!({
            "get_requests": stats.get_requests,
            "hits": stats.hits,
            "misses": stats.misses,
            "hit_rate": if stats.get_requests > 0 {
                stats.hits as f64 / stats.get_requests as f64
            } else {
                0.0
            },
            "current_keys": current_keys,
            "source_calls": stats.source_calls,
        })),
    )
        .into_response()
}

async fn reset_stats(State(state): State<AppState>) -> impl IntoResponse {
    state.stats.reset();
    (StatusCode::OK, Json(serde_json::json!({ "status": "reset" }))).into_response()
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();

    let cache = Arc::new(CacheManager::new(1000));
    let sources = Arc::new(DataSourceManager::new());
    let stats = Arc::new(CacheStats::new());

    let cache_clone = cache.clone();
    tokio::spawn(async move {
        let mut interval = tokio::time::interval(Duration::from_secs(60));
        loop {
            interval.tick().await;
            tracing::debug!("Running cache cleanup...");
            cache_clone.cleanup_expired().await;
        }
    });

    let state = AppState { cache, sources, stats };

    let app = Router::new()
        .route("/cache/:key", get(manual_get).put(manual_set).delete(manual_delete))
        .route("/read-through/:source_name/:key", get(read_through_get))
        .route("/sources", get(source_list).post(source_create))
        .route("/sources/:name", get(source_get).put(source_update).delete(source_delete))
        .route("/stats", get(get_stats).post(reset_stats))
        .with_state(state);

    let port: u16 = std::env::var("PORT")
        .unwrap_or_else(|_| "3000".to_string())
        .parse()
        .expect("PORT must be a valid number");

    let listener = tokio::net::TcpListener::bind(("0.0.0.0", port)).await.unwrap();
    tracing::info!("Server running on port {}", port);
    axum::serve(listener, app).await.unwrap();
}
