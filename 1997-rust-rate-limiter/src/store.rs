use crate::algorithms::{check_counter, init_counter_for_rule};
use crate::types::{CounterKey, CounterState, PathStatistics, RateLimitCheckResult, RateLimitRule};
use chrono::Utc;
use dashmap::DashMap;
use lru::LruCache;
use parking_lot::Mutex;
use std::collections::HashMap;
use std::net::IpAddr;
use std::num::NonZeroUsize;
use std::sync::Arc;
use std::time::Instant;
use uuid::Uuid;

pub struct CounterStore {
    counters: DashMap<CounterKey, CounterState>,
    lru_cache: Mutex<LruCache<CounterKey, ()>>,
    max_counters: usize,
}

impl CounterStore {
    pub fn new(max_counters: usize) -> Self {
        let capacity = NonZeroUsize::new(max_counters.max(1)).unwrap();
        CounterStore {
            counters: DashMap::new(),
            lru_cache: Mutex::new(LruCache::new(capacity)),
            max_counters,
        }
    }

    pub fn get_or_create(
        &self,
        key: &CounterKey,
        rule: &RateLimitRule,
        now: Instant,
    ) -> CounterState {
        if let Some(counter) = self.counters.get(key) {
            self.touch_key(key);
            return counter.clone();
        }

        let new_counter = init_counter_for_rule(rule, now);
        self.counters.insert(key.clone(), new_counter.clone());
        self.touch_key(key);
        self.evict_if_needed();
        new_counter
    }

    pub fn update(&self, key: &CounterKey, state: CounterState) {
        self.counters.insert(key.clone(), state);
        self.touch_key(key);
    }

    fn touch_key(&self, key: &CounterKey) {
        let mut cache = self.lru_cache.lock();
        cache.put(key.clone(), ());
    }

    fn evict_if_needed(&self) {
        let mut cache = self.lru_cache.lock();
        while cache.len() > self.max_counters {
            if let Some((evicted_key, _)) = cache.pop_lru() {
                self.counters.remove(&evicted_key);
            } else {
                break;
            }
        }
    }

    pub fn remove_by_rule(&self, rule_id: &Uuid) {
        let keys_to_remove: Vec<CounterKey> = self
            .counters
            .iter()
            .filter(|entry| entry.key().rule_id == *rule_id)
            .map(|entry| entry.key().clone())
            .collect();

        for key in keys_to_remove {
            self.counters.remove(&key);
            let mut cache = self.lru_cache.lock();
            cache.pop(&key);
        }
    }

    pub fn len(&self) -> usize {
        self.counters.len()
    }
}

pub struct StatisticsStore {
    path_stats: DashMap<String, PathStatistics>,
    ip_throttled_paths: DashMap<IpAddr, HashMap<String, (bool, u64, u64)>>,
}

impl StatisticsStore {
    pub fn new() -> Self {
        StatisticsStore {
            path_stats: DashMap::new(),
            ip_throttled_paths: DashMap::new(),
        }
    }

    pub fn record_request(&self, path: &str, throttled: bool) {
        let mut entry = self
            .path_stats
            .entry(path.to_string())
            .or_insert_with(|| PathStatistics {
                path_pattern: path.to_string(),
                total_requests: 0,
                throttled_requests: 0,
                last_throttled_at: None,
            });

        entry.total_requests += 1;
        if throttled {
            entry.throttled_requests += 1;
            entry.last_throttled_at = Some(Utc::now());
        }
    }

    pub fn record_ip_throttled(
        &self,
        ip: IpAddr,
        path: &str,
        is_throttled: bool,
        remaining: u64,
        retry_after_secs: u64,
    ) {
        let mut entry = self
            .ip_throttled_paths
            .entry(ip)
            .or_insert_with(|| HashMap::new());

        if is_throttled {
            entry.insert(path.to_string(), (is_throttled, remaining, retry_after_secs));
        } else {
            entry.remove(path);
        }
    }

    pub fn get_path_stats(&self) -> Vec<PathStatistics> {
        self.path_stats
            .iter()
            .map(|entry| entry.value().clone())
            .collect()
    }

    pub fn get_ip_throttled_state(
        &self,
        ip: IpAddr,
    ) -> Vec<(String, bool, u64, u64)> {
        self.ip_throttled_paths
            .get(&ip)
            .map(|entry| {
                entry
                    .iter()
                    .map(|(path, &(throttled, remaining, retry))| {
                        (path.clone(), throttled, remaining, retry)
                    })
                    .collect()
            })
            .unwrap_or_default()
    }

    pub fn get_all_throttled_ips(&self) -> Vec<IpAddr> {
        self.ip_throttled_paths
            .iter()
            .map(|entry| *entry.key())
            .collect()
    }
}

pub struct RateLimiterEngine {
    rules: DashMap<Uuid, RateLimitRule>,
    rule_order: Mutex<Vec<Uuid>>,
    counter_store: Arc<CounterStore>,
    stats_store: Arc<StatisticsStore>,
}

