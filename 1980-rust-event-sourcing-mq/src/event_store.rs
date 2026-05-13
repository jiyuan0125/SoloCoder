use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;
use chrono::Utc;
use uuid::Uuid;

use crate::models::{Event, EventQuery, PaginatedEvents};

#[derive(Debug, Clone)]
pub struct EventStore {
    store: Arc<RwLock<HashMap<Uuid, Vec<Event>>>>,
}

impl EventStore {
    pub fn new() -> Self {
        EventStore {
            store: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    pub async fn append_event(
        &self,
        aggregate_id: Uuid,
        event_type: String,
        payload: serde_json::Value,
        expected_version: u64,
    ) -> Result<Event, u64> {
        let mut store = self.store.write().await;
        let events = store.entry(aggregate_id).or_insert_with(Vec::new);

        let current_version = events.last().map(|e| e.version).unwrap_or(0);

        if expected_version != current_version {
            return Err(current_version);
        }

        let new_version = current_version + 1;
        let event = Event {
            event_id: Uuid::new_v4(),
            aggregate_id,
            event_type,
            payload,
            timestamp: Utc::now(),
            version: new_version,
        };

        events.push(event.clone());
        Ok(event)
    }

    pub async fn get_events(
        &self,
        aggregate_id: Uuid,
        query: EventQuery,
    ) -> PaginatedEvents {
        let store = self.store.read().await;
        let events = store.get(&aggregate_id).cloned().unwrap_or_default();

        let filtered: Vec<Event> = events
            .into_iter()
            .filter(|e| {
                if let Some(ref event_type) = query.event_type {
                    if &e.event_type != event_type {
                        return false;
                    }
                }
                if let Some(start) = query.start_time {
                    if e.timestamp < start {
                        return false;
                    }
                }
                if let Some(end) = query.end_time {
                    if e.timestamp > end {
                        return false;
                    }
                }
                true
            })
            .collect();

        let total = filtered.len() as u64;
        let offset = query.offset.unwrap_or(0);
        let limit = query.limit.unwrap_or(100);

        let paginated: Vec<Event> = filtered
            .into_iter()
            .skip(offset as usize)
            .take(limit as usize)
            .collect();

        PaginatedEvents {
            events: paginated,
            total,
            offset,
            limit,
        }
    }

    pub async fn get_events_from_version(
        &self,
        aggregate_id: Uuid,
        from_version: u64,
    ) -> Vec<Event> {
        let store = self.store.read().await;
        let events = store.get(&aggregate_id).cloned().unwrap_or_default();

        events
            .into_iter()
            .filter(|e| e.version > from_version)
            .collect()
    }

    pub async fn get_events_to_version(
        &self,
        aggregate_id: Uuid,
        to_version: u64,
        from_version: Option<u64>,
    ) -> Vec<Event> {
        let store = self.store.read().await;
        let events = store.get(&aggregate_id).cloned().unwrap_or_default();

        events
            .into_iter()
            .filter(|e| {
                if let Some(from) = from_version {
                    e.version > from && e.version <= to_version
                } else {
                    e.version <= to_version
                }
            })
            .collect()
    }

    pub async fn get_current_version(&self, aggregate_id: Uuid) -> u64 {
        let store = self.store.read().await;
        store
            .get(&aggregate_id)
            .and_then(|events| events.last().map(|e| e.version))
            .unwrap_or(0)
    }

    pub async fn get_event_count_since_version(
        &self,
        aggregate_id: Uuid,
        since_version: u64,
    ) -> u64 {
        let store = self.store.read().await;
        let events = store.get(&aggregate_id).cloned().unwrap_or_default();

        events
            .iter()
            .filter(|e| e.version > since_version)
            .count() as u64
    }

    pub async fn restore_events(&self, aggregate_id: Uuid, events: Vec<Event>) {
        let mut store = self.store.write().await;
        store.insert(aggregate_id, events);
    }
}
