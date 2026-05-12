use crate::backend::BackendInternal;
use crate::config::Config;
use crate::stats::{BackendStats, StatsSnapshot};
use reqwest::Client;
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock as AsyncRwLock;

#[derive(Clone)]
pub struct AppState {
    pub backends: Arc<AsyncRwLock<Vec<Arc<BackendInternal>>>>,
    pub stats: Arc<AsyncRwLock<HashMap<String, Arc<BackendStats>>>>,
    pub config: Config,
    pub http_client: Client,
}

impl AppState {
    pub fn new(config: Config) -> Self {
        AppState {
            backends: Arc::new(AsyncRwLock::new(Vec::new())),
            stats: Arc::new(AsyncRwLock::new(HashMap::new())),
            config,
            http_client: Client::builder()
                .connect_timeout(std::time::Duration::from_secs(10))
                .build()
                .expect("Failed to build HTTP client"),
        }
    }

    pub async fn add_backend(&self, path_prefix: String, target_url: String) -> Arc<BackendInternal> {
        let backend = Arc::new(BackendInternal::new(path_prefix, target_url));
        let stats = Arc::new(BackendStats::new());
        
        let mut backends = self.backends.write().await;
        backends.push(backend.clone());
        
        let mut stats_map = self.stats.write().await;
        stats_map.insert(backend.id.clone(), stats);
        
        backend
    }

    pub async fn remove_backend(&self, id: &str) -> bool {
        let mut backends = self.backends.write().await;
        let initial_len = backends.len();
        backends.retain(|b| b.id != id);
        
        let mut stats_map = self.stats.write().await;
        stats_map.remove(id);
        
        backends.len() < initial_len
    }

    pub async fn find_backend(&self, path: &str) -> Option<Arc<BackendInternal>> {
        let backends = self.backends.read().await;
        let mut best_match: Option<Arc<BackendInternal>> = None;
        let mut best_len = 0;

        for backend in backends.iter() {
            if backend.is_healthy() && path.starts_with(&backend.path_prefix) {
                if backend.path_prefix.len() > best_len {
                    best_len = backend.path_prefix.len();
                    best_match = Some(backend.clone());
                }
            }
        }

        best_match
    }

    pub async fn get_stats(&self, backend_id: &str) -> Option<Arc<BackendStats>> {
        let stats = self.stats.read().await;
        stats.get(backend_id).cloned()
    }

    pub async fn get_all_stats(&self) -> Vec<StatsSnapshot> {
        let backends = self.backends.read().await;
        let stats = self.stats.read().await;
        let mut result = Vec::new();

        for backend in backends.iter() {
            if let Some(stat) = stats.get(&backend.id) {
                result.push(StatsSnapshot {
                    backend_id: backend.id.clone(),
                    path_prefix: backend.path_prefix.clone(),
                    requests: stat.requests.load(std::sync::atomic::Ordering::SeqCst),
                    avg_response_time_ms: stat.avg_response_time_ms(),
                    timeouts: stat.timeouts.load(std::sync::atomic::Ordering::SeqCst),
                });
            }
        }

        result
    }

    pub async fn check_health(&self) {
        let backends = self.backends.read().await;
        for backend in backends.iter() {
            let health_url = format!("{}/health", backend.target_url.trim_end_matches('/'));
            let client = self.http_client.clone();
            let backend_clone = backend.clone();
            
            tokio::spawn(async move {
                let response = client.get(&health_url)
                    .timeout(std::time::Duration::from_secs(5))
                    .send()
                    .await;
                
                match response {
                    Ok(resp) => backend_clone.set_healthy(resp.status().is_success()),
                    Err(_) => backend_clone.set_healthy(false),
                }
            });
        }
    }
}
