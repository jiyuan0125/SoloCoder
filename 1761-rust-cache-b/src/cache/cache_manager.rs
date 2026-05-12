use std::collections::{HashMap, VecDeque};
use std::sync::{Arc, RwLock};

use super::cache_entry::{CacheEntry, CacheValue};
use super::stats::CacheStats;

pub const DEFAULT_CAPACITY: usize = 10_000;
pub const DEFAULT_TTL_SECS: u64 = 3600;

type KeyRef = String;

#[derive(Default)]
struct InternalCache {
    entries: HashMap<String, CacheEntry>,
    lru_order: VecDeque<KeyRef>,
}

pub struct CacheManager {
    capacity: usize,
    default_ttl: u64,
    inner: Arc<RwLock<InternalCache>>,
    stats: Arc<CacheStats>,
}

impl Clone for CacheManager {
    fn clone(&self) -> Self {
        CacheManager {
            capacity: self.capacity,
            default_ttl: self.default_ttl,
            inner: Arc::clone(&self.inner),
            stats: Arc::clone(&self.stats),
        }
    }
}

impl CacheManager {
    pub fn new() -> Self {
        CacheManager::with_capacity(DEFAULT_CAPACITY, DEFAULT_TTL_SECS)
    }

    pub fn with_capacity(capacity: usize, default_ttl: u64) -> Self {
        CacheManager {
            capacity,
            default_ttl,
            inner: Arc::new(RwLock::new(InternalCache::default())),
            stats: Arc::new(CacheStats::default()),
        }
    }

    pub fn get(&self, key: &str) -> Option<CacheValue> {
        self.get_internal(key, false).map(|e| e.value)
    }

    pub fn get_with_meta(&self, key: &str) -> Option<CacheEntry> {
        self.get_internal(key, true)
    }

    fn get_internal(&self, key: &str, with_meta: bool) -> Option<CacheEntry> {
        let result = {
            let mut inner = self.inner.write().expect("Lock poisoned");
            Self::get_inner(&mut inner, key)
        };

        match result {
            Some(entry) => {
                if entry.is_expired() {
                    self.delete(key);
                    self.stats.record_miss();
                    None
                } else {
                    self.stats.record_hit();
                    if with_meta {
                        Some(entry)
                    } else {
                        Some(entry)
                    }
                }
            }
            None => {
                self.stats.record_miss();
                None
            }
        }
    }

    fn get_inner(inner: &mut InternalCache, key: &str) -> Option<CacheEntry> {
        if let Some(entry) = inner.entries.get(key).cloned() {
            inner.lru_order.retain(|k| k != key);
            inner.lru_order.push_back(key.to_string());
            Some(entry)
        } else {
            None
        }
    }

    pub fn set(&self, key: String, value: CacheValue, ttl_seconds: Option<u64>) {
        let ttl = ttl_seconds.unwrap_or(self.default_ttl);
        let entry = CacheEntry::new(key.clone(), value, Some(ttl));

        {
            let mut inner = self.inner.write().expect("Lock poisoned");
            
            if inner.entries.len() >= self.capacity && !inner.entries.contains_key(&key) {
                Self::evict_lru_inner(&mut inner);
            }
            
            inner.entries.insert(key.clone(), entry);
            inner.lru_order.retain(|k| k != &key);
            inner.lru_order.push_back(key);
        }
        
        self.stats.record_set();
    }

    fn evict_lru_inner(inner: &mut InternalCache) {
        if let Some(key) = inner.lru_order.pop_front() {
            inner.entries.remove(&key);
        }
    }

    pub fn delete(&self, key: &str) -> bool {
        let deleted = {
            let mut inner = self.inner.write().expect("Lock poisoned");
            let existed = inner.entries.remove(key).is_some();
            if existed {
                inner.lru_order.retain(|k| k != key);
            }
            existed
        };
        
        if deleted {
            self.stats.record_delete();
        }
        
        deleted
    }

    pub fn clear(&self) {
        let mut inner = self.inner.write().expect("Lock poisoned");
        inner.entries.clear();
        inner.lru_order.clear();
    }

