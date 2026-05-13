use axum::extract::{Path, State};
use axum::http::StatusCode;
use axum::Json;

use serde::Deserialize;

use crate::AppState;
use crate::stats::StatsResponse;

#[derive(Deserialize)]
pub struct PutRequest {
    pub value: String,
}

pub async fn get_handler(
    Path(key): Path<String>,
    State(state): State<AppState>,
) -> (StatusCode, String) {
    let cache = state.cache.read().await;
    match cache.get(&key).await {
        Some(value) => (StatusCode::OK, value),
        None => (StatusCode::NOT_FOUND, String::new()),
    }
}

pub async fn put_handler(
    Path(key): Path<String>,
    Json(body): Json<PutRequest>,
    State(state): State<AppState>,
) -> StatusCode {
    let cache = state.cache.read().await;
    cache.put(key, body.value).await;
    StatusCode::NO_CONTENT
}

pub async fn delete_handler(
    Path(key): Path<String>,
    State(state): State<AppState>,
) -> StatusCode {
    let cache = state.cache.read().await;
    cache.delete(&key).await;
    StatusCode::NO_CONTENT
}

pub async fn stats_handler(
    State(state): State<AppState>,
) -> Json<StatsResponse> {
    let cache = state.cache.read().await;
    let (stats, l1_size, l2_size) = cache.get_stats().await;
    Json(StatsResponse::new(stats, l1_size, l2_size))
}
