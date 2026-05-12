use std::sync::atomic::AtomicU64;

use chrono::{DateTime, Utc};
use parking_lot::RwLock;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum BackendStatus {
    Healthy,
    Unhealthy,
    Draining,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum LoadBalancingStrategy {
    LeastConnections,
    Random,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum SelectionReason {
    LeastConnections,
    Random,
    WarmupLimit,
}

#[derive(Debug, Serialize)]
pub struct BackendStats {
    pub address: String,
    pub weight: u32,
    pub current_connections: u64,
    pub total_requests: u64,
    pub total_failures: u64,
    pub added_at: DateTime<Utc>,
    pub is_warming_up: bool,
    pub status: BackendStatus,
}

#[derive(Debug)]
pub struct Backend {
    pub address: String,
    pub weight: u32,
    pub current_connections: AtomicU64,
    pub total_requests: AtomicU64,
    pub total_failures: AtomicU64,
    pub added_at: DateTime<Utc>,
    pub warmup_end: DateTime<Utc>,
    pub status: RwLock<BackendStatus>,
    pub consecutive_failures: AtomicU64,
    pub consecutive_successes: AtomicU64,
}

#[derive(Debug, Clone, Serialize)]
pub struct AllocationRecord {
    pub timestamp: DateTime<Utc>,
    pub backend_address: String,
    pub strategy: LoadBalancingStrategy,
    pub reason: SelectionReason,
}

#[derive(Debug, Deserialize)]
pub struct AddBackendRequest {
    pub address: String,
    pub weight: Option<u32>,
}
