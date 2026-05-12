use std::sync::atomic::{AtomicU64, Ordering};

#[derive(Debug, Default)]
pub struct CacheStats {
    hits: AtomicU64,
    misses: AtomicU64,
    evictions: AtomicU64,
    sets: AtomicU64,
    deletes: AtomicU64,
    expires_removed: AtomicU64,
}

impl CacheStats {
    pub fn record_hit(&self) {
        self.hits.fetch_add(1, Ordering::Relaxed);
    }

    pub fn record_miss(&self) {
        self.misses.fetch_add(1, Ordering::Relaxed);
    }

    pub fn record_eviction(&self) {
        self.evictions.fetch_add(1, Ordering::Relaxed);
    }

    pub fn record_set(&self) {
        self.sets.fetch_add(1, Ordering::Relaxed);
    }

    pub fn record_delete(&self) {
        self.deletes.fetch_add(1, Ordering::Relaxed);
    }

    pub fn record_expire_removed(&self) {
        self.expires_removed.fetch_add(1, Ordering::Relaxed);
    }

    pub fn get_hits(&self) -> u64 {
        self.hits.load(Ordering::Relaxed)
    }

    pub fn get_misses(&self) -> u64 {
        self.misses.load(Ordering::Relaxed)
    }

    pub fn get_evictions(&self) -> u64 {
        self.evictions.load(Ordering::Relaxed)
    }

    pub fn get_sets(&self) -> u64 {
        self.sets.load(Ordering::Relaxed)
    }

    pub fn get_deletes(&self) -> u64 {
        self.deletes.load(Ordering::Relaxed)
    }

    pub fn get_expires_removed(&self) -> u64 {
        self.expires_removed.load(Ordering::Relaxed)
    }

    pub fn get_total_requests(&self) -> u64 {
        self.get_hits() + self.get_misses()
    }

    pub fn get_hit_rate(&self) -> f64 {
        let total = self.get_total_requests();
        if total == 0 {
            0.0
        } else {
            self.get_hits() as f64 / total as f64
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_stats_basic() {
        let stats = CacheStats::default();
        
        stats.record_hit();
        stats.record_hit();
        stats.record_miss();
        
        assert_eq!(stats.get_hits(), 2);
        assert_eq!(stats.get_misses(), 1);
        assert_eq!(stats.get_total_requests(), 3);
        assert!((stats.get_hit_rate() - 2.0/3.0).abs() < 0.0001);
    }
}
