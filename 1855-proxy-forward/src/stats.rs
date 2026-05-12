use serde::Serialize;
use std::sync::atomic::{AtomicU64, Ordering};

#[derive(Debug, Default)]
pub struct BackendStats {
    pub requests: AtomicU64,
    pub total_response_time_ms: AtomicU64,
    pub timeouts: AtomicU64,
}

impl BackendStats {
    pub fn new() -> Self {
        BackendStats::default()
    }

    pub fn record_request(&self, response_time_ms: u64) {
        self.requests.fetch_add(1, Ordering::SeqCst);
        self.total_response_time_ms
            .fetch_add(response_time_ms, Ordering::SeqCst);
    }

    pub fn record_timeout(&self) {
        self.timeouts.fetch_add(1, Ordering::SeqCst);
    }

    pub fn avg_response_time_ms(&self) -> f64 {
        let reqs = self.requests.load(Ordering::SeqCst);
        if reqs == 0 {
            0.0
        } else {
            let total = self.total_response_time_ms.load(Ordering::SeqCst);
            total as f64 / reqs as f64
        }
    }
}

#[derive(Debug, Serialize)]
pub struct StatsSnapshot {
    pub backend_id: String,
    pub path_prefix: String,
    pub requests: u64,
    pub avg_response_time_ms: f64,
    pub timeouts: u64,
}
