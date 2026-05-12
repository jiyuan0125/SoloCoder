use std::sync::atomic::{AtomicU64, Ordering};

pub struct StatsSnapshot {
    pub get_requests: u64,
    pub hits: u64,
    pub misses: u64,
    pub source_calls: u64,
}

pub struct CacheStats {
    get_requests: AtomicU64,
    hits: AtomicU64,
    misses: AtomicU64,
    source_calls: AtomicU64,
}

impl CacheStats {
    pub fn new() -> Self {
        Self {
            get_requests: AtomicU64::new(0),
            hits: AtomicU64::new(0),
            misses: AtomicU64::new(0),
            source_calls: AtomicU64::new(0),
        }
    }

    pub fn increment_get_requests(&self) {
        self.get_requests.fetch_add(1, Ordering::Relaxed);
    }

    pub fn increment_hits(&self) {
        self.hits.fetch_add(1, Ordering::Relaxed);
    }

    pub fn increment_misses(&self) {
        self.misses.fetch_add(1, Ordering::Relaxed);
    }

    pub fn increment_source_calls(&self) {
        self.source_calls.fetch_add(1, Ordering::Relaxed);
    }

    pub fn snapshot(&self) -> StatsSnapshot {
        StatsSnapshot {
            get_requests: self.get_requests.load(Ordering::Relaxed),
            hits: self.hits.load(Ordering::Relaxed),
            misses: self.misses.load(Ordering::Relaxed),
            source_calls: self.source_calls.load(Ordering::Relaxed),
        }
    }

    pub fn reset(&self) {
        self.get_requests.store(0, Ordering::Relaxed);
        self.hits.store(0, Ordering::Relaxed);
        self.misses.store(0, Ordering::Relaxed);
        self.source_calls.store(0, Ordering::Relaxed);
    }
}
