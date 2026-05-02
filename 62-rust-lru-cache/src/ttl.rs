use std::collections::{BTreeMap, HashMap};
use std::hash::Hash;
use std::time::{Duration, Instant};

pub struct TtlManager<K: Eq + Hash + Clone + Ord> {
    key_to_expiry: HashMap<K, Instant>,
    expiry_to_keys: BTreeMap<Instant, Vec<K>>,
}

impl<K: Eq + Hash + Clone + Ord> TtlManager<K> {
    pub fn new() -> Self {
        Self {
            key_to_expiry: HashMap::new(),
            expiry_to_keys: BTreeMap::new(),
        }
    }

    pub fn with_capacity(capacity: usize) -> Self {
        Self {
            key_to_expiry: HashMap::with_capacity(capacity),
            expiry_to_keys: BTreeMap::new(),
        }
    }

    pub fn set(&mut self, key: K, ttl: Duration) -> Instant {
        let expiry = Instant::now() + ttl;
        let key_clone = key.clone();

        if let Some(old_expiry) = self.key_to_expiry.insert(key, expiry) {
            self.remove_from_expiry_map(&old_expiry, &key_clone);
        }

        self.expiry_to_keys
            .entry(expiry)
            .or_default()
            .push(key_clone);

        expiry
    }

    pub fn refresh(&mut self, key: &K, ttl: Duration) -> Option<Instant> {
        let key_clone = key.clone();
        let old_expiry = self.key_to_expiry.get(key).cloned()?;
        
        let new_expiry = Instant::now() + ttl;
        
        self.remove_from_expiry_map(&old_expiry, key);
        self.key_to_expiry.insert(key_clone.clone(), new_expiry);
        self.expiry_to_keys
            .entry(new_expiry)
            .or_default()
            .push(key_clone);

        Some(new_expiry)
    }

    pub fn is_expired(&self, key: &K) -> bool {
        self.key_to_expiry
            .get(key)
            .map(|&expiry| expiry <= Instant::now())
            .unwrap_or(false)
    }

    pub fn get_expiry(&self, key: &K) -> Option<Instant> {
        self.key_to_expiry.get(key).cloned()
    }

    pub fn remove(&mut self, key: &K) -> bool {
        if let Some(expiry) = self.key_to_expiry.remove(key) {
            self.remove_from_expiry_map(&expiry, key);
            true
        } else {
            false
        }
    }

    pub fn collect_expired(&mut self) -> Vec<K> {
        let now = Instant::now();
        let mut expired_keys = Vec::new();

        let expired_times: Vec<Instant> = self
            .expiry_to_keys
            .range(..=now)
            .map(|(&time, _)| time)
            .collect();

        for time in expired_times {
            if let Some(keys) = self.expiry_to_keys.remove(&time) {
                for key in keys {
                    self.key_to_expiry.remove(&key);
                    expired_keys.push(key);
                }
            }
        }

        expired_keys
    }

    pub fn len(&self) -> usize {
        self.key_to_expiry.len()
    }

    pub fn is_empty(&self) -> bool {
        self.key_to_expiry.is_empty()
    }

    pub fn count_active(&self) -> usize {
        let now = Instant::now();
        self.expiry_to_keys
            .range(now..)
            .map(|(_, keys)| keys.len())
            .sum()
    }

    fn remove_from_expiry_map(&mut self, expiry: &Instant, key: &K) {
        if let Some(keys) = self.expiry_to_keys.get_mut(expiry) {
            keys.retain(|k| k != key);
            if keys.is_empty() {
                self.expiry_to_keys.remove(expiry);
            }
        }
    }
}

impl<K: Eq + Hash + Clone + Ord> Default for TtlManager<K> {
    fn default() -> Self {
        Self::new()
    }
}
