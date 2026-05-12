use std::sync::Arc;
use std::time::Duration;
use dashmap::DashMap;
use chrono::Utc;
use crate::models::{RequestRecord, RequestStatus, ResponseData, RetryHistory};

#[derive(Clone)]
pub struct IdempotencyStore {
    store: Arc<DashMap<String, RequestRecord>>,
    ttl: Duration,
}

impl IdempotencyStore {
    pub fn new(ttl_hours: u64) -> Self {
        let store = Arc::new(DashMap::new());
        let ttl = Duration::from_secs(ttl_hours * 3600);
        let store_clone = store.clone();
        let ttl_clone = ttl;
        
        tokio::spawn(async move {
            let mut interval = tokio::time::interval(Duration::from_secs(60));
            loop {
                interval.tick().await;
                clean_expired(&store_clone, ttl_clone);
            }
        });
        
        Self { store, ttl }
    }
    
    pub fn get(&self, key: &str) -> Option<RequestRecord> {
        self.store.get(key).map(|entry| entry.clone())
    }
    
    pub fn insert(&self, record: RequestRecord) {
        self.store.insert(record.idempotency_key.clone(), record);
    }
    
    pub fn update_response(&self, key: &str, response: ResponseData, status: RequestStatus) {
        if let Some(mut entry) = self.store.get_mut(key) {
            let now = Utc::now();
            entry.response = Some(response);
            entry.status = status;
            entry.completed_at = Some(now);
            entry.expires_at = now + chrono::Duration::from_std(self.ttl).unwrap();
        }
    }
    
    pub fn update_status(&self, key: &str, status: RequestStatus) {
        if let Some(mut entry) = self.store.get_mut(key) {
            entry.status = status;
            if matches!(status, RequestStatus::Success | RequestStatus::Failed) {
                let now = Utc::now();
                entry.completed_at = Some(now);
                entry.expires_at = now + chrono::Duration::from_std(self.ttl).unwrap();
            }
        }
    }
    
    pub fn increment_retry(&self, key: &str) -> u32 {
        if let Some(mut entry) = self.store.get_mut(key) {
            entry.total_retries += 1;
            entry.total_retries
        } else {
            0
        }
    }
    
    pub fn get_all(&self) -> Vec<RequestRecord> {
        self.store.iter().map(|entry| entry.clone()).collect()
    }
    
    pub fn add_retry_history(&self, key: &str, history: RetryHistory) {
        if let Some(mut entry) = self.store.get_mut(key) {
            entry.retry_history.push(history);
        }
    }
}

fn clean_expired(store: &DashMap<String, RequestRecord>, _ttl: Duration) {
    let now = Utc::now();
    store.retain(|_, record| record.expires_at > now);
}

pub fn new_idempotency_store() -> IdempotencyStore {
    IdempotencyStore::new(24)
}
