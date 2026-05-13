use crate::models::{ServiceInstance, Subscription};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;
use uuid::Uuid;

#[derive(Debug, Clone)]
pub struct AppState {
    pub services: Arc<RwLock<HashMap<String, HashMap<Uuid, ServiceInstance>>>>,
    pub subscriptions: Arc<RwLock<HashMap<Uuid, Subscription>>>,
}

impl AppState {
    pub fn new() -> Self {
        Self {
            services: Arc::new(RwLock::new(HashMap::new())),
            subscriptions: Arc::new(RwLock::new(HashMap::new())),
        }
    }
}

impl Default for AppState {
    fn default() -> Self {
        Self::new()
    }
}
