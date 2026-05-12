use std::collections::{HashMap, VecDeque};
use std::time::{Duration, Instant};

#[derive(Debug, Clone)]
pub struct CacheEntry {
    pub value: Vec<u8>,
    pub expires_at: Option<Instant>,
}

pub struct CacheService {
    max_size: usize,
    data: HashMap<String, CacheEntry>,
    lru_order: VecDeque<String>,
}

impl CacheService {
    pub fn new(max_size: usize) -> Self {
        CacheService {
            max_size,
            data: HashMap::new(),
            lru_order: VecDeque::new(),
        }
    }

    pub fn get(&mut self, key: &str) -> Option<&CacheEntry> {
        if self.data.contains_key(key) {
            self.touch_lru(key);
        }
        self.clean_expired();
        self.data.get(key)
    }

    pub fn set(&mut self, key: String, value: Vec<u8>, ttl_seconds: i64) -> bool {
        let needs_eviction = !self.data.contains_key(&key) && self.data.len() >= self.max_size;
        let mut eviction_failed = false;

        if needs_eviction {
            eviction_failed = !self.evict_lru();
        }

        let expires_at = if ttl_seconds > 0 {
            Some(Instant::now() + Duration::from_secs(ttl_seconds as u64))
        } else {
            None
        };

        let entry = CacheEntry {
            value,
            expires_at,
        };

        self.data.insert(key.clone(), entry);
        self.touch_lru(&key);

        eviction_failed
    }

    pub fn delete(&mut self, key: &str) -> bool {
        let removed = self.data.remove(key).is_some();
        if removed {
            self.remove_lru(key);
        }
        removed
    }

    pub fn batch_set(
        &mut self,
        entries: Vec<(String, Vec<u8>, i64)>,
    ) -> bool {
        let mut any_eviction_failed = false;

        for (key, value, ttl) in entries {
            if self.set(key, value, ttl) {
                any_eviction_failed = true;
            }
        }

        any_eviction_failed
    }

    pub fn batch_get(
        &mut self,
        keys: Vec<&str>,
    ) -> (HashMap<String, Vec<u8>>, Vec<String>) {
        let mut hits = HashMap::new();
        let mut misses = Vec::new();

        for key in keys {
            match self.get(key) {
                Some(entry) => {
                    hits.insert(key.to_string(), entry.value.clone());
                }
                None => {
                    misses.push(key.to_string());
                }
            }
        }

        (hits, misses)
    }

    fn touch_lru(&mut self, key: &str) {
        self.remove_lru(key);
        self.lru_order.push_front(key.to_string());
    }

    fn remove_lru(&mut self, key: &str) {
        if let Some(pos) = self.lru_order.iter().position(|k| k == key) {
            self.lru_order.remove(pos);
        }
    }

    fn evict_lru(&mut self) -> bool {
        self.clean_expired();

        if let Some(lru_key) = self.lru_order.pop_back() {
            if self.data.remove(&lru_key).is_some() {
                tracing::debug!(key = %lru_key, "Evicted key from cache");
                true
            } else {
                self.lru_order.push_back(lru_key);
                tracing::error!("LRU key not found in data during eviction");
                false
            }
        } else {
            tracing::error!("Cache is full but LRU order is empty");
            false
        }
    }

    fn clean_expired(&mut self) {
        let now = Instant::now();
        let expired_keys: Vec<String> = self
            .data
            .iter()
            .filter_map(|(key, entry)| {
                if let Some(expires_at) = entry.expires_at {
                    if now >= expires_at {
                        Some(key.clone())
                    } else {
                        None
                    }
                } else {
                    None
                }
            })
            .collect();

        for key in expired_keys {
            self.data.remove(&key);
            self.remove_lru(&key);
        }
    }

}

pub fn validate_key(key: &str) -> Result<(), String> {
    if key.is_empty() {
        return Err("key 不能为空".to_string());
    }
    if key.as_bytes().len() > 256 {
        return Err("key 长度超过 256 字节".to_string());
    }
    Ok(())
}

pub fn validate_ttl(ttl: i64) -> Result<(), String> {
    if ttl < 0 {
        return Err("过期时间不能为负".to_string());
    }
    Ok(())
}
