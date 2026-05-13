use std::{
    collections::{HashMap, VecDeque, BTreeMap},
    sync::Arc,
    time::Instant,
};

use axum::{
    extract::{Json, State},
    http::StatusCode,
    routing::{get, post},
    Router,
};
use parking_lot::Mutex;
use serde::{Deserialize, Serialize};

const MAX_VALUE_HARD_LIMIT: usize = 2 * 1024 * 1024;
const MAX_VALUE_SOFT_LIMIT: usize = 1 * 1024 * 1024;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum EvictionPolicy {
    Lru,
    Lfu,
}

#[derive(Debug, Clone)]
pub struct CacheEntry {
    key: String,
    value: String,
    size: usize,
    access_count: u64,
    insert_order: u64,
    last_access: Instant,
}

impl CacheEntry {
    fn new(key: String, value: String, insert_order: u64) -> Self {
        let size = key.len() + value.len();
        Self {
            key,
            value,
            size,
            access_count: 1,
            insert_order,
            last_access: Instant::now(),
        }
    }
}

#[derive(Debug, Clone)]
pub struct LfuIndex {
    freq_map: BTreeMap<u64, VecDeque<String>>,
    key_to_freq: HashMap<String, u64>,
}

impl LfuIndex {
    fn new() -> Self {
        Self {
            freq_map: BTreeMap::new(),
            key_to_freq: HashMap::new(),
        }
    }

    fn insert(&mut self, key: &str, freq: u64) {
        self.key_to_freq.insert(key.to_string(), freq);
        self.freq_map
            .entry(freq)
            .or_default()
            .push_back(key.to_string());
    }

    fn increment(&mut self, key: &str) {
        if let Some(&old_freq) = self.key_to_freq.get(key) {
            if let Some(queue) = self.freq_map.get_mut(&old_freq) {
                if let Some(pos) = queue.iter().position(|k| k == key) {
                    queue.remove(pos);
                }
                if queue.is_empty() {
                    self.freq_map.remove(&old_freq);
                }
            }
            let new_freq = old_freq + 1;
            self.key_to_freq.insert(key.to_string(), new_freq);
            self.freq_map
                .entry(new_freq)
                .or_default()
                .push_back(key.to_string());
        }
    }

    fn remove(&mut self, key: &str) {
        if let Some(freq) = self.key_to_freq.remove(key) {
            if let Some(queue) = self.freq_map.get_mut(&freq) {
                if let Some(pos) = queue.iter().position(|k| k == key) {
                    queue.remove(pos);
                }
                if queue.is_empty() {
                    self.freq_map.remove(&freq);
                }
            }
        }
    }

    fn evict_candidate(&self) -> Option<String> {
        self.freq_map
            .values()
            .next()
            .and_then(|queue| queue.front().cloned())
    }

    fn reset_all(&mut self, entries: &HashMap<String, CacheEntry>) {
        self.freq_map.clear();
        self.key_to_freq.clear();
        for entry in entries.values() {
            self.insert(&entry.key, 1);
        }
    }
}

#[derive(Debug, Clone)]
pub struct LruIndex {
    order: VecDeque<String>,
}

impl LruIndex {
    fn new() -> Self {
        Self {
            order: VecDeque::new(),
        }
    }

    fn touch(&mut self, key: &str) {
        if let Some(pos) = self.order.iter().position(|k| k == key) {
            self.order.remove(pos);
        }
        self.order.push_back(key.to_string());
    }

    fn remove(&mut self, key: &str) {
        if let Some(pos) = self.order.iter().position(|k| k == key) {
            self.order.remove(pos);
        }
    }

    fn evict_candidate(&self) -> Option<String> {
        self.order.front().cloned()
    }

    fn reset_all(&mut self, entries: &HashMap<String, CacheEntry>) {
        self.order.clear();
        let mut sorted: Vec<_> = entries.values().collect();
        sorted.sort_by(|a, b| a.last_access.cmp(&b.last_access));
        for entry in sorted {
            self.order.push_back(entry.key.clone());
        }
    }
}

