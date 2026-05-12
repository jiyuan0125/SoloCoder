use serde::{Deserialize, Serialize};
use std::time::Instant;

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(untagged)]
pub enum CacheValue {
    String(String),
    Integer(i64),
    Float(f64),
    Boolean(bool),
    Object(serde_json::Value),
}

impl From<String> for CacheValue {
    fn from(s: String) -> Self {
        CacheValue::String(s)
    }
}

impl From<i64> for CacheValue {
    fn from(n: i64) -> Self {
        CacheValue::Integer(n)
    }
}

impl From<f64> for CacheValue {
    fn from(n: f64) -> Self {
        CacheValue::Float(n)
    }
}

impl From<bool> for CacheValue {
    fn from(b: bool) -> Self {
        CacheValue::Boolean(b)
    }
}

impl From<serde_json::Value> for CacheValue {
    fn from(v: serde_json::Value) -> Self {
        CacheValue::Object(v)
    }
}

#[derive(Clone, Debug)]
pub struct CacheEntry {
    pub key: String,
    pub value: CacheValue,
    pub created_at: Instant,
    pub expire_at: Option<Instant>,
}

impl CacheEntry {
    pub fn new(key: String, value: CacheValue, ttl_seconds: Option<u64>) -> Self {
        let now = Instant::now();
        let expire_at = ttl_seconds.map(|ttl| now + std::time::Duration::from_secs(ttl));

        CacheEntry {
            key,
            value,
            created_at: now,
            expire_at,
        }
    }

    pub fn is_expired(&self) -> bool {
        self.expire_at.map_or(false, |expire| expire <= Instant::now())
    }

    pub fn remaining_ttl(&self) -> Option<u64> {
        self.expire_at.and_then(|expire| {
            let now = Instant::now();
            if expire > now {
                Some((expire - now).as_secs())
            } else {
                Some(0)
            }
        })
    }
}
