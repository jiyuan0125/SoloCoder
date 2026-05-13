use dashmap::DashMap;
use std::time::{Duration, Instant};
use tokio::sync::RwLock;

struct CacheEntry {
    value: String,
    expiry: Instant,
}

pub struct CacheManager {
    store: DashMap<String, CacheEntry>,
    capacity: usize,
    lru_order: RwLock<Vec<String>>,
}

impl CacheManager {
    pub fn new(capacity: usize) -> Self {
        Self {
            store: DashMap::new(),
            capacity,
            lru_order: RwLock::new(Vec::new()),
        }
    }

    pub async fn get(&self, key: &str) -> Option<String> {
        let now = Instant::now();
        
        if let Some(entry) = self.store.get(key) {
            if entry.expiry > now {
                let value = entry.value.clone();
                drop(entry);
                self.touch_lru(key).await;
                return Some(value);
            } else {
                drop(entry);
                self.store.remove(key);
            }
        }
        None
    }

    pub async fn set(&self, key: String, value: String, ttl: Duration) {
        let entry = CacheEntry {
            value,
            expiry: Instant::now() + ttl,
        };

        if self.store.len() >= self.capacity && !self.store.contains_key(&key) {
            self.evict_lru().await;
        }

        self.store.insert(key.clone(), entry);
        self.touch_lru(&key).await;
    }

    pub async fn delete(&self, key: &str) {
        self.store.remove(key);
        let mut order = self.lru_order.write().await;
        if let Some(pos) = order.iter().position(|k| k == key) {
            order.remove(pos);
        }
    }

    pub async fn len(&self) -> usize {
        self.store.len()
    }

    pub async fn cleanup_expired(&self) {
        let now = Instant::now();
        let expired_keys: Vec<String> = self
            .store
            .iter()
            .filter(|entry| entry.value().expiry <= now)
            .map(|entry| entry.key().clone())
            .collect();
        
        let mut order = self.lru_order.write().await;
        for key in expired_keys {
            self.store.remove(&key);
            if let Some(pos) = order.iter().position(|k| k == &key) {
                order.remove(pos);
            }
        }
    }

    async fn touch_lru(&self, key: &str) {
        let mut order = self.lru_order.write().await;
        if let Some(pos) = order.iter().position(|k| k == key) {
            order.remove(pos);
        }
        order.push(key.to_string());
    }

    async fn evict_lru(&self) {
        let mut order = self.lru_order.write().await;
        if !order.is_empty() {
            let oldest = order.remove(0);
            self.store.remove(&oldest);
        }
    }
}
