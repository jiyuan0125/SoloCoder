use crate::alert::AlertManager;
use crate::sliding_window::SlidingWindowLimiter;
use crate::token_bucket::TokenBucketLimiter;
use std::sync::Arc;

#[derive(Clone)]
pub struct AppState {
    pub sliding_window_limiter: Arc<SlidingWindowLimiter>,
    pub token_bucket_limiter: Arc<TokenBucketLimiter>,
    pub alert_manager: Arc<AlertManager>,
}

impl AppState {
    pub fn new() -> Self {
        AppState {
            sliding_window_limiter: Arc::new(SlidingWindowLimiter::new()),
            token_bucket_limiter: Arc::new(TokenBucketLimiter::new()),
            alert_manager: Arc::new(AlertManager::new()),
        }
    }
}

impl Default for AppState {
    fn default() -> Self {
        Self::new()
    }
}
