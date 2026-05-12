use std::collections::{BTreeMap, HashMap};
use std::sync::Arc;
use std::time::{Duration, Instant, SystemTime};

use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, put},
    Json, Router,
};
use serde::{Deserialize, Serialize};
use tokio::sync::{Mutex, Notify, RwLock};
use uuid::Uuid;

#[derive(Debug, Clone)]
struct LockHolder {
    token: String,
    priority: u32,
    acquired_at: Instant,
    expires_at: Instant,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct LockHolderResponse {
    token: String,
    priority: u32,
    expires_in_seconds: u64,
}

#[derive(Debug, Clone)]
struct Waiter {
    token: String,
    priority: u32,
    notify: Arc<Notify>,
    wait_started: Instant,
}

impl PartialEq for Waiter {
    fn eq(&self, other: &Self) -> bool {
        self.token == other.token
    }
}

impl Eq for Waiter {}

impl PartialOrd for Waiter {
    fn partial_cmp(&self, other: &Self) -> Option<std::cmp::Ordering> {
        Some(self.cmp(other))
    }
}

impl Ord for Waiter {
    fn cmp(&self, other: &Self) -> std::cmp::Ordering {
        other
            .priority
            .cmp(&self.priority)
            .then(self.wait_started.cmp(&other.wait_started))
    }
}

#[derive(Debug, Default)]
struct LockState {
    holder: Option<LockHolder>,
    waiters: BTreeMap<Waiter, ()>,
}

#[derive(Clone, Default)]
struct LockStore {
    locks: Arc<RwLock<HashMap<String, Arc<Mutex<LockState>>>>>,
}

impl LockStore {
    fn new() -> Self {
        Self {
            locks: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    async fn get_or_create_lock(&self, lock_id: &str) -> Arc<Mutex<LockState>> {
        {
            let locks = self.locks.read().await;
            if let Some(lock) = locks.get(lock_id) {
                return lock.clone();
            }
        }

        let mut locks = self.locks.write().await;
        locks
            .entry(lock_id.to_string())
            .or_insert_with(|| Arc::new(Mutex::new(LockState::default())))
            .clone()
    }

    async fn check_expired_locks(&self) {
        let now = Instant::now();
        let locks = self.locks.read().await;
        for (_, lock_mutex) in locks.iter() {
            let mut lock = lock_mutex.lock().await;
            if let Some(ref holder) = lock.holder {
                if holder.expires_at <= now {
                    lock.holder = None;
                    if let Some((waiter, _)) = lock.waiters.pop_first() {
                        waiter.notify.notify_one();
                    }
                }
            }
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct AcquireRequest {
    ttl_seconds: u64,
    priority: Option<u32>,
    wait_timeout_seconds: Option<u64>,
    no_wait: Option<bool>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct AcquireResponse {
    success: bool,
    token: Option<String>,
    lock_id: String,
    holder: Option<LockHolderResponse>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct ReleaseRequest {
    token: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct ReleaseResponse {
    success: bool,
    message: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct RenewRequest {
    token: String,
    ttl_seconds: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct RenewResponse {
    success: bool,
    message: String,
    new_expires_at: Option<i64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct MonitorLockInfo {
    lock_id: String,
    holder: Option<LockHolderInfo>,
    wait_queue_length: usize,
    held_duration_seconds: Option<u64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct LockHolderInfo {
    token: String,
    priority: u32,
    held_duration_seconds: u64,
    expires_in_seconds: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct MonitorResponse {
    total_active_locks: usize,
    total_waiters: usize,
    locks: Vec<MonitorLockInfo>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct ErrorResponse {
    error: String,
}

async fn acquire_lock(
    Path(lock_id): Path<String>,
    State(store): State<LockStore>,
    Json(req): Json<AcquireRequest>,
) -> impl IntoResponse {
    let ttl = Duration::from_secs(req.ttl_seconds);
    let priority = req.priority.unwrap_or(0);
    let no_wait = req.no_wait.unwrap_or(false);
    let wait_timeout = req.wait_timeout_seconds.map(Duration::from_secs);

    let lock_mutex = store.get_or_create_lock(&lock_id).await;
    let token = Uuid::new_v4().to_string();

    loop {
        let mut lock = lock_mutex.lock().await;

        if lock.holder.is_none() {
            let holder = LockHolder {
                token: token.clone(),
                priority,
                acquired_at: Instant::now(),
                expires_at: Instant::now() + ttl,
            };
            let holder_response = LockHolderResponse {
                token: holder.token.clone(),
                priority: holder.priority,
                expires_in_seconds: ttl.as_secs(),
            };
            lock.holder = Some(holder);
            return (
                StatusCode::OK,
                Json(AcquireResponse {
                    success: true,
                    token: Some(token),
                    lock_id: lock_id.clone(),
                    holder: Some(holder_response),
                }),
            );
        }

        if no_wait {
            let holder_response = lock.holder.as_ref().map(|h| LockHolderResponse {
                token: h.token.clone(),
                priority: h.priority,
                expires_in_seconds: h
                    .expires_at
                    .saturating_duration_since(Instant::now())
                    .as_secs(),
            });
            return (
                StatusCode::CONFLICT,
                Json(AcquireResponse {
                    success: false,
                    token: None,
                    lock_id: lock_id.clone(),
                    holder: holder_response,
                }),
            );
        }

        let notify = Arc::new(Notify::new());
        let waiter = Waiter {
            token: token.clone(),
            priority,
            notify: notify.clone(),
            wait_started: Instant::now(),
        };
        lock.waiters.insert(waiter, ());
        drop(lock);

        let wait_future = notify.notified();
        let result = if let Some(timeout) = wait_timeout {
            tokio::time::timeout(timeout, wait_future).await
        } else {
            Ok(())
        };

        match result {
            Ok(()) => continue,
            Err(_) => {
                let mut lock = lock_mutex.lock().await;
                lock.waiters.retain(|w, _| w.token != token);
                let holder_response = lock.holder.as_ref().map(|h| LockHolderResponse {
                    token: h.token.clone(),
                    priority: h.priority,
                    expires_in_seconds: h
                        .expires_at
                        .saturating_duration_since(Instant::now())
                        .as_secs(),
                });
                return (
                    StatusCode::REQUEST_TIMEOUT,
                    Json(AcquireResponse {
                        success: false,
                        token: None,
                        lock_id: lock_id.clone(),
                        holder: holder_response,
                    }),
                );
            }
        }
    }
}

async fn release_lock(
    Path(lock_id): Path<String>,
    State(store): State<LockStore>,
    Query(params): Query<ReleaseRequest>,
) -> impl IntoResponse {
    let locks = store.locks.read().await;
    let Some(lock_mutex) = locks.get(&lock_id).cloned() else {
        return (
            StatusCode::NOT_FOUND,
            Json(ReleaseResponse {
                success: false,
                message: "Lock not found".to_string(),
            }),
        );
    };
    drop(locks);

    let mut lock = lock_mutex.lock().await;
    match &lock.holder {
        Some(holder) if holder.token == params.token => {
            lock.holder = None;
            if let Some((waiter, _)) = lock.waiters.pop_first() {
                waiter.notify.notify_one();
            }
            (
                StatusCode::OK,
                Json(ReleaseResponse {
                    success: true,
                    message: "Lock released".to_string(),
                }),
            )
        }
        Some(_) => (
            StatusCode::FORBIDDEN,
            Json(ReleaseResponse {
                success: false,
                message: "Invalid token".to_string(),
            }),
        ),
        None => (
            StatusCode::NOT_FOUND,
            Json(ReleaseResponse {
                success: false,
                message: "Lock is not held".to_string(),
            }),
        ),
    }
}

async fn renew_lock(
    Path(lock_id): Path<String>,
    State(store): State<LockStore>,
    Json(req): Json<RenewRequest>,
) -> impl IntoResponse {
    let ttl = Duration::from_secs(req.ttl_seconds);
    let locks = store.locks.read().await;
    let Some(lock_mutex) = locks.get(&lock_id).cloned() else {
        return (
            StatusCode::NOT_FOUND,
            Json(RenewResponse {
                success: false,
                message: "Lock not found".to_string(),
                new_expires_at: None,
            }),
        );
    };
    drop(locks);

    let mut lock = lock_mutex.lock().await;
    match &mut lock.holder {
        Some(holder) if holder.token == req.token => {
            holder.expires_at = Instant::now() + ttl;
            let new_expires_at = SystemTime::now()
                .duration_since(SystemTime::UNIX_EPOCH)
                .unwrap()
                .as_secs() as i64
                + req.ttl_seconds as i64;
            (
                StatusCode::OK,
                Json(RenewResponse {
                    success: true,
                    message: "Lock renewed".to_string(),
                    new_expires_at: Some(new_expires_at),
                }),
            )
        }
        Some(_) => (
            StatusCode::FORBIDDEN,
            Json(RenewResponse {
                success: false,
                message: "Invalid token".to_string(),
                new_expires_at: None,
            }),
        ),
        None => (
            StatusCode::NOT_FOUND,
            Json(RenewResponse {
                success: false,
                message: "Lock is not held".to_string(),
                new_expires_at: None,
            }),
        ),
    }
}

async fn get_monitor(State(store): State<LockStore>) -> impl IntoResponse {
    let now = Instant::now();
    let locks = store.locks.read().await;
    let mut monitor_locks = Vec::new();
    let mut total_waiters = 0;
    let mut active_locks = 0;

    for (lock_id, lock_mutex) in locks.iter() {
        let lock = lock_mutex.lock().await;
        let wait_queue_length = lock.waiters.len();
        total_waiters += wait_queue_length;

        let (holder, held_duration_seconds) = if let Some(ref holder) = lock.holder {
            active_locks += 1;
            let held = now.duration_since(holder.acquired_at).as_secs();
            let expires_in = holder
                .expires_at
                .saturating_duration_since(now)
                .as_secs();
            (
                Some(LockHolderInfo {
                    token: holder.token.clone(),
                    priority: holder.priority,
                    held_duration_seconds: held,
                    expires_in_seconds: expires_in,
                }),
                Some(held),
            )
        } else {
            (None, None)
        };

        monitor_locks.push(MonitorLockInfo {
            lock_id: lock_id.clone(),
            holder,
            wait_queue_length,
            held_duration_seconds,
        });
    }

    Json(MonitorResponse {
        total_active_locks: active_locks,
        total_waiters,
        locks: monitor_locks,
    })
}

#[tokio::main]
async fn main() {
    let store = LockStore::new();

    {
        let store_clone = store.clone();
        tokio::spawn(async move {
            let mut interval = tokio::time::interval(Duration::from_secs(1));
            loop {
                interval.tick().await;
                store_clone.check_expired_locks().await;
            }
        });
    }

    let app = Router::new()
        .route("/api/locks/:lock_id/acquire", put(acquire_lock))
        .route("/api/locks/:lock_id/release", put(release_lock))
        .route("/api/locks/:lock_id/renew", put(renew_lock))
        .route("/api/monitor", get(get_monitor))
        .with_state(store);

    let port = std::env::var("PORT")
        .unwrap_or_else(|_| "8080".to_string())
        .parse::<u16>()
        .expect("PORT must be a valid number");

    let addr = std::net::SocketAddr::from(([0, 0, 0, 0], port));
    println!("Distributed Lock Service running on http://{}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
