use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::{IntoResponse, Json},
    routing::{get, post, delete},
    Router,
};
use std::sync::Arc;
use tracing::{info, warn, error};

use crate::types::{RegisterPoolRequest, AcquireResponse};
use crate::pool_manager::PoolManager;
use crate::connection::PoolError;

#[derive(Debug, Clone)]
pub struct AppState {
    pub pool_manager: Arc<PoolManager>,
}

pub fn create_router(pool_manager: Arc<PoolManager>) -> Router {
    let state = AppState { pool_manager };
    
    Router::new()
        .route("/health", get(health_check))
        .route("/pools", get(list_pools))
        .route("/pools", post(register_pool))
        .route("/pools/:name", get(get_pool_stats))
        .route("/pools/:name", delete(unregister_pool))
        .route("/pools/:name/acquire", post(acquire_connection))
        .with_state(Arc::new(state))
}

async fn health_check() -> impl IntoResponse {
    Json(serde_json::json!({ "status": "ok" }))
}

async fn list_pools(
    State(state): State<Arc<AppState>>,
) -> impl IntoResponse {
    let stats = state.pool_manager.get_all_stats();
    Json(stats)
}

async fn register_pool(
    State(state): State<Arc<AppState>>,
    Json(request): Json<RegisterPoolRequest>,
) -> impl IntoResponse {
    let name = request.name.clone();
    let config = request.into_config();
    
    match state.pool_manager.register_pool(name.clone(), config) {
        Ok(_) => {
            info!("Pool registered: {}", name);
            (
                StatusCode::CREATED,
                Json(serde_json::json!({ "message": format!("Pool '{}' registered successfully", name) }))
            )
        }
        Err(e) => {
            warn!("Failed to register pool {}: {}", name, e);
            (
                StatusCode::CONFLICT,
                Json(serde_json::json!({ "error": e.to_string() }))
            )
        }
    }
}

async fn get_pool_stats(
    State(state): State<Arc<AppState>>,
    Path(name): Path<String>,
) -> impl IntoResponse {
    match state.pool_manager.get_pool_stats(&name) {
        Some(stats) => (
            StatusCode::OK,
            Json(serde_json::to_value(stats).unwrap())
        ),
        None => (
            StatusCode::NOT_FOUND,
            Json(serde_json::json!({ "error": format!("Pool '{}' not found", name) }))
        ),
    }
}

async fn unregister_pool(
    State(state): State<Arc<AppState>>,
    Path(name): Path<String>,
) -> impl IntoResponse {
    match state.pool_manager.unregister_pool(&name).await {
        Ok(_) => {
            info!("Pool unregistered: {}", name);
            (
                StatusCode::OK,
                Json(serde_json::json!({ "message": format!("Pool '{}' unregistered successfully", name) }))
            )
        }
        Err(e) => {
            warn!("Failed to unregister pool {}: {}", name, e);
            (
                StatusCode::NOT_FOUND,
                Json(serde_json::json!({ "error": format!("Pool '{}' not found", name) }))
            )
        }
    }
}

async fn acquire_connection(
    State(state): State<Arc<AppState>>,
    Path(name): Path<String>,
) -> impl IntoResponse {
    match state.pool_manager.acquire(&name).await {
        Ok(conn) => {
            let response = AcquireResponse {
                connection_id: conn.id(),
                pool_name: name.clone(),
            };
            std::mem::forget(conn);
            
            (
                StatusCode::OK,
                Json(serde_json::to_value(response).unwrap())
            )
        }
        Err(PoolError::PoolShuttingDown) | Err(PoolError::PoolClosed) => {
            warn!("Pool {} is shutting down or closed, returning 503", name);
            (
                StatusCode::SERVICE_UNAVAILABLE,
                Json(serde_json::json!({ "error": "Service unavailable" }))
            )
        }
        Err(PoolError::PoolNotFound) => {
            (
                StatusCode::NOT_FOUND,
                Json(serde_json::json!({ "error": format!("Pool '{}' not found", name) }))
            )
        }
        Err(PoolError::AcquireTimeout) => {
            (
                StatusCode::REQUEST_TIMEOUT,
                Json(serde_json::json!({ "error": "Acquire timeout" }))
            )
        }
        Err(e) => {
            error!("Error acquiring connection from pool {}: {}", name, e);
            (
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(serde_json::json!({ "error": e.to_string() }))
            )
        }
    }
}