#[derive(Debug, Clone)]
pub struct CacheInner {
    entries: HashMap<String, CacheEntry>,
    lru: LruIndex,
    lfu: LfuIndex,
    policy: EvictionPolicy,
    max_size: usize,
    current_size: usize,
    insert_counter: u64,
    evictions: u64,
    hits: u64,
    total_requests: u64,
}

impl CacheInner {
    fn new(max_size: usize, policy: EvictionPolicy) -> Self {
        Self {
            entries: HashMap::new(),
            lru: LruIndex::new(),
            lfu: LfuIndex::new(),
            policy,
            max_size,
            current_size: 0,
            insert_counter: 0,
            evictions: 0,
            hits: 0,
            total_requests: 0,
        }
    }

    fn record_get(&mut self, hit: bool) {
        self.total_requests += 1;
        if hit {
            self.hits += 1;
        }
    }

    fn set_policy(&mut self, policy: EvictionPolicy) {
        if self.policy != policy {
            self.policy = policy;
            match policy {
                EvictionPolicy::Lru => {
                    self.lru.reset_all(&self.entries);
                }
                EvictionPolicy::Lfu => {
                    for entry in self.entries.values_mut() {
                        entry.access_count = 1;
                    }
                    self.lfu.reset_all(&self.entries);
                }
            }
        }
    }

    fn evict(&mut self) -> Option<String> {
        let key = match self.policy {
            EvictionPolicy::Lru => self.lru.evict_candidate(),
            EvictionPolicy::Lfu => self.lfu.evict_candidate(),
        };

        if let Some(ref key) = key {
            if let Some(entry) = self.entries.remove(key) {
                self.current_size -= entry.size;
                self.lru.remove(key);
                self.lfu.remove(key);
                self.evictions += 1;
            }
        }
        key
    }

    fn ensure_space(&mut self, needed: usize) -> bool {
        while self.current_size + needed > self.max_size {
            if self.evict().is_none() {
                return false;
            }
        }
        true
    }

    fn get(&mut self, key: &str) -> Option<&str> {
        let hit = self.entries.contains_key(key);
        self.record_get(hit);

        if !hit {
            return None;
        }

        self.lru.touch(key);
        self.lfu.increment(key);

        if let Some(entry) = self.entries.get_mut(key) {
            entry.access_count += 1;
            entry.last_access = Instant::now();
            Some(&entry.value)
        } else {
            None
        }
    }

    fn set(&mut self, key: String, value: String) -> Result<(), ()> {
        let value_size = value.len();
        if value_size > MAX_VALUE_HARD_LIMIT {
            return Err(());
        }

        let new_entry_size = key.len() + value_size;
        let old_entry_size = self
            .entries
            .get(&key)
            .map(|e| e.size)
            .unwrap_or(0);

        let additional_size = new_entry_size.saturating_sub(old_entry_size);

        if !self.ensure_space(additional_size) {
            return Err(());
        }

        if let Some(old_entry) = self.entries.remove(&key) {
            self.current_size -= old_entry.size;
            self.lru.remove(&key);
            self.lfu.remove(&key);
        }

        let insert_order = self.insert_counter;
        self.insert_counter += 1;

        let entry = CacheEntry::new(key.clone(), value, insert_order);
        self.current_size += entry.size;

        self.lru.touch(&key);
        self.lfu.insert(&key, 1);

        self.entries.insert(key, entry);

        Ok(())
    }

    fn batch_get(&mut self, keys: &[String]) -> Vec<Option<String>> {
        keys.iter().map(|k| self.get(k).map(|v| v.to_string())).collect()
    }

    fn stats(&self) -> CacheStats {
        let hit_rate = if self.total_requests > 0 {
            self.hits as f64 / self.total_requests as f64
        } else {
            0.0
        };
        CacheStats {
            hit_rate,
            total_requests: self.total_requests,
            evictions: self.evictions,
            current_entries: self.entries.len(),
            current_size: self.current_size,
            max_size: self.max_size,
            policy: self.policy,
        }
    }

