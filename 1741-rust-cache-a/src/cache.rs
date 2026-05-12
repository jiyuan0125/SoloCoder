use std::collections::HashMap;
use std::time::{Duration, Instant};
use lru::LruCache;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CacheEntry {
    pub value: String,
    pub ttl: Option<Duration>,
    #[serde(skip)]
    pub created_at: Option<Instant>,
}

#[derive(Debug, Clone, Serialize)]
pub struct CacheStats {
    pub total_entries: usize,
    pub hits: u64,
    pub misses: u64,
    pub evictions: u64,
}

pub struct Cache {
    max_size: usize,
    entries: LruCache<String, CacheEntry>,
    stats: CacheStats,
}

impl Cache {
    pub fn new(max_size: usize) -> Self {
        Self {
            max_size,
            entries: LruCache::unbounded(),
            stats: CacheStats {
                total_entries: 0,
                hits: 0,
                misses: 0,
                evictions: 0,
            },
        }
    }

    pub fn set(&mut self, key: String, entry: CacheEntry) {
        let entry = CacheEntry {
            created_at: Some(Instant::now()),
            ..entry
        };

        if self.entries.len() >= self.max_size && !self.entries.contains(&key) {
            self.entries.pop_lru();
            self.stats.evictions += 1;
            self.stats.total_entries -= 1;
        }

        if self.entries.put(key, entry).is_none() {
            self.stats.total_entries += 1;
        }
    }

    pub fn get(&mut self, key: &str) -> Option<&CacheEntry> {
        let should_remove = if let Some(entry) = self.entries.peek(key) {
            self.is_expired(entry)
        } else {
            false
        };

        if should_remove {
            self.entries.pop(key);
            self.stats.total_entries -= 1;
            self.stats.misses += 1;
            return None;
        }

        let entry = self.entries.get(key);
        if entry.is_some() {
            self.stats.hits += 1;
        } else {
            self.stats.misses += 1;
        }
        entry
    }

    pub fn delete(&mut self, key: &str) -> bool {
        let removed = self.entries.pop(key);
        if removed.is_some() {
            self.stats.total_entries -= 1;
        }
        removed.is_some()
    }

    pub fn batch_set(&mut self, entries: HashMap<String, CacheEntry>) -> (Vec<String>, Vec<String>) {
        let mut success = Vec::new();
        let failed = Vec::new();

        for (key, entry) in entries {
            self.set(key.clone(), entry);
            success.push(key);
        }

        (success, failed)
    }

    pub fn batch_get(&mut self, keys: &[String]) -> (HashMap<String, String>, Vec<String>) {
        let mut found = HashMap::new();
        let mut not_found = Vec::new();

        for key in keys {
            match self.get(key) {
                Some(entry) => {
                    found.insert(key.clone(), entry.value.clone());
                }
                None => {
                    not_found.push(key.clone());
                }
            }
        }

        (found, not_found)
    }

    pub fn stats(&self) -> CacheStats {
        self.stats.clone()
    }

    pub fn import(&mut self, entries: HashMap<String, CacheEntry>) -> usize {
        let mut count = 0;
        for (key, entry) in entries {
            self.set(key, entry);
            count += 1;
        }
        count
    }

    fn is_expired(&self, entry: &CacheEntry) -> bool {
        if let (Some(ttl), Some(created_at)) = (entry.ttl, entry.created_at) {
            created_at.elapsed() > ttl
        } else {
            false
        }
    }
}
