use std::collections::{HashMap, VecDeque};
use std::sync::Arc;
use std::time::{SystemTime, UNIX_EPOCH};
use tokio::sync::{Notify, RwLock};
use uuid::Uuid;

use crate::models::{
    ActiveLockInfo, AcquireResponse, HolderInfo, LockHolder, LockInfo, LockStatus, LockType,
    ResourceLock, WaitingRequest,
};

pub type LockManagerRef = Arc<RwLock<LockManager>>;

pub struct LockManager {
    locks: HashMap<String, ResourceLock>,
}

impl LockManager {
    pub fn new() -> Self {
        LockManager {
            locks: HashMap::new(),
        }
    }

    fn current_time(&self) -> u64 {
        SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap()
            .as_secs()
    }

    fn ensure_resource(&mut self, resource: &str) {
        self.locks.entry(resource.to_string()).or_insert_with(|| ResourceLock {
            resource: resource.to_string(),
            status: LockStatus::Idle,
            holders: Vec::new(),
            read_queue: VecDeque::new(),
            write_queue: VecDeque::new(),
        });
    }

    fn cleanup_expired(&mut self, resource: &str) {
        let now = self.current_time();
        if let Some(lock) = self.locks.get_mut(resource) {
            lock.holders.retain(|h| h.expires_at > now);

            if lock.holders.is_empty() && lock.status != LockStatus::Idle {
                lock.status = LockStatus::Idle;
            }

            self.try_awake_waiters(resource);
        }
    }

    fn try_awake_waiters(&mut self, resource: &str) {
        let now = self.current_time();
        let lock = match self.locks.get_mut(resource) {
            Some(l) => l,
            None => return,
        };

        lock.read_queue.retain(|w| w.requested_at + w.timeout_secs > now);
        lock.write_queue.retain(|w| w.requested_at + w.timeout_secs > now);

        if lock.status != LockStatus::Idle {
            return;
        }

        if !lock.read_queue.is_empty() {
            let mut holders = Vec::new();
            while let Some(waiter) = lock.read_queue.pop_front() {
                let holder = LockHolder {
                    id: waiter.id,
                    lock_type: LockType::Read,
                    acquired_at: now,
                    expires_at: now + waiter.timeout_secs,
                };
                holders.push(holder.clone());
                waiter.notify.notify_one();
            }
            lock.holders.extend(holders);
            lock.status = LockStatus::ReadLocked;
        } else if let Some(waiter) = lock.write_queue.pop_front() {
            let holder = LockHolder {
                id: waiter.id,
                lock_type: LockType::Write,
                acquired_at: now,
                expires_at: now + waiter.timeout_secs,
            };
            lock.holders.push(holder);
            lock.status = LockStatus::WriteLocked;
            waiter.notify.notify_one();
        }
    }

