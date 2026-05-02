use std::collections::HashMap;
use std::hash::Hash;
use std::sync::{Arc, RwLock, RwLockWriteGuard};
use std::time::Duration;

use crate::lru::LruList;
use crate::stats::Stats;
use crate::ttl::TtlManager;

struct CacheEntry<V> {
    value: V,
    ttl: Duration,
}

struct CacheInner<K, V>
where
    K: Eq + Hash + Clone + Ord,
{
    map: HashMap<K, CacheEntry<V>>,
    lru: LruList<K>,
    ttl: TtlManager<K>,
    capacity: usize,
}

pub struct Cache<K, V>
where
    K: Eq + Hash + Clone + Ord + Send + Sync + 'static,
    V: Clone + Send + Sync + 'static,
{
    inner: RwLock<CacheInner<K, V>>,
    stats: Arc<Stats>,
    on_evict: Option<Arc<dyn Fn(K, V) + Send + Sync + 'static>>,
}

impl<K, V> Cache<K, V>
where
    K: Eq + Hash + Clone + Ord + Send + Sync + 'static,
    V: Clone + Send + Sync + 'static,
{
    pub fn new(capacity: usize) -> Self {
        assert!(capacity > 0, "Capacity must be greater than 0");
        
        Self {
            inner: RwLock::new(CacheInner {
                map: HashMap::with_capacity(capacity),
                lru: LruList::with_capacity(capacity),
                ttl: TtlManager::with_capacity(capacity),
                capacity,
            }),
            stats: Arc::new(Stats::new()),
            on_evict: None,
        }
    }

    pub fn with_eviction_callback<F>(capacity: usize, on_evict: F) -> Self
    where
        F: Fn(K, V) + Send + Sync + 'static,
    {
        let mut cache = Self::new(capacity);
        cache.on_evict = Some(Arc::new(on_evict));
        cache
    }

    pub fn get(&self, key: &K) -> Option<V> {
        let mut inner = self.inner.write().unwrap();
        
        if !inner.map.contains_key(key) {
            self.stats.increment_miss();
            return None;
        }

        if inner.ttl.is_expired(key) {
            let (k, v) = self.remove_inner(&mut inner, key);
            self.stats.increment_eviction();
            self.stats.increment_miss();
            
            if let Some(ref callback) = self.on_evict {
                callback(k, v);
            }
            return None;
        }

        let entry = inner.map.get(key).unwrap();
        let value = entry.value.clone();
        let ttl = entry.ttl;

        inner.lru.access(key);
        inner.ttl.refresh(key, ttl);

        self.stats.increment_hit();
        Some(value)
    }

    pub fn put(&self, key: K, value: V, ttl: Duration) {
        let mut inner = self.inner.write().unwrap();
        let key_clone = key.clone();

        if inner.map.contains_key(&key) {
            let entry = inner.map.get_mut(&key).unwrap();
            entry.value = value;
            entry.ttl = ttl;
            inner.lru.access(&key);
            inner.ttl.set(key, ttl);
            return;
        }

        if inner.map.len() >= inner.capacity {
            self.evict_lru(&mut inner);
        }

        inner.map.insert(key_clone.clone(), CacheEntry { value, ttl });
        inner.lru.insert(key_clone.clone());
        inner.ttl.set(key_clone, ttl);
    }

    pub fn size(&self) -> usize {
        let inner = self.inner.read().unwrap();
        inner.map.len()
    }

    pub fn stats(&self) -> (usize, usize, usize, usize) {
        let current_size = self.size();
        self.stats.get(current_size)
    }

    fn evict_lru(&self, inner: &mut RwLockWriteGuard<CacheInner<K, V>>) {
        if let Some(key) = inner.lru.pop_lru() {
            let (k, v) = self.remove_inner(inner, &key);
            self.stats.increment_eviction();
            
            if let Some(ref callback) = self.on_evict {
                callback(k, v);
            }
        }
    }

    fn remove_inner(
        &self,
        inner: &mut RwLockWriteGuard<CacheInner<K, V>>,
        key: &K,
    ) -> (K, V) {
        let entry = inner.map.remove(key).unwrap();
        inner.lru.remove(key);
        inner.ttl.remove(key);
        (key.clone(), entry.value)
    }
}


