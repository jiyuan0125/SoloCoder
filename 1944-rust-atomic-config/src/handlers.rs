use axum::extract::{Path, Query, State};
use axum::http::HeaderMap;
use axum::response::IntoResponse;
use axum::Json;
use chrono::{DateTime, Utc};
use serde::Deserialize;
use std::collections::HashMap;
use std::sync::Arc;

use crate::models::{BatchUpdateRequest, SingleUpdateRequest};
use crate::storage::ConfigStore;

#[derive(Debug, Deserialize)]
pub struct AuditQueryParams {
    pub user_id: Option<String>,
    pub start_time: Option<DateTime<Utc>>,
    pub end_time: Option<DateTime<Utc>>,
}

pub type AppState = Arc<ConfigStore>;

fn get_user_id(headers: &HeaderMap) -> String {
    headers
        .get("X-User-Id")
        .and_then(|v| v.to_str().ok())
        .unwrap_or("anonymous")
        .to_string()
}

fn validate_auth(headers: &HeaderMap) -> bool {
    headers.get("X-Auth-Token").is_some()
}

pub async fn batch_update_config(
    State(store): State<AppState>,
    headers: HeaderMap,
    Json(request): Json<BatchUpdateRequest>,
) -> impl IntoResponse {
    let user_id = get_user_id(&headers);
    
    let items: HashMap<String, (String, bool)> = request
        .items
        .into_iter()
        .map(|(k, v)| (k, (v.value, v.is_secret)))
        .collect();

    match store.batch_update(&items, &user_id) {
        Ok(_) => Json(serde_json::json!({"success": true})).into_response(),
        Err(e) => (
            axum::http::StatusCode::INTERNAL_SERVER_ERROR,
            Json(serde_json::json!({"success": false, "error": e})),
        )
            .into_response(),
    }
}

pub async fn update_config(
    State(store): State<AppState>,
    Path(key): Path<String>,
    headers: HeaderMap,
    Json(request): Json<SingleUpdateRequest>,
) -> impl IntoResponse {
    let user_id = get_user_id(&headers);

    match store.single_update(&key, &request.value, request.is_secret, &user_id) {
        Ok(_) => Json(serde_json::json!({"success": true})).into_response(),
        Err(e) => (
            axum::http::StatusCode::INTERNAL_SERVER_ERROR,
            Json(serde_json::json!({"success": false, "error": e})),
        )
            .into_response(),
    }
}

pub async fn list_configs(State(store): State<AppState>) -> impl IntoResponse {
    let configs = store.list();
    Json(configs).into_response()
}

pub async fn get_config(
    State(store): State<AppState>,
    Path(key): Path<String>,
    headers: HeaderMap,
) -> impl IntoResponse {
    if !validate_auth(&headers) {
        return (
            axum::http::StatusCode::UNAUTHORIZED,
            Json(serde_json::json!({"success": false, "error": "缺少 X-Auth-Token"})),
        )
            .into_response();
    }

    match store.get(&key) {
        Ok(Some(config)) => Json(config).into_response(),
        Ok(None) => (
            axum::http::StatusCode::NOT_FOUND,
            Json(serde_json::json!({"success": false, "error": "配置不存在"})),
        )
            .into_response(),
        Err(e) => (
            axum::http::StatusCode::INTERNAL_SERVER_ERROR,
            Json(serde_json::json!({"success": false, "error": e})),
        )
            .into_response(),
    }
}

pub async fn delete_config(
    State(store): State<AppState>,
    Path(key): Path<String>,
    headers: HeaderMap,
) -> impl IntoResponse {
    let user_id = get_user_id(&headers);

    match store.delete(&key, &user_id) {
        Ok(true) => Json(serde_json::json!({"success": true})).into_response(),
        Ok(false) => (
            axum::http::StatusCode::NOT_FOUND,
            Json(serde_json::json!({"success": false, "error": "配置不存在"})),
        )
            .into_response(),
        Err(e) => (
            axum::http::StatusCode::INTERNAL_SERVER_ERROR,
            Json(serde_json::json!({"success": false, "error": e})),
        )
            .into_response(),
    }
}

pub async fn get_audit_logs(
    State(store): State<AppState>,
    Query(params): Query<AuditQueryParams>,
) -> impl IntoResponse {
    let logs = store.query_audit_logs(
        params.user_id.as_deref(),
        params.start_time,
        params.end_time,
    );
    Json(logs).into_response()
}