    fn export(&self) -> HashMap<String, String> {
        self.entries
            .iter()
            .map(|(k, v)| (k.clone(), v.value.clone()))
            .collect()
    }

    fn warm(&mut self, data: HashMap<String, String>) -> WarmResult {
        let mut loaded = 0;
        let mut skipped = 0;
        let mut failed = 0;

        for (key, value) in data {
            if value.len() > MAX_VALUE_HARD_LIMIT {
                failed += 1;
                continue;
            }
            if value.len() > MAX_VALUE_SOFT_LIMIT {
                skipped += 1;
                continue;
            }
            match self.set(key, value) {
                Ok(()) => loaded += 1,
                Err(()) => failed += 1,
            }
        }

        WarmResult {
            loaded,
            skipped,
            failed,
        }
    }
}

#[derive(Debug, Clone)]
pub struct Cache {
    inner: Arc<Mutex<CacheInner>>,
}

impl Cache {
    pub fn new(max_size: usize, policy: EvictionPolicy) -> Self {
        Self {
            inner: Arc::new(Mutex::new(CacheInner::new(max_size, policy))),
        }
    }

    pub fn get(&self, key: &str) -> Option<String> {
        let mut inner = self.inner.lock();
        inner.get(key).map(|s| s.to_string())
    }

    pub fn set(&self, key: String, value: String) -> Result<(), ()> {
        let mut inner = self.inner.lock();
        inner.set(key, value)
    }

    pub fn batch_get(&self, keys: &[String]) -> Vec<Option<String>> {
        let mut inner = self.inner.lock();
        inner.batch_get(keys)
    }

    pub fn stats(&self) -> CacheStats {
        let inner = self.inner.lock();
        inner.stats()
    }

    pub fn export(&self) -> HashMap<String, String> {
        let inner = self.inner.lock();
        inner.export()
    }

    pub fn warm(&self, data: HashMap<String, String>) -> WarmResult {
        let mut inner = self.inner.lock();
        inner.warm(data)
    }

    pub fn set_policy(&self, policy: EvictionPolicy) {
        let mut inner = self.inner.lock();
        inner.set_policy(policy);
    }
}

#[derive(Debug, Serialize)]
pub struct CacheStats {
    pub hit_rate: f64,
    pub total_requests: u64,
    pub evictions: u64,
    pub current_entries: usize,
    pub current_size: usize,
    pub max_size: usize,
    pub policy: EvictionPolicy,
}

#[derive(Debug, Serialize)]
pub struct WarmResult {
    pub loaded: usize,
    pub skipped: usize,
    pub failed: usize,
}

#[derive(Debug, Deserialize)]
pub struct BatchOperation {
    pub op: String,
    pub key: String,
    #[serde(default)]
    pub value: Option<String>,
}

