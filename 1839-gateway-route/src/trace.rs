use std::collections::HashMap;
use std::sync::Arc;

use serde::{Deserialize, Serialize};
use tokio::sync::RwLock;
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TraceRecord {
    pub trace_id: Uuid,
    pub service_name: String,
    pub status_code: u16,
    pub duration_ms: u64,
    pub timestamp: i64,
}

#[derive(Debug, Clone)]
pub struct TraceStore {
    by_trace_id: Arc<RwLock<HashMap<Uuid, TraceRecord>>>,
    by_service: Arc<RwLock<HashMap<String, Vec<TraceRecord>>>>,
}

impl TraceStore {
    pub fn new() -> Self {
        Self {
            by_trace_id: Arc::new(RwLock::new(HashMap::new())),
            by_service: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    pub async fn record(&self, record: TraceRecord) {
        let trace_id = record.trace_id;
        let service_name = record.service_name.clone();

        let mut by_id = self.by_trace_id.write().await;
        by_id.insert(trace_id, record.clone());

        let mut by_service = self.by_service.write().await;
        by_service
            .entry(service_name)
            .or_insert_with(Vec::new)
            .push(record);
    }

    pub async fn get_by_trace_id(&self, trace_id: Uuid) -> Option<TraceRecord> {
        let by_id = self.by_trace_id.read().await;
        by_id.get(&trace_id).cloned()
    }

    pub async fn get_by_service(&self, service_name: &str) -> Vec<TraceRecord> {
        let by_service = self.by_service.read().await;
        by_service.get(service_name).cloned().unwrap_or_default()
    }
}
