use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;
use std::time::{Instant, SystemTime, UNIX_EPOCH};

use crate::config::{InitialApiKey, JwtConfig, AuthMethod};

#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub struct ApiKeyMetadata {
    pub service_name: String,
    pub permissions: Vec<String>,
    pub created_at: u64,
}

pub struct ApiKeyStore {
    keys: RwLock<HashMap<String, ApiKeyMetadata>>,
}

impl ApiKeyStore {
    pub fn new() -> Self {
        Self {
            keys: RwLock::new(HashMap::new()),
        }
    }

    pub async fn from_initial(initial: &[InitialApiKey]) -> Self {
        let mut map = HashMap::new();
        for key in initial {
            map.insert(
                key.key.clone(),
                ApiKeyMetadata {
                    service_name: key.service_name.clone(),
                    permissions: key.permissions.clone(),
                    created_at: Self::now_secs(),
                },
            );
        }
        Self {
            keys: RwLock::new(map),
        }
    }

    fn now_secs() -> u64 {
        SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    }

    pub async fn add(&self, key: String, metadata: ApiKeyMetadata) {
        self.keys.write().await.insert(key, metadata);
    }

    pub async fn remove(&self, key: &str) -> bool {
        self.keys.write().await.remove(key).is_some()
    }

    pub async fn get(&self, key: &str) -> Option<ApiKeyMetadata> {
        self.keys.read().await.get(key).cloned()
    }

    pub async fn list(&self) -> Vec<(String, ApiKeyMetadata)> {
        self.keys
            .read()
            .await
            .iter()
            .map(|(k, v)| (k.clone(), v.clone()))
            .collect()
    }
}

pub struct JwtKeyStore {
    current_secret: RwLock<String>,
    old_secret: RwLock<Option<(String, Instant)>>,
    old_secret_ttl: std::time::Duration,
}

impl JwtKeyStore {
    pub fn new(config: &JwtConfig) -> Self {
        Self {
            current_secret: RwLock::new(config.current_secret.clone()),
            old_secret: RwLock::new(config.old_secret.as_ref().map(|s| {
                (s.clone(), Instant::now())
            })),
            old_secret_ttl: config.old_secret_expires_in(),
        }
    }

    pub async fn rotate_secret(&self, new_secret: String) {
        let mut current = self.current_secret.write().await;
        let mut old = self.old_secret.write().await;
        let old_value = std::mem::replace(&mut *current, new_secret);
        *old = Some((old_value, Instant::now()));
    }

    pub async fn get_current(&self) -> String {
        self.current_secret.read().await.clone()
    }

    pub async fn get_old(&self) -> Option<String> {
        let old = self.old_secret.read().await;
        if let Some((secret, created)) = old.as_ref() {
            if Instant::now().duration_since(*created) < self.old_secret_ttl {
                return Some(secret.clone());
            }
        }
        None
    }

    pub async fn get_all_valid(&self) -> Vec<String> {
        let mut result = Vec::new();
        result.push(self.get_current().await);
        if let Some(old) = self.get_old().await {
            result.push(old);
        }
        result
    }

    pub fn old_ttl(&self) -> std::time::Duration {
        self.old_secret_ttl
    }
}

#[derive(Clone)]
pub struct RouteEntry {
    pub path_pattern: String,
    pub backend: String,
    pub auth_methods: Vec<AuthMethod>,
}

pub struct RouteStore {
    routes: RwLock<Vec<RouteEntry>>,
    backends: RwLock<HashMap<String, String>>,
}

impl RouteStore {
    pub fn new(routes: Vec<RouteEntry>, backends: HashMap<String, String>) -> Self {
        Self {
            routes: RwLock::new(routes),
            backends: RwLock::new(backends),
        }
    }

    pub async fn match_route(&self, path: &str) -> Option<RouteEntry> {
        let routes = self.routes.read().await;
        for route in routes.iter() {
            if Self::path_matches(&route.path_pattern, path) {
                return Some(route.clone());
            }
        }
        None
    }

    fn path_matches(pattern: &str, path: &str) -> bool {
        if pattern.ends_with("/*") {
            let prefix = &pattern[..pattern.len() - 2];
            if path == prefix || path.starts_with(&format!("{}/", prefix)) {
                return true;
            }
        }
        pattern == path
    }

    pub async fn get_backend_url(&self, backend: &str) -> Option<String> {
        self.backends.read().await.get(backend).cloned()
    }
}

#[derive(Clone)]
pub struct AppState {
    pub api_keys: Arc<ApiKeyStore>,
    pub jwt_keys: Arc<JwtKeyStore>,
    pub routes: Arc<RouteStore>,
    pub admin_user: Arc<str>,
}
