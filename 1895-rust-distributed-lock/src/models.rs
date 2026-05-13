use serde::{Deserialize, Serialize};
use std::collections::VecDeque;
use std::sync::Arc;
use tokio::sync::Notify;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum LockStatus {
    Free,
    Locked,
    Renewing,
    Expired,
}

impl std::fmt::Display for LockStatus {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            LockStatus::Free => write!(f, "free"),
            LockStatus::Locked => write!(f, "locked"),
            LockStatus::Renewing => write!(f, "renewing"),
            LockStatus::Expired => write!(f, "expired"),
        }
    }
}

#[derive(Debug, Clone)]
pub struct Lock {
    pub name: String,
    pub status: LockStatus,
    pub holder_client_id: Option<String>,
    pub expire_at: Option<chrono::DateTime<chrono::Utc>>,
    pub timeout_seconds: u64,
    pub wait_queue: VecDeque<Waiter>,
}

impl Lock {
    pub fn new(name: String) -> Self {
        Lock {
            name,
            status: LockStatus::Free,
            holder_client_id: None,
            expire_at: None,
            timeout_seconds: 0,
            wait_queue: VecDeque::new(),
        }
    }

    pub fn remaining_time(&self) -> Option<u64> {
        match (self.status, self.expire_at) {
            (LockStatus::Locked | LockStatus::Renewing, Some(expire_at)) => {
                let now = chrono::Utc::now();
                if expire_at > now {
                    Some((expire_at - now).num_seconds() as u64)
                } else {
                    Some(0)
                }
            }
            _ => None,
        }
    }

    pub fn is_holder(&self, client_id: &str) -> bool {
        self.holder_client_id
            .as_ref()
            .map(|h| h == client_id)
            .unwrap_or(false)
    }

    pub fn is_expired(&self) -> bool {
        match (self.status, self.expire_at) {
            (LockStatus::Locked | LockStatus::Renewing, Some(expire_at)) => {
                chrono::Utc::now() >= expire_at
            }
            _ => false,
        }
    }
}

#[derive(Debug, Clone)]
pub struct Waiter {
    pub client_id: String,
    pub timeout_seconds: u64,
    pub notify: Arc<Notify>,
}

impl Waiter {
    pub fn new(client_id: String, timeout_seconds: u64) -> Self {
        Waiter {
            client_id,
            timeout_seconds,
            notify: Arc::new(Notify::new()),
        }
    }
}

#[derive(Debug, Serialize, Deserialize)]
pub struct AcquireRequest {
    pub client_id: String,
    pub timeout_seconds: u64,
    #[serde(default)]
    pub wait_timeout_seconds: u64,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct RenewRequest {
    pub client_id: String,
    pub timeout_seconds: u64,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ReleaseRequest {
    pub client_id: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ForceReleaseRequest {
    pub confirm: bool,
}

#[derive(Debug, Serialize)]
pub struct LockInfoResponse {
    pub name: String,
    pub status: String,
    pub holder_client_id: Option<String>,
    pub remaining_timeout_seconds: Option<u64>,
    pub wait_queue_length: usize,
}

impl From<&Lock> for LockInfoResponse {
    fn from(lock: &Lock) -> Self {
        LockInfoResponse {
            name: lock.name.clone(),
            status: lock.status.to_string(),
            holder_client_id: lock.holder_client_id.clone(),
            remaining_timeout_seconds: lock.remaining_time(),
            wait_queue_length: lock.wait_queue.len(),
        }
    }
}

#[derive(Debug, Serialize)]
pub struct ErrorResponse {
    pub error: String,
    pub message: String,
}
