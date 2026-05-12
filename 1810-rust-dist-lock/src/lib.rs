use std::collections::{BTreeMap, VecDeque};
use std::sync::Arc;
use std::time::{Duration, Instant};
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use tokio::sync::{Mutex, Notify};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Serialize, Deserialize)]
pub enum Priority {
    Low,
    Normal,
    High,
}

impl Default for Priority {
    fn default() -> Self {
        Priority::Normal
    }
}

impl Priority {
    pub fn from_str(s: &str) -> Self {
        match s.to_ascii_uppercase().as_str() {
            "HIGH" => Priority::High,
            "LOW" => Priority::Low,
            _ => Priority::Normal,
        }
    }
}

#[derive(Debug, Clone, Serialize)]
pub struct LockHolder {
    pub lock_id: Uuid,
    pub priority: Priority,
    pub acquired_at: DateTime<Utc>,
    pub expires_at: DateTime<Utc>,
}

struct Waiter {
    priority: Priority,
    arrival_time: Instant,
    notify: Arc<Notify>,
}

struct WaitQueue {
    queues: BTreeMap<Priority, VecDeque<Waiter>>,
    waiting_count: usize,
}

impl WaitQueue {
    fn new() -> Self {
        let mut queues = BTreeMap::new();
        queues.insert(Priority::High, VecDeque::new());
        queues.insert(Priority::Normal, VecDeque::new());
        queues.insert(Priority::Low, VecDeque::new());
        WaitQueue {
            queues,
            waiting_count: 0,
        }
    }

    fn push(&mut self, priority: Priority) -> Arc<Notify> {
        let notify = Arc::new(Notify::new());
        let waiter = Waiter {
            priority,
            arrival_time: Instant::now(),
            notify: notify.clone(),
        };
        self.queues.get_mut(&priority).unwrap().push_back(waiter);
        self.waiting_count += 1;
        notify
    }

    fn pop_next(&mut self) -> Option<Arc<Notify>> {
        for priority in [Priority::High, Priority::Normal, Priority::Low].iter() {
            if let Some(waiter) = self.queues.get_mut(priority).unwrap().pop_front() {
                self.waiting_count -= 1;
                return Some(waiter.notify);
            }
        }
        None
    }

    fn is_empty(&self) -> bool {
        self.waiting_count == 0
    }

    fn len(&self) -> usize {
        self.waiting_count
    }
}

pub struct LockState {
    holder: Option<LockHolder>,
    wait_queue: WaitQueue,
}

impl LockState {
    fn new() -> Self {
        LockState {
            holder: None,
            wait_queue: WaitQueue::new(),
        }
    }

    fn is_expired(&self) -> bool {
        match &self.holder {
            Some(holder) => Utc::now() >= holder.expires_at,
            None => false,
        }
    }
}

pub struct LockStore {
    locks: Mutex<BTreeMap<String, LockState>>,
}

impl LockStore {
    pub fn new() -> Self {
        LockStore {
            locks: Mutex::new(BTreeMap::new()),
        }
    }

    async fn get_or_create_lock<'a>(
        locks: &'a mut BTreeMap<String, LockState>,
        name: &str,
    ) -> &'a mut LockState {
        if !locks.contains_key(name) {
            locks.insert(name.to_string(), LockState::new());
        }
        locks.get_mut(name).unwrap()
    }

    pub async fn acquire_lock(
        &self,
        name: String,
        ttl_seconds: u64,
        priority: Priority,
        wait_timeout: Duration,
    ) -> Result<LockHolder, String> {
        let start_time = Instant::now();

        loop {
            let elapsed = start_time.elapsed();
            if elapsed >= wait_timeout {
                return Err("Timeout".to_string());
            }

            let mut locks = self.locks.lock().await;
            let lock_state = Self::get_or_create_lock(&mut locks, &name).await;

            if lock_state.holder.is_none() || lock_state.is_expired() {
                let lock_id = Uuid::new_v4();
                let acquired_at = Utc::now();
                let expires_at = acquired_at + chrono::Duration::seconds(ttl_seconds as i64);

                let holder = LockHolder {
                    lock_id,
                    priority,
                    acquired_at,
                    expires_at,
                };

                lock_state.holder = Some(holder.clone());

                if lock_state.wait_queue.len() > 0 {
                    if let Some(next_notify) = lock_state.wait_queue.pop_next() {
                        tokio::spawn(async move {
                            next_notify.notify_one();
                        });
                    }
                }

                return Ok(holder);
            }

            let remaining = wait_timeout.checked_sub(elapsed).unwrap_or(Duration::from_secs(0));
            if remaining.is_zero() {
                return Err("Timeout".to_string());
            }

            let notify = lock_state.wait_queue.push(priority);
            drop(locks);

            tokio::select! {
                _ = notify.notified() => {
                    continue;
                }
                _ = tokio::time::sleep(remaining) => {
                    let mut locks = self.locks.lock().await;
                    if let Some(lock_state) = locks.get_mut(&name) {
                        lock_state.wait_queue.queues.get_mut(&priority).unwrap().retain(|w| {
                            Arc::as_ptr(&w.notify) != Arc::as_ptr(&notify)
                        });
                        lock_state.wait_queue.waiting_count -= 1;
                    }
                    return Err("Timeout".to_string());
                }
            }
        }
    }

    pub async fn release_lock(
        &self,
        name: &str,
        lock_id: Uuid,
    ) -> Result<(), u16> {
        let mut locks = self.locks.lock().await;
        let lock_state = match locks.get_mut(name) {
            Some(state) => state,
            None => return Err(404),
        };

        if lock_state.is_expired() {
            lock_state.holder = None;
            if let Some(next_notify) = lock_state.wait_queue.pop_next() {
                tokio::spawn(async move {
                    next_notify.notify_one();
                });
            }
            return Err(404);
        }

        match &lock_state.holder {
            Some(holder) if holder.lock_id == lock_id => {
                lock_state.holder = None;
                if let Some(next_notify) = lock_state.wait_queue.pop_next() {
                    tokio::spawn(async move {
                        next_notify.notify_one();
                    });
                }
                Ok(())
            }
            Some(_) => Err(403),
            None => Err(404),
        }
    }

    pub async fn force_release_lock(&self, name: &str) -> bool {
        let mut locks = self.locks.lock().await;
        let lock_state = match locks.get_mut(name) {
            Some(state) => state,
            None => return false,
        };

        let had_holder = lock_state.holder.is_some();
        lock_state.holder = None;

        if let Some(next_notify) = lock_state.wait_queue.pop_next() {
            tokio::spawn(async move {
                next_notify.notify_one();
            });
        }

        had_holder
    }

    pub async fn list_active_locks(&self) -> Vec<(String, LockHolder)> {
        let locks = self.locks.lock().await;
        let now = Utc::now();
        let mut result = Vec::new();

        for (name, state) in locks.iter() {
            if let Some(holder) = &state.holder {
                if now < holder.expires_at {
                    result.push((name.clone(), holder.clone()));
                }
            }
        }

        result
    }
}

impl Default for LockStore {
    fn default() -> Self {
        Self::new()
    }
}