    pub fn len(&self) -> usize {
        let inner = self.inner.read().expect("Lock poisoned");
        inner.entries.len()
    }

    pub fn is_empty(&self) -> bool {
        let inner = self.inner.read().expect("Lock poisoned");
        inner.entries.is_empty()
    }

    pub fn capacity(&self) -> usize {
        self.capacity
    }

    pub fn stats(&self) -> &CacheStats {
        &self.stats
    }

    pub fn cleanup_expired(&self) -> usize {
        let removed_keys: Vec<String> = {
            let inner = self.inner.read().expect("Lock poisoned");
            inner.entries.iter()
                .filter(|(_, entry)| entry.is_expired())
                .map(|(key, _)| key.clone())
                .collect()
        };

        let mut count = 0;
        for key in &removed_keys {
            if self.delete(key) {
                self.stats.record_expire_removed();
                count += 1;
            }
        }

        count
    }

    pub fn batch_get(&self, keys: &[String]) -> HashMap<String, Option<CacheValue>> {
        keys.iter()
            .map(|key| (key.clone(), self.get(key)))
            .collect()
    }

    pub fn batch_set(&self, items: Vec<(String, CacheValue, Option<u64>)>) -> Vec<(String, bool)> {
        let mut results = Vec::with_capacity(items.len());
        
        for (key, value, ttl) in items {
            self.set(key.clone(), value, ttl);
            results.push((key, true));
        }
        
        results
    }

    pub fn batch_delete(&self, keys: &[String]) -> HashMap<String, bool> {
        keys.iter()
            .map(|key| (key.clone(), self.delete(key)))
            .collect()
    }
}

impl Default for CacheManager {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_basic_set_get() {
        let cache = CacheManager::with_capacity(100, 3600);
        
        cache.set("key1".into(), "value1".to_string().into(), None);
        let result = cache.get("key1");
        
        assert!(result.is_some());
        if let Some(CacheValue::String(v)) = result {
            assert_eq!(v, "value1");
        } else {
            panic!("Expected String value");
        }
    }

    #[test]
    fn test_get_missing() {
        let cache = CacheManager::new();
        assert!(cache.get("nonexistent").is_none());
    }

    #[test]
    fn test_delete() {
        let cache = CacheManager::new();
        
        cache.set("key".into(), "value".to_string().into(), None);
        assert!(cache.get("key").is_some());
        
        assert!(cache.delete("key"));
        assert!(cache.get("key").is_none());
    }

    #[test]
    fn test_lru_eviction() {
        let cache = CacheManager::with_capacity(3, 3600);
        
        cache.set("a".into(), 1i64.into(), None);
        cache.set("b".into(), 2i64.into(), None);
        cache.set("c".into(), 3i64.into(), None);
        
        assert_eq!(cache.len(), 3);
        
        cache.get("a");
        
        cache.set("d".into(), 4i64.into(), None);
        
        assert_eq!(cache.len(), 3);
        assert!(cache.get("b").is_none());
        assert!(cache.get("a").is_some());
        assert!(cache.get("c").is_some());
        assert!(cache.get("d").is_some());
    }

    #[test]
    fn test_stats_hit_miss() {
        let cache = CacheManager::new();
        
        cache.set("key".into(), "value".to_string().into(), None);
        
        cache.get("key");
        cache.get("key");
        cache.get("nonexistent");
        
        assert_eq!(cache.stats().get_hits(), 2);
        assert_eq!(cache.stats().get_misses(), 1);
    }

    #[test]
    fn test_batch_operations() {
        let cache = CacheManager::new();
        
        let items = vec![
            ("k1".to_string(), "v1".to_string().into(), None),
            ("k2".to_string(), 42i64.into(), None),
            ("k3".to_string(), true.into(), None),
        ];
        
        let results = cache.batch_set(items);
        assert_eq!(results.len(), 3);
        
        let keys = vec!["k1".to_string(), "k2".to_string(), "k99".to_string()];
        let batch_get = cache.batch_get(&keys);
        
        assert!(batch_get.get("k1").unwrap().is_some());
        assert!(batch_get.get("k2").unwrap().is_some());
        assert!(batch_get.get("k99").unwrap().is_none());
    }
}
