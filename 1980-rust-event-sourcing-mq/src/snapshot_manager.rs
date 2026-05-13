use std::fs;
use std::path::PathBuf;
use std::sync::Arc;
use tokio::sync::RwLock;
use uuid::Uuid;

use crate::models::Snapshot;

#[derive(Debug, Clone)]
pub struct SnapshotManager {
    snapshot_dir: PathBuf,
    snapshots: Arc<RwLock<std::collections::HashMap<Uuid, Snapshot>>>,
}

impl SnapshotManager {
    pub fn new(snapshot_dir: &str) -> Self {
        let path = PathBuf::from(snapshot_dir);
        if !path.exists() {
            fs::create_dir_all(&path).expect("Failed to create snapshot directory");
        }

        SnapshotManager {
            snapshot_dir: path,
            snapshots: Arc::new(RwLock::new(std::collections::HashMap::new())),
        }
    }

    pub async fn save_snapshot(&self, snapshot: &Snapshot) -> Result<(), String> {
        let file_path = self.get_snapshot_file_path(snapshot.aggregate_id);

        let json = serde_json::to_string_pretty(snapshot)
            .map_err(|e| format!("Failed to serialize snapshot: {}", e))?;

        fs::write(&file_path, json)
            .map_err(|e| format!("Failed to write snapshot file: {}", e))?;

        let mut snapshots = self.snapshots.write().await;
        snapshots.insert(snapshot.aggregate_id, snapshot.clone());

        Ok(())
    }

    pub async fn get_latest_snapshot(&self, aggregate_id: Uuid) -> Option<Snapshot> {
        let snapshots = self.snapshots.read().await;
        snapshots.get(&aggregate_id).cloned()
    }

    pub async fn load_from_disk(&self) -> Vec<Snapshot> {
        let mut loaded_snapshots = Vec::new();

        if let Ok(entries) = fs::read_dir(&self.snapshot_dir) {
            for entry in entries.flatten() {
                let path = entry.path();
                if let Some(ext) = path.extension() {
                    if ext == "json" {
                        if let Ok(content) = fs::read_to_string(&path) {
                            if let Ok(snapshot) = serde_json::from_str::<Snapshot>(&content) {
                                loaded_snapshots.push(snapshot);
                            }
                        }
                    }
                }
            }
        }

        loaded_snapshots
    }

    pub async fn load_all_snapshots(&self) -> Vec<Snapshot> {
        let loaded = self.load_from_disk().await;
        let mut snapshots = self.snapshots.write().await;

        for snapshot in &loaded {
            snapshots.insert(snapshot.aggregate_id, snapshot.clone());
        }

        loaded
    }

    fn get_snapshot_file_path(&self, aggregate_id: Uuid) -> PathBuf {
        let filename = format!("{}.json", aggregate_id);
        self.snapshot_dir.join(filename)
    }
}
