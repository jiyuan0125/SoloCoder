use std::time::Duration;

#[derive(Debug, Clone)]
pub struct SnapshotConfig {
    pub events_threshold: u64,
    pub time_interval_secs: u64,
}

impl Default for SnapshotConfig {
    fn default() -> Self {
        SnapshotConfig {
            events_threshold: 1000,
            time_interval_secs: 3600,
        }
    }
}

impl SnapshotConfig {
    pub fn from_env() -> Self {
        let events_threshold = std::env::var("SNAPSHOT_EVENTS_THRESHOLD")
            .ok()
            .and_then(|v| v.parse().ok())
            .unwrap_or(1000);

        let time_interval_secs = std::env::var("SNAPSHOT_TIME_INTERVAL_SECS")
            .ok()
            .and_then(|v| v.parse().ok())
            .unwrap_or(3600);

        SnapshotConfig {
            events_threshold,
            time_interval_secs,
        }
    }

    pub fn get_time_interval(&self) -> Duration {
        Duration::from_secs(self.time_interval_secs)
    }
}

#[derive(Debug, Clone)]
pub struct AppConfig {
    pub port: u16,
    pub snapshot_dir: String,
    pub snapshot_config: SnapshotConfig,
}

impl Default for AppConfig {
    fn default() -> Self {
        AppConfig {
            port: 8702,
            snapshot_dir: "./snapshots".to_string(),
            snapshot_config: SnapshotConfig::default(),
        }
    }
}

impl AppConfig {
    pub fn from_env() -> Self {
        let port = std::env::var("PORT")
            .ok()
            .and_then(|v| v.parse().ok())
            .unwrap_or(3000);

        let snapshot_dir = std::env::var("SNAPSHOT_DIR")
            .unwrap_or_else(|_| "./snapshots".to_string());

        AppConfig {
            port,
            snapshot_dir,
            snapshot_config: SnapshotConfig::from_env(),
        }
    }
}
