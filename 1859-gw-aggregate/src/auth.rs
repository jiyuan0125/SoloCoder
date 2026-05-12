use std::collections::HashMap;

use axum::http::HeaderMap;

use crate::types::AuthPolicy;

#[derive(Debug, Clone, Default)]
pub struct ApiKeyRegistry {
    keys: HashMap<String, String>,
}

#[derive(Debug, Clone, Default)]
pub struct AuthRegistry {
    api_keys: ApiKeyRegistry,
}

impl AuthRegistry {
    pub fn new() -> Self {
        Self {
            api_keys: ApiKeyRegistry::default(),
        }
    }

    pub fn add_api_key(&mut self, key: String, value: String) {
        self.api_keys.keys.insert(key, value);
    }

    pub fn validate(&self, policy: &AuthPolicy, headers: &HeaderMap) -> bool {
        match policy {
            AuthPolicy::None => true,
            AuthPolicy::ApiKey { key } => {
                let header_key = format!("x-api-key");
                match headers.get(&header_key) {
                    Some(header_value) => {
                        header_value.to_str().ok() == Some(key.as_str())
                    }
                    None => false,
                }
            }
        }
    }
}
