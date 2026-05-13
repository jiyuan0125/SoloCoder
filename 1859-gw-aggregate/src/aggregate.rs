use std::collections::HashMap;
use std::time::Duration;

use axum::http::HeaderMap;
use futures::future::join_all;
use serde_json::Value;
use tokio::sync::RwLock;

use crate::types::{AggregateScene, BackendCall};
use crate::auth::AuthRegistry;

#[derive(Debug, Default)]
pub struct AggregateSceneRegistry {
    scenes: HashMap<String, AggregateScene>,
}

impl AggregateSceneRegistry {
    pub fn new() -> Self {
        let mut registry = Self::default();
        registry.add(default_home_scene());
        registry.add(default_profile_scene());
        registry
    }

    pub fn add(&mut self, scene: AggregateScene) {
        self.scenes.insert(scene.name.clone(), scene);
    }

    pub fn get(&self, name: &str) -> Option<&AggregateScene> {
        self.scenes.get(name)
    }

    pub fn remove(&mut self, name: &str) -> bool {
        self.scenes.remove(name).is_some()
    }

    pub fn list(&self) -> Vec<AggregateScene> {
        self.scenes.values().cloned().collect()
    }
}

fn default_home_scene() -> AggregateScene {
    AggregateScene {
        name: "home".to_string(),
        backends: vec![
            BackendCall {
                url: "http://localhost:8501/user".to_string(),
                timeout_secs: 5,
                field_name: "user".to_string(),
            },
            BackendCall {
                url: "http://localhost:8502/orders".to_string(),
                timeout_secs: 5,
                field_name: "orders".to_string(),
            },
            BackendCall {
                url: "http://localhost:8503/recommendations".to_string(),
                timeout_secs: 5,
                field_name: "recommendations".to_string(),
            },
        ],
    }
}

fn default_profile_scene() -> AggregateScene {
    AggregateScene {
        name: "profile".to_string(),
        backends: vec![
            BackendCall {
                url: "http://localhost:8501/user".to_string(),
                timeout_secs: 5,
                field_name: "user".to_string(),
            },
            BackendCall {
                url: "http://localhost:8502/orders".to_string(),
                timeout_secs: 5,
                field_name: "orders".to_string(),
            },
        ],
    }
}

pub async fn execute_aggregate(
    scene: AggregateScene,
    client: &reqwest::Client,
    _default_timeout: Duration,
    _auth_registry: &RwLock<AuthRegistry>,
    _headers: &HeaderMap,
) -> Result<Vec<(String, Result<Value, String>)>, String> {
    let mut handles = Vec::new();

    for backend in &scene.backends {
        let url = backend.url.clone();
        let field_name = backend.field_name.clone();
        let timeout = Duration::from_secs(backend.timeout_secs);
        let client = client.clone();

        let handle = tokio::spawn(async move {
            let result = tokio::time::timeout(timeout, async {
                match client.get(&url).send().await {
                    Err(e) => Err(format!("request error: {}", e)),
                    Ok(resp) => {
                        if resp.status().is_success() {
                            resp.json::<Value>()
                                .await
                                .map_err(|e| format!("parse error: {}", e))
                        } else {
                            Err(format!("backend error: status {}", resp.status()))
                        }
                    }
                }
            }).await;

            match result {
                Ok(Ok(value)) => (field_name, Ok(value)),
                Ok(Err(e)) => (field_name, Err(e)),
                Err(_) => (field_name, Err(format!("timeout after {:?}", timeout))),
            }
        });

        handles.push(handle);
    }

    let results = join_all(handles).await;
    
    let mut output = Vec::new();
    for result in results {
        match result {
            Ok((field, value)) => output.push((field, value)),
            Err(e) => return Err(format!("task join error: {}", e)),
        }
    }

    Ok(output)
}
