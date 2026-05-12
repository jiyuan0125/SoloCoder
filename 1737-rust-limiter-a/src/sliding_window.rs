use std::collections::HashMap;
use std::time::{SystemTime, UNIX_EPOCH};
use parking_lot::RwLock;

const DEFAULT_WINDOW_SECS: u64 = 60;
const DEFAULT_LIMIT: u64 = 60;

#[derive(Debug, Clone)]
pub struct SlidingWindowConfig {
    pub window_secs: u64,
    pub limit: u64,
}

impl Default for SlidingWindowConfig {
    fn default() -> Self {
        SlidingWindowConfig {
            window_secs: DEFAULT_WINDOW_SECS,
            limit: DEFAULT_LIMIT,
        }
    }
}

struct IpCounter {
    slots: Vec<u64>,
    last_update_time: u64,
    total_count: u64,
}

impl IpCounter {
    fn new(window_secs: u64) -> Self {
        IpCounter {
            slots: vec![0; window_secs as usize],
            last_update_time: Self::current_second(),
            total_count: 0,
        }
    }

    fn current_second() -> u64 {
        SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap()
            .as_secs()
    }

    fn refresh_slots(&mut self, current_time: u64, window_secs: u64) {
        let elapsed = current_time.saturating_sub(self.last_update_time);
        if elapsed == 0 {
            return;
        }

        let window_usize = window_secs as usize;
        let shift = if elapsed >= window_secs {
            self.slots.iter().for_each(|&c| self.total_count -= c);
            self.slots.fill(0);
            self.last_update_time = current_time;
            return;
        } else {
            elapsed as usize
        };

        for i in 0..shift {
            let idx = (self.last_update_time as usize + i) % window_usize;
            self.total_count = self.total_count.saturating_sub(self.slots[idx]);
            self.slots[idx] = 0;
        }

        self.last_update_time = current_time;
    }

    fn increment(&mut self, window_secs: u64) -> bool {
        let current_time = Self::current_second();
        self.refresh_slots(current_time, window_secs);

        let idx = (current_time % window_secs) as usize;
        self.slots[idx] += 1;
        self.total_count += 1;
        true
    }

    fn check(&self, limit: u64) -> bool {
        self.total_count <= limit
    }
}

pub struct SlidingWindowLimiter {
    ip_counters: RwLock<HashMap<String, IpCounter>>,
    config: RwLock<SlidingWindowConfig>,
}

impl SlidingWindowLimiter {
    pub fn new() -> Self {
        SlidingWindowLimiter {
            ip_counters: RwLock::new(HashMap::new()),
            config: RwLock::new(SlidingWindowConfig::default()),
        }
    }

    pub fn update_config(&self, config: SlidingWindowConfig) {
        *self.config.write() = config;
    }

    pub fn get_config(&self) -> SlidingWindowConfig {
        self.config.read().clone()
    }

    pub fn try_acquire(&self, ip: &str) -> bool {
        let config = self.config.read().clone();
        let window_secs = config.window_secs;
        let limit = config.limit;

        {
            let counters = self.ip_counters.read();
            if let Some(counter) = counters.get(ip) {
                if !counter.check(limit) {
                    return false;
                }
            }
        }

        let mut counters = self.ip_counters.write();
        let counter = counters
            .entry(ip.to_string())
            .or_insert_with(|| IpCounter::new(window_secs));
        counter.increment(window_secs);
        counter.check(limit)
    }
}

impl Default for SlidingWindowLimiter {
    fn default() -> Self {
        Self::new()
    }
}