#[derive(Debug, Serialize)]
pub struct BatchResultItem {
    pub key: String,
    pub status: u16,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub value: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub error: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct WarmRequest {
    pub file_path: String,
}

#[derive(Debug, Deserialize)]
pub struct PolicyRequest {
    pub policy: EvictionPolicy,
}

async fn get_handler(
    State(cache): State<Cache>,
    axum::extract::Path(key): axum::extract::Path<String>,
) -> Result<Json<Option<String>>, StatusCode> {
    let result = cache.get(&key);
    Ok(Json(result))
}

async fn set_handler(
    State(cache): State<Cache>,
    axum::extract::Path(key): axum::extract::Path<String>,
    body: String,
) -> Result<StatusCode, StatusCode> {
    if body.len() > MAX_VALUE_HARD_LIMIT {
        return Err(StatusCode::PAYLOAD_TOO_LARGE);
    }
    if body.len() > MAX_VALUE_SOFT_LIMIT {
        return Err(StatusCode::UNPROCESSABLE_ENTITY);
    }
    match cache.set(key, body) {
        Ok(()) => Ok(StatusCode::OK),
        Err(()) => Err(StatusCode::INSUFFICIENT_STORAGE),
    }
}

async fn batch_handler(
    State(cache): State<Cache>,
    Json(ops): Json<Vec<BatchOperation>>,
) -> Json<Vec<BatchResultItem>> {
    let mut results = Vec::with_capacity(ops.len());

    for op in ops {
        match op.op.to_lowercase().as_str() {
            "get" => {
                let value = cache.get(&op.key);
                let status = if value.is_some() { 200 } else { 404 };
                results.push(BatchResultItem {
                    key: op.key,
                    status,
                    value,
                    error: None,
                });
            }
            "set" => {
                if let Some(value) = op.value {
                    if value.len() > MAX_VALUE_HARD_LIMIT {
                        results.push(BatchResultItem {
                            key: op.key,
                            status: 413,
                            value: None,
                            error: Some("Value exceeds hard limit of 2MB".to_string()),
                        });
                    } else if value.len() > MAX_VALUE_SOFT_LIMIT {
                        results.push(BatchResultItem {
                            key: op.key,
                            status: 422,
                            value: None,
                            error: Some("Value exceeds soft limit of 1MB".to_string()),
                        });
                    } else {
                        match cache.set(op.key.clone(), value) {
                            Ok(()) => results.push(BatchResultItem {
                                key: op.key,
                                status: 200,
                                value: None,
                                error: None,
                            }),
                            Err(()) => results.push(BatchResultItem {
                                key: op.key,
                                status: 507,
                                value: None,
                                error: Some("Insufficient cache storage".to_string()),
                            }),
                        }
                    }
                } else {
                    results.push(BatchResultItem {
                        key: op.key,
                        status: 400,
                        value: None,
                        error: Some("Missing value for set operation".to_string()),
                    });
                }
            }
            _ => results.push(BatchResultItem {
                key: op.key,
                status: 400,
                value: None,
                error: Some(format!("Unknown operation: {}", op.op)),
            }),
        }
    }

    Json(results)
}

async fn stats_handler(State(cache): State<Cache>) -> Json<CacheStats> {
    Json(cache.stats())
}

async fn export_handler(State(cache): State<Cache>) -> Json<HashMap<String, String>> {
    Json(cache.export())
}

async fn warm_handler(
    State(cache): State<Cache>,
    Json(req): Json<WarmRequest>,
) -> Result<Json<WarmResult>, StatusCode> {
    let content = tokio::fs::read_to_string(&req.file_path)
        .await
        .map_err(|_| StatusCode::BAD_REQUEST)?;

    let data: HashMap<String, String> =
        serde_json::from_str(&content).map_err(|_| StatusCode::BAD_REQUEST)?;

    Ok(Json(cache.warm(data)))
}

async fn policy_handler(
    State(cache): State<Cache>,
    Json(req): Json<PolicyRequest>,
) -> StatusCode {
    cache.set_policy(req.policy);
    StatusCode::OK
}

#[tokio::main]
async fn main() {
    let max_size: usize = std::env::var("CACHE_MAX_SIZE")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(64 * 1024 * 1024);

    let policy = match std::env::var("CACHE_POLICY")
        .ok()
        .map(|s| s.to_lowercase())
        .as_deref()
    {
        Some("lfu") => EvictionPolicy::Lfu,
        _ => EvictionPolicy::Lru,
    };

    let cache = Cache::new(max_size, policy);

    let app = Router::new()
        .route("/cache/:key", get(get_handler).post(set_handler))
        .route("/cache/batch", post(batch_handler))
        .route("/cache/stats", get(stats_handler))
        .route("/cache/export", get(export_handler))
        .route("/cache/warm", post(warm_handler))
        .route("/cache/policy", post(policy_handler))
        .with_state(cache);

    let port: u16 = std::env::var("PORT")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(9103);

    let listener = tokio::net::TcpListener::bind(("0.0.0.0", port))
        .await
        .expect("Failed to bind");

    println!("Cache middleware running on http://0.0.0.0:{}", port);
    println!("Cache max size: {} bytes", max_size);
    println!("Initial policy: {:?}", policy);

    axum::serve(listener, app).await.expect("Server failed");
}
