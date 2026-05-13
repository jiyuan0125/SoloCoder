use dashmap::DashMap;
use parking_lot::Mutex;
use uuid::Uuid;
use chrono::{DateTime, Utc};
use std::collections::VecDeque;
use std::sync::Arc;

use crate::models::{Task, TaskStatus};

#[derive(Clone, Default)]
pub struct TaskStore {
    pub tasks: Arc<DashMap<Uuid, Task>>,
    pub idempotency_map: Arc<DashMap<String, Uuid>>,
    pub pending_queue: Arc<Mutex<VecDeque<Uuid>>>,
    pub retry_queue: Arc<Mutex<VecDeque<Uuid>>>,
}

impl TaskStore {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn insert(&self, task: Task) -> Uuid {
        let task_id = task.task_id;
        let idempotency_key = task.idempotency_key.clone();
        self.tasks.insert(task_id, task);
        self.idempotency_map.insert(idempotency_key, task_id);
        self.pending_queue.lock().push_back(task_id);
        task_id
    }

    pub fn insert_terminal(&self, task: Task) -> Uuid {
        let task_id = task.task_id;
        let idempotency_key = task.idempotency_key.clone();
        self.tasks.insert(task_id, task);
        self.idempotency_map.insert(idempotency_key, task_id);
        task_id
    }

    pub fn get(&self, task_id: &Uuid) -> Option<Task> {
        self.tasks.get(task_id).map(|t| t.clone())
    }

    pub fn get_by_idempotency(&self, key: &str) -> Option<Task> {
        self.idempotency_map
            .get(key)
            .and_then(|id| self.tasks.get(&id).map(|t| t.clone()))
    }

    pub fn update<F>(&self, task_id: &Uuid, f: F)
    where
        F: FnOnce(&mut Task),
    {
        if let Some(mut entry) = self.tasks.get_mut(task_id) {
            f(&mut entry);
        }
    }

    pub fn pop_pending(&self) -> Option<Uuid> {
        self.pending_queue.lock().pop_front()
    }

    pub fn push_retry(&self, task_id: Uuid) {
        self.retry_queue.lock().push_back(task_id);
    }

    pub fn pop_retry(&self) -> Option<Uuid> {
        self.retry_queue.lock().pop_front()
    }

    pub fn cleanup_expired(&self, cutoff: DateTime<Utc>) -> usize {
        let mut removed = 0;
        self.tasks.retain(|_, task| {
            let keep = if let Some(completed_at) = task.completed_at {
                completed_at > cutoff
            } else {
                true
            };
            if !keep {
                removed += 1;
            }
            keep
        });
        if removed > 0 {
            self.idempotency_map.retain(|_, task_id| self.tasks.contains_key(task_id));
        }
        removed
    }

    pub fn all_tasks(&self) -> Vec<Task> {
        self.tasks.iter().map(|t| t.clone()).collect()
    }

    pub fn count_by_status(&self) -> (u64, u64, u64, u64, u64) {
        let mut pending = 0u64;
        let mut running = 0u64;
        let mut retrying = 0u64;
        let mut succeeded = 0u64;
        let mut failed = 0u64;
        for task in self.tasks.iter() {
            match task.status {
                TaskStatus::Pending => pending += 1,
                TaskStatus::Running => running += 1,
                TaskStatus::Retrying => retrying += 1,
                TaskStatus::Succeeded => succeeded += 1,
                TaskStatus::Failed => failed += 1,
            }
        }
        (pending, running, retrying, succeeded, failed)
    }
}
