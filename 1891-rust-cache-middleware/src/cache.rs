use std::collections::{HashMap, HashSet, VecDeque};
use std::time::{Duration, Instant};

#[derive(Debug, Clone)]
pub struct CacheEntry {
    pub value: Vec<u8>,
    pub expires_at: Option<Instant>,
    pub namespace: Option<String>,
}

#[derive(Debug, Default, Clone)]
pub struct NamespaceStats {
    pub total_queries: u64,
    pub cache_hits: u64,
}

#[derive(Debug, Default, Clone)]
pub struct CacheStats {
    pub total_queries: u64,
    pub cache_hits: u64,
    pub evictions: u64,
    pub expired_cleanups: u64,
    pub namespace_stats: HashMap<String, NamespaceStats>,
}

impl CacheStats {
    pub fn hit_rate(&self) -> f64 {
        if self.total_queries == 0 {
            0.0
        } else {
            self.cache_hits as f64 / self.total_queries as f64
        }
    }

    pub fn namespace_hit_rate(&self, namespace: &str) -> f64 {
        match self.namespace_stats.get(namespace) {
            Some(ns) if ns.total_queries > 0 => ns.cache_hits as f64 / ns.total_queries as f64,
            _ => 0.0,
        }
    }
}

#[derive(Debug)]
pub struct CacheEngine {
    capacity: usize,
    default_ttl: u64,
    data: HashMap<String, CacheEntry>,
    lru_order: VecDeque<String>,
    stats: CacheStats,
}

impl CacheEngine {
    pub fn new(capacity: usize, default_ttl: u64) -> Self {
        Self {
            capacity,
            default_ttl,
            data: HashMap::new(),
            lru_order: VecDeque::new(),
            stats: CacheStats::default(),
        }
    }

    pub fn capacity(&self) -> usize {
        self.capacity
    }

    pub fn set_capacity(&mut self, new_capacity: usize) {
        self.capacity = new_capacity;
        self.evict_to_capacity();
    }

    pub fn default_ttl(&self) -> u64 {
        self.default_ttl
    }

    pub fn set_default_ttl(&mut self, ttl: u64) {
        self.default_ttl = ttl;
    }

    pub fn entry_count(&self) -> usize {
        self.data.len()
    }

    pub fn stats(&self) -> CacheStats {
        self.stats.clone()
    }

    fn extract_namespace(key: &str) -> Option<String> {
        key.find(':').map(|idx| key[..idx].to_string())
    }

    fn is_expired(&self, entry: &CacheEntry) -> bool {
        entry.expires_at.map_or(false, |e| Instant::now() >= e)
    }

    fn cleanup_expired(&mut self, key: &str) -> bool {
        if let Some(entry) = self.data.get(key) {
            if self.is_expired(entry) {
                self.remove_internal(key);
                self.stats.expired_cleanups += 1;
                return true;
            }
        }
        false
    }

    fn update_lru_on_access(&mut self, key: &str) {
        if let Some(pos) = self.lru_order.iter().position(|k| k == key) {
            self.lru_order.remove(pos);
        }
        self.lru_order.push_front(key.to_string());
    }

    fn remove_internal(&mut self, key: &str) {
        if self.data.remove(key).is_some() {
            if let Some(pos) = self.lru_order.iter().position(|k| k == key) {
                self.lru_order.remove(pos);
            }
        }
    }

    fn evict_to_capacity(&mut self) {
        while self.data.len() > self.capacity {
            if let Some(lru_key) = self.lru_order.pop_back() {
                self.data.remove(&lru_key);
                self.stats.evictions += 1;
            } else {
                break;
            }
        }
    }

    pub fn get(&mut self, key: &str) -> Option<&Vec<u8>> {
        self.stats.total_queries += 1;

        let namespace = Self::extract_namespace(key);
        if let Some(ref ns) = namespace {
            self.stats
                .namespace_stats
                .entry(ns.clone())
                .or_default()
                .total_queries += 1;
        }

        if self.cleanup_expired(key) {
            return None;
        }

        let exists = self.data.contains_key(key);
        if exists {
            self.update_lru_on_access(key);
            self.stats.cache_hits += 1;
            if let Some(ref ns) = namespace {
                self.stats
                    .namespace_stats
                    .entry(ns.clone())
                    .or_default()
                    .cache_hits += 1;
            }
        }

        self.data.get(key).map(|e| &e.value)
    }

    pub fn set(&mut self, key: String, value: Vec<u8>, ttl_seconds: Option<u64>) -> bool {
        if key.len() > 256 || value.len() > 1_048_576 {
            return false;
        }

        let namespace = Self::extract_namespace(&key);
        let expires_at = ttl_seconds.or(Some(self.default_ttl)).map(|ttl| {
            if ttl == 0 {
                Instant::now()
            } else {
                Instant::now() + Duration::from_secs(ttl)
            }
        });

        let entry = CacheEntry {
            value,
            expires_at,
            namespace,
        };

        if self.data.contains_key(&key) {
            self.update_lru_on_access(&key);
        } else {
            self.lru_order.push_front(key.clone());
        }

        self.data.insert(key, entry);
        self.evict_to_capacity();
        true
    }

