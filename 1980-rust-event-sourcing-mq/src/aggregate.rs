use chrono::{DateTime, Utc};
use serde_json::Value;
use std::sync::Arc;
use tokio::sync::RwLock;
use uuid::Uuid;

use crate::config::{AppConfig, SnapshotConfig};
use crate::event_store::EventStore;
use crate::models::{AggregateState, Event, Snapshot};
use crate::snapshot_manager::SnapshotManager;

#[derive(Debug, Clone)]
pub struct AggregateManager {
    event_store: EventStore,
    snapshot_manager: SnapshotManager,
    snapshot_config: SnapshotConfig,
    last_snapshot_time: Arc<RwLock<std::collections::HashMap<Uuid, DateTime<Utc>>>>,
}

impl AggregateManager {
    pub fn new(event_store: EventStore, snapshot_manager: SnapshotManager, config: &AppConfig) -> Self {
        AggregateManager {
            event_store,
            snapshot_manager,
            snapshot_config: config.snapshot_config.clone(),
            last_snapshot_time: Arc::new(RwLock::new(std::collections::HashMap::new())),
        }
    }

    pub async fn get_state(&self, aggregate_id: Uuid) -> AggregateState {
        let snapshot = self.snapshot_manager.get_latest_snapshot(aggregate_id).await;

        let (from_snapshot_version, mut state, events) = if let Some(snap) = snapshot {
            (
                Some(snap.version),
                snap.state,
                self.event_store
                    .get_events_from_version(aggregate_id, snap.version)
                    .await,
            )
        } else {
            (
                None,
                Value::Object(serde_json::Map::new()),
                self.event_store
                    .get_events_from_version(aggregate_id, 0)
                    .await,
            )
        };

        let events_replayed = events.len() as u64;
        let version = events.last().map(|e| e.version).unwrap_or(from_snapshot_version.unwrap_or(0));

        for event in events {
            state = apply_event(&state, &event);
        }

        AggregateState {
            aggregate_id,
            state,
            version,
            from_snapshot_version,
            events_replayed,
        }
    }

    pub async fn get_state_at_version(&self, aggregate_id: Uuid, target_version: u64) -> AggregateState {
        let current_version = self.event_store.get_current_version(aggregate_id).await;

        if target_version > current_version {
            return self.get_state(aggregate_id).await;
        }

        let all_snapshots = self.snapshot_manager.load_from_disk().await;
        let relevant_snapshots: Vec<_> = all_snapshots
            .into_iter()
            .filter(|s| s.aggregate_id == aggregate_id && s.version <= target_version)
            .collect();

        let latest_snapshot = relevant_snapshots
            .into_iter()
            .max_by_key(|s| s.version);

        let (from_snapshot_version, mut state, events) = if let Some(snap) = latest_snapshot {
            (
                Some(snap.version),
                snap.state,
                self.event_store
                    .get_events_to_version(aggregate_id, target_version, Some(snap.version))
                    .await,
            )
        } else {
            (
                None,
                Value::Object(serde_json::Map::new()),
                self.event_store
                    .get_events_to_version(aggregate_id, target_version, None)
                    .await,
            )
        };

        let events_replayed = events.len() as u64;
        let version = events.last().map(|e| e.version).unwrap_or(from_snapshot_version.unwrap_or(0));

        for event in events {
            state = apply_event(&state, &event);
        }

        AggregateState {
            aggregate_id,
            state,
            version,
            from_snapshot_version,
            events_replayed,
        }
    }

    pub async fn append_event(
        &self,
        aggregate_id: Uuid,
        event_type: String,
        payload: Value,
        expected_version: u64,
    ) -> Result<Event, u64> {
        let event = self.event_store.append_event(aggregate_id, event_type, payload, expected_version).await?;

        self.check_and_create_snapshot(aggregate_id).await;

        Ok(event)
    }

    async fn check_and_create_snapshot(&self, aggregate_id: Uuid) {
        let snapshot = self.snapshot_manager.get_latest_snapshot(aggregate_id).await;
        let last_snapshot_version = snapshot.as_ref().map(|s| s.version).unwrap_or(0);
        let events_since_snapshot = self.event_store.get_event_count_since_version(aggregate_id, last_snapshot_version).await;

        let should_snapshot_by_count = events_since_snapshot >= self.snapshot_config.events_threshold;

        let last_snapshot_time = {
            let times = self.last_snapshot_time.read().await;
            times.get(&aggregate_id).copied()
        };

        let now = Utc::now();
        let should_snapshot_by_time = match last_snapshot_time {
            None => true,
            Some(last_time) => {
                let elapsed = (now - last_time).num_seconds() as u64;
                elapsed >= self.snapshot_config.time_interval_secs
            }
        };

        if should_snapshot_by_count || should_snapshot_by_time {
            let state = self.get_state(aggregate_id).await;
            let events_count = self.event_store.get_event_count_since_version(aggregate_id, last_snapshot_version).await;

            let new_snapshot = Snapshot {
                snapshot_id: Uuid::new_v4(),
                aggregate_id,
                state: state.state,
                version: state.version,
                timestamp: now,
                events_count,
            };

            if self.snapshot_manager.save_snapshot(&new_snapshot).await.is_ok() {
                let mut times = self.last_snapshot_time.write().await;
                times.insert(aggregate_id, now);
            }
        }
    }

    pub fn event_store(&self) -> &EventStore {
        &self.event_store
    }

    pub fn snapshot_manager(&self) -> &SnapshotManager {
        &self.snapshot_manager
    }
}

fn apply_event(current_state: &Value, event: &Event) -> Value {
    let mut new_state = current_state.clone();

    if let Value::Object(ref mut obj) = new_state {
        obj.insert("last_event_type".to_string(), Value::String(event.event_type.clone()));
        obj.insert("last_event_id".to_string(), Value::String(event.event_id.to_string()));
        obj.insert("last_event_timestamp".to_string(), Value::String(event.timestamp.to_rfc3339()));

        if let Value::Object(ref payload_obj) = event.payload {
            for (key, value) in payload_obj {
                obj.insert(key.clone(), value.clone());
            }
        }
    }

    new_state
}
