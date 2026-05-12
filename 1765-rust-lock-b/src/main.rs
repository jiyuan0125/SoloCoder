use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, SystemTime, UNIX_EPOCH};

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::{IntoResponse, Json},
    routing::{delete, get, post},
    Router,
};
use serde::{Deserialize, Serialize};
use tokio::sync::RwLock;
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
struct LockInfo {
    name: String,
    holder_id: Uuid,
    acquired_at: u64,
    expires_at: u64,
    ttl_seconds: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct AcquireLockRequest {
    holder_id: Option<Uuid>,
    ttl_seconds: Option<u64>,
    timeout_seconds: Option<u64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct AcquireLockResponse {
    success: bool,
    lock: Option<LockInfo>,
    message: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct ReleaseLockRequest {
    holder_id: Uuid,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct ReleaseLockResponse {
    success: bool,
    message: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct LockStatusResponse {
    exists: bool,
    lock: Option<LockInfo>,
}

type LockStore = Arc<RwLock<HashMap<String, LockInfo>>>;

fn current_time_seconds() -> u64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap_or_default()
        .as_secs()
}

fn is_lock_expired(lock: &LockInfo) -> bool {
    current_time_seconds() >= lock.expires_at
}

async fn acquire_lock(
    Path(lock_name): Path<String>,
    State(store): State<LockStore>,
    Json(request): Json<AcquireLockRequest>,
) -> impl IntoResponse {
    let holder_id = request.holder_id.unwrap_or_else(Uuid::new_v4);
    let ttl_seconds = request.ttl_seconds.unwrap_or(60);
    let timeout_seconds = request.timeout_seconds.unwrap_or(10);

    if lock_name.is_empty() {
        return (
            StatusCode::BAD_REQUEST,
            Json(AcquireLockResponse {
                success: false,
                lock: None,
                message: "Lock name cannot be empty".to_string(),
            }),
        );
    }

    if ttl_seconds == 0 {
        return (
            StatusCode::BAD_REQUEST,
            Json(AcquireLockResponse {
                success: false,
                lock: None,
                message: "TTL must be greater than 0".to_string(),
            }),
        );
    }

    let start_time = current_time_seconds();
    let deadline = start_time + timeout_seconds;

    loop {
        let now = current_time_seconds();
        if now >= deadline {
            return (
                StatusCode::CONFLICT,
                Json(AcquireLockResponse {
                    success: false,
                    lock: None,
                    message: "Failed to acquire lock: timeout".to_string(),
                }),
            );
        }

        let mut store_write = store.write().await;

        if let Some(existing_lock) = store_write.get(&lock_name) {
            if existing_lock.holder_id == holder_id {
                let updated_lock = LockInfo {
                    name: lock_name.clone(),
                    holder_id,
                    acquired_at: now,
                    expires_at: now + ttl_seconds,
                    ttl_seconds,
                };
                store_write.insert(lock_name.clone(), updated_lock.clone());
                return (
                    StatusCode::OK,
                    Json(AcquireLockResponse {
                        success: true,
                        lock: Some(updated_lock),
                        message: "Lock renewed successfully".to_string(),
                    }),
                );
            }

            if is_lock_expired(existing_lock) {
                let new_lock = LockInfo {
                    name: lock_name.clone(),
                    holder_id,
                    acquired_at: now,
                    expires_at: now + ttl_seconds,
                    ttl_seconds,
                };
                store_write.insert(lock_name.clone(), new_lock.clone());
                return (
                    StatusCode::OK,
                    Json(AcquireLockResponse {
                        success: true,
                        lock: Some(new_lock),
                        message: "Lock acquired successfully (expired lock replaced)".to_string(),
                    }),
                );
            }

            drop(store_write);
            tokio::time::sleep(Duration::from_millis(100)).await;
        } else {
            let new_lock = LockInfo {
                name: lock_name.clone(),
                holder_id,
                acquired_at: now,
                expires_at: now + ttl_seconds,
                ttl_seconds,
            };
            store_write.insert(lock_name.clone(), new_lock.clone());
            return (
                StatusCode::OK,
                Json(AcquireLockResponse {
                    success: true,
                    lock: Some(new_lock),
                    message: "Lock acquired successfully".to_string(),
                }),
            );
        }
    }
}

async fn release_lock(
    Path(lock_name): Path<String>,
    State(store): State<LockStore>,
    Json(request): Json<ReleaseLockRequest>,
) -> impl IntoResponse {
    let mut store_write = store.write().await;

    match store_write.get(&lock_name) {
        Some(existing_lock) => {
            if existing_lock.holder_id == request.holder_id {
                store_write.remove(&lock_name);
                (
                    StatusCode::OK,
                    Json(ReleaseLockResponse {
                        success: true,
                        message: "Lock released successfully".to_string(),
                    }),
                )
            } else {
                (
                    StatusCode::FORBIDDEN,
                    Json(ReleaseLockResponse {
                        success: false,
                        message: "Cannot release lock: not the holder".to_string(),
                    }),
                )
            }
        }
        None => (
            StatusCode::NOT_FOUND,
            Json(ReleaseLockResponse {
                success: false,
                message: "Lock not found".to_string(),
            }),
        ),
    }
}

async fn get_lock_status(
    Path(lock_name): Path<String>,
    State(store): State<LockStore>,
) -> impl IntoResponse {
    let store_read = store.read().await;

    match store_read.get(&lock_name) {
        Some(lock) => {
            if is_lock_expired(lock) {
                (
                    StatusCode::OK,
                    Json(LockStatusResponse {
                        exists: false,
                        lock: None,
                    }),
                )
            } else {
                (
                    StatusCode::OK,
                    Json(LockStatusResponse {
                        exists: true,
                        lock: Some(lock.clone()),
                    }),
                )
            }
        }
        None => (
            StatusCode::OK,
            Json(LockStatusResponse {
                exists: false,
                lock: None,
            }),
        ),
    }
}

async fn get_all_locks(State(store): State<LockStore>) -> impl IntoResponse {
    let store_read = store.read().await;
    let now = current_time_seconds();

    let active_locks: Vec<LockInfo> = store_read
        .values()
        .filter(|lock| now < lock.expires_at)
        .cloned()
        .collect();

    Json(active_locks)
}

async fn health_check() -> impl IntoResponse {
    Json(serde_json::json!({
        "status": "healthy",
        "timestamp": current_time_seconds()
    }))
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_env_filter(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "info".into()),
        )
        .init();

    let store: LockStore = Arc::new(RwLock::new(HashMap::new()));

    let app = Router::new()
        .route("/health", get(health_check))
        .route("/locks", get(get_all_locks))
        .route("/locks/:name", post(acquire_lock))
        .route("/locks/:name", get(get_lock_status))
        .route("/locks/:name", delete(release_lock))
        .with_state(store);

    let port = std::env::var("PORT")
        .unwrap_or_else(|_| "3000".to_string())
        .parse::<u16>()
        .unwrap_or(3000);

    let addr = std::net::SocketAddr::from(([0, 0, 0, 0], port));
    tracing::info!("Lock service starting on {}", addr);

    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}
