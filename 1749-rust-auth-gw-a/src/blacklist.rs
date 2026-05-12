use std::collections::HashMap;
use std::sync::Arc;
use parking_lot::RwLock;
use chrono::Utc;
use tokio::time::{interval, Duration};

#[derive(Clone)]
pub struct TokenBlacklist {
    inner: Arc<RwLock<HashMap<String, i64>>>,
}

impl TokenBlacklist {
    pub fn new() -> Self {
        Self {
            inner: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    pub fn add(&self, jti: String, exp: i64) {
        let mut map = self.inner.write();
        map.insert(jti, exp);
    }

    pub fn contains(&self, jti: &str) -> bool {
        let map = self.inner.read();
        map.contains_key(jti)
    }

    pub fn cleanup_expired(&self) -> usize {
        let now = Utc::now().timestamp();
        let mut map = self.inner.write();
        let before = map.len();
        map.retain(|_, exp| *exp > now);
        let after = map.len();
        before - after
    }

    pub fn len(&self) -> usize {
        self.inner.read().len()
    }

    pub fn is_empty(&self) -> bool {
        self.inner.read().is_empty()
    }
}

impl Default for TokenBlacklist {
    fn default() -> Self {
        Self::new()
    }
}

pub fn start_cleanup_task(blacklist: TokenBlacklist, interval_secs: u64) {
    tokio::spawn(async move {
        let mut ticker = interval(Duration::from_secs(interval_secs));
        loop {
            ticker.tick().await;
            let removed = blacklist.cleanup_expired();
            if removed > 0 {
                tracing::info!("清理了 {} 个过期的黑名单条目", removed);
            }
        }
    });
}
