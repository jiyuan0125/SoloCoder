use std::collections::{HashMap, VecDeque};
use std::sync::Arc;
use parking_lot::RwLock;
use chrono::{DateTime, Utc};
use serde::Serialize;

const MAX_LOGS_PER_USER: usize = 100;

#[derive(Debug, Clone, Serialize)]
pub struct AccessLogEntry {
    pub user_id: String,
    pub username: String,
    pub method: String,
    pub path: String,
    pub remote_addr: Option<String>,
    pub user_agent: Option<String>,
    pub timestamp: DateTime<Utc>,
    pub status: u16,
}

#[derive(Clone)]
pub struct AccessLogger {
    inner: Arc<RwLock<HashMap<String, VecDeque<AccessLogEntry>>>>,
    max_entries: usize,
}

impl AccessLogger {
    pub fn new() -> Self {
        Self {
            inner: Arc::new(RwLock::new(HashMap::new())),
            max_entries: MAX_LOGS_PER_USER,
        }
    }

    pub fn log(&self, entry: AccessLogEntry) {
        if entry.user_id.is_empty() {
            return;
        }

        let mut map = self.inner.write();
        let queue = map.entry(entry.user_id.clone()).or_insert_with(VecDeque::new);
        queue.push_back(entry);
        if queue.len() > self.max_entries {
            queue.pop_front();
        }
    }

    pub fn get_user_logs(&self, user_id: &str) -> Vec<AccessLogEntry> {
        let map = self.inner.read();
        map.get(user_id)
            .map(|q| q.iter().cloned().collect())
            .unwrap_or_default()
    }

    pub fn all_user_ids(&self) -> Vec<String> {
        let map = self.inner.read();
        map.keys().cloned().collect()
    }
}

impl Default for AccessLogger {
    fn default() -> Self {
        Self::new()
    }
}