    pub async fn acquire(
        &mut self,
        resource: &str,
        lock_type: LockType,
        timeout_secs: u64,
    ) -> Result<AcquireResponse, LockStatus> {
        self.ensure_resource(resource);
        self.cleanup_expired(resource);

        let now = self.current_time();
        let lock = self.locks.get_mut(resource).unwrap();

        match lock.status {
            LockStatus::Idle => {
                let holder = LockHolder {
                    id: Uuid::new_v4(),
                    lock_type,
                    acquired_at: now,
                    expires_at: now + timeout_secs,
                };
                lock.holders.push(holder.clone());
                lock.status = match lock_type {
                    LockType::Read => LockStatus::ReadLocked,
                    LockType::Write => LockStatus::WriteLocked,
                };
                Ok(AcquireResponse {
                    lock_id: holder.id,
                    resource: resource.to_string(),
                    lock_type,
                    expires_at: holder.expires_at,
                })
            }
            LockStatus::ReadLocked => match lock_type {
                LockType::Read => {
                    let holder = LockHolder {
                        id: Uuid::new_v4(),
                        lock_type: LockType::Read,
                        acquired_at: now,
                        expires_at: now + timeout_secs,
                    };
                    lock.holders.push(holder.clone());
                    Ok(AcquireResponse {
                        lock_id: holder.id,
                        resource: resource.to_string(),
                        lock_type: LockType::Read,
                        expires_at: holder.expires_at,
                    })
                }
                LockType::Write => Err(LockStatus::ReadLocked),
            },
            LockStatus::WriteLocked | LockStatus::Expired => {
                let notify = Arc::new(Notify::new());
                let req_id = Uuid::new_v4();
                let waiter = WaitingRequest {
                    id: req_id,
                    lock_type,
                    timeout_secs,
                    requested_at: now,
                    notify: notify.clone(),
                };

                match lock_type {
                    LockType::Read => lock.read_queue.push_back(waiter),
                    LockType::Write => lock.write_queue.push_back(waiter),
                }

                let _ = notify.notified().await;

                self.cleanup_expired(resource);
                let lock = self.locks.get_mut(resource).unwrap();

                if let Some(holder) = lock.holders.iter().find(|h| h.id == req_id) {
                    Ok(AcquireResponse {
                        lock_id: holder.id,
                        resource: resource.to_string(),
                        lock_type: holder.lock_type,
                        expires_at: holder.expires_at,
                    })
                } else {
                    Err(LockStatus::Expired)
                }
            }
        }
    }

    pub fn renew(&mut self, resource: &str, lock_id: Uuid) -> Result<u64, &'static str> {
        self.cleanup_expired(resource);
        let now = self.current_time();

        let lock = self
            .locks
            .get_mut(resource)
            .ok_or("resource not found")?;

        let holder = lock
            .holders
            .iter_mut()
            .find(|h| h.id == lock_id)
            .ok_or("lock not found")?;

        if holder.lock_type != LockType::Write {
            return Err("only write locks can be renewed");
        }

        let original_timeout = holder.expires_at - holder.acquired_at;
        holder.expires_at = now + original_timeout;

        Ok(holder.expires_at)
    }

    pub fn force_unlock(&mut self, resource: &str) {
        if let Some(lock) = self.locks.get_mut(resource) {
            lock.holders.clear();
            lock.status = LockStatus::Idle;
            self.try_awake_waiters(resource);
        }
    }

    pub fn get_lock_info(&mut self, resource: &str) -> Option<LockInfo> {
        self.cleanup_expired(resource);
        let now = self.current_time();

        let lock = self.locks.get(resource)?;

        let ttl_remaining = if lock.holders.is_empty() {
            None
        } else {
            lock.holders
                .iter()
                .map(|h| h.expires_at.saturating_sub(now))
                .min()
        };

        Some(LockInfo {
            resource: lock.resource.clone(),
            status: lock.status,
            holders: lock
                .holders
                .iter()
                .map(|h| HolderInfo {
                    lock_id: h.id,
                    lock_type: h.lock_type,
                    acquired_at: h.acquired_at,
                    expires_at: h.expires_at,
                })
                .collect(),
            read_queue_len: lock.read_queue.len(),
            write_queue_len: lock.write_queue.len(),
            ttl_remaining_secs: ttl_remaining,
        })
    }

    pub fn list_active_locks(&mut self) -> Vec<ActiveLockInfo> {
        let resources: Vec<String> = self.locks.keys().cloned().collect();
        for resource in resources {
            self.cleanup_expired(&resource);
        }

        self.locks
            .iter()
            .filter(|(_, lock)| !lock.holders.is_empty() || !lock.read_queue.is_empty() || !lock.write_queue.is_empty())
            .map(|(resource, lock)| ActiveLockInfo {
                resource: resource.clone(),
                status: lock.status,
                holders_count: lock.holders.len(),
            })
            .collect()
    }

    pub fn cleanup_all_expired(&mut self) {
        let resources: Vec<String> = self.locks.keys().cloned().collect();
        for resource in resources {
            self.cleanup_expired(&resource);
        }
    }
}

impl Default for LockManager {
    fn default() -> Self {
        Self::new()
    }
}
