use crate::models::*;
use std::collections::{HashMap, BTreeMap};
use std::sync::Arc;
use tokio::sync::RwLock;

pub const HEALTH_CHECK_INTERVAL_SECS: u64 = 15;
pub const RECENT_REQUESTS_WINDOW: usize = 100;
pub const ERROR_RATE_THRESHOLD: f64 = 0.20;
pub const DEGRADED_RECOVERY_CHECKS: u32 = 5;

#[derive(Default)]
pub struct AppStateInner {
    pub backends: HashMap<String, BackendService>,
    pub routes_exact: HashMap<String, Route>,
    pub routes_prefix: BTreeMap<String, Route>,
    pub api_keys: HashMap<String, ApiKey>,
    pub stats: HashMap<String, BackendStats>,
}

#[derive(Clone, Default)]
pub struct AppState {
    pub inner: Arc<RwLock<AppStateInner>>,
}

impl AppState {
    pub async fn register_backend(&self, name: String, url: String) -> BackendService {
        let backend = BackendService {
            id: uuid::Uuid::new_v4(),
            name: name.clone(),
            url,
            status: BackendStatus::Pending,
            consecutive_healthy_checks: 0,
        };
        
        let mut inner = self.inner.write().await;
        inner.backends.insert(name, backend.clone());
        inner.stats.insert(backend.name.clone(), BackendStats::default());
        backend
    }

    pub async fn unregister_backend(&self, name: &str) -> Option<BackendService> {
        let mut inner = self.inner.write().await;
        inner.backends.remove(name)
    }

    pub async fn get_backend(&self, name: &str) -> Option<BackendService> {
        let inner = self.inner.read().await;
        inner.backends.get(name).cloned()
    }

    pub async fn get_all_backends(&self) -> Vec<BackendService> {
        let inner = self.inner.read().await;
        inner.backends.values().cloned().collect()
    }

    pub async fn update_backend_status(&self, name: &str, status: BackendStatus) {
        let mut inner = self.inner.write().await;
        if let Some(backend) = inner.backends.get_mut(name) {
            backend.status = status;
            if status == BackendStatus::Online {
                backend.consecutive_healthy_checks = 0;
            }
        }
    }

    pub async fn increment_healthy_checks(&self, name: &str) -> u32 {
        let mut inner = self.inner.write().await;
        if let Some(backend) = inner.backends.get_mut(name) {
            backend.consecutive_healthy_checks += 1;
            backend.consecutive_healthy_checks
        } else {
            0
        }
    }

    pub async fn reset_healthy_checks(&self, name: &str) {
        let mut inner = self.inner.write().await;
        if let Some(backend) = inner.backends.get_mut(name) {
            backend.consecutive_healthy_checks = 0;
        }
    }

    pub async fn add_route(&self, route: Route) {
        let mut inner = self.inner.write().await;
        match route.match_type {
            MatchType::Exact => {
                inner.routes_exact.insert(route.path.clone(), route);
            }
            MatchType::Prefix => {
                inner.routes_prefix.insert(route.path.clone(), route);
            }
        }
    }

    pub async fn remove_route(&self, id: &uuid::Uuid) -> bool {
        let mut inner = self.inner.write().await;
        let mut removed = false;
        
        let exact_keys: Vec<String> = inner
            .routes_exact
            .iter()
            .filter(|(_, r)| r.id == *id)
            .map(|(k, _)| k.clone())
            .collect();
        
        for key in exact_keys {
            inner.routes_exact.remove(&key);
            removed = true;
        }
        
        let prefix_keys: Vec<String> = inner
            .routes_prefix
            .iter()
            .filter(|(_, r)| r.id == *id)
            .map(|(k, _)| k.clone())
            .collect();
        
        for key in prefix_keys {
            inner.routes_prefix.remove(&key);
            removed = true;
        }
        
        removed
    }

    pub async fn find_route(&self, path: &str) -> Option<Route> {
        let inner = self.inner.read().await;
        
        if let Some(route) = inner.routes_exact.get(path) {
            return Some(route.clone());
        }
        
        let mut best_match: Option<Route> = None;
        for (prefix, route) in inner.routes_prefix.iter() {
            if path.starts_with(prefix) {
                let is_better = match &best_match {
                    Some(best) => prefix.len() > best.path.len(),
                    None => true,
                };
                if is_better {
                    best_match = Some(route.clone());
                }
            }
        }
        
        best_match
    }

    pub async fn get_all_routes(&self) -> Vec<Route> {
        let inner = self.inner.read().await;
        let mut routes = Vec::new();
        routes.extend(inner.routes_exact.values().cloned());
        routes.extend(inner.routes_prefix.values().cloned());
        routes
    }

    pub async fn register_api_key(&self, key: String) -> ApiKey {
        let api_key = ApiKey {
            key: key.clone(),
            created_at: std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .map(|d| d.as_secs())
                .unwrap_or(0),
        };
        
        let mut inner = self.inner.write().await;
        inner.api_keys.insert(key, api_key.clone());
        api_key
    }

    pub async fn validate_api_key(&self, key: &str) -> bool {
        let inner = self.inner.read().await;
        inner.api_keys.contains_key(key)
    }

    pub async fn get_all_api_keys(&self) -> Vec<ApiKey> {
        let inner = self.inner.read().await;
        inner.api_keys.values().cloned().collect()
    }

    pub async fn record_request(&self, backend_name: &str, is_error: bool) {
        let mut inner = self.inner.write().await;
        if let Some(stats) = inner.stats.get_mut(backend_name) {
            stats.total_requests += 1;
            if is_error {
                stats.error_requests += 1;
            }
            stats.recent_requests.push(!is_error);
            if stats.recent_requests.len() > RECENT_REQUESTS_WINDOW {
                stats.recent_requests.remove(0);
            }
        }
    }

    pub async fn get_stats(&self, backend_name: &str) -> Option<BackendStats> {
        let inner = self.inner.read().await;
        inner.stats.get(backend_name).cloned()
    }

    pub async fn get_all_stats(&self) -> Vec<(String, BackendStats)> {
        let inner = self.inner.read().await;
        inner
            .stats
            .iter()
            .map(|(k, v)| (k.clone(), v.clone()))
            .collect()
    }

    pub async fn calculate_error_rate(&self, backend_name: &str) -> f64 {
        let inner = self.inner.read().await;
        if let Some(stats) = inner.stats.get(backend_name) {
            if stats.recent_requests.is_empty() {
                return 0.0;
            }
            let total = stats.recent_requests.len();
            let errors = stats.recent_requests.iter().filter(|&&ok| !ok).count();
            errors as f64 / total as f64
        } else {
            0.0
        }
    }
}