    pub fn delete(&mut self, key: &str) -> bool {
        let existed = self.data.contains_key(key);
        self.remove_internal(key);
        existed
    }

    pub fn delete_by_prefix(&mut self, prefix: &str) -> usize {
        let keys_to_delete: Vec<String> = self
            .data
            .keys()
            .filter(|k| k.starts_with(prefix))
            .cloned()
            .collect();

        let count = keys_to_delete.len();
        for key in keys_to_delete {
            self.remove_internal(&key);
        }
        count
    }

    pub fn clear_namespace(&mut self, namespace: &str) -> usize {
        let prefix = format!("{}:", namespace);
        self.delete_by_prefix(&prefix)
    }

    pub fn scan_by_prefix(&self, prefix: &str) -> Vec<(String, Vec<u8>)> {
        self.data
            .iter()
            .filter(|(k, v)| k.starts_with(prefix) && !self.is_expired(v))
            .map(|(k, v)| (k.clone(), v.value.clone()))
            .collect()
    }

    pub fn list_namespaces(&self) -> Vec<String> {
        let mut namespaces = HashSet::new();
        for entry in self.data.values() {
            if let Some(ns) = &entry.namespace {
                if entry.expires_at.map_or(true, |e| Instant::now() < e) {
                    namespaces.insert(ns.clone());
                }
            }
        }
        namespaces.into_iter().collect()
    }

    pub fn namespace_stats(&self, namespace: &str) -> NamespaceStats {
        self.stats
            .namespace_stats
            .get(namespace)
            .cloned()
            .unwrap_or_default()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::thread::sleep;

    #[test]
    fn test_basic_set_get() {
        let mut cache = CacheEngine::new(100, 300);
        assert!(cache.set("key1".into(), b"value1".to_vec(), None));
        assert_eq!(cache.get("key1"), Some(&b"value1".to_vec()));
    }

    #[test]
    fn test_ttl_expiration() {
        let mut cache = CacheEngine::new(100, 1);
        assert!(cache.set("key1".into(), b"value1".to_vec(), Some(1)));
        assert_eq!(cache.get("key1"), Some(&b"value1".to_vec()));
        sleep(Duration::from_secs(2));
        assert_eq!(cache.get("key1"), None);
    }

    #[test]
    fn test_lru_eviction() {
        let mut cache = CacheEngine::new(2, 300);
        cache.set("k1".into(), b"v1".to_vec(), None);
        cache.set("k2".into(), b"v2".to_vec(), None);
        cache.get("k1");
        cache.set("k3".into(), b"v3".to_vec(), None);
        assert_eq!(cache.get("k2"), None);
        assert_eq!(cache.get("k1"), Some(&b"v1".to_vec()));
        assert_eq!(cache.get("k3"), Some(&b"v3".to_vec()));
    }

    #[test]
    fn test_size_limits() {
        let mut cache = CacheEngine::new(100, 300);
        let long_key = "x".repeat(257);
        assert!(!cache.set(long_key, b"v".to_vec(), None));
        let large_value = vec![0u8; 1_048_577];
        assert!(!cache.set("k".into(), large_value, None));
    }

    #[test]
    fn test_delete_by_prefix() {
        let mut cache = CacheEngine::new(100, 300);
        cache.set("user:1".into(), b"a".to_vec(), None);
        cache.set("user:2".into(), b"b".to_vec(), None);
        cache.set("product:1".into(), b"c".to_vec(), None);
        assert_eq!(cache.delete_by_prefix("user:"), 2);
        assert_eq!(cache.entry_count(), 1);
    }

    #[test]
    fn test_namespace_stats() {
        let mut cache = CacheEngine::new(100, 300);
        cache.set("ns1:k1".into(), b"v1".to_vec(), None);
        cache.set("ns2:k1".into(), b"v2".to_vec(), None);
        cache.get("ns1:k1");
        cache.get("ns1:k1");
        cache.get("ns1:missing");
        cache.get("ns2:missing");

        let stats = cache.stats();
        assert_eq!(stats.total_queries, 4);
        assert_eq!(stats.cache_hits, 2);
        assert_eq!(stats.namespace_hit_rate("ns1"), 2.0 / 3.0);
        assert_eq!(stats.namespace_hit_rate("ns2"), 0.0);
    }

    #[test]
    fn test_capacity_dynamic() {
        let mut cache = CacheEngine::new(10, 300);
        for i in 0..10 {
            cache.set(format!("k{}", i), vec![i as u8], None);
        }
        assert_eq!(cache.entry_count(), 10);
        cache.set_capacity(3);
        assert_eq!(cache.entry_count(), 3);
    }
}
