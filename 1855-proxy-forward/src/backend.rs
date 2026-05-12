use serde::{Deserialize, Serialize};
use std::sync::atomic::{AtomicBool, Ordering};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Backend {
    pub id: String,
    pub path_prefix: String,
    pub target_url: String,
    pub healthy: bool,
}

#[derive(Debug)]
pub struct BackendInternal {
    pub id: String,
    pub path_prefix: String,
    pub target_url: String,
    pub healthy: AtomicBool,
}

impl BackendInternal {
    pub fn new(path_prefix: String, target_url: String) -> Self {
        BackendInternal {
            id: Uuid::new_v4().to_string(),
            path_prefix,
            target_url,
            healthy: AtomicBool::new(true),
        }
    }

    pub fn is_healthy(&self) -> bool {
        self.healthy.load(Ordering::SeqCst)
    }

    pub fn set_healthy(&self, healthy: bool) {
        self.healthy.store(healthy, Ordering::SeqCst);
    }

    pub fn to_external(&self) -> Backend {
        Backend {
            id: self.id.clone(),
            path_prefix: self.path_prefix.clone(),
            target_url: self.target_url.clone(),
            healthy: self.is_healthy(),
        }
    }
}
