use std::sync::atomic::{AtomicUsize, Ordering};

#[derive(Default)]
pub struct Stats {
    hits: AtomicUsize,
    misses: AtomicUsize,
    evictions: AtomicUsize,
}

impl Stats {
    pub fn new() -> Self {
        Self {
            hits: AtomicUsize::new(0),
            misses: AtomicUsize::new(0),
            evictions: AtomicUsize::new(0),
        }
    }

    pub fn increment_hit(&self) {
        self.hits.fetch_add(1, Ordering::Relaxed);
    }

    pub fn increment_miss(&self) {
        self.misses.fetch_add(1, Ordering::Relaxed);
    }

    pub fn increment_eviction(&self) {
        self.evictions.fetch_add(1, Ordering::Relaxed);
    }

    pub fn get(&self, current_size: usize) -> (usize, usize, usize, usize) {
        (
            self.hits.load(Ordering::Relaxed),
            self.misses.load(Ordering::Relaxed),
            self.evictions.load(Ordering::Relaxed),
            current_size,
        )
    }

    pub fn hits(&self) -> usize {
        self.hits.load(Ordering::Relaxed)
    }

    pub fn misses(&self) -> usize {
        self.misses.load(Ordering::Relaxed)
    }

    pub fn evictions(&self) -> usize {
        self.evictions.load(Ordering::Relaxed)
    }
}
