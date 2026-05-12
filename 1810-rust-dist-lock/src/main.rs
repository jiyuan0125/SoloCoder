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
use uuid::Uuid;
use dist_lock_service::{LockStore, Priority};

#[derive(Debug, Deserialize)]
struct AcquireLockRequest {
    ttl_seconds: u64,
    #[serde(default)]
    priority: Option<String>,
}

#[derive(Debug, Serialize)]
struct AcquireLockResponse {
    lock_id: Uuid,
    acquired_at: String,
    expires_at: String,
    priority: String,
}

#[derive(Debug, Serialize)]
struct LockInfo {
    name: String,
    lock_id: Uuid,
    priority: String,
    acquired_at: String,
    expires_at: String,
}

#[derive(Debug, Serialize)]
struct ListLocksResponse {
    locks: Vec<LockInfo>,
}

#[derive(Debug, Serialize)]
struct ErrorResponse {
    error: String,
}

type AppState = Arc<LockStore>;

fn priority_to_string(p: Priority) -> String {
    match p {
        Priority::High => "HIGH".to_string(),
        Priority::Normal => "NORMAL".to_string(),
        Priority::Low => "LOW".to_string(),
    }
}

async fn acquire_lock(
    State(store): State<AppState>,
    Path(name): Path<String>,
    Json(req): Json<AcquireLockRequest>,
) -> impl IntoResponse {
    let priority_str = req.priority.as_deref().unwrap_or("NORMAL");
    let priority = Priority::from_str(priority_str);
    
    let wait_timeout = Duration::from_secs(30);

    match store.acquire_lock(name, req.ttl_seconds, priority, wait_timeout).await {
        Ok(holder) => {
            let response = AcquireLockResponse {
                lock_id: holder.lock_id,
                acquired_at: holder.acquired_at.to_rfc3339(),
                expires_at: holder.expires_at.to_rfc3339(),
                priority: priority_to_string(holder.priority),
            };
            (StatusCode::OK, Json(response)).into_response()
        }
        Err(_) => (
            StatusCode::REQUEST_TIMEOUT,
            Json(ErrorResponse {
                error: "Timeout waiting for lock".to_string(),
            }),
        )
            .into_response(),
    }
}

async fn release_lock(
    State(store): State<AppState>,
    Path((name, lock_id)): Path<(String, String)>,
) -> impl IntoResponse {
    let lock_id = match Uuid::parse_str(&lock_id) {
        Ok(id) => id,
        Err(_) => {
            return (
                StatusCode::BAD_REQUEST,
                Json(ErrorResponse {
                    error: "Invalid lock_id format".to_string(),
                }),
            )
                .into_response();
        }
    };

    match store.release_lock(&name, lock_id).await {
        Ok(()) => StatusCode::NO_CONTENT.into_response(),
        Err(403) => (
            StatusCode::FORBIDDEN,
            Json(ErrorResponse {
                error: "Lock ID mismatch".to_string(),
            }),
        )
            .into_response(),
        Err(404) => (
            StatusCode::NOT_FOUND,
            Json(ErrorResponse {
                error: "Lock not found or already expired".to_string(),
            }),
        )
            .into_response(),
        Err(_) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(ErrorResponse {
                error: "Internal server error".to_string(),
            }),
        )
            .into_response(),
    }
}

async fn force_release_lock(
    State(store): State<AppState>,
    Path(name): Path<String>,
) -> impl IntoResponse {
    if store.force_release_lock(&name).await {
        StatusCode::NO_CONTENT.into_response()
    } else {
        (
            StatusCode::NOT_FOUND,
            Json(ErrorResponse {
                error: "Lock not found".to_string(),
            }),
        )
            .into_response()
    }
}

async fn list_locks(State(store): State<AppState>) -> impl IntoResponse {
    let active_locks = store.list_active_locks().await;
    let locks: Vec<LockInfo> = active_locks
        .into_iter()
        .map(|(name, holder)| LockInfo {
            name,
            lock_id: holder.lock_id,
            priority: priority_to_string(holder.priority),
            acquired_at: holder.acquired_at.to_rfc3339(),
            expires_at: holder.expires_at.to_rfc3339(),
        })
        .collect();

    (
        StatusCode::OK, Json(ListLocksResponse { locks })).into_response()
}

#[tokio::main]
async fn main() {
    let store: AppState = Arc::new(LockStore::new());

    let app = Router::new()
        .route("/locks", get(list_locks))
        .route("/locks/:name", post(acquire_lock))
        .route("/locks/:name/:lock_id", delete(release_lock))
        .route("/locks/:name/force", delete(force_release_lock))
        .with_state(store);

    let port = std::env::var("PORT")
        .unwrap_or_else(|_| "8080".to_string())
        .parse::<u16>()
        .expect("Invalid PORT");

    let addr = std::net::SocketAddr::from(([0, 0, 0, 0], port));
    let listener = tokio::net::TcpListener::bind(addr).await.expect("Failed to bind");

    println!("Lock service listening on {}", addr);

    axum::serve(listener, app).await.expect("Server failed to start");
}
