use std::collections::HashMap;
use std::sync::{Arc, Condvar, Mutex, RwLock};
use std::thread;
use std::time::{Duration, Instant};

use lru::LruCache;
use rand::Rng;

use crate::stats::Stats;

const L2_TTL: Duration = Duration::from_secs(300);
const L2_MIN_DELAY_MS: u64 = 1;
const L2_MAX_DELAY_MS: u64 = 10;

pub struct L1Cache {
    inner: LruCache<String, String>,
}

impl L1Cache {
    pub fn new(capacity: usize) -> Self {
        Self {
            inner: LruCache::new(std::num::NonZeroUsize::new(capacity).unwrap()),
        }
    }

    pub fn get(&mut self, key: &str) -> Option<String> {
        self.inner.get(key).cloned()
    }

    pub fn put(&mut self, key: String, value: String) {
        self.inner.put(key, value);
    }

    pub fn delete(&mut self, key: &str) {
        self.inner.pop(key);
    }

    pub fn len(&self) -> usize {
        self.inner.len()
    }
}

pub struct L2Entry {
    value: String,
    expire_at: Instant,
}

pub struct L2Cache {
    inner: HashMap<String, L2Entry>,
    capacity: usize,
}

impl L2Cache {
    pub fn new(capacity: usize) -> Self {
        Self {
            inner: HashMap::with_capacity(capacity),
            capacity,
        }
    }

    pub fn get(&mut self, key: &str) -> Option<String> {
        let now = Instant::now();
        let entry = self.inner.get(key);
        match entry {
            Some(e) if e.expire_at > now => Some(e.value.clone()),
            Some(_) => {
                self.inner.remove(key);
                None
            }
            None => None,
        }
    }

    pub fn put(&mut self, key: String, value: String) {
        if self.inner.len() >= self.capacity && !self.inner.contains_key(&key) {
            if let Some((k, _)) = self.inner.iter().next() {
                let k = k.clone();
                self.inner.remove(&k);
            }
        }
        let expire_at = Instant::now() + L2_TTL;
        self.inner.insert(key, L2Entry { value, expire_at });
    }

    pub fn delete(&mut self, key: &str) {
        self.inner.remove(key);
    }

    pub fn len(&self) -> usize {
        self.inner.len()
    }
}

fn simulate_network_delay() {
    let mut rng = rand::thread_rng();
    let delay_ms = rng.gen_range(L2_MIN_DELAY_MS..=L2_MAX_DELAY_MS);
    thread::sleep(Duration::from_millis(delay_ms));
}

struct InFlightState {
    done: bool,
    value: Option<String>,
}

pub struct TwoLevelCache {
    l1: Mutex<L1Cache>,
    l2: Mutex<L2Cache>,
    stats: RwLock<Stats>,
    in_flight: Mutex<HashMap<String, Arc<(Mutex<InFlightState>, Condvar)>>>,
}

impl TwoLevelCache {
    pub fn new(l1_capacity: usize, l2_capacity: usize) -> Self {
        Self {
            l1: Mutex::new(L1Cache::new(l1_capacity)),
            l2: Mutex::new(L2Cache::new(l2_capacity)),
            stats: RwLock::new(Stats::new()),
            in_flight: Mutex::new(HashMap::new()),
        }
    }

    pub fn get(&self, key: &str) -> Option<String> {
        {
            let mut l1 = self.l1.lock().unwrap();
            if let Some(value) = l1.get(key) {
                let mut stats = self.stats.write().unwrap();
                stats.l1_hits += 1;
                return Some(value);
            }
        }

        {
            let mut stats = self.stats.write().unwrap();
            stats.l1_misses += 1;
        }

        let key_owned = key.to_string();
        let (state_arc, is_leader) = {
            let mut in_flight = self.in_flight.lock().unwrap();
            if let Some(existing) = in_flight.get(&key_owned) {
                (existing.clone(), false)
            } else {
                let new_state = Arc::new((
                    Mutex::new(InFlightState {
                        done: false,
                        value: None,
                    }),
                    Condvar::new(),
                ));
                in_flight.insert(key_owned.clone(), new_state.clone());
                (new_state, true)
            }
        };

        let (lock, cvar) = &*state_arc;

        if !is_leader {
            let mut state = lock.lock().unwrap();
            while !state.done {
                state = cvar.wait(state).unwrap();
            }
            return state.value.clone();
        }

        let result = {
            simulate_network_delay();
            let mut l2 = self.l2.lock().unwrap();
            let l2_result = l2.get(key);
            drop(l2);

            if let Some(value) = l2_result {
                let mut l1 = self.l1.lock().unwrap();
                l1.put(key.to_string(), value.clone());
                let mut stats = self.stats.write().unwrap();
                stats.l2_hits += 1;
                stats.l1_miss_l2_hit += 1;
                Some(value)
            } else {
                let mut stats = self.stats.write().unwrap();
                stats.l2_misses += 1;
                None
            }
        };

        {
            let mut state = lock.lock().unwrap();
            state.done = true;
            state.value = result.clone();
            cvar.notify_all();
        }

        {
            let mut in_flight = self.in_flight.lock().unwrap();
            in_flight.remove(&key_owned);
        }

        result
    }

    pub fn put(&self, key: String, value: String) {
        {
            let mut l1 = self.l1.lock().unwrap();
            l1.put(key.clone(), value.clone());
        }

        simulate_network_delay();
        {
            let mut l2 = self.l2.lock().unwrap();
            l2.put(key, value);
        }
    }

    pub fn delete(&self, key: &str) {
        {
            let mut l1 = self.l1.lock().unwrap();
            l1.delete(key);
        }

        simulate_network_delay();
        {
            let mut l2 = self.l2.lock().unwrap();
            l2.delete(key);
        }
    }

    pub fn get_stats(&self) -> (Stats, usize, usize) {
        let stats = self.stats.read().unwrap().clone();
        let l1_size = self.l1.lock().unwrap().len();
        let l2_size = self.l2.lock().unwrap().len();
        (stats, l1_size, l2_size)
    }
}
