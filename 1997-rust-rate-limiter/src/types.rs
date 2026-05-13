use serde::{Deserialize, Serialize};
use std::collections::BTreeMap;
use std::net::IpAddr;
use std::time::{Duration, Instant};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum AlgorithmType {
    FixedWindow,
    SlidingWindow,
    TokenBucket,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum PathPatternType {
    Exact,
    Prefix,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RuleConfig {
    pub id: Option<Uuid>,
    pub algorithm: AlgorithmType,
    pub path_pattern: String,
    pub path_type: PathPatternType,
    pub per_ip: bool,
    pub limit: u64,
    pub window_secs: Option<u64>,
    pub capacity: Option<u64>,
    pub refill_rate: Option<u64>,
    pub refill_interval_ms: Option<u64>,
}

#[derive(Debug, Clone)]
pub struct RateLimitRule {
    pub id: Uuid,
    pub algorithm: AlgorithmType,
    pub path_pattern: String,
    pub path_type: PathPatternType,
    pub per_ip: bool,
    pub limit: u64,
    pub window_secs: u64,
    pub capacity: u64,
    pub refill_rate: u64,
    pub refill_interval: Duration,
    pub created_at: Instant,
}

impl From<RuleConfig> for RateLimitRule {
    fn from(config: RuleConfig) -> Self {
        let id = config.id.unwrap_or_else(Uuid::new_v4);
        RateLimitRule {
            id,
            algorithm: config.algorithm,
            path_pattern: config.path_pattern,
            path_type: config.path_type,
            per_ip: config.per_ip,
            limit: config.limit,
            window_secs: config.window_secs.unwrap_or(60),
            capacity: config.capacity.unwrap_or(config.limit),
            refill_rate: config.refill_rate.unwrap_or(1),
            refill_interval: Duration::from_millis(config.refill_interval_ms.unwrap_or(1000)),
            created_at: Instant::now(),
        }
    }
}

#[derive(Debug, Clone)]
pub struct FixedWindowCounter {
    pub count: u64,
    pub window_start: Instant,
}

#[derive(Debug, Clone)]
pub struct SlidingWindowCounter {
    pub requests: BTreeMap<u64, u64>,
    pub total: u64,
}

#[derive(Debug, Clone)]
pub struct TokenBucketState {
    pub tokens: u64,
    pub last_refill: Instant,
}

#[derive(Debug, Clone)]
pub enum CounterState {
    FixedWindow(FixedWindowCounter),
    SlidingWindow(SlidingWindowCounter),
    TokenBucket(TokenBucketState),
}

#[derive(Debug, Clone)]
pub struct CounterKey {
    pub rule_id: Uuid,
    pub ip: Option<IpAddr>,
}

impl PartialEq for CounterKey {
    fn eq(&self, other: &Self) -> bool {
        self.rule_id == other.rule_id && self.ip == other.ip
    }
}

impl Eq for CounterKey {}

impl std::hash::Hash for CounterKey {
    fn hash<H: std::hash::Hasher>(&self, state: &mut H) {
        self.rule_id.hash(state);
        self.ip.hash(state);
    }
}

#[derive(Debug, Clone, Serialize)]
pub struct PathStatistics {
    pub path_pattern: String,
    pub total_requests: u64,
    pub throttled_requests: u64,
    pub last_throttled_at: Option<chrono::DateTime<chrono::Utc>>,
}

#[derive(Debug, Clone, Serialize)]
pub struct IpThrottledState {
    pub ip: IpAddr,
    pub paths: Vec<PathThrottledInfo>,
}

#[derive(Debug, Clone, Serialize)]
pub struct PathThrottledInfo {
    pub path_pattern: String,
    pub is_throttled: bool,
    pub remaining: u64,
    pub retry_after_secs: u64,
}

#[derive(Debug, Clone)]
pub struct RateLimitCheckResult {
    pub allowed: bool,
    pub remaining: u64,
    pub retry_after: Option<Duration>,
}

impl Default for RateLimitCheckResult {
    fn default() -> Self {
        RateLimitCheckResult {
            allowed: true,
            remaining: u64::MAX,
            retry_after: None,
        }
    }
}
