use crate::models::{Lock, LockStatus, Waiter};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::Mutex;
use tokio::time::{timeout, Duration};

pub const MAX_QUEUE_LENGTH: usize = 20;

#[derive(Clone)]
pub struct LockManager {
    locks: Arc<Mutex<HashMap<String, Lock>>>,
}

impl LockManager {
    pub fn new() -> Self {
        LockManager {
            locks: Arc::new(Mutex::new(HashMap::new())),
        }
    }

    pub async fn try_acquire(
        &self,
        lock_name: &str,
        client_id: &str,
        timeout_seconds: u64,
        wait_timeout_seconds: u64,
    ) -> Result<(), AcquireError> {
        if timeout_seconds == 0 {
            return Err(AcquireError::InvalidTimeout);
        }

        let waiter = {
            let mut locks = self.locks.lock().await;
            let lock_entry = locks
                .entry(lock_name.to_string())
                .or_insert_with(|| Lock::new(lock_name.to_string()));

            if lock_entry.status == LockStatus::Free {
                self.acquire_lock_internal(lock_entry, client_id, timeout_seconds);
                return Ok(());
            }

            if lock_entry.is_expired() {
                self.acquire_lock_internal(lock_entry, client_id, timeout_seconds);
                return Ok(());
            }

            if wait_timeout_seconds == 0 {
                return Err(AcquireError::LockBusy);
            }

            if lock_entry.wait_queue.len() >= MAX_QUEUE_LENGTH {
                return Err(AcquireError::QueueFull);
            }

            let waiter = Waiter::new(client_id.to_string(), timeout_seconds);
            lock_entry.wait_queue.push_back(waiter.clone());
            waiter
        };

        let wait_duration = Duration::from_secs(wait_timeout_seconds);
        let notified = timeout(wait_duration, waiter.notify.notified()).await;

        let mut locks = self.locks.lock().await;
        let lock_entry = locks.get_mut(lock_name).unwrap();

        match notified {
            Ok(_) => {
                if lock_entry
                        .holder_client_id
                        .as_ref()
                        .map(|h| h == client_id)
                        .unwrap_or(false)
                {
                    Ok(())
                } else {
                    lock_entry
                        .wait_queue
                        .retain(|w| w.client_id != client_id);
                    Err(AcquireError::LockBusy)
                }
            }
            Err(_) => {
                lock_entry
                    .wait_queue
                    .retain(|w| w.client_id != client_id);
                Err(AcquireError::WaitTimeout)
            }
        }
    }

    fn acquire_lock_internal(&self, lock: &mut Lock, client_id: &str, timeout_seconds: u64) {
        lock.status = LockStatus::Locked;
        lock.holder_client_id = Some(client_id.to_string());
        lock.timeout_seconds = timeout_seconds;
        lock.expire_at = Some(chrono::Utc::now() + chrono::Duration::seconds(timeout_seconds as i64));
    }

    pub async fn release(
        &self,
        lock_name: &str,
        client_id: &str,
    ) -> Result<(), ReleaseError> {
        let mut locks = self.locks.lock().await;
        let lock_entry = match locks.get_mut(lock_name) {
            Some(l) => l,
            None => return Err(ReleaseError::NotFound),
        };

        if lock_entry.is_expired() {
            return Err(ReleaseError::AlreadyExpired);
        }

        if !lock_entry.is_holder(client_id) {
            return Err(ReleaseError::NotHolder);
        }

        self.release_and_awaken_next(lock_entry);
        Ok(())
    }

    pub async fn force_release(&self, lock_name: &str) -> Result<(), ForceReleaseError> {
        let mut locks = self.locks.lock().await;
        let lock_entry = match locks.get_mut(lock_name) {
            Some(l) => l,
            None => return Err(ForceReleaseError::NotFound),
        };

        if lock_entry.status == LockStatus::Free {
            return Err(ForceReleaseError::AlreadyFree);
        }

        self.release_and_awaken_next(lock_entry);
        Ok(())
    }

    fn release_and_awaken_next(&self, lock: &mut Lock) {
        lock.status = LockStatus::Free;
        lock.holder_client_id = None;
        lock.expire_at = None;
        lock.timeout_seconds = 0;

        while let Some(next_waiter) = lock.wait_queue.pop_front() {
            self.acquire_lock_internal(
                lock,
                &next_waiter.client_id,
                next_waiter.timeout_seconds,
            );
            next_waiter.notify.notify_one();
            return;
        }
    }

    pub async fn renew(
        &self,
        lock_name: &str,
        client_id: &str,
        timeout_seconds: u64,
    ) -> Result<(), RenewError> {
        if timeout_seconds == 0 {
            return Err(RenewError::InvalidTimeout);
        }

        let mut locks = self.locks.lock().await;
        let lock_entry = match locks.get_mut(lock_name) {
            Some(l) => l,
            None => return Err(RenewError::NotFound),
        };

        if !lock_entry.is_holder(client_id) {
            return Err(RenewError::NotHolder);
        }

        if lock_entry.is_expired() {
            lock_entry.status = LockStatus::Expired;
            return Err(RenewError::AlreadyExpired);
        }

        lock_entry.status = LockStatus::Renewing;
        lock_entry.timeout_seconds = timeout_seconds;
        lock_entry.expire_at = Some(chrono::Utc::now() + chrono::Duration::seconds(timeout_seconds as i64));
        Ok(())
    }

    pub async fn get_all_locks(&self) -> Vec<Lock> {
        let locks = self.locks.lock().await;
        locks.values().cloned().collect()
    }

    pub async fn cleanup_expired(&self) {
        let mut locks = self.locks.lock().await;
        let mut expired_locks: Vec<String> = Vec::new();

        for (name, lock) in locks.iter_mut() {
            if lock.is_expired()
                && (lock.status == LockStatus::Locked || lock.status == LockStatus::Renewing)
            {
                expired_locks.push(name.clone());
            }
        }

        for name in expired_locks {
            if let Some(lock_entry) = locks.get_mut(&name) {
                lock_entry.status = LockStatus::Expired;
                self.release_and_awaken_next(lock_entry);
            }
        }
    }
}

#[derive(Debug)]
pub enum AcquireError {
    LockBusy,
    QueueFull,
    WaitTimeout,
    InvalidTimeout,
}

#[derive(Debug)]
pub enum ReleaseError {
    NotFound,
    NotHolder,
    AlreadyExpired,
}

#[derive(Debug)]
pub enum ForceReleaseError {
    NotFound,
    AlreadyFree,
}

#[derive(Debug)]
pub enum RenewError {
    NotFound,
    NotHolder,
    AlreadyExpired,
    InvalidTimeout,
}