impl RateLimiterEngine {
    pub fn new(max_counters: usize) -> Self {
        RateLimiterEngine {
            rules: DashMap::new(),
            rule_order: Mutex::new(Vec::new()),
            counter_store: Arc::new(CounterStore::new(max_counters)),
            stats_store: Arc::new(StatisticsStore::new()),
        }
    }

    pub fn add_rule(&self, rule: RateLimitRule) {
        let rule_id = rule.id;
        self.rules.insert(rule_id, rule);
        let mut order = self.rule_order.lock();
        if !order.contains(&rule_id) {
            order.push(rule_id);
        }
    }

    pub fn update_rule(&self, rule: RateLimitRule) -> bool {
        if self.rules.contains_key(&rule.id) {
            self.rules.insert(rule.id, rule);
            true
        } else {
            false
        }
    }

    pub fn remove_rule(&self, rule_id: &Uuid) -> bool {
        if self.rules.remove(rule_id).is_some() {
            let mut order = self.rule_order.lock();
            order.retain(|id| id != rule_id);
            self.counter_store.remove_by_rule(rule_id);
            true
        } else {
            false
        }
    }

    pub fn get_rule(&self, rule_id: &Uuid) -> Option<RateLimitRule> {
        self.rules.get(rule_id).map(|entry| entry.value().clone())
    }

    pub fn get_all_rules(&self) -> Vec<RateLimitRule> {
        let order = self.rule_order.lock();
        order
            .iter()
            .filter_map(|id| self.rules.get(id).map(|entry| entry.value().clone()))
            .collect()
    }

    pub fn match_rules(&self, path: &str) -> Vec<RateLimitRule> {
        let order = self.rule_order.lock();
        let mut matched_exact: Vec<(usize, RateLimitRule)> = Vec::new();
        let mut matched_prefix: Vec<(usize, String, RateLimitRule)> = Vec::new();

        for (idx, rule_id) in order.iter().enumerate() {
            if let Some(rule) = self.rules.get(rule_id) {
                let rule = rule.value();
                match rule.path_type {
                    crate::types::PathPatternType::Exact => {
                        if rule.path_pattern == path {
                            matched_exact.push((idx, rule.clone()));
                        }
                    }
                    crate::types::PathPatternType::Prefix => {
                        if path.starts_with(&rule.path_pattern) {
                            matched_prefix
                                .push((idx, rule.path_pattern.clone(), rule.clone()));
                        }
                    }
                }
            }
        }

        matched_prefix.sort_by(|a, b| {
            let len_cmp = b.1.len().cmp(&a.1.len());
            if len_cmp != std::cmp::Ordering::Equal {
                len_cmp
            } else {
                a.0.cmp(&b.0)
            }
        });

        let mut result: Vec<RateLimitRule> = matched_exact.into_iter().map(|(_, r)| r).collect();
        result.extend(matched_prefix.into_iter().map(|(_, _, r)| r));
        result
    }

    pub fn check_request(
        &self,
        path: &str,
        ip: Option<IpAddr>,
    ) -> RateLimitCheckResult {
        let now = Instant::now();
        let rules = self.match_rules(path);

        if rules.is_empty() {
            return RateLimitCheckResult::default();
        }

        let mut min_remaining = u64::MAX;
        let mut max_retry_after = None;
        let mut any_throttled = false;

        for rule in &rules {
            let key = CounterKey {
                rule_id: rule.id,
                ip: if rule.per_ip { ip } else { None },
            };

            let mut counter = self.counter_store.get_or_create(&key, rule, now);
            let result = check_counter(rule, &mut counter, now);
            self.counter_store.update(&key, counter);

            if !result.allowed {
                any_throttled = true;
                if result.remaining < min_remaining {
                    min_remaining = result.remaining;
                }
                if let Some(retry) = result.retry_after {
                    if max_retry_after.map_or(true, |current: std::time::Duration| retry > current)
                    {
                        max_retry_after = Some(retry);
                    }
                }
            }

            if any_throttled {
                break;
            }
        }

        self.stats_store.record_request(path, any_throttled);

        if any_throttled {
            if let Some(client_ip) = ip {
                for rule in &rules {
                    self.stats_store.record_ip_throttled(
                        client_ip,
                        &rule.path_pattern,
                        true,
                        min_remaining,
                        max_retry_after.map_or(0, |d| d.as_secs()),
                    );
                }
            }
            RateLimitCheckResult {
                allowed: false,
                remaining: min_remaining,
                retry_after: max_retry_after,
            }
        } else {
            if let Some(client_ip) = ip {
                for rule in &rules {
                    self.stats_store
                        .record_ip_throttled(client_ip, &rule.path_pattern, false, 0, 0);
                }
            }
            RateLimitCheckResult::default()
        }
    }

    pub fn get_stats_store(&self) -> Arc<StatisticsStore> {
        self.stats_store.clone()
    }

    pub fn get_counter_store(&self) -> Arc<CounterStore> {
        self.counter_store.clone()
    }
}
