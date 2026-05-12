use std::collections::HashMap;
use std::sync::Arc;
use std::time::Duration;

use reqwest::Client;
use serde::{Deserialize, Serialize};
use tokio::sync::RwLock;
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Backend {
    pub id: Uuid,
    pub path_prefix: String,
    pub target_url: String,
    pub available: bool,
    pub last_checked: Option<i64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateBackendRequest {
    pub path_prefix: String,
    pub target_url: String,
}

#[derive(Debug, Clone, Serialize)]
pub struct Stats {
    pub backend_id: Uuid,
    pub path_prefix: String,
    pub target_url: String,
    pub request_count: u64,
}

#[derive(Debug, Clone, Default)]
pub struct Config {
    pub timeout_seconds: u64,
}

pub struct AppStateInner {
    pub backends: RwLock<HashMap<Uuid, Backend>>,
    pub request_counts: RwLock<HashMap<Uuid, u64>>,
    pub config: RwLock<Config>,
    pub http_client: Client,
}

pub type AppState = Arc<AppStateInner>;

impl AppStateInner {
    pub fn new() -> AppState {
        Arc::new(AppStateInner {
            backends: RwLock::new(HashMap::new()),
            request_counts: RwLock::new(HashMap::new()),
            config: RwLock::new(Config { timeout_seconds: 30 }),
            http_client: Client::builder()
                .connect_timeout(Duration::from_secs(5))
                .build()
                .expect("Failed to create HTTP client"),
        })
    }

    pub async fn add_backend(&self, req: CreateBackendRequest) -> Backend {
        let id = Uuid::new_v4();
        let backend = Backend {
            id,
            path_prefix: normalize_path_prefix(&req.path_prefix),
            target_url: normalize_target_url(&req.target_url),
            available: true,
            last_checked: None,
        };
        self.backends.write().await.insert(id, backend.clone());
        self.request_counts.write().await.insert(id, 0);
        backend
    }

    pub async fn remove_backend(&self, id: Uuid) -> bool {
        self.backends.write().await.remove(&id).is_some()
            && self.request_counts.write().await.remove(&id).is_some()
    }

    pub async fn get_backends(&self) -> Vec<Backend> {
        self.backends.read().await.values().cloned().collect()
    }

    pub async fn find_backend_for_path(&self, path: &str) -> Option<Backend> {
        let backends = self.backends.read().await;
        let mut best_match: Option<Backend> = None;
        let mut best_match_len = 0;

        for backend in backends.values() {
            if !backend.available {
                continue;
            }
            if path.starts_with(&backend.path_prefix) && backend.path_prefix.len() > best_match_len {
                best_match = Some(backend.clone());
                best_match_len = backend.path_prefix.len();
            }
        }

        best_match
    }

    pub async fn increment_request_count(&self, backend_id: Uuid) {
        let mut counts = self.request_counts.write().await;
        *counts.entry(backend_id).or_insert(0) += 1;
    }

    pub async fn get_stats(&self) -> Vec<Stats> {
        let backends = self.backends.read().await;
        let counts = self.request_counts.read().await;
        backends
            .values()
            .map(|b| Stats {
                backend_id: b.id,
                path_prefix: b.path_prefix.clone(),
                target_url: b.target_url.clone(),
                request_count: *counts.get(&b.id).unwrap_or(&0),
            })
            .collect()
    }

    pub async fn update_timeout(&self, seconds: u64) {
        let mut config = self.config.write().await;
        config.timeout_seconds = seconds;
    }

    pub async fn get_timeout(&self) -> Duration {
        let config = self.config.read().await;
        Duration::from_secs(config.timeout_seconds)
    }

    pub async fn check_health(&self) {
        let backend_ids: Vec<Uuid> = self.backends.read().await.keys().cloned().collect();

        for id in backend_ids {
            if let Some(backend) = self.backends.read().await.get(&id).cloned() {
                let health_url = format!("{}/health", backend.target_url.trim_end_matches('/'));
                let available = match self.http_client.get(&health_url).send().await {
                    Ok(resp) => resp.status().is_success(),
                    Err(_) => false,
                };

                if let Some(b) = self.backends.write().await.get_mut(&id) {
                    b.available = available;
                    b.last_checked = Some(chrono::Utc::now().timestamp());
                }
            }
        }
    }
}

fn normalize_path_prefix(prefix: &str) -> String {
    let mut normalized = prefix.trim().to_string();
    if !normalized.starts_with('/') {
        normalized.insert(0, '/');
    }
    if !normalized.ends_with('/') {
        normalized.push('/');
    }
    normalized
}

fn normalize_target_url(url: &str) -> String {
    url.trim().trim_end_matches('/').to_string()
}
