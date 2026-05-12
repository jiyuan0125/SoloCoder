use std::collections::HashMap;
use std::sync::Arc;
use std::time::Duration;

use tokio::sync::RwLock;

const HEALTH_CHECK_INTERVAL: u64 = 15;
const MAX_FAILURES: u32 = 3;

#[derive(Debug, Clone)]
pub struct HealthStatus {
    pub is_healthy: bool,
    pub failure_count: u32,
}

#[derive(Debug, Clone)]
pub struct HealthChecker {
    statuses: Arc<RwLock<HashMap<String, HealthStatus>>>,
    client: reqwest::Client,
}

impl HealthChecker {
    pub fn new() -> Self {
        Self {
            statuses: Arc::new(RwLock::new(HashMap::new())),
            client: reqwest::Client::builder()
                .timeout(Duration::from_secs(5))
                .build()
                .unwrap_or_default(),
        }
    }

    pub async fn is_healthy(&self, backend_url: &str) -> bool {
        let statuses = self.statuses.read().await;
        statuses
            .get(backend_url)
            .map(|s| s.is_healthy)
            .unwrap_or(true)
    }

    pub async fn register_backend(&self, backend_url: &str) {
        let mut statuses = self.statuses.write().await;
        if !statuses.contains_key(backend_url) {
            statuses.insert(
                backend_url.to_string(),
                HealthStatus {
                    is_healthy: true,
                    failure_count: 0,
                },
            );
        }
    }

    pub async fn start(&self) {
        let statuses = self.statuses.clone();
        let client = self.client.clone();

        tokio::spawn(async move {
            loop {
                tokio::time::sleep(Duration::from_secs(HEALTH_CHECK_INTERVAL)).await;

                let backends: Vec<String> = {
                    let s = statuses.read().await;
                    s.keys().cloned().collect()
                };

                for backend in backends {
                    let health_url = format!("{}/health", backend.trim_end_matches('/'));

                    let result = client.get(&health_url).send().await;
                    let is_success = match result {
                        Ok(response) => response.status().is_success(),
                        Err(_) => false,
                    };

                    let mut s = statuses.write().await;
                    if let Some(status) = s.get_mut(&backend) {
                        if is_success {
                            status.failure_count = 0;
                            status.is_healthy = true;
                        } else {
                            status.failure_count += 1;
                            if status.failure_count >= MAX_FAILURES {
                                status.is_healthy = false;
                            }
                        }
                    }
                }
            }
        });
    }
}
