use std::{
    sync::{
        atomic::{AtomicU64, Ordering},
        Arc,
    },
    time::{Duration, Instant},
};

use chrono::{DateTime, Utc};
use dashmap::DashMap;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq, Hash)]
pub enum LimitDimension {
    Ip,
    Path,
    IpPath,
    User,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum LimitAlgorithm {
    FixedWindow,
    SlidingWindow,
    TokenBucket,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LimitRule {
    pub id: String,
    pub dimension: LimitDimension,
    pub algorithm: LimitAlgorithm,
    pub max_requests: u64,
    pub window_seconds: u64,
    pub tokens_per_second: Option<u64>,
    pub bucket_capacity: Option<u64>,
    pub enabled: bool,
}

impl Default for LimitRule {
    fn default() -> Self {
        LimitRule {
            id: "default".to_string(),
            dimension: LimitDimension::IpPath,
            algorithm: LimitAlgorithm::TokenBucket,
            max_requests: 60,
            window_seconds: 60,
            tokens_per_second: Some(1),
            bucket_capacity: Some(10),
            enabled: true,
        }
    }
}

struct WindowState {
    window_start: Instant,
    count: AtomicU64,
}

struct TokenBucketState {
    tokens: AtomicU64,
    last_refill: Instant,
}

#[derive(Debug, Clone, Serialize)]
pub struct LimitCounter {
    pub key: String,
    pub rule_id: String,
    pub allowed: u64,
    pub rejected: u64,
    pub last_updated: DateTime<Utc>,
}

#[derive(Clone)]
pub struct RateLimiter {
    rules: Arc<DashMap<String, LimitRule>>,
    window_states: Arc<DashMap<String, WindowState>>,
    bucket_states: Arc<DashMap<String, TokenBucketState>>,
    counters: Arc<DashMap<String, LimitCounter>>,
}

impl RateLimiter {
    pub fn new() -> Self {
        let rules = Arc::new(DashMap::new());
        rules.insert("default".to_string(), LimitRule::default());
        RateLimiter {
            rules,
            window_states: Arc::new(DashMap::new()),
            bucket_states: Arc::new(DashMap::new()),
            counters: Arc::new(DashMap::new()),
        }
    }

    pub fn upsert_rule(&self, rule: LimitRule) {
        self.rules.insert(rule.id.clone(), rule);
    }

    pub fn get_rule(&self, id: &str) -> Option<LimitRule> {
        self.rules.get(id).map(|r| r.clone())
    }

    pub fn list_rules(&self) -> Vec<LimitRule> {
        self.rules.iter().map(|r| r.clone()).collect()
    }

    pub fn delete_rule(&self, id: &str) -> Option<LimitRule> {
        self.rules.remove(id).map(|(_, r)| r)
    }

    pub fn list_counters(&self) -> Vec<LimitCounter> {
        self.counters
            .iter()
            .map(|entry| LimitCounter {
                key: entry.key().clone(),
                rule_id: entry.rule_id.clone(),
                allowed: entry.allowed,
                rejected: entry.rejected,
                last_updated: entry.last_updated,
            })
            .collect()
    }

    pub fn check(&self, key: &str) -> bool {
        let rules: Vec<LimitRule> = self.rules.iter().filter(|r| r.enabled).map(|r| r.clone()).collect();
        for rule in rules {
            let scoped_key = format!("{}:{}", rule.id, key);
            let allowed = match rule.algorithm {
                LimitAlgorithm::FixedWindow => self.check_fixed_window(&scoped_key, &rule),
                LimitAlgorithm::SlidingWindow => self.check_sliding_window(&scoped_key, &rule),
                LimitAlgorithm::TokenBucket => self.check_token_bucket(&scoped_key, &rule),
            };
            self.record(&scoped_key, &rule.id, allowed);
            if !allowed {
                return false;
            }
        }
        true
    }

    fn record(&self, key: &str, rule_id: &str, allowed: bool) {
        let now = Utc::now();
        self.counters
            .entry(key.to_string())
            .and_modify(|c| {
                c.last_updated = now;
                if allowed {
                    c.allowed += 1;
                } else {
                    c.rejected += 1;
                }
            })
            .or_insert_with(|| LimitCounter {
                key: key.to_string(),
                rule_id: rule_id.to_string(),
                allowed: if allowed { 1 } else { 0 },
                rejected: if allowed { 0 } else { 1 },
                last_updated: now,
            });
    }

    fn check_fixed_window(&self, key: &str, rule: &LimitRule) -> bool {
        let now = Instant::now();
        let window_dur = Duration::from_secs(rule.window_seconds);
        let mut entry = self
            .window_states
            .entry(key.to_string())
            .or_insert_with(|| WindowState {
                window_start: now,
                count: AtomicU64::new(0),
            });
        if now.duration_since(entry.window_start) >= window_dur {
            entry.window_start = now;
            entry.count.store(0, Ordering::SeqCst);
        }
        let current = entry.count.fetch_add(1, Ordering::SeqCst);
        current < rule.max_requests
    }

    fn check_sliding_window(&self, key: &str, rule: &LimitRule) -> bool {
        self.check_fixed_window(key, rule)
    }

    fn check_token_bucket(&self, key: &str, rule: &LimitRule) -> bool {
        let per_sec = rule.tokens_per_second.unwrap_or(1);
        let capacity = rule.bucket_capacity.unwrap_or(rule.max_requests);
        let now = Instant::now();
        let mut entry = self
            .bucket_states
            .entry(key.to_string())
            .or_insert_with(|| TokenBucketState {
                tokens: AtomicU64::new(capacity),
                last_refill: now,
            });
        let elapsed = now.duration_since(entry.last_refill).as_secs();
        if elapsed > 0 {
            let refill = per_sec.saturating_mul(elapsed).min(capacity);
            let current = entry.tokens.load(Ordering::SeqCst);
            let new_tokens = current.saturating_add(refill).min(capacity);
            entry.tokens.store(new_tokens, Ordering::SeqCst);
            entry.last_refill = now;
        }
        let current = entry.tokens.fetch_sub(1, Ordering::SeqCst);
        current > 0
    }
}

impl Default for RateLimiter {
    fn default() -> Self {
        Self::new()
    }
}
