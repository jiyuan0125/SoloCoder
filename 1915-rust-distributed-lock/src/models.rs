use serde::{Deserialize, Serialize};
use std::collections::VecDeque;
use std::sync::Arc;
use tokio::sync::Notify;
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum LockType {
    Read,
    Write,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize)]
pub enum LockStatus {
    Idle,
    ReadLocked,
    WriteLocked,
    Expired,
}

#[derive(Debug, Clone)]
pub struct LockHolder {
    pub id: Uuid,
    pub lock_type: LockType,
    pub acquired_at: u64,
    pub expires_at: u64,
}

#[derive(Debug, Clone)]
pub struct WaitingRequest {
    pub id: Uuid,
    pub lock_type: LockType,
    pub timeout_secs: u64,
    pub requested_at: u64,
    pub notify: Arc<Notify>,
}

#[derive(Debug, Clone)]
pub struct ResourceLock {
    pub resource: String,
    pub status: LockStatus,
    pub holders: Vec<LockHolder>,
    pub read_queue: VecDeque<WaitingRequest>,
    pub write_queue: VecDeque<WaitingRequest>,
}

#[derive(Debug, Deserialize)]
pub struct AcquireRequest {
    pub lock_type: LockType,
    pub timeout: u64,
}

#[derive(Debug, Serialize)]
pub struct AcquireResponse {
    pub lock_id: Uuid,
    pub resource: String,
    pub lock_type: LockType,
    pub expires_at: u64,
}

#[derive(Debug, Serialize)]
pub struct LockInfo {
    pub resource: String,
    pub status: LockStatus,
    pub holders: Vec<HolderInfo>,
    pub read_queue_len: usize,
    pub write_queue_len: usize,
    pub ttl_remaining_secs: Option<u64>,
}

#[derive(Debug, Serialize)]
pub struct HolderInfo {
    pub lock_id: Uuid,
    pub lock_type: LockType,
    pub acquired_at: u64,
    pub expires_at: u64,
}

#[derive(Debug, Serialize)]
pub struct ActiveLockInfo {
    pub resource: String,
    pub status: LockStatus,
    pub holders_count: usize,
}
