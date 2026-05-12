use std::collections::HashSet;
use std::sync::Arc;

use tokio::sync::RwLock;
use uuid::Uuid;

#[derive(Debug, Clone)]
pub struct ApiKeyRegistry {
    keys: Arc<RwLock<HashSet<String>>>,
}

impl ApiKeyRegistry {
    pub fn new() -> Self {
        Self {
            keys: Arc::new(RwLock::new(HashSet::new())),
        }
    }

    pub async fn create_key(&self) -> String {
        let key = Uuid::new_v4().to_string();
        let mut keys = self.keys.write().await;
        keys.insert(key.clone());
        key
    }

    pub async fn validate_key(&self, key: &str) -> bool {
        let keys = self.keys.read().await;
        keys.contains(key)
    }
}
