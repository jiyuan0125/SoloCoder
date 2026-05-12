use std::collections::HashMap;
use std::time::{SystemTime, UNIX_EPOCH};
use parking_lot::RwLock;

#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub struct TokenBucketConfig {
    pub capacity: u64,
    pub fill_rate: u64,
}

impl Default for TokenBucketConfig {
    fn default() -> Self {
        TokenBucketConfig {
            capacity: 100,
            fill_rate: 10,
        }
    }
}

struct TokenBucket {
    capacity: u64,
    fill_rate: u64,
    tokens: u64,
    last_fill_time: u64,
}

impl TokenBucket {
    fn new(config: &TokenBucketConfig) -> Self {
        TokenBucket {
            capacity: config.capacity,
            fill_rate: config.fill_rate,
            tokens: config.capacity,
            last_fill_time: Self::current_second(),
        }
    }

    fn current_second() -> u64 {
        SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap()
            .as_secs()
    }

    fn update_config(&mut self, config: &TokenBucketConfig) {
        self.capacity = config.capacity;
        self.fill_rate = config.fill_rate;
        if self.tokens > self.capacity {
            self.tokens = self.capacity;
        }
    }

    fn fill(&mut self) {
        let now = Self::current_second();
        let elapsed = now.saturating_sub(self.last_fill_time);
        if elapsed > 0 {
            let new_tokens = elapsed.saturating_mul(self.fill_rate);
            self.tokens = std::cmp::min(self.capacity, self.tokens.saturating_add(new_tokens));
            self.last_fill_time = now;
        }
    }

    fn try_acquire(&mut self, tokens: u64) -> bool {
        self.fill();
        if self.tokens >= tokens {
            self.tokens -= tokens;
            true
        } else {
            false
        }
    }
}

pub struct TokenBucketLimiter {
    buckets: RwLock<HashMap<String, (TokenBucket, TokenBucketConfig)>>,
}

impl TokenBucketLimiter {
    pub fn new() -> Self {
        TokenBucketLimiter {
            buckets: RwLock::new(HashMap::new()),
        }
    }

    pub fn set_rule(&self, path: &str, config: TokenBucketConfig) {
        let mut buckets = self.buckets.write();
        match buckets.get_mut(path) {
            Some((bucket, stored_config)) => {
                *stored_config = config.clone();
                bucket.update_config(&config);
            }
            None => {
                buckets.insert(path.to_string(), (TokenBucket::new(&config), config));
            }
        }
    }

    pub fn remove_rule(&self, path: &str) {
        self.buckets.write().remove(path);
    }

    pub fn get_rule(&self, path: &str) -> Option<TokenBucketConfig> {
        self.buckets.read().get(path).map(|(_, config)| config.clone())
    }

    pub fn get_all_rules(&self) -> HashMap<String, TokenBucketConfig> {
        self.buckets
            .read()
            .iter()
            .map(|(path, (_, config))| (path.clone(), config.clone()))
            .collect()
    }

    pub fn try_acquire(&self, path: &str, tokens: u64) -> bool {
        let mut buckets = self.buckets.write();
        match buckets.get_mut(path) {
            Some((bucket, _)) => bucket.try_acquire(tokens),
            None => true,
        }
    }
}

impl Default for TokenBucketLimiter {
    fn default() -> Self {
        Self::new()
    }
}
