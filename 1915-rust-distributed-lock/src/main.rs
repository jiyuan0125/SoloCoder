mod lock_manager;
mod models;

use std::env;
use std::sync::Arc;
use std::time::Duration;

use axum::extract::{Path, State};
use axum::http::StatusCode;
use axum::response::IntoResponse;
use axum::routing::{delete, get, post, put};
use axum::{Json, Router};
use serde::Serialize;
use tokio::sync::RwLock;
use uuid::Uuid;

use crate::lock_manager::{LockManager, LockManagerRef};
use crate::models::{AcquireRequest, LockStatus};

#[derive(Serialize)]
struct ErrorResponse {
    error: String,
}

async fn acquire_lock(
    State(manager): State<LockManagerRef>,
    Path(resource): Path<String>,
    Json(req): Json<AcquireRequest>,
) -> impl IntoResponse {
    if req.timeout == 0 {
        return (
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                error: "timeout must be greater than 0".to_string(),
            }),
        )
            .into_response();
    }

    let mut mgr = manager.write().await;
    match mgr.acquire(&resource, req.lock_type, req.timeout).await {
        Ok(resp) => (StatusCode::OK, Json(resp)).into_response(),
        Err(LockStatus::ReadLocked) => (
            StatusCode::CONFLICT,
            Json(ErrorResponse {
                error: "resource is read-locked, write lock cannot be acquired".to_string(),
            }),
        )
            .into_response(),
        Err(LockStatus::Expired) => (
            StatusCode::REQUEST_TIMEOUT,
            Json(ErrorResponse {
                error: "lock request expired while waiting".to_string(),
            }),
        )
            .into_response(),
        Err(_) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(ErrorResponse {
                error: "unexpected error".to_string(),
            }),
        )
            .into_response(),
    }
}

#[derive(Serialize)]
struct RenewResponse {
    resource: String,
    lock_id: Uuid,
    expires_at: u64,
}

#[derive(serde::Deserialize)]
struct RenewRequest {
    lock_id: Uuid,
}

async fn renew_lock(
    State(manager): State<LockManagerRef>,
    Path(resource): Path<String>,
    Json(req): Json<RenewRequest>,
) -> impl IntoResponse {
    let mut mgr = manager.write().await;
    match mgr.renew(&resource, req.lock_id) {
        Ok(expires_at) => (
            StatusCode::OK,
            Json(RenewResponse {
                resource,
                lock_id: req.lock_id,
                expires_at,
            }),
        )
            .into_response(),
        Err(msg) => (
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                error: msg.to_string(),
            }),
        )
            .into_response(),
    }
}

async fn get_lock_info(
    State(manager): State<LockManagerRef>,
    Path(resource): Path<String>,
) -> impl IntoResponse {
    let mut mgr = manager.write().await;
    match mgr.get_lock_info(&resource) {
        Some(info) => (StatusCode::OK, Json(info)).into_response(),
        None => (
            StatusCode::NOT_FOUND,
            Json(ErrorResponse {
                error: "resource not found".to_string(),
            }),
        )
            .into_response(),
    }
}

async fn list_locks(State(manager): State<LockManagerRef>) -> impl IntoResponse {
    let mut mgr = manager.write().await;
    let locks = mgr.list_active_locks();
    (StatusCode::OK, Json(locks)).into_response()
}

async fn force_unlock(
    State(manager): State<LockManagerRef>,
    Path(resource): Path<String>,
) -> impl IntoResponse {
    let mut mgr = manager.write().await;
    mgr.force_unlock(&resource);
    (
        StatusCode::OK,
        Json(serde_json::json!({
            "message": "force unlocked",
            "resource": resource
        })),
    )
        .into_response()
}

fn build_router(manager: LockManagerRef) -> Router {
    Router::new()
        .route("/locks", get(list_locks))
        .route("/locks/:resource", post(acquire_lock).get(get_lock_info))
        .route("/locks/:resource/renew", put(renew_lock))
        .route("/locks/:resource/force", delete(force_unlock))
        .with_state(manager)
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();

    let manager: LockManagerRef = Arc::new(RwLock::new(LockManager::new()));

    {
        let manager = manager.clone();
        tokio::spawn(async move {
            let mut interval = tokio::time::interval(Duration::from_secs(1));
            loop {
                interval.tick().await;
                let mut mgr = manager.write().await;
                mgr.cleanup_all_expired();
            }
        });
    }

    let app = build_router(manager);

    let port = env::var("PORT")
        .unwrap_or_else(|_| "3000".to_string())
        .parse::<u16>()
        .expect("PORT must be a valid number");

    let addr = format!("0.0.0.0:{}", port);
    tracing::info!("listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
