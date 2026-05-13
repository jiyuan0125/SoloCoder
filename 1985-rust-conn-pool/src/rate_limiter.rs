use std::time::{Duration, Instant};
use parking_lot::Mutex;
use std::sync::Arc;

#[derive(Debug, Clone)]
pub struct RateLimiter {
    inner: Arc<Mutex<RateLimiterInner>>,
}

#[derive(Debug)]
struct RateLimiterInner {
    max_per_second: u32,
    tokens: u32,
    last_refill: Instant,
}

impl RateLimiter {
    pub fn new(max_per_second: u32) -> Self {
        Self {
            inner: Arc::new(Mutex::new(RateLimiterInner {
                max_per_second,
                tokens: max_per_second,
                last_refill: Instant::now(),
            })),
        }
    }
    
    pub fn try_acquire(&self) -> bool {
        let mut inner = self.inner.lock();
        let now = Instant::now();
        let elapsed = now.duration_since(inner.last_refill);
        
        if elapsed >= Duration::from_secs(1) {
            inner.tokens = inner.max_per_second;
            inner.last_refill = now;
        }
        
        if inner.tokens > 0 {
            inner.tokens -= 1;
            true
        } else {
            false
        }
    }
}
