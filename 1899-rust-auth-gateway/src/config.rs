use serde::{Deserialize, Deserializer, Serialize};
use std::collections::HashMap;
use std::time::Duration;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AppConfig {
    pub routes: HashMap<String, RouteConfig>,
    pub backends: HashMap<String, BackendConfig>,
    pub jwt: JwtConfig,
    pub initial_api_keys: Vec<InitialApiKey>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RouteConfig {
    pub backend: String,
    pub auth_methods: Vec<AuthMethod>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize)]
pub enum AuthMethod {
    ApiKey,
    Jwt,
}

impl<'de> Deserialize<'de> for AuthMethod {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: Deserializer<'de>,
    {
        let s = String::deserialize(deserializer)?;
        match s.to_lowercase().as_str() {
            "api_key" | "apikey" | "api-key" => Ok(AuthMethod::ApiKey),
            "jwt" => Ok(AuthMethod::Jwt),
            _ => Err(serde::de::Error::custom(format!("unknown auth method: {}", s))),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BackendConfig {
    pub url: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct JwtConfig {
    pub current_secret: String,
    pub old_secret: Option<String>,
    pub old_secret_expires_in_seconds: Option<u64>,
    pub admin_user: Option<String>,
}

impl JwtConfig {
    pub fn old_secret_expires_in(&self) -> Duration {
        Duration::from_secs(self.old_secret_expires_in_seconds.unwrap_or(3600))
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InitialApiKey {
    pub key: String,
    pub service_name: String,
    pub permissions: Vec<String>,
}

impl Default for AppConfig {
    fn default() -> Self {
        let mut routes = HashMap::new();
        routes.insert(
            "/api/*".to_string(),
            RouteConfig {
                backend: "default".to_string(),
                auth_methods: vec![AuthMethod::ApiKey, AuthMethod::Jwt],
            },
        );

        let mut backends = HashMap::new();
        backends.insert(
            "default".to_string(),
            BackendConfig {
                url: "http://localhost:8081".to_string(),
            },
        );

        Self {
            routes,
            backends,
            jwt: JwtConfig {
                current_secret: "super-secret-key-change-me-please".to_string(),
                old_secret: None,
                old_secret_expires_in_seconds: Some(3600),
                admin_user: Some("admin".to_string()),
            },
            initial_api_keys: vec![],
        }
    }
}

impl AppConfig {
    pub fn load() -> Result<Self, String> {
        Ok(Self::default())
    }
}
